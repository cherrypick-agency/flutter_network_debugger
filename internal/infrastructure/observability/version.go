package observability

// Binary versioning for logs and metrics.
// Values are overwritten via -ldflags during build.
var (
	Version = "dev"  // release version
	Commit  = "none" // short commit
	Date    = ""     // ISO8601 UTC build time
)
