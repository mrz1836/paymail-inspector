package cli

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
)

// Placeholders for build metadata that was not injected at build time.
const (
	devVersion     = "dev"
	unknownCommit  = "unknown"
	unknownDate    = "unknown"
	develModuleVer = "(devel)"
	shortSHALength = 7
)

// BuildInfo identifies the running build. Release builds receive it through
// ldflags (see .goreleaser.yml and .mage.yaml); everything else falls back to
// the module metadata the Go toolchain embeds, so `go install …@vX.Y.Z` still
// reports a real version.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// pseudoVersion matches the timestamp-and-hash suffix of a Go pseudo-version,
// e.g. 0.4.5-0.20261123174806-980d3c0abcde.
var pseudoVersion = regexp.MustCompile(`\d{14}-[0-9a-f]{12}$`)

// readBuildInfo is the toolchain metadata lookup, swappable in tests.
type readBuildInfo func() (*debug.BuildInfo, bool)

// String renders the one-line version summary used by --version.
func (b BuildInfo) String() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", b.Version, b.Commit, b.Date)
}

// resolve fills empty fields from the toolchain metadata and then from
// placeholders, so every field is always printable.
func (b BuildInfo) resolve(read readBuildInfo) BuildInfo {
	if read != nil && (b.Version == "" || b.Commit == "" || b.Date == "") {
		if info, ok := read(); ok && info != nil {
			b = b.fillFrom(info)
		}
	}

	if b.Version == "" {
		b.Version = devVersion
	}

	if b.Commit == "" {
		b.Commit = unknownCommit
	}

	if b.Date == "" {
		b.Date = unknownDate
	}

	return b
}

// fillFrom copies the module version and VCS stamps from info into any
// field that is still empty.
func (b BuildInfo) fillFrom(info *debug.BuildInfo) BuildInfo {
	if b.Version == "" && info.Main.Version != develModuleVer {
		b.Version = strings.TrimPrefix(info.Main.Version, "v")
	}

	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}

	if b.Commit == "" {
		b.Commit = shortSHA(settings["vcs.revision"])
	}

	if b.Date == "" {
		b.Date = settings["vcs.time"]
	}

	return b
}

// isRelease reports whether the build carries a real release version: not a
// development placeholder, a pseudo-version, or a version with build
// metadata (Go stamps a modified checkout "+dirty"). Only releases check for
// updates.
func (b BuildInfo) isRelease() bool {
	if b.Version == "" || b.Version == devVersion || strings.Contains(b.Version, "+") {
		return false
	}

	return !pseudoVersion.MatchString(b.Version)
}

// shortSHA abbreviates a commit hash for display.
func shortSHA(sha string) string {
	if len(sha) > shortSHALength {
		return sha[:shortSHALength]
	}

	return sha
}
