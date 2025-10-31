package httpapi

import (
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/internal/infrastructure/config"
)

// byteTokenBucket — простой токен‑бакет для ограничения по байтам в секунду.
type byteTokenBucket struct {
	rateBytesPerSec float64
	capacity        float64
	mu              sync.Mutex
	tokens          float64
	last            time.Time
}

func newBucket(bytesPerSec int) *byteTokenBucket {
	if bytesPerSec <= 0 {
		return nil
	}
	r := float64(bytesPerSec)
	// небольшая ёмкость, чтобы позволить мелкие всплески
	cap := 2 * r
	return &byteTokenBucket{rateBytesPerSec: r, capacity: cap, tokens: cap, last: time.Now()}
}

// waitN блокирует текущую горутину, пока не накопится минимум n токенов.
func (b *byteTokenBucket) waitN(n int) {
	if b == nil || n <= 0 {
		return
	}
	need := float64(n)
	for {
		b.mu.Lock()
		// пополним
		now := time.Now()
		elapsed := now.Sub(b.last).Seconds()
		if elapsed > 0 {
			b.tokens += elapsed * b.rateBytesPerSec
			if b.tokens > b.capacity {
				b.tokens = b.capacity
			}
			b.last = now
		}
		if b.tokens >= need {
			b.tokens -= need
			b.mu.Unlock()
			return
		}
		// сколько ждать до появления недостающих токенов
		deficit := need - b.tokens
		// минимальный слип 1мс, чтобы не крутить цикл
		sleep := time.Duration(deficit / b.rateBytesPerSec * float64(time.Second))
		if sleep < time.Millisecond {
			sleep = time.Millisecond
		}
		b.mu.Unlock()
		time.Sleep(sleep)
	}
}

// throttledReader — читает из base и ограничивает пропускную способность.
type throttledReader struct {
	base io.Reader
	bkt  *byteTokenBucket
}

func (r *throttledReader) Read(p []byte) (int, error) {
	n, err := r.base.Read(p)
	if n > 0 {
		r.bkt.waitN(n)
	}
	return n, err
}

// lossyReader — имитирует потерю пакетов: иногда прерывает поток ошибкой.
type lossyReader struct {
	base    io.Reader
	rng     *rand.Rand
	losePct int // 0..100
}

func (r *lossyReader) Read(p []byte) (int, error) {
	// мелкая вероятность прервать чтение до реального I/O
	if r.losePct > 0 && r.rng.Intn(100) < r.losePct {
		return 0, io.ErrUnexpectedEOF
	}
	return r.base.Read(p)
}

// throttledConnRead — оборачивает net.Conn и ограничивает скорость чтения.
type throttledConnRead struct {
	net.Conn
	bkt *byteTokenBucket
	// для имитации потерь
	rng     *rand.Rand
	losePct int
}

func (c *throttledConnRead) Read(p []byte) (int, error) {
	if c.losePct > 0 && c.rng.Intn(100) < c.losePct {
		// имитируем обрыв соединения
		return 0, io.ErrUnexpectedEOF
	}
	n, err := c.Conn.Read(p)
	if n > 0 && c.bkt != nil {
		c.bkt.waitN(n)
	}
	return n, err
}

// helpers
func kbpsToBytesPerSec(kbps int) int { // kbit/s -> bytes/s
	if kbps <= 0 {
		return 0
	}
	// 1000/8 ~= 125; используем 1000 для привычной сетевой величины
	return int(float64(kbps) * 1000.0 / 8.0)
}

func wrapReaderThrottleLoss(r io.Reader, bytesPerSec int, lossPct int) io.Reader {
	if r == nil {
		return r
	}
	out := r
	if bytesPerSec > 0 {
		out = &throttledReader{base: out, bkt: newBucket(bytesPerSec)}
	}
	if lossPct > 0 {
		out = &lossyReader{base: out, rng: rand.New(rand.NewSource(time.Now().UnixNano())), losePct: lossPct}
	}
	return out
}

func wrapConnForThrottle(cfg *config.Config, conn net.Conn, direction string) net.Conn {
	if conn == nil || cfg == nil {
		return conn
	}
	if cfg.ThrottleOffline {
		// немедленный обрыв даст клиенту представление об оффлайне
		_ = conn.Close()
		return conn
	}

	if !cfg.ThrottleEnabled {
		return conn
	}

	// Применяем latency wrapper первым (самый внутренний слой)
	// Latency должна быть на самом низком уровне, чтобы влиять на каждый пакет
	if cfg.ThrottleLatencyMs > 0 {
		conn = wrapConnWithLatency(conn, cfg.ThrottleLatencyMs, cfg.ThrottleLatencyJitter)
	}

	// Затем применяем bandwidth throttling и packet loss (если настроены)
	var bps int
	switch direction {
	case "down": // upstream -> client
		bps = kbpsToBytesPerSec(cfg.ThrottleDownKbps)
	case "up": // client -> upstream
		bps = kbpsToBytesPerSec(cfg.ThrottleUpKbps)
	default:
		bps = 0
	}

	// Применяем bandwidth/packet loss wrapper если есть что применять
	if bps > 0 || cfg.ThrottlePacketLoss > 0 {
		conn = &throttledConnRead{
			Conn:    conn,
			bkt:     newBucket(bps),
			rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
			losePct: cfg.ThrottlePacketLoss,
		}
	}

	return conn
}

// throttleSleepWS — для путей, где нет прямого доступа к net.Conn (gorilla/websocket).
// Засыпает пропорционально размеру сообщения.
func (d *Deps) throttleSleepWS(direction domain.Direction, n int) {
	if n <= 0 || !d.Cfg.ThrottleEnabled {
		return
	}
	var kbps int
	if direction == domain.DirectionUpstreamToClient {
		kbps = d.Cfg.ThrottleDownKbps
	} else {
		kbps = d.Cfg.ThrottleUpKbps
	}
	bps := kbpsToBytesPerSec(kbps)
	if bps <= 0 {
		return
	}
	sleep := time.Duration(float64(n) / float64(bps) * float64(time.Second))
	if sleep < time.Millisecond {
		sleep = time.Millisecond
	}
	time.Sleep(sleep)
}

// injectLatency — добавляет искусственную задержку для симуляции высокой latency (RTT/ping).
// В отличие от bandwidth throttling, который замедляет передачу данных пропорционально их размеру,
// latency injection добавляет фиксированную задержку к каждой операции независимо от размера данных.
// Это симулирует географическое расстояние, спутниковые каналы или перегруженные сети.
func (d *Deps) injectLatency() {
	if !d.Cfg.ThrottleEnabled || d.Cfg.ThrottleLatencyMs <= 0 {
		return
	}
	baseLatency := time.Duration(d.Cfg.ThrottleLatencyMs) * time.Millisecond

	// Добавляем jitter (случайную вариацию) если настроен
	if d.Cfg.ThrottleLatencyJitter > 0 {
		jitter := rand.Intn(d.Cfg.ThrottleLatencyJitter*2+1) - d.Cfg.ThrottleLatencyJitter
		baseLatency += time.Duration(jitter) * time.Millisecond
	}

	// Не даём latency стать отрицательной из-за jitter
	if baseLatency < 0 {
		baseLatency = 0
	}

	time.Sleep(baseLatency)
}

// latencyConnWrapper — обёртка для net.Conn, добавляющая latency к каждой операции Read/Write.
// Это симулирует RTT (Round Trip Time) для каждого пакета данных.
type latencyConnWrapper struct {
	net.Conn
	latencyMs int
	jitterMs  int
	rng       *rand.Rand
}

func (c *latencyConnWrapper) Read(p []byte) (int, error) {
	c.injectDelay()
	return c.Conn.Read(p)
}

func (c *latencyConnWrapper) Write(p []byte) (int, error) {
	c.injectDelay()
	return c.Conn.Write(p)
}

func (c *latencyConnWrapper) injectDelay() {
	if c.latencyMs <= 0 {
		return
	}
	delay := time.Duration(c.latencyMs) * time.Millisecond
	if c.jitterMs > 0 {
		jitter := c.rng.Intn(c.jitterMs*2+1) - c.jitterMs
		delay += time.Duration(jitter) * time.Millisecond
	}
	if delay > 0 {
		time.Sleep(delay)
	}
}

// wrapConnWithLatency — оборачивает connection для добавления latency injection.
func wrapConnWithLatency(conn net.Conn, latencyMs int, jitterMs int) net.Conn {
	if conn == nil || latencyMs <= 0 {
		return conn
	}
	return &latencyConnWrapper{
		Conn:      conn,
		latencyMs: latencyMs,
		jitterMs:  jitterMs,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
