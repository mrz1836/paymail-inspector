package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/spf13/cobra"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

// newVerifyCmd returns `verify <paymail> <pubkey>`.
func newVerifyCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "verify <paymail> <pubkey>",
		Short: "Verify that a pubkey belongs to a paymail address",
		Long: `
                   .__  _____
___  __ ___________|__|/ ____\__.__.
\  \/ // __ \_  __ \  \   __<   |  |
 \   /\  ___/|  | \/  ||  |  \___  |
  \_/  \___  >__|  |__||__|  / ____|
           \/                \/

Ask the provider of a paymail address whether a public key is a valid identity
key for it (when the provider supports the capability). The arguments may be
given in either order.

The key returned by pki for a handle can change over time, which makes data
signed with an old key hard to verify: holding the key is not proof that it
belonged to the user. This capability answers exactly that question.

The command exits non-zero when the provider reports no match.

Read more at: ` + specMirror + "05-verify-public-key-owner.html",
		Aliases:    []string{"verification"},
		SuggestFor: []string{"pubkey"},
		Example: a.examples(
			"verify "+exampleVerifyAddress+" "+examplePubKey,
			"verify "+examplePubKey+" "+exampleVerifyAddress,
		),
		Args: exactArgs(2, "a paymail address and a pubkey"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runVerify(cmd.Context(), args[0], args[1])
		},
	}
}

// runVerify checks one address/pubkey pair.
func (a *app) runVerify(ctx context.Context, first, second string) error {
	addr, pubKey, err := splitVerifyArgs(first, second)
	if err != nil {
		return err
	}

	s, err := a.newSession("")
	if err != nil {
		return err
	}

	caps, err := s.capabilities(ctx, addr.Domain, true)
	if err != nil {
		return err
	}

	verifyURL, err := requireCapability(caps, addr.Domain, paymail.BRFCVerifyPublicKeyOwner, "")
	if err != nil {
		return err
	}

	verification, err := s.verifyPubKey(verifyURL, addr, pubKey)
	if err != nil {
		return err
	}

	a.out.Header(output.LevelBold, "Verify response for "+a.out.Value(addr.Address))
	field(a.out, "Paymail", addr.Address)
	field(a.out, "PubKey", pubKey)

	if !verification.Match {
		a.out.Error("DO NOT MATCH! (service responded: match=false)")
		return fmt.Errorf("%w: %s", errPubKeyMismatch, addr.Address)
	}

	a.out.Success("Paymail & PubKey match! (service responded: match=true)")

	return nil
}

// splitVerifyArgs accepts the address and pubkey in either order. A short
// argument without "@" is read as a provider handle ($handle, 1handle).
func splitVerifyArgs(first, second string) (address, string, error) {
	rawAddr, pubKey := first, second
	if !looksLikeAddress(first) && looksLikeAddress(second) {
		rawAddr, pubKey = second, first
	}

	addr, err := parseAddress(rawAddr)
	if err != nil {
		return address{}, "", err
	}

	pubKey = strings.TrimSpace(pubKey)
	if len(pubKey) != paymail.PubKeyLength {
		return address{}, "", fmt.Errorf("%w: expected %d characters but got %d", errInvalidPubKey, paymail.PubKeyLength, len(pubKey))
	}

	return addr, pubKey, nil
}

// looksLikeAddress reports whether s is a paymail address or handle rather
// than a pubkey.
func looksLikeAddress(s string) bool {
	return strings.Contains(s, "@") || strings.HasPrefix(s, "$") || len(s) < paymail.PubKeyLength
}
