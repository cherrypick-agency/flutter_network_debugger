package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/rs/zerolog"

	"network-debugger/cmd/process-helper/ipc"
)

const (
	// MaxConcurrentConnections - maximum concurrent connections
	MaxConcurrentConnections = 10
	// MaxRequestSize - maximum size of a single request in bytes (1 MB)
	MaxRequestSize = 1024 * 1024
)

// Server - IPC server for handling requests from main app
type Server struct {
	listener net.Listener
	handler  *Handler
	logger   zerolog.Logger
	wg       sync.WaitGroup
	connSem  chan struct{} // semaphore for limiting concurrent connections
}

// NewServer - create new server
func NewServer(listener net.Listener, handler *Handler, logger zerolog.Logger) *Server {
	return &Server{
		listener: listener,
		handler:  handler,
		logger:   logger,
		connSem:  make(chan struct{}, MaxConcurrentConnections),
	}
}

// Serve - start accept loop
func (s *Server) Serve(ctx context.Context) error {
	s.logger.Info().Msg("Helper daemon server started, waiting for connections")

	// Create errCh for handling errors from accept()
	errCh := make(chan error, 1)

	// Start accept loop in separate goroutine
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				// Check if this is shutdown
				select {
				case <-ctx.Done():
					errCh <- nil
					return
				default:
					s.logger.Error().Err(err).Msg("Failed to accept connection")
					continue
				}
			}

			// Try to acquire semaphore (non-blocking)
			select {
			case s.connSem <- struct{}{}:
				// Successfully acquired slot, start goroutine
				s.wg.Add(1)
				go s.handleConnection(conn)
			case <-ctx.Done():
				// Shutdown while waiting for slot
				conn.Close()
				errCh <- nil
				return
			default:
				// No available slots - reject connection
				s.logger.Warn().Msg("Max concurrent connections reached, rejecting connection")
				conn.Close()
			}
		}
	}()

	// Wait for shutdown
	select {
	case <-ctx.Done():
		s.logger.Info().Msg("Shutting down server, waiting for active connections...")
		// Close listener to interrupt Accept()
		s.listener.Close()
		// Wait for all goroutines to complete
		s.wg.Wait()
		s.logger.Info().Msg("All connections closed, server stopped")
		return ctx.Err()
	case err := <-errCh:
		s.logger.Info().Msg("Accept loop ended")
		s.wg.Wait()
		return err
	}
}

// handleConnection - handle single connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		<-s.connSem // release slot in semaphore
		s.wg.Done()
	}()

	s.logger.Debug().Str("remote", conn.RemoteAddr().String()).Msg("Connection accepted")

	scanner := bufio.NewScanner(conn)
	// Set maximum buffer size for OOM protection
	scanner.Buffer(make([]byte, 4096), MaxRequestSize)
	writer := bufio.NewWriter(conn)

	for scanner.Scan() {
		line := scanner.Bytes()

		// Parse request
		var req ipc.Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.logger.Error().Err(err).Msg("Failed to parse request")
			s.sendError(writer, "", ipc.ErrInvalidRequest, "Invalid JSON request")
			continue
		}

		// Validate request ID
		if req.ID == "" {
			s.logger.Error().Msg("Request ID is empty")
			s.sendError(writer, "", ipc.ErrInvalidRequest, "Request ID is required")
			continue
		}

		s.logger.Debug().Str("id", req.ID).Str("method", req.Method).Msg("Received request")

		// Handle request
		resp := s.handler.Handle(&req)

		// Send response
		if err := s.sendResponse(writer, resp); err != nil {
			s.logger.Error().Err(err).Msg("Failed to send response")
		}
	}

	// Check reason for scan completion
	if err := scanner.Err(); err != nil {
		s.logger.Error().Err(err).Msg("Scanner error during connection")
	} else {
		s.logger.Debug().Msg("Connection closed by client (EOF)")
	}
}

// sendResponse - send response to connection
func (s *Server) sendResponse(w *bufio.Writer, resp *ipc.Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	if err := w.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	return nil
}

// sendError - send error response
func (s *Server) sendError(w *bufio.Writer, reqID string, code int, message string) {
	resp := &ipc.Response{
		ID: reqID,
		Error: &ipc.ErrorData{
			Code:    code,
			Message: message,
		},
	}
	if err := s.sendResponse(w, resp); err != nil {
		s.logger.Error().Err(err).Str("reqID", reqID).Msg("Failed to send error response")
	}
}
