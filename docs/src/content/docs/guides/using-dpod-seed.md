---
title: Using dpod-seed
description: The full init → sync → validate → eject workflow for managing your DevPod environment.
draft: false
---

This guide walks through the complete lifecycle of a `dpod-seed`-managed project.

## 1. Browse available distros

Before initialising, see what distros are available in the upstream registry:

```sh
dpod-seed list
```

Output:

```
NAME         DESCRIPTION                   LATEST TAG
example      Minimal placeholder distro    v0.1.0
devops-k8s   Kubernetes development setup  v1.2.0
```

## 2. Initialise a project

Run `init` from your project root. It requires an interactive terminal (TTY):

```sh
cd my-project
dpod-seed init
```

`init` will:

1. Fetch the list of available distros
2. Show a numbered menu — enter the number of the distro you want
3. Preview the files that will be written
4. Ask for confirmation: `[y]es / [r]etry / [c]ancel`
5. Write `dpod.yaml` and run `sync` to materialise files

After `init`, your project will contain:

```
dpod.yaml   ← your distro pin
dpod.lock   ← record of every file dpod-seed owns
.devcontainer/
  devcontainer.json   ← materialised by dpod-seed
```

Commit both `dpod.yaml` and `dpod.lock` to version control.

:::note
If you're running in a non-interactive environment (e.g. a script or CI), populate `dpod.yaml` manually and run `dpod-seed sync`.
:::

## 3. Sync

After editing `dpod.yaml` — for example to upgrade the distro tag — run:

```sh
dpod-seed sync
```

`sync` reads `dpod.yaml`, resolves the distro at the pinned tag, computes a diff against `dpod.lock`, and writes only the changed files. It then updates `dpod.lock`.

```
2 updated, 0 added, 0 removed
```

`sync` is idempotent: running it twice in a row with no changes produces `0 added, 0 updated, 0 removed`.

## 4. Validate in CI

Use `validate` in CI pipelines to confirm that the resolved distro version still exists without writing any files:

```sh
dpod-seed validate
```

Exit code 0 means the pin in `dpod.yaml` is resolvable. Any non-zero exit code should fail the pipeline.

Example GitHub Actions step:

```yaml
- name: Validate dpod environment
  run: dpod-seed validate
```

## 5. Eject

If you want to take ownership of the materialised files yourself — stopping `dpod-seed` from managing them — run:

```sh
dpod-seed eject
```

You'll be asked to confirm. On confirmation, `dpod.lock` is removed. The materialised files (e.g. `.devcontainer/devcontainer.json`) are left untouched and are now yours to edit freely.

## dpod.yaml reference

See the [dpod.yaml reference](/dpod-seed/reference/dpod-yaml/) for all supported fields including overrides and patches.
