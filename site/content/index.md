---
title: Share a Windows keyboard and mouse
description: Use PowerToys Mouse Without Borders with a Linux or macOS client.
hide:
  - navigation
  - toc
---

<div class="fy-hero" markdown>
<div class="fy-hero__text" markdown>
<p class="fy-eyebrow">mwb</p>

# Share a Windows keyboard and mouse

<p class="fy-hero__lede">mwb is a Linux and macOS client for Microsoft PowerToys Mouse Without Borders. Use your Windows machine's keyboard and mouse on a Linux or macOS machine over the network.</p>

[Install mwb](getting-started/install.md){ .md-button .md-button--primary }
[How it works](concepts/how-it-works.md){ .md-button }

</div>
</div>

## Start here

1. [Install mwb](getting-started/install.md), or update an existing installation.
2. Set up [Linux uinput access](getting-started/linux.md) or [macOS Accessibility permission](getting-started/macos.md).
3. [Configure](guides/configuration.md) the Windows host and matching PowerToys security key.
4. [Run mwb](guides/run.md) manually or at login, then select **Refresh connections** in PowerToys.

## Read more

- [Configuration](guides/configuration.md): the 4 TOML settings and their defaults.
- [Run mwb](guides/run.md): service setup, command flags, and logs.
- [How mwb works](concepts/how-it-works.md): the connection and local input devices.
- [Build from source](contributing/build.md): contributor checks and site development.

!!! note "Pre-release"

    Linux needs uinput access. macOS needs Accessibility permission. A Windows machine must run PowerToys Mouse Without Borders.
