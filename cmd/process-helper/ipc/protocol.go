package ipc

import "network-debugger/internal/features/process/domain"

// Request - IPC request from main app to helper daemon
type Request struct {
	ID     string      `json:"id"`     // request ID for matching responses
	Method string      `json:"method"` // "detect", "icon", "ping"
	Params RequestData `json:"params,omitempty"`
}

// RequestData - parameters for IPC requests
type RequestData struct {
	Port uint32 `json:"port,omitempty"` // for detect method
	PID  int32  `json:"pid,omitempty"`  // for icon method
	Path string `json:"path,omitempty"` // optional path for icon
}

// Response - IPC response from helper daemon to main app
type Response struct {
	ID     string      `json:"id"`               // matching request ID
	Result *ResultData `json:"result,omitempty"` // success result
	Error  *ErrorData  `json:"error,omitempty"`  // error if any
}

// ResultData - result payload for successful responses
type ResultData struct {
	// For "detect" method
	ProcessInfo *domain.ProcessInfo `json:"processInfo,omitempty"`

	// For "icon" method
	Icon *domain.AppIcon `json:"icon,omitempty"`

	// For "ping" method
	Pong string `json:"pong,omitempty"`
}

// ErrorData - error information
type ErrorData struct {
	Code    int    `json:"code"`    // error code (1=invalid request, 2=not found, 3=permission denied, 4=internal error)
	Message string `json:"message"` // human-readable error message
}

// Error codes
const (
	ErrInvalidRequest   = 1
	ErrNotFound         = 2
	ErrPermissionDenied = 3
	ErrInternalError    = 4
)
