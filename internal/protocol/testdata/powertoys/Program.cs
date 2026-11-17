// Copyright (c) Microsoft Corporation
// The Microsoft Corporation licenses the copied PowerToys code under the MIT license.
// See README.md for the license, upstream sources, and adaptation boundaries.
using System;
using System.Collections.Concurrent;
using System.Globalization;
using System.IO;
using System.Net.Sockets;
using System.Security.Cryptography;
using System.Text;

internal static class Program
{
    private static void Main(string[] args)
    {
        if (args.Length != 1)
        {
            throw new ArgumentException("Usage: dotnet run -- <plaintext-hex>");
        }

        const string key = "MwbEncryptionTestKey1234";
        byte[] plaintext = Convert.FromHexString(args[0]);
        byte[] header = Convert.FromHexString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f");
        Console.WriteLine($"magic={LegacyEncryption.Get24BitHash(key):x8}");
        Print("legacy-key", LegacyEncryption.KeyVector(key, 50000));
        Print("header-50k-key", HeaderEncryption.KeyVector(key, header[..16], 50000));
        Print("header-100k-key", HeaderEncryption.KeyVector(key, header[..16], 100000));
        PrintWire("legacy-wire", plaintext, output => LegacyEncryption.GetEncryptedStream(output, key, 50000));
        PrintWire("header-50k-wire", plaintext, output => HeaderEncryption.GetEncryptedStream(output, key, 50000, header));
        PrintWire("header-100k-wire", plaintext, output => HeaderEncryption.GetEncryptedStream(output, key, 100000, header));
    }

    private static void Print(string name, byte[] value) => Console.WriteLine($"{name}={Convert.ToHexString(value).ToLowerInvariant()}");

    private static void PrintWire(string name, byte[] plaintext, Func<Stream, Stream> create)
    {
        using var output = new MemoryStream();
        using (Stream encrypted = create(output))
        {
            encrypted.Write(plaintext, 0, plaintext.Length);
        }
        Print(name, output.ToArray());
    }

    // Same AES setup as Encryption.InitEncryption in both upstream versions.
    internal static AesCryptoServiceProvider CreateCipher()
    {
        var cipher = new AesCryptoServiceProvider();
        cipher.KeySize = 256;
        cipher.BlockSize = 16 * 8;
        cipher.Padding = PaddingMode.Zeros;
        cipher.Mode = CipherMode.CBC;
        cipher.GenerateIV();
        return cipher;
    }
}

internal static class LegacyEncryption
{
    private static readonly AesCryptoServiceProvider symAl = Program.CreateCipher();
    private static readonly string InitialIV = ulong.MaxValue.ToString(CultureInfo.InvariantCulture);
    private static readonly ConcurrentDictionary<string, byte[]> LegalKeyDictionary = new(StringComparer.OrdinalIgnoreCase);

    internal static byte[] KeyVector(string key, int iterations) => GenLegalKey(key, iterations);

    // Copied from microsoft/PowerToys src/modules/MouseWithoutBorders/App/Core/Encryption.cs,
    // commit 1ce0509fc71b4ce9df39e3ce64a48d254a1f20ba (MIT).
    // Only the security key and iteration count are parameters instead of globals/literals.
    private static byte[] GenLegalKey(string key, int iterations)
    {
        byte[] rv;
        string myKey = key;

        if (!LegalKeyDictionary.TryGetValue(myKey, out byte[] value))
        {
            rv = Rfc2898DeriveBytes.Pbkdf2(
                myKey,
                Common.GetBytesU(InitialIV),
                iterations,
                HashAlgorithmName.SHA512,
                32);
            _ = LegalKeyDictionary.AddOrUpdate(myKey, rv, (k, v) => rv);
        }
        else
        {
            rv = value;
        }

        return rv;
    }

    // Copied unchanged from the same Encryption.cs at
    // 1ce0509fc71b4ce9df39e3ce64a48d254a1f20ba (MIT).
    private static byte[] GenLegalIV()
    {
        string st = InitialIV;
        int ivLength = symAl.IV.Length;
        if (st.Length > ivLength)
        {
            st = st[..ivLength];
        }
        else if (st.Length < ivLength)
        {
            st = st.PadRight(ivLength, ' ');
        }

        return Common.GetBytes(st);
    }

    // Copied from the same Encryption.cs at
    // 1ce0509fc71b4ce9df39e3ce64a48d254a1f20ba (MIT); key/iterations parameterized.
    internal static Stream GetEncryptedStream(Stream encryptedStream, string key, int iterations)
    {
        ICryptoTransform encryptor;
        encryptor = symAl.CreateEncryptor(GenLegalKey(key, iterations), GenLegalIV());
        return new CryptoStream(encryptedStream, encryptor, CryptoStreamMode.Write);
    }

    // Copied unchanged from the same Encryption.cs at
    // 1ce0509fc71b4ce9df39e3ce64a48d254a1f20ba (MIT).
    // The same method is present at a64c400b7ec7668a84844fe3b95626db46eb6bcb.
    internal static uint Get24BitHash(string st)
    {
        if (string.IsNullOrEmpty(st))
        {
            return 0;
        }

        byte[] bytes = new byte[Package.PACKAGE_SIZE];
        for (int i = 0; i < Package.PACKAGE_SIZE; i++)
        {
            if (i < st.Length)
            {
                bytes[i] = (byte)st[i];
            }
        }

        var hash = SHA512.Create();
        byte[] hashValue = hash.ComputeHash(bytes);

        for (int i = 0; i < 50000; i++)
        {
            hashValue = hash.ComputeHash(hashValue);
        }

        Logger.LogDebug(string.Format(CultureInfo.CurrentCulture, "magic: {0},{1},{2}", hashValue[0], hashValue[1], hashValue[^1]));
        hash.Clear();
        return (uint)((hashValue[0] << 23) + (hashValue[1] << 16) + (hashValue[^1] << 8) + hashValue[2]);
    }
}

internal static class HeaderEncryption
{
    private static readonly AesCryptoServiceProvider symAl = Program.CreateCipher();
    private const int SaltSize = 16;
    private const int SymAlBlockSize = 16;
    private const int DerivedKeyLength = 32;

    // Copied from microsoft/PowerToys src/modules/MouseWithoutBorders/App/Core/Encryption.cs,
    // a64c400b7ec7668a84844fe3b95626db46eb6bcb (MIT), with key parameterized.
    private static void EnsurePairingKeyAvailable(string key)
    {
        if (string.IsNullOrWhiteSpace(key))
        {
            throw new CryptographicException("Pairing setup is required.");
        }
    }

    internal static byte[] KeyVector(string key, byte[] salt, int iterations) => GenLegalKey(key, salt, iterations);

    // Copied from the same Encryption.cs at a64c400b7ec7668a84844fe3b95626db46eb6bcb
    // (MIT), with key and iterations parameterized. The 50k variant comes from
    // a9e1f0f20e9dc3ddd6093839cbda2be7f5daae44; its iteration constant is 50000.
    private static byte[] GenLegalKey(string key, byte[] salt, int iterations)
    {
        return Rfc2898DeriveBytes.Pbkdf2(
            key,
            salt,
            iterations,
            HashAlgorithmName.SHA512,
            DerivedKeyLength);
    }

    // Copied from the same Encryption.cs at a64c400b7ec7668a84844fe3b95626db46eb6bcb
    // (MIT); a9e1f0f20e9dc3ddd6093839cbda2be7f5daae44 uses 50000 iterations.
    // Only key/iterations and the fixed input header replace production inputs.
    internal static Stream GetEncryptedStream(Stream encryptedStream, string key, int iterations, byte[] header)
    {
        EnsurePairingKeyAvailable(key);

        // Production allocates the header and calls RandomNumberGenerator.Fill(header).
        ExchangeEncryptionHeader(encryptedStream, header, send: true);

        byte[] salt = header[..SaltSize];
        byte[] iv = header[SaltSize..];

        ICryptoTransform encryptor = symAl.CreateEncryptor(GenLegalKey(key, salt, iterations), iv);
        return new CryptoStream(encryptedStream, encryptor, CryptoStreamMode.Write);
    }

    // Copied unchanged from the same Encryption.cs at
    // a64c400b7ec7668a84844fe3b95626db46eb6bcb (MIT).
    private static void ExchangeEncryptionHeader(Stream stream, byte[] header, bool send)
    {
        try
        {
            if (send)
            {
                stream.Write(header, 0, header.Length);
            }
            else
            {
                stream.ReadExactly(header);
            }
        }
        catch (IOException e)
        {
            Logger.Log($"{nameof(ExchangeEncryptionHeader)}: Exception {(send ? "writing" : "reading")} the encryption header to the socket stream: {e.InnerException?.GetType()}/{e.Message}. (This is expected when the remote machine closes the connection during desktop switch or reconnection.)");

            if (e is not EndOfStreamException && e.InnerException is not (SocketException or ObjectDisposedException))
            {
                throw;
            }
        }
    }
}

// Minimal dependencies for the copied methods: Common.GetBytesU uses UTF-16LE;
// Common.GetBytes uses ASCII; the on-wire packet header size is 32 bytes.
internal static class Common
{
    internal static byte[] GetBytesU(string value) => ASCIIEncoding.Unicode.GetBytes(value);
    internal static byte[] GetBytes(string value) => Encoding.ASCII.GetBytes(value);
}

internal static class Package
{
    internal const int PACKAGE_SIZE = 32;
}

internal static class Logger
{
    internal static void LogDebug(string message) { }
    internal static void Log(string message) { }
}
