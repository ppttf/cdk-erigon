# Security Context Configuration

## Overview

CDK Erigon container runs as non-root user for security best practices. This document reviews the current configuration and provides Kubernetes security context recommendations.

## Current Dockerfile Configuration

### User Configuration

```dockerfile
ARG UID=1000
ARG GID=1000

RUN adduser -D -u ${UID} -g ${GID} erigon
USER erigon
WORKDIR /home/erigon
```

**Review:** ✅ Correct
- Non-root user (UID 1000)
- Dedicated user account
- Home directory ownership

### File Permissions

**Directories Created:**
- `/home/erigon` - User home (owned by erigon:erigon)
- `~/.local/share/erigon` - Default datadir (owned by erigon:erigon)

**Review:** ✅ Correct
- User can read/write to required directories
- No root privileges needed for normal operation

### Capabilities

**Required:** None
**Dockerfile:** No special capabilities added

**Review:** ✅ Correct
- No CAP_NET_ADMIN, CAP_SYS_ADMIN, etc.
- Standard network ports (>1024) don't require privileges
- Port 8545 accessible without root

## Kubernetes Security Context

### Pod Security Context (Recommended)

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cdk-erigon-sequencer
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault
```

**Explanation:**
- `runAsNonRoot: true` - Enforce non-root (fail if root)
- `runAsUser: 1000` - Explicit UID (matches Dockerfile)
- `runAsGroup: 1000` - Primary GID
- `fsGroup: 1000` - Volume ownership (critical for PVs)
- `seccompProfile` - Restrict syscalls (security hardening)

### Container Security Context (Recommended)

```yaml
containers:
- name: erigon
  securityContext:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: false  # NATS needs writable datadir
    capabilities:
      drop:
      - ALL
```

**Explanation:**
- `allowPrivilegeEscalation: false` - Prevent setuid/setgid
- `readOnlyRootFilesystem: false` - NATS writes to datadir
- `drop: ALL` - Remove all capabilities (none needed)

### Why Not Read-Only Root Filesystem?

**NATS JetStream Requirements:**
- Writes to datadir continuously
- Creates/updates stream files
- Manages consumer state

**Options:**
1. **Use volume mount** (recommended):
```yaml
volumeMounts:
- name: data
  mountPath: /home/erigon/.local/share/erigon
readOnlyRootFilesystem: true  # Now possible
```

2. **Use emptyDir for NATS**:
```yaml
volumeMounts:
- name: nats-data
  mountPath: /home/erigon/.local/share/erigon/datastream
readOnlyRootFilesystem: true  # Root is read-only
```

3. **Keep root writable**:
```yaml
readOnlyRootFilesystem: false  # Simpler, less secure
```

## Volume Permissions

### PersistentVolume Configuration

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: erigon-data
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
  storageClassName: fast-ssd
```

### Volume Ownership

**With fsGroup:**
```yaml
spec:
  securityContext:
    fsGroup: 1000  # Automatically chowns volume to GID 1000
```

**Volumes mounted with group ownership:**
- `/data` → `root:1000` with `rwxrwxr-x`
- User `erigon` (GID 1000) can write

**Without fsGroup:**
- Volume owned by `root:root`
- User `erigon` cannot write
- Pod fails to start

### Storage Class Considerations

**Some storage classes require fsGroup:**
- AWS EBS
- GCE Persistent Disk
- Azure Disk

**Some ignore fsGroup:**
- NFS (uses root_squash)
- HostPath (uses host permissions)

## Network Security

### Port Exposure

**Container Ports:**
```yaml
ports:
- name: rpc
  containerPort: 8545
  protocol: TCP
- name: nats
  containerPort: 4222
  protocol: TCP
- name: nats-monitor
  containerPort: 8222
  protocol: TCP
```

**Review:** ✅ All ports >1024 (no root needed)

### Network Policies (Optional)

**Sequencer:**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: sequencer-netpol
spec:
  podSelector:
    matchLabels:
      app: cdk-erigon-sequencer
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: cdk-erigon-rpc
    ports:
    - port: 4222  # NATS only from RPC nodes
  - from: []  # RPC port open to all
    ports:
    - port: 8545
```

## Pod Security Standards

### Restricted Profile (Recommended)

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: cdk-erigon
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

**Requirements Met:**
- ✅ Runs as non-root
- ✅ No privilege escalation
- ✅ Drops all capabilities
- ✅ Uses seccomp
- ⚠️ May need volume exceptions for writability

### Baseline Profile (Fallback)

If restricted fails due to volume issues:

```yaml
labels:
  pod-security.kubernetes.io/enforce: baseline
```

## Helm Values

### Security Configuration

```yaml
securityContext:
  pod:
    runAsNonRoot: true
    runAsUser: 1000
    runAsGroup: 1000
    fsGroup: 1000
    seccompProfile:
      type: RuntimeDefault

  container:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: false
    capabilities:
      drop:
      - ALL

persistence:
  enabled: true
  storageClass: fast-ssd
  size: 100Gi
  accessMode: ReadWriteOnce
```

### Helm Template

```yaml
securityContext:
  runAsNonRoot: {{ .Values.securityContext.pod.runAsNonRoot }}
  runAsUser: {{ .Values.securityContext.pod.runAsUser }}
  runAsGroup: {{ .Values.securityContext.pod.runAsGroup }}
  fsGroup: {{ .Values.securityContext.pod.fsGroup }}
  {{- with .Values.securityContext.pod.seccompProfile }}
  seccompProfile:
    {{- toYaml . | nindent 4 }}
  {{- end }}

containers:
- name: erigon
  securityContext:
    allowPrivilegeEscalation: {{ .Values.securityContext.container.allowPrivilegeEscalation }}
    readOnlyRootFilesystem: {{ .Values.securityContext.container.readOnlyRootFilesystem }}
    {{- with .Values.securityContext.container.capabilities }}
    capabilities:
      {{- toYaml . | nindent 6 }}
    {{- end }}
```

## Validation

### Testing Security Context

```bash
# 1. Deploy with security context
kubectl apply -f statefulset.yaml

# 2. Verify user
kubectl exec <pod> -- id
# Expected: uid=1000(erigon) gid=1000(erigon)

# 3. Verify no root
kubectl exec <pod> -- whoami
# Expected: erigon

# 4. Verify cannot escalate
kubectl exec <pod> -- su root
# Expected: su: must be suid to work properly (DENIED)

# 5. Verify volume permissions
kubectl exec <pod> -- ls -ld /data
# Expected: drwxrwxr-x 2 root erigon ... /data

# 6. Test write access
kubectl exec <pod> -- touch /data/test.txt
# Expected: Success

# 7. Verify capabilities
kubectl exec <pod> -- cat /proc/1/status | grep Cap
# Expected: All zeros (no capabilities)
```

### Common Issues

#### Permission Denied Writing to Volume

**Symptom:**
```
[ERROR] Failed to write to /data: permission denied
```

**Cause:** Missing `fsGroup`

**Fix:**
```yaml
securityContext:
  fsGroup: 1000  # Add this
```

#### Pod Security Standards Violation

**Symptom:**
```
Error: pods "cdk-erigon-0" is forbidden: violates PodSecurity "restricted:latest"
```

**Causes:**
1. Missing `seccompProfile`
2. `allowPrivilegeEscalation` not set to false
3. Capabilities not dropped

**Fix:** Use complete security context from above

## Security Best Practices

1. **Always run as non-root**
   - Dockerfile: `USER erigon`
   - K8s: `runAsNonRoot: true`

2. **Drop all capabilities**
   - None required for normal operation
   - `drop: [ALL]`

3. **Use fsGroup for volumes**
   - Ensures user can write
   - `fsGroup: 1000`

4. **Enable seccomp**
   - Restricts syscalls
   - `type: RuntimeDefault`

5. **Prevent privilege escalation**
   - No setuid/setgid
   - `allowPrivilegeEscalation: false`

6. **Use Network Policies**
   - Restrict NATS to RPC nodes only
   - Allow RPC from authorized sources

7. **Regular security scans**
   - Scan image for CVEs
   - Update base image regularly

8. **Secrets management**
   - Never hardcode secrets
   - Use Kubernetes Secrets
   - Consider External Secrets Operator

## Compliance

### CIS Kubernetes Benchmark

**Compliance Status:**
- ✅ 5.2.1: Minimize admission of privileged containers
- ✅ 5.2.2: Minimize admission of containers with capabilities
- ✅ 5.2.3: Minimize admission of containers with root privileges
- ✅ 5.2.4: Minimize admission of containers with privilege escalation
- ✅ 5.2.5: Minimize admission of containers with seccomp unconfined
- ⚠️ 5.2.6: Minimize admission of containers with read-only root filesystem (NATS requires writable)

### Pod Security Standards

**Compliance Level:** Baseline (✅) / Restricted (⚠️)

**Restricted Exceptions:**
- `readOnlyRootFilesystem: false` required for NATS
- Consider using volume mount workaround for full compliance

## References

- [Kubernetes Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Container Requirements](container-requirements.md)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
