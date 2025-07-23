# Patternizer

[![Quay Repository](https://img.shields.io/badge/Quay.io-patternizer-blue?logo=quay)](https://quay.io/repository/dminnear/patternizer)
[![CI Pipeline](https://github.com/dminnear-rh/patternizer/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/dminnear-rh/patternizer/actions/workflows/ci.yaml)

**Patternizer** is a command-line tool that bootstraps a Git repository containing Helm charts into a ready-to-use Validated Pattern. It automatically generates the necessary scaffolding, configuration files, and utility scripts, so you can get your pattern up and running in minutes.

> **Note:** This tool was developed with assistance from [Cursor](https://cursor.sh), an AI-powered code editor.

- [Patternizer](#patternizer)
  - [Features](#features)
  - [Quick Start](#quick-start)
  - [Example Workflow](#example-workflow)
  - [Usage Details](#usage-details)
    - [Container Usage (Recommended)](#container-usage-recommended)
      - [**Initialize without secrets:**](#initialize-without-secrets)
      - [**Initialize with secrets support:**](#initialize-with-secrets-support)
    - [Understanding Secrets Management](#understanding-secrets-management)
    - [Generated Files](#generated-files)
  - [Development \& Contributing](#development--contributing)
    - [Prerequisites](#prerequisites)
    - [Local Development Workflow](#local-development-workflow)
    - [Common Makefile Targets](#common-makefile-targets)
    - [Testing Strategy](#testing-strategy)
    - [Architecture](#architecture)
    - [CI/CD Pipeline](#cicd-pipeline)
    - [How to Contribute](#how-to-contribute)

## Features

  - 🚀 **CLI-first design** with intuitive commands and help system
  - 📦 **Container-native** for consistent execution across all environments
  - 🔍 **Auto-discovery** of Helm charts and Git repository metadata
  - 🔐 **Optional secrets integration** with Vault and External Secrets
  - 🏗️ **Makefile-driven** utility scripts for easy pattern management

## Quick Start

This guide assumes you have a Git repository containing one or more Helm charts.

**Prerequisites:**

  * Podman or Docker
  * A Git repository you want to convert into a pattern

Navigate to your repository's root directory and run the initialization command:

```bash
# In the root of your pattern-repo
podman run --pull=always --rm -it -v .:/repo:z quay.io/dminnear/patternizer init
```

This single command will generate all the necessary files to turn your repository into a Validated Pattern.

## Example Workflow

1.  **Clone or create your pattern repository:**

    ```bash
    git clone https://github.com/your-org/your-pattern.git
    cd your-pattern
    git checkout -b initialize-pattern
    ```

2.  **Initialize the pattern using Patternizer:**

    ```bash
    podman run --pull=always --rm -it -v .:/repo:z quay.io/dminnear/patternizer init
    ```

3.  **Review, commit, and push the generated files:**

    ```bash
    git status
    git add .
    git commit -m 'initialize pattern using patternizer'
    git push -u origin initialize-pattern
    ```

4.  **Install the pattern:**

    ```bash
    ./pattern.sh make install
    ```

## Usage Details

### Container Usage (Recommended)

Using the prebuilt container is the easiest way to run Patternizer, as it requires no local installation. The `-v .:/repo:z` flag mounts your current directory into the container's `/repo` workspace.

#### **Initialize without secrets:**

```bash
podman run --pull=always --rm -it -v .:/repo:z quay.io/dminnear/patternizer init
```

#### **Initialize with secrets support:**

```bash
podman run --pull=always --rm -it -v .:/repo:z quay.io/dminnear/patternizer init --with-secrets
```

### Understanding Secrets Management

You can start simple and add secrets management later.

  * By default, `patternizer init` disables secret loading.
  * To add secrets scaffolding, run `patternizer init --with-secrets` at any time. This will update your configuration to enable secrets.
  * **Important:** This action is not easily reversible. We recommend committing your work to Git *before* adding secrets support.

For more details on how secrets work in the framework, see the [Secrets Management Documentation](https://validatedpatterns.io/learn/secrets-management-in-the-validated-patterns-framework/).

### Generated Files

Running `patternizer init` creates the following:

  * `values-global.yaml`: Global pattern configuration.
  * `values-<cluster_group>.yaml`: Cluster group-specific values.
  * `pattern.sh`: A utility script for common pattern operations (`install`, `upgrade`, etc.).
  * `Makefile`: A simple Makefile that includes `Makefile-pattern`.
  * `Makefile-pattern`: The core Makefile with all pattern-related targets.

Using the `--with-secrets` flag additionally creates:

  * `values-secret.yaml.template`: A template for defining your secrets.
  * It also updates `values-global.yaml` to set `global.secretLoader.disabled: false` and adds Vault and External Secrets Operator to the cluster group values.

## Development & Contributing

This section is for developers who want to contribute to the Patternizer project itself.

### Prerequisites

  * Go (see `go.mod` for version)
  * Podman or Docker
  * Git
  * Make

### Local Development Workflow

```bash
# 1. Clone the repository
git clone https://github.com/dminnear-rh/patternizer.git
cd patternizer

# 2. Set up the development environment (installs tools)
make dev-setup

# 3. Make your changes...

# 4. Run the full CI suite locally before committing
make ci
```

### Common Makefile Targets

The `Makefile` is the single source of truth for all development and CI tasks.

  * `make help`: Show all available targets.
  * `make check`: Quick feedback loop (format, vet, build, unit tests).
  * `make build`: Build the `patternizer` binary.
  * `make test`: Run all tests (unit and integration).
  * `make test-unit`: Run unit tests only.
  * `make test-integration`: Run integration tests only.
  * `make lint`: Run all code quality checks.
  * `make local-container-build`: Build the container image locally.

### Testing Strategy

Patternizer has a comprehensive test suite to ensure stability and correctness.

  * **Unit Tests:** Located alongside the code they test (e.g., `src/internal/helm/helm_test.go`), these tests cover individual functions and packages in isolation. They validate Helm chart detection, Git URL parsing, and YAML processing logic.
  * **Integration Tests:** The integration test suite (`test/integration_test.sh`) validates the end-to-end CLI workflow against a real Git repository. Key scenarios include:
    1.  **Basic Init:** Validates default file generation without secrets.
    2.  **Init with Secrets:** Ensures secrets-related applications and files are correctly added.
    3.  **Configuration Preservation:** Verifies that existing custom values are preserved when the tool is re-run.
    4.  **Sequential Execution:** Tests running `init` and then `init --with-secrets` to ensure a clean upgrade.

### Architecture

The CLI is organized into focused packages following Go best practices, with a clean separation of concerns between command-line logic (`cmd`), core business logic (`internal`), and file operations (`fileutils`). This modular design makes the codebase maintainable, testable, and extensible.

### CI/CD Pipeline

The GitHub Actions pipeline (`.github/workflows/ci.yaml`) runs on every push and pull request. It uses the same `Makefile` targets that developers use locally, ensuring perfect consistency between local and CI environments.

### How to Contribute

1.  Fork the repository.
2.  Create a feature branch for your changes.
3.  Make your changes and ensure they pass the local CI check (`make ci`).
4.  Submit a pull request.
