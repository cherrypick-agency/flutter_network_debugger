package domain

import "context"

// ScriptRepository is the port/interface for script persistence
// This follows Clean Architecture - this is a PORT that adapters implement
type ScriptRepository interface {
	// Save creates or updates a script
	Save(ctx context.Context, script *Script) error

	// Get retrieves a script by ID
	Get(ctx context.Context, id string) (*Script, error)

	// List retrieves scripts matching the filter
	List(ctx context.Context, filter ScriptFilter) ([]*Script, error)

	// Delete removes a script
	Delete(ctx context.Context, id string) error

	// UpdateEnabled toggles script enabled state
	UpdateEnabled(ctx context.Context, id string, enabled bool) error
}
