# Patternizer

[![Quay Repository](https://img.shields.io/badge/Quay.io-patternizer-blue?logo=quay)](https://quay.io/repository/dminnear/patternizer)
[![CI Pipeline](https://github.com/dminnear-rh/patternizer/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/dminnear-rh/patternizer/actions/workflows/ci.yaml)

**Patternizer** is a CLI tool and container utility designed to bootstrap Validated Pattern repositories. It automatically generates the necessary `values-global.yaml` and `values-<cluster_group>.yaml` files by inspecting Git repositories, discovering Helm charts, and applying sensible defaults.

The tool provides both a standalone CLI and containerized execution for maximum flexibility and consistency across environments.

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

# Build and test
make build
make test
```

---

## CLI Usage

### Available Commands

```bash
# Show help and available commands
patternizer help

# Initialize pattern files (without secrets)
patternizer init

# Initialize pattern files with secrets support
patternizer init --with-secrets

# Show help for the init command
patternizer init help
```

### Output Files

The `patternizer init` command generates:

- `values-global.yaml` - Global pattern configuration
- `values-<cluster_group>.yaml` - Cluster group-specific configuration
- `pattern.sh` - Utility script for pattern operations

When using `--with-secrets`:
- `values-secret.yaml.template` - Template for secrets configuration
- Modified `pattern.sh` with `USE_SECRETS=true` as default

---

## Container Usage

Use the prebuilt container from Quay without needing to install anything locally:

### Basic Initialization

```bash
# Navigate to your pattern repository
cd /path/to/your/pattern-repo

# Initialize without secrets
podman run --rm -it -v .:/repo:z quay.io/dminnear/patternizer init

# Initialize with secrets support
podman run --rm -it -v .:/repo:z quay.io/dminnear/patternizer init --with-secrets
```

---

## Example Workflow

1. **Clone or create a pattern repository:**
   ```bash
   git clone https://github.com/your-org/your-pattern.git
   cd your-pattern
   ```

2. **Initialize the pattern:**
   ```bash
   podman run --rm -it -v .:/repo:z quay.io/dminnear/patternizer init
   ```

3. **Review generated files:**
   ```bash
   ls -la values-*.yaml pattern.sh
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

# Build container image locally
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

The project includes comprehensive unit tests located in `src/*_test.go` and `src/internal/*/` packages:

**Core Functionality Tests:**
- Resource path resolution and environment variable handling
- Default values generation for global and cluster group configurations
- Secrets integration logic and template handling

**Helm Chart Discovery Tests:**
- `FindTopLevelCharts()` correctly identifies top-level Helm charts
- Properly skips sub-charts, hidden directories, and invalid chart structures
- `IsHelmChart()` validates chart structure (Chart.yaml, values.yaml, templates/)

**Pattern Processing Tests:**
- URL parsing for SSH, HTTPS, and HTTP Git repository formats
- Error handling for invalid or unsupported URL formats
- Field preservation during YAML processing (custom user fields are never overwritten)
- Proper merging of defaults with existing configuration files

**Field Preservation Verification:**
- Tests ensure that custom fields in `values-global.yaml` are preserved
- Tests verify that custom fields in cluster group values files are maintained
- Tests confirm that nested custom fields and arrays are properly handled

### Integration Tests

The integration test (`test/integration_test.sh`) validates the complete workflow:
- Clones the [trivial-pattern](https://github.com/dminnear-rh/trivial-pattern) repository
- Runs `patternizer init`
- Verifies generated files match expected output
- Ensures `pattern.sh` is created with correct configuration

Run integration tests locally:
```bash
# Run integration tests (automatically builds binary first)
make test-integration

# Or run all tests (unit + integration)
make test
```

---

## CI/CD Pipeline

The project uses a comprehensive CI pipeline with three stages:

1. **Lint & Format**: Code quality checks with `gofmt`, `go vet`, and `golangci-lint`
2. **Build & Test**: Binary compilation, unit tests, and integration tests
3. **Container Build**: Multi-stage container build and push to Quay.io

All code must pass linting and tests before being merged or deployed.

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
   make check      # Quick feedback loop
   make test       # Run all tests
   make lint       # Run all linting checks
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
