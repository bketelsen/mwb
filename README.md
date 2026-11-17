<!-- readme-kind: tool -->
# mwb

[![CI](https://github.com/bketelsen/mwb/actions/workflows/ci.yml/badge.svg)](https://github.com/bketelsen/mwb/actions/workflows/ci.yml)

mwb is a Linux and macOS client for people who use [Microsoft PowerToys Mouse Without Borders](https://learn.microsoft.com/en-us/windows/powertoys/mouse-without-borders) on Windows. Share the Windows machine's keyboard and mouse with your Linux or macOS machine over the network.

**Status:** pre-release. Linux needs uinput access, and macOS needs Accessibility permission.

**Documentation:** [bketelsen.github.io/mwb](https://bketelsen.github.io/mwb/)

## What it does

- **Receive keyboard and mouse input.** Control Linux or macOS from the Windows machine. [How it works](https://bketelsen.github.io/mwb/concepts/how-it-works/)
- **Run at login.** Use a systemd user service on Linux or a LaunchAgent on macOS. [Run mwb](https://bketelsen.github.io/mwb/guides/run/)

## Install

```sh
brew tap bketelsen/tap
brew install --cask bketelsen/tap/mwb
```

The [install guide](https://bketelsen.github.io/mwb/getting-started/install/) covers release packages, source builds, updates, and platform requirements.

## Use

First, complete [Linux setup](https://bketelsen.github.io/mwb/getting-started/linux/) or [macOS setup](https://bketelsen.github.io/mwb/getting-started/macos/). Create `~/.config/mwb/config.toml` with your Windows host and PowerToys security key as the [configuration guide](https://bketelsen.github.io/mwb/guides/configuration/) shows. Then run:

```sh
mwb
```

In PowerToys Mouse Without Borders, select **Refresh connections** and place the client in the device layout. [Run mwb](https://bketelsen.github.io/mwb/guides/run/) explains services, flags, and logs.

## Develop

```sh
mise install   # the pinned tools
make ci        # the local CI gate
```

[Build from source](https://bketelsen.github.io/mwb/contributing/build/) explains the build, checks, and site targets.

## Where to go next

- [AGENTS.md](AGENTS.md): the rules for agents and contributors.
- [docs/README.md](docs/README.md): the decisions, designs, specs, and plans.
- [SECURITY.md](SECURITY.md): report a vulnerability privately.
- [LICENSE](LICENSE): MIT.
