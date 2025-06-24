# Patternizer

[![Quay Repository](https://img.shields.io/badge/Quay.io-patternizer-blue?logo=quay)](https://quay.io/repository/dminnear/patternizer)
[![Container Build Status](https://github.com/dminnear-rh/patternizer/actions/workflows/push-to-quay.yaml/badge.svg?branch=main)](https://github.com/dminnear-rh/patternizer/actions/workflows/push-to-quay.yaml)

Patternizer is a container-based utility designed to bootstrap a new Validated Pattern repository. It automatically generates the necessary `values-global.yaml` and `values-<cluster_group>.yaml` files by inspecting the Git repository, discovering Helm charts, and applying sensible defaults.

The entire process is wrapped in a container for simplicity and consistency, ensuring that anyone can generate a new pattern with a single command.

---
## Prerequisites

Before you begin, ensure you have the following tools installed on your system:

* **Git**: For version control.
* **Go**: (Version 1.24+) Required only for building the `patternizer` binary from source.
* **Podman** (or Docker): For building and running the container.

---
## Quickstart Guide

Follow these steps to set up a new pattern repository.

### Step 1: Build the Go Binary

The `patternizer` utility is a Go application. First, you need to compile it from the source located in the `src` directory.

```bash
cd src
go build
```

This will create an executable file named `patternizer` inside the `src` directory.

### Step 2: Build the Container Image

Next, build the container image. The `Conatinerfile` in the root of the repository packages the `patternizer` binary and all necessary dependencies.

From the **root directory** of the `patternizer` repository, run:

```bash
podman build -t patternizer:local .
```

### Step 3: Run the Patternizer

Now you can use the container to initialize any Git repository as a Validated Pattern.

1.  Clone the Git repository you want to turn into a pattern.
2.  Navigate into that repository's directory.
3.  Run the container, mounting your repository as a volume.

```bash
# git clone https://github.com/dminnear-rh/trivial-pattern.git
# cd /path/to/repo/trivial-pattern
podman run --rm -it -v .:/repo:z patternizer:local
```

---
## How It Works

When you execute the `podman run` command, the container performs the following actions on the repository mounted at `/repo`:

1.  **Runs Patternizer**: Executes the Go binary to create or update:
    * `values-global.yaml`: The pattern name is automatically detected from your Git remote URL.
    * `values-hub.yaml` (or similar): This file is populated with default applications and any top-level Helm charts found in your repository.
2.  **Copies Common Files**: The entire `validatedpatterns/common` repository is copied into a `common/` subfolder for access to shared scripts and resources.
3.  **Adds Templates**: A `Makefile` and `values-secret.yaml.template` are added to your repository root.
4.  **Creates Utility Script**: A symbolic link named `pattern.sh` is created in your repository root, pointing to `common/scripts/pattern-util.sh` for easy access.

After the container finishes, your local repository will be fully initialized and ready for you to customize and push to Git.
