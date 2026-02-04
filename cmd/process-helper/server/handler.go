package server

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"network-debugger/cmd/process-helper/ipc"
	"network-debugger/internal/features/process/domain"
)

const (
	// RequestTimeout - maximum time to process a single request
	RequestTimeout = 30 * time.Second
)

// Handler - IPC request handler
type Handler struct {
	detector  domain.IProcessDetector
	extractor domain.IIconExtractor
	logger    zerolog.Logger
}

// NewHandler - create new handler
func NewHandler(detector domain.IProcessDetector, extractor domain.IIconExtractor, logger zerolog.Logger) *Handler {
	return &Handler{
		detector:  detector,
		extractor: extractor,
		logger:    logger,
	}
}

// Handle - process request and return response
func (h *Handler) Handle(req *ipc.Request) *ipc.Response {
	switch req.Method {
	case "detect":
		return h.handleDetect(req)
	case "icon":
		return h.handleIcon(req)
	case "ping":
		return h.handlePing(req)
	default:
		return &ipc.Response{
			ID: req.ID,
			Error: &ipc.ErrorData{
				Code:    ipc.ErrInvalidRequest,
				Message: fmt.Sprintf("unknown method: %s", req.Method),
			},
		}
	}
}

// handleDetect - detect process by port
func (h *Handler) handleDetect(req *ipc.Request) *ipc.Response {
	port := req.Params.Port
	if port == 0 {
		return &ipc.Response{
			ID: req.ID,
			Error: &ipc.ErrorData{
				Code:    ipc.ErrInvalidRequest,
				Message: "port is required",
			},
		}
	}

	h.logger.Debug().Uint32("port", port).Msg("Detecting process")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	processInfo, err := h.detector.DetectByPort(ctx, port)
	if err != nil {
		h.logger.Warn().Err(err).Uint32("port", port).Msg("Process detection failed")
		return &ipc.Response{
			ID: req.ID,
			Error: &ipc.ErrorData{
				Code:    ipc.ErrNotFound,
				Message: fmt.Sprintf("process not found for port %d: %v", port, err),
			},
		}
	}

	h.logger.Info().
		Uint32("port", port).
		Int32("pid", processInfo.PID).
		Str("name", processInfo.Name).
		Msg("Process detected")

	return &ipc.Response{
		ID: req.ID,
		Result: &ipc.ResultData{
			ProcessInfo: processInfo,
		},
	}
}

// handleIcon - extract process icon
func (h *Handler) handleIcon(req *ipc.Request) *ipc.Response {
	pid := req.Params.PID
	path := req.Params.Path

	if pid == 0 && path == "" {
		return &ipc.Response{
			ID: req.ID,
			Error: &ipc.ErrorData{
				Code:    ipc.ErrInvalidRequest,
				Message: "pid or path is required",
			},
		}
	}

	h.logger.Debug().Int32("pid", pid).Str("path", path).Msg("Extracting icon")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	var icon *domain.AppIcon
	var err error

	if path != "" {
		icon, err = h.extractor.ExtractByPath(ctx, path)
	} else {
		icon, err = h.extractor.ExtractByPID(ctx, pid)
	}

	if err != nil {
		h.logger.Warn().Err(err).Int32("pid", pid).Msg("Icon extraction failed")
		return &ipc.Response{
			ID: req.ID,
			Error: &ipc.ErrorData{
				Code:    ipc.ErrNotFound,
				Message: fmt.Sprintf("icon extraction failed: %v", err),
			},
		}
	}

	h.logger.Info().Int32("pid", pid).Str("format", icon.Format).Msg("Icon extracted")

	return &ipc.Response{
		ID: req.ID,
		Result: &ipc.ResultData{
			Icon: icon,
		},
	}
}

// handlePing - health check
func (h *Handler) handlePing(req *ipc.Request) *ipc.Response {
	h.logger.Debug().Msg("Ping received")

	return &ipc.Response{
		ID: req.ID,
		Result: &ipc.ResultData{
			Pong: "pong",
		},
	}
}
