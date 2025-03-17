# CDK-Erigon Documentation

Welcome to the documentation for CDK-Erigon, a zkEVM implementation built on the Erigon Ethereum client. CDK-Erigon provides a Layer 2 scaling solution with zero-knowledge proofs for high throughput and security.

## Documentation Index

### Architecture and Design

- [Architecture Overview](ARCHITECTURE.md) - High-level architecture of CDK-Erigon
- [zkEVM Implementation](ZKEVM.md) - Detailed documentation of the zkEVM component
- [Consensus Mechanism](CONSENSUS.md) - Explanation of the consensus system
- [Bridges and Interoperability](BRIDGES.md) - Details on L1/L2 bridging mechanisms
- [Execution Flow](EXECUTION_FLOW.md) - Comprehensive sequence diagram of the entire execution flow

### Configuration and Setup

- [Configuration Guide](CONFIGURATION.md) - Complete guide to configuring CDK-Erigon
- [Synchronization](SYNC.md) - Details on the L1/L2 synchronization process

### API Reference

- [API Documentation](API.md) - Comprehensive reference for JSON-RPC APIs

### Development

- [Development Guide](DEVELOPMENT.md) - Guide for developers contributing to CDK-Erigon

## Getting Started

For new users, we recommend the following reading path:

1. Start with the [Architecture Overview](ARCHITECTURE.md) to understand the system
2. Review the [Execution Flow](EXECUTION_FLOW.md) to understand how everything fits together
3. Review the [Configuration Guide](CONFIGURATION.md) to set up your node
4. Understand the [Synchronization](SYNC.md) process for running a node
5. Learn about the [Bridges and Interoperability](BRIDGES.md) for cross-chain interactions
6. Explore the [API Documentation](API.md) for interacting with your node

## Common Use Cases

### Running a Full Node

To run a full node that syncs with the network:

1. Follow the [Configuration Guide](CONFIGURATION.md)
2. Focus on the sections about node configuration and synchronization
3. Monitor your node using the metrics described in [Synchronization](SYNC.md)

### Operating a Sequencer

For those wanting to run a sequencer:

1. Review the [Consensus Mechanism](CONSENSUS.md) document
2. Study the sequencer operation in the [Execution Flow](EXECUTION_FLOW.md)
3. Configure your node according to the sequencer section in [Configuration Guide](CONFIGURATION.md)
4. Understand the security implications detailed in [Consensus Mechanism](CONSENSUS.md)

### Bridging Assets Between L1 and L2

For users wanting to bridge assets:

1. Read the [Bridges and Interoperability](BRIDGES.md) document
2. Understand the bridging operations and security mechanisms
3. Follow the bridge API examples for common use cases

### Development and Testing

For developers:

1. Start with the [Development Guide](DEVELOPMENT.md)
2. Understand the [zkEVM Implementation](ZKEVM.md)
3. Learn about the APIs through the [API Documentation](API.md)

## Support and Community

If you need help, consider these resources:

- GitHub Issues: [https://github.com/0xPolygon/cdk-erigon/issues](https://github.com/0xPolygon/cdk-erigon/issues)
- Discord: [Polygon Discord](https://discord.gg/polygon)
- Polygon Forum: [https://forum.polygon.technology/](https://forum.polygon.technology/)

## License

CDK-Erigon is licensed under the GNU Lesser General Public License v3.0 or later (LGPL-v3+).

## Acknowledgements

CDK-Erigon builds upon:
- The Erigon Ethereum client
- The Polygon zkEVM project 