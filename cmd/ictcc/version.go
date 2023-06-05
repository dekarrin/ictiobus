package main

const (
	// Version is the version of this release of ictiobus.
	Version = "1.0.0"
)

// GetVersionString returns the current version of ictcc, styled with the name
// of the binary itself.
func GetVersionString() string {
	return "ictcc v" + Version
}
