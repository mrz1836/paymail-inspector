package cli

import (
	"bytes"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

func TestBuildInfo_Resolve(t *testing.T) {
	stamped := func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v0.5.0"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "0123456789abcdef"},
				{Key: "vcs.time", Value: "2026-03-04T05:06:07Z"},
			},
		}, true
	}

	tests := []struct {
		name string
		in   BuildInfo
		read readBuildInfo
		want BuildInfo
	}{
		{"injected", BuildInfo{Version: "1.0.0", Commit: "abc", Date: "now"}, stamped, BuildInfo{Version: "1.0.0", Commit: "abc", Date: "now"}},
		{"from toolchain", BuildInfo{}, stamped, BuildInfo{Version: "0.5.0", Commit: "0123456", Date: "2026-03-04T05:06:07Z"}},
		{"partial", BuildInfo{Version: "2.0.0"}, stamped, BuildInfo{Version: "2.0.0", Commit: "0123456", Date: "2026-03-04T05:06:07Z"}},
		{"devel", BuildInfo{}, func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
		}, BuildInfo{Version: devVersion, Commit: unknownCommit, Date: unknownDate}},
		{"unavailable", BuildInfo{}, func() (*debug.BuildInfo, bool) { return nil, false }, BuildInfo{Version: devVersion, Commit: unknownCommit, Date: unknownDate}},
		{"no reader", BuildInfo{}, nil, BuildInfo{Version: devVersion, Commit: unknownCommit, Date: unknownDate}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.resolve(tt.read))
		})
	}
}

func TestBuildInfo_IsRelease(t *testing.T) {
	for version, want := range map[string]bool{
		"1.2.3":                                 true,
		"0.5.0-rc1":                             true,
		"":                                      false,
		devVersion:                              false,
		"0.4.4+dirty":                           false,
		"0.4.5-0.20261123174806-980d3c0abcde":   false,
		"0.4.5-0.20261123174806-980d3c0abcde+x": false,
	} {
		assert.Equal(t, want, BuildInfo{Version: version}.isRelease(), version)
	}
}

func TestBuildInfo_StringAndShortSHA(t *testing.T) {
	assert.Equal(t, "1.0.0 (commit: abc, built: today)", BuildInfo{Version: "1.0.0", Commit: "abc", Date: "today"}.String())
	assert.Equal(t, "abc", shortSHA("abc"))
	assert.Equal(t, "0123456", shortSHA("0123456789"))
}

func TestInvokedName(t *testing.T) {
	assert.Equal(t, aliasName, invokedName("/usr/local/bin/paymail"))
	assert.Equal(t, aliasName, invokedName(`C:\bin\paymail.exe`[7:]))
	assert.Equal(t, aliasName, invokedName("paymail.exe"))
	assert.Equal(t, appName, invokedName("/opt/paymail-inspector"))
	assert.Equal(t, appName, invokedName("cli.test"))
}

func TestNewApp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	a := newApp(aliasName, BuildInfo{Version: "3.0.0", Commit: "c", Date: "d"}, &stdout, &stderr)

	assert.Equal(t, aliasName, a.name)
	assert.Equal(t, "3.0.0", a.build.Version)
	assert.Equal(t, "paymail-inspector/3.0.0", a.userAgent())
	assert.Equal(t, cacheName, filepath.Base(a.cachePath))
	assert.NotEmpty(t, a.configDirs)
	assert.Len(t, a.providers, len(defaultProviders()))
	assert.Equal(t, "3.0.0", a.update.CurrentVersion)
	assert.NotNil(t, a.out)
	assert.NotNil(t, a.errOut)

	renamed := a.renamed(appName)
	assert.Equal(t, appName, renamed.name)
	assert.Equal(t, a.build, renamed.build)
	assert.NotSame(t, a.cfg, renamed.cfg)
}

func TestNewPaymailClient(t *testing.T) {
	client, err := newPaymailClient(clientOptions{userAgent: "test", nameServer: "127.0.0.1", tracing: true})
	require.NoError(t, err)
	assert.NotEmpty(t, client.GetBRFCs(), "go-paymail bundles the known BRFC specifications")

	client, err = newPaymailClient(clientOptions{userAgent: "test"})
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestDefaultPaths(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	configDirs, cachePath := defaultPaths()
	require.Len(t, configDirs, 2)
	assert.Equal(t, appName, filepath.Base(configDirs[0]))
	assert.Equal(t, legacyDir, filepath.Base(configDirs[1]))
	assert.Equal(t, filepath.Join(appName, cacheName), filepath.Join(filepath.Base(filepath.Dir(cachePath)), filepath.Base(cachePath)))
}

func TestRandomHex(t *testing.T) {
	got, err := randomHex(8)
	require.NoError(t, err)
	assert.Len(t, got, 16)
}

func TestProviderFor(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, "Example", h.providerFor("EXAMPLE.COM").Name)
	assert.Equal(t, provider{Domain: "unknown.io", Name: "unknown.io", Link: "https://unknown.io"}, h.providerFor("unknown.io"))

	for _, p := range defaultProviders() {
		assert.True(t, strings.HasPrefix(p.Link, "https://"), p.Domain)
		assert.NotEmpty(t, p.Name, p.Domain)
	}
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		in, alias, domain string
		wantErr           error
	}{
		{"Alice@Example.com", "alice", "example.com", nil},
		{" alice@example.com ", "alice", "example.com", nil},
		{"$alice", "alice", "handcash.io", nil},
		{"1alice", "alice", "relayx.io", nil},
		{"alice", "", "", errInvalidPaymail},
		{"alice@", "", "", errInvalidPaymail},
		{"alice@localhost", "", "", errInvalidPaymail},
	}

	for _, tt := range tests {
		got, err := parseAddress(tt.in)
		if tt.wantErr != nil {
			require.ErrorIs(t, err, tt.wantErr, tt.in)
			continue
		}

		require.NoError(t, err, tt.in)
		assert.Equal(t, tt.alias, got.Alias, tt.in)
		assert.Equal(t, tt.domain, got.Domain, tt.in)
		assert.Equal(t, tt.alias+"@"+tt.domain, got.Address, tt.in)
	}
}

func TestParseDomain(t *testing.T) {
	for in, want := range map[string]string{
		"Example.com":        "example.com",
		"https://www.x.io/a": "x.io",
		"alice@example.com":  "example.com",
		" sub.example.com\t": "sub.example.com",
	} {
		got, err := parseDomain(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}

	for _, in := range []string{"", "localhost", "@@", "bad@"} {
		_, err := parseDomain(in)
		require.Error(t, err, in)
	}
}

func TestParseHandle(t *testing.T) {
	for in, want := range map[string]string{
		"mrz":           "mrz",
		"MrZ":           "mrz",
		"$mr-z":         "mr-z",
		"1mrz":          "mrz",
		"mrz@relayx.io": "mrz",
		"a.b+c_d":       "a.b+c_d",
	} {
		got, err := parseHandle(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}

	_, err := parseHandle("%%%")
	require.ErrorIs(t, err, errInvalidHandle)

	_, err = parseHandle(strings.Repeat("a", maxHandleLength+1))
	require.ErrorIs(t, err, errInvalidHandle)
}

func TestSpecURL(t *testing.T) {
	for in, want := range map[string]string{
		"http://bsvalias.org/04-01-basic-address-resolution.html": specMirror + "04-01-basic-address-resolution.html",
		"https://www.bsvalias.org/index.html":                     specMirror + "index.html",
		"https://github.com/bitcoin-sv-specs/brfc-misc":           "https://github.com/bitcoin-sv-specs/brfc-misc",
		"": "",
	} {
		assert.Equal(t, want, specURL(in), in)
	}
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "abc", truncate("abc", 5))
	assert.Equal(t, "abcde...", truncate("abcdefgh", 5))
	assert.Empty(t, truncate("", 3))
}

func TestRenderTracing(t *testing.T) {
	var buf bytes.Buffer

	renderTracing(output.New(&buf, false), resty.TraceInfo{
		DNSLookup:     time.Millisecond,
		TotalTime:     2 * time.Second,
		IsConnWasIdle: true,
		ConnIdleTime:  time.Second,
		IsConnReused:  true,
	}, 200)

	out := buf.String()
	assert.Contains(t, out, "DNSLookup")
	assert.Contains(t, out, "1ms")
	assert.Contains(t, out, "2s [200]")
	assert.Contains(t, out, "IsConnWasIdle | true")
	assert.Contains(t, out, "IsConnReused  | true")

	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		assert.False(t, strings.HasSuffix(line, "|"), "no dangling separators: %q", line)
	}
}

func TestDetailsRender(t *testing.T) {
	var buf bytes.Buffer

	(&paymailDetails{Address: address{Address: testAddress}}).render(output.New(&buf, false))
	assert.Equal(t, "\n==========| Results for "+testAddress+"\n", buf.String(), "empty details print only the header")
}

func FuzzParseHandle(f *testing.F) {
	for _, seed := range []string{"mrz", "$mr-z", "1mrz", "mrz@relayx.io", "", "@", "$", "1", "%%%"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		handle, err := parseHandle(raw)
		if err != nil {
			return
		}

		require.NotEmpty(t, handle)
		require.LessOrEqual(t, len(handle), maxHandleLength)
		require.Equal(t, strings.ToLower(handle), handle)
		require.NotContains(t, handle, "@")
	})
}

func FuzzParseAddress(f *testing.F) {
	for _, seed := range []string{"alice@example.com", "$alice", "1alice", "alice", "@", "a@b", "a@b.c", " A@B.CO "} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		addr, err := parseAddress(raw)
		if err != nil {
			return
		}

		require.Equal(t, addr.Alias+"@"+addr.Domain, addr.Address)
		require.Contains(t, addr.Domain, ".")
	})
}
