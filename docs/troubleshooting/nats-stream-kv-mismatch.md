# NATS Stream/KV Mismatch Troubleshooting

## Symptom

Sequencer crashes on startup with error:

```
Resuming: using KV metadata as source of truth kvCount=X streamCount=Y nextEntry=X
failed to get entry N from NATS: nats: API error: code=404 err_code=10037 description=message not found
```

Or the inverse:

```
Failed to update total entries in metadata store after publish error="failed to store total entries in metadata: context deadline exceeded"
```

## Root Cause

The NATS datastream uses two data structures that must stay in sync:

| Component | Purpose | Location |
|-----------|---------|----------|
| **JetStream** | Stores actual datastream messages | `DATASTREAM` stream |
| **KV Store** | Tracks metadata (entry count, bookmarks) | `METADATA` KV bucket |

When the sequencer crashes mid-operation (OOM, timeout, etc.), these can become inconsistent:

- **KV > Stream**: KV recorded entries that weren't fully published
- **Stream > KV**: Messages published but KV update failed/timed out

## Common Causes

1. **Memory pressure** - Container OOMKilled during write operations
2. **KV write timeout** - `context deadline exceeded` when updating metadata
3. **Unclean shutdown** - SIGKILL before graceful shutdown completes
4. **Disk I/O latency** - Slow PVC causing JetStream operations to timeout

## Diagnosis

Check the sequencer logs for:

```bash
kubectl logs cdk-erigon-sequencer-0 | grep -E "kvCount|streamCount|mismatch|deadline exceeded"
```

Look for:
- `kvCount=X streamCount=Y` where X ≠ Y
- `context deadline exceeded` errors
- `message not found` errors

Check pod termination reason:

```bash
kubectl get pod cdk-erigon-sequencer-0 -o jsonpath='{.status.containerStatuses[0].lastState.terminated.reason}'
```

## Resolution

### Option 1: Reset NATS State (Recommended for Dev/Test)

Delete the NATS data to force a full resync from the upstream datastream:

```bash
# Delete the sequencer pod and its PVC
kubectl delete pvc nats-data-cdk-erigon-sequencer-0
kubectl delete pod cdk-erigon-sequencer-0

# The StatefulSet will recreate both
```

Or exec into the pod and clear NATS data:

```bash
kubectl exec -it cdk-erigon-sequencer-0 -- rm -rf /data/nats-data/*
kubectl delete pod cdk-erigon-sequencer-0
```

### Option 2: Reset KV Only

If you want to preserve stream messages but reset the metadata:

```bash
# Port-forward to NATS
kubectl port-forward cdk-erigon-sequencer-0 4222:4222 &

# Use nats CLI to delete the KV bucket
nats kv rm METADATA --force

# Restart the pod
kubectl delete pod cdk-erigon-sequencer-0
```

### Option 3: Increase Resources (Prevention)

If crashes are caused by memory pressure, increase container limits:

```yaml
# values-dev.yaml or values-local.yaml
sequencer:
  resources:
    limits:
      memory: 4Gi  # Increase from 2Gi
    requests:
      memory: 2Gi
```

Then upgrade:

```bash
helm upgrade cdk-erigon k8s/helm -f k8s/helm/values-bali.yaml -f k8s/helm/values-dev.yaml -f k8s/helm/values-local.yaml
```

## Prevention

### 1. Adequate Memory

The sequencer needs memory for:
- Erigon execution (~1-1.5GB during sync)
- NATS JetStream (configured for 1GB max)
- SMT (Sparse Merkle Tree) computation (spikes during IntermediateHashes stage)

**Recommended minimum: 4GB** for active sync, 2GB once synced.

### 2. Monitor Memory Usage

Use the Grafana dashboard "cdk-erigon Sync & Metrics" to monitor:
- `go_memstats_heap_inuse_bytes` - Heap memory
- `process_resident_memory_bytes` - RSS
- Container memory limits vs actual usage

### 3. Graceful Shutdown

Ensure pods have adequate termination grace period:

```yaml
sequencer:
  terminationGracePeriodSeconds: 60
```

## Technical Details

### How Sync Works

1. **stream-catchup** stage reads executed blocks from mdbx database
2. Batches ~80,000 messages for publishing to NATS
3. For each message:
   - Publish to JetStream (`DATASTREAM` stream)
   - Update entry count in KV (`METADATA` bucket)
4. If any step fails mid-batch, inconsistency occurs

### Timeout Configuration

The current timeout is 10 seconds for the entire batch with a nested 5-second timeout per KV operation:

```go
// stream_server.go line 244
txCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

// metadata_manager.go line 238
timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
```

Under memory pressure or slow I/O, these timeouts may be insufficient.

### Non-Atomic Operations

The publish loop in `CommitAtomicOp` is not truly atomic:

1. `PublishMsg` to JetStream
2. Increment `nextEntry`  
3. `SetTotalEntries` to KV

If step 3 fails after step 1 succeeds, the stream has messages that KV doesn't know about.

Conversely, if the KV Put succeeds on the server but the response times out, the client retries but gets an error - resulting in KV being ahead of stream.

These edge cases require manual intervention (see Resolution section above). Automatic recovery is planned - see Linear issue RD-614.

### Related Files

- `zk/datastream/natsstream/stream_server.go` - Stream publishing logic
- `zk/datastream/natsstream/metadata_manager.go` - KV metadata operations
- `zk/datastream/natsstream/manager.go` - NATS server lifecycle

## See Also

- [NATS JetStream Documentation](https://docs.nats.io/nats-concepts/jetstream)
- [cdk-erigon Helm Chart](../../k8s/helm/README.md)
