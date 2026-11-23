package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errCobraUnknownCommand mimics cobra's root-level unknown command error.
var errCobraUnknownCommand = errors.New(`unknown command "x" for "paymail-inspector"`)

func TestRoot_HelpWithoutArguments(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run())

	out := h.output()
	assert.Contains(t, out, "Inspect, validate and resolve paymail addresses")
	assert.Contains(t, out, "paymail-inspector resolve "+exampleAddress)

	for _, name := range []string{"brfc", cmdCaps, "p2p", cmdResolve, "validate", cmdVerify, "whois", "version", "update"} {
		assert.Contains(t, out, "  "+name, "help should list %s", name)
	}

	assert.NotContains(t, out, "gen-docs", "the docs generator is hidden")
}

func TestRoot_VersionFlag(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("--version"))
	assert.Equal(t, "paymail-inspector 1.2.3 (commit: abc1234, built: 2026-01-02T03:04:05Z)\n", h.output())
}

func TestRoot_UsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"unknown command", []string{"bogus"}, `unknown command "bogus"`},
		{"unknown flag", []string{"--bogus"}, "unknown flag: --bogus"},
		{"missing argument", []string{cmdResolve}, "resolve requires a paymail address"},
		{"extra argument", []string{cmdCaps, "a.com", "b.com"}, "capabilities requires a domain or paymail address"},
		{"unknown subcommand", []string{"brfc", "bogus"}, `unknown command "bogus"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)

			assert.Equal(t, ExitUsage, h.run(tt.args...))
			assert.Contains(t, h.errOutput(), errorPrefix+tt.want)
			assert.Contains(t, h.errOutput(), usagePointer)
		})
	}
}

func TestRoot_CommandErrorExitsOne(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, ExitError, h.run(cmdResolve, "not-an-address"))
	assert.Contains(t, h.errOutput(), errorPrefix+"invalid paymail address")
	assert.NotContains(t, h.errOutput(), usagePointer)
}

func TestRoot_FlushCache(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run(cmdCaps, testDomain))
	require.FileExists(t, h.cachePath)

	h.reset()
	require.Equal(t, ExitOK, h.run("--flush-cache"))
	assert.Contains(t, h.output(), "Flushed the local response cache")
	assert.NoFileExists(t, h.cachePath)
	assert.NotContains(t, h.output(), "Usage:", "flushing alone should not print help")
}

func TestRoot_FlushCacheFailure(t *testing.T) {
	h := newHarness(t)

	// A directory where the cache file should be cannot be opened as a
	// cache; --flush-cache removes it rather than reporting a no-op success.
	require.NoError(t, os.MkdirAll(h.cachePath, 0o750))

	assert.Equal(t, ExitOK, h.run("--flush-cache"))
	assert.Contains(t, h.errOutput(), "Response cache disabled")
	assert.Contains(t, h.output(), "Flushed the local response cache")
	assert.NoFileExists(t, h.cachePath)
	assert.NoDirExists(t, h.cachePath)

	// When even removal fails, the flush fails loudly.
	require.NoError(t, os.MkdirAll(filepath.Join(h.cachePath, "child"), 0o750))
	h.reset()
	assert.Equal(t, ExitError, h.run("--flush-cache"))
	assert.Contains(t, h.errOutput(), "flush response cache")
	assert.NotContains(t, h.output(), "Flushed")
}

func TestRoot_FlushUnreadableCache(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions differ on windows")
	}

	h := newHarness(t)
	require.NoError(t, os.MkdirAll(filepath.Dir(h.cachePath), 0o750))
	require.NoError(t, os.WriteFile(h.cachePath, []byte("{}"), 0o000))

	if f, err := os.Open(h.cachePath); err == nil {
		_ = f.Close()

		t.Skip("running with permissions that ignore file modes")
	}

	require.Equal(t, ExitOK, h.run("--flush-cache"))
	assert.NoFileExists(t, h.cachePath, "an unreadable cache is removed")
}

func TestRoot_CachedResponsesAreReused(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress))
	require.Equal(t, 1, h.client.callCount("GetCapabilities"))
	require.Equal(t, 1, h.client.callCount("GetPKI"))

	h.reset()
	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress))
	assert.Equal(t, 1, h.client.callCount("GetCapabilities"), "capabilities come from the cache")
	assert.Equal(t, 1, h.client.callCount("GetPKI"), "pki comes from the cache")
	assert.Equal(t, 2, h.client.callCount("ResolveAddress"), "resolution is never cached")
	assert.Contains(t, h.output(), "(from cache)")

	h.reset()
	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress, "--no-cache"))
	assert.Equal(t, 2, h.client.callCount("GetCapabilities"), "--no-cache skips cached reads")
	assert.NotContains(t, h.output(), "(from cache)")
}

func TestRoot_FallbackSRVIsNotCached(t *testing.T) {
	h := newHarness(t)
	h.client.caps[testDomain] = h.client.caps[testSRVHost]
	delete(h.client.srv, testDomain) // the fake answers like go-paymail: <domain>:443

	require.Equal(t, ExitOK, h.run("resolve", testAddress))
	require.Equal(t, ExitOK, h.run("resolve", testAddress))
	assert.Equal(t, 2, h.client.callCount("GetSRVRecord"), "a fallback SRV record is looked up again")
	assert.Equal(t, 1, h.client.callCount("GetCapabilities"), "capabilities are still cached")
}

func TestRoot_UnreadableCacheEntryIsIgnored(t *testing.T) {
	h := newHarness(t)

	require.NoError(t, os.MkdirAll(filepath.Dir(h.cachePath), 0o750))
	require.NoError(t, os.WriteFile(h.cachePath, []byte(`{"capabilities:example.com":{"value":"not an object"}}`), 0o600))

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress))
	assert.Contains(t, h.errOutput(), "Ignoring unreadable cache entry")
	assert.Equal(t, 1, h.client.callCount("GetCapabilities"))
}

func TestRoot_CacheOpenFailureIsAWarning(t *testing.T) {
	h := newHarness(t)

	// The cache directory's parent is a file, so the cache cannot open.
	parent := filepath.Join(h.dir, "file")
	require.NoError(t, os.WriteFile(parent, nil, 0o600))
	h.cachePath = filepath.Join(parent, "cache", cacheName)

	require.Equal(t, ExitOK, h.run(cmdCaps, testDomain))
	assert.Contains(t, h.errOutput(), "Response cache disabled")
	assert.Contains(t, h.output(), "Listing 6 capabilities")
}

func TestRoot_CacheSaveFailureIsAWarning(t *testing.T) {
	h := newHarness(t)

	// Swap the cache directory for a file after the cache opened, so only
	// the final save fails.
	h.client.onCapabilities = func() {
		dir := filepath.Dir(h.cachePath)
		_ = os.RemoveAll(dir)
		_ = os.WriteFile(dir, nil, 0o600)
	}

	require.Equal(t, ExitOK, h.run(cmdCaps, testDomain))
	assert.Contains(t, h.errOutput(), "save response cache")
}

func TestRoot_NoColor(t *testing.T) {
	h := newHarness(t)
	h.setColor(true)

	require.Equal(t, ExitOK, h.run("version", "--no-color"))
	assert.NotContains(t, h.output(), "\x1b[")

	h.setColor(true)
	h.reset()
	require.Equal(t, ExitOK, h.run("version"))
	assert.Contains(t, h.output(), "\x1b[")
}

func TestRoot_ConfigFile(t *testing.T) {
	t.Run("explicit --config", func(t *testing.T) {
		h := newHarness(t)
		path := h.writeConfig(t, filepath.Join(h.dir, "custom"), "sender-handle: bob@"+testDomain+"\nsender-name: Bob\n")
		h.client.pki["bob@"+testDomain] = pkiFor("bob@"+testDomain, testPubKey)

		require.Equal(t, ExitOK, h.run(cmdResolve, testAddress, "--config", path))
		require.NotNil(t, h.client.lastSender)
		assert.Equal(t, "bob@"+testDomain, h.client.lastSender.SenderHandle)
		assert.Equal(t, "Bob", h.client.lastSender.SenderName)
	})

	t.Run("default search path", func(t *testing.T) {
		h := newHarness(t)
		h.writeConfig(t, h.configDirs[0], "bsvalias: \"2.0\"\n")

		assert.Equal(t, ExitError, h.run(cmdCaps, testDomain))
		assert.Contains(t, h.errOutput(), "bsvalias version mismatch")
	})

	t.Run("legacy ~/paymail fallback", func(t *testing.T) {
		h := newHarness(t)
		h.writeConfig(t, h.configDirs[1], "sender-name: Legacy\n")

		require.Equal(t, ExitOK, h.run(cmdResolve, testAddress))
		assert.Equal(t, "Legacy", h.client.lastSender.SenderName)
	})

	t.Run("missing explicit config", func(t *testing.T) {
		h := newHarness(t)

		assert.Equal(t, ExitError, h.run("version", "--config", filepath.Join(h.dir, "missing.yaml")))
		assert.Contains(t, h.errOutput(), "read config")
	})

	t.Run("invalid config", func(t *testing.T) {
		h := newHarness(t)
		h.writeConfig(t, h.configDirs[0], "sender-name: [unterminated\n")

		// A discovered config is optional: a broken one is reported and
		// skipped instead of failing every command.
		assert.Equal(t, ExitOK, h.run("version"))
		assert.Contains(t, h.errOutput(), "Ignoring config: read config")
	})

	t.Run("invalid explicit config", func(t *testing.T) {
		h := newHarness(t)
		path := h.writeConfig(t, filepath.Join(h.dir, "custom"), "sender-name: [unterminated\n")

		assert.Equal(t, ExitError, h.run("version", "--config", path))
		assert.Contains(t, h.errOutput(), "read config")
	})

	t.Run("config path cannot be inspected", func(t *testing.T) {
		h := newHarness(t)

		// A config path whose parent is a file fails with ENOTDIR, which is
		// not "no config": it is reported, and the command still runs.
		parent := filepath.Join(h.dir, "file")
		require.NoError(t, os.WriteFile(parent, nil, 0o600))
		h.configDirs = []string{filepath.Join(parent, "sub")}

		assert.Equal(t, ExitOK, h.run("version"))
		assert.Contains(t, h.errOutput(), "Ignoring config: inspect config")
	})

	t.Run("legacy BSVALIAS variable", func(t *testing.T) {
		h := newHarness(t)
		t.Setenv("BSVALIAS", "2.0")

		assert.Equal(t, ExitError, h.run(cmdCaps, testDomain))
		assert.Contains(t, h.errOutput(), `expected "2.0"`)
	})

	t.Run("environment variable", func(t *testing.T) {
		h := newHarness(t)
		t.Setenv("PAYMAIL_INSPECTOR_SENDER_NAME", "From Env")

		require.Equal(t, ExitOK, h.run(cmdResolve, testAddress))
		assert.Equal(t, "From Env", h.client.lastSender.SenderName)
	})
}

func TestRoot_Suggestions(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		suggestion string
		helpFor    string
	}{
		{"root typo", []string{"rsolve", testAddress}, "resolve", "paymail-inspector"},
		{"suggest for", []string{"lookup"}, "whois", "paymail-inspector"},
		{"group typo", []string{"brfc", "lsit"}, "list", "paymail-inspector brfc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)

			assert.Equal(t, ExitUsage, h.run(tt.args...))
			assert.Contains(t, h.errOutput(), "Did you mean this?")
			assert.Contains(t, h.errOutput(), "\t"+tt.suggestion)
			assert.Contains(t, h.errOutput(), "Run '"+tt.helpFor+" --help' for usage.")
		})
	}
}

func TestRoot_UsageHelpPointsAtFailingCommand(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, ExitUsage, h.run(cmdResolve))
	assert.Contains(t, h.errOutput(), "Run 'paymail-inspector resolve --help' for usage.")
}

func TestRoot_RetiredDocsFlag(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("--docs"))
	assert.Contains(t, h.errOutput(), "Flag --docs is retired: generate the command reference with the gen-docs command")
	assert.Contains(t, h.output(), "Usage:")
}

func TestRoot_CompletionSkipsHooks(t *testing.T) {
	h := newHarness(t)
	h.writeConfig(t, h.configDirs[0], "sender-name: [unterminated\n")

	require.Equal(t, ExitOK, h.run(cobra.ShellCompRequestCmd, cmdResolve, "--sender"))
	assert.Contains(t, h.output(), "--sender-handle")
	assert.NotContains(t, h.errOutput(), "Ignoring config", "completion does not load config")
	assert.NoFileExists(t, h.cachePath)

	h.reset()
	require.Equal(t, ExitOK, h.run(cobra.ShellCompNoDescRequestCmd, "wh"))
	assert.Contains(t, h.output(), "whois")
}

func TestRoot_DefaultConfigHint(t *testing.T) {
	h := newHarness(t)
	assert.Equal(t, filepath.Join(h.configDirs[0], configName), h.defaultConfigHint())

	h.configDirs = nil
	assert.Equal(t, configName, h.defaultConfigHint())
}

func TestExecute(t *testing.T) {
	saved := os.Args

	t.Cleanup(func() { os.Args = saved })

	os.Args = []string{"/usr/local/bin/paymail", "--version"}

	assert.Equal(t, ExitOK, Execute(context.Background(), BuildInfo{Version: testVersion}))

	os.Args = []string{"paymail-inspector", "--definitely-not-a-flag"}

	assert.Equal(t, ExitUsage, Execute(context.Background(), BuildInfo{}))
}

func TestIsUsageError(t *testing.T) {
	assert.False(t, isUsageError(nil))
	assert.False(t, isUsageError(errFake))
	assert.True(t, isUsageError(newUsageError("bad %s", "input")))
	assert.True(t, isUsageError(usageError{err: errFake}))
	assert.Equal(t, errFake.Error(), usageError{err: errFake}.Error())
	assert.True(t, isUsageError(errCobraUnknownCommand))
}
