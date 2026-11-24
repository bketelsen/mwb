---
description: Run mwb manually, as a Linux systemd user service, or as a macOS LaunchAgent, and read its logs.
---

# Run mwb

First, [install mwb](../getting-started/install.md), complete your platform's input permissions, and [create the configuration](configuration.md).

## Manual run

```sh
mwb
mwb -config /path/to/config.toml
mwb -debug
```

| Flag | Effect |
| --- | --- |
| `-config` | Select another configuration file. The default is `~/.config/mwb/config.toml`. |
| `-debug` | Enable debug logging. Normal logging uses the info level. |

Logs go to standard error. Use `./mwb` instead of `mwb` when running the source build from the repository.
Press Ctrl+C to stop a manual run.
The client retries failed connections, with a delay that starts at 1 second and increases to a maximum of 30 seconds.

## Linux systemd user service

The source installation command `make install` places the root `mwb.service` in `~/.config/systemd/user/`.
That unit runs `~/go/bin/mwb`. The deb/rpm packages instead supply a user unit that runs `/usr/bin/mwb`.
A user-local unit takes precedence; do not keep an old source unit when switching to the packaged binary.

After installation and configuration:

```sh
systemctl --user daemon-reload
systemctl --user enable --now mwb
journalctl --user -u mwb -f
```

The unit restarts mwb on failure after 5 seconds.
To stop it and disable startup:

```sh
systemctl --user disable --now mwb
```

For a source installation, `make uninstall` removes the source-installed binary and user unit.
A Homebrew installation can use the [manual run](#manual-run); the cask has no service configuration.

## macOS LaunchAgent

The repository supplies [packaging/com.mwb.agent.plist](https://github.com/bketelsen/mwb/blob/main/packaging/com.mwb.agent.plist).
Its program path is `/usr/local/bin/mwb`.
Install the binary there, or edit your local plist copy to use your installed binary's absolute path.

Use 1 of the 2 blocks below: the first for a source checkout, the second for a Homebrew installation.
If the agent is already loaded, `launchctl load` fails. Run `launchctl unload ~/Library/LaunchAgents/com.mwb.agent.plist` first.

From the source checkout:

```sh
mkdir -p ~/Library/LaunchAgents
cp packaging/com.mwb.agent.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.mwb.agent.plist
```

For a Homebrew installation, download the plist and set its program path before loading it.
The `curl` command takes the plist from the `main` branch, which can be newer than your installed release.
To match your release, replace `main` in the URL with the release tag: `v` followed by the version that `brew info --cask bketelsen/tap/mwb` shows.

```sh
mkdir -p ~/Library/LaunchAgents
curl -fsSL -o ~/Library/LaunchAgents/com.mwb.agent.plist https://raw.githubusercontent.com/bketelsen/mwb/main/packaging/com.mwb.agent.plist
plutil -replace ProgramArguments.0 -string "$(brew --prefix)/bin/mwb" ~/Library/LaunchAgents/com.mwb.agent.plist
launchctl load ~/Library/LaunchAgents/com.mwb.agent.plist
```

The agent starts at login and restarts after an unsuccessful exit.
Its log paths are:

```sh
tail -f /tmp/mwb.log
tail -f /tmp/mwb.err.log
```

mwb writes its logs to standard error, so check `/tmp/mwb.err.log` first.
To stop and unload it:

```sh
launchctl unload ~/Library/LaunchAgents/com.mwb.agent.plist
```

## Refresh connections in PowerToys

On Windows, open PowerToys Mouse Without Borders and select **Refresh connections**.
When the client appears, drag it to the position you want in the device layout.

If input devices cannot be created, check [Linux uinput access](../getting-started/linux.md) or [macOS Accessibility permission](../getting-started/macos.md).
For connection failures, compare the configured host, base port, and security key with PowerToys, and read the logs above.
