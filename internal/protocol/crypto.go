// internal/protocol/crypto.go
package protocol

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	aesKeySize   = 32
	aesBlockSize = 16
	pbkdf2Iter   = 50000
	initialIVStr = "18446744073709551615"
)

// pbkdf2Salt returns the UTF-16LE encoding of "18446744073709551615" (40 bytes).
func pbkdf2Salt() []byte {
	s := initialIVStr
	salt := make([]byte, len(s)*2)
	for i, c := range s {
		binary.LittleEndian.PutUint16(salt[i*2:], uint16(c))
	}
	return salt
}

// DeriveKey derives a 32-byte AES-256 key using PBKDF2-SHA512 with 50,000 iterations.
func DeriveKey(securityKey string) []byte {
	return pbkdf2.Key([]byte(securityKey), pbkdf2Salt(), pbkdf2Iter, aesKeySize, sha512.New)
}

// FixedIV returns the 16-byte IV for AES-CBC (ASCII "1844674407370955").
func FixedIV() []byte {
	s := initialIVStr
	if len(s) > aesBlockSize {
		s = s[:aesBlockSize]
	}
	return []byte(s)
}

// Scheme identifies the stream encryption variant a PowerToys MWB peer uses.
type Scheme int

const (
	// SchemeLegacy uses the fixed salt and IV with 50,000 PBKDF2 iterations.
	SchemeLegacy Scheme = iota
	// SchemeHeader50k sends a cleartext random salt+IV header before the
	// cipher text and derives the key with 50,000 iterations (PowerToys #48742).
	SchemeHeader50k
	// SchemeHeader100k is SchemeHeader50k with 100,000 iterations (PowerToys #49600).
	SchemeHeader100k
)

func (s Scheme) String() string {
	switch s {
	case SchemeLegacy:
		return "legacy"
	case SchemeHeader50k:
		return "header-50k"
	case SchemeHeader100k:
		return "header-100k"
	}
	return fmt.Sprintf("Scheme(%d)", int(s))
}

// HasHeader reports whether streams start with a cleartext salt+IV header.
func (s Scheme) HasHeader() bool {
	return s != SchemeLegacy
}

func (s Scheme) iterations() int {
	if s == SchemeHeader100k {
		return 100000
	}
	return pbkdf2Iter
}

const saltSize = 16

// HeaderSize is the size of the cleartext salt+IV header used by header schemes.
const HeaderSize = saltSize + aesBlockSize

// ProbeSize is how much of a peer's stream DetectScheme needs: the optional
// header, the 16-byte random initial block, and the first packet header.
const ProbeSize = HeaderSize + aesBlockSize + PacketSize

// DeriveSchemeKey derives the AES-256 key for a header scheme from the salt
// carried in the stream header.
func DeriveSchemeKey(securityKey string, salt []byte, s Scheme) []byte {
	return pbkdf2.Key([]byte(securityKey), salt, s.iterations(), aesKeySize, sha512.New)
}

// DetectScheme determines which scheme produced the start of a peer's stream
// by decrypting its first packet and validating it against magic. prefix must
// hold at least ProbeSize bytes. It returns the scheme, the AES key and the IV
// for decrypting the peer's stream, and how many bytes of prefix precede the
// cipher text.
func DetectScheme(prefix []byte, securityKey string, magic uint32) (s Scheme, key, iv []byte, offset int, err error) {
	if len(prefix) < ProbeSize {
		return 0, nil, nil, 0, fmt.Errorf("probe too short: %d bytes, need %d", len(prefix), ProbeSize)
	}

	firstPacketValid := func(key, iv, ct []byte) bool {
		block, err := aes.NewCipher(key)
		if err != nil {
			return false
		}
		pt := make([]byte, aesBlockSize+PacketSize)
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(pt, ct[:len(pt)])
		return ValidatePacket(pt[aesBlockSize:], magic) == nil
	}

	salt, headerIV := prefix[:saltSize], prefix[saltSize:HeaderSize]
	for _, s := range []Scheme{SchemeHeader100k, SchemeHeader50k} {
		key := DeriveSchemeKey(securityKey, salt, s)
		if firstPacketValid(key, headerIV, prefix[HeaderSize:]) {
			return s, key, headerIV, HeaderSize, nil
		}
	}

	key = DeriveKey(securityKey)
	if firstPacketValid(key, FixedIV(), prefix) {
		return SchemeLegacy, key, FixedIV(), 0, nil
	}

	return 0, nil, nil, 0, fmt.Errorf("cannot decrypt peer stream with any known scheme (security key mismatch?)")
}

// Get24BitHash computes the MWB magic number from the security key.
func Get24BitHash(key string) uint32 {
	if key == "" {
		return 0
	}
	buf := make([]byte, PacketSize)
	for i := 0; i < len(key) && i < PacketSize; i++ {
		buf[i] = byte(key[i])
	}
	h := sha512.New()
	h.Write(buf)
	hash := h.Sum(nil)
	for i := 0; i < 50000; i++ {
		h.Reset()
		h.Write(hash)
		hash = h.Sum(nil)
	}
	return uint32((int(hash[0]) << 23) + (int(hash[1]) << 16) + (int(hash[len(hash)-1]) << 8) + int(hash[2]))
}

// StampPacket sets the magic number and checksum in a marshaled packet buffer.
func StampPacket(buf []byte, magic uint32) {
	buf[3] = byte((magic >> 24) & 0xFF)
	buf[2] = byte((magic >> 16) & 0xFF)
	buf[1] = 0
	for i := 2; i < PacketSize; i++ {
		buf[1] = buf[1] + buf[i]
	}
}

// ValidatePacket checks the magic number and checksum.
func ValidatePacket(buf []byte, magic uint32) error {
	wireMagic := uint32(buf[3])<<24 + uint32(buf[2])<<16
	expectedMagic := magic & 0xFFFF0000
	if wireMagic != expectedMagic {
		return fmt.Errorf("magic mismatch: wire=0x%08X expected=0x%08X", wireMagic, expectedMagic)
	}
	var checksum byte
	for i := 2; i < PacketSize; i++ {
		checksum += buf[i]
	}
	if buf[1] != checksum {
		return fmt.Errorf("checksum mismatch: wire=0x%02X computed=0x%02X", buf[1], checksum)
	}
	return nil
}

// ClearStamp zeroes bytes 1-3 after validation.
func ClearStamp(buf []byte) {
	buf[1] = 0
	buf[2] = 0
	buf[3] = 0
}
