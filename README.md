# Patternizer

[![Quay Repository](https://img.shields.io/badge/Quay.io-patternizer-blue?logo=quay)](https://quay.io/repository/dminnear/patternizer)
[![Container Build Status](https://github.com/dminnear-rh/patternizer/actions/workflows/push-to-quay.yaml/badge.svg?branch=main)](https://github.com/dminnear-rh/patternizer/actions/workflows/push-to-quay.yaml)

**Patternizer** is a container-based utility designed to bootstrap a new Validated Pattern repository. It automatically generates the necessary `values-global.yaml` and `values-<cluster_group>.yaml` files by inspecting the Git repository, discovering Helm charts, and applying sensible defaults.

The utility also provides a `pattern.sh` script for installing, loading secrets, and other common operations. The entire process is containerized to ensure consistency and ease of use.

---

## Quickstart Guide

You can use the prebuilt container image from Quay to initialize a new pattern repository without building anything locally.

1. Clone the Git repository you want to patternize:

   ```bash
   git clone https://github.com/dminnear-rh/trivial-pattern.git
   cd trivial-pattern
   ````

2. Run the Patternizer container to initialize the repository:

   * If **you do not need secrets support**, run:

     ```bash
     podman run --rm -it -v .:/repo:z quay.io/dminnear/patternizer
     ```

   * If **you want to enable secrets support** (Vault and External Secrets), run instead:

     ```bash
     podman run --rm -it -e USE_SECRETS=true -v .:/repo:z quay.io/dminnear/patternizer
     ```

   This will generate:

   * `values-global.yaml`
   * `values-<cluster_group>.yaml`
   * `pattern.sh` (utility script)
   * `values-secret.yaml.template` (only if `USE_SECRETS=true`)

3. Install the Pattern

   Log into the OpenShift 4 cluster where you want to install the pattern, then run:

   ```bash
   ./pattern.sh make install
   ```

   This uses the common utility container (`quay.io/dminnear/common-utility-container`) to handle shared scripts and resources, so no `common/` directory or Makefile is added directly to your repo.

---

## Development (Optional)

You only need to build the Go binary or container if you're modifying or developing Patternizer itself.

### Build the Go Binary

```bash
cd src
go build
```

### Build the Container Image

```bash
podman build -t patternizer:local .
```
