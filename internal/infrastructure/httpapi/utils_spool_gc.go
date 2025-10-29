package httpapi

import (
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// startSpoolGC periodically deletes orphan spool files older than a TTL.
func startSpoolGC(d *Deps) {
    ttl := 24 * time.Hour
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()
    // initial run
    cleanupSpoolDir(d.Cfg.BodySpoolDir, ttl)
    for range ticker.C {
        cleanupSpoolDir(d.Cfg.BodySpoolDir, ttl)
    }
}

func cleanupSpoolDir(dir string, olderThan time.Duration) {
    if dir == "" {
        dir = os.TempDir()
    }
    now := time.Now()
    _ = filepath.WalkDir(dir, func(path string, de fs.DirEntry, err error) error {
        if err != nil { return nil }
        if de.IsDir() { return nil }
        name := de.Name()
        // only our patterns
        if !(strings.HasPrefix(name, "gpx-req-") || strings.HasPrefix(name, "gpx-resp-") || strings.HasPrefix(name, "gpx-ws-")) {
            return nil
        }
        if info, err := de.Info(); err == nil {
            if now.Sub(info.ModTime()) > olderThan {
                _ = os.Remove(path)
            }
        }
        return nil
    })
}


