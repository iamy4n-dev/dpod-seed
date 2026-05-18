---
title: dpod.yaml
description: Schema reference for the dpod.yaml project configuration file.
draft: false
---

`dpod.yaml` is the project-level configuration file that pins your project to a specific distro version. It lives at the root of your project and should be committed to version control.

## Schema

```yaml
distro: <name>@<tag>

overrides:
  patches:
    - path: <dest-path>
      patch: <patch-file>
```

## Fields

### `distro`

**Required.** Pins the project to a named distro at a specific tag.

```yaml
distro: devops-k8s@v1.2.0
```

Format: `<name>@<tag>` where:
- `<name>` is a distro name from the registry (`dpod-seed list`)
- `<tag>` is a version tag from the `distros` repo (e.g. `v1.2.0`)

To upgrade, edit the tag and run `dpod-seed sync`.

### `overrides.patches`

**Optional.** A list of patch files to apply on top of the materialised distro files.

```yaml
overrides:
  patches:
    - path: .devcontainer/devcontainer.json
      patch: patches/devcontainer.patch
```

Each entry has:

| Field | Description |
|-------|-------------|
| `path` | Destination path of the file to patch (relative to project root) |
| `patch` | Path to a unified diff patch file (relative to project root) |

Patches are applied after `sync` writes the distro files. This lets you customise distro-owned files without forking the distro.

## Example

```yaml
distro: devops-k8s@v1.2.0

overrides:
  patches:
    - path: .devcontainer/devcontainer.json
      patch: patches/devcontainer.patch
```

## Workflow

1. `dpod-seed init` writes `dpod.yaml` for you.
2. Edit `distro: name@tag` to upgrade or switch distros.
3. Run `dpod-seed sync` to apply the change.
4. Commit `dpod.yaml` and `dpod.lock` together.
