// internal/network/client_test.go
package network

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/bketelsen/mwb/internal/protocol"
)

// newServerDecryptReader reads the client's stream preamble for scheme and
// returns a reader that decrypts the rest of the client's stream.
func newServerDecryptReader(conn net.Conn, securityKey string, scheme protocol.Scheme) (*protocol.DecryptReader, error) {
	if !scheme.HasHeader() {
		return protocol.NewDecryptReader(conn, protocol.DeriveKey(securityKey), protocol.FixedIV())
	}
	header := make([]byte, protocol.HeaderSize)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("read client header: %w", err)
	}
	salt, iv := header[:16], header[16:]
	return protocol.NewDecryptReader(conn, protocol.DeriveSchemeKey(securityKey, salt, scheme), iv)
}

// runFakeServer behaves like a Windows MWB server: it sends its preamble and
// handshakes before reading anything, then acknowledges the client's handshake.
func runFakeServer(conn net.Conn, securityKey string, scheme protocol.Scheme) error {
	magic := protocol.Get24BitHash(securityKey)

	enc, err := protocol.NewSchemeEncryptWriter(conn, securityKey, scheme)
	if err != nil {
		return err
	}

	// Server: send random IV block
	ranData := make([]byte, 16)
	_, _ = rand.Read(ranData)
	if _, err := enc.Write(ranData); err != nil {
		return err
	}

	// Server: send its own handshakes, as MainTCPRoutine does on connect
	for i := 0; i < 10; i++ {
		hs := &protocol.Packet{Type: protocol.Handshake, ID: int32(i + 1), Src: 7}
		hs.SetMachineName("WINHOST")
		buf := hs.Marshal()
		protocol.StampPacket(buf, magic)
		if _, err := enc.Write(buf); err != nil {
			return err
		}
	}

	dec, err := newServerDecryptReader(conn, securityKey, scheme)
	if err != nil {
		return err
	}

	// Server: read random IV block from client
	clientRan := make([]byte, 16)
	if _, err := io.ReadFull(dec, clientRan); err != nil {
		return err
	}

	// Server: read the first handshake packet from the client
	pktBuf := make([]byte, protocol.PacketSizeEx)
	if _, err := io.ReadFull(dec, pktBuf); err != nil {
		return err
	}
	if err := protocol.ValidatePacket(pktBuf, magic); err != nil {
		return err
	}
	protocol.ClearStamp(pktBuf)

	pkt, err := protocol.UnmarshalPacket(pktBuf)
	if err != nil {
		return err
	}
	if pkt.Type != protocol.Handshake {
		return fmt.Errorf("expected Handshake, got %d", pkt.Type)
	}

	// Server: send HandshakeAck with inverted machine fields
	ack := &protocol.Packet{
		Type: protocol.HandshakeAck,
		Src:  7,
		Des:  pkt.Src,
	}
	ack.Handshake.Machine1 = ^pkt.Handshake.Machine1
	ack.Handshake.Machine2 = ^pkt.Handshake.Machine2
	ack.Handshake.Machine3 = ^pkt.Handshake.Machine3
	ack.Handshake.Machine4 = ^pkt.Handshake.Machine4
	ack.SetMachineName("WINHOST")

	ackBuf := ack.Marshal()
	protocol.StampPacket(ackBuf, magic)
	_, err = enc.Write(ackBuf)
	return err
}

func TestConnectionHandshake(t *testing.T) {
	securityKey := "TestSecurityKey!!"

	for _, scheme := range []protocol.Scheme{protocol.SchemeLegacy, protocol.SchemeHeader50k, protocol.SchemeHeader100k} {
		t.Run(scheme.String(), func(t *testing.T) {
			// Start a fake MWB server
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = ln.Close() }()

			serverDone := make(chan error, 1)
			go func() {
				conn, err := ln.Accept()
				if err != nil {
					serverDone <- err
					return
				}
				// Keep the connection open until the client is done with it
				serverDone <- runFakeServer(conn, securityKey, scheme)
				time.Sleep(500 * time.Millisecond)
				_ = conn.Close()
			}()

			// Client: connect and handshake
			conn, err := Connect(ln.Addr().String(), securityKey, "linux-test", 5*time.Second)
			if err != nil {
				t.Fatalf("Connect: %v", err)
			}
			defer func() { _ = conn.Close() }()

			if conn.RemoteName != "WINHOST" {
				t.Errorf("remote name = %q, want %q", conn.RemoteName, "WINHOST")
			}

			select {
			case err := <-serverDone:
				if err != nil {
					t.Fatalf("server error: %v", err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("timeout waiting for server")
			}
		})
	}
}

func TestConnectWrongKey(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_ = runFakeServer(conn, "ServerSecurityKey", protocol.SchemeHeader100k)
	}()

	_, err = Connect(ln.Addr().String(), "ClientSecurityKey", "linux-test", 5*time.Second)
	if err == nil {
		t.Fatal("Connect succeeded with mismatched security key")
	}
}
