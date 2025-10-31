package usecase

import (
	"context"
	"network-debugger/internal/domain"
)

// ComposeLibraryRepository is a persistence interface for compose library.
// Implementation stores full library snapshot atomically (e.g., JSON file).
type ComposeLibraryRepository interface {
	LoadLibrary(ctx context.Context) (domain.ComposeLibrary, error)
	SaveLibrary(ctx context.Context, lib domain.ComposeLibrary) error
}

// ComposeHistoryRepository persists history of sent requests.
type ComposeHistoryRepository interface {
	AppendHistory(ctx context.Context, e domain.ComposeHistoryEntry) error
	ListHistory(ctx context.Context, from string, limit int) ([]domain.ComposeHistoryEntry, string, error)
}
