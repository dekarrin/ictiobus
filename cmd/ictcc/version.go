package main

const (
	// Version is the version of this release of ictiobus.
	Version = "0.8.0+dev"
)

// GetVersionString returns the current version of ictcc, styled with the name
// of the binary itself.
func GetVersionString() string {
	return "ictcc v" + Version
}
