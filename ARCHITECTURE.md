# CDK-Erigon Architecture Overview

## Introduction

CDK-Erigon is a zkEVM implementation, built on top of the Erigon Ethereum client. It extends Erigon's capabilities to support zkEVM-specific functionality, which enables Layer 2 scaling solutions with zero-knowledge proofs.

This document provides a high-level overview of the CDK-Erigon architecture, its main components, and how they interact with each other.

## System Architecture

```
+-------------------------------+
|                               |
|        CDK-Erigon Node        |
|                               |
+-------------------------------+
           |       |
           |       |
           v       v
+---------------+  +---------------+
|  Layer 1      |  | Layer 2       |
|  (Ethereum)   |  | (zkEVM Chain) |
+---------------+  +---------------+
           |       |
           |       |
           v       v
+---------------+  +---------------+
| External RPCs |  | Datastreamer  |
| and Services  |  | Service       |
+---------------+  +---------------+
```

CDK-Erigon interacts with both Layer 1 (Ethereum) and Layer 2 (zkEVM) networks. It communicates with external RPCs for Ethereum data and may use a Datastreamer service for efficient data retrieval.

## Main Components

```
+-------------------------------------------------------------+
|                                                             |
|                      CDK-Erigon                             |
|                                                             |
| +-------------------+  +------------------+  +-----------+  |
| |                   |  |                  |  |           |  |
| | Core Erigon       |  | zkEVM Extension  |  | RPC API   |  |
| | Components        |  | Components       |  | Server    |  |
| |                   |  |                  |  |           |  |
| +-------------------+  +------------------+  +-----------+  |
|          |                      |                  |        |
|          v                      v                  v        |
| +-------------------+  +------------------+  +-----------+  |
| |                   |  |                  |  |           |  |
| | Storage Layer     |  | L1/L2 Sync       |  | HTTP/WS   |  |
| | (Database)        |  | Mechanisms       |  | Interface |  |
| |                   |  |                  |  |           |  |
| +-------------------+  +------------------+  +-----------+  |
|                                                             |
+-------------------------------------------------------------+
```

### 1. Core Erigon Components

These are the foundational components inherited from the Erigon Ethereum client:
- **Node Management**: Handles peer connections, network protocols
- **Chain Management**: Processes blocks and transactions
- **State Management**: Maintains the Ethereum state

### 2. zkEVM Extension Components

These components extend Erigon to support zkEVM functionality:
- **L1 Contract Interface**: Interacts with zkEVM contracts on Ethereum
- **ZK Verifier**: Verifies zero-knowledge proofs
- **Batch Processor**: Handles batches of L2 transactions

### 3. RPC API Server

Provides interfaces for external interaction:
- **Ethereum APIs**: Standard Ethereum JSON-RPC endpoints
- **zkEVM APIs**: Additional endpoints specific to zkEVM functionality

### 4. Storage Layer

Manages data persistence:
- **Chain Data**: Blocks, transactions, receipts
- **State Data**: Account states, storage, code
- **ZK-specific Data**: Proof data, batch information

### 5. L1/L2 Sync Mechanisms

Handles synchronization between layers:
- **L1 Sync**: Fetches data from Ethereum
- **L2 Sync**: Maintains the zkEVM chain state
- **Batch Sync**: Synchronizes transaction batches

### 6. HTTP/WS Interface

External communication channels:
- **HTTP Server**: REST-like JSON-RPC interface
- **WebSocket Server**: For subscription-based services

## Execution Flow

```
+---------------+     +---------------+     +----------------+
| 1. Start Up & |     | 2. Initialize |     | 3. Connect to  |
| Parse Config  | --> | Node & Chain  | --> | Networks       |
+---------------+     +---------------+     +----------------+
         |                                           |
         v                                           v
+----------------+     +---------------+     +----------------+
| 6. Process     |     | 5. Start RPC  |     | 4. Initialize  |
| Transactions   | <-- | Servers       | <-- | zkEVM          |
+----------------+     +---------------+     +----------------+
         |
         v
+----------------+     +---------------+
| 7. Sync with   |     | 8. Verify &   |
| L1 & L2        | --> | Process       |
+----------------+     | Batches       |
                       +---------------+
```

### Main Execution Sequence

1. **Startup & Configuration**: 
   - Parse command-line arguments
   - Load configuration file
   - Set up logging and debugging

2. **Node & Chain Initialization**:
   - Initialize the node with network settings
   - Set up the chain with genesis configuration
   - Prepare the database and state management

3. **Network Connections**:
   - Connect to L1 Ethereum network
   - Connect to L2 zkEVM network (if applicable)
   - Establish connections to datastreamer services

4. **zkEVM Initialization**:
   - Set up zkEVM-specific components
   - Initialize verification mechanisms
   - Prepare for batch processing

5. **Start RPC Servers**:
   - Initialize HTTP and WebSocket servers
   - Register API handlers
   - Start listening on configured ports

6. **Transaction Processing**:
   - Receive transactions from mempool or API
   - Validate and execute transactions
   - Update state accordingly

7. **Layer Synchronization**:
   - Sync with L1 Ethereum (contract events, blocks)
   - Maintain L2 state based on L1 data
   - Handle reorgs and chain reorganizations

8. **Batch Verification & Processing**:
   - Process transaction batches
   - Verify zero-knowledge proofs
   - Update state based on verified batches

## Configuration System

CDK-Erigon uses a flexible configuration system that supports:
- Command-line arguments
- YAML configuration files
- TOML configuration files
- Environment variables

Key configuration categories include:
- Network settings (chain ID, RPC URLs)
- zkEVM-specific settings (addresses, gas parameters)
- Node operation parameters (API, logging, debugging)
- Performance tuning options (memory limits, cache sizes)

## Directory Structure

```
cdk-erigon/
├── cmd/                 # Command-line tools and entry points
│   ├── cdk-erigon/      # Main application entry point
│   └── hack/            # Utility tools and scripts
├── eth/                 # Ethereum core implementation
├── node/                # Node management code
├── zkevm/               # zkEVM specific implementation
├── turbo/               # Performance-optimized components
├── common/              # Shared utilities and types
├── rlp/                 # RLP encoding/decoding
└── params/              # Chain parameters and constants
```

## For More Detailed Information

For deeper information about specific components, please refer to the following documentation:

- [Core Node Architecture](./docs/CORE_NODE.md) - Details on the base Erigon components
- [zkEVM Extension](./docs/ZKEVM.md) - Complete details on the zkEVM implementation
- [API Reference](./docs/API.md) - Comprehensive API documentation
- [Configuration Guide](./docs/CONFIGURATION.md) - Detailed configuration options
- [Synchronization](./docs/SYNC.md) - How the L1/L2 synchronization works
- [Developer Guide](./docs/DEVELOPMENT.md) - Guide for developers working on the codebase

## zkEVM Specific Components

### Contracts Interaction

CDK-Erigon interacts with several smart contracts deployed on L1:

```
+-----------------------------------+
|                                   |
|           Ethereum (L1)           |
|                                   |
+-----------------------------------+
              |     |     |
              |     |     |
              v     v     v
+-------------+ +-------+ +---------+
| Rollup      | | zkEVM | | Global  |
| Contract    | | Bridge | | Exit   |
+-------------+ +-------+ | Root    |
                          +---------+
```

- **Rollup Contract**: Manages state transitions and batches
- **zkEVM Bridge**: Handles asset transfers between L1 and L2
- **Global Exit Root**: Maintains proof of L2 state for withdrawals

### Verification Flow

```
+------------+    +--------------+    +------------+
| L2 Batch   | -> | Zero         | -> | State      |
| Submission |    | Knowledge    |    | Update     |
+------------+    | Verification |    +------------+
                  +--------------+
```

- **Batch Submission**: L2 transactions are batched and submitted to L1
- **Verification**: Zero-knowledge proofs verify the correctness of state transitions
- **State Update**: The verified state is then committed

## Key Interfaces and Extension Points

CDK-Erigon provides several extension points for developers:

- **RPC Extensions**: Add custom JSON-RPC methods
- **Event Hooks**: Plug into various parts of the execution pipeline
- **Custom Verification**: Implement custom verification logic
- **Storage Adapters**: Customize how data is stored and retrieved

## Performance Considerations

CDK-Erigon is optimized for:
- Memory efficiency through smart data structures
- Disk I/O optimization via batched writes
- Parallel execution where possible
- Caching of frequently accessed data

Typical resource requirements:
- **Memory**: 8-16GB minimum, 32GB+ recommended
- **Disk**: Fast SSD with 500GB+ capacity
- **CPU**: Modern multi-core processor
- **Network**: High-bandwidth, low-latency connection 