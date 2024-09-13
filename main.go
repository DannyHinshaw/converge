// Package main is the executable entrypoint for running the converge tool.
package main

import (
	"io"
	"os"

	"github.com/dannyhinshaw/converge/cmd"
)

// version is the version of the converge tool.
// This is set at build time using the -ldflags flag.
var version = "(dev)"

func main() {
	if err := cmd.NewRoot(version).Execute(); err != nil {
		_, _ = io.WriteString(os.Stderr, err.Error())
	}
}
