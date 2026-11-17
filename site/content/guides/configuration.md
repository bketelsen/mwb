---
description: Configure the Windows host, security key, machine name, and base port.
---

# Configuration

By default, mwb reads `~/.config/mwb/config.toml`.
Create the directory and file before running the client:

```sh
mkdir -p ~/.config/mwb
```

```toml
host = "192.0.2.10"
key = "YourSecurityKey"
name = "linux"
# port = 15100
```

Replace the documentation address with the Windows machine's address.
Copy the security key from **PowerToys > Mouse Without Borders > Security key**.
The value must match PowerToys.

## Settings

| Key | Required | Default and behavior |
| --- | --- | --- |
| `host` | Yes | No default. The Windows host address must not be empty. |
| `key` | Yes | No default. The PowerToys security key must not be empty. |
| `name` | No | The local hostname. If the hostname is empty, `linux`. Names longer than 15 characters are cut; the implementation keeps the first 15 bytes, so use a short ASCII name. |
| `port` | No | `15100` when omitted or set to `0`. This is the base port, not the TCP message port. |

The message port is `port + 1`. With the default base port, mwb connects to TCP port `15101`.
The name appears in the PowerToys device layout.

!!! warning "Keep the key private"

    Do not post the PowerToys security key in an issue, pull request, or discussion. Use a placeholder when sharing configuration examples.

Use `mwb -config /path/to/config.toml` for another file.
Restart mwb after changing its configuration. See [Run mwb](run.md) for service and manual commands.
