//go:build darwin
// +build darwin

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"

	"network-debugger/cmd/process-helper/ipc"
	"network-debugger/cmd/process-helper/server"
	processdetector "network-debugger/internal/features/process/infrastructure/detector"
	processicon "network-debugger/internal/features/process/infrastructure/icon"
)

const Version = "1.0.0"

func main() {
	// Create logger
	logger := zerolog.New(os.Stderr).With().
		Timestamp().
		Str("component", "process-helper").
		Logger()

	logger.Info().Str("version", Version).Msg("Starting network-debugger helper daemon")

	// Check that running with root privileges
	if os.Geteuid() != 0 {
		logger.Fatal().Msg("Helper daemon must be run as root (sudo)")
	}

	logger.Info().Msg("Running with root privileges")

	// Create detector (privileged mode)
	detector, err := processdetector.NewDetector(true)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create process detector")
	}

	// Create icon extractor
	extractor, err := processicon.NewExtractor()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create icon extractor")
	}

	// Create handler
	handler := server.NewHandler(detector, extractor, logger)

	// Create Unix socket listener
	listener, err := ipc.CreateListener()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create socket listener")
	}
	defer ipc.CleanupSocket()

	logger.Info().Str("socket", ipc.SocketPath).Msg("Unix socket created")

	// Create server
	srv := server.NewServer(listener, handler, logger)

	// Context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	// Buffer size 2 to not lose signals if they arrive simultaneously
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
	}()

	// Start server
	if err := srv.Serve(ctx); err != nil && err != context.Canceled {
		logger.Fatal().Err(err).Msg("Server error")
	}

	logger.Info().Msg("Helper daemon stopped")
}
