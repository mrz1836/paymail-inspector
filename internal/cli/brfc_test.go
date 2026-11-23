package cli

import (
	"testing"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrfc_GroupShowsHelp(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("brfc"))
	assert.Contains(t, h.output(), "list      show every known BRFC specification")
	assert.Contains(t, h.output(), "Available Commands:")
}

func TestBrfc_List(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("brfc", "list"))

	out := h.output()
	assert.Contains(t, out, "bsvalias Public Key Infrastructure v1")
	assert.Contains(t, out, "URL       : "+specMirror+"03-public-key-infrastructure.html", "lapsed bsvalias.org links are rewritten")
	assert.NotContains(t, out, "bsvalias.org")
	assert.Contains(t, out, "Alias     : pki")
	assert.Contains(t, out, "c9ca939e2f69 (valid)")
	assert.Contains(t, out, "deadbeef0000 (invalid ID, expected")
	assert.NotContains(t, out, "No ID", "malformed entries are skipped")
	assert.Contains(t, out, "Total BRFC specifications found: 3")
}

func TestBrfc_ListSkipValidation(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("brfc", "ls", "--skip-validation"))
	assert.Contains(t, h.output(), "deadbeef0000 (valid)", "validation is skipped")
}

func TestBrfc_SearchValidationError(t *testing.T) {
	h := newHarness(t)

	// list skips untitled specs, but search shows any match, and an ID
	// cannot be regenerated without a title.
	h.client.brfcs = []*paymail.BRFCSpec{{ID: "abc123", Author: "someone", Version: "1"}}

	require.Equal(t, ExitOK, h.run("brfc", "search", "abc123"))
	assert.Contains(t, h.output(), "(validation error:")
}

func TestBrfc_ListWithoutSpecs(t *testing.T) {
	h := newHarness(t)
	h.client.brfcs = nil

	assert.Equal(t, ExitError, h.run("brfc", "list"))
	assert.Contains(t, h.errOutput(), "no BRFC specifications available")
}

func TestBrfc_ClientFailure(t *testing.T) {
	h := newHarness(t)
	h.newClient = func(clientOptions) (paymailClient, error) { return nil, errFake }

	assert.Equal(t, ExitError, h.run("brfc", "list"))
	assert.Contains(t, h.errOutput(), errFake.Error())
}

func TestBrfc_Search(t *testing.T) {
	tests := []struct {
		name  string
		term  string
		found string
		total string
	}{
		{"by author", "NCHAIN", "Basic Address Resolution", "found: 2 searching: NCHAIN"},
		{"by id", "ba4ba8", "Public Key Infrastructure", "found: 1 searching: ba4ba8"},
		{"by title", "bogus", "deadbeef0000", "found: 1 searching: bogus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)

			require.Equal(t, ExitOK, h.run("brfc", "search", tt.term))
			assert.Contains(t, h.output(), tt.found)
			assert.Contains(t, h.output(), tt.total)
		})
	}
}

func TestBrfc_SearchFailures(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, ExitError, h.run("brfc", "find", "no-such-spec"))
	assert.Contains(t, h.errOutput(), `no BRFC specifications matched: "no-such-spec"`)

	h.reset()
	assert.Equal(t, ExitUsage, h.run("brfc", "search"))
	assert.Contains(t, h.errOutput(), "search requires a search term")

	h.reset()
	h.client.brfcs = nil
	assert.Equal(t, ExitError, h.run("brfc", "search", "x"))
}

func TestBrfc_Generate(t *testing.T) {
	h := newHarness(t)

	require.Equal(t, ExitOK, h.run("brfc", "generate", "--title", "BRFC Specifications", "--author", "andy (nChain)", "--version", "1"))

	out := h.output()
	assert.Contains(t, out, "Generated BRFC ID")
	assert.Contains(t, out, "ID           : 57dd1f54fc67")
	assert.Contains(t, out, "Title        : BRFC Specifications")
	assert.Contains(t, out, "Author       : andy (nChain)")
	assert.Contains(t, out, "Version      : 1")
}

func TestBrfc_GenerateFailures(t *testing.T) {
	h := newHarness(t)

	assert.Equal(t, ExitUsage, h.run("brfc", "gen"))
	assert.Contains(t, h.errOutput(), "brfc generate requires --title")

	h.reset()
	assert.Equal(t, ExitError, h.run("brfc", "new", "--title", "bsvalias Public Key Infrastructure", "--author", "andy (nChain)", "--version", "1"))
	assert.Contains(t, h.errOutput(), "BRFC already exists: ba4ba843a7d6")

	h.reset()
	h.client.brfcs = nil
	assert.Equal(t, ExitError, h.run("brfc", "generate", "--title", "Anything"))
	assert.Contains(t, h.errOutput(), "no BRFC specifications available")
}

func TestContainsFold(t *testing.T) {
	assert.True(t, containsFold("Hello World", "WORLD"))
	assert.True(t, containsFold("abc", ""))
	assert.False(t, containsFold("abc", "d"))
}
