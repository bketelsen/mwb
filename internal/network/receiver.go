// internal/network/receiver.go
package network

import (
	"errors"
	"io"
	"log/slog"

	"github.com/bjk/mwb/internal/protocol"
)

// ReceiveLoop reads packets from the connection and dispatches them.
func ReceiveLoop(conn *Conn, handler *Handler) error {
	for {
		pkt, err := conn.RecvPacket()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				slog.Info("connection closed by remote")
				return nil
			}
			return err
		}

		switch pkt.Type {
		case protocol.Heartbeat, protocol.HeartbeatEx, protocol.HeartbeatExL2, protocol.HeartbeatExL3:
			slog.Debug("heartbeat received", "type", pkt.Type, "from", pkt.MachineName())
			resp := &protocol.Packet{
				Type: pkt.Type,
				Src:  conn.MachineID,
				Des:  pkt.Src,
			}
			resp.SetMachineName(conn.RemoteName)
			if err := conn.SendPacket(resp); err != nil {
				slog.Error("send heartbeat response", "err", err)
			}
		case protocol.ByeBye:
			slog.Info("remote disconnected (ByeBye)")
			return nil
		case protocol.Invalid:
			slog.Warn("invalid packet received")
		case protocol.Handshake:
			slog.Debug("late handshake packet, ignoring")
		default:
			handler.HandlePacket(pkt)
		}
	}
}
