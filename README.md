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
