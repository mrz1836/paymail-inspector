package cli

import (
	"strings"
	"testing"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestP2P_Success(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("p2p", testAddress))

	out := h.output()
	assert.Contains(t, out, "Found 1 payment output(s)")
	assert.Contains(t, out, "P2P information for "+testAddress)
	assert.Contains(t, out, "Name         : Alice")
	assert.Contains(t, out, "Reference    : ref-123")
	assert.Contains(t, out, "Output #1")
	assert.Contains(t, out, "Script       : "+testOutput)
	assert.Contains(t, out, "Satoshis     : 1000")
	assert.Contains(t, out, "Address      : "+testBTCAddr)

	require.NotNil(t, h.client.lastPayment)
	assert.Equal(t, uint64(defaultSatoshis), h.client.lastPayment.Satoshis)
}

func TestP2P_CustomAmountWithoutProfile(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("send", testAddress, "--satoshis", "5000", "--skip-public-profile", "--skip-bitpic"))
	assert.Equal(t, uint64(5000), h.client.lastPayment.Satoshis)
	assert.Zero(t, h.client.callCount("GetPublicProfile"))
	assert.NotContains(t, h.output(), "Name         :")
}

func TestP2P_Failures(t *testing.T) {
	tests := []struct {
		name  string
		setup func(h *harness)
		args  []string
		want  string
	}{
		{"invalid address", nil, []string{"p2p", "nope"}, "invalid paymail address"},
		{cmdCaps, func(h *harness) { h.client.capsErr = errFake }, []string{"p2p", testAddress}, "get capabilities"},
		{"missing capability", func(h *harness) {
			delete(h.client.caps[testSRVHost].Capabilities, paymail.BRFCP2PPaymentDestination)
		}, []string{"p2p", testAddress}, "does not publish " + paymail.BRFCP2PPaymentDestination},
		{"provider", func(h *harness) { h.client.p2pErr = errFake }, []string{"p2p", testAddress}, "p2p payment destination for " + testAddress},
		{"client", func(h *harness) {
			h.newClient = func(clientOptions) (paymailClient, error) { return nil, errFake }
		}, []string{"p2p", testAddress}, errFake.Error()},
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

func TestVerify_Match(t *testing.T) {
	for _, args := range [][]string{
		{cmdVerify, testAddress, testPubKey},
		{"verification", testPubKey, testAddress},
	} {
		h := newHarness(t)

		require.Equal(t, ExitOK, h.run(args...), "args %v", args)
		assert.Contains(t, h.output(), "Verify response for "+testAddress)
		assert.Contains(t, h.output(), "PubKey       : "+testPubKey)
		assert.Contains(t, h.output(), "Paymail & PubKey match!")
	}
}

func TestVerify_Mismatch(t *testing.T) {
	h := newHarness(t)
	h.client.verifyMatch = false

	assert.Equal(t, ExitError, h.run(cmdVerify, testAddress, testPubKey))
	assert.Contains(t, h.output(), "DO NOT MATCH!")
	assert.Contains(t, h.errOutput(), "pubkey does not match the paymail address: "+testAddress)
}

func TestVerify_Failures(t *testing.T) {
	tests := []struct {
		name  string
		setup func(h *harness)
		args  []string
		code  int
		want  string
	}{
		{"one argument", nil, []string{cmdVerify, testAddress}, ExitUsage, "verify requires a paymail address and a pubkey"},
		{"short pubkey", nil, []string{cmdVerify, testAddress, "02abc"}, ExitError, "invalid pubkey: expected 66 characters but got 5"},
		{"long pubkey", nil, []string{cmdVerify, testAddress, testPubKey + "00"}, ExitError, "invalid pubkey: expected 66 characters but got 68"},
		{"invalid address", nil, []string{cmdVerify, "not@valid", testPubKey}, ExitError, "invalid paymail address"},
		{cmdCaps, func(h *harness) { h.client.capsErr = errFake }, []string{cmdVerify, testAddress, testPubKey}, ExitError, "get capabilities"},
		{"missing capability", func(h *harness) {
			delete(h.client.caps[testSRVHost].Capabilities, paymail.BRFCVerifyPublicKeyOwner)
		}, []string{cmdVerify, testAddress, testPubKey}, ExitError, "does not publish " + paymail.BRFCVerifyPublicKeyOwner},
		{"provider", func(h *harness) { h.client.verifyErr = errFake }, []string{cmdVerify, testAddress, testPubKey}, ExitError, "verify pubkey for " + testAddress},
		{"client", func(h *harness) {
			h.newClient = func(clientOptions) (paymailClient, error) { return nil, errFake }
		}, []string{cmdVerify, testAddress, testPubKey}, ExitError, errFake.Error()},
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

func TestLooksLikeAddress(t *testing.T) {
	assert.True(t, looksLikeAddress(testAddress))
	assert.True(t, looksLikeAddress("$handle"))
	assert.True(t, looksLikeAddress("1handle"))
	assert.False(t, looksLikeAddress(testPubKey))
	assert.False(t, looksLikeAddress(strings.Repeat("a", 70)))
}
