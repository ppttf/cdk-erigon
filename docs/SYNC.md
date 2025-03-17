# Synchronization in CDK-Erigon

This document explains the synchronization process in CDK-Erigon, focusing on how the node synchronizes with both Layer 1 (Ethereum) and Layer 2 (zkEVM) chains.

## Overview

CDK-Erigon implements a complex synchronization system that maintains consistency between L1 and L2 chains. The synchronization process ensures that:

1. Layer 2 state remains consistent with the data published on Layer 1
2. All transactions are processed in the correct order
3. The node can verify the correctness of the state using zero-knowledge proofs

```
+--------------------+        +--------------------+
|                    |        |                    |
|  Layer 1           |        |  Layer 2           |
|  (Ethereum)        | <----> |  (zkEVM)           |
|                    |        |                    |
+--------------------+        +--------------------+
          ^                            ^
          |                            |
          |                            |
          v                            v
+--------------------+        +--------------------+
|                    |        |                    |
|  L1 Sync Process   |        |  L2 Sync Process   |
|                    |        |                    |
+--------------------+        +--------------------+
          ^                            ^
          |                            |
          |                            |
          +------------+---------------+
                       |
                       v
               +----------------+
               |                |
               |  CDK-Erigon    |
               |  Node          |
               |                |
               +----------------+
```

## Synchronization Components

### 1. L1 Synchronization

The L1 synchronization process fetches data from the Ethereum chain to track the zkEVM rollup's state.

#### Key Components

- **L1 Block Fetcher**: Retrieves blocks from the Ethereum chain - [`zk/syncer/l1_syncer.go`](../zk/syncer/l1_syncer.go)
- **Event Filter**: Filters for relevant rollup events - [`zk/syncer/l1_syncer.go:queryBlocks()`](../zk/syncer/l1_syncer.go)
- **Contract Monitor**: Monitors zkEVM contracts for state updates - [`zk/stages/stage_l1_syncer.go`](../zk/stages/stage_l1_syncer.go)

#### L1 Events Tracked

- **Batch Submissions**: New L2 batches submitted by the sequencer - [`zk/stages/stage_l1_syncer.go:parseLogType()`](../zk/stages/stage_l1_syncer.go)
- **Batch Verifications**: Batches that have been verified with ZK proofs - [`zk/stages/stage_l1_syncer.go:parseLogType()`](../zk/stages/stage_l1_syncer.go)
- **Global Exit Root Updates**: Changes to the Global Exit Root - [`zk/stages/stage_l1_info_tree.go`](../zk/stages/stage_l1_info_tree.go)
- **Contract Address Updates**: Changes to contract addresses - [`zk/contracts/rollup.go`](../zk/contracts/rollup.go)

### 2. L2 Synchronization

The L2 synchronization process maintains the zkEVM chain state based on data from L1 and directly from L2 sources like the datastreamer.

#### Key Components

- **Batch Processor**: Processes L2 transaction batches - [`zk/stages/stage_batches.go`](../zk/stages/stage_batches.go)
- **State Manager**: Updates the L2 state based on processed batches - [`zk/stages/stage_zk_interhashes.go`](../zk/stages/stage_zk_interhashes.go)
- **Proof Verifier**: Verifies zero-knowledge proofs (if available) - [`zk/stages/stage_witness.go`](../zk/stages/stage_witness.go)
- **Datastreamer Client**: Fetches data from datastreamer service - [`zk/datastream/client/client.go`](../zk/datastream/client/client.go)

## Synchronization Flow

### Initial Synchronization

```
+-------------------+     +-------------------+     +-------------------+
| 1. Fetch L1 Blocks| --> | 2. Extract Rollup | --> | 3. Retrieve       |
| & Rollup Events   |     | Batch Information |     | Batch Data        |
+-------------------+     +-------------------+     +-------------------+
           |                                                  |
           v                                                  v
+-------------------+     +-------------------+     +-------------------+
| 6. Update Local   | <-- | 5. Apply State    | <-- | 4. Process        |
| State             |     | Transitions       |     | Transactions      |
+-------------------+     +-------------------+     +-------------------+
```

1. **Fetch L1 Blocks & Rollup Events**:
   - Connect to the L1 node specified in `zkevm.l1-rpc-url`
   - Fetch blocks starting from `zkevm.l1-first-block`
   - Filter for events from rollup contracts
   - Implementation: [`zk/syncer/l1_syncer.go:RunQueryBlocks()`](../zk/syncer/l1_syncer.go)

2. **Extract Rollup Batch Information**:
   - Parse events to identify batch submissions and verifications
   - Build a chronological sequence of batches
   - Implementation: [`zk/stages/stage_l1_syncer.go:SpawnStageL1Syncer()`](../zk/stages/stage_l1_syncer.go)

3. **Retrieve Batch Data**:
   - Get transaction data for each batch
   - Sources can be L1 contracts, datastreamer, or RPC endpoints
   - Implementation: [`zk/stages/stage_batches.go:SpawnBatchesStage()`](../zk/stages/stage_batches.go)

4. **Process Transactions**:
   - Execute transactions in each batch
   - Compute state transitions
   - Implementation: [`eth/stagedsync/stage_execute.go:SpawnExecuteBlocksStage()`](../eth/stagedsync/stage_execute.go)

5. **Apply State Transitions**:
   - Update the node's local state
   - Calculate and verify state roots
   - Implementation: [`zk/stages/stage_zk_interhashes.go:SpawnIntermediateHashesStage()`](../zk/stages/stage_zk_interhashes.go)

6. **Update Local State**:
   - Commit the new state to the database
   - Update chain metadata
   - Implementation: [`eth/stagedsync/stage_finish.go:FinishForward()`](../eth/stagedsync/stage_finish.go)

### Ongoing Synchronization

Once the initial synchronization is complete, the node continues to monitor L1 for new events and updates its state accordingly.

```
        +---------------------+
        |                     |
        |  Monitor L1 Chain   |<---------+
        |                     |          |
        +----------+----------+          |
                   |                     |
                   v                     |
        +----------+----------+          |
        |                     |          |
        |  New Batch Event?   |---No---->+
        |                     |          
        +----------+----------+          
                   |                     
                   | Yes                 
                   v                     
        +----------+----------+          
        |                     |          
        |  Process New Batch  |          
        |                     |          
        +----------+----------+          
                   |                     
                   v                     
        +----------+----------+          
        |                     |          
        |  Update Local State |          
        |                     |          
        +---------------------+          
```

The ongoing synchronization loop is implemented in [`eth/stagedsync/sync.go:Run()`](../eth/stagedsync/sync.go).

## Synchronization Modes

CDK-Erigon supports different synchronization modes:

### Standard Sync

The default synchronization mode:
- Fetches all L1 data
- Executes all transactions
- Verifies state transitions
- Configuration: Default mode, no special flags
- Implementation: [`zk/stages/stages.go:DefaultZkStages()`](../zk/stages/stages.go)

### Fast Sync

A faster synchronization mode that uses trusted sources:
- Fetches state directly from datastreamer
- Skips certain verifications for speed
- Configuration: `zkevm.l2-datastreamer-url`
- Implementation: [`zk/datastream/client/client.go`](../zk/datastream/client/client.go)

### Trusted Sync

For certain configurations that trust the sequencer:
- Uses the sequencer RPC for state
- Performs minimal verification
- Configuration: `zkevm.l2-sequencer-rpc-url`
- Implementation: [`zk/syncer/sequencer_rpc.go`](../zk/syncer/sequencer_rpc.go)

## Handling Reorgs

Blockchain reorganizations (reorgs) can occur on L1, which affect L2 state. CDK-Erigon handles reorgs as follows:

1. **Detect L1 Reorg**:
   - Monitor L1 chain for reorgs
   - Identify affected L2 batches
   - Implementation: [`zk/stages/stage_l1_syncer.go:detectedReorg()`](../zk/stages/stage_l1_syncer.go)

2. **Revert Affected Batches**:
   - Roll back state to before the reorg
   - Configuration: `zkevm.witness-unwind-limit`
   - Implementation: [`eth/stagedsync/sync.go:Unwind()`](../eth/stagedsync/sync.go)

3. **Reapply New Chain**:
   - Process batches from the new canonical chain
   - Update state accordingly
   - Implementation: [`eth/stagedsync/sync.go:Run()`](../eth/stagedsync/sync.go)

## Datastreamer Integration

The datastreamer component provides efficient data transfer for synchronization:

```
+------------------+        +------------------+        +------------------+
|                  |        |                  |        |                  |
|  Sequencer Node  |------->|  Datastreamer    |------->|  Full Node       |
|                  |        |  Service         |        |                  |
+------------------+        +------------------+        +------------------+
```

- **Efficient Data Transfer**: Uses optimized protocols for block and transaction data
- **Delta Compression**: Only transfers changes rather than full state
- **Multiplexing**: Supports multiple nodes from a single source
- **Configuration**: Set via `zkevm.l2-datastreamer-url`

Implementation:
- Client: [`zk/datastream/client/client.go`](../zk/datastream/client/client.go)
- Server: [`zk/datastream/server/server.go`](../zk/datastream/server/server.go)
- Integration: [`zk/stages/stage_datastream_catchup.go`](../zk/stages/stage_datastream_catchup.go)

## Configuration Parameters

### L1 Synchronization

| Parameter | Description | Code Reference |
|-----------|-------------|----------------|
| `zkevm.l1-rpc-url` | URL of the L1 Ethereum node | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l1-first-block` | First L1 block to scan for zkEVM events | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l1-chain-id` | Chain ID of the L1 network | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l1-rollup-id` | Rollup ID on the L1 network | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l1-contract-address-check` | Whether to check L1 contract addresses | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l1-block-range` | Number of blocks to process in a single batch | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l1-query-delay` | Delay between L1 queries (in milliseconds) | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |

### L2 Synchronization

| Parameter | Description | Code Reference |
|-----------|-------------|----------------|
| `zkevm.l2-sequencer-rpc-url` | URL of the L2 sequencer RPC endpoint | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l2-datastreamer-url` | URL of the datastreamer service | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l2-datastreamer-timeout` | Timeout for datastreamer connections | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l2-chain-id` | Chain ID for the L2 network | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.l2-short-circuit-to-verified-batch` | Whether to short-circuit to verified batches | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |

### Advanced Parameters

| Parameter | Description | Code Reference |
|-----------|-------------|----------------|
| `zkevm.sync-limit` | Maximum number of batches to sync | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.sync-limit-verified-enabled` | Whether to enable sync limit for verified batches | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.sync-limit-unverified-count` | Maximum number of unverified batches to sync | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.debug-no-sync` | Disable synchronization (for debugging) | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |
| `zkevm.witness-unwind-limit` | Maximum batches to unwind | [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) |

## Troubleshooting Synchronization Issues

### Common Sync Problems

#### 1. L1 Connection Issues

**Symptoms**:
- Logs showing L1 connection errors
- Sync progress halted at L1 block fetching

**Solutions**:
- Check `zkevm.l1-rpc-url` configuration
- Verify L1 node is operational
- Ensure network connectivity
- Relevant code: [`zk/syncer/l1_syncer.go:RunQueryBlocks()`](../zk/syncer/l1_syncer.go)

#### 2. Batch Processing Errors

**Symptoms**:
- Error logs during batch processing
- Sync progress halted at specific batch

**Solutions**:
- Check batch data integrity
- Verify datastreamer connectivity if enabled
- Check execution logs for specific errors
- Relevant code: [`zk/stages/stage_batches.go:SpawnBatchesStage()`](../zk/stages/stage_batches.go)

#### 3. State Root Mismatches

**Symptoms**:
- Error logs showing state root mismatch
- Sync halted with verification errors

**Solutions**:
- Check for potential L1 reorgs
- Verify executor service is functioning correctly
- Consider resetting to a trusted checkpoint
- Relevant code: [`zk/stages/stage_zk_interhashes.go`](../zk/stages/stage_zk_interhashes.go)

## Monitoring Synchronization Progress

The node's synchronization progress can be monitored through:

1. **Log Output**: Progress indicators in log files
   - Implementation: [`zk/stages/stage_l1_syncer.go:SpawnStageL1Syncer()`](../zk/stages/stage_l1_syncer.go)

2. **Metrics**: Prometheus metrics for detailed monitoring
   - Implementation: [`zk/metrics/metrics.go`](../zk/metrics/metrics.go)

3. **RPC Methods**: JSON-RPC methods to query sync status
   - Implementation: [`turbo/jsonrpc/zkevm_api.go`](../turbo/jsonrpc/zkevm_api.go)

## Performance Tuning

Synchronization performance can be optimized through:

1. **Hardware Resources**: 
   - Recommended: 16+ CPU cores, 32GB+ RAM, SSD storage
   - Implementation: Resource checks in [`eth/backend.go`](../eth/backend.go)

2. **Network Bandwidth**:
   - Higher bandwidth improves L1 and datastreamer sync
   - Implementation: Connection handling in [`zk/syncer/l1_syncer.go`](../zk/syncer/l1_syncer.go)

3. **Configuration Optimization**:
   - Adjust `zkevm.l1-block-range` for batch size
   - Tune `zkevm.l1-query-delay` for rate limiting
   - Implementation: Usage in [`zk/syncer/l1_syncer.go`](../zk/syncer/l1_syncer.go) 