# mwb - Mouse Without Borders Client (Linux & macOS)

## Project Structure

```
cmd/mwb/           - CLI entry point
internal/
  config/           - TOML config loading (~/.config/mwb/config.toml)
  input/            - platform-specific virtual mouse/keyboard devices
                      uinput.go (Linux), coregraphics.go (macOS)
  network/          - TCP connection, handshake, packet receive loop
  protocol/         - Packet types, binary serialization, AES-256-CBC encryption
docs/plans/         - Implementation plan
```

## Build & Test

```bash
make build   # builds ./mwb
make test    # runs go test ./...
make fmt     # format all Go files
make lint    # run golangci-lint
make check   # fmt + lint + test (run before committing)
```

## Before Committing

Always run `make check` before committing any changes. All code must pass formatting, linting, and tests.

## Protocol Notes

- Connects to Windows PowerToys MWB on port 15101 (base 15100 + 1 for messages)
- AES-256-CBC encryption with PBKDF2-SHA512 key derivation (50k iterations)
- Handshake: send 10 Handshake packets (type 126), respond to peer's with HandshakeAck (type 127)
- **Critical**: HandshakeAck must have `Src = MachineID` (not 0), otherwise server stores ID.NONE and never routes packets to the socket
- **Critical**: All packet IDs must be non-zero and unique — server has a zero-initialized dedup ring buffer that silently drops ID=0 packets
- After handshake, must send HeartbeatEx (type 51) with `Des=255` (ID.ALL) to trigger `AddToMachinePool()` on the server
- Must respond to Hi (type 2) with Hello (type 3) for device discovery
- Heartbeat responses must include the local machine name (not the remote's)

## System Requirements

### Linux

/dev/uinput must be accessible. Requires:
1. `uinput` kernel module loaded (`modprobe uinput`)
2. udev rule: `KERNEL=="uinput", GROUP="input", MODE="0660"` in `/etc/udev/rules.d/99-uinput.rules`
3. User in `input` group

### macOS

Accessibility permission must be granted:
1. Open System Settings > Privacy & Security > Accessibility
2. Add and enable the `mwb` binary (or Terminal.app if running from terminal)
3. Build requires CGo: `CGO_ENABLED=1 go build ./cmd/mwb`
