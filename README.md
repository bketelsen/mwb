# mwb

A Linux and macOS client for [Microsoft PowerToys Mouse Without Borders](https://learn.microsoft.com/en-us/windows/powertoys/mouse-without-borders). Share your Windows keyboard and mouse with a Linux or macOS machine over the network.

## Prerequisites

- Go 1.25+
- A Windows machine running PowerToys Mouse Without Borders
- **Linux:** uinput support, Wayland display
- **macOS:** Accessibility permission granted to the mwb binary

## Linux Setup

Before running mwb on Linux, you need to configure uinput access:

```bash
# Load the uinput kernel module
sudo modprobe uinput

# Load uinput automatically on boot
echo 'uinput' | sudo tee /etc/modules-load.d/uinput.conf

# Allow the input group to access /dev/uinput
echo 'KERNEL=="uinput", GROUP="input", MODE="0660"' | sudo tee /etc/udev/rules.d/99-uinput.rules
sudo udevadm control --reload-rules && sudo udevadm trigger /dev/uinput

# Add your user to the input group (log out and back in after this)
sudo usermod -aG input $USER
```

## macOS Setup

mwb needs Accessibility permission to inject mouse and keyboard events:

1. Build or install the `mwb` binary
2. Run it once — macOS will prompt for Accessibility access
3. Grant permission in **System Settings > Privacy & Security > Accessibility**

If running from Terminal, you may need to grant the permission to Terminal.app instead.

## Installation

### Linux

```bash
make install
```

This builds the binary, installs it to `~/go/bin/mwb`, and sets up a systemd user service.

### macOS

```bash
make build
cp mwb /usr/local/bin/mwb
```

## Configuration

Create `~/.config/mwb/config.toml`:

```toml
host = "192.168.1.100"   # IP of the Windows machine
key = "YourSecurityKey"   # Must match the key set in PowerToys MWB
name = "linux"            # Name shown in the MWB device layout (max 15 chars)
# port = 15100            # Base port (default: 15100, message port = base + 1)
```

The security key is found in PowerToys > Mouse Without Borders > Security key.

## Usage

### Linux: systemd service (recommended)

```bash
systemctl --user enable --now mwb
```

View logs:

```bash
journalctl --user -u mwb -f
```

To uninstall:

```bash
make uninstall
```

### macOS: LaunchAgent (recommended)

```bash
cp packaging/com.mwb.agent.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.mwb.agent.plist
```

This starts mwb at login and restarts it on failure. View logs:

```bash
tail -f /tmp/mwb.log
tail -f /tmp/mwb.err.log
```

To stop and unload:

```bash
launchctl unload ~/Library/LaunchAgents/com.mwb.agent.plist
```

### Manual

```bash
mwb                          # uses default config path
mwb -config /path/to/config  # custom config path
mwb -debug                   # enable debug logging
```

On the Windows side, open PowerToys Mouse Without Borders and click "Refresh connections" to discover the Linux client. Then drag it into your desired position in the device layout.

## How It Works

mwb connects to the Windows MWB server over TCP, performs an AES-256-CBC encrypted handshake, and receives mouse/keyboard events which it injects locally via Linux uinput virtual devices or macOS CoreGraphics events.
