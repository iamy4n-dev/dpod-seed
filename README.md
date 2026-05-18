# dpod-seed

CLI for managing reproducible DevPod environments from verified upstream distros.

**[Documentation](https://duyanh-y4n.github.io/dpod-seed)**

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/duyanh-y4n/dpod-seed/master/install.sh | sh
```

## Quick start

```sh
dpod-seed list    # browse available distros
dpod-seed init    # bootstrap a project interactively
dpod-seed sync    # re-sync after editing dpod.yaml
dpod-seed validate  # verify in CI
```

Binaries for `linux/amd64`, `linux/arm64`, `darwin/amd64`, and `darwin/arm64` are attached to every [GitHub Release](https://github.com/duyanh-y4n/dpod-seed/releases).
