---
title: Creating your own distro
description: End-to-end guide for creating a personal or team-namespaced distro outside the upstream registry.
draft: false
---

This guide covers creating a distro that lives in your own forks — not contributed upstream. It fills the gap between [Writing a distro](/dpod-seed/guides/writing-a-distro/) (upstream contribution) and [Self-hosting](/dpod-seed/guides/self-hosting/) (infrastructure setup).

## 1. Naming convention

Personal and team distros use a `namespace/role-flavor` convention:

```
y4n/platform-eng-base
y4n/my-own-base
acme/frontend-node
```

The namespace is your GitHub username or organisation. The remainder is free-form but should describe the role or tool set the distro targets.

## 2. Fork the upstream repos

You need your own forks of the three upstream repos. Fork them once; all your distros and packages live inside them.

| Repo to fork | Your fork URL (example) |
|---|---|
| `iamy4n-dev/distros` | `github.com/y4n/distros` |
| `iamy4n-dev/devcontainer` | `github.com/y4n/devcontainer` |
| `iamy4n-dev/packages` | `github.com/y4n/packages` |

See [Self-hosting](/dpod-seed/guides/self-hosting/) for how to point `dpod-seed` at these forks via `~/.dpod-seed/config.yaml`.

## 3. Scaffold the distro

Clone your fork of `distros` and run:

```sh
dpod-seed scaffold distro y4n/my-own-base
```

This creates:

```
distros/y4n/my-own-base/
  distro.yaml   ← composition spec
  README.md     ← frontmatter + stub body
```

If you also need a new package bundle, clone your fork of `packages` and scaffold it there:

```sh
dpod-seed scaffold package my-tools
```

```
packages/my-tools/
  dotfiles/       ← files placed under .devcontainer/ by default
  manifest.yaml   ← placement overrides (optional)
  README.md       ← frontmatter + stub body
```

## 4. Fill in the README frontmatter

The `README.md` frontmatter is parsed by the registry UI and `dpod-seed list`. Fill it in before publishing.

**Distro README** (`distros/y4n/my-own-base/README.md`):

```yaml
---
name: "y4n/my-own-base"
display_name: "y4n · My Own Base"
description: "Personal base environment with Zsh, Neovim, and cloud CLIs."
status: stable
devcontainer: arch-base@v2.0.0
tags: [personal, zsh, neovim, cloud]
packages:
  - shell-zsh@v1.3.0
  - cloud-tools@v1.0.0
---
```

**Package README** (`packages/my-tools/README.md`):

```yaml
---
name: "my-tools"
display_name: "My Tools"
description: "Personal CLI utilities: fzf, ripgrep, jq."
status: stable
tags: [cli, utilities]
tools:
  - fzf
  - ripgrep
  - jq
---
```

Valid `status` values used in the upstream registry are `stable` and `experimental`. Use `experimental` while iterating, flip to `stable` when the distro is settled.

## 5. Fill in distro.yaml

Edit `distros/y4n/my-own-base/distro.yaml` to pin real versions:

```yaml
name: y4n/my-own-base
description: Personal base environment with Zsh, Neovim, and cloud CLIs.
devcontainer: arch-base@v2.0.0
packages:
  - shell-zsh@v1.3.0
  - cloud-tools@v1.0.0
```

All `@version` pins must reference tags that exist in your `devcontainer` and `packages` forks. See the [distro.yaml reference](/dpod-seed/reference/distro-yaml/) for the full schema.

## 6. Point your project at your personal registry

In `~/.dpod-seed/config.yaml` on every machine where you run `dpod-seed`:

```yaml
registryURL: https://raw.githubusercontent.com/y4n/distros/main/registry.yaml
repos:
  distro: github.com/y4n/distros
  devcontainer: github.com/y4n/devcontainer
  packages: github.com/y4n/packages
```

Then add this to your project's `dpod.yaml`:

```yaml
distro: y4n/my-own-base@v0.1.0
```

`dpod-seed` will resolve `y4n/my-own-base@v0.1.0` from your personal registry instead of the public one.

## 7. Tag and publish

Once you're happy with the distro, push a version tag in your `distros` fork:

```sh
git tag v0.1.0 && git push origin v0.1.0
```

If your fork has the same CI workflow as the upstream, it will regenerate `registry.yaml` automatically on tag push. If not, regenerate it manually and commit:

```sh
dpod-seed registry generate   # if the generate command is available in your build
git add registry.yaml && git commit -m "chore: regenerate registry for v0.1.0"
git push origin main
```

After `registry.yaml` is updated, `dpod-seed list` (with your personal `registryURL` configured) will show your distro:

```
NAME                DESCRIPTION                                      LATEST TAG  STATUS
y4n/my-own-base     Personal base environment with Zsh, Neovim...   v0.1.0      stable
```
