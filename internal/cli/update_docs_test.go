package cli

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	selfupdate "github.com/mrz1836/go-selfupdate"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRelease returns a release whose archive and checksums point at baseURL.
func fakeRelease(baseURL, tag string) *selfupdate.Release {
	ver := strings.TrimPrefix(tag, "v")

	return &selfupdate.Release{
		TagName: tag,
		Assets: []selfupdate.ReleaseAsset{
			{Name: fmt.Sprintf("%s_%s_%s_%s.tar.gz", appName, ver, runtime.GOOS, runtime.GOARCH), BrowserDownloadURL: baseURL + "/archive"},
			{Name: fmt.Sprintf("%s_%s_checksums.txt", appName, ver), BrowserDownloadURL: baseURL + "/checksums"},
		},
	}
}

// withUpdateServer points the harness's update config at a fake release
// served by a server that answers 503 to every download, and returns the
// target binary path, which must never be replaced.
func withUpdateServer(t *testing.T, h *harness, tag string) string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	target := filepath.Join(h.dir, appName)
	require.NoError(t, os.WriteFile(target, []byte("old"), 0o600))

	h.update.TargetPath = target
	h.update.Client = srv.Client()
	h.update.Source = fakeReleaseSource{rel: fakeRelease(srv.URL, tag)}

	return target
}

func TestUpdate_CheckReportsNewVersion(t *testing.T) {
	h := newHarness(t)
	withUpdateServer(t, h, "v9.9.9")

	require.Equal(t, ExitOK, h.run("update", "--check"))
	assert.Contains(t, h.output(), "is available: 1.2.3 -> v9.9.9")
}

func TestUpdate_CheckUpToDate(t *testing.T) {
	h := newHarness(t)
	withUpdateServer(t, h, "v1.2.3")

	require.Equal(t, ExitOK, h.run("upgrade", "--check"))
	assert.Contains(t, h.output(), "is up to date")
}

func TestUpdate_InstallFailureLeavesBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("self-update install is not supported on windows")
	}

	h := newHarness(t)
	target := withUpdateServer(t, h, "v9.9.9")

	assert.Equal(t, ExitError, h.run("update"))
	assert.Contains(t, h.errOutput(), errorPrefix)

	data, err := os.ReadFile(target) //nolint:gosec // G304: test reads its own temp file
	require.NoError(t, err)
	assert.Equal(t, "old", string(data), "a failed update must not touch the binary")
}

func TestUpdate_SourceFailure(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, ExitError, h.run("update", "--check"))
	assert.Contains(t, h.errOutput(), errFake.Error())
}

func TestUpdate_DefaultConfig(t *testing.T) {
	h := newHarness(t)

	cfg := h.defaultUpdateConfig()
	assert.Equal(t, repoOwner, cfg.Owner)
	assert.Equal(t, appName, cfg.Repo)
	assert.Equal(t, appName, cfg.BinaryName)
	assert.Equal(t, testVersion, cfg.CurrentVersion)
	assert.Equal(t, "PAYMAIL_INSPECTOR_GITHUB_TOKEN", cfg.TokenEnvVar)

	h.build.Version = devVersion
	assert.Empty(t, h.defaultUpdateConfig().CurrentVersion, "development builds never look like a release")
}

func TestUpgradeCommand(t *testing.T) {
	dir := t.TempDir()

	// Homebrew: the binary lives in a Cellar.
	cellar := filepath.Join(dir, "Cellar", appName, "1.0.0", "bin")
	require.NoError(t, os.MkdirAll(cellar, 0o750))
	brewBin := filepath.Join(cellar, appName)
	require.NoError(t, os.WriteFile(brewBin, nil, 0o700)) //nolint:gosec // G306: a fake executable in a temp dir

	// go install: the binary lives in GOBIN.
	gobin := filepath.Join(dir, "gobin")
	require.NoError(t, os.MkdirAll(gobin, 0o750))
	goBin := filepath.Join(gobin, appName)
	require.NoError(t, os.WriteFile(goBin, nil, 0o700)) //nolint:gosec // G306: a fake executable in a temp dir
	t.Setenv("GOBIN", gobin)

	// A release install: a writable file anywhere else.
	releaseBin := filepath.Join(dir, "local", appName)
	require.NoError(t, os.MkdirAll(filepath.Dir(releaseBin), 0o750))
	require.NoError(t, os.WriteFile(releaseBin, nil, 0o700)) //nolint:gosec // G306: a fake executable in a temp dir

	assert.Equal(t, "brew upgrade paymail-inspector", upgradeCommand(aliasName, brewBin))
	assert.Equal(t, "go install github.com/mrz1836/paymail-inspector/cmd/paymail-inspector@latest", upgradeCommand(aliasName, goBin))
	assert.Equal(t, "paymail update", upgradeCommand(aliasName, releaseBin))
	assert.Equal(t, "paymail-inspector update", upgradeCommand(appName, ""))

	if runtime.GOOS != "windows" {
		readOnly := filepath.Join(dir, "system", appName)
		require.NoError(t, os.MkdirAll(filepath.Dir(readOnly), 0o750))
		require.NoError(t, os.WriteFile(readOnly, nil, 0o500)) //nolint:gosec // G306: a fake read-only executable

		assert.Equal(t, "paymail update", upgradeCommand(aliasName, readOnly), "other managers keep the default")
	}
}

func TestBannerEnv(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, "1", h.bannerEnv(noUpdateEnv))
	assert.Empty(t, h.bannerEnv("NO_COLOR"))

	h.flags.noColor = true
	assert.Equal(t, "1", h.bannerEnv("NO_COLOR"), "--no-color reaches the update notice")

	h.getenv = nil

	t.Setenv("PAYMAIL_INSPECTOR_TEST_VAR", "value")
	assert.Equal(t, "value", h.bannerEnv("PAYMAIL_INSPECTOR_TEST_VAR"))
}

func TestUpdate_HelpUsesInvokedName(t *testing.T) {
	h := newHarness(t)
	h.name = aliasName

	require.Equal(t, ExitOK, h.run("update", "--help"))
	assert.Contains(t, h.output(), "paymail update --check")
}

func TestDocs_GeneratesReference(t *testing.T) {
	h := newHarness(t)
	dir := filepath.Join(h.dir, "docs")

	require.Equal(t, ExitOK, h.run("gen-docs", "--dir", dir))
	assert.Contains(t, h.output(), "Generated the command reference in "+dir)

	for _, name := range []string{"paymail.md", "paymail_resolve.md", "paymail_brfc_generate.md", "paymail_update.md", "paymail_completion_zsh.md"} {
		assert.FileExists(t, filepath.Join(dir, name))
	}

	assert.NoFileExists(t, filepath.Join(dir, "paymail_gen-docs.md"), "hidden commands are not documented")

	data, err := os.ReadFile(filepath.Join(dir, "paymail_resolve.md")) //nolint:gosec // G304: test reads its own temp file
	require.NoError(t, err)
	assert.Contains(t, string(data), "paymail resolve "+exampleAddress)
	assert.Contains(t, string(data), "```text\n", "the ASCII banner is fenced")
	assert.Contains(t, string(data), docsConfigHint, "docs show a portable config path")
	assert.NotContains(t, string(data), h.configDirs[0], "docs never show this machine's paths")
	assert.NotContains(t, string(data), "Auto generated", "docs are reproducible")
}

func TestFenceArt(t *testing.T) {
	root := &cobra.Command{Use: "root", Long: "\nART\nMORE\n\nText"}
	plain := &cobra.Command{Use: "plain", Long: "No art here"}
	noBreak := &cobra.Command{Use: "nobreak", Long: "\nonly art"}
	root.AddCommand(plain, noBreak)

	fenceArt(root)

	assert.Equal(t, "```text\nART\nMORE\n```\n\nText", root.Long)
	assert.Equal(t, "No art here", plain.Long)
	assert.Equal(t, "\nonly art", noBreak.Long)
}

func TestDocs_CustomNameAndFailure(t *testing.T) {
	h := newHarness(t)
	dir := filepath.Join(h.dir, "docs")

	require.Equal(t, ExitOK, h.run("gen-docs", "--dir", dir, "--name", appName))
	assert.FileExists(t, filepath.Join(dir, "paymail-inspector_whois.md"))

	blocker := filepath.Join(h.dir, "blocker")
	require.NoError(t, os.WriteFile(blocker, nil, 0o600))

	h.reset()
	assert.Equal(t, ExitError, h.run("gen-docs", "--dir", filepath.Join(blocker, "docs")))
	assert.Contains(t, h.errOutput(), "create docs directory")
}

func TestDocs_GenerateFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions differ on windows")
	}

	h := newHarness(t)
	dir := filepath.Join(h.dir, "readonly")
	require.NoError(t, os.MkdirAll(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) }) //nolint:gosec // G302: restore so the temp dir can be removed

	if f, err := os.Create(filepath.Join(dir, "probe")); err == nil { //nolint:gosec // G304: probe file in a temp dir
		_ = f.Close()

		t.Skip("running with permissions that ignore directory modes")
	}

	assert.Equal(t, ExitError, h.run("gen-docs", "--dir", dir))
	assert.Contains(t, h.errOutput(), "generate docs")
}

func TestVersion_Command(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("version"))

	out := h.output()
	assert.Contains(t, out, "Version      : "+testVersion)
	assert.Contains(t, out, "Commit       : abc1234")
	assert.Contains(t, out, "Built        : 2026-01-02T03:04:05Z")
	assert.Contains(t, out, "Go           : "+runtime.Version())
	assert.Contains(t, out, "Platform     : "+runtime.GOOS+"/"+runtime.GOARCH)
}
