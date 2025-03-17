# Bridges and Interoperability in CDK-Erigon

This document explains the bridge mechanisms and interoperability features of CDK-Erigon, which enable assets and messages to move between Layer 1 (Ethereum) and Layer 2 (zkEVM).

## Overview

Bridges are essential components of any Layer 2 solution, as they allow users to move assets and data between the base layer (Ethereum) and the scaling layer (zkEVM). CDK-Erigon implements a secure bridging mechanism based on cryptographic proofs.

```
+--------------------+                                +--------------------+
|                    |                                |                    |
|  Layer 1           |                                |  Layer 2           |
|  (Ethereum)        |                                |  (zkEVM)           |
|                    |                                |                    |
+--------+-----------+                                +-----------+--------+
         |                                                        |
         |              +------------------------+                |
         |              |                        |                |
         +------------->|  Bridge Infrastructure |<---------------+
                        |                        |
                        +------------------------+
```

## Bridge Architecture

The CDK-Erigon bridge system consists of several components:

### 1. Global Exit Root Manager

The Global Exit Root Manager is a smart contract that stores and verifies cryptographic commitments to the state of both L1 and L2. It acts as the central coordination point for bridging operations.

### 2. Bridge Contracts

- **L1 Bridge Contract**: Deployed on Ethereum, handles asset deposits and claim verifications
- **L2 Bridge Contract**: Deployed on zkEVM, handles asset withdrawals and claim processing

### 3. Message Relay System

The message relay system transmits data between L1 and L2:

- **L1 to L2 Messages**: Sent through the sequencer
- **L2 to L1 Messages**: Verified with zero-knowledge proofs

## Bridging Operations

### L1 to L2 Transfers (Deposits)

When a user deposits assets from L1 to L2:

1. User initiates a deposit transaction on the L1 Bridge contract
2. The L1 Bridge locks the assets and emits a deposit event
3. The L1 state is updated in the Global Exit Root Manager
4. The sequencer includes this information in an L2 batch
5. The L2 Bridge mints or releases equivalent assets on L2

```
+---------------------+     +--------------------+     +---------------------+
| User Deposits on L1 | --> | L1 Bridge Locks    | --> | Global Exit Root    |
| Bridge              |     | Assets & Emits     |     | Updated             |
+---------------------+     +--------------------+     +---------------------+
                                                                 |
                                                                 v
                            +---------------------+     +---------------------+
                            | Assets Released     | <-- | Sequencer Processes |
                            | on L2               |     | Deposit on L2       |
                            +---------------------+     +---------------------+
```

### L2 to L1 Transfers (Withdrawals)

When a user withdraws assets from L2 to L1:

1. User initiates a withdrawal transaction on the L2 Bridge contract
2. The L2 Bridge locks or burns the assets and registers the claim
3. The claim is included in a batch and the L2 state root is updated
4. The batch is published to L1 along with the new L2 state root
5. The Global Exit Root Manager is updated with the new L2 state root
6. User can claim their assets on L1 after the appropriate waiting period

```
+---------------------+     +--------------------+     +----------------------+
| User Initiates      | --> | L2 Bridge Locks    | --> | Claim Included in    |
| Withdrawal on L2    |     | Assets & Records   |     | Batch & Published    |
+---------------------+     +--------------------+     +----------------------+
                                                                 |
                                                                 v
+---------------------+     +--------------------+     +----------------------+
| User Claims Assets  | <-- | Claim Verified     | <-- | Global Exit Root     |
| on L1 Bridge        |     | Against Exit Root  |     | Updated on L1        |
+---------------------+     +--------------------+     +----------------------+
```

## Bridge Security Mechanisms

### Cryptographic Verification

The bridge security relies on cryptographic verification of state roots:

1. **Merkle Trees**: State is organized in Merkle trees for efficient verification
2. **Global Exit Root**: Combines roots from both L1 and L2 states
3. **Zero-Knowledge Proofs**: Verify the validity of L2 state transitions

### Security Guarantees

The bridge provides the following security guarantees:

1. **Asset Safety**: Assets cannot be double-spent or created without corresponding lock/burn
2. **Censorship Resistance**: Users can force claim execution even if sequencers misbehave
3. **Finality Alignment**: Bridge operations respect the finality mechanisms of both chains

## Bridge Configuration

### L1 Bridge Configuration

```yaml
zkevm:
  bridge:
    l1-contract-address: "0x123...abc"  # Address of the L1 bridge contract
    l1-global-exit-root-manager-address: "0x456...def"  # Address of the exit root manager
    l1-claim-timeout: "1h"  # Timeout for L1 claims
```

### L2 Bridge Configuration

```yaml
zkevm:
  bridge:
    l2-contract-address: "0x789...ghi"  # Address of the L2 bridge contract
    l2-claim-gas-limit: 500000  # Gas limit for L2 claim transactions
```

## Supported Token Standards

The bridge supports various token standards:

1. **Native Tokens**: ETH on L1 and the equivalent native token on L2
2. **ERC-20 Tokens**: Fungible tokens following the ERC-20 standard
3. **ERC-721 Tokens**: Non-fungible tokens (NFTs)
4. **ERC-1155 Tokens**: Multi-tokens standard

## Bridge API

CDK-Erigon exposes several API methods to interact with the bridge:

### Deposit Information

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "zkevm_getDepositsByAddress",
  "params": ["0xuser_address", "pending"],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    {
      "depositIndex": 42,
      "originNetwork": 1,
      "destinationNetwork": 1001,
      "originTokenAddress": "0xtoken_address_on_l1",
      "amount": "1000000000000000000",
      "destinationAddress": "0xuser_address",
      "status": "pending",
      "timestamp": 1647352486,
      "claimTxHash": null
    }
  ]
}
```

### Bridge Transaction Status

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "zkevm_getBridgeTransactionStatus",
  "params": ["0xtx_hash"],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "depositIndex": 42,
    "status": "finalized",
    "lastUpdated": 1647352586,
    "claim": {
      "claimTxHash": "0xclaim_tx_hash",
      "claimTimestamp": 1647352540
    }
  }
}
```

### Global Exit Root Information

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "zkevm_getLatestGlobalExitRoot",
  "params": [],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "globalExitRoot": "0x1234...abcd",
    "l1ExitRoot": "0x2345...bcde",
    "l2ExitRoot": "0x3456...cdef",
    "lastUpdated": 1647352600
  }
}
```

## Message Passing

In addition to asset transfers, the bridge supports arbitrary message passing between L1 and L2:

### L1 to L2 Messages

Users can send messages from L1 to L2 by:

1. Calling the `sendMessage` function on the L1 Bridge contract
2. Specifying the destination address and calldata
3. The message is relayed to L2 and executed

### L2 to L1 Messages

Messages from L2 to L1 follow a similar pattern to asset withdrawals:

1. The message is registered on the L2 Bridge
2. Once included in a batch, it can be claimed on L1
3. Upon verification, the message is executed on L1

## Advanced Bridge Features

### Force Transactions

In case the sequencer censors transactions, users can submit force transactions on L1:

1. User submits transaction data to the Force Transactions contract on L1
2. After a timeout, the transaction must be included by sequencers
3. This ensures censorship resistance

### Batched Transfers

For gas efficiency, users can batch multiple transfers in a single transaction:

1. Group multiple token transfers in one bridge transaction
2. Save on gas costs for both L1 and L2 operations

## Bridge Economics

### Fee Structure

Bridge operations involve various fees:

1. **L1 Gas Fees**: Paid for transactions on Ethereum
2. **L2 Gas Fees**: Paid for transactions on zkEVM
3. **Bridge Service Fee**: Optional fee for bridge service providers

### Liquidity Providers

To improve user experience, liquidity providers can:

1. Offer instant withdrawals for a fee
2. Maintain liquidity on both L1 and L2
3. Earn fees from users who want to avoid waiting periods

## Common Bridge Use Cases

### Asset Transfer

The most common use case is transferring assets between L1 and L2:

```javascript
// Deposit ETH from L1 to L2
await l1Bridge.deposit(
  destinationAddress,
  ethers.constants.AddressZero, // ETH address (0x0)
  amount,
  { value: amount }
);

// Withdraw ETH from L2 to L1
await l2Bridge.withdraw(
  destinationAddress,
  ethers.constants.AddressZero, // ETH address (0x0)
  amount,
  { value: amount }
);
```

### Cross-Chain Dapp Communication

Dapps can communicate across layers:

```javascript
// Send message from L1 to L2
await l1Bridge.sendMessage(
  destinationAddress,
  calldata,
  { value: messageFee }
);

// Send message from L2 to L1
await l2Bridge.sendMessage(
  destinationAddress,
  calldata,
  { value: messageFee }
);
```

## Bridge Security Best Practices

When using the bridge:

1. **Wait for Finality**: Wait for appropriate finality periods before considering transactions final
2. **Verify Receipts**: Always verify transaction receipts
3. **Use Official Interfaces**: Use official bridge interfaces or trusted dapps
4. **Monitor Bridge Status**: Check bridge status before performing large transfers

## Troubleshooting Bridge Issues

### Common Bridge Problems

#### 1. Deposits Not Arriving on L2

**Symptoms**:
- Deposit transaction is confirmed on L1, but assets don't appear on L2

**Solutions**:
- Check the deposit status using `zkevm_getDepositsByAddress`
- Verify that the sequencer is processing batches
- Allow more time for L1 finality and batch processing

#### 2. Withdrawals Stuck in Pending

**Symptoms**:
- Withdrawal initiated on L2 but cannot be claimed on L1

**Solutions**:
- Verify the batch containing your withdrawal was published to L1
- Check if the waiting period has elapsed
- Ensure the Global Exit Root was updated correctly

#### 3. Claim Transaction Failures

**Symptoms**:
- Claim transaction reverts on L1

**Solutions**:
- Verify you're claiming the correct withdrawal
- Check if someone already claimed this withdrawal
- Ensure sufficient gas for the claim transaction

## Interoperability with Other Networks

CDK-Erigon can be configured to bridge with:

1. **Ethereum Mainnet**: Primary L1 for production
2. **Ethereum Testnets**: For testing (Goerli, Sepolia)
3. **Other EVM-Compatible Chains**: Through custom bridge deployments

## Bridge Monitoring and Analytics

### Key Metrics to Monitor

1. **Bridge Volume**: Total value bridged in both directions
2. **Active Deposits/Withdrawals**: Number of pending operations
3. **Bridge Utilization**: Usage patterns over time
4. **Bridge Status**: Overall health and performance

### Monitoring Tools

1. **Bridge Explorer**: Web interface to track bridge transactions
2. **Prometheus Metrics**: Expose bridge metrics from your node
3. **Alert System**: Configure alerts for bridge issues

## Future Bridge Improvements

Planned improvements to the bridge include:

1. **Multi-Hop Bridges**: Direct bridging between different L2s
2. **Optimistic Withdrawals**: Faster withdrawals with economic security
3. **Enhanced Security**: Additional verification mechanisms
4. **Improved UX**: Better user experience and simpler interfaces

## References

- [Ethereum Bridges Documentation](https://ethereum.org/en/bridges/)
- [zkEVM Bridge Specification](https://docs.polygon.technology/zkEVM/protocol/components/bridge/)
- [EIP-1559: Fee Market Change](https://eips.ethereum.org/EIPS/eip-1559)
- [EIP-2718: Typed Transaction Envelope](https://eips.ethereum.org/EIPS/eip-2718) 