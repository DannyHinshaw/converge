package main

// The following vars will be set at build time through ldflags.
var (
	// Version is the current version of gonverge.
	Version = "" //nolint: gochecknoglobals // these are defined by ldflags

	// BuildDate is the date the binary was built.
	BuildDate = "" //nolint: gochecknoglobals // these are defined by ldflags
)
