package detector

import (
	"testing"

	"network-debugger/internal/features/process/domain"
)

// Composer 1.
func TestNewDetector(t *testing.T) {
	// Тест без привилегий
	detector, err := NewDetector(false)
	if err != nil {
		t.Fatalf("NewDetector(false) error = %v", err)
	}

	if detector == nil {
		t.Fatal("NewDetector(false) returned nil")
	}

	// Проверяем что это реализует интерфейс
	_, ok := detector.(domain.IProcessDetector)
	if !ok {
		t.Error("Detector does not implement IProcessDetector interface")
	}

	// Тест с привилегиями
	detector2, err := NewDetector(true)
	if err != nil {
		t.Fatalf("NewDetector(true) error = %v", err)
	}

	if detector2 == nil {
		t.Fatal("NewDetector(true) returned nil")
	}
}

// Composer 1.
func TestNewDetectorForPlatform(t *testing.T) {
	detector, err := newDetectorForPlatform()
	if err != nil {
		t.Fatalf("newDetectorForPlatform() error = %v", err)
	}

	if detector == nil {
		t.Fatal("newDetectorForPlatform() returned nil")
	}

	// Проверяем что возвращается gopsutilAdapter
	_, ok := detector.(*gopsutilAdapter)
	if !ok {
		t.Error("newDetectorForPlatform() did not return gopsutilAdapter")
	}
}

// Composer 1.
func TestGopsutilAdapter_RequiresPrivileges(t *testing.T) {
	adapter := &gopsutilAdapter{}

	if adapter.RequiresPrivileges() {
		t.Error("RequiresPrivileges() should return false for gopsutilAdapter")
	}
}
