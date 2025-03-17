# Consensus in CDK-Erigon

This document explains the consensus mechanisms used in CDK-Erigon, focusing on how the zkEVM achieves agreement on the state of the Layer 2 blockchain.

## Overview

CDK-Erigon implements a specialized consensus mechanism that differs significantly from traditional blockchain consensus systems like Proof of Work (PoW) or Proof of Stake (PoS). Instead, it uses a hybrid approach that leverages:

1. **Sequencer-based Transaction Ordering**
2. **Layer 1 Ethereum for Finality**
3. **Zero-Knowledge Proofs for Validity**

```
+---------------------+       +----------------------+       +---------------------+
|                     |       |                      |       |                     |
| Transaction         |       | L1 Data              |       | Zero-Knowledge      |
| Sequencing          | ----> | Availability         | ----> | Verification        |
|                     |       |                      |       |                     |
+---------------------+       +----------------------+       +---------------------+
```

## Consensus Components

### 1. Sequencer

The sequencer is responsible for:

- Receiving user transactions
- Ordering transactions into batches
- Executing transactions to compute state transitions
- Publishing transaction data and state roots to Layer 1

#### Sequencer Operation

```
+-------------------+      +-------------------+      +-------------------+
| User Transactions | ---> | Sequencer Orders  | ---> | Execute & Compute |
| Received          |      | Into Batches      |      | State Transitions |
+-------------------+      +-------------------+      +-------------------+
                                                               |
                                                               v
                    +-------------------+      +-------------------+
                    | Publish to Layer 1|<---- | Generate State    |
                    | (Data & Roots)    |      | Roots             |
                    +-------------------+      +-------------------+
```

### 2. Data Availability on Layer 1

L1 Ethereum is used as a secure data availability layer:

- Batch data is published to L1 contracts
- This ensures all network participants can access transaction data
- Data availability is crucial for fraud detection and verification

### 3. Zero-Knowledge Proofs

ZK proofs provide cryptographic guarantees of state validity:

- Proves that all state transitions follow the correct execution rules
- Eliminates the need for nodes to re-execute all transactions
- Enables efficient verification of computational integrity

## Consensus Flow

### 1. Transaction Submission and Execution

1. Users submit transactions to the sequencer
2. The sequencer orders transactions into batches
3. Batches are executed to compute state transitions
4. State roots are calculated for each batch

### 2. Data Publishing

1. Batch data is published to Layer 1
2. Publication includes:
   - Transaction data
   - State roots
   - Batch metadata

### 3. Proof Generation and Verification

1. Zero-knowledge proofs are generated for batches
2. Proofs verify that state transitions are valid
3. Verified proofs are submitted to L1 smart contracts

### 4. Finality

1. Batches become "pending" when published to L1
2. Batches become "verified" when valid proofs are accepted
3. L1 Ethereum consensus provides ultimate finality

## Trust Assumptions

CDK-Erigon's consensus relies on several trust assumptions:

1. **Sequencer Trust**:
   - By default, the sequencer is trusted to order transactions
   - Censorship resistance is provided by the ability to submit transactions directly to L1

2. **Layer 1 Security**:
   - The security of the consensus mechanism depends on Ethereum's security
   - L1 reorgs can affect L2 state

3. **Cryptographic Assumptions**:
   - Security of the ZK proof system
   - Collision resistance of hash functions

## Consensus Parameters

| Parameter | Description |
|-----------|-------------|
| `zkevm.sequencer.enabled` | Whether this node acts as a sequencer |
| `zkevm.sequencer.max-batch-size` | Maximum transactions per batch |
| `zkevm.sequencer.max-batches-per-publish` | Maximum batches to publish in one L1 transaction |
| `zkevm.sequencer.batch-timeout` | Time before closing a batch even if not full |
| `zkevm.sequencer.max-l1-gas-price` | Maximum L1 gas price to pay for publishing |
| `zkevm.aggregator.enabled` | Whether this node acts as a proof aggregator |
| `zkevm.aggregator.max-verify-batch-size` | Maximum batches to verify in one proof |

## Liveness and Safety Guarantees

### Liveness Guarantees

CDK-Erigon ensures liveness through:

1. **Sequencer Redundancy**:
   - Multiple sequencers can be operated to prevent single points of failure
   - Configuration: `zkevm.sequencer.failover-enabled`

2. **Forced Inclusion**:
   - Users can submit transactions directly to L1 for forced inclusion
   - Protects against sequencer censorship

3. **Transaction Timeouts**:
   - Batches are published even if not full after timeout
   - Configuration: `zkevm.sequencer.batch-timeout`

### Safety Guarantees

Safety is ensured through:

1. **Zero-Knowledge Proofs**:
   - Cryptographic verification of state transitions
   - Invalid state transitions cannot be verified

2. **Layer 1 Security**:
   - All critical data is anchored to Ethereum
   - Inherits Ethereum's security properties

3. **Global Exit Root**:
   - Cryptographic commitment to L2 state
   - Enables secure bridging between L1 and L2

## Permissionless vs. Permissioned Operation

CDK-Erigon supports both permissionless and permissioned operation modes:

### Permissionless Mode

- Anyone can run a node and verify the L2 state
- Full verification requires:
  - L1 data access
  - ZK verification capabilities

### Permissioned Mode

- Restricted node operation
- Used for private or consortium networks
- Configuration: `zkevm.permissioned-contracts` and `zkevm.allowed-operators`

## Fork Choice Rules

Unlike traditional blockchains, CDK-Erigon's fork choice is deterministic and based on L1 data:

1. **L1 Canonical Chain**:
   - L2 follows the canonical L1 chain
   - L1 reorgs trigger L2 reorgs

2. **Batch Ordering**:
   - Batches are ordered by their position in L1 blocks
   - Within an L1 block, batches are ordered by transaction index

3. **Verification Priority**:
   - Verified batches take precedence over unverified batches
   - Configuration: `zkevm.l2-short-circuit-to-verified-batch`

## Sequencer Selection and Operation

### Sequencer Configuration

To configure a node as a sequencer:

```yaml
zkevm:
  sequencer:
    enabled: true
    private-key: "0x..."  # Sequencer private key
    max-batch-size: 300
    batch-timeout: "10s"
    max-l1-gas-price: 100000000000  # 100 Gwei
```

### Sequencer Rotation

For networks with multiple sequencers:

1. **Time-based Rotation**:
   - Sequencers operate in pre-defined time slots
   - Configuration: `zkevm.sequencer.rotation-period`

2. **Contract-based Selection**:
   - Smart contract manages sequencer selection
   - Configuration: `zkevm.sequencer.selection-contract`

## Prover and Aggregator

The prover and aggregator components are responsible for generating and submitting zero-knowledge proofs:

```
+-------------------+     +-------------------+     +-------------------+
| Batches Published | --> | Prover Generates  | --> | Aggregator        |
| to L1             |     | ZK Proofs         |     | Submits Proofs    |
+-------------------+     +-------------------+     +-------------------+
```

### Prover Configuration

```yaml
zkevm:
  prover:
    enabled: true
    url: "http://localhost:50052"
    max-batch-size: 150
```

### Aggregator Configuration

```yaml
zkevm:
  aggregator:
    enabled: true
    private-key: "0x..."
    max-verify-batch-size: 10
    polling-interval: "1m"
```

## Performance and Scaling

CDK-Erigon's consensus is designed for high performance:

1. **High Transaction Throughput**:
   - Sequencer can order thousands of transactions per second
   - Execution is the main bottleneck

2. **Batching Efficiency**:
   - Multiple transactions are batched together
   - Reduces L1 data costs

3. **Proof Aggregation**:
   - Multiple batches can be verified in a single proof
   - Amortizes verification costs across batches

## Consensus State Verification

CDK-Erigon nodes verify consensus state through:

1. **Chain Synchronization**:
   - Nodes download and verify batch data from L1
   - Configuration: See `SYNC.md`

2. **State Verification**:
   - Nodes verify state transitions match published roots
   - Configuration: `zkevm.debug-disable-state-root-check`

3. **Proof Verification**:
   - Nodes verify zero-knowledge proofs
   - Enhanced security but higher computational requirements

## Security Considerations

When operating CDK-Erigon with consensus, consider:

1. **Sequencer Key Security**:
   - Protect sequencer private keys
   - Use HSMs or secure key management solutions

2. **Layer 1 Connectivity**:
   - Maintain reliable connections to multiple L1 providers
   - Monitor for L1 reorgs

3. **Proof System Vulnerabilities**:
   - Stay updated with the latest security patches
   - Monitor cryptographic research for vulnerabilities

## Fault Tolerance and Recovery

CDK-Erigon implements fault tolerance mechanisms:

1. **Sequencer Failover**:
   - Automatic switch to backup sequencers
   - Configuration: `zkevm.sequencer.failover-peers`

2. **State Restoration**:
   - Ability to reconstruct state from L1 data
   - Command: `erigon --zkEVM.rebuild-l2-state`

3. **Consensus Recovery**:
   - Automatic handling of L1 reorgs
   - Configuration: `zkevm.witness-unwind-limit`

## Troubleshooting Consensus Issues

### Common Consensus Problems

#### 1. Sequencer Issues

**Symptoms**:
- No new batches being created
- Transactions pending for extended periods

**Solutions**:
- Check sequencer logs for errors
- Verify L1 connectivity and gas prices
- Restart sequencer with `--zkEVM.sequencer.force-batch-close`

#### 2. Proof Verification Failures

**Symptoms**:
- Batches remain unverified
- Proof submission transactions fail

**Solutions**:
- Check prover connectivity
- Verify aggregator has sufficient ETH
- For testing, use `--zkEVM.debug-disable-proof-check`

#### 3. State Root Mismatches

**Symptoms**:
- Logs showing "state root mismatch"
- Sync stops at specific batch

**Solutions**:
- Verify executor configuration
- Check for incompatible software versions
- Try rebuilding L2 state: `--zkEVM.rebuild-l2-state`

## FAQs

**Q: How does CDK-Erigon's consensus compare to traditional blockchains?**

A: Unlike traditional blockchains which use PoW or PoS for consensus, CDK-Erigon relies on a sequencer for ordering, Ethereum L1 for data availability, and ZK proofs for validity. This hybrid approach enables high throughput while inheriting security from Ethereum.

**Q: Can multiple sequencers run simultaneously?**

A: Yes, multiple sequencers can run simultaneously, but only one is active at a time to provide transaction ordering. Sequencers can follow a rotation schedule or selection mechanism.

**Q: How are conflicting transactions handled?**

A: Since a single sequencer handles transaction ordering, there are no conflicting transactions in normal operation. If multiple sequencers attempt to propose batches, the L1 contract enforces a deterministic ordering based on the L1 block inclusion.

**Q: What happens during an L1 reorg?**

A: When an L1 reorg occurs, CDK-Erigon automatically reorgs the L2 chain to maintain consistency. Batches that were part of the reorged L1 blocks are unwound, and batches from the new canonical chain are applied.

**Q: How are consensus parameters updated?**

A: Consensus parameters can be updated through:
1. Configuration file changes
2. Smart contract upgrades for on-chain parameters
3. Protocol upgrades for fundamental consensus changes 