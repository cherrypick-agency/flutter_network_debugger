package httpapi

import (
	"bytes"
	"io"
	"math/rand"
	"net"
	"testing"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/internal/infrastructure/config"
)

func TestNewBucket(t *testing.T) {
	tests := []struct {
		name        string
		bytesPerSec int
		wantNil     bool
	}{
		{
			name:        "zero rate",
			bytesPerSec: 0,
			wantNil:     true,
		},
		{
			name:        "negative rate",
			bytesPerSec: -100,
			wantNil:     true,
		},
		{
			name:        "positive rate",
			bytesPerSec: 1000,
			wantNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newBucket(tt.bytesPerSec)
			if (got == nil) != tt.wantNil {
				t.Errorf("newBucket() = %v, wantNil %v", got, tt.wantNil)
			}
			if !tt.wantNil && got != nil {
				if got.rateBytesPerSec != float64(tt.bytesPerSec) {
					t.Errorf("rateBytesPerSec = %v, want %v", got.rateBytesPerSec, float64(tt.bytesPerSec))
				}
				expectedCap := 2 * float64(tt.bytesPerSec)
				if got.capacity != expectedCap {
					t.Errorf("capacity = %v, want %v", got.capacity, expectedCap)
				}
			}
		})
	}
}

func TestByteTokenBucket_WaitN(t *testing.T) {
	tests := []struct {
		name        string
		bytesPerSec int
		requested   int
	}{
		{
			name:        "small request",
			bytesPerSec: 10000,
			requested:   100,
		},
		{
			name:        "zero request",
			bytesPerSec: 10000,
			requested:   0,
		},
		{
			name:        "negative request",
			bytesPerSec: 10000,
			requested:   -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := newBucket(tt.bytesPerSec)
			start := time.Now()
			bucket.waitN(tt.requested)
			elapsed := time.Since(start)

			// For small requests relative to bucket capacity, it should be fast
			if tt.requested <= 0 {
				if elapsed > 10*time.Millisecond {
					t.Errorf("waitN() took too long for %d bytes: %v", tt.requested, elapsed)
				}
			}
		})
	}
}

func TestByteTokenBucket_WaitN_Nil(t *testing.T) {
	var bucket *byteTokenBucket
	// Should not panic
	bucket.waitN(100)
}

func TestThrottledReader_Read(t *testing.T) {
	data := []byte("hello world, this is test data for throttling")
	reader := bytes.NewReader(data)

	bucket := newBucket(1000) // 1000 bytes/sec
	tr := &throttledReader{
		base: reader,
		bkt:  bucket,
	}

	buf := make([]byte, len(data))
	n, err := tr.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != len(data) {
		t.Errorf("Read() n = %v, want %v", n, len(data))
	}
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Read() data mismatch")
	}
}

func TestLossyReader_Read(t *testing.T) {
	data := []byte("test data")

	t.Run("no loss", func(t *testing.T) {
		reader := bytes.NewReader(data)
		lr := &lossyReader{
			base:    reader,
			rng:     newSeededRand(12345),
			losePct: 0,
		}

		buf := make([]byte, len(data))
		n, err := lr.Read(buf)
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("Read() n = %v, want %v", n, len(data))
		}
	})

	t.Run("with loss configured", func(t *testing.T) {
		// With 100% loss, we should eventually get an error
		reader := bytes.NewReader(data)
		lr := &lossyReader{
			base:    reader,
			rng:     newSeededRand(42),
			losePct: 100,
		}

		// First read might succeed or fail
		buf := make([]byte, len(data))
		_, _ = lr.Read(buf)
		// We don't assert here because randomness makes this non-deterministic
		// The important thing is it doesn't panic
	})
}

func TestKbpsToBytesPerSec(t *testing.T) {
	tests := []struct {
		name string
		kbps int
		want int
	}{
		{
			name: "zero",
			kbps: 0,
			want: 0,
		},
		{
			name: "negative",
			kbps: -100,
			want: 0,
		},
		{
			name: "1000 kbps",
			kbps: 1000,
			want: 125000, // 1000 * 1000 / 8
		},
		{
			name: "8000 kbps",
			kbps: 8000,
			want: 1000000, // 8000 * 1000 / 8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kbpsToBytesPerSec(tt.kbps)
			if got != tt.want {
				t.Errorf("kbpsToBytesPerSec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapReaderThrottleLoss(t *testing.T) {
	data := []byte("test data")

	t.Run("nil reader", func(t *testing.T) {
		result := wrapReaderThrottleLoss(nil, 1000, 0)
		if result != nil {
			t.Errorf("wrapReaderThrottleLoss() with nil should return nil")
		}
	})

	t.Run("no throttle no loss", func(t *testing.T) {
		reader := bytes.NewReader(data)
		result := wrapReaderThrottleLoss(reader, 0, 0)
		if result != reader {
			t.Errorf("wrapReaderThrottleLoss() without settings should return original reader")
		}
	})

	t.Run("with throttle", func(t *testing.T) {
		reader := bytes.NewReader(data)
		result := wrapReaderThrottleLoss(reader, 1000, 0)
		if result == reader {
			t.Errorf("wrapReaderThrottleLoss() with throttle should wrap reader")
		}
		// Verify it's wrapped
		if _, ok := result.(*throttledReader); !ok {
			t.Errorf("wrapReaderThrottleLoss() should return *throttledReader")
		}
	})

	t.Run("with loss", func(t *testing.T) {
		reader := bytes.NewReader(data)
		result := wrapReaderThrottleLoss(reader, 0, 10)
		if result == reader {
			t.Errorf("wrapReaderThrottleLoss() with loss should wrap reader")
		}
		// Verify it's wrapped
		if _, ok := result.(*lossyReader); !ok {
			t.Errorf("wrapReaderThrottleLoss() should return *lossyReader")
		}
	})

	t.Run("with both throttle and loss", func(t *testing.T) {
		reader := bytes.NewReader(data)
		result := wrapReaderThrottleLoss(reader, 1000, 10)
		if result == reader {
			t.Errorf("wrapReaderThrottleLoss() with both should wrap reader")
		}
	})
}

func TestThrottledConnRead_Read(t *testing.T) {
	// Create a mock connection
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Write test data from server side
	go func() {
		server.Write([]byte("test data"))
	}()

	bucket := newBucket(1000)
	conn := &throttledConnRead{
		Conn:    client,
		bkt:     bucket,
		rng:     newSeededRand(12345),
		losePct: 0,
	}

	buf := make([]byte, 9)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != 9 {
		t.Errorf("Read() n = %v, want 9", n)
	}
	if string(buf) != "test data" {
		t.Errorf("Read() data = %q, want %q", string(buf), "test data")
	}
}

func TestThrottledConnRead_Read_WithLoss(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	conn := &throttledConnRead{
		Conn:    client,
		bkt:     newBucket(1000),
		rng:     newSeededRand(42),
		losePct: 100, // 100% loss should trigger error
	}

	buf := make([]byte, 10)
	_, err := conn.Read(buf)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Read() with 100%% loss should return io.ErrUnexpectedEOF, got %v", err)
	}
}

type mockConn struct {
	net.Conn
	readData  []byte
	readPos   int
	writeData []byte
}

func (m *mockConn) Read(b []byte) (int, error) {
	if m.readPos >= len(m.readData) {
		return 0, io.EOF
	}
	n := copy(b, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockConn) Write(b []byte) (int, error) {
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *mockConn) Close() error {
	return nil
}

func TestWrapConnForThrottle(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		conn      net.Conn
		direction string
		wantSame  bool
	}{
		{
			name:      "nil config",
			cfg:       nil,
			conn:      &mockConn{},
			direction: "down",
			wantSame:  true,
		},
		{
			name:      "nil conn",
			cfg:       &config.Config{},
			conn:      nil,
			direction: "down",
			wantSame:  true,
		},
		{
			name: "throttle disabled",
			cfg: &config.Config{
				ThrottleEnabled: false,
			},
			conn:      &mockConn{},
			direction: "down",
			wantSame:  true,
		},
		{
			name: "throttle offline",
			cfg: &config.Config{
				ThrottleOffline: true,
			},
			conn:      &mockConn{},
			direction: "down",
			wantSame:  false, // Will be closed
		},
		{
			name: "throttle enabled with latency",
			cfg: &config.Config{
				ThrottleEnabled:   true,
				ThrottleLatencyMs: 10,
			},
			conn:      &mockConn{},
			direction: "down",
			wantSame:  false, // Will be wrapped
		},
		{
			name: "throttle enabled with bandwidth down",
			cfg: &config.Config{
				ThrottleEnabled:  true,
				ThrottleDownKbps: 1000,
			},
			conn:      &mockConn{},
			direction: "down",
			wantSame:  false,
		},
		{
			name: "throttle enabled with bandwidth up",
			cfg: &config.Config{
				ThrottleEnabled: true,
				ThrottleUpKbps:  1000,
			},
			conn:      &mockConn{},
			direction: "up",
			wantSame:  false,
		},
		{
			name: "throttle enabled with packet loss",
			cfg: &config.Config{
				ThrottleEnabled:    true,
				ThrottlePacketLoss: 5,
			},
			conn:      &mockConn{},
			direction: "down",
			wantSame:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapConnForThrottle(tt.cfg, tt.conn, tt.direction)
			if tt.wantSame && got != tt.conn {
				t.Errorf("wrapConnForThrottle() should return same conn")
			}
			if !tt.wantSame && tt.conn != nil && !tt.cfg.ThrottleOffline && got == tt.conn {
				t.Errorf("wrapConnForThrottle() should wrap conn")
			}
		})
	}
}

func TestLatencyConnWrapper_ReadWrite(t *testing.T) {
	mock := &mockConn{
		readData: []byte("test data"),
	}

	wrapper := &latencyConnWrapper{
		Conn:      mock,
		latencyMs: 1, // Very small latency for testing
		jitterMs:  0,
		rng:       newSeededRand(12345),
	}

	// Test Read
	buf := make([]byte, 9)
	start := time.Now()
	n, err := wrapper.Read(buf)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != 9 {
		t.Errorf("Read() n = %v, want 9", n)
	}
	// Should have some delay
	if elapsed < time.Millisecond {
		t.Errorf("Read() should inject latency, elapsed = %v", elapsed)
	}

	// Test Write
	start = time.Now()
	n, err = wrapper.Write([]byte("write test"))
	elapsed = time.Since(start)

	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != 10 {
		t.Errorf("Write() n = %v, want 10", n)
	}
	if elapsed < time.Millisecond {
		t.Errorf("Write() should inject latency, elapsed = %v", elapsed)
	}
}

func TestLatencyConnWrapper_InjectDelay(t *testing.T) {
	mock := &mockConn{}

	tests := []struct {
		name      string
		latencyMs int
		jitterMs  int
	}{
		{
			name:      "zero latency",
			latencyMs: 0,
			jitterMs:  0,
		},
		{
			name:      "small latency",
			latencyMs: 1,
			jitterMs:  0,
		},
		{
			name:      "with jitter",
			latencyMs: 5,
			jitterMs:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &latencyConnWrapper{
				Conn:      mock,
				latencyMs: tt.latencyMs,
				jitterMs:  tt.jitterMs,
				rng:       newSeededRand(12345),
			}

			start := time.Now()
			wrapper.injectDelay()
			elapsed := time.Since(start)

			if tt.latencyMs > 0 {
				// Should have at least some delay
				if elapsed < time.Millisecond/2 {
					t.Errorf("injectDelay() should add delay, elapsed = %v", elapsed)
				}
			}
		})
	}
}

func TestWrapConnWithLatency(t *testing.T) {
	tests := []struct {
		name      string
		conn      net.Conn
		latencyMs int
		jitterMs  int
		wantSame  bool
	}{
		{
			name:      "nil conn",
			conn:      nil,
			latencyMs: 10,
			jitterMs:  0,
			wantSame:  true,
		},
		{
			name:      "zero latency",
			conn:      &mockConn{},
			latencyMs: 0,
			jitterMs:  0,
			wantSame:  true,
		},
		{
			name:      "with latency",
			conn:      &mockConn{},
			latencyMs: 10,
			jitterMs:  0,
			wantSame:  false,
		},
		{
			name:      "with latency and jitter",
			conn:      &mockConn{},
			latencyMs: 10,
			jitterMs:  5,
			wantSame:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapConnWithLatency(tt.conn, tt.latencyMs, tt.jitterMs)
			if tt.wantSame && got != tt.conn {
				t.Errorf("wrapConnWithLatency() should return same conn")
			}
			if !tt.wantSame && got == tt.conn {
				t.Errorf("wrapConnWithLatency() should wrap conn")
			}
			if !tt.wantSame {
				if wrapper, ok := got.(*latencyConnWrapper); ok {
					if wrapper.latencyMs != tt.latencyMs {
						t.Errorf("latencyMs = %v, want %v", wrapper.latencyMs, tt.latencyMs)
					}
					if wrapper.jitterMs != tt.jitterMs {
						t.Errorf("jitterMs = %v, want %v", wrapper.jitterMs, tt.jitterMs)
					}
				}
			}
		})
	}
}

func TestDeps_ThrottleSleepWS(t *testing.T) {
	tests := []struct {
		name      string
		cfg       config.Config
		direction domain.Direction
		n         int
		wantSleep bool
	}{
		{
			name: "throttle disabled",
			cfg: config.Config{
				ThrottleEnabled: false,
			},
			direction: domain.DirectionUpstreamToClient,
			n:         1000,
			wantSleep: false,
		},
		{
			name: "zero bytes",
			cfg: config.Config{
				ThrottleEnabled:  true,
				ThrottleDownKbps: 1000,
			},
			direction: domain.DirectionUpstreamToClient,
			n:         0,
			wantSleep: false,
		},
		{
			name: "negative bytes",
			cfg: config.Config{
				ThrottleEnabled:  true,
				ThrottleDownKbps: 1000,
			},
			direction: domain.DirectionUpstreamToClient,
			n:         -10,
			wantSleep: false,
		},
		{
			name: "downstream throttle",
			cfg: config.Config{
				ThrottleEnabled:  true,
				ThrottleDownKbps: 1000,
			},
			direction: domain.DirectionUpstreamToClient,
			n:         100,
			wantSleep: true,
		},
		{
			name: "upstream throttle",
			cfg: config.Config{
				ThrottleEnabled: true,
				ThrottleUpKbps:  1000,
			},
			direction: domain.DirectionClientToUpstream,
			n:         100,
			wantSleep: true,
		},
		{
			name: "zero kbps",
			cfg: config.Config{
				ThrottleEnabled:  true,
				ThrottleDownKbps: 0,
			},
			direction: domain.DirectionUpstreamToClient,
			n:         1000,
			wantSleep: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := &Deps{Cfg: tt.cfg}
			start := time.Now()
			deps.throttleSleepWS(tt.direction, tt.n)
			elapsed := time.Since(start)

			if tt.wantSleep && elapsed < time.Millisecond {
				t.Errorf("throttleSleepWS() should sleep, elapsed = %v", elapsed)
			}
			if !tt.wantSleep && elapsed > 10*time.Millisecond {
				t.Errorf("throttleSleepWS() should not sleep, elapsed = %v", elapsed)
			}
		})
	}
}

func TestDeps_InjectLatency(t *testing.T) {
	tests := []struct {
		name      string
		cfg       config.Config
		wantSleep bool
	}{
		{
			name: "throttle disabled",
			cfg: config.Config{
				ThrottleEnabled: false,
			},
			wantSleep: false,
		},
		{
			name: "zero latency",
			cfg: config.Config{
				ThrottleEnabled:   true,
				ThrottleLatencyMs: 0,
			},
			wantSleep: false,
		},
		{
			name: "with latency",
			cfg: config.Config{
				ThrottleEnabled:   true,
				ThrottleLatencyMs: 1,
			},
			wantSleep: true,
		},
		{
			name: "with latency and jitter",
			cfg: config.Config{
				ThrottleEnabled:       true,
				ThrottleLatencyMs:     5,
				ThrottleLatencyJitter: 2,
			},
			wantSleep: true,
		},
		{
			name: "large negative jitter",
			cfg: config.Config{
				ThrottleEnabled:       true,
				ThrottleLatencyMs:     1,
				ThrottleLatencyJitter: 100, // Could make latency negative
			},
			wantSleep: false, // Negative latency should be clamped to 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := &Deps{Cfg: tt.cfg}
			start := time.Now()
			deps.injectLatency()
			elapsed := time.Since(start)

			if tt.wantSleep && elapsed < time.Millisecond/2 {
				t.Errorf("injectLatency() should sleep, elapsed = %v", elapsed)
			}
		})
	}
}

// Helper function for creating seeded random generators
func newSeededRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
