# mwb

Go client for Microsoft PowerToys Mouse Without Borders on Linux (Wayland,
uinput) and macOS (CoreGraphics). It connects to a Windows PC that runs
PowerToys and injects the shared mouse and keyboard into the local session.

This file (`AGENTS.md`) is the canonical agent instructions file. `CLAUDE.md`,
`GEMINI.md`, `.cursorrules` and `.github/copilot-instructions.md` are relative
symlinks to it (frostyard/core ADR-0018, ADR-0029). Edit `AGENTS.md`, never an
alias.

## Map

```
cmd/mwb/            CLI entry point (flags -config and -debug)
internal/
  config/           TOML config loading (~/.config/mwb/config.toml)
  input/            platform-specific virtual mouse and keyboard devices:
                    uinput.go (Linux), coregraphics.go (macOS)
  network/          TCP connection, handshake, packet receive loop
  protocol/         packet types, binary serialization, AES-256-CBC encryption
packaging/          systemd user unit, udev rule, package scripts, macOS LaunchAgent
site/               the user documentation site (ProperDocs, MaterialX)
docs/               engineering docs: adr/, design/, specs/, plans/
.memory/            the corrections inbox
```

## Gate

- `mise install` provisions the pinned tools (`mise.toml`, checksums in
  `mise.lock`): golangci-lint and svu. Go comes from `go.mod` only. Bump
  golangci-lint in `mise.toml` and run `mise lock`; never pin it in a workflow
  or the `Makefile`.
- `make verify` is the credential-free gate and changes no file: tidy diff,
  vet, gofmt check, the pinned golangci-lint and the tests. `make check` runs
  `make fmt`, then `make verify`. `make ci` adds the race tests and the
  builds.
- `mise.toml` also pins GoReleaser Pro. `make ci` runs `make release-check`:
  `goreleaser check --soft` with the pinned GoReleaser, plus a self-test on
  `testdata/goreleaser-invalid.yml`.
- Run `make check` before each commit and `make ci` before each push
  (frostyard/core ADR-0083). `.github/workflows/ci.yml` runs `make ci` on Linux
  and macOS. `make vulncheck` needs the network, so it stays outside `make ci`
  and runs in its own CI job.

## Protocol notes

- Connects to Windows PowerToys MWB on port 15101 (base 15100 + 1 for messages)
- AES-256-CBC encryption with PBKDF2-SHA512 key derivation. Three stream schemes exist across PowerToys versions (`protocol.Scheme`):
  - legacy: fixed salt/IV (`"18446744073709551615"`), 50k iterations
  - header-50k (PowerToys #48742, 2026-06): each stream starts with a cleartext 32-byte header (16-byte random salt + 16-byte random IV), 50k iterations
  - header-100k (PowerToys #49600, 2026-07): same header, 100k iterations
- The server sends its preamble and 10 Handshakes without waiting for the client (after ~5s spent resolving the client's name), so `Connect` reads the first `protocol.ProbeSize` bytes, detects the scheme by decrypting the first packet, then replies with the same scheme
- Handshake: send 10 Handshake packets (type 126), respond to peer's with HandshakeAck (type 127)
- **Critical**: HandshakeAck must have `Src = MachineID` (not 0), otherwise server stores ID.NONE and never routes packets to the socket
- **Critical**: All packet IDs must be non-zero and unique — server has a zero-initialized dedup ring buffer that silently drops ID=0 packets
- After handshake, must send HeartbeatEx (type 51) with `Des=255` (ID.ALL) to trigger `AddToMachinePool()` on the server
- Must respond to Hi (type 2) with Hello (type 3) for device discovery
- Heartbeat responses must include the local machine name (not the remote's)

## System requirements

### Linux

/dev/uinput must be accessible. Requires:
1. `uinput` kernel module loaded (`modprobe uinput`)
2. udev rule: `KERNEL=="uinput", GROUP="input", MODE="0660"` in `/etc/udev/rules.d/99-mwb.rules`
3. User in `input` group

### macOS

Accessibility permission must be granted:
1. Open System Settings > Privacy & Security > Accessibility
2. Add and enable the `mwb` binary (or Terminal.app if running from terminal)
3. Build requires CGo: `CGO_ENABLED=1 go build ./cmd/mwb`

## Releases

- Tag format: `v<major>.<minor>.<patch>`. svu computes the next version from
  conventional commits (`.svu.yml`: `tag.prefix: "v"`, `v0: true`).
- `make bump` checks the svu pin, a clean `main` equal to `origin/main` and
  `make ci`, then tags `svu next` and pushes the tag. Only the maintainer runs
  it: agents never tag, push a tag or publish a release.
- `.github/workflows/release.yml` runs on a tag: GoReleaser Pro builds Linux
  (CGO off) and macOS (CGO on) in a split matrix, then merges and publishes the
  GitHub release, the deb and rpm packages and the `bketelsen/homebrew-tap`
  cask.
- `release.yml` reads the GoReleaser pin with `make goreleaser-version`.
  Bump GoReleaser only in `mise.toml`, then run `mise lock`.

## CI workflows

Every external action in `.github/workflows` is pinned to a full commit SHA
with a `# vX.Y.Z` comment, each workflow has top-level `permissions: {}`, and
each checkout sets `persist-credentials: false` (frostyard/core ADR-0021). A
call of a frostyard/core reusable workflow is pinned to a core commit SHA with
a comment that names its source. Dependabot (`.github/dependabot.yml`) opens
at most 1 grouped version-update PR for each ecosystem and run (ADR-0075).

## Security and risk paths

- `.github/risk-paths.json` lists the high-risk paths: governance and CI files,
  tool pins, the release configuration, the packaging files that run as root,
  and the code that holds the security key, the session key and the
  handshake. The `why` of an L1 or L2 entry starts with 1 category of core's
  `docs/specs/risk-paths.md` (section B) and a colon. `risk_paths_test.go`
  checks the categories and the coverage. A new privileged file gets an entry
  in the same change.
- `.github/workflows/pr-triage.yml` calls core's `pr-triage.yml`: it adds a
  level label (`level:L1` to `level:L4`) and `risk:high` for L1 and L2
  (ADR-0068, ADR-0078), and removes `ready-for-human` when the head moves
  (ADR-0077). Never remove `risk:high` or a level label to make a pull request
  ready; the PR body names each matched path and why the change is safe.
- Report vulnerabilities as `SECURITY.md` says (GitHub private vulnerability
  reporting). Never put a security key or vulnerability details in a public
  issue or in a record under `.memory/corrections/`.

## Corrections

`.memory/corrections/` holds 1 JSON file for each record, with exactly `date`,
`scope`, `correction`, `evidence` and `promoted_to` (`.memory/README.md`,
frostyard/core ADR-0081). A blocking review point adds 1 record file in the
pull request that fixes it, with a guard (a test that fails without the fix,
or an `AGENTS.md` rule). Set `promoted_to` to the file that now holds the
rule; never edit an older record.

## Documentation rules

- Engineering docs live in `docs/` in frostyard/core's four-category shape:
  see the category table, the index and the conventions in
  [docs/README.md](docs/README.md). A new doc starts from its category's
  `TEMPLATE.md` and gets a line in the index.
- A repo-local decision gets an ADR in `docs/adr/`; a decision that binds
  more than this repository goes to frostyard/core, with a line in
  [docs/org-adrs.md](docs/org-adrs.md).
- User documentation lives in the site (`site/content/`); `make site` builds
  it in strict mode. The root `README.md` follows core's README spec (kind
  `tool`).
