// Package version holds release metadata injected at build time.
package version

// Set via -ldflags at release build; defaults suit local development.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns the application version (release tag or "dev").
func String() string {
	return Version
}
