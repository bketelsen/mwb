# mwb

A Linux client for [Microsoft PowerToys Mouse Without Borders](https://learn.microsoft.com/en-us/windows/powertoys/mouse-without-borders). Share your Windows keyboard and mouse with a Linux machine over the network.

## Prerequisites

- Go 1.25+
- Linux with uinput support
- A Windows machine running PowerToys Mouse Without Borders

## System Setup

Before running mwb, you need to configure uinput access:

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

## Installation

```bash
make build
```

Or install directly:

```bash
make install
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

```bash
mwb                          # uses default config path
mwb -config /path/to/config  # custom config path
mwb -debug                   # enable debug logging
```

On the Windows side, open PowerToys Mouse Without Borders and click "Refresh connections" to discover the Linux client. Then drag it into your desired position in the device layout.

## How It Works

mwb connects to the Windows MWB server over TCP, performs an AES-256-CBC encrypted handshake, and receives mouse/keyboard events which it injects locally via Linux uinput virtual devices.
