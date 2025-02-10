# Prover Inputs Generation

This page describes the process for generating Prover Input the minimal necessary data for ZK EVM prover engines to prove a block. 

It consists of 3 steps: [1. Preflight](#step-1-preflight), [2. Prepare](#step-2-prepare) and [3. Execute](#step-3-execute)

## About Ethereum

This documentation assumes some knowledge about Ethereum state transition. If you are not familiar with it, you can refer to the [About Ethereum](about-ethereum.md) page.

## What Are Prover Inputs?

ZK proving engines operate in isolated & stateless environments without a direct access to a full blockchain node. 

**Prover Inputs** refer to the minimal data required by a ZK EVM proving engine to effectively prove a block in an isolated & stateless environment (including processing the block, computing the final state root, and validating both the block and the final state). They include:

- **Block**: The Ethereum block to be executed, including the block header and all transactions.
- **Chain Configuration**: The chain identifier and fork configurations.
- **Witness**: chain and state data accessed during block execution 
  - **Ancestors**: At minimum, the parent header, and optionally all block headers up to the oldest ancestor block accessed during execution (maximum 256 entries).
  - **Codes**: Bytecode of all smart contracts called during execution.
  - **PreState**: The partial pre-state accessed during execution, represented as a list of RLP-encoded MPT nodes (both account storage and all storage tries in the same list).
- **AccessList**: A mapping of accessed state entries (accounts and storage) during block execution. While this data is actually redundant, it currently facilitate some ZK EVM prover engines by enabling to perform some pre-state validations before executing the block. On the long run we may re-assess if this is absolutely needed.

## Generation of Prover Inputs

### Approach Overview

The current approach consists in collecting Prover Inputs data from a remote Ethereum-compatible JSON-RPC node (full node for recent blocks or archive node for older blocks). This method ensures compatibility across chains exposing Ethereum compatible JSON-RPC API and is relatively lightweight to maintain. Although it incurs performance overhead due to multiple API calls, it offers an acceptable trade-off given current proving times.

Alternative approaches, such as integrating prover input generation directly within a full node, offer better performance but require new implementations for each EVM chain.

### Generation Flow

Generating Prover Inputs for a block involves three consecutive steps, each requiring an EVM block execution:

1. **Preflight**: Executes the block online using a remote RPC backend, it collects intermediary preflight data `PreflightData`.
2. **Prepare**: Executes the block offline using a memory backend and optimizes the inputs into final `ProverInput`.
3. **Execute**: Validates the generated `ProverInput` by executing the block offline with them.

#### Diagram

![Prover Inputs Generation Flow](./zk-pig.png)

#### Step 1: Preflight

This step retrieves necessary data from a remote JSON-RPC node. It runs in an online environment.

It performs an EVM block execution using an RPC backend. During this execution we tracks state accesses (accounts and storage slots) so we can later fetch MPT proofs for every access. 

Any of the following EVM operations results in JSON-RPC call as follow

| **Operation**                     | **RPC Call**                                     |
|---------------------------------|-------------------------------------------------|
| Access to an account            | `eth_getProof`                                  |
| Opcode `SLOAD`                  | `eth_getStorageAt`                              |
| Opcode `BLOCKHASH`              | Series of `eth_getBlockByHash` calls            |
| Smart contract call             | `eth_getCode`                                   |

> 💡 Preflight EVM execution only processes the block but it does not validate the final state

At the end of the EVM execution, preflight fetches proofs for:

- **Pre-state**: All accessed state entries via `eth_getProof(account, accessedSlots, parent.Number)`.
- **Post-state**: Destructed accounts (`eth_getProof(destructedAccount, [], block.number)`) and deleted storage slots (`eth_getProof(account, deletedStorage, block.number)`)


> 💡 For more details you can refer to the [Pre-State Preparation Documentation](modified-mpt.md#pre-state-preparation-workflow)

The intermediary `PreflightData` contains:

- **Block**: The Ethereum block to be executed, including the block header and all transactions.
- **Chain Configuration**: The chain identifier and fork configurations.
- **Ancestors**: At minimum, the parent header, and optionally all block headers up to the oldest ancestor block accessed during execution (maximum 256 entries).
- **Codes**: Bytecode of all smart contracts called during execution.
- **Pre-State Proofs**: The list of pre-state proofs for every accounts and storage accessed during block execution, obtained via `eth_getProof(account, accessedSlots, parent.Number)` after preflight block execution
- **Post-State Proofs**: The list of post-state proofs for every destructed accounts and deleted storage during block execution, obtained via `eth_getProof(..., block.Number)` after preflight block execution

The `PreflightData` contain redundant data which is later optimized during [Prepare](#step-2-prepare).

#### Step 2: Prepare

This step optimizes `PreflightData` into final `ProverInput` offline. It reduces the `Pre-State Proofs` and `Post-State Proofs` that contain redundant and unecessary data into `Pre-State` a minimal list of MPT nodes necessary for the EVM block execution. 

It:

- Initializes a chain and state in memory using `PreflightData` (codes, ancestors, and proofs).
- Executes the EVM by processing block AND validating final state.
- Generates `ProverInput` based on the witness obtained from EVM execution.

During this step, a [modified MPT](modified-mpt.md#modified-mpt-implementation) is used, ensuring effective and compatible deletions.

#### Step 3: Execute

This step validates the generated `ProverInput`. It consists in running an EVM execution in an off-line isolated environment basing only on `ProverInput` data.

It:

- Initializes a chain and state in memory using `ProverInput` (codes, ancestors, and preState).
- Executes the EVM by processing block AND validating final state.

During this step, a [modified MPT](modified-mpt.md#modified-mpt-implementation) is used, ensuring effective and compatible deletions.

## Definitions

- **EVM**: The Ethereum Virtual Machine, responsible for executing smart contracts.
- **Smart Contract**: Program running on the Ethereum Virtual Machine (EVM).
- **Block**: A unit of data containing transactions, pre-state roots, and metadata, forming the blockchain. Each blocks induces a transition of state.
- **Transaction**: An Ethereum transaction inducing a state transitions. It triggers the EVM execution of a smart contract.
- **Merkle Patricia Trie**: A data structure combining Merkle trees and Patricia tries, used to efficiently store and verify Ethereum states.
- **State**: The complete data of accounts and storage on the Ethereum blockchain at a given block.
- **Account**: An entity in Ethereum that can hold ETH and interact with contracts, represented as an object in the state trie.
- **Storage**: Key-value pairs representing data for smart contracts, stored in a dedicated MPT.
- **Partial Pre-state**: The subset of the state required to execute a block successfully.
- **Stateless**: Generally refers to EVM executions without access to a full blockchain state  
- **Witness**: Supplemental ancestor headers, smart contract codes, pre-state MPT nodes accessed during an EVM block execution.



