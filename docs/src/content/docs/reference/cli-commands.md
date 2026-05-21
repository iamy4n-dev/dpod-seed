---
title: CLI commands
description: All dpod-seed commands, flags, and exit codes.
draft: false
---

## dpod-seed init

Interactively select a distro and materialise DevPod config into the current project.

```sh
dpod-seed init
```

Requires an interactive terminal (TTY). In non-interactive environments, populate `dpod.yaml` manually and run `dpod-seed sync`.

**Behaviour:**
- If `dpod.yaml` already exists, prompts to overwrite before continuing.
- Displays a numbered list of available distros from the registry.
- Shows a preview of files that will be written before asking for confirmation.
- On confirmation, writes `dpod.yaml` and runs `sync`.

**Prompts:**
| Prompt | Values |
|--------|--------|
| Select distro | Number 1–N |
| Confirm write | `y` / `r` (retry) / `c` (cancel) |

---

## dpod-seed sync

Re-materialise DevPod config from `dpod.yaml`, overwriting only CLI-owned files.

```sh
dpod-seed sync
```

**Behaviour:**
- Reads `dpod.yaml` and resolves the distro at the pinned tag.
- Diffs the resolved manifest against `dpod.lock`.
- Writes only added or updated files; removes files that are no longer in the distro.
- Updates `dpod.lock`.

**Output:**
```
<N> added, <N> updated, <N> removed
```

Idempotent — running twice with no changes produces `0 added, 0 updated, 0 removed`.

---

## dpod-seed validate

Resolve `dpod.yaml` without writing any files. Suitable for CI.

```sh
dpod-seed validate
```

**Behaviour:**
- Reads `dpod.yaml` and resolves the distro at the pinned tag via the GitHub API.
- Exits 0 if the pin is resolvable; non-zero otherwise.

**Output on success:**
```
OK — resolved <N> files
```

---

## dpod-seed list

List all verified distros from the upstream registry.

```sh
dpod-seed list [--registry <url>]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--registry` | `https://raw.githubusercontent.com/iamy4n-dev/distros/main/registry.yaml` | URL of the `registry.yaml` file |

**Output:**
```
NAME         DESCRIPTION                   LATEST TAG
devops-k8s   Kubernetes development setup  v1.2.0
```

---

## dpod-seed eject

Remove `dpod.lock`, transferring file ownership back to the project.

```sh
dpod-seed eject
```

**Behaviour:**
- Asks for confirmation before proceeding.
- Removes `dpod.lock`. Materialised files are left untouched.
- After ejecting, `dpod-seed sync` and `dpod-seed validate` will no longer work until a new `dpod.yaml` is created.

---

## dpod-seed scaffold distro \<name\>

Generate a new distro directory layout with a `distro.yaml` template.

```sh
dpod-seed scaffold distro <name>
```

Creates:
```
distros/<name>/
  distro.yaml
  README.md
```

Fails if `distros/<name>` already exists.

---

## dpod-seed scaffold package \<name\>

Generate a new package bundle directory layout with a `manifest.yaml` template.

```sh
dpod-seed scaffold package <name>
```

Creates:
```
packages/<name>/
  manifest.yaml
  dotfiles/
```

Fails if `packages/<name>` already exists.
