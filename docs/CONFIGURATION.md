# CDK-Erigon Configuration Guide

This document provides a comprehensive guide to configuring CDK-Erigon, including all available options, their meanings, and recommended values.

## Configuration Methods

CDK-Erigon can be configured using:

1. **Command-line flags**: Specified when starting the node
2. **Configuration files**: YAML or TOML files that contain all settings
3. **Environment variables**: For certain options in container environments

The configuration options are defined and processed in [`cmd/cdk-erigon/main.go`](../cmd/cdk-erigon/main.go) and [`cmd/cdk-erigon/config.go`](../cmd/cdk-erigon/config.go).

### Command-line Usage

```bash
./build/bin/cdk-erigon --config=config.yaml [additional flags]
```

### Example Configuration File

```yaml
# Network and chain settings
datadir: /path/to/data
chain: hermez-bali
http: true
private.api.addr: localhost:9092

# zkEVM specific settings
zkevm.l2-chain-id: 2440
zkevm.l2-sequencer-rpc-url: https://rpc.example.com/
zkevm.l2-datastreamer-url: datastream.example.com:6900
zkevm.l1-chain-id: 11155111
zkevm.l1-rpc-url: https://sepolia.infura.io/v3/your-api-key

# Contract addresses
zkevm.address-sequencer: "0x9aeCf44E36f20DC407d1A580630c9a2419912dcB"
zkevm.address-zkevm: "0x89BA0Ed947a88fe43c22Ae305C0713eC8a7Eb361"
zkevm.address-rollup: "0xE2EF6215aDc132Df6913C8DD16487aBF118d1764"
zkevm.address-ger-manager: "0x2968D6d736178f8FE7393CC33C87f29D9C287e78"

# Gas parameters
zkevm.default-gas-price: 1000000000
zkevm.max-gas-price: 0
zkevm.gas-price-factor: 0.12

# Synchronization settings
zkevm.l1-rollup-id: 1
zkevm.l1-first-block: 4794475
txpool.disable: true

# API settings
http.api: [eth, debug, net, trace, web3, erigon, zkevm]
http.addr: 0.0.0.0
http.vhosts: any
http.corsdomain: any
ws: true
```

## Core Configuration Options

### Node Settings

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `datadir` | string | `~/.local/share/erigon` | Data directory for the database and keystore | [`cmd/flags/flags.go:SetDataDir()`](../cmd/flags/flags.go) |
| `chain` | string | | Chain to use: `hermez-bali`, `hermez-cardona`, etc. | [`cmd/utils/flags.go:ChainFlag`](../cmd/utils/flags.go) |
| `config` | string | | Path to the configuration file | [`cmd/utils/flags.go:ConfigFileFlag`](../cmd/utils/flags.go) |
| `private.api.addr` | string | `localhost:9090` | Private API network address | [`cmd/utils/flags.go:PrivateApiAddr`](../cmd/utils/flags.go) |

### Network Settings

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `http` | bool | `false` | Enable the HTTP-RPC server | [`cmd/utils/flags.go:HTTPEnabledFlag`](../cmd/utils/flags.go) |
| `http.addr` | string | `localhost` | HTTP-RPC server listening interface | [`cmd/utils/flags.go:HTTPListenAddrFlag`](../cmd/utils/flags.go) |
| `http.port` | int | `8545` | HTTP-RPC server listening port | [`cmd/utils/flags.go:HTTPPortFlag`](../cmd/utils/flags.go) |
| `http.vhosts` | string | `localhost` | Comma separated list of virtual hostnames | [`cmd/utils/flags.go:HTTPVirtualHostsFlag`](../cmd/utils/flags.go) |
| `http.corsdomain` | string | | Comma separated list of domains for CORS | [`cmd/utils/flags.go:HTTPCORSDomainFlag`](../cmd/utils/flags.go) |
| `http.api` | []string | `["eth", "erigon"]` | API modules to enable via HTTP-RPC | [`cmd/utils/flags.go:HTTPApiFlag`](../cmd/utils/flags.go) |
| `ws` | bool | `false` | Enable the WS-RPC server | [`cmd/utils/flags.go:WSEnabledFlag`](../cmd/utils/flags.go) |
| `ws.addr` | string | `localhost` | WS-RPC server listening interface | [`cmd/utils/flags.go:WSListenAddrFlag`](../cmd/utils/flags.go) |
| `ws.port` | int | `8546` | WS-RPC server listening port | [`cmd/utils/flags.go:WSPortFlag`](../cmd/utils/flags.go) |
| `ws.api` | []string | `["eth", "erigon"]` | API modules to enable via WS-RPC | [`cmd/utils/flags.go:WSApiFlag`](../cmd/utils/flags.go) |

### Logging and Debugging

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `verbosity` | int | `3` | Logging verbosity (0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=trace) | [`cmd/utils/flags.go:VerbosityFlag`](../cmd/utils/flags.go) |
| `log.console` | bool | `false` | Output logs to console | [`cmd/utils/flags.go:LogConsoleFlag`](../cmd/utils/flags.go) |
| `log.console.verbosity` | string | `info` | Console log verbosity | [`cmd/utils/flags.go:LogConsoleVerbosityFlag`](../cmd/utils/flags.go) |
| `debug.timers` | bool | `false` | Enable debug timers | [`cmd/utils/flags.go:TimersFlag`](../cmd/utils/flags.go) |
| `pprof` | bool | `false` | Enable pprof HTTP server | [`cmd/utils/flags.go:PprofFlag`](../cmd/utils/flags.go) |
| `pprof.addr` | string | `localhost` | pprof HTTP server listening interface | [`cmd/utils/flags.go:PprofAddrFlag`](../cmd/utils/flags.go) |
| `pprof.port` | int | `6060` | pprof HTTP server listening port | [`cmd/utils/flags.go:PprofPortFlag`](../cmd/utils/flags.go) |

## zkEVM Configuration Options

The zkEVM configuration options are defined in [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) within the `Zk` struct.

### Network Configuration

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.l2-chain-id` | uint64 | | Layer 2 chain ID | [`eth/ethconfig/config.go:Zk.L2ChainID`](../eth/ethconfig/config.go) |
| `zkevm.l2-sequencer-rpc-url` | string | | URL of the L2 sequencer RPC endpoint | [`eth/ethconfig/config.go:Zk.L2SequencerRpcUrl`](../eth/ethconfig/config.go) |
| `zkevm.l2-datastreamer-url` | string | | URL of the datastreamer service | [`eth/ethconfig/config.go:Zk.L2DatastreamerUrl`](../eth/ethconfig/config.go) |
| `zkevm.l1-chain-id` | uint64 | | Layer 1 chain ID | [`eth/ethconfig/config.go:Zk.L1ChainID`](../eth/ethconfig/config.go) |
| `zkevm.l1-rpc-url` | string | | URL of the L1 Ethereum node | [`eth/ethconfig/config.go:Zk.L1RpcUrl`](../eth/ethconfig/config.go) |
| `zkevm.l1-first-block` | uint64 | | First L1 block to scan for zkEVM events | [`eth/ethconfig/config.go:Zk.L1FirstBlock`](../eth/ethconfig/config.go) |

### Contract Addresses

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.address-sequencer` | string | | Address of the sequencer | [`eth/ethconfig/config.go:Zk.AddressSequencer`](../eth/ethconfig/config.go) |
| `zkevm.address-zkevm` | string | | Address of the zkEVM contract on L1 | [`eth/ethconfig/config.go:Zk.AddressZkevm`](../eth/ethconfig/config.go) |
| `zkevm.address-rollup` | string | | Address of the Rollup contract on L1 | [`eth/ethconfig/config.go:Zk.AddressRollup`](../eth/ethconfig/config.go) |
| `zkevm.address-ger-manager` | string | | Address of the Global Exit Root Manager | [`eth/ethconfig/config.go:Zk.AddressGerManager`](../eth/ethconfig/config.go) |

### Gas and Transaction Parameters

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.default-gas-price` | uint64 | `1000000000` | Default gas price for L2 transactions | [`eth/ethconfig/config.go:Zk.DefaultGasPrice`](../eth/ethconfig/config.go) |
| `zkevm.max-gas-price` | uint64 | `0` | Maximum allowed gas price (0 = no limit) | [`eth/ethconfig/config.go:Zk.MaxGasPrice`](../eth/ethconfig/config.go) |
| `zkevm.gas-price-factor` | float64 | `0.12` | Factor for calculating gas prices | [`eth/ethconfig/config.go:Zk.GasPriceFactor`](../eth/ethconfig/config.go) |
| `zkevm.l1-rollup-id` | uint64 | `1` | Rollup ID on the L1 network | [`eth/ethconfig/config.go:Zk.L1RollupId`](../eth/ethconfig/config.go) |

### Sequencer Settings

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.sequencer-batch-seal-time` | duration | `10s` | Time interval for sealing batches | [`eth/ethconfig/config.go:Zk.SequencerBatchSealTime`](../eth/ethconfig/config.go) |
| `zkevm.sequencer-non-empty-batch-seal-time` | duration | `60s` | Time interval for sealing non-empty batches | [`eth/ethconfig/config.go:Zk.SequencerNonEmptyBatchSealTime`](../eth/ethconfig/config.go) |
| `zkevm.sequencer-timeout-on-empty-tx-pool` | duration | `1s` | Timeout when transaction pool is empty | [`eth/ethconfig/config.go:Zk.SequencerTimeoutOnEmptyTxPool`](../eth/ethconfig/config.go) |
| `zkevm.sequencer-halt-on-batch-number` | uint64 | `0` | Batch number to halt sequencer (0 = don't halt) | [`eth/ethconfig/config.go:Zk.SequencerHaltOnBatchNumber`](../eth/ethconfig/config.go) |
| `zkevm.sequencer-resequence` | bool | `false` | Enable transaction resequencing | [`eth/ethconfig/config.go:Zk.SequencerResequence`](../eth/ethconfig/config.go) |
| `zkevm.sequencer-resequence-strict` | bool | `false` | Enforce strict resequencing | [`eth/ethconfig/config.go:Zk.SequencerResequenceStrict`](../eth/ethconfig/config.go) |
| `zkevm.sequencer-decoded-tx-cache-size` | int | `100000` | Size of decoded transaction cache | [`eth/ethconfig/config.go:Zk.SequencerDecodedTxCacheSize`](../eth/ethconfig/config.go) |
| `zkevm.sequencer-decoded-tx-cache-ttl` | duration | `10m` | TTL for decoded transaction cache | [`eth/ethconfig/config.go:Zk.SequencerDecodedTxCacheTTL`](../eth/ethconfig/config.go) |

### Executor Configuration

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.executor-urls` | []string | | URLs for external executors | [`eth/ethconfig/config.go:Zk.ExecutorUrls`](../eth/ethconfig/config.go) |
| `zkevm.executor-strict-mode` | bool | `false` | Enforce strict executor mode | [`eth/ethconfig/config.go:Zk.ExecutorStrictMode`](../eth/ethconfig/config.go) |
| `zkevm.executor-request-timeout` | duration | `10s` | Timeout for executor requests | [`eth/ethconfig/config.go:Zk.ExecutorRequestTimeout`](../eth/ethconfig/config.go) |
| `zkevm.executor-enabled` | bool | `true` | Enable executor functionality | [`eth/ethconfig/config.go:Zk.ExecutorEnabled`](../eth/ethconfig/config.go) |
| `zkevm.executor-max-concurrent-requests` | int | `10` | Maximum concurrent executor requests | [`eth/ethconfig/config.go:Zk.ExecutorMaxConcurrentRequests`](../eth/ethconfig/config.go) |

### Debug and Performance Options

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.debug-no-sync` | bool | `false` | Disable synchronization (for debugging) | [`eth/ethconfig/config.go:Zk.DebugNoSync`](../eth/ethconfig/config.go) |
| `zkevm.disable-virtual-counters` | bool | `false` | Disable virtual counters for performance | [`eth/ethconfig/config.go:Zk.DisableVirtualCounters`](../eth/ethconfig/config.go) |
| `zkevm.l1-contract-address-check` | bool | `true` | Check L1 contract addresses | [`eth/ethconfig/config.go:Zk.L1ContractAddressCheck`](../eth/ethconfig/config.go) |
| `zkevm.debug-disable-state-root-check` | bool | `false` | Disable state root verification | [`eth/ethconfig/config.go:Zk.DebugDisableStateRootCheck`](../eth/ethconfig/config.go) |
| `zkevm.witness-memdb-size` | bytesize | `4GB` | Memory allocated for witness generation | [`eth/ethconfig/config.go:Zk.WitnessMemdbSize`](../eth/ethconfig/config.go) |

### Datastreamer Configuration

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.l2-datastreamer-timeout` | duration | `60s` | Timeout for datastreamer connections | [`eth/ethconfig/config.go:Zk.L2DatastreamerTimeout`](../eth/ethconfig/config.go) |
| `zkevm.l2-datastreamer-useTLS` | bool | `false` | Use TLS for datastreamer connections | [`eth/ethconfig/config.go:Zk.L2DatastreamerUseTLS`](../eth/ethconfig/config.go) |
| `zkevm.l2-datastreamer-max-entry-chan` | uint64 | `1000` | Maximum entries in datastreamer channel | [`eth/ethconfig/config.go:Zk.L2DatastreamerMaxEntryChan`](../eth/ethconfig/config.go) |
| `zkevm.datastream-version` | string | `v1` | Datastreamer protocol version | [`eth/ethconfig/config.go:Zk.DatastreamVersion`](../eth/ethconfig/config.go) |

### RPC Configuration

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.rpc-ratelimits` | int | `0` | Rate limits for RPC calls (0 = no limit) | [`eth/ethconfig/config.go:Zk.RpcRatelimits`](../eth/ethconfig/config.go) |
| `zkevm.rpc-get-batch-witness-concurrency-limit` | int | `5` | Concurrency limit for batch witness requests | [`eth/ethconfig/config.go:Zk.RpcGetBatchWitnessConcurrencyLimit`](../eth/ethconfig/config.go) |

### Witness Cache Settings

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.witness-cache-enabled` | bool | `true` | Enable witness caching | [`eth/ethconfig/config.go:Zk.WitnessCacheEnabled`](../eth/ethconfig/config.go) |
| `zkevm.witness-cache-purge` | bool | `false` | Purge witness cache on startup | [`eth/ethconfig/config.go:Zk.WitnessCachePurge`](../eth/ethconfig/config.go) |
| `zkevm.witness-cache-batch-ahead-offset` | uint64 | `10` | Number of batches to cache ahead | [`eth/ethconfig/config.go:Zk.WitnessCacheBatchAheadOffset`](../eth/ethconfig/config.go) |
| `zkevm.witness-cache-batch-behind-offset` | uint64 | `50` | Number of batches to keep behind | [`eth/ethconfig/config.go:Zk.WitnessCacheBatchBehindOffset`](../eth/ethconfig/config.go) |
| `zkevm.witness-unwind-limit` | uint64 | `10` | Maximum batches to unwind | [`eth/ethconfig/config.go:Zk.WitnessUnwindLimit`](../eth/ethconfig/config.go) |

### Miscellaneous Settings

| Option | Type | Default | Description | Code Reference |
|--------|------|---------|-------------|---------------|
| `zkevm.bad-batches` | []uint64 | | List of known bad batches to skip | [`eth/ethconfig/config.go:Zk.BadBatches`](../eth/ethconfig/config.go) |
| `zkevm.ignore-bad-batches-check` | bool | `false` | Ignore bad batch checks | [`eth/ethconfig/config.go:Zk.IgnoreBadBatchesCheck`](../eth/ethconfig/config.go) |
| `zkevm.smt-regenerate-in-memory` | bool | `false` | Regenerate SMT in memory | [`eth/ethconfig/config.go:Zk.SmtRegenerateInMemory`](../eth/ethconfig/config.go) |
| `zkevm.log-level` | string | `info` | Log level for zkEVM components | [`eth/ethconfig/config.go:Zk.LogLevel`](../eth/ethconfig/config.go) |
| `zkevm.panel-manager-url` | string | | URL for the pool manager service | [`eth/ethconfig/config.go:Zk.PanelManagerUrl`](../eth/ethconfig/config.go) |

## Example Configurations

### Full Node Configuration

```yaml
datadir: /data/erigon
chain: hermez-bali
http: true
private.api.addr: localhost:9092
zkevm.l2-chain-id: 2440
zkevm.l1-chain-id: 11155111
zkevm.l1-rpc-url: https://sepolia.infura.io/v3/your-api-key

zkevm.address-sequencer: "0x9aeCf44E36f20DC407d1A580630c9a2419912dcB"
zkevm.address-zkevm: "0x89BA0Ed947a88fe43c22Ae305C0713eC8a7Eb361"
zkevm.address-rollup: "0xE2EF6215aDc132Df6913C8DD16487aBF118d1764"
zkevm.address-ger-manager: "0x2968D6d736178f8FE7393CC33C87f29D9C287e78"

zkevm.l1-rollup-id: 1
zkevm.l1-first-block: 4794475
txpool.disable: false

http.api: [eth, debug, net, trace, web3, erigon, zkevm]
http.addr: 0.0.0.0
http.vhosts: any
http.corsdomain: any
ws: true
```

### Sequencer Configuration

```yaml
datadir: /data/erigon-sequencer
chain: hermez-bali
http: true
private.api.addr: localhost:9092
zkevm.l2-chain-id: 2440
zkevm.l1-chain-id: 11155111
zkevm.l1-rpc-url: https://sepolia.infura.io/v3/your-api-key

zkevm.address-sequencer: "0x9aeCf44E36f20DC407d1A580630c9a2419912dcB"
zkevm.address-zkevm: "0x89BA0Ed947a88fe43c22Ae305C0713eC8a7Eb361"
zkevm.address-rollup: "0xE2EF6215aDc132Df6913C8DD16487aBF118d1764"
zkevm.address-ger-manager: "0x2968D6d736178f8FE7393CC33C87f29D9C287e78"

zkevm.default-gas-price: 1000000000
zkevm.max-gas-price: 100000000000
zkevm.gas-price-factor: 0.12

zkevm.l1-rollup-id: 1
zkevm.l1-first-block: 4794475
txpool.disable: false

zkevm.sequencer-batch-seal-time: 30s
zkevm.sequencer-non-empty-batch-seal-time: 10s
zkevm.sequencer-resequence: true

http.api: [eth, debug, net, trace, web3, erigon, zkevm]
http.addr: 0.0.0.0
http.vhosts: any
http.corsdomain: any
ws: true
```

### Low Resource Configuration

For machines with limited resources:

```yaml
datadir: /data/erigon-light
chain: hermez-bali
http: true
private.api.addr: localhost:9092
zkevm.l2-chain-id: 2440
zkevm.l1-chain-id: 11155111
zkevm.l1-rpc-url: https://sepolia.infura.io/v3/your-api-key

zkevm.address-sequencer: "0x9aeCf44E36f20DC407d1A580630c9a2419912dcB"
zkevm.address-zkevm: "0x89BA0Ed947a88fe43c22Ae305C0713eC8a7Eb361"
zkevm.address-rollup: "0xE2EF6215aDc132Df6913C8DD16487aBF118d1764"
zkevm.address-ger-manager: "0x2968D6d736178f8FE7393CC33C87f29D9C287e78"

zkevm.l1-rollup-id: 1
zkevm.l1-first-block: 4794475
txpool.disable: true
zkevm.disable-virtual-counters: true
zkevm.witness-memdb-size: 1GB

http.api: [eth, web3, erigon, zkevm]
http.addr: 0.0.0.0
http.vhosts: any
ws: false
```

## Configuration Implementation

The configuration handling is implemented in the following key files:

1. [`cmd/cdk-erigon/main.go`](../cmd/cdk-erigon/main.go) - Entry point that processes command-line flags and config files
2. [`cmd/cdk-erigon/config.go`](../cmd/cdk-erigon/config.go) - Config file parsing and validation
3. [`cmd/utils/flags.go`](../cmd/utils/flags.go) - Definition of command-line flags
4. [`eth/ethconfig/config.go`](../eth/ethconfig/config.go) - Core configuration structures and defaults
5. [`eth/backend.go`](../eth/backend.go) - Applies configuration to the node components

## Troubleshooting Configuration Issues

### Memory-Related Crashes

If you're experiencing memory-related crashes:

1. Reduce memory usage with `zkevm.disable-virtual-counters: true` - Implemented in [`zk/legacy_executor_verifier/executor_client.go`](../zk/legacy_executor_verifier/executor_client.go)
2. Adjust `zkevm.witness-memdb-size` to a lower value - Implemented in [`zk/stages/stage_witness.go`](../zk/stages/stage_witness.go)
3. Run with environment variable `GOGC=50` to increase garbage collection frequency

### Synchronization Issues

For synchronization problems:

1. Verify your `zkevm.l1-rpc-url` is operational and has sufficient rate limits - Used in [`zk/syncer/l1_syncer.go`](../zk/syncer/l1_syncer.go)
2. Check `zkevm.l1-first-block` is set correctly for your network - Used in [`zk/stages/stage_l1_syncer.go`](../zk/stages/stage_l1_syncer.go)
3. For temporary testing, use `zkevm.debug-disable-state-root-check: true` - Implemented in [`zk/stages/stage_zk_interhashes.go`](../zk/stages/stage_zk_interhashes.go)