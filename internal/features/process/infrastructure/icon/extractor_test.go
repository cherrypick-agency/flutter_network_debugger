package icon

import (
	"testing"

	"network-debugger/internal/features/process/domain"
)

// Composer 1.
func TestNewExtractor(t *testing.T) {
	extractor, err := NewExtractor()

	if err != nil {
		t.Fatalf("NewExtractor() error = %v, want nil", err)
	}

	if extractor == nil {
		t.Fatal("NewExtractor() returned nil")
	}

	// Проверяем что это реализует интерфейс
	_, ok := extractor.(domain.IIconExtractor)
	if !ok {
		t.Error("Extractor does not implement IIconExtractor interface")
	}
}
