package cli

import (
	"net/http"
	"testing"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_DomainPasses(t *testing.T) {
	h := newHarness(t)
	h.client.dnssec.Answer.DSRecordCount = 2

	require.Equal(t, ExitOK, h.run("validate", testDomain))

	out := h.output()
	assert.Contains(t, out, "Domain detected: "+testDomain)
	assert.Contains(t, out, "SRV record passed all validations")
	assert.Contains(t, out, "Checking "+testSRVHost+" for DNSSEC")
	assert.Contains(t, out, "DNSSEC is valid (2 DS record(s))")
	assert.Contains(t, out, "SSL certificate is valid for "+testSRVHost)
	assert.Contains(t, out, "Found required capabilities: [pki] [paymentDestination]")
	assert.Contains(t, out, testDomain+" passed validation (0 warning(s))")
	assert.Equal(t, 1, h.client.callCount("ValidateSRVRecord"))
	assert.Equal(t, 1, h.client.callCount("GetSRVRecord"), "the checked SRV record is reused for capabilities")
}

func TestValidate_AddressPasses(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("v", testAddress))
	assert.Contains(t, h.output(), "Paymail address detected: "+testAddress)
	assert.Contains(t, h.output(), "Paymail information for "+testAddress)
	assert.Contains(t, h.output(), "PubKey       : "+testPubKey)
}

func TestValidate_SkipsAreWarnings(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("validate", testDomain, "--skip-srv", "--skip-dnssec", "--skip-ssl"))

	out := h.output()
	assert.Contains(t, out, "Skipping SRV record check for "+testDomain)
	assert.Contains(t, out, "Skipping DNSSEC check for "+testDomain)
	assert.Contains(t, out, "Skipping SSL check for "+testDomain)
	assert.Contains(t, out, "passed validation (3 warning(s))")
	assert.Zero(t, h.client.callCount("ValidateSRVRecord"))
	assert.Zero(t, h.client.callCount("CheckDNSSEC"))
	assert.Zero(t, h.client.callCount("CheckSSL"))
}

func TestValidate_SRVExpectationsAreConfigurable(t *testing.T) {
	h := newHarness(t)

	var got clientOptions

	h.newClient = func(opts clientOptions) (paymailClient, error) {
		got = opts
		return h.client, nil
	}

	require.Equal(t, ExitOK, h.run("validate", testDomain, "-n", "1.1.1.1", "-p", "8443", "--priority", "5", "-w", "7", "-s", "custom", "--protocol", "udp"))
	assert.Equal(t, "1.1.1.1", got.nameServer)
	assert.Equal(t, "paymail-inspector/"+testVersion, got.userAgent)
	assert.True(t, got.tracing)
}

func TestValidate_DNSSECWarnings(t *testing.T) {
	tests := []struct {
		name   string
		result *paymail.DNSCheckResult
		want   string
	}{
		{"missing", &paymail.DNSCheckResult{Domain: testSRVHost, ErrorMessage: "no DS records"}, "DNSSEC detail: no DS records"},
		{"no result", nil, "DNSSEC check returned no result"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			h.client.dnssec = tt.result

			require.Equal(t, ExitOK, h.run("validate", testDomain))
			assert.Contains(t, h.output(), tt.want)
			assert.Contains(t, h.output(), "passed validation (1 warning(s))")
		})
	}
}

func TestValidate_P2POnlyProviderIsAWarning(t *testing.T) {
	h := newHarness(t)
	delete(h.client.caps[testSRVHost].Capabilities, paymail.BRFCPaymentDestination)

	require.Equal(t, ExitOK, h.run("validate", testDomain))
	assert.Contains(t, h.output(), "the provider offers P2P payment destinations")
	assert.NotContains(t, h.output(), "Found required capabilities")
}

func TestValidate_Failures(t *testing.T) {
	tests := []struct {
		name  string
		setup func(h *harness)
		args  []string
		want  string
	}{
		{"srv lookup", func(h *harness) { h.client.srvErr = errFake }, []string{"validate", testDomain}, "lookup SRV record for " + testDomain},
		{"srv invalid", func(h *harness) { h.client.srvValidateErr = errFake }, []string{"validate", testDomain}, "SRV record for " + testDomain + " is invalid"},
		{"ssl error", func(h *harness) { h.client.sslErr = errFake }, []string{"validate", testDomain}, "SSL check failed for " + testSRVHost},
		{"ssl invalid", func(h *harness) { h.client.sslValid = false }, []string{"validate", testDomain}, "No valid SSL certificate found"},
		{cmdCaps, func(h *harness) { h.client.capsErr = errFake }, []string{"validate", testDomain}, "get capabilities for " + testDomain},
		{"no capabilities", func(h *harness) {
			h.client.caps[testSRVHost] = capabilities(map[string]any{})
		}, []string{"validate", testDomain}, "Missing required capability: pki; Missing required capability: paymentDestination"},
		{"pki", func(h *harness) { h.client.pkiErr[testAddress] = errFake }, []string{"validate", testAddress}, "get pki for " + testAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			tt.setup(h)

			assert.Equal(t, ExitError, h.run(tt.args...))
			assert.Contains(t, h.errOutput(), "validation failed: "+testDomain)
			assert.Contains(t, h.errOutput(), tt.want)
			assert.Contains(t, h.output(), "failed")
		})
	}
}

func TestValidate_InputErrors(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, ExitError, h.run("validate", "nodot"))
	assert.Contains(t, h.errOutput(), "invalid domain")

	h.reset()
	assert.Equal(t, ExitError, h.run("validate", "bad@"))
	assert.Contains(t, h.errOutput(), "invalid paymail address")

	h.reset()
	h.newClient = func(clientOptions) (paymailClient, error) { return nil, errFake }
	assert.Equal(t, ExitError, h.run("validate", testDomain))
	assert.Contains(t, h.errOutput(), errFake.Error())
}

func TestWhois_Results(t *testing.T) {
	h := newHarness(t)
	h.client.pkiErr["alice@other.test"] = statusError{status: http.StatusBadRequest, msg: "code 400, message: Paymail not found"}

	require.Equal(t, ExitOK, h.run("whois", testAlias))

	out := h.output()
	assert.Contains(t, out, "Searching 2 providers for "+testAlias)
	assert.Contains(t, out, "Results for "+testAddress)
	assert.Contains(t, out, "Name         : Alice")
	assert.Contains(t, out, "The handle alice might be available! Reserve it now: https://other.test")
	assert.Less(t, indexOf(out, testAddress), indexOf(out, "alice@other.test"), "results follow provider order")
	assert.NotContains(t, out, "Retrieving", "whois suppresses per-request progress")
}

func TestWhois_HandleForms(t *testing.T) {
	for _, raw := range []string{"alice", "Alice", "alice@anything.io", "$alice"} {
		h := newHarness(t)

		require.Equal(t, ExitOK, h.run("w", raw, "--skip-public-profile"), raw)
		assert.Contains(t, h.output(), "Results for "+testAddress, raw)
		assert.Zero(t, h.client.callCount("GetPublicProfile"), raw)
	}
}

func TestWhois_PartialFailure(t *testing.T) {
	h := newHarness(t)
	h.client.pkiErr["alice@other.test"] = errFake

	require.Equal(t, ExitOK, h.run("whois", testAlias))
	assert.Contains(t, h.output(), "Lookup failed: get pki for alice@other.test")
}

func TestWhois_ProviderWithoutPKI(t *testing.T) {
	h := newHarness(t)
	h.client.caps["other.test"] = capabilities(map[string]any{})

	require.Equal(t, ExitOK, h.run("whois", testAlias))
	assert.Contains(t, h.output(), "other.test does not publish pki")
}

func TestWhois_AllProvidersFail(t *testing.T) {
	h := newHarness(t)
	h.client.capsErr = errFake

	assert.Equal(t, ExitError, h.run("whois", testAlias))
	assert.Contains(t, h.errOutput(), "no paymail provider responded for "+testAlias)
}

func TestWhois_InputErrors(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, ExitError, h.run("whois", "!!!"))
	assert.Contains(t, h.errOutput(), "invalid handle")

	h.reset()
	h.newClient = func(clientOptions) (paymailClient, error) { return nil, errFake }
	assert.Equal(t, ExitError, h.run("whois", testAlias))
}

func TestIsNotFound(t *testing.T) {
	status := func(code int) *paymail.PKIResponse {
		return &paymail.PKIResponse{StandardResponse: paymail.StandardResponse{StatusCode: code}}
	}

	assert.False(t, isNotFound(status(http.StatusNotFound), nil))
	assert.False(t, isNotFound(nil, errFake))
	assert.True(t, isNotFound(status(http.StatusNotFound), errFake))
	assert.True(t, isNotFound(status(http.StatusBadRequest), statusError{msg: "Paymail NOT FOUND"}))
	assert.False(t, isNotFound(status(http.StatusBadRequest), errFake))
	assert.False(t, isNotFound(status(http.StatusInternalServerError), errFake))
}
