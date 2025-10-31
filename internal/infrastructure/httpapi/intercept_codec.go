package httpapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"strings"
)

// decodeForIntercept пытается декомпрессировать тело для превью/редактирования
// поддерживаем gzip/deflate. Возвращает декодированные байты (или исходные, если декодирование не удалось).
func decodeForIntercept(body []byte, contentEncoding string, limit int) ([]byte, bool) {
	enc := strings.ToLower(strings.TrimSpace(contentEncoding))
	if limit <= 0 {
		limit = 1 << 20
	}
	switch enc {
	case "gzip":
		zr, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return body, false
		}
		defer zr.Close()
		r := io.LimitReader(zr, int64(limit))
		out, err := io.ReadAll(r)
		if err != nil {
			return body, false
		}
		return out, true
	case "deflate":
		fr := flate.NewReader(bytes.NewReader(body))
		if fr == nil {
			return body, false
		}
		defer fr.Close()
		r := io.LimitReader(fr, int64(limit))
		out, err := io.ReadAll(r)
		if err != nil {
			return body, false
		}
		return out, true
	default:
		return body, false
	}
}

// encodeForIntercept перекодирует тело обратно в исходный Content-Encoding.
func encodeForIntercept(body []byte, contentEncoding string) ([]byte, bool) {
	enc := strings.ToLower(strings.TrimSpace(contentEncoding))
	switch enc {
	case "gzip":
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write(body); err != nil {
			_ = zw.Close()
			return nil, false
		}
		if err := zw.Close(); err != nil {
			return nil, false
		}
		return buf.Bytes(), true
	case "deflate":
		var buf bytes.Buffer
		zw, err := flate.NewWriter(&buf, flate.DefaultCompression)
		if err != nil {
			return nil, false
		}
		if _, err := zw.Write(body); err != nil {
			_ = zw.Close()
			return nil, false
		}
		if err := zw.Close(); err != nil {
			return nil, false
		}
		return buf.Bytes(), true
	default:
		return body, false
	}
}
