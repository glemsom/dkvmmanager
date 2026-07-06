// Package version provides version information for DKVM Manager
package version

// Version is provided by version_gen.go (generated from VERSION file).
// Can be overridden at build time using -ldflags: -X github.com/glemsom/dkvmmanager/internal/version.Version=...

// Commit holds the Git commit hash
// Can be overridden at build time using -ldflags
var Commit = "none"

// Date holds the build timestamp
// Can be overridden at build time using -ldflags
var Date = "unknown"
