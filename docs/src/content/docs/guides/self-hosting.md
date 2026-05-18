---
title: Self-hosting
description: Run dpod-seed against your own private forks of the upstream repositories.
draft: false
---

By default `dpod-seed` fetches distros from the public `github.com/duyanh-y4n/distros` registry. If your organisation needs private distros, you can point `dpod-seed` at private forks of all upstream repos.

## Host config file

Create `~/.dpod-seed/config.yaml` on any machine where `dpod-seed` runs:

```yaml
registryURL: https://raw.githubusercontent.com/your-org/distros/main/registry.yaml
repos:
  distro: github.com/your-org/distros
  devcontainer: github.com/your-org/devcontainer
  packages: github.com/your-org/packages
```

All fields are optional — omit any field to fall back to the public default.

| Field | Default |
|-------|---------|
| `registryURL` | `https://raw.githubusercontent.com/duyanh-y4n/distros/main/registry.yaml` |
| `repos.distro` | `github.com/duyanh-y4n/distros` |
| `repos.devcontainer` | `github.com/duyanh-y4n/devcontainer` |
| `repos.packages` | `github.com/duyanh-y4n/packages` |

## What to fork

To fully self-host, fork these three repos:

| Repo | Purpose |
|------|---------|
| `duyanh-y4n/distros` | Registry of named distros; hosts `registry.yaml` |
| `duyanh-y4n/devcontainer` | Devcontainer profiles referenced by distros |
| `duyanh-y4n/packages` | Package bundles referenced by distros |

Each fork works identically to the upstream — add your own distros, devcontainer profiles, and packages following the same contribution workflow.

## Registry URL

`registryURL` must point to the raw content of a valid `registry.yaml` file. GitHub raw URLs follow the pattern:

```
https://raw.githubusercontent.com/<org>/<repo>/<branch>/registry.yaml
```

## CI and devcontainers

If your CI runners or devcontainer builds need `dpod-seed` to use private repos, distribute `~/.dpod-seed/config.yaml` as part of your base image or via a secrets manager. The file contains no credentials — authentication is handled by the GitHub API token in the environment (`GITHUB_TOKEN`).
