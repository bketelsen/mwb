---
description: Grant macOS Accessibility permission to mwb or Terminal.app.
---

# macOS setup

mwb needs Accessibility permission to inject mouse and keyboard events.

1. [Install or build mwb](install.md).
2. Run it once. macOS can prompt for Accessibility access.
3. Open **System Settings > Privacy & Security > Accessibility**.
4. Add and enable the mwb binary. If you run it from Terminal, you may need to enable Terminal.app instead.

A source build needs CGo:

```sh
CGO_ENABLED=1 make build
```

Next, [configure mwb](../guides/configuration.md) and [run it](../guides/run.md).
The supplied [LaunchAgent instructions](../guides/run.md#macos-launchagent) use `/usr/local/bin/mwb`.
