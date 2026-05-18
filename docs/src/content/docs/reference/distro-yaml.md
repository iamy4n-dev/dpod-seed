---
title: distro.yaml
description: Schema reference for distro.yaml — the composition spec for a dpod-seed distro.
draft: false
---

`distro.yaml` lives inside the [`distros` repository](https://github.com/duyanh-y4n/distros) at `distros/<name>/distro.yaml`. It defines what a distro includes and is the source of truth for the registry.

## Schema

```yaml
name: <string>
description: <string>
devcontainer: <profile>@<tag>
packages:
  - <bundle>@<tag>
```

## Fields

### `name`

**Required.** The canonical identifier for the distro. Must match the parent directory name.

```yaml
name: devops-k8s
```

Used as the key in `registry.yaml` and in `dpod.yaml` pins (`devops-k8s@v1.2.0`).

### `description`

**Required.** Short human-readable description shown in `dpod-seed list` output.

```yaml
description: Kubernetes development environment with kubectl, helm, and k9s
```

### `devcontainer`

**Required.** The devcontainer profile to use, pinned to a specific tag.

```yaml
devcontainer: arch-base@v2.0.0
```

Format: `<profile>@<tag>` where the profile name and tag reference a file in `github.com/duyanh-y4n/devcontainer`.

### `packages`

**Optional.** A list of package bundles to include, each pinned to a specific tag.

```yaml
packages:
  - shell-zsh@v1.3.0
  - k8s-tools@v1.1.0
```

Format: `<bundle>@<tag>` where the bundle name and tag reference a file in `github.com/duyanh-y4n/packages`.

## Example

```yaml
name: devops-k8s
description: Kubernetes development environment with kubectl, helm, and k9s
devcontainer: arch-base@v2.0.0
packages:
  - shell-zsh@v1.3.0
  - k8s-tools@v1.1.0
```

## Validation

CI validates every `distro.yaml` on pull request. Common errors:

| Error | Cause |
|-------|-------|
| Missing `name` | `name` field is empty or absent |
| Unresolvable pin | Referenced tag does not exist in the devcontainer or packages repo |
