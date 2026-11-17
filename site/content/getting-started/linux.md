---
description: Give the Linux client access to uinput for virtual mouse and keyboard devices.
---

# Linux setup

mwb uses uinput to create virtual mouse and keyboard devices on Linux.
It needs access to `/dev/uinput`. The Linux client is intended for a Wayland desktop.

## Load uinput

```sh
sudo modprobe uinput

# Load uinput automatically on boot.
echo 'uinput' | sudo tee /etc/modules-load.d/uinput.conf
```

## Allow device access

The deb/rpm packages supply `/usr/lib/udev/rules.d/99-mwb.rules`.
For a source or Homebrew installation, create a local rule:

```sh
echo 'KERNEL=="uinput", GROUP="input", MODE="0660"' | sudo tee /etc/udev/rules.d/99-uinput.rules
```

Reload the rules and add your account to the `input` group:

```sh
sudo udevadm control --reload-rules
sudo udevadm trigger /dev/uinput
sudo usermod -aG input "$USER"
```

Log out and log back in so the group change takes effect.

!!! warning "Input access"

    This rule gives members of the input group access to uinput. Grant that membership only to accounts that need to create input devices.

Next, [configure mwb](../guides/configuration.md) and [run it](../guides/run.md).
