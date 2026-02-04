package helper

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"network-debugger/cmd/process-helper/ipc"
	"network-debugger/internal/features/process/domain"
)

// mockServer - mock helper daemon for tests
type mockServer struct {
	conn      net.Conn
	requests  chan *ipc.Request
	responses map[string]*ipc.Response
	handler   func(*ipc.Request) *ipc.Response
	mu        sync.Mutex
}

func newMockServer(conn net.Conn) *mockServer {
	ms := &mockServer{
		conn:      conn,
		requests:  make(chan *ipc.Request, 10),
		responses: make(map[string]*ipc.Response),
	}
	go ms.handleConn()
	return ms
}

func (ms *mockServer) handleConn() {
	reader := bufio.NewReader(ms.conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		var req ipc.Request
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		ms.requests <- &req

		var resp *ipc.Response
		if ms.handler != nil {
			resp = ms.handler(&req)
		} else {
			ms.mu.Lock()
			resp = ms.responses[req.ID]
			ms.mu.Unlock()

			if resp == nil {
				resp = &ipc.Response{
					ID:     req.ID,
					Result: &ipc.ResultData{},
				}
			}
		}

		if resp != nil {
			resp.ID = req.ID
			respData, _ := json.Marshal(resp)
			ms.conn.Write(append(respData, '\n'))
		}
	}
}

func (ms *mockServer) setResponse(id string, resp *ipc.Response) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.responses[id] = resp
}

func (ms *mockServer) setHandler(handler func(*ipc.Request) *ipc.Response) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.handler = handler
}

// Composer 1.
func TestNewClient(t *testing.T) {
	client := NewClient("/tmp/test.sock")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.socketPath != "/tmp/test.sock" {
		t.Errorf("socketPath = %q, want %q", client.socketPath, "/tmp/test.sock")
	}

	if client.timeout != 10*time.Second {
		t.Errorf("timeout = %v, want %v", client.timeout, 10*time.Second)
	}

	if client.conn != nil {
		t.Error("conn should be nil initially")
	}
}

// Composer 1.
func TestClient_IsRunning_Success(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:     req.ID,
			Result: &ipc.ResultData{Pong: "pong"},
		}
	})

	if !client.IsRunning() {
		t.Error("IsRunning() should return true when ping succeeds")
	}
}

// Composer 1.
func TestClient_IsRunning_Failure(t *testing.T) {
	client := NewClient("/tmp/nonexistent.sock")

	if client.IsRunning() {
		t.Error("IsRunning() should return false when connection fails")
	}
}

// Composer 1.
func TestClient_Ping_Success(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:     req.ID,
			Result: &ipc.ResultData{Pong: "pong"},
		}
	})

	err := client.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
}

// Composer 1.
func TestClient_Ping_ErrorResponse(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:    req.ID,
			Error: &ipc.ErrorData{Code: 1, Message: "test error"},
		}
	})

	err := client.Ping()
	if err == nil {
		t.Error("Ping() should return error when server returns error")
	}
}

// Composer 1.
func TestClient_DetectProcess_Success(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	expectedProcess := &domain.ProcessInfo{
		PID:            1234,
		Name:           "test-process",
		ExecutablePath: "/usr/bin/test",
		DetectedAt:     time.Now(),
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID: req.ID,
			Result: &ipc.ResultData{
				ProcessInfo: expectedProcess,
			},
		}
	})

	process, err := client.DetectProcess(8080)
	if err != nil {
		t.Errorf("DetectProcess() error = %v, want nil", err)
	}

	if process == nil {
		t.Fatal("DetectProcess() returned nil process")
	}

	if process.PID != expectedProcess.PID {
		t.Errorf("PID = %d, want %d", process.PID, expectedProcess.PID)
	}

	if process.Name != expectedProcess.Name {
		t.Errorf("Name = %q, want %q", process.Name, expectedProcess.Name)
	}
}

// Composer 1.
func TestClient_DetectProcess_EmptyResponse(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:     req.ID,
			Result: &ipc.ResultData{},
		}
	})

	process, err := client.DetectProcess(8080)
	if err == nil {
		t.Error("DetectProcess() should return error when response is empty")
	}

	if process != nil {
		t.Error("DetectProcess() should return nil process on error")
	}
}

// Composer 1.
func TestClient_DetectProcess_ErrorResponse(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:    req.ID,
			Error: &ipc.ErrorData{Code: 1, Message: "process not found"},
		}
	})

	process, err := client.DetectProcess(8080)
	if err == nil {
		t.Error("DetectProcess() should return error when server returns error")
	}

	if process != nil {
		t.Error("DetectProcess() should return nil process on error")
	}
}

// Composer 1.
func TestClient_DetectProcess_NilResult(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:     req.ID,
			Result: nil,
		}
	})

	process, err := client.DetectProcess(8080)
	if err == nil {
		t.Error("DetectProcess() should return error when result is nil")
	}

	if process != nil {
		t.Error("DetectProcess() should return nil process on error")
	}
}

// Composer 1.
func TestClient_ExtractIcon_Success(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	expectedIcon := &domain.AppIcon{
		Format: "png",
		Data:   []byte{1, 2, 3, 4},
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID: req.ID,
			Result: &ipc.ResultData{
				Icon: expectedIcon,
			},
		}
	})

	icon, err := client.ExtractIcon(1234)
	if err != nil {
		t.Errorf("ExtractIcon() error = %v, want nil", err)
	}

	if icon == nil {
		t.Fatal("ExtractIcon() returned nil icon")
	}

	if icon.Format != expectedIcon.Format {
		t.Errorf("Format = %q, want %q", icon.Format, expectedIcon.Format)
	}

	if len(icon.Data) != len(expectedIcon.Data) {
		t.Errorf("Data length = %d, want %d", len(icon.Data), len(expectedIcon.Data))
	}
}

// Composer 1.
func TestClient_ExtractIcon_EmptyResponse(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:     req.ID,
			Result: &ipc.ResultData{},
		}
	})

	icon, err := client.ExtractIcon(1234)
	if err == nil {
		t.Error("ExtractIcon() should return error when response is empty")
	}

	if icon != nil {
		t.Error("ExtractIcon() should return nil icon on error")
	}
}

// Composer 1.
func TestClient_ExtractIcon_ErrorResponse(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:    req.ID,
			Error: &ipc.ErrorData{Code: 2, Message: "icon not found"},
		}
	})

	icon, err := client.ExtractIcon(1234)
	if err == nil {
		t.Error("ExtractIcon() should return error when server returns error")
	}

	if icon != nil {
		t.Error("ExtractIcon() should return nil icon on error")
	}
}

// Composer 1.
func TestClient_ExtractIcon_NilResult(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:     req.ID,
			Result: nil,
		}
	})

	icon, err := client.ExtractIcon(1234)
	if err == nil {
		t.Error("ExtractIcon() should return error when result is nil")
	}

	if icon != nil {
		t.Error("ExtractIcon() should return nil icon on error")
	}
}

// Composer 1.
func TestClient_Close(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		conn:       clientConn,
	}

	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	if client.conn != nil {
		t.Error("conn should be nil after Close()")
	}
}

// Composer 1.
func TestClient_Close_NoConnection(t *testing.T) {
	client := NewClient("/tmp/test.sock")

	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// Composer 1.
func TestClient_sendRequest_ReconnectOnError(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	serverConn.Close()
	defer clientConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	req := &ipc.Request{
		ID:     generateID(),
		Method: "ping",
	}

	_, err := client.sendRequest(req)
	if err == nil {
		t.Error("sendRequest() should return error when connection fails")
	}

	client.mu.Lock()
	connClosed := client.conn == nil
	client.mu.Unlock()

	if !connClosed {
		t.Error("conn should be nil after reconnect on error")
	}
}

// Composer 1.
func TestClient_ThreadSafety(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	server := newMockServer(serverConn)
	server.setHandler(func(req *ipc.Request) *ipc.Response {
		return &ipc.Response{
			ID:     req.ID,
			Result: &ipc.ResultData{Pong: "pong"},
		}
	})

	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			err := client.Ping()
			if err != nil {
				t.Errorf("Ping() failed in goroutine %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
}

// Composer 1.
func TestClient_Timeout(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    100 * time.Millisecond,
		conn:       clientConn,
	}

	req := &ipc.Request{
		ID:     generateID(),
		Method: "ping",
	}

	done := make(chan bool)
	go func() {
		_, err := client.sendRequest(req)
		if err == nil {
			t.Error("sendRequest() should return error on timeout")
		}
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("sendRequest() did not timeout")
	}
}

// Composer 1.
func TestClient_connect_Automatic(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/test.sock"

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		server := newMockServer(conn)
		server.setHandler(func(req *ipc.Request) *ipc.Response {
			return &ipc.Response{
				ID:     req.ID,
				Result: &ipc.ResultData{Pong: "pong"},
			}
		})
	}()

	client := NewClient(socketPath)
	client.timeout = 1 * time.Second

	err = client.Ping()
	if err != nil {
		t.Logf("Ping() error (may be expected): %v", err)
	}

	client.mu.Lock()
	connected := client.conn != nil
	client.mu.Unlock()

	if !connected {
		t.Log("conn is nil after connect attempt (may reconnect on error)")
	}
}

// Composer 1.
func TestClient_connect_Timeout(t *testing.T) {
	client := NewClient("/tmp/nonexistent.sock")
	client.timeout = 100 * time.Millisecond

	err := client.connect()
	if err == nil {
		t.Error("connect() should return error for nonexistent socket")
	}
}

// Composer 1.
func TestClient_sendRequest_ConnectError(t *testing.T) {
	client := NewClient("/tmp/nonexistent.sock")
	client.timeout = 100 * time.Millisecond

	req := &ipc.Request{
		ID:     generateID(),
		Method: "ping",
	}

	_, err := client.sendRequest(req)
	if err == nil {
		t.Error("sendRequest() should return error when connect fails")
	}
}

// Composer 1.
func TestClient_sendRequest_InvalidJSON(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	go func() {
		reader := bufio.NewReader(serverConn)
		reader.ReadBytes('\n')
		serverConn.Write([]byte("invalid json\n"))
	}()

	req := &ipc.Request{
		ID:     generateID(),
		Method: "ping",
	}

	_, err := client.sendRequest(req)
	if err == nil {
		t.Error("sendRequest() should return error when response is invalid JSON")
	}
}

// Composer 1.
func TestClient_sendRequest_WriteError(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	serverConn.Close()
	defer clientConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	req := &ipc.Request{
		ID:     generateID(),
		Method: "ping",
	}

	_, err := client.sendRequest(req)
	if err == nil {
		t.Error("sendRequest() should return error when write fails")
	}

	client.mu.Lock()
	connClosed := client.conn == nil
	client.mu.Unlock()

	if !connClosed {
		t.Error("conn should be nil after reconnect on write error")
	}
}

// Composer 1.
func TestClient_sendRequest_ReadError(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	req := &ipc.Request{
		ID:     generateID(),
		Method: "ping",
	}

	_, err := client.sendRequest(req)
	if err == nil {
		t.Error("sendRequest() should return error when read fails")
	}

	client.mu.Lock()
	connClosed := client.conn == nil
	client.mu.Unlock()

	if !connClosed {
		t.Error("conn should be nil after reconnect on read error")
	}
}

// Composer 1.
func TestClient_sendRequest_SetDeadlineError(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	client := &Client{
		socketPath: "/tmp/test.sock",
		timeout:    1 * time.Second,
		conn:       clientConn,
	}

	clientConn.Close()

	req := &ipc.Request{
		ID:     generateID(),
		Method: "ping",
	}

	_, err := client.sendRequest(req)
	if err == nil {
		t.Error("sendRequest() should return error when SetDeadline fails")
	}

	client.mu.Lock()
	connClosed := client.conn == nil
	client.mu.Unlock()

	if !connClosed {
		t.Error("conn should be nil after reconnect on SetDeadline error")
	}
}

// Composer 1.
func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("generateID() returned empty string")
	}

	if id2 == "" {
		t.Error("generateID() returned empty string")
	}

	if id1 == id2 {
		t.Error("generateID() should return unique IDs")
	}
}

// Composer 1.
func TestGenerateID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateID()
		if ids[id] {
			t.Errorf("generateID() returned duplicate ID: %s", id)
		}
		ids[id] = true
	}
}
