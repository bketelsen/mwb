// cmd/mwb/main.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bjk/mwb/internal/config"
	"github.com/bjk/mwb/internal/input"
	"github.com/bjk/mwb/internal/network"
)

func main() {
	configPath := flag.String("config", "", "path to config.toml")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	if *configPath == "" {
		home, _ := os.UserHomeDir()
		*configPath = filepath.Join(home, ".config", "mwb", "config.toml")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nCreate config at %s with:\n\n", *configPath)
		fmt.Fprintf(os.Stderr, "  host = \"192.168.1.100\"\n  key = \"YourSecurityKey\"\n  name = \"linux\"\n\n")
		os.Exit(1)
	}

	slog.Info("mwb starting", "host", cfg.Host, "port", cfg.MessagePort(), "name", cfg.Name)

	mouse, err := input.CreateVirtualMouse("mwb-mouse")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating virtual mouse: %v\n", err)
		fmt.Fprintf(os.Stderr, "Ensure your user is in the 'input' group: sudo usermod -aG input $USER\n")
		os.Exit(1)
	}
	defer mouse.Close()

	keyboard, err := input.CreateVirtualKeyboard("mwb-keyboard")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating virtual keyboard: %v\n", err)
		os.Exit(1)
	}
	defer keyboard.Close()

	slog.Info("virtual input devices created")

	handler := &network.Handler{
		Mouse:    mouse,
		Keyboard: keyboard,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		backoff := 1 * time.Second
		maxBackoff := 30 * time.Second

		for {
			addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.MessagePort())
			slog.Info("connecting", "addr", addr)

			conn, err := network.Connect(addr, cfg.Key, cfg.Name, 10*time.Second)
			if err != nil {
				slog.Error("connection failed", "err", err, "retry_in", backoff)
				time.Sleep(backoff)
				backoff = min(backoff*2, maxBackoff)
				continue
			}

			slog.Info("connected", "remote", conn.RemoteName)
			backoff = 1 * time.Second

			if err := network.ReceiveLoop(conn, handler); err != nil {
				slog.Error("receive loop error", "err", err)
			}

			conn.Close()
			slog.Info("disconnected, will reconnect", "in", backoff)
			time.Sleep(backoff)
		}
	}()

	sig := <-sigCh
	slog.Info("shutting down", "signal", sig)
}
