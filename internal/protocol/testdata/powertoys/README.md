# PowerToys known-answer vectors

This package-free .NET 10 console harness runs PowerToys encryption methods.
The expected Go test values come from this harness, not from Go encryption.
`Program.cs` identifies the source of each copied method.

## Sources

All copies come from Microsoft's MIT-licensed
`src/modules/MouseWithoutBorders/App/Core/Encryption.cs`:

- [Legacy, 1ce0509fc71b4ce9df39e3ce64a48d254a1f20ba](https://github.com/microsoft/PowerToys/blob/1ce0509fc71b4ce9df39e3ce64a48d254a1f20ba/src/modules/MouseWithoutBorders/App/Core/Encryption.cs),
  the parent of PR #48742: fixed UTF-16LE salt, 50,000 iterations, fixed ASCII IV.
- [Header-50k, a9e1f0f20e9dc3ddd6093839cbda2be7f5daae44](https://github.com/microsoft/PowerToys/blob/a9e1f0f20e9dc3ddd6093839cbda2be7f5daae44/src/modules/MouseWithoutBorders/App/Core/Encryption.cs),
  PR #48742: per-stream salt and IV header, 50,000 iterations.
- [Header-100k, a64c400b7ec7668a84844fe3b95626db46eb6bcb](https://github.com/microsoft/PowerToys/blob/a64c400b7ec7668a84844fe3b95626db46eb6bcb/src/modules/MouseWithoutBorders/App/Core/Encryption.cs),
  PR #49600: the header scheme with 100,000 iterations.

The copied algorithms retain `Rfc2898DeriveBytes.Pbkdf2`, SHA-512,
AES-256-CBC, `PaddingMode.Zeros`, and the write-mode `CryptoStream`.
Only the key and iteration count are parameters, and a fixed header replaces
`RandomNumberGenerator.Fill`. Minimal `Common`, `Package`, and no-op `Logger`
dependencies keep the copied methods standalone. The legacy cache remains;
this harness uses only the original 50,000 iterations with that cache.
The header methods support both historical iteration counts.

The security key, `MwbEncryptionTestKey1234`, is the key from PowerToys
`EncryptionTests.cs`. These are public test inputs, not credentials.

## Regenerate

From the repository root, with the SDK image available:

```sh
podman run --rm \
  -v "$PWD/internal/protocol/testdata/powertoys:/src:Z" -w /src \
  mcr.microsoft.com/dotnet/sdk:10.0 dotnet run -- \
  202122232425262728292a2b2c2d2e2f7e67302f0100000007000000000000000000000000000000000000000000000057494e484f535420202020202020202020202020202020202020202020202020
```

The plaintext is an 80-byte, block-aligned stream:

1. A fixed 16-byte initial block, `202122232425262728292a2b2c2d2e2f`.
2. A Go-serialized 64-byte `Handshake` with `ID: 1`, `Src: 7`, and
   `SetMachineName("WINHOST")`, stamped with the key's magic and checksum.

A throwaway Go test generated only this input; it did not generate expected
keys or ciphertext. The .NET magic is `2f30262c`, consistent with the input's
stamp. Header salt is `000102030405060708090a0b0c0d0e0f`; header IV is
`101112131415161718191a1b1c1d1e1f`. Dispose closes the `CryptoStream` and emits
all ciphertext. Zero padding adds no block to this aligned plaintext.

Copy the harness output into `../../powertoys_vectors_test.go`. Remove only
this harness's generated `bin/` and `obj/` directories after the run.
Go ignores this folder because it is under `testdata/`.

## Independent Python key check

The following third implementation must match all 3 .NET keys:

```sh
python3 - <<'PY'
import hashlib
key = b'MwbEncryptionTestKey1234'
vectors = [
    ('legacy', '18446744073709551615'.encode('utf-16le'), 50000,
     '49d5cef88a62dabd3076c13d3b483cf4f6893b2cbf089beefda7542671d70b0d'),
    ('header-50k', bytes(range(16)), 50000,
     '6fc17832aed42c03078957301c4f35ac32df7a578093f5cca87f8cf73b059acf'),
    ('header-100k', bytes(range(16)), 100000,
     '9ec1e292f2a92a2d2dd62b1a74420dc6f9927ca44acea96359cd9010e8cbbf95'),
]
for name, salt, iterations, expected in vectors:
    actual = hashlib.pbkdf2_hmac('sha512', key, salt, iterations, 32).hex()
    assert actual == expected, (name, actual, expected)
    print(name, actual, 'MATCH .NET')
PY
```

## License for the copied PowerToys code

Copyright (c) Microsoft Corporation

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
