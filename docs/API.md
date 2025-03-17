# CDK-Erigon API Reference

This document provides a comprehensive reference for the JSON-RPC APIs exposed by CDK-Erigon, including both standard Ethereum APIs and zkEVM-specific extensions.

## API Implementation Structure

The CDK-Erigon API implementation is organized as follows:

- **API Interfaces**: Defined in various files in the `turbo/jsonrpc` directory
  - `eth_api.go`: Standard Ethereum API interface and implementation
  - `zkevm_api.go`: zkEVM-specific API interface and implementation
  - `debug_api.go`: Debug API interface and implementation
  - `trace_api.go`: Trace API interface and implementation
  - `erigon_api.go`: Erigon-specific API interface and implementation

- **API Registration**: APIs are registered in `turbo/jsonrpc/daemon.go` and exposed via the HTTP and WebSocket servers

## API Overview

CDK-Erigon supports multiple API modules that can be enabled independently:

- `eth`: Standard Ethereum API ([Implementation](../turbo/jsonrpc/eth_api.go))
- `net`: Network information API ([Implementation](../turbo/jsonrpc/net_api.go))
- `web3`: Web3 compatibility API ([Implementation](../turbo/jsonrpc/web3_api.go))
- `debug`: Debugging API ([Implementation](../turbo/jsonrpc/debug_api.go))
- `trace`: Transaction tracing API ([Implementation](../turbo/jsonrpc/trace_api.go))
- `erigon`: Erigon-specific extensions ([Implementation](../turbo/jsonrpc/erigon_api.go))
- `zkevm`: zkEVM-specific extensions ([Implementation](../turbo/jsonrpc/zkevm_api.go))

## Enabling APIs

APIs can be enabled via configuration:

```yaml
# HTTP API
http: true
http.api: [eth, debug, net, trace, web3, erigon, zkevm]

# WebSocket API
ws: true
ws.api: [eth, debug, net, trace, web3, erigon, zkevm]
```

## Standard Ethereum APIs

CDK-Erigon implements the standard Ethereum JSON-RPC API as specified in the [Ethereum JSON-RPC documentation](https://ethereum.org/en/developers/docs/apis/json-rpc/).

### Common Ethereum Methods

#### Block and Chain Information

| Method | Description | Implementation |
|--------|-------------|----------------|
| `eth_blockNumber` | Returns the number of the most recent block | [`BlockNumber`](../turbo/jsonrpc/eth_api.go) |
| `eth_getBlockByNumber` | Returns information about a block by number | [`GetBlockByNumber`](../turbo/jsonrpc/eth_api.go) |
| `eth_getBlockByHash` | Returns information about a block by hash | [`GetBlockByHash`](../turbo/jsonrpc/eth_api.go) |
| `eth_chainId` | Returns the current chain ID | [`ChainId`](../turbo/jsonrpc/eth_api.go) |
| `eth_syncing` | Returns an object with data about the sync status or false | [`Syncing`](../turbo/jsonrpc/eth_api.go) |

#### Transaction Methods

| Method | Description | Implementation |
|--------|-------------|----------------|
| `eth_getTransactionByHash` | Returns information about a transaction by hash | [`GetTransactionByHash`](../turbo/jsonrpc/eth_api.go) |
| `eth_getTransactionByBlockHashAndIndex` | Returns information about a transaction by block hash and index | [`GetTransactionByBlockHashAndIndex`](../turbo/jsonrpc/eth_api.go) |
| `eth_getTransactionByBlockNumberAndIndex` | Returns information about a transaction by block number and index | [`GetTransactionByBlockNumberAndIndex`](../turbo/jsonrpc/eth_api.go) |
| `eth_getTransactionReceipt` | Returns the receipt of a transaction by hash | [`GetTransactionReceipt`](../turbo/jsonrpc/eth_api.go) |
| `eth_sendRawTransaction` | Sends a signed transaction | [`SendRawTransaction`](../turbo/jsonrpc/eth_api.go) |
| `eth_sendTransaction` | Creates and sends a transaction | [`SendTransaction`](../turbo/jsonrpc/eth_api.go) |
| `eth_estimateGas` | Estimates the gas required for a transaction | [`EstimateGas`](../turbo/jsonrpc/eth_api.go) |
| `eth_getTransactionCount` | Returns the number of transactions sent from an address | [`GetTransactionCount`](../turbo/jsonrpc/eth_api.go) |

#### Account Methods

| Method | Description | Implementation |
|--------|-------------|----------------|
| `eth_getBalance` | Returns the balance of an account | [`GetBalance`](../turbo/jsonrpc/eth_api.go) |
| `eth_getCode` | Returns the code at a given address | [`GetCode`](../turbo/jsonrpc/eth_api.go) |
| `eth_getStorageAt` | Returns the value from a storage position | [`GetStorageAt`](../turbo/jsonrpc/eth_api.go) |
| `eth_call` | Executes a new message call without creating a transaction | [`Call`](../turbo/jsonrpc/eth_api.go) |

#### Network Methods

| Method | Description | Implementation |
|--------|-------------|----------------|
| `net_version` | Returns the current network ID | [`Version`](../turbo/jsonrpc/net_api.go) |
| `net_listening` | Returns true if client is actively listening for network connections | [`Listening`](../turbo/jsonrpc/net_api.go) |
| `net_peerCount` | Returns number of peers currently connected to the client | [`PeerCount`](../turbo/jsonrpc/net_api.go) |
| `web3_clientVersion` | Returns the current client version | [`ClientVersion`](../turbo/jsonrpc/web3_api.go) |

## zkEVM-Specific APIs

The `zkevm` namespace provides methods specific to the zkEVM Layer 2 solution.

### Batch Methods

| Method | Description | Implementation |
|--------|-------------|----------------|
| `zkevm_batchNumber` | Returns the latest batch number | [`BatchNumber`](../turbo/jsonrpc/zkevm_api.go) |
| `zkevm_batchByNumber` | Returns batch information by number | [`BatchByNumber`](../turbo/jsonrpc/zkevm_api.go) |
| `zkevm_getBatchByBlockNumber` | Returns the batch containing a specific block | [`GetBatchByBlockNumber`](../turbo/jsonrpc/zkevm_api.go) |
| `zkevm_verifiedBatchNumber` | Returns the latest verified batch number | [`VerifiedBatchNumber`](../turbo/jsonrpc/zkevm_api.go) |

### State Methods

| Method | Description | Implementation |
|--------|-------------|----------------|
| `zkevm_getGlobalExitRoot` | Returns the current Global Exit Root | [`GetGlobalExitRoot`](../turbo/jsonrpc/zkevm_api.go) |
| `zkevm_isBlockVirtualized` | Checks if a block has been virtualized | [`IsBlockVirtualized`](../turbo/jsonrpc/zkevm_api.go) |
| `zkevm_getL1BatchInfo` | Returns L1 information for a given batch | [`GetL1BatchInfo`](../turbo/jsonrpc/zkevm_api.go) |

### Status Methods

| Method | Description | Implementation |
|--------|-------------|----------------|
| `zkevm_synced` | Returns true if node is fully synced | [`Synced`](../turbo/jsonrpc/zkevm_api.go) |
| `zkevm_networkConfig` | Returns network configuration parameters | [`NetworkConfig`](../turbo/jsonrpc/zkevm_api.go) |
| `zkevm_getLatestVerifiedBatchNum` | Returns the latest verified batch number | [`GetLatestVerifiedBatchNum`](../turbo/jsonrpc/zkevm_api.go) |

### Bridge Methods

| Method | Description | Implementation |
|--------|-------------|----------------|
| `zkevm_bridgeAssets` | Returns assets bridged between L1 and L2 | [`BridgeAssets`](../turbo/jsonrpc/zkevm_api.go) |
| `zkevm_claimStatus` | Returns the status of a claim | [`ClaimStatus`](../turbo/jsonrpc/zkevm_api.go) |

## Data Structures

### Batch Object

```json
{
  "number": "0x1b4",
  "globalExitRoot": "0x8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65d6b4b0f7ffc18957aa",
  "timestamp": "0x62fdd7c8",
  "stateRoot": "0xca3149fa9e37db08d1cd49c9061db1002ef1cd58db2210f2115c8c989b2bdf45",
  "transactions": [
    "0x5e77a04531c7c107af1882d76cbff9486d0a9aa53701c30888509d4f5f2b003a",
    "0x5e77a04531c7c107af1882d76cbff9486d0a9aa53701c30888509d4f5f2b003b"
  ],
  "sequencerAddress": "0x617b3a3528F9cDd6630fd3301B9c8911F7Bf063D",
  "verified": true,
  "verifiedAt": "0x63e0b882"
}
```

### L1 Batch Info Object

```json
{
  "blockNumber": "0x493e00",
  "transactionHash": "0x8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65d6b4b0f7ffc18957aa",
  "transactionIndex": "0x1",
  "receivedAt": "0x62fdd7c8",
  "verifiedAt": "0x63e0b882"
}
```

### Network Config Object

```json
{
  "chainId": "0x978",
  "l1ChainId": "0xaa36a7",
  "rollupId": "0x1",
  "sequencerAddress": "0x617b3a3528F9cDd6630fd3301B9c8911F7Bf063D",
  "zkevmAddress": "0x89BA0Ed947a88fe43c22Ae305C0713eC8a7Eb361",
  "rollupAddress": "0xE2EF6215aDc132Df6913C8DD16487aBF118d1764",
  "gerAddress": "0x2968D6d736178f8FE7393CC33C87f29D9C287e78"
}
```

## Debug API Extensions

The `debug` namespace provides methods for debugging and introspection.

### Common Debug Methods

| Method | Description | Implementation |
|--------|-------------|----------------|
| `debug_traceTransaction` | Returns the trace of a transaction | [`TraceTransaction`](../turbo/jsonrpc/debug_api.go) |
| `debug_traceBlockByNumber` | Returns the traces in the specified block | [`TraceBlockByNumber`](../turbo/jsonrpc/debug_api.go) |
| `debug_traceBlockByHash` | Returns the traces in the specified block | [`TraceBlockByHash`](../turbo/jsonrpc/debug_api.go) |
| `debug_traceBadBlock` | Returns the traces of a bad block | [`TraceBadBlock`](../turbo/jsonrpc/debug_api.go) |

### zkEVM Debug Extensions

| Method | Description | Implementation |
|--------|-------------|----------------|
| `debug_zkevmTraceTransaction` | Returns detailed zkEVM-specific trace of a transaction | [`TraceTransaction`](../turbo/jsonrpc/tracing_zkevm.go) |
| `debug_zkevmStateRoot` | Returns the state root at a given block | [`ZkEvmStateRoot`](../turbo/jsonrpc/debug_api.go) |
| `debug_zkevmBatchTotalElements` | Returns the total elements in a batch | [`ZkEvmBatchTotalElements`](../turbo/jsonrpc/debug_api.go) |

## Trace API

The `trace` namespace provides methods for tracing transactions.

| Method | Description | Implementation |
|--------|-------------|----------------|
| `trace_transaction` | Returns all traces produced at the given transaction | [`Transaction`](../turbo/jsonrpc/trace_api.go) |
| `trace_get` | Returns specific trace identified by index in a block | [`Get`](../turbo/jsonrpc/trace_api.go) |
| `trace_block` | Returns all traces produced at the given block | [`Block`](../turbo/jsonrpc/trace_api.go) |
| `trace_call` | Executes a call and returns a number of transaction traces | [`Call`](../turbo/jsonrpc/trace_api.go) |
| `trace_callMany` | Performs multiple call traces in a single request | [`CallMany`](../turbo/jsonrpc/trace_api.go) |
| `trace_rawTransaction` | Traces a raw signed transaction | [`RawTransaction`](../turbo/jsonrpc/trace_api.go) |

## Erigon-Specific API

The `erigon` namespace provides methods specific to Erigon.

| Method | Description | Implementation |
|--------|-------------|----------------|
| `erigon_forks` | Returns the block numbers of all hardforks | [`Forks`](../turbo/jsonrpc/erigon_api.go) |
| `erigon_issuance` | Returns the total issuance of a chain | [`Issuance`](../turbo/jsonrpc/erigon_api.go) |
| `erigon_getHeaderByNumber` | Returns a block header by number | [`GetHeaderByNumber`](../turbo/jsonrpc/erigon_api.go) |
| `erigon_getHeaderByHash` | Returns a block header by hash | [`GetHeaderByHash`](../turbo/jsonrpc/erigon_api.go) |
| `erigon_getLogsByHash` | Returns logs by transaction or block hash | [`GetLogsByHash`](../turbo/jsonrpc/erigon_api.go) |

## WebSocket API

The WebSocket API supports all HTTP RPC methods plus subscription methods:

| Method | Description | Implementation |
|--------|-------------|----------------|
| `eth_subscribe` | Creates a subscription to specific events | [`Subscribe`](../turbo/jsonrpc/eth_api.go) |
| `eth_unsubscribe` | Cancels an existing subscription | [`Unsubscribe`](../turbo/jsonrpc/eth_api.go) |

### Subscription Types

| Type | Description | Implementation |
|------|-------------|----------------|
| `newHeads` | Subscribes to new block headers | [`NewHeads`](../turbo/jsonrpc/eth_api.go) |
| `logs` | Subscribes to new logs matching the filter | [`Logs`](../turbo/jsonrpc/eth_api.go) |
| `newPendingTransactions` | Subscribes to pending transactions | [`NewPendingTransactions`](../turbo/jsonrpc/eth_api.go) |
| `syncing` | Subscribes to sync status events | [`Syncing`](../turbo/jsonrpc/zkevm_api.go) |

## Authentication

By default, RPC connections do not require authentication. However, for production environments, it's recommended to secure the API endpoints.

## Rate Limiting

CDK-Erigon supports rate limiting for RPC requests:

```yaml
zkevm.rpc-ratelimits: 100  # Limit to 100 requests per second
```

## API Examples

### Ethereum Examples

#### Getting the Latest Block Number

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "eth_blockNumber",
  "params": [],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x4b7" // 1207
}
```

#### Sending a Transaction

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "eth_sendRawTransaction",
  "params": ["0xf86d82010c8504a817c800825208943a8d39553410cc30e9ffa1e24a262de1fdb0bd0880de0b6b3a764000080820217a04fc8bcb847ac25ff9c453f2d225c8ceae66ac9d7178a3a38e4550a5b59fd3cb1a06c5b0da17fdaec2877da3d1c6bcf2a48ade6b56eb5dbe43e6b07ae3fe4029e3"],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x27d9a0e07da58a8c2b15b11fee299e2c25eab745daf90f92a6d80e8d2a1e0fcf"
}
```

### zkEVM Examples

#### Getting the Latest Batch Number

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "zkevm_batchNumber",
  "params": [],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x1b4" // 436
}
```

#### Getting Batch Information

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "zkevm_batchByNumber",
  "params": ["0x1b4", true],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "number": "0x1b4",
    "globalExitRoot": "0x8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65d6b4b0f7ffc18957aa",
    "timestamp": "0x62fdd7c8",
    "stateRoot": "0xca3149fa9e37db08d1cd49c9061db1002ef1cd58db2210f2115c8c989b2bdf45",
    "transactions": [{...}, {...}],
    "sequencerAddress": "0x617b3a3528F9cDd6630fd3301B9c8911F7Bf063D",
    "verified": true,
    "verifiedAt": "0x63e0b882"
  }
}
```

#### Getting the Global Exit Root

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "zkevm_getGlobalExitRoot",
  "params": [],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65d6b4b0f7ffc18957aa"
}
```

## Common Errors

| Code | Message | Description |
|------|---------|-------------|
| -32700 | Parse error | Invalid JSON |
| -32600 | Invalid request | The JSON sent is not a valid request object |
| -32601 | Method not found | The method does not exist / is not available |
| -32602 | Invalid params | Invalid method parameters |
| -32603 | Internal error | Internal JSON-RPC error |
| -32000 | Server error | Generic server-side error |
| -32001 | Execution error | Error during execution (for eth_call, etc.) |
| -32099 | Timeout | Request timeout |

## Performance Considerations

- Use batch requests to reduce network overhead
- Subscribe to events using WebSockets instead of polling
- Limit the number of blocks requested in range-based calls
- Use pagination for methods that return large result sets

## Further Reading

For more detailed information about specific APIs:
- [Ethereum JSON-RPC Documentation](https://ethereum.org/en/developers/docs/apis/json-rpc/)
- [Web3.js Documentation](https://web3js.readthedocs.io/)
- [Ethers.js Documentation](https://docs.ethers.io/) 