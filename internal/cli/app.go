package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/fatih/color"
	selfupdate "github.com/mrz1836/go-selfupdate"
	"github.com/spf13/viper"

	"github.com/mrz1836/paymail-inspector/internal/cache"
	"github.com/mrz1836/paymail-inspector/internal/output"
)

// Application identity.
const (
	appName    = "paymail-inspector" // binary and repository name
	aliasName  = "paymail"           // short name the binary answers to via a symlink
	repoOwner  = "mrz1836"
	envPrefix  = "PAYMAIL_INSPECTOR" // prefix for every environment variable
	configName = "config.yaml"
	cacheName  = "cache.json"
	legacyDir  = "paymail" // ~/paymail: where releases before the rename kept their config
	cacheTTL   = time.Hour
)

// Configuration keys shared by flags, the config file, and the environment.
const (
	keyBsvAlias     = "bsvalias"
	keySenderHandle = "sender-handle"
	keySenderName   = "sender-name"
)

// paymailClient is the slice of the go-paymail client the CLI uses. The real
// paymail.ClientInterface satisfies it; tests supply a fake.
type paymailClient interface {
	CheckDNSSEC(domain string) (result *paymail.DNSCheckResult)
	CheckSSL(host string) (valid bool, err error)
	GetBRFCs() []*paymail.BRFCSpec
	GetCapabilities(target string, port int) (response *paymail.CapabilitiesResponse, err error)
	GetP2PPaymentDestination(p2pURL, alias, domain string, paymentRequest *paymail.PaymentRequest) (response *paymail.PaymentDestinationResponse, err error)
	GetPKI(pkiURL, alias, domain string) (response *paymail.PKIResponse, err error)
	GetPublicProfile(publicProfileURL, alias, domain string) (response *paymail.PublicProfileResponse, err error)
	GetSRVRecord(service, protocol, domainName string) (srv *net.SRV, err error)
	ResolveAddress(resolutionURL, alias, domain string, senderRequest *paymail.SenderRequest) (response *paymail.ResolutionResponse, err error)
	ValidateSRVRecord(ctx context.Context, srv *net.SRV, port, priority, weight uint16) error
	VerifyPubKey(verifyURL, alias, domain, pubKey string) (response *paymail.VerificationResponse, err error)
}

// clientOptions configures one paymail client.
type clientOptions struct {
	userAgent  string
	nameServer string
	tracing    bool
}

// clientFactory builds a paymail client.
type clientFactory func(opts clientOptions) (paymailClient, error)

// newPaymailClient is the production clientFactory.
func newPaymailClient(opts clientOptions) (paymailClient, error) {
	ops := []paymail.ClientOps{paymail.WithUserAgent(opts.userAgent)}
	if opts.tracing {
		ops = append(ops, paymail.WithRequestTracing())
	}

	if opts.nameServer != "" {
		ops = append(ops, paymail.WithNameServer(opts.nameServer))
	}

	client, err := paymail.NewClient(ops...)
	if err != nil {
		return nil, fmt.Errorf("create paymail client: %w", err)
	}

	return client, nil
}

// globalFlags holds the persistent flags every command shares.
type globalFlags struct {
	configFile  string
	noCache     bool
	flushCache  bool
	skipTracing bool
	noColor     bool
}

// app carries everything a command needs. One app serves one run.
type app struct {
	name   string // the name the binary was invoked as
	build  BuildInfo
	stdout io.Writer
	stderr io.Writer
	out    *output.Printer // stdout
	errOut *output.Printer // stderr
	cfg    *viper.Viper
	flags  globalFlags

	// colorOK reports whether stdout can show color at all; --no-color
	// can only turn it off.
	colorOK bool

	newClient  clientFactory
	configDirs []string // searched in order when --config is not given
	configHint string   // overrides the config path shown in help (used by gen-docs)
	cachePath  string
	providers  []provider
	update     selfupdate.Config
	exePath    string              // the running binary, to tailor the update notice to its installer
	getenv     func(string) string // environment lookup for the update notice; nil means os.Getenv
	randomHex  func(n int) (string, error)
	now        func() time.Time

	storeOnce sync.Once
	store     cache.Store
	storeErr  error
}

// newApp returns the production app writing to stdout and stderr.
func newApp(name string, info BuildInfo, stdout, stderr io.Writer) *app {
	build := info.resolve(debug.ReadBuildInfo)
	configDirs, cachePath := defaultPaths()
	exe, _ := os.Executable() // best effort: an unknown path just keeps the default notice

	a := &app{
		name:       name,
		build:      build,
		stdout:     stdout,
		stderr:     stderr,
		colorOK:    !color.NoColor,
		newClient:  newPaymailClient,
		configDirs: configDirs,
		cachePath:  cachePath,
		providers:  defaultProviders(),
		exePath:    exe,
		randomHex:  randomHex,
		now:        time.Now,
	}
	a.cfg = newConfig()
	a.update = a.defaultUpdateConfig()
	a.setColor(a.colorOK)

	return a
}

// renamed returns a fresh app identical to a but presenting itself as name.
func (a *app) renamed(name string) *app {
	b := &app{
		name:       name,
		build:      a.build,
		stdout:     a.stdout,
		stderr:     a.stderr,
		cfg:        newConfig(),
		colorOK:    a.colorOK,
		newClient:  a.newClient,
		configDirs: a.configDirs,
		cachePath:  a.cachePath,
		providers:  a.providers,
		update:     a.update,
		exePath:    a.exePath,
		getenv:     a.getenv,
		randomHex:  a.randomHex,
		now:        a.now,
	}
	b.setColor(a.colorOK)

	return b
}

// setColor rebuilds both printers with color on or off.
func (a *app) setColor(on bool) {
	a.out = output.New(a.stdout, on)
	a.errOut = output.New(a.stderr, on)
}

// errLine writes a formatted error line to stderr.
func (a *app) errLine(format string, args ...any) {
	a.errOut.Error(format, args...)
}

// newConfig returns a configuration registry reading PAYMAIL_INSPECTOR_*
// environment variables, e.g. PAYMAIL_INSPECTOR_SENDER_HANDLE.
func newConfig() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	v.SetDefault(keyBsvAlias, paymail.DefaultBsvAliasVersion)

	// Releases before the rename read an unprefixed BSVALIAS variable.
	_ = v.BindEnv(keyBsvAlias, envPrefix+"_BSVALIAS", "BSVALIAS")

	return v
}

// userAgent identifies the CLI to paymail providers.
func (a *app) userAgent() string {
	return appName + "/" + a.build.Version
}

// cacheStore returns the response cache, opening it on first use. A cache
// that cannot be opened is reported once and replaced by a no-op store, so
// a broken cache never fails a command.
func (a *app) cacheStore() cache.Store {
	a.storeOnce.Do(func() {
		store, err := cache.Open(a.cachePath)
		if err != nil {
			a.storeErr = err
			a.store = cache.Nop{}
			a.errOut.Warn("Response cache disabled: %s", err)

			return
		}

		a.store = store
	})

	return a.store
}

// closeCache persists the cache if it was opened during the run.
func (a *app) closeCache() error {
	if a.store == nil {
		return nil
	}

	if err := a.store.Close(); err != nil {
		return fmt.Errorf("save response cache: %w", err)
	}

	return nil
}

// defaultPaths returns the config search path and cache file location. The
// legacy ~/paymail directory is still searched for a config file so upgrades
// keep their settings.
func defaultPaths() (configDirs []string, cachePath string) {
	if dir, err := os.UserConfigDir(); err == nil {
		configDirs = append(configDirs, filepath.Join(dir, appName))
	}

	if home, err := os.UserHomeDir(); err == nil {
		configDirs = append(configDirs, filepath.Join(home, legacyDir))
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	return configDirs, filepath.Join(cacheDir, appName, cacheName)
}

// randomHex returns n random bytes, hex-encoded.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return hex.EncodeToString(b), nil
}

// invokedName returns the command name to present: "paymail" when the binary
// was started through a symlink of that name, otherwise paymail-inspector.
func invokedName(arg0 string) string {
	base := strings.TrimSuffix(filepath.Base(arg0), ".exe")
	if base == aliasName {
		return aliasName
	}

	return appName
}
