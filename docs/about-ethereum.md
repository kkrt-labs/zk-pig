# About Ethereum

## Generalities about state transition

EVM chains are programmable blockchains in which each block `B` (including a set of transactions) induces a state transition from a pre-state `S` as of the parent block to `S'` the final state after execution of the block. Given that block transitions are deterministic and that every node in the network executes the same sequence of blocks starting from the same initial state, it results in every node having the same identical state at any given block. This is what is commonly called a decentralized state machine.

## Ethereum State

The EVM state is composed of:

- **Account State (World State):** Contains account data, stored as a Merkle Patricia Trie (MPT), where:
  - **Key** is the Keccak hash of the account address.
  - **Value** is the RLP encoding of account data including
    - Balance: The ETH balance of the account.
    - Nonce: The count of transactions executed by the account.
    - Code Hash: In case, the account is a Smart Contract, this is the hash of corresponding bytecode.
    - Storage Root: In case, the account is a Smart Contract, this is the root hash of the MPT for storage of the account
- **Storage of Account:** Stored as an MPT, where:
  - **Key** is the Keccak hash of the slot index.
  - **Value** is the slot values.

The Ethereum state root is the Merkle Patricia root of the account state trie.

## EVM Block Execution

The execution of an EVM block involves operations such as processing transactions, handling system calls, accounting for block rewards... which results in a set of elementary state transitions that can be categorized as follows:

| **State Transition**            | **Description**                                                                                                    | **Example**                                                                                      |
|---------------------------|--------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| **Create an account**     | Create a new account object in the account state trie.                                                             | Deploying a new smart contract.                                                                 |
| **Update an account**     | Update one or more account fields: balance, nonce, or storage hash (Note: protocol does not allow to change `codeHash`).           | Updating balance when transferring ETH, updating nonce when executing a transaction, etc.       |
| **SelfDestruct an account** | Delete an account.                                                                                                | Self-destructing a smart contract.                                                              |
| **Create a storage**      | Set an empty storage slot to a non-null value.                                                                     | Any call to `SSTORE(slot, value)` on an empty slot with `value ≠ 0`.                        |
| **Update a storage**      | Set a non-empty storage slot to a non-null value.                                                                  | Any call to `SSTORE(slot, value)` on a non-empty slot with `value ≠ 0`.                     |
| **Delete a storage**      | Set a non-empty storage slot to zero.                                                                              | Any call to `SSTORE(slot, 0)` on a non-empty slot.                                              |

Starting from parent state `S`, applying all state transitions induced by the execution of block `B` results in the final state `S'`.

## Adding new blocks

Practically, the final state root `S'` is first computed by the block proposer and included in the block header at block proposal time. When other nodes receive the proposed block, they re-execute the block on their local state, validate the computed final state root against the one in the block header, and reject the block in case of a mismatch.

## Block witness

Block witness encompasses the minimal state and chain data required for a stateless EVM block execution (meaning without access to a a database containing the full state) which includes
- performing all block operations (apply transactions, apply fee rewards, apply system calls, etc.) 
- deriving the post state root

This includes
- **partial pre-state** containing the list of MPTS nodes from the accounts MPT and storage MPTs which have been resolved during block execution either 
   - when accessing a state data (either an account or a storage slot)
   - when deleting some state entries which may result in a [MPT branch node reduction](modified-mpt.md#branch-node-reduction) which resolves extra nodes
- **codes** of all smart contracts called during block execution
- **ancestors headers**, minimally containing the direct parent of the executed block and optionally older ancestors if accessed with opcode `BLOCKHASH` during block execution. For instance, the opcode `BLOCKHASH`enables smart contracts to access the hash of any of the 256 most recent blocks (excluding the current block, as its hash is computed post-execution)

### Witness examples

The table below is a non-exhaustive list describing the witness for some common operations of the block execution.

| **Operation**                                                                                                                                 	| **Account pre-state**                                                                                          	| **Storage pre-state**                                                                                          	| **Codes**                    	| **Ancestors**                                                                 	|
|-----------------------------------------------------------------------------------------------------------------------------------------------	|----------------------------------------------------------------------------------------------------------------	|----------------------------------------------------------------------------------------------------------------	|------------------------------	|-------------------------------------------------------------------------------	|
| Any Transaction                                                                                                                               	| - All nodes to EOA sender's account (for nonce checks and fee payment)<br>- Possibly more depending on the tx  	| - Depends on the tx                                                                                            	| - All smart contracts called 	| - Depends on the tx                                                           	|
| ETH Transfer from Bob to Alice                                                                                                                	| - All nodes to Bob's account (for balance update)<br>- All nodes to Alice's account (for balance update)       	| None                                                                                                           	| None                         	| None                                                                          	|
| [Simple ERC20](https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/token/ERC20/ERC20.sol) transfer from Bob to Alice 	| - All nodes to ERC20 account (for codeHash & storageHash)                                                      	| - ERC20: All nodes to Bob's balance slot<br>- ERC20: All nodes to Alice's balance slot                         	| - ERC20 bytecode             	| None                                                                          	|
| [EIP-4788 BeaconRoot](https://eips.ethereum.org/EIPS/eip-4788) system call                                                                    	| - All nodes to BeaconRoot account (for codeHash & storageHash)                                                 	| - BeaconRoot: All nodes to timestampid slot<br>- BeaconRoot: All nodes to rootid slot                          	| - BeaconRoot bytecode        	| None                                                                          	|
| Any contract call                                                                                                                             	| - All nodes to contract account (for codeHash & storageHash)                                                   	| - All nodes to all slots accessed during the call                                                              	| - Contract bytecode          	| - Depends on the call                                                         	|
| Transaction fees (or priority fees) to coinbase                                                                                               	| - All nodes to coinbase account (for balance update)                                                           	| None                                                                                                           	| None                         	| None                                                                          	|
| Op-code BLOCKHASH <blocknum>                                                                                                                  	| None                                                                                                           	| None                                                                                                           	| None                         	| - All headers from `currentBlockNum-1` down to `blockNum` (max. 256 headers)  	|
| For account destructions or storage deletions (see [Modified MPT Implementation](modified-mpt.md) for more details)                           	| - All nodes to the destructed account<br>- Possibly, the remaining child node in case of branch node reduction 	| - All nodes to the destructed storage<br>- Possibly, the remaining child node in case of branch node reduction 	| None                         	| None    