# Fix GoReleaser Homebrew Formula Generation

Date: 2026-03-04

## Problem

The goreleaser merge step fails with:

```
no linux/macos archives found matching goos=[darwin linux] goarch=[amd64 arm64] ids=[mwb-linux mwb-darwin]
```

Two issues:
1. `.goreleaser.yml` uses `homebrew_casks` (for macOS GUI apps) instead of `brews` (for CLI tools)
2. The `ids` filter references build IDs (`mwb-linux`, `mwb-darwin`) but goreleaser filters by archive IDs — the single archive has no explicit ID (defaults to `"default"`), so nothing matches

## Fix (Approach A1)

Replace `homebrew_casks` with `brews` per the original design doc. Drop the `ids` filter since there is only one archive definition and no need to filter.

### Changes

**`.goreleaser.yml`** — replace `homebrew_casks` block (lines 56-88) with:

```yaml
brews:
  - name: mwb
    repository:
      owner: bketelsen
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
      directory: Formula
    homepage: https://github.com/bketelsen/mwb
    description: Mouse Without Borders client for Linux and macOS
    license: MIT
    install: |
      bin.install "mwb"
    service: |
      run opt_bin/"mwb"
      keep_alive true
      log_path var/"log/mwb.log"
      error_log_path var/"log/mwb.log"
    caveats: |
      On macOS, grant Accessibility permissions before starting:
        System Settings > Privacy & Security > Accessibility > Add mwb

      On Linux, set up uinput access:
        sudo modprobe uinput
        echo 'KERNEL=="uinput", GROUP="input", MODE="0660"' | \
          sudo tee /etc/udev/rules.d/99-uinput.rules
        sudo udevadm control --reload-rules && sudo udevadm trigger
        sudo usermod -aG input $USER
```

### No changes needed

- `.github/workflows/release.yml` — `HOMEBREW_TAP_TOKEN` already passed in merge job
- Archive format, nfpms, split/merge CI structure all unchanged
