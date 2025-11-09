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
	// Создать logger
	logger := zerolog.New(os.Stderr).With().
		Timestamp().
		Str("component", "process-helper").
		Logger()

	logger.Info().Str("version", Version).Msg("Starting network-debugger helper daemon")

	// Проверить что запущен с root правами
	if os.Geteuid() != 0 {
		logger.Fatal().Msg("Helper daemon must be run as root (sudo)")
	}

	logger.Info().Msg("Running with root privileges")

	// Создать detector (привилегированный режим)
	detector, err := processdetector.NewDetector(true)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create process detector")
	}

	// Создать icon extractor
	extractor, err := processicon.NewExtractor()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create icon extractor")
	}

	// Создать handler
	handler := server.NewHandler(detector, extractor, logger)

	// Создать Unix socket listener
	listener, err := ipc.CreateListener()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create socket listener")
	}
	defer ipc.CleanupSocket()

	logger.Info().Str("socket", ipc.SocketPath).Msg("Unix socket created")

	// Создать server
	srv := server.NewServer(listener, handler, logger)

	// Context с cancellation для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка shutdown signals
	// Буфер размером 2 чтобы не потерять сигналы если приходят одновременно
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
	}()

	// Запустить server
	if err := srv.Serve(ctx); err != nil && err != context.Canceled {
		logger.Fatal().Err(err).Msg("Server error")
	}

	logger.Info().Msg("Helper daemon stopped")
}
