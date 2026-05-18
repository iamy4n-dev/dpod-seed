# dpod-seed — Product Requirements Document

## Problem Statement

Setting up a reproducible development environment for a project is tedious and inconsistent. Developers
manually copy devcontainer configs, dotfiles, and tool configurations across projects, leading to drift
between environments. When tooling needs to change (new tool version, different OS base, updated shell
config), there is no mechanism to propagate that change back to existing projects in a controlled,
reviewable way.

Teams that want to standardise on a shared dev environment have no single source of truth — configs
are scattered, unversioned, and impossible to verify.

## Solution

`dpod-seed` is a Go CLI that manages the lifecycle of a project's DevPod environment by pulling
verified configuration from a set of upstream repositories into the project. Developers declare what
environment they want in a `dpod.yaml` file, and the CLI materialises the correct files. A
`dpod.lock` file records exactly what was written so future syncs can overwrite, update, and remove
only what the CLI owns — never touching project files.

The upstream configuration is split across three purpose-built repositories:

- **`devcontainer/`** — base OS profiles and system tooling (package manager setup, etc.)
- **`packages/`** — dotfile bundles grouped by purpose (shell, editor, k8s, AWS, databases, etc.)
- **`distro/`** — named, verified compositions of a devcontainer profile and a set of package bundles

A distro follows the Linux distro mental model: it is a sealed preset that is verified by CI and
published under a git tag. Teams that want structural changes fork the distro repo. Small per-project
tweaks are expressed as patch files declared in `dpod.yaml`.

## User Stories

1. As a developer, I want to run `dpod-seed init` in my project and interactively pick a distro, so that my project gets a working `.devcontainer/` without me manually copying files.
2. As a developer, I want to review a summary of what `dpod-seed init` will write before it executes, so that I can confirm or cancel before any files are touched.
3. As a developer, I want to retry distro and package selection during `dpod-seed init` without restarting the command, so that I can correct mistakes without friction.
4. As a developer, I want `dpod-seed list` to show me all verified distros with their latest tag and a short description, so that I can pick the right one for my role.
5. As a developer, I want to pin my project to a specific distro tag in `dpod.yaml`, so that my environment is reproducible and does not change unexpectedly.
6. As a developer, I want to run `dpod-seed sync` to upgrade to a newer distro tag, so that I get updated tooling without manually copying files.
7. As a developer, I want files I have not declared in `dpod.yaml` to never be touched by the CLI, so that my project files are always safe.
8. As a developer, I want files removed from a distro in a newer version to be deleted from my project on sync, so that my `.devcontainer/` does not accumulate orphaned files.
9. As a developer, I want to add a small override (e.g. swap one editor package) in `dpod.yaml` using a patch file, without forking the whole distro, so that I can customise without maintaining a fork.
10. As a developer, I want `dpod-seed validate` to resolve my `dpod.yaml` and check all referenced versions exist, without writing any files, so that I can run it in CI to catch broken configs early.
11. As a developer, I want `dpod-seed eject` to remove `dpod.lock` (with a confirmation prompt) so that the CLI stops managing my files and I take full ownership, without losing any files.
12. As a distro author, I want to run `dpod-seed scaffold distro <name>` to generate the correct directory layout and a `distro.yaml` template, so that I can start a new distro without looking up the structure.
13. As a distro author, I want to run `dpod-seed scaffold package <name>` to generate the correct directory layout and a `manifest.yaml` template for a new package bundle.
14. As a distro author, I want CI to verify my distro (build, boot, tool version checks) on every PR, so that only passing distros get tagged and published.
15. As a distro author, I want to publish a new verified distro by tagging `vX.Y.Z` on the `distro/` repo, so that downstream projects can pin to it.
16. As a team lead, I want to point `dpod-seed` at a private fork of the upstream repos via `~/.dpod-seed/config.yaml`, so that my team uses our own verified distros without touching public upstream.
17. As a team lead, I want team members to be able to run `dpod-seed sync` in CI (non-interactive), so that the environment config is validated on every PR.
18. As a platform engineer, I want `dpod-seed` to be available as a devcontainer feature, so that it is pre-installed inside any devpod and available without a manual install step.
19. As a developer, I want `dpod.lock` to record the source SHA of every file the CLI materialised, so that I can audit exactly where each file came from.
20. As a developer, I want `dpod-seed sync` to print a clear summary of added, updated, and removed files after completing, so that I can review what changed.
21. As a developer, I want to install `dpod-seed` with a one-liner `curl | sh` script that downloads the correct binary for my platform from GitHub Releases, so that setup takes seconds.
22. As a developer working on a project that uses an upstream distro as-is, I want to know that running `dpod-seed sync` will cleanly overwrite distro-owned files, so that I understand the update contract before upgrading.
23. As a developer who wants structural customisation, I want to fork the `distro/` repo and point `dpod.yaml` at my fork, so that I fully control the sealed preset without depending on upstream.

## Implementation Decisions

### Three upstream repositories

- **`devcontainer/`** — one directory per base OS profile under `profiles/`. Each profile contains a `.devcontainer/` directory (devcontainer.json + Dockerfile or equivalent).
- **`packages/`** — one directory per bundle under `packages/`. Each bundle contains a `dotfiles/` directory (files to materialise) and an optional `manifest.yaml` for non-convention placement.
- **`distro/`** — one directory per distro under `distros/`. Each distro contains a `distro.yaml` manifest pinning exact tags from the other two repos.

Example `distro.yaml`:
```yaml
name: devops-k8s
description: Arch-based OS with AWS and Kubernetes tooling
devcontainer: arch-base@v2.1.0
packages:
  - shell-zsh@v1.3.0
  - k8s-tools@v1.1.0
  - aws-cli@v2.0.0
```

### Per-project files

**`dpod.yaml`** — checked into the project repo. Declares the distro + optional patch-file overrides:
```yaml
distro: devops-k8s@v1.2.0
overrides:
  packages:
    add: [neovim]
    remove: [vscode]
  patches:
    - overrides/add-custom-alias.patch
```

**`dpod.lock`** — generated by the CLI. Records every file the CLI materialised, its source repo, and its source SHA. The CLI only overwrites or deletes files listed here. Everything else is untouched.

### Registry

A `registry.yaml` file is published to a known stable URL in the `distro/` repo by CI whenever a new tag is created. `dpod-seed list` fetches only this file — no cloning, no API tokens required.

### File placement convention

Files under `dotfiles/` in a package bundle map to `.devcontainer/` in the project, preserving relative path. A `manifest.yaml` in the bundle overrides destination for specific files. This convention is the default; the manifest is only needed for exceptions.

### Sync behaviour

1. Resolve `dpod.yaml` → full file manifest via `distro.yaml` pin chain
2. Diff against `dpod.lock` → sets of (add, update, remove)
3. Fetch required files from upstream at pinned SHAs
4. Materialise files to project per placement convention
5. Apply patch files from `overrides.patches`
6. Delete files in the remove set
7. Write updated `dpod.lock`
8. Print summary (N added, N updated, N removed)

### Eject behaviour

Deletes `dpod.lock` only. `dpod.yaml` is preserved as a record of origin. A confirmation prompt is shown before deletion: `"This will remove CLI ownership of all managed files. Files themselves are untouched. Continue? [y/N]"`. Irreversible.

### CLI command surface

| Command | Behaviour |
|---|---|
| `dpod-seed init` | Interactive: list distros → pick → review summary → confirm/retry/cancel → materialise |
| `dpod-seed sync` | Re-materialise from `dpod.yaml`, overwrite owned files, delete dropped files |
| `dpod-seed list` | Fetch `registry.yaml`, print distro names, descriptions, latest tags |
| `dpod-seed validate` | Resolve `dpod.yaml` without writing anything; exits non-zero if unresolvable |
| `dpod-seed eject` | Delete `dpod.lock` with confirmation; files stay, CLI no longer manages them |
| `dpod-seed scaffold distro <name>` | Generate distro directory layout + `distro.yaml` template |
| `dpod-seed scaffold package <name>` | Generate package directory layout + `manifest.yaml` template |

### Modules

**Resolver** — pure function. Input: distro name + tag + registry. Output: fully resolved file manifest (list of source repo, path, SHA for every file to materialise). No I/O. All network calls happen before and after this step.

**LockManager** — pure function. Input: current `dpod.lock` state + new resolved manifest. Output: diff (added files, updated files, removed files). No I/O.

**Fetcher** — downloads a file tree from a GitHub repo at a specific SHA and path. Interface: `Fetch(repo, sha, path) → []File`. Self-hosted repos configurable via `~/.dpod-seed/config.yaml` or env vars.

**PatchApplier** — applies unified diff patch files to in-memory file content. Input: file content + patch file content. Output: patched content. No filesystem access — called by Materializer.

**Materializer** — writes resolved files to disk using the placement convention, calls PatchApplier for overrides, delegates the remove set to file deletion. Owns all filesystem writes.

**RegistryClient** — thin HTTP client. Fetches `registry.yaml` from a configured URL and parses it.

**ConfigParser** — reads and writes `dpod.yaml`. Thin YAML wrapper.

**CLI** — command routing and interactive prompts (numbered lists, review/confirm/retry/cancel flow).

### Distribution

- Binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 published to GitHub Releases by CI on tag
- Install script: `curl -fsSL .../install.sh | sh`
- Devcontainer feature published to the `devcontainer/` repo so `dpod-seed` is pre-installed in any managed devpod

### Upstream access

Public GitHub repos by default. Override via `~/.dpod-seed/config.yaml`:
```yaml
registryUrl: https://raw.githubusercontent.com/org/distro/main/registry.yaml
repos:
  devcontainer: https://github.com/org/devcontainer
  packages: https://github.com/org/packages
  distro: https://github.com/org/distro
```

Private repos use the user's existing git credentials — no special auth built into the CLI.

## Testing Decisions

A good test exercises observable external behaviour only — what files were written, what the lock contains, what errors were returned — not internal call sequences or intermediate state.

### Resolver
- Given a valid `distro.yaml` pinning known component tags, returns the correct file manifest
- Given an unknown distro name, returns a descriptive error
- Given a tag that does not exist in the upstream repo, returns a descriptive error
- Overrides (add/remove packages) are reflected correctly in the resolved manifest

### LockManager
- Diff against an empty lock returns all files as "added"
- Diff where a file's SHA changed returns it as "updated"
- Files in the old lock absent from the new manifest are returned as "removed"
- Files unchanged between lock and manifest are returned in neither set

### PatchApplier
- A valid unified diff is applied correctly to the base content
- An invalid patch returns a descriptive error
- A patch that does not apply cleanly (context mismatch) returns a descriptive error

### Fetcher
- Correct repo, SHA, and path returns expected file tree
- Non-existent SHA returns a descriptive error
- Network failure returns a wrapped error

### Materializer
- Files are written to the correct paths per the placement convention
- `manifest.yaml` overrides produce the correct destination paths
- Files in the remove set are deleted and not present after materialisation
- Patch overrides are applied before files are written to disk

## Out of Scope

- **Monorepo support** — multiple `dpod.yaml` files in subdirectories pointing at different distros. Deferred to v2.
- **3-way merge on sync** — files are overwritten. Users who want to preserve customisation fork the distro or use `overrides.patches`.
- **Windows support** — initial release targets Linux and macOS only.
- **Built-in dotfiles management** — personal dotfiles (outside the devcontainer context) are handled by DevPod's native dotfiles feature, not `dpod-seed`.
- **Provider configuration** — DevPod provider configs are machine-specific and out of scope.
- **Interactive TUI** — init uses simple numbered prompts, not an arrow-key TUI.
- **GUI or web interface** — CLI only.

## Further Notes

- The three upstream repos (`devcontainer/`, `packages/`, `distro/`) are separate repositories. `dpod-seed` itself lives in a fourth repo (`dpod-seed/`) containing the CLI source.
- The project follows the Linux distro mental model deliberately: upstream is authoritative, divergence is handled by forking, and the upgrade contract is explicit and clean.
- The `dpod.lock` file should be committed to the project repo so that the exact materialised state is reproducible by any team member or CI run.
- `dpod-seed validate` is the recommended CI gate: run it on every PR to catch broken distro references before they land.
- When running non-interactively (no TTY detected), `init` should fail fast with a clear error directing the user to populate `dpod.yaml` manually, rather than hanging on prompts.
