# PreStop Hook Specification

## Overview

PreStop hooks execute before SIGTERM, allowing graceful connection draining before shutdown begins.

## Recommended Configuration

### Simple Delay (Recommended)

```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 5"]
```

**Why:** 5 second delay allows:
- Load balancers to detect pod removal
- In-flight requests to complete
- New connections to stop arriving

### HTTP Endpoint (Alternative)

```yaml
lifecycle:
  preStop:
    httpGet:
      path: /shutdown
      port: 8545
```

**Requires:** CDK Erigon implementation of `/shutdown` endpoint

## Sequencer Configuration

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cdk-erigon-sequencer
spec:
  template:
    spec:
      terminationGracePeriodSeconds: 60
      containers:
      - name: erigon
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 5"]
```

**Timing Budget:**
- PreStop: 5s (connection drain)
- SIGTERM → NATS flush: 10-15s
- Process exit: 5s
- Total: 25s (within 60s grace period)

## RPC Node Configuration

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cdk-erigon-rpc
spec:
  template:
    spec:
      terminationGracePeriodSeconds: 30
      containers:
      - name: erigon
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 3"]
```

**Timing Budget:**
- PreStop: 3s (connection drain)
- SIGTERM → shutdown: 5-10s
- Process exit: 2s
- Total: 15s (within 30s grace period)

## When NOT to Use PreStop

- Pods with no external traffic
- Batch jobs
- Init containers
- Sidecar containers (use main container's lifecycle)

## Troubleshooting

### Pod Takes Full Grace Period to Terminate

**Symptom:** Pod always takes 60s to terminate

**Cause:** PreStop timeout or SIGTERM handler hung

**Solution:**
1. Check PreStop duration: `kubectl logs <pod> --previous`
2. Verify SIGTERM handling: Look for graceful shutdown logs
3. Reduce grace period if PreStop + shutdown < 30s

### Connections Still Dropping During Shutdown

**Symptom:** Clients see connection errors during rolling updates

**Cause:** Insufficient PreStop delay

**Solution:**
1. Increase PreStop delay: `sleep 10`
2. Ensure load balancer respects pod readiness
3. Add `preStop` to readiness probe removal

## Helm Values

```yaml
sequencer:
  lifecycle:
    preStop:
      enabled: true
      delay: 5  # seconds

rpc:
  lifecycle:
    preStop:
      enabled: true
      delay: 3  # seconds
```

## References

- [Kubernetes Container Lifecycle Hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/)
- [Shutdown Sequence Documentation](shutdown-sequence.md)
- [Graceful Shutdown Best Practices](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
