package httpapi

import "sync/atomic"

// previewMaxBytes controls how many bytes of payload are kept in Frame/Event previews.
// <= 0 means no truncation (keep full payload). Default is 1024 and can be overridden from config.
var previewMaxBytes atomic.Int32

// exposeSensitiveHeaders controls whether previews include raw (unmasked) header values
// alongside masked headers map in `headersRaw` field. This is used by the frontend
// to reveal sensitive headers on demand (eye icon) while keeping masked view by default.
// It is configured from config.ExposeSensitiveHeaders in router.
var exposeSensitiveHeaders atomic.Bool

// Whether to decompress preview payload for common encodings (gzip/deflate/br)
var previewDecompress atomic.Bool

func init() {
	previewMaxBytes.Store(1024)
	exposeSensitiveHeaders.Store(true)
	previewDecompress.Store(true)
}

// formatBinaryPreview returns a short hexdump-like preview for binary data.
func formatBinaryPreview(b []byte, max int) string {
	if max <= 0 || max > len(b) {
		max = len(b)
	}
	if max > 256 {
		max = 256
	}
	const hexdigits = "0123456789ABCDEF"
	n := max
	out := make([]byte, 0, n*3+16)
	for i := 0; i < n; i++ {
		v := b[i]
		out = append(out, hexdigits[v>>4], hexdigits[v&0x0F])
		if i+1 < n {
			out = append(out, ' ')
		}
	}
	return string(out)
}
