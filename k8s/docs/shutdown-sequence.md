# Shutdown Sequence Documentation

## Overview

Complete shutdown flow for CDK Erigon with embedded NATS JetStream, ensuring zero message loss.

## Sequencer Shutdown Sequence

### Timeline

```
t=0s    Kubernetes initiates pod termination
        ├─ Pod marked for deletion
        ├─ Removed from Service endpoints
        └─ PreStop hook executes

t=5s    PreStop completes
        └─ SIGTERM sent to cdk-erigon process

t=5-10s NATS JetStream flush begins
        ├─ Stop accepting new connections
        ├─ Complete in-flight transactions
        ├─ Flush pending messages to disk
        └─ Close JetStream streams cleanly

t=10-15s Sequencer shutdown
        ├─ Stop block production
        ├─ Finalize last block
        ├─ Close database connections
        └─ Write shutdown marker

t=15-20s Process exits cleanly
        └─ Exit code 0

t=60s   SIGKILL (if process hung)
        └─ Force termination
```

### Detailed Steps

#### 1. PreStop Hook (t=0-5s)

**Action:** Execute `/bin/sh -c "sleep 5"`

**Purpose:**
- Allow load balancer to detect pod removal
- Drain existing connections
- Stop new traffic from arriving

**Logs:**
```
No logs - simple sleep command
```

#### 2. SIGTERM Received (t=5s)

**Action:** CDK Erigon receives SIGTERM signal

**Purpose:** Initiate graceful shutdown

**Expected Logs:**
```
[INFO] Received SIGTERM, initiating graceful shutdown
[INFO] Stopping block production
[INFO] Flushing NATS JetStream...
```

#### 3. NATS JetStream Flush (t=5-15s)

**Action:** Persist all pending messages

**Critical Operations:**
- Flush all streams to disk
- Complete pending writes
- Close consumer connections cleanly
- Write stream state metadata

**Expected Logs:**
```
[INFO] NATS server shutting down
[INFO] JetStream flushing streams...
[INFO] Stream 'datastream' flushed: 12345 messages
[INFO] All streams persisted successfully
```

**Failure Indicators:**
```
[ERROR] Stream flush timeout
[ERROR] Failed to persist message sequence 12345
[WARN] Forcing shutdown, potential data loss
```

#### 4. Sequencer Shutdown (t=10-15s)

**Action:** Stop block production gracefully

**Operations:**
- Finalize current block if partially complete
- Close state database connections
- Persist sequencer state
- Write shutdown timestamp

**Expected Logs:**
```
[INFO] Finalizing block #12345
[INFO] Block #12345 complete
[INFO] Closing database connections
[INFO] Sequencer state persisted
```

#### 5. Process Exit (t=15-20s)

**Action:** Clean process termination

**Expected:**
- Exit code: 0
- All resources released
- No orphaned connections

**Logs:**
```
[INFO] Shutdown complete
[INFO] Exit code: 0
```

#### 6. SIGKILL Failsafe (t=60s)

**Action:** Kubernetes force-kills if still running

**Indicates:**
- Shutdown timeout exceeded
- Process hung or deadlocked
- Requires investigation

## RPC Node Shutdown Sequence

### Timeline

```
t=0s    PreStop hook (3s connection drain)
t=3s    SIGTERM received
t=3-8s  Disconnect from NATS, close connections
t=8-10s Process exits
t=30s   SIGKILL failsafe
```

### Key Differences from Sequencer

- **No NATS flush** (consumer, not publisher)
- **Shorter grace period** (30s vs 60s)
- **Faster shutdown** (10s vs 20s)

### RPC Shutdown Logs

```
[INFO] Received SIGTERM
[INFO] Stopping RPC server
[INFO] Disconnecting from NATS server
[INFO] Closing NATS consumer
[INFO] Database connections closed
[INFO] Shutdown complete
```

## Configuration

### Kubernetes StatefulSet

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

### Helm Values

```yaml
sequencer:
  terminationGracePeriodSeconds: 60
  lifecycle:
    preStop:
      delay: 5

rpc:
  terminationGracePeriodSeconds: 30
  lifecycle:
    preStop:
      delay: 3
```

## Validation

### Successful Shutdown Checklist

- [ ] PreStop delay observed in pod events
- [ ] SIGTERM logged in container logs
- [ ] NATS flush messages present (sequencer)
- [ ] Exit code 0
- [ ] No orphaned connections (check with `ss -tunap`)
- [ ] NATS data directory intact
- [ ] Restart recovers all messages

### Testing Shutdown

```bash
# 1. Start sequencer pod
kubectl apply -f sequencer.yaml
kubectl wait --for=condition=Ready pod/cdk-erigon-sequencer-0

# 2. Trigger shutdown
kubectl delete pod cdk-erigon-sequencer-0

# 3. Monitor logs in real-time
kubectl logs -f cdk-erigon-sequencer-0

# 4. Check exit code
kubectl get pod cdk-erigon-sequencer-0 -o jsonpath='{.status.containerStatuses[0].state.terminated.exitCode}'

# 5. Verify NATS data persisted
kubectl exec cdk-erigon-sequencer-0 -- ls -lh /data/nats
```

## Troubleshooting

### Shutdown Takes Full 60s

**Symptom:** Pod always terminates at exactly 60s

**Likely Causes:**
1. NATS flush hanging
2. Database connection not closing
3. Process not handling SIGTERM

**Debug Steps:**
```bash
# Check if SIGTERM was received
kubectl logs <pod> | grep SIGTERM

# Check for flush errors
kubectl logs <pod> | grep -i "flush\|error"

# Verify process handling signals
kubectl exec <pod> -- kill -TERM 1
# Should see graceful shutdown logs
```

### Messages Lost After Restart

**Symptom:** NATS consumer skips messages after restart

**Likely Causes:**
1. SIGKILL before flush completed
2. Filesystem corruption
3. Consumer state not persisted

**Debug Steps:**
```bash
# Check JetStream stream state
kubectl exec <sequencer-pod> -- nats stream info datastream

# Verify sequence numbers
kubectl logs <rpc-pod> | grep "sequence"

# Check for gaps
kubectl exec <rpc-pod> -- nats consumer info datastream rpc-consumer
```

### Pod Restarts During Shutdown

**Symptom:** Pod restarts instead of clean termination

**Likely Causes:**
1. Liveness probe killing pod
2. OOMKilled during flush
3. terminationGracePeriodSeconds too short

**Solutions:**
```yaml
# Disable liveness probe during shutdown (not recommended)
# OR increase grace period
terminationGracePeriodSeconds: 90

# OR increase memory limits
resources:
  limits:
    memory: 4Gi
```

## Recovery After Unclean Shutdown

### Sequencer Recovery

1. Check NATS stream integrity:
```bash
kubectl exec <pod> -- nats stream info datastream
```

2. Verify last published sequence:
```bash
kubectl logs <pod> | grep "last sequence"
```

3. Repair if needed:
```bash
# JetStream auto-repairs on restart
# No manual intervention usually needed
```

### RPC Node Recovery

1. Check consumer state:
```bash
kubectl exec <pod> -- nats consumer info datastream rpc-consumer
```

2. Consumer will resume from last ack
3. No data loss if sequencer flush succeeded

## Best Practices

1. **Always set terminationGracePeriodSeconds > expected shutdown time**
   - Sequencer: 60s minimum
   - RPC: 30s minimum

2. **Monitor shutdown duration**
   - Alert if shutdown > 50% grace period
   - Investigate if consistently slow

3. **Test shutdown under load**
   - During active block production
   - With multiple RPC consumers
   - With large pending message queue

4. **Validate after every restart**
   - Check exit code
   - Verify no gaps in NATS sequences
   - Confirm all consumers reconnected

## References

- [PreStop Hook Specification](prestop-hook.md)
- [Health Probes Configuration](health-probes.md)
- [Kubernetes Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [NATS JetStream Persistence](https://docs.nats.io/nats-concepts/jetstream/streams)
