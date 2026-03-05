# Fix GoReleaser Homebrew Formula Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix the goreleaser merge step so it generates a Homebrew formula instead of failing on homebrew_casks.

**Architecture:** Replace the `homebrew_casks` block in `.goreleaser.yml` with a `brews` block. Drop the `ids` filter. No other files change.

**Tech Stack:** GoReleaser Pro v2, GitHub Actions

---

### Task 1: Replace homebrew_casks with brews in goreleaser config

**Files:**
- Modify: `.goreleaser.yml:56-88`

**Step 1: Replace the homebrew_casks block**

Replace lines 56-88 (the entire `homebrew_casks` block) with:

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

Key differences from the old block:
- `brews` instead of `homebrew_casks`
- No `ids` filter (picks up all archives)
- `directory: Formula` inside `repository` (not top-level `Casks`)
- Ruby DSL `install` block instead of raw hooks
- `service` block for `brew services` support
- No `post.install` hook (not needed for formulae)

**Step 2: Validate goreleaser config syntax**

Run: `goreleaser check` (if installed locally), or just verify YAML is valid:
```bash
python3 -c "import yaml; yaml.safe_load(open('.goreleaser.yml'))" && echo "YAML valid"
```
Expected: "YAML valid"

**Step 3: Run make check**

Run: `make check`
Expected: All formatting, linting, and tests pass (goreleaser config change doesn't affect Go code)

**Step 4: Commit**

```bash
git add .goreleaser.yml
git commit -m "fix: replace homebrew_casks with brews in goreleaser config

homebrew_casks is for macOS GUI apps, not CLI tools. The ids filter
also referenced build IDs instead of archive IDs, causing the merge
step to fail with 'no linux/macos archives found'."
```

### Task 2: Tag and release to verify fix

This task is manual — push a new tag to trigger the release workflow and confirm the merge step succeeds.

```bash
git tag v0.X.Y  # next version
git push origin main --tags
```

Monitor: https://github.com/bketelsen/mwb/actions — the merge job should now complete and push `Formula/mwb.rb` to `bketelsen/homebrew-tap`.
