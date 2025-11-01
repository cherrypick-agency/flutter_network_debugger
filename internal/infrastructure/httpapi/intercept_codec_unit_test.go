package httpapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"testing"
)

func TestDecodeForIntercept(t *testing.T) {
	testData := []byte("hello world, this is test data for decompression testing")

	// Compress test data with gzip
	var gzipBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuf)
	gzipWriter.Write(testData)
	gzipWriter.Close()
	gzipData := gzipBuf.Bytes()

	// Compress test data with deflate
	var deflateBuf bytes.Buffer
	deflateWriter, _ := flate.NewWriter(&deflateBuf, flate.DefaultCompression)
	deflateWriter.Write(testData)
	deflateWriter.Close()
	deflateData := deflateBuf.Bytes()

	tests := []struct {
		name            string
		body            []byte
		contentEncoding string
		limit           int
		wantDecoded     bool
		wantData        []byte
	}{
		{
			name:            "gzip decode",
			body:            gzipData,
			contentEncoding: "gzip",
			limit:           0,
			wantDecoded:     true,
			wantData:        testData,
		},
		{
			name:            "deflate decode",
			body:            deflateData,
			contentEncoding: "deflate",
			limit:           0,
			wantDecoded:     true,
			wantData:        testData,
		},
		{
			name:            "no encoding",
			body:            testData,
			contentEncoding: "",
			limit:           0,
			wantDecoded:     false,
			wantData:        testData,
		},
		{
			name:            "unknown encoding",
			body:            testData,
			contentEncoding: "br",
			limit:           0,
			wantDecoded:     false,
			wantData:        testData,
		},
		{
			name:            "gzip with limit",
			body:            gzipData,
			contentEncoding: "gzip",
			limit:           5,
			wantDecoded:     true,
			wantData:        testData[:5],
		},
		{
			name:            "gzip uppercase",
			body:            gzipData,
			contentEncoding: "GZIP",
			limit:           0,
			wantDecoded:     true,
			wantData:        testData,
		},
		{
			name:            "invalid gzip data",
			body:            []byte("not gzip data"),
			contentEncoding: "gzip",
			limit:           0,
			wantDecoded:     false,
			wantData:        []byte("not gzip data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, decoded := decodeForIntercept(tt.body, tt.contentEncoding, tt.limit)

			if decoded != tt.wantDecoded {
				t.Errorf("decodeForIntercept() decoded = %v, want %v", decoded, tt.wantDecoded)
			}

			if !bytes.Equal(result, tt.wantData) {
				t.Errorf("decodeForIntercept() result = %q, want %q", string(result), string(tt.wantData))
			}
		})
	}
}

func TestEncodeForIntercept(t *testing.T) {
	testData := []byte("hello world, this is test data for compression testing")

	tests := []struct {
		name            string
		body            []byte
		contentEncoding string
		wantCompressed  bool
		wantLen         int // -1 means don't check length, just verify it's different
	}{
		{
			name:            "gzip encoding",
			body:            testData,
			contentEncoding: "gzip",
			wantCompressed:  true,
			wantLen:         -1,
		},
		{
			name:            "deflate encoding",
			body:            testData,
			contentEncoding: "deflate",
			wantCompressed:  true,
			wantLen:         -1,
		},
		{
			name:            "no encoding",
			body:            testData,
			contentEncoding: "",
			wantCompressed:  false,
			wantLen:         len(testData),
		},
		{
			name:            "unknown encoding",
			body:            testData,
			contentEncoding: "br",
			wantCompressed:  false,
			wantLen:         len(testData),
		},
		{
			name:            "gzip with whitespace",
			body:            testData,
			contentEncoding: "  gzip  ",
			wantCompressed:  true,
			wantLen:         -1,
		},
		{
			name:            "GZIP uppercase",
			body:            testData,
			contentEncoding: "GZIP",
			wantCompressed:  true,
			wantLen:         -1,
		},
		{
			name:            "deflate uppercase",
			body:            testData,
			contentEncoding: "DEFLATE",
			wantCompressed:  true,
			wantLen:         -1,
		},
		{
			name:            "empty body gzip",
			body:            []byte{},
			contentEncoding: "gzip",
			wantCompressed:  true,
			wantLen:         -1,
		},
		{
			name:            "empty body deflate",
			body:            []byte{},
			contentEncoding: "deflate",
			wantCompressed:  true,
			wantLen:         -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, compressed := encodeForIntercept(tt.body, tt.contentEncoding)

			if compressed != tt.wantCompressed {
				t.Errorf("encodeForIntercept() compressed = %v, want %v", compressed, tt.wantCompressed)
			}

			if result == nil {
				t.Fatal("encodeForIntercept() returned nil result")
			}

			if tt.wantLen >= 0 && len(result) != tt.wantLen {
				t.Errorf("encodeForIntercept() result length = %d, want %d", len(result), tt.wantLen)
			}

			// Verify we can decompress what we compressed
			if tt.wantCompressed {
				var decompressed []byte
				var err error

				switch tt.contentEncoding {
				case "gzip", "  gzip  ", "GZIP":
					gr, err := gzip.NewReader(bytes.NewReader(result))
					if err != nil {
						t.Fatalf("failed to create gzip reader: %v", err)
					}
					decompressed, err = io.ReadAll(gr)
					gr.Close()
					if err != nil {
						t.Fatalf("failed to decompress gzip: %v", err)
					}

				case "deflate", "DEFLATE":
					fr := flate.NewReader(bytes.NewReader(result))
					decompressed, err = io.ReadAll(fr)
					fr.Close()
					if err != nil {
						t.Fatalf("failed to decompress deflate: %v", err)
					}
				}

				if !bytes.Equal(decompressed, tt.body) {
					t.Errorf("decompressed data doesn't match original")
				}
			} else {
				// For non-compressed, result should be the same as input
				if !bytes.Equal(result, tt.body) {
					t.Errorf("non-compressed result doesn't match input")
				}
			}
		})
	}
}
