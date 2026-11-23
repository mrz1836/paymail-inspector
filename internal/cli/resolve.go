package cli

import (
	"context"
	"fmt"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/spf13/cobra"
)

// fakeSignatureBytes is the size of the placeholder signature sent when a
// provider enforces sender validation and no --signature is given.
const fakeSignatureBytes = 64

// resolveOptions are the flags of the resolve command.
type resolveOptions struct {
	amount            uint64
	purpose           string
	signature         string
	skipPublicProfile bool
}

// newResolveCmd returns `resolve <paymail>`.
func newResolveCmd(a *app) *cobra.Command {
	opts := &resolveOptions{}

	cmd := &cobra.Command{
		Use:   "resolve <paymail>",
		Short: "Resolve a paymail address to its pubkey, output script, and profile",
		Long: `
                            .__
_______   ____   __________ |  |___  __ ____
\_  __ \_/ __ \ /  ___/  _ \|  |\  \/ // __ \
 |  | \/\  ___/ \___ (  <_> )  |_\   /\  ___/
 |__|    \___  >____  >____/|____/\_/  \___  >
             \/     \/                     \/

Resolve a paymail address into its public key, a hex-encoded Bitcoin output
script and address, and its public profile (when published).

Given a sender and a receiver, where the sender knows the receiver's paymail
handle <alias>@<domain>.<tld>, the sender performs service discovery against
the receiver and requests a payment destination from the receiver's paymail
service.

Read more at: ` + specMirror + "04-01-basic-address-resolution.html",
		Aliases:    []string{"r", "resolution"},
		SuggestFor: []string{"address", "destination", "payment", "addressing"},
		Example: a.examples(
			"resolve "+exampleAddress,
			"r "+exampleAddress+" --sender-handle you@"+exampleDomain,
			"r 1"+exampleAlias,
		),
		Args: exactArgs(1, "a paymail address"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runResolve(cmd.Context(), args[0], opts)
		},
	}

	flags := cmd.Flags()
	flags.Uint64VarP(&opts.amount, "amount", "a", 0, "Amount in satoshis for the payment request")
	flags.StringVarP(&opts.purpose, "purpose", "p", "", "Purpose of the payment")
	flags.StringVarP(&opts.signature, "signature", "s", "", "Signature of the sender request (when the provider enforces sender validation)")
	flags.BoolVar(&opts.skipPublicProfile, "skip-public-profile", false, "Skip the public profile request")
	flags.String(keySenderHandle, "", "Sender's paymail address (default: the receiver's address)")
	flags.String(keySenderName, "", "Sender's name")
	_ = a.cfg.BindPFlag(keySenderHandle, flags.Lookup(keySenderHandle))
	_ = a.cfg.BindPFlag(keySenderName, flags.Lookup(keySenderName))

	deprecateFlags(cmd, reasonAlwaysRequired, "skip-pki")
	deprecateFlags(cmd, reasonServiceShutDown, retiredIntegrationFlags()...)

	return cmd
}

// runResolve resolves one paymail address and prints the results.
func (a *app) runResolve(ctx context.Context, raw string, opts *resolveOptions) error {
	addr, err := parseAddress(raw)
	if err != nil {
		return err
	}

	sender, err := a.senderAddress(addr)
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

	pkiURL, resolveURL, err := resolveEndpoints(caps, addr.Domain)
	if err != nil {
		return err
	}

	signature := opts.signature
	if capBool(caps, paymail.BRFCSenderValidation) {
		if signature, err = s.senderValidation(ctx, addr, sender, signature); err != nil {
			return err
		}
	}

	details := &paymailDetails{Address: addr, Provider: a.providerFor(addr.Domain)}

	if details.PKI, err = s.pki(pkiURL, addr, true); err != nil {
		return err
	}

	details.Resolution, err = s.resolveAddress(resolveURL, addr, &paymail.SenderRequest{
		Amount:       opts.amount,
		Purpose:      opts.purpose,
		SenderHandle: sender.Address,
		SenderName:   a.cfg.GetString(keySenderName),
		Signature:    signature,
	})
	if err != nil {
		return err
	}

	details.PublicProfile = s.optionalProfile(caps, addr, opts.skipPublicProfile)
	details.render(a.out)

	return nil
}

// resolveEndpoints returns the pki and address-resolution URLs a provider
// must publish for resolve to work.
func resolveEndpoints(caps *paymail.CapabilitiesResponse, domain string) (pkiURL, resolveURL string, err error) {
	if pkiURL, err = requireCapability(caps, domain, paymail.BRFCPki, paymail.BRFCPkiAlternate); err != nil {
		return "", "", err
	}

	if resolveURL, err = requireCapability(caps, domain, paymail.BRFCPaymentDestination, paymail.BRFCBasicAddressResolution); err != nil {
		return "", "", err
	}

	return pkiURL, resolveURL, nil
}

// senderAddress returns the configured sender, defaulting to the receiver.
func (a *app) senderAddress(receiver address) (address, error) {
	raw := a.cfg.GetString(keySenderHandle)
	if raw == "" {
		a.out.Dim("--%s is not set; using the receiver's address %s", keySenderHandle, receiver.Address)
		return receiver, nil
	}

	sender, err := parseAddress(raw)
	if err != nil {
		return address{}, fmt.Errorf("--%s: %w", keySenderHandle, err)
	}

	return sender, nil
}

// senderValidation handles a provider that enforces sender validation: it
// confirms the sender's pubkey can be found and returns the signature to
// send. Without --signature a random placeholder is sent, which lets the
// rest of the flow be inspected but will be rejected by a strict provider.
func (s *session) senderValidation(ctx context.Context, receiver, sender address, signature string) (string, error) {
	s.out.Warn("Sender validation is ENFORCED by %s", receiver.Domain)
	s.step("Running sender validation for %s...", s.out.Value(sender.Address))

	if signature == "" {
		s.out.Warn("No --signature given; sending a random placeholder signature")

		var err error
		if signature, err = s.randomHex(fakeSignatureBytes); err != nil {
			return "", err
		}
	}

	if sender.Address != receiver.Address {
		caps, err := s.capabilities(ctx, sender.Domain, true)
		if err != nil {
			return "", fmt.Errorf("sender %s: %w", sender.Address, err)
		}

		pkiURL, err := requireCapability(caps, sender.Domain, paymail.BRFCPki, paymail.BRFCPkiAlternate)
		if err != nil {
			return "", err
		}

		pki, err := s.pki(pkiURL, sender, true)
		if err != nil {
			return "", fmt.Errorf("sender %s: %w", sender.Address, err)
		}

		s.out.Info("Found --%s %s's pubkey: %s", keySenderHandle, sender.Address, s.out.Value(pki.PubKey))
	}

	s.out.Success("Sender pre-validation passed")

	return signature, nil
}
