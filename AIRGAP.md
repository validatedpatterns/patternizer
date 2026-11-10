# Airgap / Disconnected Patterns

This page summarizes Patternizer’s disconnected (airgap) workflow. For the full walkthrough and lab notes, see the [Disconnected Patterns UX Update](https://validatedpatterns.io/blog/2026-10-26-disconnected-updates/) blog post. For general disconnected install guidance, see the [disconnected installation guide](https://validatedpatterns.io/learn/disconnected-installation/).

- [Airgap / Disconnected Patterns](#airgap--disconnected-patterns)
  - [Overview](#overview)
  - [Quick Start](#quick-start)
  - [Example Workflow](#example-workflow)
  - [Usage Details](#usage-details)
    - [Collect artifacts (`airgap init`)](#collect-artifacts-airgap-init)
    - [Move content into the airgap](#move-content-into-the-airgap)
    - [Prepare the pattern](#prepare-the-pattern)
    - [Configure secrets and bootstrap](#configure-secrets-and-bootstrap)
    - [Install](#install)
  - [Further reading](#further-reading)

## Overview

Identifying every operator, container image, and Helm chart for a pattern is tedious—especially when remote charts use wildcard versions. Patternizer’s airgap commands scrape a pattern repository for those artifacts, merge optional extras from a config file, and generate an oc-mirror v2 `imagesetconfig.yaml` plus helpers for disconnected install.

The same tooling can include operators, images, and charts for workflows that are not pattern-specific via `--extra-config`.

## Quick Start

**Prerequisites:**

- Podman or Docker (or a local Patternizer binary)
- A pattern repository (local path or Git URL)
- Access to a connected machine for mirroring, then a bastion in the airgap

```bash
# Optional shell helper (see README for the full function)
patternizer airgap init https://github.com/validatedpatterns/multicloud-gitops \
  --extra-config "$HOME/extra-images.yaml" \
  --ocp-version 4.22 \
  --output-dir "$HOME/mcgodemo"
```

## Example Workflow

Using [multicloud-gitops](https://github.com/validatedpatterns/multicloud-gitops) as the example pattern:

1. **On a connected host:** run `patternizer airgap init` to scrape the pattern, merge `--extra-config`, and write outputs under `--output-dir`.
2. **Mirror** images into your disconnected registry with oc-mirror (use generated `mirror-commands.sh`). Install OpenShift in the airgap if needed.
3. **Copy** the `--output-dir` tree to the bastion (`scp -r …`).
4. **Publish** the cloned pattern into your airgapped Git service (for example Gitea).
5. **On the bastion:** `patternizer upgrade`, configure `~/values-secret-<pattern>.yaml` with an `airgap:` section, run `patternizer airgap bootstrap`, commit/push, then `./pattern.sh make install-airgap`.

## Usage Details

### Collect artifacts (`airgap init`)

Patternizer scrapes the repository for operators, container images, and Helm charts. Path-based charts (`path: charts/...`) contribute images under `additionalImages`. Remote OCI charts (`chart:` / `chartVersion:`) are resolved (wildcards take the latest) and scanned. Git clones use full history to avoid shallow-clone issues when pushing into your Git service.

| Flag | Purpose |
|------|---------|
| `--extra-config` | YAML merge of extra operators, `additionalImages`, and `helm.local` |
| `--ocp-version` | Include OpenShift platform channel `stable-<version>` and tag OLM catalogs / CatalogSources (default catalog tag `4.22` when omitted) |
| `--output-dir` | Generated artifacts (default `./airgap-out`) |
| `--registry` | Optional mirror hostname for `values-disconnected.yaml` / `mirror-commands.sh` |
| `--download-oc-mirror` | Download oc-mirror archives into `--output-dir` |
| `--git-ref` / `--clone-dir` | Git branch/tag and clone location when source is a URL |

Example `--extra-config`:

```yaml
operators:
  - catalog: registry.redhat.io/redhat/redhat-operator-index:v4.22
    packages:
      - name: lvms-operator
additionalImages:
  - name: registry.example.com/app:1.0
```

Outputs under `--output-dir`:

- `imagesetconfig.yaml`
- `values-disconnected.yaml`
- `mirror-commands.sh`
- `CHECKLIST.md`
- cloned pattern repo (when source is a Git URL)

### Move content into the airgap

Copy `--output-dir` to the disconnected bastion, mirror images into the registry, and push the pattern Git history into your internal SCM.

### Prepare the pattern

```bash
cd ~/mcgodemo/multicloud-gitops
patternizer upgrade
```

`upgrade` refreshes `Makefile-common` and `pattern.sh` with airgap-aware targets (including `install-airgap`).

### Configure secrets and bootstrap

Copy the secrets template and add an `airgap:` section (same `path` / `value` style as Vault secrets; **not** loaded into Vault). Declare Git/registry CA paths, registry credentials (`dockerconfig` or username/password), `registry`, and `ocpVersion` (or explicit `catalogSources` / `discoverCatalogs`).

```bash
cp values-secret.yaml.template ~/values-secret-multicloud-gitops.yaml
# edit airgap: certificates, registries, registry, ocpVersion, …

patternizer airgap bootstrap /path/to/multicloud-gitops \
  --output-dir ./airgap-bootstrap \
  --discover-catalogs \
  --secrets-file ~/values-secret-multicloud-gitops.yaml
```

Bootstrap ingests that file, optionally discovers CatalogSources, and updates `values-global.yaml` and variant values (`helmRepoUrl`, operator sources, and related disconnected fields).

Lower-level helpers remain available without a values-secret file: `airgap trust-ca` and `airgap repo-creds`.

### Install

```bash
git add . && git commit -m 'patternizer airgap bootstrap' && git push
./pattern.sh make install-airgap
```

`install-airgap` applies proxy / user-CA trust, runs `make install` in the background, waits for GitOps, then applies TLS trust and registry repo-creds so Argo CD can sync—avoiding the deadlock where install never finishes until post-trust is applied.

## Further reading

<!-- - [Disconnected Patterns UX Update](https://validatedpatterns.io/blog/2026-10-26-disconnected-updates/) -->
- [Disconnected installation guide](https://validatedpatterns.io/learn/disconnected-installation/)
- [Patternizer README](./README.md) — full CLI examples for `airgap init`, `trust-ca`, `repo-creds`, and bootstrap
