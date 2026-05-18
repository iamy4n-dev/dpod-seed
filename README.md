# dpod-seed

CLI for managing reproducible DevPod environments from verified upstream distros.

You declare which environment you want in `dpod.yaml`; `dpod-seed` materialises the correct devcontainer files, records what it wrote in `dpod.lock`, and keeps everything in sync as the distro evolves.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/duyanh-y4n/dpod-seed/master/install.sh | sh
```

To install a specific version:

```sh
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/duyanh-y4n/dpod-seed/master/install.sh | sh
```

Binaries for `linux/amd64`, `linux/arm64`, `darwin/amd64`, and `darwin/arm64` are attached to every [GitHub Release](https://github.com/duyanh-y4n/dpod-seed/releases).

## Quick start

```sh
# 1. Browse available distros
dpod-seed list

# 2. Bootstrap a project interactively
cd my-project
dpod-seed init

# 3. Re-sync after editing dpod.yaml (e.g. upgrading the tag)
dpod-seed sync

# 4. Validate in CI (non-interactive, no writes)
dpod-seed validate
```

## Commands

| Command | Description |
|---------|-------------|
| `dpod-seed init` | Interactively select a distro and materialise files into the project |
| `dpod-seed sync` | Re-materialise from `dpod.yaml`, overwriting only CLI-owned files |
| `dpod-seed list` | List all verified distros from the upstream registry |
| `dpod-seed validate` | Resolve `dpod.yaml` and verify all referenced versions exist |
| `dpod-seed eject` | Remove `dpod.lock` and hand ownership of files back to you |
| `dpod-seed scaffold distro <name>` | Generate a new distro layout + `distro.yaml` template |
| `dpod-seed scaffold package <name>` | Generate a new package bundle layout + `manifest.yaml` template |

## dpod.yaml

```yaml
distro: devops-k8s@v1.2.0

# Optional: patch files to apply on top of the distro
overrides:
  patches:
    - path: .devcontainer/devcontainer.json
      patch: patches/devcontainer.patch
```

The `distro` field pins the project to a specific tag. Run `dpod-seed sync` after bumping the tag to apply the update.

## dpod.lock

`dpod.lock` records the source repo and SHA of every file the CLI materialised. Commit it alongside `dpod.yaml` so your environment is fully reproducible and auditable. Do not edit it by hand.

## Self-hosting

Point `dpod-seed` at a private fork of the upstream repos by creating `~/.dpod-seed/config.yaml`:

```yaml
registryURL: https://raw.githubusercontent.com/your-org/distros/main/registry.yaml
repos:
  distro: github.com/your-org/distros
  devcontainer: github.com/your-org/devcontainer
  packages: github.com/your-org/packages
```

## Devcontainer feature

`dpod-seed` is available as a devcontainer feature so it is pre-installed inside any devpod without a manual install step:

```jsonc
// .devcontainer/devcontainer.json
{
  "features": {
    "ghcr.io/duyanh-y4n/devcontainer/dpod-seed:1": {}
  }
}
```

Pin a specific version:

```jsonc
{
  "features": {
    "ghcr.io/duyanh-y4n/devcontainer/dpod-seed:1": { "version": "v0.1.0" }
  }
}
```
