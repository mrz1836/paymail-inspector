package cli

import (
	"context"
	"fmt"
	"sync"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/spf13/cobra"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

// whoisResult is the outcome of one provider lookup.
type whoisResult struct {
	details   *paymailDetails
	available bool  // the provider answered that the handle is not registered
	err       error // the lookup itself failed
}

// whoisOptions are the flags of the whois command.
type whoisOptions struct {
	skipPublicProfile bool
}

// newWhoisCmd returns `whois <handle>`.
func newWhoisCmd(a *app) *cobra.Command {
	opts := &whoisOptions{}

	cmd := &cobra.Command{
		Use:   "whois <handle>",
		Short: "Find a handle across the public paymail providers",
		Long: fmt.Sprintf(`
        .__           .__
__  _  _|  |__   ____ |__| ______
\ \/ \/ /  |  \ /  _ \|  |/  ___/
 \     /|   Y  (  <_> )  |\___ \
  \/\_/ |___|  /\____/|__/____  >
             \/               \/

Search %d public paymail providers for a handle, in parallel, and show who
owns it on each one, or where it is still available.

The handle may be bare (mrz), a provider shortcut ($handle for HandCash,
1handle for RelayX), or a full paymail address, whose alias is searched.`, len(a.providers)),
		Aliases:    []string{"who", "w"},
		SuggestFor: []string{"lookup"},
		Example:    a.examples("whois "+exampleAlias, "w 1"+exampleAlias, `w \$handcash`),
		Args:       exactArgs(1, "a handle"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runWhois(cmd.Context(), args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.skipPublicProfile, "skip-public-profile", false, "Skip the public profile requests")
	deprecateFlags(cmd, reasonServiceShutDown, retiredIntegrationFlags()...)

	return cmd
}

// runWhois searches every provider concurrently and prints the results in
// provider order.
func (a *app) runWhois(ctx context.Context, raw string, opts *whoisOptions) error {
	handle, err := parseHandle(raw)
	if err != nil {
		return err
	}

	s, err := a.newSession("")
	if err != nil {
		return err
	}

	s.quiet = true

	a.out.Header(output.LevelDefault, fmt.Sprintf("Searching %d providers for %s...", len(a.providers), a.out.Value(handle)))

	// Each goroutine writes only its own slot, so no lock is needed.
	results := make([]whoisResult, len(a.providers))

	var wg sync.WaitGroup
	for i, p := range a.providers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			results[i] = s.whoisLookup(ctx, handle, p, opts)
		}()
	}

	wg.Wait()

	failed := 0

	for i, result := range results {
		p := a.providers[i]

		switch {
		case result.err != nil:
			failed++

			a.out.Header(output.LevelBold, handle+"@"+p.Domain)
			a.out.Error("Lookup failed: %s", result.err)
		case result.available:
			a.out.Header(output.LevelBold, handle+"@"+p.Domain)
			a.out.Success("The handle %s might be available! Reserve it now: %s", handle, p.Link)
		default:
			result.details.render(a.out)
		}
	}

	if failed == len(a.providers) {
		return fmt.Errorf("%w for %s", errNoProviderResponded, handle)
	}

	return nil
}

// whoisLookup checks one provider for handle.
func (s *session) whoisLookup(ctx context.Context, handle string, p provider, opts *whoisOptions) whoisResult {
	addr := address{Alias: handle, Domain: p.Domain, Address: handle + "@" + p.Domain}

	caps, err := s.capabilities(ctx, p.Domain, true)
	if err != nil {
		return whoisResult{err: err}
	}

	pkiURL, err := requireCapability(caps, p.Domain, paymail.BRFCPki, paymail.BRFCPkiAlternate)
	if err != nil {
		return whoisResult{err: err}
	}

	pki, err := s.pki(pkiURL, addr, true)
	if isNotFound(pki, err) {
		return whoisResult{available: true}
	}

	if err != nil {
		return whoisResult{err: err}
	}

	return whoisResult{details: &paymailDetails{
		Address:       addr,
		Provider:      p,
		PKI:           pki,
		PublicProfile: s.optionalProfile(caps, addr, opts.skipPublicProfile),
	}}
}
