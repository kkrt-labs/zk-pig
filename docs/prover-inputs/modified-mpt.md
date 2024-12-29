
# Modified Patricia Merkle Trie (MPT)

This section describes the modified Merkle Patricia Trie (MPT) implementation necessary during [prepare](prover-inputs-generation.md#step-2-prepare) and [execute](prover-inputs-generation.md#step-3-prepare) of Prover Inputs generation.

## Background on Merkle Patricia Trie

The Ethereum state (a representation of all accounts and smart contract storage) is encoded using a specialized version of a **Merkle Tree**, called the **Merkle Patricia Trie (MPT)**. This data structure enables cryptographically verifiable relationships between its components, ensuring that a single root value can be used to validate the entirety of the data within the tree.

### MPT Nodes and Lookup Process

The Ethereum MPT is a data structure designed to store and retrieve data in a `path → value` format. It is composed of nodes that reference each other to enable efficient path-based lookups.

#### MPT Node Types

1. **Leaf Node**: A two-item node `[path, value]` where:  
   - `path` is the "partial path" to skip ahead in the trie.  
   - `value` is the user data stored in the MPT.  

2. **Extension Node**: A two-item node `[path, key]` where:  
   - `path` is the "partial path" to skip ahead in the trie.  
   - `key` is the reference to the child node.  

3. **Branch Node**: A 17-item node `[k0, ..., k15, value]` where:  
   - `ki` is the reference to the child node corresponding to the nibble `i` of the path.  
   - `value` is the data stored if the branch itself corresponds to a valid path 

4. **Null Node**: Represented as an empty string, indicating the absence of a node.

#### MPT Lookup Process

To find a value at a given `path`, the MPT traverses nodes recursively starting from the root:
- **Leaf Node**: If the leaf's `path` matches the remaining untraversed portion of the `path`, return the `value`; otherwise, return 0.
- **Extension Node**: If the extension's `path` matches the remaining untraversed portion of the `path`, resolve the child node and lookup from child node; otherwise, return 0.
- **Branch Node**: Resolve the child node referenced by the next nibble of the untraversed portion of the `path` and lookup from child node
- **Null Node**: Return 0.

#### MPT Node Keys

Nodes are referenced by a `key` that depends on their serialized RLP representation:
- **Short Nodes**: If the RLP encoded node's size is less than 32 bytes (`len(rlp(node)) < 32`), the `key` is the RLP encoding itself. These nodes are not stored in the underlying database.
- **Long Nodes**: For larger nodes (which is the vast majority of nodes), the `key` is the Keccak-256 hash of the RLP encoding (`keccak(rlp(node))`). These nodes are stored in the underlying Key-Value database.

When looking up an extension or branch node, a child node may need to be resolved:
- If the `key` is less than 32 bytes, decode it directly.
- Otherwise, fetch the corresponding node from the Key-Value database and decode it.

### MPT Operations

#### Types of Operations

1. **Addition**: Add a new `value` at a specified `path`.  
2. **Update**: Modify an existing `value` at a specified `path`.  
3. **Deletion**: Remove a `value` from a specified `path`.

#### Operation Details

##### Addition
To add a `value` at `path`:
1. Traverse the trie to find the deepest existing node matching the longest prefix of `path`.
2. Depending on the type of node encountered:
   - **Branch Node**: Add a new leaf node and update the branch by adding the new leaf.
   - **Short Node** (Leaf/Extension): Create a new branch node with two children: the new leaf and the converted short node.

##### Update
To update a `value` at `path`:
1. Locate the leaf node corresponding to `path`.
2. Add a new leaf with the same leaf's path and the updated `value`.
3. Update the parent node to reference the new leaf.

##### Deletion

Deleting is a more complex process as it possibly requires to reduce nodes, as MPTs ensure that 
1. a branch node have at least 2 non-null children otherwise it must be converted in a short node (i.e. leaf or extension node)
2. there cannot be 2 consecutive short nodes otherwise they must be merged together

To delete a value at a given `path` starts by looking up the for the corresponding leaf and update the parent's node by replacing the deleted child with a null node.

###### Extension node reduction

During deletion if an extension node is left with an null child, then it is deleted as well.

###### Branch node reduction

During deletion if a branch node is left with a single non-null child, then it is reduced as follow
1. The branch node is converted in a one-nibble extension node with the remaining child node
2. Then MPT resolves the remaining child node. If the child node is a short node (i.e., an extension or a leaf), then the child node is merged into the reduced branch node
3. Then if the parent of the branch node is a short node, the reduced branch node is merged into the parent node

## Problem with Partial States

In the context, of EVM block proving, blocks are executed in a stateless environment with access only to a partial portion of the pre-state which only includes a subset of nodes.

To generate the partial pre-state we use data fetched from a remote JSON-RPC node during [Preflight](prover-inputs-generation.md#step-1-preflight) using `eth_getProof` JSON-RPC calls. 

While preflight enables to retrieve all the MPT nodes on the path of every accessed accounts and storage, and re-constitute the partial pre-state necessary for the block processing, it may miss the MPT nodes resolved during [branch node reduction](#branch-node-reduction) following deletions. This creates a significant issue: the inability to compute the final state root after processing the block.

## Solution

### Modified MPT Implementation

We use a modified MPT implementation with a single difference from the standard MPT implementation: when handling [branch node reduction on a deletion](#branch-node-reduction), if the remaining child cannot be resolved (because it is missing from Key-Value database), then, instead of raising an error (as per standard implementation), the branch node is reduced into a one-nibble short node (exactly as if the remaining child was not a short node as per standard implementation).

This modified MPT is used during both [Prepare](prover-inputs-generation.md#step-2-prepare) and [Execute](prover-inputs-generation.md#step-3-execute).

### Supplemental Technique: Hypothetise child short-nodes using post-state

To further address this issue, we use a supplemental technique during the [Prepare](prover-inputs-generation.md#step-2-prepare) phase. 

For every deletion resulting in a branch node reduction, we pre-inject some hypothetized short-nodes into the pre-state, ensuring that if the remaining child is a short-node, then it resolve's to one of the pre-injected hypothetized short nodes. Consequently, ensuring that, during branch node reduction, the remaining child node resolution will succeed.

To compute the hypothetized short-nodes, we base on the post-state proof for every deleted MPT path. Indeed, in case of a deletion, the post-state proof is actually an exclusion proof, in which the last proof's element is an MPT node proving that there is no value at the given path. If this last proof's item is a short-node, this means that the deletion triggered a branch node reduction, and it is possible to compute all the short-nodes that could have potentially reduced into the last proof's element (see [below](#pre-state-preparation-workflow)).

#### Pre-State Preparation Workflow

To prepare a pre-state for a deletion at `path` (resulting from an account destruction or a storage value set to zero during EVM block execution).

1. Retrieve from a full node (using `eth_getProof` JSON-RPC):

- Pre-state proof: Inclusion proof `[n_1, n_2, ..., n_p]` for `path`
- Post-state proof: Exclusion proof `[n'_1, n'_2, ..., n'_q]` for `path`

2. Inject `[n_1, n_2, ..., n_p]` into the pre-state for the base state access.
3. If `n'_q` is not a short node (i.e an extension or a leaf) or it is a one-nibble short node `[[nib_1],key)]`, no further action is needed.
4. If `n'_q` is a short node `[[nib_1,nib_2,...,nib_l],key)]` with path length `l` greater or equal 2, then compute all potential short nodes that could reduce into `n'_q` i.e all short nodes with shorter prefix key `[[nib_2,...,nib_l],key], [[nib_3,...,nib_l],key]`, ...,`[[nib_l],key]`
5. Inject all short nodes in pre-state

This ensures that all short nodes that could potentially reduce into `n'_q` have been pre-injected into the pre-state.

> **NOTE** This approach is actually a bit overkill, as `[n_1, n_2, ..., n_p]` and `[n'_1, n'_2, ..., n'_q]` would allow to infer the exact reduction scenario and missing child node of the reduced branch node thus injecting a single node in pre-state instead of a list. We leave it for future optimization.

### Result

As a result, during prepare the following scenarios can occur, when reducing a branch node and resolving the remaining child node
- child is found, which means that the remaining child node is a short node and it has been pre-injected during [pre-state preparation](#supplemental-technique-pre-injected-short-nodes-from-post-state-in-pre-state) then the branch node is reduced as per standard implementation.
- child is not found, then the [Modified MPT implementation](#modified-mpt-implementation) reduces branch node in a one-nibble short-node. This is exactly what the standard MPT implementation would have done with full-state access, indeed if the remaining child is not found, it implies that it is not a short-node (or it would have been pre-injected and found), thus the standard implementation would reduce the branch node into in a one-nibbe short-node.

So in both scenarios, the behavior is the same as a standard Ethereum MPT implementation with full-state access.

### Considerations

This solution requires access to **both pre-state and post-state data**, making it unsuitable for live block execution. It is intended for proving blocks after execution on a full node.