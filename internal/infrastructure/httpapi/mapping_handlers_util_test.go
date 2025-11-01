package httpapi

import "testing"

func TestGuessContentTypeByName(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"script.js", "application/javascript"},
		{"SCRIPT.JS", "application/javascript"},
		{"style.css", "text/css"},
		{"data.json", "application/json"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"icon.svg", "image/svg+xml"},
		{"unknown.xyz", "application/octet-stream"},
		{"noext", "application/octet-stream"},
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		got := guessContentTypeByName(tt.filename)
		if got != tt.want {
			t.Errorf("guessContentTypeByName(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}
}
