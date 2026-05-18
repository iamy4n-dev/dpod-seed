---
title: dpod.lock
description: Schema reference for dpod.lock — the record of every file dpod-seed manages.
draft: false
---

`dpod.lock` records the source repository and exact SHA of every file that `dpod-seed` materialised into your project. It lives at the root of your project alongside `dpod.yaml`.

**Commit `dpod.lock` to version control.** It makes your environment fully reproducible and auditable — you can always trace any devcontainer file back to the exact upstream commit that produced it.

**Do not edit `dpod.lock` by hand.** It is written and updated exclusively by `dpod-seed sync`.

## Schema

```yaml
files:
  - path: <string>
    repo: <string>
    sha: <string>
```

## Fields

### `files`

A list of file records, one per file that `dpod-seed` currently owns.

Each record has:

| Field | Description |
|-------|-------------|
| `path` | Destination path of the file relative to the project root |
| `repo` | Source repository (e.g. `github.com/duyanh-y4n/devcontainer`) |
| `sha` | Git commit SHA the file was fetched from |

## Example

```yaml
files:
  - path: .devcontainer/devcontainer.json
    repo: github.com/duyanh-y4n/devcontainer
    sha: a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
  - path: .devcontainer/init.sh
    repo: github.com/duyanh-y4n/packages
    sha: 9f8e7d6c5b4a9f8e7d6c5b4a9f8e7d6c5b4a9f8e
```

## Lifecycle

| Command | Effect on dpod.lock |
|---------|---------------------|
| `dpod-seed init` | Created (via `sync`) |
| `dpod-seed sync` | Updated to reflect current materialised state |
| `dpod-seed validate` | Read-only — not modified |
| `dpod-seed eject` | Deleted |

## Ownership model

Files listed in `dpod.lock` are owned by `dpod-seed`. Running `sync` may overwrite them. If you need to customise an owned file, use [overrides.patches](/dpod-seed/reference/dpod-yaml/#overridespatch) in `dpod.yaml` rather than editing the file directly — otherwise `sync` will revert your changes.

After `dpod-seed eject`, `dpod.lock` is removed and the files become yours to edit freely.
