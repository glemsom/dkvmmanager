// Package version provides version information for DKVM Manager
package version

// Version holds the current version of the application
// Can be overridden at build time using -ldflags
var Version = "0.1.9"

// Commit holds the Git commit hash
// Can be overridden at build time using -ldflags
var Commit = "none"

// Date holds the build timestamp
// Can be overridden at build time using -ldflags
var Date = "unknown"
