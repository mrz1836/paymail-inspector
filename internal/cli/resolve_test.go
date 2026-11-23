package cli

import (
	"testing"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_Success(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress, "--amount", "500", "--purpose", "coffee"))

	out := h.output()
	assert.Contains(t, out, "--sender-handle is not set; using the receiver's address "+testAddress)
	assert.Contains(t, out, "Found pubkey 02ead23149...")
	assert.Contains(t, out, "Resolved output script 76a914da31...")
	assert.Contains(t, out, "Found public profile [name, avatar]")
	assert.Contains(t, out, "Results for "+testAddress)
	assert.Contains(t, out, "Name         : Alice")
	assert.Contains(t, out, "Avatar       : https://example.com/alice.png")
	assert.Contains(t, out, "PubKey       : "+testPubKey)
	assert.Contains(t, out, "Output Script: "+testOutput)
	assert.Contains(t, out, "Address      : "+testBTCAddr)

	sender := h.client.lastSender
	require.NotNil(t, sender)
	assert.Equal(t, testAddress, sender.SenderHandle)
	assert.Equal(t, uint64(500), sender.Amount)
	assert.Equal(t, "coffee", sender.Purpose)
	assert.Equal(t, "2026-01-02T03:04:05Z", sender.Dt)
	assert.Empty(t, sender.Signature)
}

func TestResolve_HandleShortcut(t *testing.T) {
	h := newHarness(t)
	h.client.srv["relayx.io"] = h.client.srv[testDomain]
	h.client.pki["mrz@relayx.io"] = pkiFor("mrz@relayx.io", testPubKey)

	require.Equal(t, ExitOK, h.run("r", "1mrz", "--skip-public-profile"))
	assert.Contains(t, h.output(), "Results for mrz@relayx.io")
	assert.Zero(t, h.client.callCount("GetPublicProfile"), "--skip-public-profile")
}

func TestResolve_DeprecatedFlags(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress, "--skip-bitpic", "--skip-pki", "--skip-roundesk", "--skip-baemail", "--skip-powping"))
	assert.Contains(t, h.errOutput(), "Flag --skip-bitpic is retired: the service it queried has shut down")
	assert.Contains(t, h.errOutput(), "Flag --skip-pki is retired: the pki lookup is required")
	assert.NotContains(t, h.output(), "Flag --", "notices never pollute stdout")
}

func TestResolve_DeprecatedFlagsAreHidden(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run(cmdResolve, "--help"))
	assert.NotContains(t, h.output(), "skip-bitpic")
	assert.Contains(t, h.output(), "--sender-handle")
}

func TestResolve_EmptyProfileIsNoted(t *testing.T) {
	h := newHarness(t)
	delete(h.client.profiles, testAddress)

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress))
	assert.Contains(t, h.output(), "Public profile is empty")
	assert.NotContains(t, h.output(), "Name         :")
}

func TestResolve_ProfileFailureIsAWarning(t *testing.T) {
	h := newHarness(t)
	h.client.profileErr = errFake

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress))
	assert.Contains(t, h.output(), "Public profile unavailable")
}

func TestResolve_SenderValidation(t *testing.T) {
	h := newHarness(t)
	h.client.caps[testSRVHost].Capabilities[paymail.BRFCSenderValidation] = true
	h.client.pki["bob@other.test"] = pkiFor("bob@other.test", senderPub)

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress, "--sender-handle", "bob@other.test", "--sender-name", "Bob"))

	out := h.output()
	assert.Contains(t, out, "Sender validation is ENFORCED by "+testDomain)
	assert.Contains(t, out, "No --signature given; sending a random placeholder signature")
	assert.Contains(t, out, "Found --sender-handle bob@other.test's pubkey")
	assert.Contains(t, out, "Sender pre-validation passed")

	sender := h.client.lastSender
	require.NotNil(t, sender)
	assert.Equal(t, "bob@other.test", sender.SenderHandle)
	assert.Equal(t, "Bob", sender.SenderName)
	assert.Len(t, sender.Signature, fakeSignatureBytes*2)
}

func TestResolve_SenderValidationWithSignature(t *testing.T) {
	h := newHarness(t)
	h.client.caps[testSRVHost].Capabilities[paymail.BRFCSenderValidation] = true

	require.Equal(t, ExitOK, h.run(cmdResolve, testAddress, "--signature", "sig"))
	assert.Equal(t, "sig", h.client.lastSender.Signature)
	assert.NotContains(t, h.output(), "placeholder")
}

func TestResolve_SenderValidationFailures(t *testing.T) {
	tests := []struct {
		name  string
		setup func(h *harness)
		want  string
	}{
		{"random failure", func(h *harness) {
			h.randomHex = func(int) (string, error) { return "", errFake }
		}, errFake.Error()},
		{"sender capabilities", func(h *harness) {
			delete(h.client.caps, "other.test")
		}, "sender bob@other.test: get capabilities"},
		{"sender without pki", func(h *harness) {
			h.client.caps["other.test"] = capabilities(map[string]any{})
		}, "other.test does not publish pki"},
		{"sender pki", func(h *harness) {
			h.client.pkiErr["bob@other.test"] = errFake
		}, "sender bob@other.test: get pki"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			h.client.caps[testSRVHost].Capabilities[paymail.BRFCSenderValidation] = true
			tt.setup(h)

			assert.Equal(t, ExitError, h.run(cmdResolve, testAddress, "--sender-handle", "bob@other.test"))
			assert.Contains(t, h.errOutput(), tt.want)
		})
	}
}

func TestResolve_Failures(t *testing.T) {
	tests := []struct {
		name  string
		setup func(h *harness)
		args  []string
		want  string
	}{
		{"invalid address", nil, []string{cmdResolve, "nope"}, "invalid paymail address"},
		{"invalid sender", nil, []string{cmdResolve, testAddress, "--sender-handle", "nope"}, "--sender-handle: invalid paymail address"},
		{cmdCaps, func(h *harness) { h.client.capsErr = errFake }, []string{cmdResolve, testAddress}, "get capabilities"},
		{"missing pki", func(h *harness) {
			delete(h.client.caps[testSRVHost].Capabilities, paymail.BRFCPki)
		}, []string{cmdResolve, testAddress}, "does not publish pki"},
		{"missing resolution", func(h *harness) {
			delete(h.client.caps[testSRVHost].Capabilities, paymail.BRFCPaymentDestination)
		}, []string{cmdResolve, testAddress}, "does not publish paymentDestination"},
		{"pki", func(h *harness) { h.client.pkiErr[testAddress] = errFake }, []string{cmdResolve, testAddress}, "get pki for " + testAddress},
		{"resolution", func(h *harness) { h.client.resolveErr = errFake }, []string{cmdResolve, testAddress}, "resolve " + testAddress},
		{"client", func(h *harness) {
			h.newClient = func(clientOptions) (paymailClient, error) { return nil, errFake }
		}, []string{cmdResolve, testAddress}, errFake.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			if tt.setup != nil {
				tt.setup(h)
			}

			assert.Equal(t, ExitError, h.run(tt.args...))
			assert.Contains(t, h.errOutput(), tt.want)
		})
	}
}
