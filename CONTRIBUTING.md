# Contributing Guidelines

## Getting Started

### Clone the Repository

Begin by cloning the repository to your local machine:

```bash
git clone git@github.com:kkrt-labs/zk-pig.git
```

### Install Dependencies

Ensure you have the following tools installed:

- [Go 1.24.3](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/get-started/get-docker/)

### Verify Installation

Run the following command to confirm everything is set up correctly:

```bash
make test
```

## Developing in the Project

This repository includes a `Makefile` to streamline development tasks.

To see the available commands, run:

```bash
make help
```

## How to Contribute

We follow the [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow) branching strategy with the default branch named `main`.

### Create a New Branch

- Always create a new branch from the default `main` branch:

    ```bash
    git checkout main
    git pull
    git checkout -b feat/your-branch-name
    ```
- Use clear and descriptive branch names, such as:
  - `feat/add-login`
  - `fix/bug-1234`
  - `chore/fix-typo`
  - `test/generator`
  - `devops/update-workflow`
  - `docs/add-readme`
- While you can choose any prefix, we recommend picking from the following list
  - `feat/*` for a new feature (e.g. API route, major perf improvement)
  - `fix/*` for a bug fix
  - `chore/*` for a package upgrade, a configuration change, a refactor, a small perf improvement, a typo fix...
  - `test/*` for adding tests
  - `devops/*` for CI-CD, development environment upgrade, security upgrade, automation addition...
  - `docs/*` for documentation

### Make Changes

- Implement your changes, ensuring relevant tests and documentation are updated.
- Commit your work using the [Commit Message Convention](#commit-message-convention). Each commit should represent an isolated and complete change.
- Push your branch:

    ```bash
    git push origin feat/your-branch-name
    ```
- Regularly rebase and push your branch on `main` to integrate the latest changes and minimize merge conflicts:

    ```bash
    git pull
    git rebase main
    ```

#### Commit Message Convention

Commits can either adhere to
- Preferably [gitmoji](https://github.com/carloscuesta/gitmoji) standard
- Or [Conventional Commit v1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) standard

This ensures consistency and enables automated tooling.

### Submit a Pull Request

1. [Open a Pull Request](https://github.com/kkrt-labs/zk-pig/compare) from your branch targeting the `main` branch.
  - Provide a clear title respecting [gitmoji](https://github.com/carloscuesta/gitmoji) message convention
  - Provide a detailed description of the change
  - Reference associated issues resolved by the Pull Request.
  - If it is a breaking change, provide migration path for the change.
  - Set Pull Request labels (those are used for automated generation of release notes)
     - one of `type.feat/type.fix/type.chore/type.test/type.devops/type.docs`
     - one of `breaking-change/non-breaking-change` 
4. Ensure your Pull Request passes all CI checks. If it fails, make the necessary updates to resolve any issues.
5. Request reviews from relevant team members.
6. Address feedback and iterate on your changes as needed. And repeat the review process until your Pull Request is approved.
8. Merge your Pull Request following to the [merging strategy](#merging-strategy)

#### Merging Strategy

We use the [Squash & Merge](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges#squash-and-merge-your-commits) strategy for merging changes. This ensures that all commits in your branch are combined into a single, concise commit in the `main` branch.
