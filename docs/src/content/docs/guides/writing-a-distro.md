---
title: Writing a distro
description: How to author, test, and contribute a new distro to the upstream registry.
draft: false
---

A distro is a named, versioned composition of a devcontainer profile and a set of package bundles. Distros live in the [`distros` repository](https://github.com/duyanh-y4n/distros).

## Directory layout

```
distros/
  <name>/
    distro.yaml   ← composition spec
    README.md     ← human-readable description
registry.yaml     ← auto-generated on tag push — do not edit by hand
```

## 1. Scaffold a new distro

From inside a clone of the `distros` repo, use the `scaffold` command:

```sh
dpod-seed scaffold distro <name>
```

This creates:

```
distros/<name>/
  distro.yaml   ← template with placeholder values
  README.md     ← stub
```

## 2. Edit distro.yaml

Fill in the `distro.yaml` with real pins:

```yaml
name: devops-k8s
description: Kubernetes development environment with kubectl, helm, and k9s
devcontainer: arch-base@v2.0.0
packages:
  - shell-zsh@v1.3.0
  - k8s-tools@v1.1.0
```

See the [distro.yaml reference](/dpod-seed/reference/distro-yaml/) for the full schema.

:::caution
All `@version` pins must reference tags that exist in the upstream `devcontainer` and `packages` repos. A pull request with unresolvable pins will fail CI.
:::

## 3. Write README.md

Add `distros/<name>/README.md` describing:

- What this distro is for
- Which tools are included
- Any notable configuration choices

## 4. Open a pull request

Push your branch and open a PR against the `distros` repo. CI will:

- Run `go test ./...` on the registry generator
- Validate all `distro.yaml` files are well-formed
- Dry-run the registry generator to confirm output is valid

## 5. Release

After the PR is merged, a maintainer pushes a version tag:

```sh
git tag v0.2.0 && git push origin v0.2.0
```

CI automatically regenerates `registry.yaml` and commits it to `main`. The new distro is immediately available to all `dpod-seed` users.

## Package bundles and devcontainer profiles

A distro references:

- **devcontainer profile** — from `github.com/duyanh-y4n/devcontainer`, e.g. `arch-base@v2.0.0`
- **package bundles** — from `github.com/duyanh-y4n/packages`, e.g. `k8s-tools@v1.1.0`

Each bundle contributes files under `.devcontainer/`. Pinning by tag ensures the composition is reproducible and auditable.
