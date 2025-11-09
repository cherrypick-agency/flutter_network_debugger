package dart

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// DartExecutor implements domain.ScriptExecutor for Dart scripts
// Uses subprocess communication via JSON-RPC 2.0 over stdin/stdout
type DartExecutor struct {
	processPool      *ProcessPool
	scriptRunnerPath string
	enabled          bool
}

// NewDartExecutor creates a new Dart executor with process pooling
// scriptRunnerPath should point to the Dart script runner (e.g., "scripts/dart/script_runner.dart")
func NewDartExecutor(maxProcesses int, scriptRunnerPath string) (*DartExecutor, error) {
	// Check if Dart is available
	if !isDartAvailable() {
		log.Printf("[Dart] Dart SDK not found in PATH. Dart script support disabled.")
		return &DartExecutor{
			enabled: false,
		}, nil
	}

	pool, err := NewProcessPool(maxProcesses, scriptRunnerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Dart process pool: %w", err)
	}

	log.Printf("[Dart] Initialized with %d processes", maxProcesses)

	return &DartExecutor{
		processPool:      pool,
		scriptRunnerPath: scriptRunnerPath,
		enabled:          true,
	}, nil
}

// Execute runs a Dart script with JSON-RPC communication
func (e *DartExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	log.Printf("[Dart] Execute called for script: %s", req.Script.Name)

	if !e.enabled {
		log.Printf("[Dart] Executor disabled, Dart SDK not available")
		return domain.ExecutionResult{
			Error: "Dart runtime not available. Install Dart SDK to enable Dart scripts.",
		}, nil
	}

	start := time.Now()

	// Get process from pool
	proc, err := e.processPool.Get(ctx)
	if err != nil {
		log.Printf("[Dart] Failed to get process from pool: %v", err)
		return domain.ExecutionResult{}, err
	}
	defer e.processPool.Release(proc)
	log.Printf("[Dart] Got process from pool")

	// Create JSON-RPC request
	rpcReq := map[string]any{
		"jsonrpc": "2.0",
		"method":  "execute",
		"params": map[string]any{
			"code":  string(req.Script.Code),
			"input": string(req.Input),
		},
		"id": time.Now().UnixNano(),
	}

	// Send request
	log.Printf("[Dart] Sending RPC request: %+v", rpcReq)
	if err := json.NewEncoder(proc.stdin).Encode(rpcReq); err != nil {
		log.Printf("[Dart] Failed to send request: %v", err)
		return domain.ExecutionResult{}, fmt.Errorf("failed to send request: %w", err)
	}
	log.Printf("[Dart] Request sent, waiting for response")

	// Read response with timeout
	type response struct {
		result domain.ExecutionResult
		err    error
	}

	respChan := make(chan response, 1)
	go func() {
		result, err := proc.readResponse()
		log.Printf("[Dart] Read response: result=%+v, err=%v", result, err)
		respChan <- response{result, err}
	}()

	select {
	case <-ctx.Done():
		log.Printf("[Dart] Execution timed out")
		return domain.ExecutionResult{
			Duration: time.Since(start),
			Error:    "timeout",
		}, nil
	case resp := <-respChan:
		resp.result.Duration = time.Since(start)
		log.Printf("[Dart] Returning result: %+v, err=%v", resp.result, resp.err)
		return resp.result, resp.err
	}
}

// Runtime returns the runtime type
func (e *DartExecutor) Runtime() domain.ScriptRuntime {
	return domain.RuntimeDart
}

// Validate checks if the Dart script is valid
func (e *DartExecutor) Validate(ctx context.Context, script domain.Script) error {
	if !e.enabled {
		return errors.New("Dart runtime not available")
	}
	// TODO: Implement Dart syntax validation
	return nil
}

// Close cleans up the process pool
func (e *DartExecutor) Close() error {
	if e.processPool != nil {
		return e.processPool.Close()
	}
	return nil
}

// isDartAvailable checks if Dart SDK is installed
func isDartAvailable() bool {
	cmd := exec.Command("dart", "--version")
	err := cmd.Run()
	return err == nil
}

// ProcessPool manages a pool of Dart VM processes for reuse
type ProcessPool struct {
	processes        chan *DartProcess
	maxSize          int
	scriptRunnerPath string
	currentCount     int // Number of currently active processes
	mu               sync.Mutex
}

// DartProcess represents a running Dart VM process
type DartProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
}

// NewProcessPool creates a new pool of Dart processes
// Processes are created lazily on-demand, not pre-spawned
func NewProcessPool(maxSize int, scriptRunnerPath string) (*ProcessPool, error) {
	pool := &ProcessPool{
		processes:        make(chan *DartProcess, maxSize),
		maxSize:          maxSize,
		scriptRunnerPath: scriptRunnerPath,
		currentCount:     0, // Start with no processes
	}

	return pool, nil
}

// Get retrieves a process from the pool, creating one lazily if needed
func (p *ProcessPool) Get(ctx context.Context) (*DartProcess, error) {
	// Try to get a process from the pool (non-blocking)
	select {
	case proc := <-p.processes:
		return proc, nil
	default:
	}

	// No process available, try to create a new one if under limit
	p.mu.Lock()
	if p.currentCount < p.maxSize {
		p.currentCount++
		p.mu.Unlock()

		proc, err := p.startProcess()
		if err != nil {
			// Failed to create process, decrement counter
			p.mu.Lock()
			p.currentCount--
			p.mu.Unlock()
			return nil, err
		}
		return proc, nil
	}
	p.mu.Unlock()

	// Pool exhausted, wait for a process to be returned
	select {
	case proc := <-p.processes:
		return proc, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Release returns a process to the pool
func (p *ProcessPool) Release(proc *DartProcess) {
	select {
	case p.processes <- proc:
		// Successfully returned to pool
	default:
		// Pool full, kill process and decrement counter
		p.mu.Lock()
		p.currentCount--
		p.mu.Unlock()
		if proc.cmd != nil && proc.cmd.Process != nil {
			proc.cmd.Process.Kill()
		}
	}
}

// Close shuts down all processes in the pool
func (p *ProcessPool) Close() error {
	close(p.processes)
	for proc := range p.processes {
		if proc.cmd != nil && proc.cmd.Process != nil {
			proc.cmd.Process.Kill()
		}
	}
	return nil
}

// startProcess spawns a new Dart VM process
func (p *ProcessPool) startProcess() (*DartProcess, error) {
	log.Printf("[Dart] Starting process with script_runner: %s", p.scriptRunnerPath)
	cmd := exec.Command("dart", "run", p.scriptRunnerPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("[Dart] Failed to create stdin pipe: %v", err)
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[Dart] Failed to create stdout pipe: %v", err)
		return nil, err
	}

	// Capture stderr for debugging
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("[Dart] Failed to create stderr pipe: %v", err)
		return nil, err
	}

	// Read stderr in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[Dart script_runner stderr] %s", scanner.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		log.Printf("[Dart] Failed to start process: %v", err)
		return nil, err
	}

	log.Printf("[Dart] Process started successfully (PID: %d)", cmd.Process.Pid)

	return &DartProcess{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
	}, nil
}

// readResponse reads a JSON-RPC response from the Dart process
func (p *DartProcess) readResponse() (domain.ExecutionResult, error) {
	if !p.stdout.Scan() {
		return domain.ExecutionResult{}, errors.New("no response from Dart process")
	}

	var rpcResp struct {
		Result struct {
			Output string   `json:"output"`
			Logs   []string `json:"logs"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(p.stdout.Bytes(), &rpcResp); err != nil {
		return domain.ExecutionResult{}, err
	}

	if rpcResp.Error != nil {
		return domain.ExecutionResult{
			Error: rpcResp.Error.Message,
		}, nil
	}

	return domain.ExecutionResult{
		Output: []byte(rpcResp.Result.Output),
		Logs:   rpcResp.Result.Logs,
	}, nil
}
