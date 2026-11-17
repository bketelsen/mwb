---
description: How mwb connects to PowerToys and delivers input to Linux or macOS.
---

# How mwb works

mwb connects over TCP to the Windows machine running PowerToys Mouse Without Borders.
The default message port is 15101: the [configured base port](../guides/configuration.md) plus 1.

The client and Windows peer exchange an encrypted handshake using AES-256-CBC.
The matching PowerToys security key supplies the key material, derived with PBKDF2-SHA512 and 50,000 iterations.
After the handshake, the client announces its machine identity and name with heartbeat and discovery messages.
This lets PowerToys add it to the device layout.

Windows sends mouse and keyboard events over the connection.
mwb injects them through Linux uinput virtual devices or macOS CoreGraphics events.
That is why [Linux device access](../getting-started/linux.md) or [macOS Accessibility permission](../getting-started/macos.md) is necessary.

If the connection ends, the client attempts to reconnect.
Use [Run mwb](../guides/run.md) to start the client, inspect logs, and refresh the device list in PowerToys.
