package cli

import (
	"net"
	"testing"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities_ListsSortedWithTitles(t *testing.T) {
	h := newHarness(t)
	h.client.caps[testSRVHost].Capabilities["custom"] = 42

	require.Equal(t, ExitOK, h.run(cmdCaps, testDomain))

	out := h.output()
	assert.Contains(t, out, "Retrieving SRV record for "+testDomain)
	assert.Contains(t, out, "SRV target: "+testSRVHost+":443 --weight 10 --priority 10")
	assert.Contains(t, out, "Retrieving capabilities for "+testSRVHost+":443")
	assert.Contains(t, out, "Found 7 capabilities")
	assert.Contains(t, out, "Listing 7 capabilities (bsvalias 1.0)")
	assert.Contains(t, out, "Disabled")
	assert.Contains(t, out, pkiTemplate+"  (bsvalias Public Key Infrastructure)")
	assert.Contains(t, out, "custom")
	assert.Contains(t, out, "42")

	// Sorted: BRFC IDs (digits) before named capabilities.
	assert.Less(t, indexOf(out, paymail.BRFCP2PPaymentDestination), indexOf(out, "paymentDestination   "))
	assert.Less(t, indexOf(out, "paymentDestination   "), indexOf(out, "pki   "))
}

func TestCapabilities_EnabledBoolean(t *testing.T) {
	h := newHarness(t)
	h.client.caps[testSRVHost].Capabilities[paymail.BRFCSenderValidation] = true

	require.Equal(t, ExitOK, h.run("c", testAddress, "-t"))
	assert.Contains(t, h.output(), "Enabled")
	assert.NotContains(t, h.output(), "DNSLookup", "-t hides tracing")
}

func TestCapabilities_Tracing(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run(cmdCaps, testDomain))
	assert.Contains(t, h.output(), "DNSLookup")
	assert.Contains(t, h.output(), "TotalTime")
}

func TestCapabilities_SRVFallback(t *testing.T) {
	h := newHarness(t)
	h.client.srvErr = errFake
	h.client.caps[testDomain] = h.client.caps[testSRVHost]

	require.Equal(t, ExitOK, h.run(cmdCaps, testDomain))
	assert.Contains(t, h.output(), "No usable SRV record")
	assert.Contains(t, h.output(), "using "+testDomain+":443")
}

func TestCapabilities_NilSRVFallsBack(t *testing.T) {
	h := newHarness(t)
	h.client.srv[testDomain] = nil
	h.client.caps[testDomain] = h.client.caps[testSRVHost]

	require.Equal(t, ExitOK, h.run(cmdCaps, testDomain))
	assert.Contains(t, h.output(), "No usable SRV record")
}

func TestCapabilities_Failures(t *testing.T) {
	tests := []struct {
		name  string
		setup func(h *harness)
		args  []string
		code  int
		want  string
	}{
		{"invalid domain", nil, []string{cmdCaps, "nodot"}, ExitError, "invalid domain"},
		{"empty domain", nil, []string{cmdCaps, "!!!"}, ExitError, "invalid domain"},
		{"invalid address", nil, []string{cmdCaps, "@" + testDomain}, ExitError, "invalid paymail address"},
		{"provider error", func(h *harness) { h.client.capsErr = errFake }, []string{cmdCaps, testDomain}, ExitError, "get capabilities for " + testDomain},
		{"version mismatch", nil, []string{cmdCaps, testDomain, "--bsvalias", "2.0"}, ExitError, `expected "2.0" but got "1.0"`},
		{"client error", func(h *harness) {
			h.newClient = func(clientOptions) (paymailClient, error) { return nil, errFake }
		}, []string{cmdCaps, testDomain}, ExitError, errFake.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			if tt.setup != nil {
				tt.setup(h)
			}

			assert.Equal(t, tt.code, h.run(tt.args...))
			assert.Contains(t, h.errOutput(), tt.want)
		})
	}
}

func TestCapabilities_CachedVersionIsRechecked(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress))

	h.reset()
	assert.Equal(t, ExitError, h.run(cmdResolve, testAddress, "--bsvalias", "9.9"))
	assert.Contains(t, h.errOutput(), "bsvalias version mismatch")
}

func TestCapHelpers(t *testing.T) {
	caps := capabilities(map[string]any{
		"str":  "value",
		"alt":  "alternate",
		"bool": true,
		"no":   false,
		"num":  1,
	})

	assert.Equal(t, "value", capString(caps, "str", ""))
	assert.Equal(t, "alternate", capString(caps, "missing", "alt"))
	assert.Empty(t, capString(caps, "bool", ""), "a boolean is not a URL")
	assert.Empty(t, capString(caps, "missing", ""))

	assert.True(t, capBool(caps, "bool"))
	assert.False(t, capBool(caps, "no"))
	assert.False(t, capBool(caps, "str"), "a string is not a boolean")
	assert.False(t, capBool(caps, "missing"))

	url, err := requireCapability(caps, testDomain, "str", "")
	require.NoError(t, err)
	assert.Equal(t, "value", url)

	_, err = requireCapability(caps, testDomain, "missing", "")
	require.ErrorIs(t, err, errMissingCapability)
}

func TestBrfcTitle(t *testing.T) {
	specs := []*paymail.BRFCSpec{{ID: "abc", Title: "By ID"}, {ID: "def", Alias: "pki", Title: "By Alias"}}

	assert.Equal(t, "By ID", brfcTitle(specs, "abc"))
	assert.Equal(t, "By Alias", brfcTitle(specs, "pki"))
	assert.Empty(t, brfcTitle(specs, "zzz"))
	assert.Empty(t, brfcTitle(specs, ""), "an empty alias never matches")
}

func TestFormatSRV(t *testing.T) {
	assert.Equal(t, "host:443 --weight 1 --priority 2", formatSRV(&net.SRV{Target: "host", Port: 443, Weight: 1, Priority: 2}))
}

// indexOf returns the byte offset of substr in s, or -1.
func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}
