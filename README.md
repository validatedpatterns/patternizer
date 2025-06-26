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

### Building the CLI

```bash
cd src
go build -o patternizer .
```

### Running Tests

```bash
# Run unit tests
cd src
go test -v ./...

# Run integration tests (requires built binary)
./test/integration_test.sh
```

### Building the Container

```bash
# Build with default settings
podman build -t patternizer:local .

# Build with custom alpine version
podman build --build-arg ALPINE_VERSION=3.22 -t patternizer:local .
```

### Code Quality

The project uses comprehensive linting and formatting:

```bash
cd src

# Format code
gofmt -s -w .

# Run linter
golangci-lint run

# Run vet
go vet ./...
```

---

## Testing

### Unit Tests

Located in `src/*_test.go`, these test core functionality:
- Resource path resolution
- Default values generation
- Secrets integration logic

### Integration Tests

The integration test (`test/integration_test.sh`) validates the complete workflow:
- Clones the [trivial-pattern](https://github.com/dminnear-rh/trivial-pattern) repository
- Runs `patternizer init`
- Verifies generated files match expected output
- Ensures `pattern.sh` is created with correct configuration

Run integration tests locally:
```bash
# Build the binary first
cd src && go build -o patternizer .

# Run integration tests
cd .. && ./test/integration_test.sh
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

The CLI is organized into focused modules:

- `main.go` - CLI setup and command definitions (Cobra)
- `commands.go` - Command logic and orchestration
- `fileutils.go` - File operations and resource management
- `pattern.go` - Core pattern processing and Git operations
- `helm.go` - Helm chart discovery
- `values_*.go` - YAML structure definitions

This modular design makes the codebase maintainable and testable.

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./... && ./test/integration_test.sh`
5. Submit a pull request

All contributions must pass the CI pipeline including linting, formatting, and comprehensive testing.
