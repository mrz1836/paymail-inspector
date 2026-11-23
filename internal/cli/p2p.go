package cli

import (
	"context"
	"fmt"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/spf13/cobra"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

// defaultSatoshis is the amount requested when --satoshis is not given.
const defaultSatoshis = 1000

// p2pOptions are the flags of the p2p command.
type p2pOptions struct {
	satoshis          uint64
	skipPublicProfile bool
}

// newP2PCmd returns `p2p <paymail>`.
func newP2PCmd(a *app) *cobra.Command {
	opts := &p2pOptions{}

	cmd := &cobra.Command{
		Use:   "p2p <paymail>",
		Short: "Start a P2P payment request and list its outputs",
		Long: `
       ________
______ \_____  \______
\____ \ /  ____/\____ \
|  |_> >       \|  |_> >
|   __/\_______ \   __/
|__|           \/__|

Start a P2P payment destination request with the receiver for an amount in
satoshis, and list the outputs (script, satoshis, address) and the reference
it returns.

P2P payment destination is the successor to basic address resolution: rather
than one address it returns a set of outputs for a P2P transaction.

Read more at: ` + paymailStandard + " (BRFC 2a40af698840)",
		Aliases:    []string{"send"},
		SuggestFor: []string{"sending"},
		Example:    a.examples(`p2p \$handcash`, `p2p \$handcash --satoshis 5000`, "p2p handcash@handcash.io"),
		Args:       exactArgs(1, "a paymail address"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runP2P(cmd.Context(), args[0], opts)
		},
	}

	cmd.Flags().Uint64Var(&opts.satoshis, "satoshis", 0, fmt.Sprintf("Amount in satoshis to request (default %d)", defaultSatoshis))
	cmd.Flags().BoolVar(&opts.skipPublicProfile, "skip-public-profile", false, "Skip the public profile request")
	deprecateFlags(cmd, reasonServiceShutDown, "skip-bitpic")

	return cmd
}

// runP2P requests a P2P payment destination and prints the outputs.
func (a *app) runP2P(ctx context.Context, raw string, opts *p2pOptions) error {
	addr, err := parseAddress(raw)
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

	destinationURL, err := requireCapability(caps, addr.Domain, paymail.BRFCP2PPaymentDestination, "")
	if err != nil {
		return err
	}

	satoshis := opts.satoshis
	if satoshis == 0 {
		satoshis = defaultSatoshis
	}

	destination, err := s.p2pDestination(destinationURL, addr, satoshis)
	if err != nil {
		return err
	}

	a.renderP2P(addr, destination, s.optionalProfile(caps, addr, opts.skipPublicProfile))

	return nil
}

// renderP2P prints a P2P payment destination response.
func (a *app) renderP2P(addr address, destination *paymail.PaymentDestinationResponse, profile *paymail.PublicProfileResponse) {
	a.out.Header(output.LevelBold, "P2P information for "+a.out.Value(addr.Address))

	if profile != nil {
		field(a.out, "Name", profile.Name)
		field(a.out, "Avatar", profile.Avatar)
	}

	field(a.out, "Reference", destination.Reference)

	for i, out := range destination.Outputs {
		a.out.Header(output.LevelDefault, fmt.Sprintf("Output #%d", i+1))
		field(a.out, "Script", out.Script)
		field(a.out, "Satoshis", fmt.Sprintf("%d", out.Satoshis))
		field(a.out, "Address", out.Address)
	}
}
