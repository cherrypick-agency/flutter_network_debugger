package httpapi

import (
	"network-debugger/internal/infrastructure/config"
	"time"
)

func sleepResponseDelay(cfg config.Config) {
	if cfg.ResponseDelayMinMs > 0 && cfg.ResponseDelayMaxMs > 0 {
		delta := cfg.ResponseDelayMaxMs - cfg.ResponseDelayMinMs
		if delta < 0 {
			delta = 0
		}
		// Небольшая псевдослучайность на основе времени — для дев/демо достаточно
		n := time.Now().UnixNano()
		rnd := int(n % int64(delta+1))
		time.Sleep(time.Duration(cfg.ResponseDelayMinMs+rnd) * time.Millisecond)
		return
	}
	if cfg.ResponseDelayMs > 0 {
		time.Sleep(time.Duration(cfg.ResponseDelayMs) * time.Millisecond)
	}
}
