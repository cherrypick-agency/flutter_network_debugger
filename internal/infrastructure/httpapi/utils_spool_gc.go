package httpapi

import (
	"context"
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
	cleanupSpoolDir(d, d.Cfg.BodySpoolDir, ttl)
	for range ticker.C {
		cleanupSpoolDir(d, d.Cfg.BodySpoolDir, ttl)
	}
}

func cleanupSpoolDir(d *Deps, dir string, olderThan time.Duration) {
	if dir == "" {
		dir = os.TempDir()
	}
	keep := keepSpoolPaths(d, dir)
	_, keepAllMaps := keep[keepAllMapSpoolKey]
	now := time.Now()
	_ = filepath.WalkDir(dir, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if de.IsDir() {
			return nil
		}
		name := de.Name()
		// only our patterns
		isKnown := strings.HasPrefix(name, "gpx-req-") ||
			strings.HasPrefix(name, "gpx-resp-") ||
			strings.HasPrefix(name, "gpx-ws-") ||
			strings.HasPrefix(name, "gpx-map-")
		if !isKnown {
			return nil
		}
		if info, err := de.Info(); err == nil {
			if now.Sub(info.ModTime()) > olderThan {
				if strings.HasPrefix(name, "gpx-map-") {
					if keepAllMaps {
						return nil
					}
					abs, _ := filepath.Abs(path)
					if _, ok := keep[abs]; ok {
						return nil
					}
				}
				_ = os.Remove(path)
			}
		}
		return nil
	})
}

func keepSpoolPaths(d *Deps, spoolDir string) map[string]struct{} {
	out := map[string]struct{}{}
	if d == nil || d.Mapping == nil {
		out[keepAllMapSpoolKey] = struct{}{}
		return out
	}
	spoolAbs, _ := filepath.Abs(spoolDir)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rules, err := d.Mapping.List(ctx)
	if err != nil {
		out[keepAllMapSpoolKey] = struct{}{}
		return out
	}
	for _, r := range rules {
		if r.BlobPath != nil && strings.TrimSpace(*r.BlobPath) != "" {
			abs, _ := filepath.Abs(*r.BlobPath)
			out[abs] = struct{}{}
		}
		// Just in case: if someone uses filePath inside spool dir.
		if r.FilePath != nil && strings.TrimSpace(*r.FilePath) != "" {
			abs, _ := filepath.Abs(*r.FilePath)
			// keep only if this is really our spool file
			if strings.HasPrefix(filepath.Base(abs), "gpx-map-") && isWithinDir(spoolAbs, abs) {
				out[abs] = struct{}{}
			}
		}
	}
	return out
}

const keepAllMapSpoolKey = "\x00KEEP_ALL_GPX_MAP"
