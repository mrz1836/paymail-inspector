package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bsv-blockchain/go-paymail"
	selfupdate "github.com/mrz1836/go-selfupdate"
	"github.com/stretchr/testify/require"
)

// Fixtures shared by the command tests. Nothing here resolves to a real host:
// every request is answered by fakeClient.
const (
	testDomain   = "example.com"
	testAlias    = "alice"
	testAddress  = testAlias + "@" + testDomain
	testPubKey   = "02ead23149a1e33df17325ec7a7ba9e0b20c674c57c630f527d69b866aa9b65b10"
	testOutput   = "76a914da31dec79fe8e1d9af38ee883540f286ac99520f88ac"
	testBTCAddr  = "1Lti3s6AQNKTSgxnTyBREMa6XdHLBnPSKa"
	testVersion  = "1.2.3"
	testSRVHost  = "paymail.example.com"
	pkiTemplate  = "https://" + testSRVHost + "/id/{alias}@{domain.tld}"
	resolveTmpl  = "https://" + testSRVHost + "/address/{alias}@{domain.tld}"
	profileTmpl  = "https://" + testSRVHost + "/profile/{alias}@{domain.tld}"
	p2pTemplate  = "https://" + testSRVHost + "/p2p/{alias}@{domain.tld}"
	verifyTmpl   = "https://" + testSRVHost + "/verify/{alias}@{domain.tld}/{pubkey}"
	noUpdateEnv  = "NO_UPDATE_CHECK"
	senderPub    = "0352530c305378fd9dfd99f8c8c44e9092efa7c1674b61d4e9be65f92aa7a77bbe"
	cmdCaps      = "capabilities"
	cmdResolve   = "resolve"
	cmdVerify    = "verify"
	errorPrefix  = "Error: "
	usagePointer = "--help' for usage."
)

// errFake is returned by fakeClient when a test asks it to fail.
var errFake = errors.New("fake failure")

// TestMain makes sure no test can trigger a real update check.
func TestMain(m *testing.M) {
	_ = os.Setenv(noUpdateEnv, "1")

	os.Exit(m.Run())
}

// fakeClient is an in-memory paymailClient. Lookups are keyed by domain or
// address; a missing entry is answered like a provider that has nothing.
type fakeClient struct {
	mu    sync.Mutex
	calls map[string]int

	brfcs []*paymail.BRFCSpec

	srv            map[string]*net.SRV
	srvErr         error
	srvValidateErr error

	caps           map[string]*paymail.CapabilitiesResponse // keyed by target host
	capsErr        error
	onCapabilities func() // optional hook, run on every GetCapabilities call

	pki    map[string]*paymail.PKIResponse // keyed by address
	pkiErr map[string]error

	profiles   map[string]*paymail.PublicProfileResponse
	profileErr error

	resolution *paymail.ResolutionResponse
	resolveErr error
	lastSender *paymail.SenderRequest

	destination *paymail.PaymentDestinationResponse
	p2pErr      error
	lastPayment *paymail.PaymentRequest

	verifyMatch bool
	verifyErr   error

	dnssec   *paymail.DNSCheckResult
	sslValid bool
	sslErr   error
}

// newFakeClient returns a client for a healthy provider of testAddress.
func newFakeClient() *fakeClient {
	return &fakeClient{
		calls: map[string]int{},
		brfcs: []*paymail.BRFCSpec{
			{ID: "c9ca939e2f69", Title: "bsvalias Payment Addressing (Basic Address Resolution)", Author: "andy (nChain)", Version: "1", Alias: "paymentDestination"},
			{ID: "ba4ba843a7d6", Title: "bsvalias Public Key Infrastructure", Author: "andy (nChain)", Version: "1", Alias: "pki", URL: "http://bsvalias.org/03-public-key-infrastructure.html"},
			{ID: "deadbeef0000", Title: "Bogus ID", Author: "nobody", Version: "1"},
			{ID: "", Title: "No ID", Version: "1"},
		},
		srv: map[string]*net.SRV{
			testDomain: {Target: testSRVHost, Port: 443, Priority: 10, Weight: 10},
		},
		caps: map[string]*paymail.CapabilitiesResponse{
			testSRVHost: capabilities(map[string]any{
				paymail.BRFCPki:                   pkiTemplate,
				paymail.BRFCPaymentDestination:    resolveTmpl,
				paymail.BRFCPublicProfile:         profileTmpl,
				paymail.BRFCP2PPaymentDestination: p2pTemplate,
				paymail.BRFCVerifyPublicKeyOwner:  verifyTmpl,
				paymail.BRFCSenderValidation:      false,
			}),
		},
		pki: map[string]*paymail.PKIResponse{
			testAddress: pkiFor(testAddress, testPubKey),
		},
		pkiErr: map[string]error{},
		profiles: map[string]*paymail.PublicProfileResponse{
			testAddress: {PublicProfilePayload: paymail.PublicProfilePayload{Name: "Alice", Avatar: "https://example.com/alice.png"}},
		},
		resolution: &paymail.ResolutionResponse{
			StandardResponse:  paymail.StandardResponse{StatusCode: http.StatusOK},
			ResolutionPayload: paymail.ResolutionPayload{Output: testOutput, Address: testBTCAddr},
		},
		destination: &paymail.PaymentDestinationResponse{
			StandardResponse: paymail.StandardResponse{StatusCode: http.StatusOK},
			PaymentDestinationPayload: paymail.PaymentDestinationPayload{
				Reference: "ref-123",
				Outputs:   []*paymail.PaymentOutput{{Script: testOutput, Satoshis: 1000, Address: testBTCAddr}},
			},
		},
		verifyMatch: true,
		dnssec:      &paymail.DNSCheckResult{DNSSEC: true, Domain: testSRVHost},
		sslValid:    true,
	}
}

// capabilities builds a bsvalias 1.0 capabilities response.
func capabilities(values map[string]any) *paymail.CapabilitiesResponse {
	return &paymail.CapabilitiesResponse{
		StandardResponse:    paymail.StandardResponse{StatusCode: http.StatusOK},
		CapabilitiesPayload: paymail.CapabilitiesPayload{BsvAlias: paymail.DefaultBsvAliasVersion, Capabilities: values},
	}
}

// pkiFor builds a PKI response.
func pkiFor(handle, pubKey string) *paymail.PKIResponse {
	return &paymail.PKIResponse{
		StandardResponse: paymail.StandardResponse{StatusCode: http.StatusOK},
		PKIPayload:       paymail.PKIPayload{BsvAlias: paymail.DefaultBsvAliasVersion, Handle: handle, PubKey: pubKey},
	}
}

func (f *fakeClient) CheckDNSSEC(string) *paymail.DNSCheckResult {
	f.count("CheckDNSSEC")
	return f.dnssec
}

func (f *fakeClient) CheckSSL(string) (bool, error) {
	f.count("CheckSSL")
	return f.sslValid, f.sslErr
}

func (f *fakeClient) GetBRFCs() []*paymail.BRFCSpec {
	f.count("GetBRFCs")
	return f.brfcs
}

func (f *fakeClient) GetCapabilities(target string, _ int) (*paymail.CapabilitiesResponse, error) {
	f.count("GetCapabilities")

	if f.onCapabilities != nil {
		f.onCapabilities()
	}

	if f.capsErr != nil {
		return nil, f.capsErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if caps, ok := f.caps[target]; ok {
		return caps, nil
	}

	return nil, fmt.Errorf("no capabilities for %s: %w", target, errFake)
}

func (f *fakeClient) GetP2PPaymentDestination(_, _, _ string, request *paymail.PaymentRequest) (*paymail.PaymentDestinationResponse, error) {
	f.count("GetP2PPaymentDestination")
	f.lastPayment = request

	return f.destination, f.p2pErr
}

func (f *fakeClient) GetPKI(_, alias, domain string) (*paymail.PKIResponse, error) {
	f.count("GetPKI")

	f.mu.Lock()
	defer f.mu.Unlock()

	addr := alias + "@" + domain
	if err, ok := f.pkiErr[addr]; ok {
		return &paymail.PKIResponse{StandardResponse: paymail.StandardResponse{StatusCode: statusFor(err)}}, err
	}

	if pki, ok := f.pki[addr]; ok {
		return pki, nil
	}

	return &paymail.PKIResponse{StandardResponse: paymail.StandardResponse{StatusCode: http.StatusNotFound}},
		fmt.Errorf("code 404, message: not found: %w", paymail.ErrPKIBadResponse)
}

func (f *fakeClient) GetPublicProfile(_, alias, domain string) (*paymail.PublicProfileResponse, error) {
	f.count("GetPublicProfile")

	if f.profileErr != nil {
		return nil, f.profileErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if profile, ok := f.profiles[alias+"@"+domain]; ok {
		return profile, nil
	}

	return &paymail.PublicProfileResponse{}, nil
}

func (f *fakeClient) GetSRVRecord(_, _, domain string) (*net.SRV, error) {
	f.count("GetSRVRecord")

	if f.srvErr != nil {
		return nil, f.srvErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if srv, ok := f.srv[domain]; ok {
		return srv, nil
	}

	// Like go-paymail: no SRV record means <domain>:443.
	return &net.SRV{Target: domain, Port: 443, Priority: 10, Weight: 10}, nil
}

func (f *fakeClient) ResolveAddress(_, _, _ string, sender *paymail.SenderRequest) (*paymail.ResolutionResponse, error) {
	f.count("ResolveAddress")
	f.lastSender = sender

	return f.resolution, f.resolveErr
}

func (f *fakeClient) ValidateSRVRecord(context.Context, *net.SRV, uint16, uint16, uint16) error {
	f.count("ValidateSRVRecord")
	return f.srvValidateErr
}

func (f *fakeClient) VerifyPubKey(_, _, _, _ string) (*paymail.VerificationResponse, error) {
	f.count("VerifyPubKey")

	if f.verifyErr != nil {
		return nil, f.verifyErr
	}

	return &paymail.VerificationResponse{VerificationPayload: paymail.VerificationPayload{Match: f.verifyMatch}}, nil
}

// count records one call.
func (f *fakeClient) count(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls[name]++
}

// callCount returns how many times name was called.
func (f *fakeClient) callCount(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls[name]
}

// statusError carries an HTTP status for a fake PKI failure.
type statusError struct {
	status int
	msg    string
}

func (e statusError) Error() string { return e.msg }

// statusFor returns the HTTP status a fake PKI error stands for.
func statusFor(err error) int {
	var se statusError
	if errors.As(err, &se) {
		return se.status
	}

	return http.StatusInternalServerError
}

// fakeReleaseSource answers the update check without touching GitHub.
type fakeReleaseSource struct {
	rel *selfupdate.Release
	err error
}

func (f fakeReleaseSource) Latest(context.Context) (*selfupdate.Release, error) {
	return f.rel, f.err
}

// harness is an app wired to fakes, with captured output.
type harness struct {
	*app

	client *fakeClient
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	dir    string
}

// newHarness returns an isolated app: temp config and cache paths, a fake
// paymail client, a fixed clock, and an update source that never calls out.
func newHarness(t *testing.T) *harness {
	t.Helper()

	dir := t.TempDir()
	client := newFakeClient()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	a := &app{
		name:       appName,
		build:      BuildInfo{Version: testVersion, Commit: "abc1234", Date: "2026-01-02T03:04:05Z"},
		stdout:     stdout,
		stderr:     stderr,
		cfg:        newConfig(),
		newClient:  func(clientOptions) (paymailClient, error) { return client, nil },
		configDirs: []string{filepath.Join(dir, "config"), filepath.Join(dir, "legacy")},
		cachePath:  filepath.Join(dir, "cache", cacheName),
		providers: []provider{
			{Domain: testDomain, Name: "Example", Link: "https://example.com"},
			{Domain: "other.test", Name: "Other", Link: "https://other.test"},
		},
		getenv:    func(key string) string { return map[string]string{noUpdateEnv: "1"}[key] },
		randomHex: func(n int) (string, error) { return fmt.Sprintf("%0*d", n*2, 7), nil },
		now:       func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
	}
	a.update = a.defaultUpdateConfig()
	a.update.Source = fakeReleaseSource{err: errFake}
	a.setColor(false)

	client.caps["other.test"] = capabilities(map[string]any{paymail.BRFCPki: pkiTemplate})

	return &harness{app: a, client: client, stdout: stdout, stderr: stderr, dir: dir}
}

// run executes args and returns the exit code. Every run gets a fresh
// config registry, as a real process would.
func (h *harness) run(args ...string) int {
	h.cfg = newConfig()
	h.flags = globalFlags{}
	h.storeOnce = sync.Once{}
	h.store = nil

	return h.app.run(context.Background(), args)
}

// output returns everything written to stdout so far.
func (h *harness) output() string { return h.stdout.String() }

// errOutput returns everything written to stderr so far.
func (h *harness) errOutput() string { return h.stderr.String() }

// reset clears captured output between runs.
func (h *harness) reset() {
	h.stdout.Reset()
	h.stderr.Reset()
}

// writeConfig writes a config file in the primary config directory.
func (h *harness) writeConfig(t *testing.T, dir, body string) string {
	t.Helper()

	path := filepath.Join(dir, configName)
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	return path
}
