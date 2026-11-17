---
description: Install mwb with Homebrew, a Linux release package, or a source build, and update each installation.
---

# Install mwb

You need a Windows machine running [PowerToys Mouse Without Borders](https://learn.microsoft.com/en-us/windows/powertoys/mouse-without-borders).
The client runs on Linux with Wayland and uinput, or on macOS with Accessibility permission.

The release configuration builds Linux binaries for amd64 and arm64, and macOS binaries for arm64.
Linux releases include deb/rpm packages. Builds from source need Go 1.27.4 or newer (the `go` line of `go.mod`); macOS builds also need CGo.

## Homebrew

The Homebrew installation is experimental. The cask is published in [bketelsen/homebrew-tap](https://github.com/bketelsen/homebrew-tap).

```sh
brew tap bketelsen/tap
brew install --cask bketelsen/tap/mwb
```

Complete [Linux setup](linux.md) or [macOS setup](macos.md) after installation.
Use the [manual run instructions](../guides/run.md#manual-run) first.

### Update

```sh
brew update
brew upgrade --cask bketelsen/tap/mwb
```

Stop the running client before the update, then start it again with your existing configuration.

## Linux deb/rpm packages

Download the package for your architecture from [GitHub releases](https://github.com/bketelsen/mwb/releases).
In these commands, replace the example filename with the file you downloaded:

```sh
# Debian-based systems
sudo apt install ./mwb_VERSION_ARCH.deb

# RPM-based systems
sudo rpm -Uvh ./mwb_VERSION_ARCH.rpm
```

The packages install the binary at `/usr/bin/mwb`, a systemd user unit at `/usr/lib/systemd/user/mwb.service`, and a udev rule at `/usr/lib/udev/rules.d/99-mwb.rules`.
They do not remove the need to [load uinput and join the input group](linux.md).
Then [configure mwb](../guides/configuration.md) and [start the user service](../guides/run.md#linux-systemd-user-service).

### Update

Download the newer package from GitHub releases and repeat the matching package command above.
Then restart the client:

```sh
systemctl --user restart mwb
```

## From source

```sh
git clone https://github.com/bketelsen/mwb.git
cd mwb
make build
```

The binary is `./mwb`. On Linux, `make install` builds it, installs it to `~/go/bin/mwb`, and installs a systemd user service.
It does not start the service. Complete [Linux setup](linux.md) and [configuration](../guides/configuration.md), then follow [Run mwb](../guides/run.md).

On macOS, build with CGo enabled and copy the binary to the path in the supplied LaunchAgent:

```sh
CGO_ENABLED=1 make build
sudo cp mwb /usr/local/bin/mwb
```

Grant [Accessibility permission](macos.md) before normal use.
The [build guide](../contributing/build.md) lists the contributor gates and pinned tool setup.

### Update

Update the source checkout to the version you want, then rebuild with the same commands.
On Linux, run `make install` again and `systemctl --user restart mwb`.
On macOS, [unload the LaunchAgent](../guides/run.md#macos-launchagent), replace `/usr/local/bin/mwb`, and load it again.
For a manual run, stop the old process and start the new binary.
