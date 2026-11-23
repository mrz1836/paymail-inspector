// Package main is the paymail-inspector entry point. All logic lives in
// internal/cli; this file only carries the build metadata injected by
// goreleaser and magex.
package main

import (
	"context"
	"os"

	"github.com/mrz1836/paymail-inspector/internal/cli"
)

// Build metadata injected via ldflags:
//
//	-X main.version=...   -X main.commit=...   -X main.buildDate=...
//
//nolint:gochecknoglobals // ldflags can only target package-level variables
var (
	version   string
	commit    string
	buildDate string
)

func main() {
	os.Exit(cli.Execute(context.Background(), cli.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    buildDate,
	}))
}
