package cli

import (
	"fmt"
	"sort"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/spf13/cobra"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

// capabilityWidth aligns capability names in the listing.
const capabilityWidth = 28

// newCapabilitiesCmd returns `capabilities <domain|paymail>`.
func newCapabilitiesCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:     "capabilities <domain|paymail>",
		Short:   "List the capabilities a paymail domain publishes",
		Aliases: []string{"c", "inspect"},
		Long: `
                          ___.   .__.__  .__  __  .__
  ____ _____  ___________ \_ |__ |__|  | |__|/  |_|__| ____   ______
_/ ___\\__  \ \____ \__  \ | __ \|  |  | |  \   __\  |/ __ \ /  ___/
\  \___ / __ \|  |_> > __ \| \_\ \  |  |_|  ||  | |  \  ___/ \___ \
 \___  >____  /   __(____  /___  /__|____/__||__| |__|\___  >____  >
     \/     \/|__|       \/    \/                         \/     \/

Capability discovery is how a paymail client learns the features a paymail
service supports, and the endpoint of each. Drawing on RFC 5785, the service
publishes a machine-readable document at /.well-known/bsvalias on the host
named by its SRV record.

Each capability is listed with the BRFC specification it implements, when the
specification is known.

Read more at: ` + specMirror + "02-02-capability-discovery.html",
		Example: a.examples("capabilities "+exampleDomain, "c "+exampleAddress),
		Args:    exactArgs(1, "a domain or paymail address"),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain, err := parseDomain(args[0])
			if err != nil {
				return err
			}

			s, err := a.newSession("")
			if err != nil {
				return err
			}

			caps, err := s.capabilities(cmd.Context(), domain, false)
			if err != nil {
				return err
			}

			a.renderCapabilities(caps, s.client.GetBRFCs())

			return nil
		},
	}
}

// renderCapabilities lists capabilities in a stable, sorted order.
func (a *app) renderCapabilities(caps *paymail.CapabilitiesResponse, specs []*paymail.BRFCSpec) {
	a.out.Header(output.LevelBold, fmt.Sprintf("Listing %d capabilities (bsvalias %s)...", len(caps.Capabilities), caps.BsvAlias))

	keys := make([]string, 0, len(caps.Capabilities))
	for key := range caps.Capabilities {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		name := fmt.Sprintf("%-*s", capabilityWidth, key)

		var value string

		switch v := caps.Capabilities[key].(type) {
		case string:
			value = a.out.Caution(v)
		case bool:
			value = a.out.Alert("Disabled")
			if v {
				value = a.out.Good("Enabled")
			}
		default:
			value = a.out.Caution(fmt.Sprintf("%v", v))
		}

		line := a.out.Value(name) + " " + value
		if title := brfcTitle(specs, key); title != "" {
			line += "  " + a.out.Label("("+title+")")
		}

		a.out.Info("%s", line)
	}
}

// brfcTitle names the BRFC specification behind a capability key, matching
// either its ID or its alias (e.g. "pki").
func brfcTitle(specs []*paymail.BRFCSpec, key string) string {
	for _, spec := range specs {
		if spec.ID == key || (spec.Alias != "" && spec.Alias == key) {
			return spec.Title
		}
	}

	return ""
}
