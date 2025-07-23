# Patternizer

[![Quay Repository](https://img.shields.io/badge/Quay.io-patternizer-blue?logo=quay)](https://quay.io/repository/dminnear/patternizer)
[![CI Pipeline](https://github.com/dminnear-rh/patternizer/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/dminnear-rh/patternizer/actions/workflows/ci.yaml)

> **Note:** This tool was developed with assistance from [Cursor](https://cursor.sh), an AI-powered code editor.

**Patternizer** is a CLI tool and container utility designed to bootstrap Validated Pattern repositories. It automatically generates the necessary files for a Validated Pattern from a Git repo containing Helm charts.

Patternizer is available as a standalone CLI, however, it is most conventiently used from a container [as described below](#container-usage).

By default, Patternizer will not add the necessary files and values for loading secrets as part of your pattern. This is to make it simpler and clearer as you develop patterns which pieces are needed. You are able to run [`init --with-secrets`](#basic-initialization) at any point, even if you already ran the init without the secrets flag, to add the scaffolding for secrets into your pattern. So don't be afraid to start without secrets and then add them in as you want more control over DB credentials, ssh keys, etc.. When/if you do decide to add secrets please reference the [Secrets Management](https://validatedpatterns.io/learn/secrets-management-in-the-validated-patterns-framework/) section of our site. Finally, while you can always run init again with the --with-secrets flag to add secrets to your pattern, you cannot then run the init command without that flag to remove the secrets scaffolding that was added. If you're worried you may have such a case, then you should commit your changes to Git before running the init command with the --with-secrets flag.

---

## Features

- 🚀 **CLI-first design** with intuitive commands and help system
- 📦 **Container support** for consistent execution across environments
- 🔍 **Auto-discovery** of Helm charts and Git repository metadata
- 🔐 **Secrets integration** with Vault and External Secrets support
- ✅ **Comprehensive testing** with unit and integration tests
- 🏗️ **Multi-stage builds** for minimal container images
- 🛠️ **Makefile-driven development** for consistent local development and CI

## Quick Start for Developers

```bash
# Clone the repository
git clone https://github.com/dminnear-rh/patternizer.git
cd patternizer

# Set up development environment
make dev-setup

# See all available targets
make help

# Quick feedback loop (format check, vet, build, unit tests)
make check
```

---

## CLI Usage

[Container usage](#container-usage) is recommended but you are also able to run the CLI directly by following the steps below.

```bash
# Build the binary
make build

# Copy the binary somewhere onto your $PATH
cp src/patternizer /usr/local/bin
```

### Available Commands

```bash
# Show help and available commands
patternizer help

# Initialize pattern files (without secret loading)
patternizer init

# Initialize pattern files with secret loading
patternizer init --with-secrets

# Show help for specific commands
patternizer init help
```

### Output Files

The `patternizer init` command generates:

- `values-global.yaml` - Global pattern configuration with `global.secretLoader.disabled: true`
- `values-<cluster_group>.yaml` - Cluster group-specific configuration
- `pattern.sh` - Utility script for pattern operations
- `Makefile` - Simple include-based Makefile that includes `Makefile-pattern`
- `Makefile-pattern` - Contains all pattern targets and dynamically reads secrets config from `values-global.yaml`

When using `--with-secrets`:
- `values-secret.yaml.template` - Template for secrets configuration
- `values-global.yaml` with `global.secretLoader.disabled: false` (enables secrets)
- Additional applications (vault, golang-external-secrets) in cluster group values

The secrets loading behavior is controlled entirely by the `global.secretLoader.disabled` field in `values-global.yaml`.

---

## Container Usage

Use the prebuilt container from Quay without needing to install anything locally:

### Basic Initialization

```bash
# Navigate to your pattern repository
cd /path/to/your/pattern-repo

# Initialize without secrets
podman run --pull=always --rm -it -v .:/repo:z quay.io/dminnear/patternizer init

# Initialize with secrets support
podman run --pull=always --rm -it -v .:/repo:z quay.io/dminnear/patternizer init --with-secrets
```

---

## Example Workflow

1. **Clone or create a pattern repository:**
   ```bash
   git clone https://github.com/your-org/your-pattern.git
   cd your-pattern
   git checkout -b run-patternizer
   ```

2. **Initialize the pattern:**
   ```bash
   podman run --pull=always --rm -it -v .:/repo:z quay.io/dminnear/patternizer init
   ```

3. **Commit and push generated files:**
   ```bash
   git add .
   git commit -m 'initialize pattern using patternizer'
   git push -u origin run-patternizer
   ```

4. **Install the pattern:**
   ```bash
   ./pattern.sh make install
   ```

---

## Development

### Prerequisites

- Go 1.24+
- Podman or Docker
- Git
- Make

### Quick Start

```bash
# Set up development environment (installs dependencies and tools)
make dev-setup

# Show all available targets
make help
```

### Common Development Tasks

```bash
# Build the CLI
make build

# Run all tests (unit + integration)
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Build container image locally (as `patternizer:local`)
make local-container-build

# Run full CI pipeline locally
make ci

# Quick feedback loop (format check, vet, build, unit tests)
make check
```

### Code Quality

The project uses comprehensive linting and formatting:

```bash
# Run all linting checks (gofmt, go vet, golangci-lint)
make lint

# Format code
make fmt

# Run individual lint checks
make lint-fmt     # gofmt check
make lint-vet     # go vet
make lint-golangci # golangci-lint
```

---

## Testing

### Unit Tests

The project includes comprehensive unit tests across multiple packages:

**Main Package Tests (`src/main_test.go`):**
- `TestGetResourcePath()` - Resource path resolution with and without environment variables
- `TestNewDefaultValuesGlobal()` - Global configuration default values validation
- `TestNewDefaultValuesClusterGroup()` - Cluster group configuration generation and secrets integration

**Helm Package Tests (`src/internal/helm/helm_test.go`):**
- `TestFindTopLevelCharts()` - Helm chart discovery functionality with comprehensive test scenarios:
  - Correctly identifies valid top-level charts (chart1, chart2)
  - Properly skips sub-charts in `charts/` directories
  - Ignores hidden directories (`.hidden-chart`) and invalid chart structures
  - Handles edge cases: missing Chart.yaml, missing values.yaml, missing templates directory, templates as file
- `TestIsHelmChart()` - Chart structure validation:
  - Validates required files: Chart.yaml, values.yaml, templates/ directory
  - Tests various invalid configurations and edge cases

**Pattern Package Tests (`src/internal/pattern/pattern_test.go`):**
- `TestExtractPatternNameFromURL()` - Git URL parsing for multiple formats:
  - SSH URLs: `git@github.com:user/repo.git`, `git@gitlab.com:group/subgroup/repo.git`
  - HTTPS/HTTP URLs: `https://github.com/user/repo.git`, `http://github.com/user/repo`
  - Error handling for invalid URLs and unsupported protocols
- `TestProcessGlobalValuesPreservesFields()` - Field preservation during YAML processing:
  - Preserves existing custom fields at all nesting levels
  - Maintains custom arrays, nested objects, and primitive values
  - Intelligently merges new defaults with existing configurations
- `TestProcessClusterGroupValuesPreservesFields()` - Cluster group values field preservation:
  - Preserves custom applications, subscriptions, and cluster-level fields
  - Adds new applications while maintaining existing ones
  - Maintains custom fields within applications and subscriptions
- `TestProcessGlobalValuesWithNewFile()` - New file creation with proper defaults
- `TestProcessGlobalValuesWithSecrets()` - Validates secrets configuration:
  - Tests `ProcessGlobalValues` with `withSecrets=true`
  - Verifies `global.secretLoader.disabled: false` is set correctly
  - Ensures secrets-enabled configuration is properly generated

### Integration Tests

The integration test suite (`test/integration_test.sh`) validates the complete CLI workflow with four comprehensive test scenarios:

**Test 1: Basic Initialization (Without Secrets)**
- Clones the [trivial-pattern](https://github.com/dminnear-rh/trivial-pattern) repository
- Runs `patternizer init` and validates generated files
- Verifies `values-global.yaml` contains `global.secretLoader.disabled: true`
- Validates `values-prod.yaml` content using YAML normalization
- Ensures `pattern.sh` is created and executable
- Validates `Makefile` (include-based) and `Makefile-pattern` are created

**Test 2: Initialization with Secrets**
- Tests `patternizer init --with-secrets` functionality
- Verifies `values-global.yaml` contains `global.secretLoader.disabled: false`
- Validates secrets-specific applications (vault, golang-external-secrets) are added
- Verifies additional namespaces and `values-secret.yaml.template` are created
- Ensures `pattern.sh` and both Makefile files are properly generated

**Test 3: Custom Pattern and Cluster Group Names**
- Tests field preservation and intelligent merging of existing configurations
- Pre-populates custom `values-global.yaml` with renamed pattern and cluster group
- Verifies custom names are preserved while adding missing default configurations
- Validates custom cluster group file generation (e.g., `values-renamed-cluster-group.yaml`)
- Ensures `global.secretLoader.disabled: false` is set correctly with `--with-secrets`

**Test 4: Sequential Execution**
- Tests running `patternizer init` followed by `patternizer init --with-secrets`
- Validates that the second command properly upgrades the configuration
- Ensures `global.secretLoader.disabled` transitions from `true` to `false`
- Verifies final state matches direct `--with-secrets` execution

Run integration tests locally:
```bash
# Run integration tests (automatically builds binary first)
make test-integration

# Or run all tests (unit + integration)
make test
```

---

## CI/CD Pipeline

The project uses a comprehensive CI pipeline with three stages that leverage the Makefile for consistency:

1. **Lint & Format**: `make lint` - Code quality checks with `gofmt`, `go vet`, and `golangci-lint`
2. **Build & Test**: `make build`, `make test-unit`, `make test-coverage`, `make test-integration`
3. **Container Build**: Multi-stage container build and push to Quay.io

All code must pass linting and tests before being merged or deployed.

The CI pipeline uses the same Makefile targets that developers use locally, ensuring perfect consistency between local development and CI environments. You can run the same checks locally with `make ci`.

---

## Architecture

The CLI is organized into focused packages following Go best practices:

**Main Package (`src/`):**
- `main.go` - Application entry point

**Command Package (`src/cmd/`):**
- `root.go` - Cobra CLI setup and root command
- `init.go` - Initialization command logic and orchestration

**Internal Packages (`src/internal/`):**
- `fileutils/` - File operations, resource management, and path resolution
- `helm/` - Helm chart discovery and validation
- `pattern/` - Core pattern processing, Git operations, and URL parsing
- `types/` - YAML structure definitions and default value constructors

**Key Design Principles:**
- **Separation of Concerns**: Each package has a single, well-defined responsibility
- **Testability**: All packages are thoroughly unit tested with comprehensive coverage
- **Field Preservation**: YAML processing preserves all user-defined custom fields
- **Error Handling**: Comprehensive error handling with descriptive messages
- **Modularity**: Clean interfaces between packages for maintainability

This modular design makes the codebase maintainable, testable, and extensible.

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run the development workflow:
   ```bash
   make dev-setup  # Set up development environment
   make ci         # Lint/build/test your changes in one step
   ```
5. Submit a pull request

All contributions must pass the CI pipeline including linting, formatting, and comprehensive testing.

### Development Workflow

For the best development experience:
```bash
# Initial setup
make dev-setup

# During development (fast feedback)
make check

# Before committing
make ci  # Runs the full CI pipeline locally
```
