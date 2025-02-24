# ZK-PIG

![Test](https://github.com/kkrt-labs/zk-pig/actions/workflows/test.yml/badge.svg?branch=main)  
[![codecov](https://codecov.io/gh/kkrt-labs/zk-pig/graph/badge.svg?token=ML8SpNgYm1)](https://codecov.io/gh/kkrt-labs/zk-pig)  
[![API Reference](https://pkg.go.dev/badge/github.com/kkrt-labs/zk-pig)](https://pkg.go.dev/github.com/kkrt-labs/zk-pig?tab=doc)  
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/kkrt-labs/zk-pig/blob/master/LICENSE.md)  
[![Twitter](https://img.shields.io/twitter/follow/KakarotZkEvm.svg)](https://x.com/intent/follow?screen_name=KakarotZkEvm)

**ZK-PIG** is a ZK-EVM Prover Input generator responsible for generating the data inputs necessary for proving Execution Layer (EL) blocks. These prover inputs can later be consumed by proving infrastructures to generate EL block proofs.

From an architecture perspective, ZK-PIG connects to an Ethereum-compatible EL full or archive node via JSON-RPC to fetch the necessary data.

> **Note about Prover Inputs:** ZK-EVM proving engines operate in isolated & stateless environments without direct access to a full blockchain node. The **Prover Input** refers to the minimal EVM data required by such a ZK-EVM proving engine to effectively prove an EL block. For more information on prover inputs, you can refer to this [article](https://ethresear.ch/t/zk-evm-prover-input-standardization/21626).

The **Kakarot Controller** is a monorepo housing all the services necessary for managing and orchestrating Kakarot proving operations.

## Installation

### Homebrew

`zkpig` is distributed with Homebrew.

If installing for the first time, you'll need to add the `kkrt-labs/kkrt` tap:

```sh
brew tap kkrt-labs/kkrt
```

Then to install `zkpig`, run:

```sh
brew install zkpig
```

To test the installation, run:

```sh
zkpig version
```

## Architecture

For more detailed architecture documentation, you can refer to the [Documentation](https://github.com/kkrt-labs/zk-pig/blob/main/docs/prover-input-generation.md).

## Contributing

Interested in contributing? Check out our [Contributing Guidelines](CONTRIBUTING.md) to get started!

## Usage

### Prerequisites

- You have an Ethereum Execution Layer-compatible node accessible via JSON-RPC (e.g., Geth, Erigon, Infura, etc.).

    > **⚠️ Warning ⚠️:** If generating prover inputs for an old block, you must use an Ethereum archive node that effectively exposes `eth_getProof` JSON-RPC for the block in question. Otherwise, ZK-PIG will fail at generating the prover inputs due to missing data.

    > **Note:** ZK-PIG is compatible with both HTTP and WebSocket JSON-RPC endpoints.

### Generate Prover Inputs

First, set the `CHAIN_RPC_URL` environment variable to the URL of the Ethereum node from which to collect data:

```sh
export CHAIN_RPC_URL=<rpc-url>
```

To generate prover inputs for a given block, use the following command:

```sh
zkpig generate --block-number <block-number>
```

> **Note:** The command takes around 1 minute to complete, mainly due to the time it takes to fetch the necessary data from the Ethereum node (around 2,000 requests/block).

On successful completion, the prover inputs are stored in the `/data` directory.

To generate prover inputs for the `latest` block, use the following command:

```sh
zkpig generate
```

For more information on the commands, you can use the following command:

```sh
zkpig generate --help
```

### Logging

To configure logging, you can set:
- `--log-level` to configure verbosity (`debug`, `info`, `warn`, `error`)
- `--log-format` to switch between `json` and `text`. For example:

```sh
zkpig generate \
  --block-number 1234 \
  --log-level debug \
  --log-format text
```

## Commands Overview

To get the list of all available commands and flags, you can run:

```sh
zkpig help
```

### `zkpig generate`

> Description: Generates prover inputs for a given block. It consists of running preflight, prepare, and execute in a single run.

#### Usage

```sh
zkpig generate \
  --block-number 1234 \
  --chain-rpc-url http://127.0.0.1:8545 \
  --data-dir ./data \
  --inputs-content-type json
```

### `zkpig preflight`

> Description: Only fetches and locally stores the necessary data (e.g., pre-state, block, transactions, state proofs, etc.) but does not run block validation. This is useful if you want to collect the data for a block and run block validation separately. It is also useful for debugging purposes.

#### Usage

```sh
zkpig preflight \
  --block-number 1234 \
  --chain-rpc-url http://127.0.0.1:8545 \
  --data-dir ./data \
  --inputs-content-type json
```

### `zkpig prepare`

> Description: Converts the data collected during preflight into the minimal, final prover input.  
> Can be run offline without a chain-rpc-url. In that case, it needs to be provided with a chain-id.

#### Usage

```sh
zkpig prepare \
  --block-number 1234 \
  --chain-id 1 \
  --chain-rpc-url http://127.0.0.1:8545 \
  --data-dir ./data \
  --inputs-content-type json
```

### `zkpig execute`

> Description: Re-executes the block over the previously generated prover inputs.  
> Can be run offline without a chain-rpc-url. In that case, it needs to be provided with a chain-id.

#### Usage

```sh
zkpig execute \
  --chain-id 1 \
  --block-number 1234 \
  --chain-rpc-url http://127.0.0.1:8545 \
  --data-dir ./data \
  --inputs-content-type json
```

## Caveats and Known Issues

### Flaky Behavior

- **Description**: There are instances where the system may exhibit flaky behavior. This can manifest as intermittent failures or unexpected results during execution.
- **Potential Causes**: 
  - Network instability
  - Resource constraints
  - Timing issues in asynchronous operations
- **Mitigation Strategies**:
  - Ensure stable network conditions.
  - Try running the command again.
- **Reporting**: If you encounter flaky behavior, please report it with detailed logs and steps to reproduce the issue.
