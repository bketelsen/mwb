// internal/network/client_test.go
package network

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
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

func TestSetupTimesOutWhenPeerStopsReading(t *testing.T) {
	oldTimeout := setupTimeout
	setupTimeout = 1 * time.Second
	t.Cleanup(func() { setupTimeout = oldTimeout })

	raw, peer := net.Pipe()
	t.Cleanup(func() { _ = raw.Close() })
	t.Cleanup(func() { _ = peer.Close() })

	const securityKey = "TestSecurityKey!!"
	var preamble bytes.Buffer
	enc, err := protocol.NewSchemeEncryptWriter(&preamble, securityKey, protocol.SchemeLegacy)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := enc.Write(make([]byte, 16)); err != nil {
		t.Fatal(err)
	}
	magic := protocol.Get24BitHash(securityKey)
	for i := 0; i < 10; i++ {
		hs := &protocol.Packet{Type: protocol.Handshake, ID: int32(i + 1), Src: 7}
		hs.SetMachineName("WINHOST")
		buf := hs.Marshal()
		protocol.StampPacket(buf, magic)
		if _, err := enc.Write(buf); err != nil {
			t.Fatal(err)
		}
	}
	go func() { _, _ = peer.Write(preamble.Bytes()) }()

	result := make(chan error, 1)
	go func() {
		_, err := setup(raw, securityKey, "linux-test")
		result <- err
	}()
	select {
	case err := <-result:
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("setup error = %v, want deadline exceeded", err)
		}
		if !strings.Contains(err.Error(), "send IV block") {
			t.Fatalf("setup error = %v, want the send IV block path", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("setup did not time out within 5 seconds")
	}
}

func TestSetupTimesOutWhenPeerSendsNothing(t *testing.T) {
	oldTimeout := setupTimeout
	setupTimeout = 200 * time.Millisecond
	t.Cleanup(func() { setupTimeout = oldTimeout })

	raw, peer := net.Pipe()
	t.Cleanup(func() { _ = raw.Close() })
	t.Cleanup(func() { _ = peer.Close() })
	go func() { _, _ = io.Copy(io.Discard, peer) }()

	result := make(chan error, 1)
	go func() {
		_, err := setup(raw, "TestSecurityKey!!", "linux-test")
		result <- err
	}()
	select {
	case err := <-result:
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("setup error = %v, want deadline exceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("setup did not time out within 2 seconds")
	}
}

// queuedSetupPeer provides the write buffering that runFakeServer expects from
// a TCP socket. The underlying net.Pipe still blocks reads and writes normally.
// lastWrite preserves the CBC chaining block for a packet after runFakeServer.
type queuedSetupPeer struct {
	net.Conn
	writes    chan []byte
	done      <-chan struct{}
	lastWrite []byte
}

func (p *queuedSetupPeer) Write(buf []byte) (int, error) {
	data := bytes.Clone(buf)
	select {
	case p.writes <- data:
		p.lastWrite = data
		return len(buf), nil
	case <-p.done:
		return 0, net.ErrClosed
	}
}

type deadlineRecordingConn struct {
	net.Conn
	mu        sync.Mutex
	deadlines []time.Time
}

func (c *deadlineRecordingConn) SetDeadline(deadline time.Time) error {
	c.mu.Lock()
	c.deadlines = append(c.deadlines, deadline)
	c.mu.Unlock()
	return c.Conn.SetDeadline(deadline)
}

func TestSetupClearsDeadline(t *testing.T) {
	raw, peer := net.Pipe()
	recorded := &deadlineRecordingConn{Conn: raw}
	t.Cleanup(func() { _ = raw.Close() })
	t.Cleanup(func() { _ = peer.Close() })
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })
	server := &queuedSetupPeer{Conn: peer, writes: make(chan []byte, 16), done: done}
	go func() {
		for {
			select {
			case buf := <-server.writes:
				if _, err := peer.Write(buf); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	const securityKey = "TestSecurityKey!!"
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- runFakeServer(server, securityKey, protocol.SchemeLegacy)
		_, _ = io.Copy(io.Discard, peer)
	}()

	c, err := setup(recorded, securityKey, "linux-test")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := <-serverDone; err != nil {
		t.Fatalf("server: %v", err)
	}

	recorded.mu.Lock()
	deadlines := append([]time.Time(nil), recorded.deadlines...)
	recorded.mu.Unlock()
	if len(deadlines) < 2 {
		t.Fatalf("SetDeadline calls = %d, want at least 2", len(deadlines))
	}
	if deadlines[0].IsZero() {
		t.Fatal("first SetDeadline argument is zero, want a setup deadline")
	}
	if !deadlines[len(deadlines)-1].IsZero() {
		t.Fatalf("last SetDeadline argument = %v, want zero", deadlines[len(deadlines)-1])
	}

	// Continue the legacy CBC stream from the server's final ciphertext block.
	iv := server.lastWrite[len(server.lastWrite)-16:]
	enc, err := protocol.NewEncryptWriter(server, protocol.DeriveKey(securityKey), iv)
	if err != nil {
		t.Fatal(err)
	}
	pkt := &protocol.Packet{Type: protocol.HeartbeatEx, ID: 12, Src: 7}
	buf := pkt.Marshal()
	protocol.StampPacket(buf, protocol.Get24BitHash(securityKey))
	if _, err := enc.Write(buf); err != nil {
		t.Fatal(err)
	}
	got, err := c.RecvPacket()
	if err != nil {
		t.Fatalf("RecvPacket after setup deadline: %v", err)
	}
	if got.Type != pkt.Type || got.ID != pkt.ID {
		t.Fatalf("packet = %+v, want type %d and ID %d", got, pkt.Type, pkt.ID)
	}
}
