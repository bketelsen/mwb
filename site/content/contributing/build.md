---
description: Build mwb, run contributor checks, and edit the documentation site.
---

# Build from source

## Tools and checks

The source build needs Go 1.27.4 or newer (the `go` line of `go.mod`). macOS builds need CGo enabled.
Use [mise](https://mise.jdx.dev) for the repository's pinned contributor tools.

```sh
mise install        # the pinned tools
make ci             # the local CI gate
make check          # format, lint, and test
make release-check  # check the GoReleaser config
make build          # writes ./mwb
```

Run `make ci` before a push and `make check` before a commit.
The CI gate and tool pins are part of the repository baseline.

For a macOS build:

```sh
CGO_ENABLED=1 make build
```

## Install a source build

On Linux, `make install` builds and installs the binary at `~/go/bin/mwb` and the systemd user unit at `~/.config/systemd/user/mwb.service`:

```sh
make install
```

It reloads the user service manager, but does not start mwb.
Complete the [Linux setup](../getting-started/linux.md) and [configuration](../guides/configuration.md), then [run the service](../guides/run.md#linux-systemd-user-service).
Do not use this Linux systemd target on macOS; follow the [macOS source installation](../getting-started/install.md#from-source) instead.

## This site

The user site lives in `site/`: pages in `site/content/` and configuration in `site/properdocs.yml`.
[ProperDocs](https://properdocs.org) and [MaterialX](https://jaywhj.github.io/mkdocs-materialx/) build it.
GitHub Pages publishes it from `main`.
Python 3 with venv support is required.

```sh
make site        # strict build into site/dist
make site-serve  # preview at http://127.0.0.1:8000/mwb/
```

Both targets first install the hash-locked requirements into `build/site-venv`.
A warning fails the strict build. Update the affected pages in the same PR as any user-visible behavior change.

To refresh the dependency lock after editing `site/requirements.in`, use uv:

```sh
uv pip compile --universal --generate-hashes site/requirements.in -o site/requirements.txt
```

## Maintainer documents

Decisions, designs, specs, and plans live in [docs/](https://github.com/bketelsen/mwb/tree/main/docs).
Start at the [documentation index](https://github.com/bketelsen/mwb/blob/main/docs/README.md).
Read [AGENTS.md](https://github.com/bketelsen/mwb/blob/main/AGENTS.md) before changing the repository.
