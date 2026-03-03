// internal/network/handler.go
package network

import (
	"log/slog"

	"github.com/bjk/mwb/internal/input"
	"github.com/bjk/mwb/internal/protocol"
)

// MouseDevice is the interface for mouse injection.
type MouseDevice interface {
	MoveTo(x, y int32) error
	ButtonDown(button uint16) error
	ButtonUp(button uint16) error
	Wheel(delta int32) error
}

// KeyboardDevice is the interface for keyboard injection.
type KeyboardDevice interface {
	KeyDown(code uint16) error
	KeyUp(code uint16) error
}

// Handler processes incoming MWB packets and injects input events.
type Handler struct {
	Mouse    MouseDevice
	Keyboard KeyboardDevice
}

// HandlePacket dispatches a packet to the appropriate handler.
func (h *Handler) HandlePacket(pkt *protocol.Packet) {
	switch pkt.Type {
	case protocol.Mouse:
		h.handleMouse(pkt)
	case protocol.Keyboard:
		h.handleKeyboard(pkt)
	case protocol.Hello, protocol.Awake, protocol.HandshakeAck:
		// expected control packets — ignore silently
	default:
		slog.Debug("unhandled packet type", "type", pkt.Type)
	}
}

func (h *Handler) handleMouse(pkt *protocol.Packet) {
	md := pkt.Mouse
	var err error
	switch int(md.DwFlags) {
	case protocol.WM_MOUSEMOVE:
		err = h.Mouse.MoveTo(md.X, md.Y)
	case protocol.WM_LBUTTONDOWN:
		if err = h.Mouse.MoveTo(md.X, md.Y); err == nil {
			err = h.Mouse.ButtonDown(input.BTN_LEFT)
		}
	case protocol.WM_LBUTTONUP:
		if err = h.Mouse.MoveTo(md.X, md.Y); err == nil {
			err = h.Mouse.ButtonUp(input.BTN_LEFT)
		}
	case protocol.WM_RBUTTONDOWN:
		if err = h.Mouse.MoveTo(md.X, md.Y); err == nil {
			err = h.Mouse.ButtonDown(input.BTN_RIGHT)
		}
	case protocol.WM_RBUTTONUP:
		if err = h.Mouse.MoveTo(md.X, md.Y); err == nil {
			err = h.Mouse.ButtonUp(input.BTN_RIGHT)
		}
	case protocol.WM_MBUTTONDOWN:
		if err = h.Mouse.MoveTo(md.X, md.Y); err == nil {
			err = h.Mouse.ButtonDown(input.BTN_MIDDLE)
		}
	case protocol.WM_MBUTTONUP:
		if err = h.Mouse.MoveTo(md.X, md.Y); err == nil {
			err = h.Mouse.ButtonUp(input.BTN_MIDDLE)
		}
	case protocol.WM_MOUSEWHEEL:
		delta := md.WheelDelta / 120
		if delta == 0 && md.WheelDelta > 0 {
			delta = 1
		} else if delta == 0 && md.WheelDelta < 0 {
			delta = -1
		}
		err = h.Mouse.Wheel(delta)
	default:
		slog.Debug("unhandled mouse event", "flags", md.DwFlags)
		return
	}
	if err != nil {
		slog.Error("mouse input error", "err", err)
	}
}

func (h *Handler) handleKeyboard(pkt *protocol.Packet) {
	kd := pkt.Keyboard
	evdevCode, ok := input.VKToEvdev(kd.WVk)
	if !ok {
		slog.Debug("unknown VK code", "vk", kd.WVk)
		return
	}
	var err error
	isUp := (kd.DwFlags & protocol.LLKHF_UP) != 0
	if isUp {
		err = h.Keyboard.KeyUp(evdevCode)
	} else {
		err = h.Keyboard.KeyDown(evdevCode)
	}
	if err != nil {
		slog.Error("keyboard input error", "err", err)
	}
}
