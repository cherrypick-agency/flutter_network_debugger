package dart

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
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

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(req.Script.Config.TimeoutMs)*time.Millisecond)
	defer cancel()

	start := time.Now()

	// Get process from pool
	proc, err := e.processPool.Get(execCtx)
	if err != nil {
		log.Printf("[Dart] Failed to get process from pool: %v", err)
		return domain.ExecutionResult{}, err
	}
	log.Printf("[Dart] Got process from pool")

	// Create JSON-RPC request
	code := req.Script.SourceCode
	if code == "" {
		code = string(req.Script.Code)
	}

	rpcReq := map[string]any{
		"jsonrpc": "2.0",
		"method":  "execute",
		"params": map[string]any{
			"code":  code,
			"input": string(req.Input),
		},
		"id": time.Now().UnixNano(),
	}

	// Send request
	log.Printf("[Dart] Sending RPC request: %+v", rpcReq)
	if err := json.NewEncoder(proc.stdin).Encode(rpcReq); err != nil {
		log.Printf("[Dart] Failed to send request: %v", err)
		e.processPool.Discard(proc)
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
	case <-execCtx.Done():
		log.Printf("[Dart] Execution timed out")
		e.processPool.Discard(proc)
		return domain.ExecutionResult{
			Duration: time.Since(start),
			Error:    "timeout",
		}, nil
	case resp := <-respChan:
		resp.result.Duration = time.Since(start)
		log.Printf("[Dart] Returning result: %+v, err=%v", resp.result, resp.err)
		if resp.err != nil {
			e.processPool.Discard(proc)
		} else {
			e.processPool.Release(proc)
		}
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

	// Skip validation if source code is empty
	if script.SourceCode == "" {
		return nil
	}

	// Create temporary file for validation
	tmpFile, err := os.CreateTemp("", "dart-validate-*.dart")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write source code to temp file
	if _, err := tmpFile.Write([]byte(script.SourceCode)); err != nil {
		return fmt.Errorf("failed to write source: %w", err)
	}
	tmpFile.Close()

	// Для синтаксической проверки достаточно `dart format`:
	// - не требует pubspec.yaml
	// - падает на ошибках парсинга
	analyzeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(analyzeCtx, "dart", "format", "--output", "none", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("dart validation failed: %s", string(output))
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "dart", "--version")
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
	closed           atomic.Bool
}

func (p *ProcessPool) decrementCount() {
	p.mu.Lock()
	if p.currentCount > 0 {
		p.currentCount--
	}
	p.mu.Unlock()
}

// DartProcess represents a running Dart VM process
type DartProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
}

func killAndWait(proc *DartProcess) {
	if proc == nil || proc.cmd == nil || proc.cmd.Process == nil {
		return
	}
	_ = proc.cmd.Process.Kill()
	// Важно: Wait нужен, иначе процесс может остаться зомби на unix.
	_ = proc.cmd.Wait()
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
	if p.closed.Load() {
		return nil, errors.New("process pool is closed")
	}

	// Try to get a process from the pool (non-blocking)
	select {
	case proc, ok := <-p.processes:
		if !ok {
			return nil, errors.New("process pool is closed")
		}
		if proc == nil {
			return nil, errors.New("process pool returned nil process")
		}
		return proc, nil
	default:
	}

	// No process available, try to create a new one if under limit
	p.mu.Lock()
	if p.closed.Load() {
		p.mu.Unlock()
		return nil, errors.New("process pool is closed")
	}
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
	case proc, ok := <-p.processes:
		if !ok {
			return nil, errors.New("process pool is closed")
		}
		if proc == nil {
			return nil, errors.New("process pool returned nil process")
		}
		return proc, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Release returns a process to the pool
func (p *ProcessPool) Release(proc *DartProcess) {
	if proc == nil {
		return
	}
	if p.closed.Load() {
		p.decrementCount()
		killAndWait(proc)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			// Канал могли закрыть на shutdown между Load() и send.
			p.decrementCount()
			killAndWait(proc)
		}
	}()
	select {
	case p.processes <- proc:
		// Successfully returned to pool
	default:
		// Pool full, kill process and decrement counter
		p.decrementCount()
		killAndWait(proc)
	}
}

// Discard kills a process and removes it from the pool accounting.
// Use this when the process may be in a bad state (timeout/cancel/protocol desync).
func (p *ProcessPool) Discard(proc *DartProcess) {
	if proc == nil {
		return
	}

	p.decrementCount()

	killAndWait(proc)
}

// Close shuts down all processes in the pool
func (p *ProcessPool) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(p.processes)
	for proc := range p.processes {
		p.decrementCount()
		killAndWait(proc)
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
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			log.Printf("[Dart script_runner stderr] %s", scanner.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		log.Printf("[Dart] Failed to start process: %v", err)
		return nil, err
	}

	log.Printf("[Dart] Process started successfully (PID: %d)", cmd.Process.Pid)

	proc := &DartProcess{
		cmd:   cmd,
		stdin: stdin,
		stdout: func() *bufio.Scanner {
			scanner := bufio.NewScanner(stdout)
			scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
			return scanner
		}(),
	}

	// Ping handshake: если dart run завис/не поднялся, не возвращаем процесс в пул.
	{
		pingReq := map[string]any{
			"jsonrpc": "2.0",
			"method":  "ping",
			"id":      time.Now().UnixNano(),
		}

		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := json.NewEncoder(proc.stdin).Encode(pingReq); err != nil {
			killAndWait(proc)
			return nil, fmt.Errorf("failed to send ping to Dart process: %w", err)
		}

		type pingResp struct {
			err error
		}
		ch := make(chan pingResp, 1)
		go func() {
			if !proc.stdout.Scan() {
				if err := proc.stdout.Err(); err != nil {
					ch <- pingResp{err: fmt.Errorf("failed to read ping response: %w", err)}
					return
				}
				ch <- pingResp{err: errors.New("no ping response")}
				return
			}
			var resp struct {
				Result any `json:"result"`
				Error  any `json:"error"`
			}
			if err := json.Unmarshal(proc.stdout.Bytes(), &resp); err != nil {
				ch <- pingResp{err: fmt.Errorf("invalid ping response: %w", err)}
				return
			}
			if resp.Error != nil {
				ch <- pingResp{err: errors.New("ping returned error")}
				return
			}
			ch <- pingResp{err: nil}
		}()

		select {
		case <-pingCtx.Done():
			killAndWait(proc)
			return nil, fmt.Errorf("dart process ping timeout: %w", pingCtx.Err())
		case r := <-ch:
			if r.err != nil {
				killAndWait(proc)
				return nil, r.err
			}
		}
	}

	return proc, nil
}

// readResponse reads a JSON-RPC response from the Dart process
func (p *DartProcess) readResponse() (domain.ExecutionResult, error) {
	if !p.stdout.Scan() {
		if err := p.stdout.Err(); err != nil {
			return domain.ExecutionResult{}, fmt.Errorf("failed to read response from Dart process: %w", err)
		}
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
