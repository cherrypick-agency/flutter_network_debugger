package helper

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"network-debugger/cmd/process-helper/ipc"
	"network-debugger/internal/features/process/domain"
)

// Client - IPC client for communicating with helper daemon
type Client struct {
	socketPath string
	timeout    time.Duration
	conn       net.Conn
	mu         sync.Mutex
}

// NewClient - create new IPC client
func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		timeout:    10 * time.Second, // 10 seconds timeout per request
	}
}

// IsRunning - check if helper daemon is available
func (c *Client) IsRunning() bool {
	err := c.Ping()
	return err == nil
}

// DetectProcess - detect process by port via helper
func (c *Client) DetectProcess(port uint32) (*domain.ProcessInfo, error) {
	req := &ipc.Request{
		ID:     generateID(),
		Method: "detect",
		Params: ipc.RequestData{
			Port: port,
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("helper error: %s", resp.Error.Message)
	}

	if resp.Result == nil || resp.Result.ProcessInfo == nil {
		return nil, fmt.Errorf("empty response from helper")
	}

	return resp.Result.ProcessInfo, nil
}

// ExtractIcon - extract icon via helper
func (c *Client) ExtractIcon(pid int32) (*domain.AppIcon, error) {
	req := &ipc.Request{
		ID:     generateID(),
		Method: "icon",
		Params: ipc.RequestData{
			PID: pid,
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("helper error: %s", resp.Error.Message)
	}

	if resp.Result == nil || resp.Result.Icon == nil {
		return nil, fmt.Errorf("empty response from helper")
	}

	return resp.Result.Icon, nil
}

// Ping - health check helper daemon
func (c *Client) Ping() error {
	req := &ipc.Request{
		ID:     generateID(),
		Method: "ping",
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("helper error: %s", resp.Error.Message)
	}

	return nil
}

// Close - close connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}

	return nil
}

// sendRequest - send request and receive response
func (c *Client) sendRequest(req *ipc.Request) (*ipc.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Connect if not yet connected
	if c.conn == nil {
		if err := c.connect(); err != nil {
			return nil, fmt.Errorf("failed to connect to helper: %w", err)
		}
	}

	// Set deadline
	deadline := time.Now().Add(c.timeout)
	if err := c.conn.SetDeadline(deadline); err != nil {
		c.reconnect()
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Send request
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	writer := bufio.NewWriter(c.conn)
	if _, err := writer.Write(reqData); err != nil {
		c.reconnect()
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	if err := writer.WriteByte('\n'); err != nil {
		c.reconnect()
		return nil, fmt.Errorf("failed to write newline: %w", err)
	}
	if err := writer.Flush(); err != nil {
		c.reconnect()
		return nil, fmt.Errorf("failed to flush: %w", err)
	}

	// Receive response
	reader := bufio.NewReader(c.conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		c.reconnect()
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp ipc.Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// connect - connect to Unix socket
func (c *Client) connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

// reconnect - reconnect (after error)
// IMPORTANT: must only be called when c.mu is locked
func (c *Client) reconnect() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

// generateID - generate unique ID for request (thread-safe)
func generateID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback to timestamp if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	randomPart := binary.BigEndian.Uint64(b[:])
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), randomPart)
}
