# Health Probes Configuration

## Overview

Kubernetes health probes ensure containers are running correctly and ready to serve traffic.

## Probe Types

### Readiness Probe

**Purpose:** Determine if pod should receive traffic

**Behavior:**
- Pod removed from Service endpoints when failing
- Traffic stops routing to pod
- Pod remains running

**Use Case:** Node is alive but not ready (syncing, recovering, overloaded)

### Liveness Probe

**Purpose:** Detect hung or deadlocked processes

**Behavior:**
- Pod restarted when failing
- Drastic action - data loss possible

**Use Case:** Process completely stuck, not responding

### Startup Probe

**Purpose:** Allow long initialization without false failures

**Behavior:**
- Disables liveness/readiness until succeeds
- Generous timeout for initial sync

**Use Case:** Initial blockchain sync can take minutes

## Sequencer Configuration

### Recommended Setup

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cdk-erigon-sequencer
spec:
  template:
    spec:
      containers:
      - name: erigon
        ports:
        - name: rpc
          containerPort: 8545

        startupProbe:
          httpGet:
            path: /
            port: rpc
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 30  # 5 minutes total

        livenessProbe:
          httpGet:
            path: /
            port: rpc
          initialDelaySeconds: 60
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3  # 90s to recover

        readinessProbe:
          httpGet:
            path: /
            port: rpc
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3  # 30s to become ready
```

### Timing Explanation

**startupProbe:**
- `failureThreshold: 30` × `periodSeconds: 10` = 5 minutes
- Allows initial sync without restarts
- After success, liveness/readiness take over

**livenessProbe:**
- `periodSeconds: 30` - Check every 30s (not too aggressive)
- `timeoutSeconds: 10` - Allow slow responses under load
- `failureThreshold: 3` - 90s total before restart

**readinessProbe:**
- `periodSeconds: 10` - Frequent checks for quick recovery
- `timeoutSeconds: 5` - Fast detection of issues
- `failureThreshold: 3` - 30s to become ready

## RPC Node Configuration

### Recommended Setup

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cdk-erigon-rpc
spec:
  template:
    spec:
      containers:
      - name: erigon
        ports:
        - name: rpc
          containerPort: 8545

        startupProbe:
          httpGet:
            path: /
            port: rpc
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 60  # 10 minutes for full sync

        livenessProbe:
          httpGet:
            path: /
            port: rpc
          initialDelaySeconds: 60
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /
            port: rpc
          initialDelaySeconds: 20
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2  # Quick detection for load balancing
```

### Key Differences from Sequencer

- **Longer startupProbe** (10 min vs 5 min) - RPC sync takes longer
- **Faster readinessProbe** (5s vs 10s) - More responsive load balancing
- **Lower readiness failureThreshold** (2 vs 3) - Quick removal from pool

## Probe Endpoints

### HTTP Probe (Recommended)

**Endpoint:** `http://localhost:8545/`

**Expected Response:**
- Status: 200 OK or 405 Method Not Allowed
- Body: JSON-RPC error (GET not supported)

**Why It Works:**
- RPC server always responds to HTTP requests
- Alive = responds (even with error)
- Not ready = timeout or connection refused

### Alternative: JSON-RPC Probe

**More Accurate but Heavier:**

```yaml
readinessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - |
      curl -s -X POST http://localhost:8545 \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
        | grep -q '"result"'
```

**Pros:** Tests actual RPC functionality
**Cons:** Higher overhead, slower

## Helm Values

### Values Schema

```yaml
sequencer:
  probes:
    startup:
      enabled: true
      initialDelaySeconds: 10
      periodSeconds: 10
      timeoutSeconds: 5
      failureThreshold: 30

    liveness:
      enabled: true
      initialDelaySeconds: 60
      periodSeconds: 30
      timeoutSeconds: 10
      failureThreshold: 3

    readiness:
      enabled: true
      initialDelaySeconds: 30
      periodSeconds: 10
      timeoutSeconds: 5
      failureThreshold: 3

rpc:
  probes:
    startup:
      enabled: true
      initialDelaySeconds: 10
      periodSeconds: 10
      timeoutSeconds: 5
      failureThreshold: 60

    liveness:
      enabled: true
      initialDelaySeconds: 60
      periodSeconds: 30
      timeoutSeconds: 10
      failureThreshold: 3

    readiness:
      enabled: true
      initialDelaySeconds: 20
      periodSeconds: 5
      timeoutSeconds: 3
      failureThreshold: 2
```

### Helm Template

```yaml
{{- if .Values.sequencer.probes.startupProbe.enabled }}
startupProbe:
  httpGet:
    path: /
    port: rpc
  initialDelaySeconds: {{ .Values.sequencer.probes.startup.initialDelaySeconds }}
  periodSeconds: {{ .Values.sequencer.probes.startup.periodSeconds }}
  timeoutSeconds: {{ .Values.sequencer.probes.startup.timeoutSeconds }}
  failureThreshold: {{ .Values.sequencer.probes.startup.failureThreshold }}
{{- end }}
```

## Tuning Guidelines

### When to Increase Timeouts

**Symptom:** Probes failing during normal operation

**Scenarios:**
- High load causing slow responses
- Database queries taking longer
- Network latency issues

**Solution:**
```yaml
readinessProbe:
  timeoutSeconds: 10  # Increase from 5
  failureThreshold: 5  # More tolerance
```

### When to Decrease Periods

**Symptom:** Slow detection of issues

**Scenarios:**
- Need faster load balancer updates
- Quick failover requirements
- High availability priority

**Solution:**
```yaml
readinessProbe:
  periodSeconds: 5  # Decrease from 10
  failureThreshold: 2  # Quicker removal
```

### When to Disable Startup Probe

**Scenario:** Fast startup, no initial sync needed

**Example:** Pre-synced datadir mounted

```yaml
startupProbe:
  enabled: false  # Skip for already-synced nodes
```

## Troubleshooting

### Pod Stuck in Init State

**Symptom:** Pod never becomes ready

**Debug:**
```bash
# Check probe status
kubectl describe pod <pod-name>

# Look for probe failures
kubectl get events --field-selector involvedObject.name=<pod-name>

# Test endpoint manually
kubectl exec <pod-name> -- curl -v http://localhost:8545/
```

**Common Causes:**
1. Port not exposed: Add `ports:` section
2. Service not starting: Check logs
3. Firewall blocking: Verify network policy

### Frequent Restarts

**Symptom:** Pod restarts every few minutes

**Debug:**
```bash
# Check restart reason
kubectl get pod <pod-name> -o jsonpath='{.status.containerStatuses[0].lastState.terminated.reason}'

# Review liveness probe config
kubectl get pod <pod-name> -o yaml | grep -A 10 livenessProbe
```

**Common Causes:**
1. Liveness probe too aggressive: Increase timeout/failureThreshold
2. Process actually hung: Check logs for deadlocks
3. OOMKilled: Increase memory limits

### Slow Traffic Restoration

**Symptom:** Pod takes too long to receive traffic after restart

**Debug:**
```bash
# Check readiness probe timing
kubectl describe pod <pod-name> | grep -A 5 Readiness

# Test endpoint response time
time kubectl exec <pod-name> -- curl http://localhost:8545/
```

**Solutions:**
1. Decrease `periodSeconds` for faster checks
2. Decrease `initialDelaySeconds` if startup is fast
3. Optimize application startup time

## Best Practices

1. **Always use startupProbe for blockchain nodes**
   - Initial sync takes time
   - Prevents liveness probe from killing during startup

2. **Keep liveness probe conservative**
   - Restarting loses state
   - Only restart when truly hung
   - Use 30s+ periods, 3+ failures

3. **Make readiness probe responsive**
   - Quick detection improves load balancing
   - Use 5-10s periods
   - Lower failureThreshold acceptable

4. **Test probes under load**
   - Ensure timeouts sufficient during peak
   - Validate no false positives
   - Check probe impact on performance

5. **Monitor probe failures**
   - Alert on repeated probe failures
   - Track probe latency
   - Correlate with application metrics

## Validation

### Testing Probes

```bash
# 1. Deploy with probes
kubectl apply -f statefulset.yaml

# 2. Wait for startup probe to succeed
kubectl wait --for=condition=Ready pod/<pod-name> --timeout=5m

# 3. Verify readiness
kubectl get pod <pod-name> -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'

# 4. Test liveness detection (simulate hang)
kubectl exec <pod-name> -- kill -STOP 1
# Should restart after ~90s

# 5. Test readiness detection (stop RPC)
kubectl exec <pod-name> -- kill -STOP <rpc-pid>
# Should remove from endpoints after ~30s
```

### Expected Behavior

**Startup:**
- Pod starts → startupProbe checks every 10s
- After success → liveness/readiness take over
- Total startup time: < failureThreshold × periodSeconds

**Running:**
- Readiness checks every 5-10s
- Liveness checks every 30s
- Pod in Service endpoints when ready

**Shutdown:**
- Probes continue during preStop
- Pod removed from endpoints immediately
- Graceful shutdown proceeds

## References

- [Kubernetes Probes Documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Container Requirements](container-requirements.md)
- [Shutdown Sequence](shutdown-sequence.md)
