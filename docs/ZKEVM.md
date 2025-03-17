# zkEVM Components and Implementation

This document provides a detailed explanation of the zkEVM-specific components in CDK-Erigon, their implementation, and how they interact with each other and the rest of the system.

## Overview

The zkEVM components extend the standard Erigon Ethereum client to support a Layer 2 scaling solution based on zero-knowledge proofs. This enables high-throughput transaction processing with Ethereum-level security guarantees.

## Architecture

```
+-----------------------------------------------------------+
|                       zkEVM Module                        |
+-----------------------------------------------------------+
         |               |                |
         v               v                v
+----------------+ +------------+ +----------------+
| Etherman       | | State      | | Executor       |
| (L1 Interface) | | Management | | Interface      |
+----------------+ +------------+ +----------------+
         |               |                |
         v               v                v
+----------------+ +------------+ +----------------+
| Contract       | | SMT        | | Proof          |
| Interaction    | | (Merkle    | | Verification   |
|                | | Trees)     | |                |
+----------------+ +------------+ +----------------+
```

## Core Components

### 1. Etherman (L1 Interface)

The Etherman component is responsible for interacting with the Ethereum Layer 1 chain. It handles:

- Reading contract events from Rollup contracts
- Monitoring Global Exit Root updates
- Submitting batches and proofs to L1
- Handling L1 reorgs and their impact on L2

**Key Files:**
- [`zk/syncer/l1_syncer.go`](../zk/syncer/l1_syncer.go): Main implementation for L1 synchronization
- [`zk/contracts/rollup.go`](../zk/contracts/rollup.go): Rollup contract interaction
- [`zk/contracts/eth_txs.go`](../zk/contracts/eth_txs.go): Ethereum transaction handling

**Key Interfaces:**
```go
// Interface is defined in zk/syncer/l1_syncer.go
type IEtherman interface {
    HeaderByNumber(ctx context.Context, blockNumber *big.Int) (*ethTypes.Header, error)
    BlockByNumber(ctx context.Context, blockNumber *big.Int) (*ethTypes.Block, error)
    FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]ethTypes.Log, error)
    // additional methods...
}
```

**Configuration Options:**
- `zkevm.l1-rpc-url`: URL for the L1 Ethereum node - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.l1-chain-id`: Chain ID of the L1 network - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.address-rollup`: Address of the Rollup contract on L1 - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.address-zkevm`: Address of the zkEVM contract on L1 - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.address-ger-manager`: Address of the Global Exit Root Manager - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)

### 2. State Management

The State Management component maintains the current state of the zkEVM chain. It:

- Tracks account states and storage
- Manages the Sparse Merkle Tree (SMT) for state verification
- Processes state transitions from executed transactions
- Generates state roots for proof verification

**Key Files:**
- [`zk/state/state.go`](../zk/state/state.go): State management implementation
- [`zk/smt/smt.go`](../zk/smt/smt.go): Sparse Merkle Tree implementation
- [`zk/stages/stage_zk_interhashes.go`](../zk/stages/stage_zk_interhashes.go): State root calculation

**Key Data Structures:**
```go
// State is defined in various state management files
type State struct {
    // Core state tracking
    batch  uint64
    block  uint64
    
    // Trees for state verification
    accountsTree  *smt.SparseMerkleTree
    storageTree   *smt.SparseMerkleTree
    
    // Caches
    accountCache  map[common.Address]*Account
    storageCache  map[common.Address]map[common.Hash]common.Hash
}
```

**State Root Calculation:**
The state root is calculated as a Merkle root of all accounts and their storage, which allows for compact proofs of specific state elements. Implementation in [`zk/stages/stage_zk_interhashes.go`](../zk/stages/stage_zk_interhashes.go).

### 3. Executor Interface

The Executor Interface interacts with the zkEVM execution engine, which can be either:
- An integrated executor running in the same process
- An external executor service accessed via RPC

It handles:
- Executing transaction batches
- Computing state transitions
- Generating zero-knowledge proofs (or interfacing with a prover)
- Validating execution results

**Key Files:**
- [`zk/executor/executor.go`](../zk/executor/executor.go): Main executor interface
- [`zk/legacy_executor_verifier/executor_client.go`](../zk/legacy_executor_verifier/executor_client.go): Client for external executor
- [`zk/stages/stage_sequence_execute.go`](../zk/stages/stage_sequence_execute.go): Transaction execution via executor

**Executor Protocol:**
```
+---------------+                 +---------------+
| CDK-Erigon    |  1. Tx Batch    | Executor      |
| Node          | --------------> | Service       |
|               |                 |               |
|               |  2. Execution   |               |
|               | <-------------- |               |
|               |     Result      |               |
+---------------+                 +---------------+
```

**Configuration Options:**
- `zkevm.executor-urls`: URLs for external executors - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.executor-strict`: Enforce strict executor mode - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.executor-request-timeout`: Timeout for executor requests - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)

### 4. Batch Processing

The Batch Processing component manages transaction batches that are processed by the zkEVM:

- Groups transactions into batches
- Schedules batches for execution
- Manages batch sequencing
- Handles batch verification

**Key Files:**
- [`zk/stages/stage_batches.go`](../zk/stages/stage_batches.go): Batch processing implementation
- [`zk/sequencer/sequencer.go`](../zk/sequencer/sequencer.go): Sequencer implementation (for sequencer nodes)
- [`zk/stages/stage_sequence_execute.go`](../zk/stages/stage_sequence_execute.go): Batch execution

**Batch Structure:**
```go
// Defined in zk/types/batch.go
type Batch struct {
    BatchNumber    uint64
    GlobalExitRoot common.Hash
    Timestamp      uint64
    Transactions   [][]byte
    Coinbase       common.Address
    StateRoot      common.Hash
}
```

**Sequencing Parameters:**
- `zkevm.sequencer-batch-seal-time`: Time interval for sealing batches - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.sequencer-empty-batch-seal-time`: Time interval for sealing empty batches - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)

### 5. Datastreamer Integration

The Datastreamer component provides efficient transfer of blockchain data:

- Streams L2 blocks and transactions
- Optimizes network usage with delta compression
- Provides faster synchronization for new nodes

**Key Files:**
- [`zk/datastream/client/client.go`](../zk/datastream/client/client.go): Client implementation
- [`zk/datastream/server/server.go`](../zk/datastream/server/server.go): Server implementation

**Configuration Options:**
- `zkevm.l2-datastreamer-url`: URL for the datastreamer service - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.l2-datastreamer-timeout`: Connection timeout - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)

## Synchronization Process

The zkEVM synchronization follows these steps:

1. **L1 Synchronization**:
   - Fetch L1 blocks and filter for relevant events - [`zk/stages/stage_l1_syncer.go`](../zk/stages/stage_l1_syncer.go)
   - Identify batch submissions and verifications - [`zk/stages/stage_l1_syncer.go`](../zk/stages/stage_l1_syncer.go)
   - Extract Global Exit Roots - [`zk/syncer/l1_syncer.go`](../zk/syncer/l1_syncer.go)

2. **Batch Retrieval**:
   - Get batch data either from L1 or datastreamer - [`zk/stages/stage_batches.go`](../zk/stages/stage_batches.go)
   - Validate batch consistency - [`zk/stages/stage_batches.go`](../zk/stages/stage_batches.go)

3. **State Synchronization**:
   - Execute transactions in each batch - [`eth/stagedsync/stage_execute.go`](../eth/stagedsync/stage_execute.go)
   - Update local state - [`zk/stages/stage_zk_interhashes.go`](../zk/stages/stage_zk_interhashes.go)
   - Verify state transitions with proofs (if available) - [`zk/stages/stage_witness.go`](../zk/stages/stage_witness.go)

4. **Finality Tracking**:
   - Track verified batches on L1 - [`zk/stages/stage_l1_syncer.go`](../zk/stages/stage_l1_syncer.go)
   - Handle L1 reorgs and their impact on L2 - [`zk/stages/stage_l1_syncer.go`](../zk/stages/stage_l1_syncer.go)

## Verification Flow

```
+----------------+       +----------------+       +----------------+
| Transaction    |       | ZK Proof       |       | State Root     |
| Execution      | ----> | Generation     | ----> | Verification   |
+----------------+       +----------------+       +----------------+
```

1. **Transaction Execution**:
   - Transactions are executed in deterministic order - [`zk/stages/stage_sequence_execute.go`](../zk/stages/stage_sequence_execute.go)
   - State transitions are recorded - [`eth/stagedsync/stage_execute.go`](../eth/stagedsync/stage_execute.go)

2. **ZK Proof Generation**:
   - A zero-knowledge proof is generated (externally or internally) - [`zk/prover/prover.go`](../zk/prover/prover.go)
   - The proof attests to the correctness of state transitions - [`zk/prover/prover.go`](../zk/prover/prover.go)

3. **State Root Verification**:
   - The new state root is computed - [`zk/stages/stage_zk_interhashes.go`](../zk/stages/stage_zk_interhashes.go)
   - The state root and proof are submitted to L1 - [`zk/sequencer/sequencer.go`](../zk/sequencer/sequencer.go)
   - The L1 contract verifies the proof - [`zk/contracts/rollup.go`](../zk/contracts/rollup.go)

## JSON-RPC API Extensions

CDK-Erigon extends the standard Ethereum JSON-RPC API with zkEVM-specific methods:

- `zkevm_batchNumber`: Returns the latest batch number
- `zkevm_batchByNumber`: Returns batch information by number
- `zkevm_getBatchByBlockNumber`: Returns the batch containing a specific block
- `zkevm_getGlobalExitRoot`: Returns the current Global Exit Root
- `zkevm_isBlockVirtualized`: Checks if a block has been virtualized
- `zkevm_verifiedBatchNumber`: Returns the latest verified batch number

**API Implementation:**
- [`turbo/jsonrpc/zkevm_api.go`](../turbo/jsonrpc/zkevm_api.go): Implementation of zkEVM-specific API methods

## zkEVM Specific Configuration

The zkEVM module is configured through various parameters, which can be set in the config file or via command-line arguments:

### Network Configuration
- `zkevm.l2-chain-id`: Chain ID for the L2 network - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.l2-rpc-url`: URL for connecting to L2 (for non-sequencers) - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.l1-chain-id`: Chain ID for the L1 network - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.l1-rpc-url`: URL for the L1 Ethereum node - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)

### Contract Addresses
- `zkevm.address-sequencer`: Address of the sequencer - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.address-zkevm`: Address of the zkEVM contract on L1 - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.address-rollup`: Address of the Rollup contract on L1 - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.address-ger-manager`: Address of the Global Exit Root Manager - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)

### Gas and Transaction Parameters
- `zkevm.default-gas-price`: Default gas price for L2 transactions - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.max-gas-price`: Maximum allowed gas price - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)
- `zkevm.gas-price-factor`: Factor for calculating gas prices - [`eth/ethconfig/config.go`](../eth/ethconfig/config.go)

### Sequencer Configuration
- `zkevm.sequencer-batch-seal-time`: Time interval for sealing batches
- `zkevm.sequencer-empty-batch-seal-time`: Time interval for sealing empty batches

### Debug and Performance Options
- `zkevm.debug-no-sync`: Disable synchronization (for debugging)
- `zkevm.disable-virtual-counters`: Disable virtual counters for performance
- `zkevm.l1-contract-address-check`: Check L1 contract addresses
- `zkevm.debug-disable-state-root-check`: Disable state root verification

## Performance Tuning

The zkEVM module includes several parameters that can be tuned for optimal performance:

- **Memory Usage**:
  - `zkevm.witness-memdb-size`: Memory allocated for witness generation
  - `zkevm.sequencer-decoded-tx-cache-size`: Size of decoded transaction cache

- **Concurrency**:
  - `zkevm.executor-max-concurrent-requests`: Maximum concurrent executor requests
  - `zkevm.rpc-get-batch-witness-concurrency-limit`: Concurrency limit for batch witness requests

- **Timeouts**:
  - `zkevm.executor-request-timeout`: Timeout for executor requests
  - `zkevm.l2-datastreamer-timeout`: Timeout for datastreamer connections

## Advanced Features

### Resequencing

The resequencing feature allows for reorganizing transaction batches:

- **Configuration**:
  - `zkevm.sequencer-resequence`: Enable resequencing
  - `zkevm.sequencer-resequence-strict`: Enforce strict resequencing

- **Mechanism**:
  - Transactions are reordered to optimize execution
  - State is updated accordingly
  - New batches are created with the reordered transactions

### Witness Generation

Witness generation is a critical part of the zero-knowledge proof system:

- Witnesses contain the execution trace and state accesses
- They are used to generate the zero-knowledge proofs
- Can be cached for performance optimization

- **Configuration**:
  - `zkevm.witness-cache-enabled`: Enable witness caching
  - `zkevm.witness-cache-purge`: Purge witness cache on startup

## Error Handling and Recovery

The zkEVM module implements robust error handling and recovery mechanisms:

- **Unwind Operations**:
  - `zkevm.witness-unwind-limit`: Maximum batches to unwind
  - Allows rolling back to a previous valid state

- **Bad Batch Handling**:
  - `zkevm.bad-batches`: List of known bad batches to skip
  - `zkevm.ignore-bad-batches-check`: Ignore bad batch checks 