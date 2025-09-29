package e2e

import (
    "fmt"
    "os"
    "strconv"
    "time"
    "testing"
)

// Global timeout guard for the e2e test package to avoid indefinite hangs.
func TestMain(m *testing.M) {
    // default 2 minutes, overridable via E2E_TIMEOUT_SECONDS
    timeout := 2 * time.Minute
    if v := os.Getenv("E2E_TIMEOUT_SECONDS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            timeout = time.Duration(n) * time.Second
        }
    }
    timer := time.AfterFunc(timeout, func() {
        fmt.Fprintf(os.Stderr, "\n[E2E] global timeout %s reached, aborting tests\n", timeout)
        os.Exit(3)
    })
    code := m.Run()
    _ = timer.Stop()
    os.Exit(code)
}


