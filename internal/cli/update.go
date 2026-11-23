package cli

import (
	"os"
	"strings"

	selfupdate "github.com/mrz1836/go-selfupdate"
	"github.com/mrz1836/go-selfupdate/cobracmd"
	"github.com/mrz1836/go-selfupdate/notify"
	"github.com/spf13/cobra"
)

// tokenEnvVar is consulted before GITHUB_TOKEN and GH_TOKEN when the update
// check hits GitHub's API rate limits.
const tokenEnvVar = envPrefix + "_GITHUB_TOKEN"

// defaultUpdateConfig returns the production self-update config. The
// compiled-in version is passed explicitly, so a binary run from outside
// PATH still updates itself; a development build reports an empty version,
// which every real release outranks.
func (a *app) defaultUpdateConfig() selfupdate.Config {
	version := ""
	if a.build.isRelease() {
		version = a.build.Version
	}

	return selfupdate.Config{
		Owner:          repoOwner,
		Repo:           appName,
		BinaryName:     appName,
		CurrentVersion: version,
		TokenEnvVar:    tokenEnvVar,
	}
}

// attachUpdateCommand registers `update` (alias `upgrade`, flags --check,
// --force, --verbose) and the passive "a new version is available" notice.
//
// Updates install only from the GitHub release archives: the SHA-256 checksum
// is verified and the binary is replaced atomically, and a binary owned by
// `go install` or Homebrew is refused rather than overwritten. The binary
// behind a `paymail` symlink is resolved and updated in place, so the alias
// keeps working. The notice derives its opt-out from the binary name
// (PAYMAIL_INSPECTOR_NO_UPDATE_CHECK); the shared NO_UPDATE_CHECK and CI also
// disable it.
func attachUpdateCommand(root *cobra.Command, a *app) *cobra.Command {
	cfg := a.update

	// Errors are reported once, by run, like every other command's.
	cmd := cobracmd.New(cfg)
	cmd.Example = a.examples("update", "update --check", "update --force")
	root.AddCommand(cmd)

	cobracmd.AttachBanner(root, notify.Config{
		Owner:          cfg.Owner,
		Repo:           cfg.Repo,
		BinaryName:     cfg.BinaryName,
		CurrentVersion: cfg.CurrentVersion,
		Source:         cfg.Source,
		Client:         cfg.Client,
		TokenEnvVar:    cfg.TokenEnvVar,
		UpgradeCommand: upgradeCommand(a.name, a.exePath),
		BannerOut:      a.stderr,
		Getenv:         a.bannerEnv,
	})

	return cmd
}

// modulePath is the Go module path, used in the go install upgrade hint.
const modulePath = "github.com/mrz1836/" + appName

// upgradeCommand returns the command the update notice suggests, matching
// how the binary at exe was installed: `update` refuses to replace a binary
// that Homebrew or `go install` owns, so those users are pointed at the
// installer that does own it.
func upgradeCommand(name, exe string) string {
	managed, reason := selfupdate.DetectManaged(exe)

	switch {
	case !managed:
		return name + " update"
	case strings.Contains(reason, "Homebrew"):
		return "brew upgrade " + appName
	case strings.Contains(reason, "Go binary directory"):
		return "go install " + modulePath + "/cmd/" + appName + "@latest"
	default:
		return name + " update"
	}
}

// bannerEnv is the environment the update notice sees. It honors --no-color
// by reporting NO_COLOR as set, since the notice reads only the environment.
func (a *app) bannerEnv(key string) string {
	if key == "NO_COLOR" && a.flags.noColor {
		return "1"
	}

	if a.getenv != nil {
		return a.getenv(key)
	}

	return os.Getenv(key)
}
