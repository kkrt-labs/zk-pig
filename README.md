# Kakarot Controller

The **Kakarot Controller** is a monorepo housing all the services necessary for managing and orchestrating Kakarot proving operations.

## Installation

The `kkrtctl` application is distributed with Homebrew. 

Given the application is private, you need to
- have access to the private repository
- configure a [GitHub personal access token](https://github.com/settings/tokens/new) with scope `repo`

If installing for the first time you'll need to add the `kkrtlabs/kkrt` tap

```sh
brew tap kkrtlabs/kkrt
```

Then run 

```sh
export HOMEBREW_GITHUB_API_TOKEN=<access token>
brew install kkrtcl
```

Test installation

```sh
kkrtctl version
```

## Contributing

Interested in contributing? Check out our [Contributing Guidelines](CONTRIBUTING.md) to get started! 


## Local Usage

The `kkrtctl` CLI now supports multiple commands related to generating EVM prover inputs. Below are the key commands and how to use them locally with minimal setup.

### Prerequisites

1. You have an Ethereum node accessible via JSON-RPC (e.g., Geth, Erigon, Infura, etc.).
1. You have set up your environment variables (if needed):
    - `RPC_URL` for the default Ethereum node URL (e.g., http://127.0.0.1:8545)
    - `DATA_DIR` for where locally generated block data will be stored (e.g., ./data)

### Commands Overview 
1. `kkrtctl prover-inputs generate`
> Description: Automatically performs all three major phases:
> - Preflight: fetches necessary pre-state data.
> - Prepare: converts that data into final ProverInputs.
> - Execute: validates the prepared inputs by re-running the block execution

#### Usage
```sh
kkrtctl prover-inputs generate \
  --block-number 1234 \
  --rpc-url http://127.0.0.1:8545 \
  --data-dir ./data
```

2. `kkrtctl prover-inputs preflight`
> Description: Only fetches and stores the heavy pre-state (e.g., proofs, code). This is useful if you want to run the first step of proving separately.

#### Usage
```sh
kkrtctl prover-inputs preflight \
  --block-number 1234 \
  --rpc-url http://127.0.0.1:8545 \
  --data-dir ./data
```

3. `kkrtctl prover-inputs prepare`
> Description: Converts the “heavy” preflight data into the minimal, final ProverInputs.
> Requires a valid --chain-id, since it uses chain metadata for final block validation.

#### Usage
```sh
kkrtctl prover-inputs prepare \
  --chain-id 1 \
  --block-number 1234 \
  --rpc-url http://127.0.0.1:8545 \
  --data-dir ./data
```

4. `kkrtctl prover-inputs execute`
> Description: Re-executes the block with the final ProverInputs to verify correctness.
> Also requires --chain-id.

#### Usage
```sh
kkrtctl prover-inputs execute \
  --chain-id 1 \
  --block-number 1234 \
  --rpc-url http://127.0.0.1:8545 \
  --data-dir ./data
```

### Logging
Use `--log-level` to configure verbosity (`debug`, `info`, `warn`, `error`) and `--log-format` to switch between json and text. For example:

```sh
kkrtctl prover-inputs generate \
  --block-number 1234 \
  --rpc-url http://127.0.0.1:8545 \
  --data-dir ./data \
  --log-level debug \
  --log-format text
```

### Environment Variables Fallback
- `--rpc-url` falls back to `RPC_URL` if not explicitly set.
- `--data-dir` falls back to `DATA_DIR`.
- `--log-level` falls back to `LOG_LEVEL`.
- `--log-format` falls back to `LOG_FORMAT`.


### Commands

1. `kkrtctl version` - Print the version number
1. `kkrtctl prover-inputs generate --block-number <block-number> --rpc-url <rpc-url> --data-dir <data-dir>` - Generate prover inputs for a specific block
1. `kkrtctl prover-inputs preflight --block-number <block-number> --rpc-url <rpc-url> --data-dir <data-dir>` - Preflight the prover inputs generation
1. `kkrtctl prover-inputs prepare --block-number <block-number> --rpc-url <rpc-url> --data-dir <data-dir>` - Prepare the prover inputs generation
1. `kkrtctl prover-inputs execute --block-number <block-number> --rpc-url <rpc-url> --data-dir <data-dir>` - Execute the prover inputs generation
