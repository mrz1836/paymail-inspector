package cli

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/spf13/cobra"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

// defaultNameServer resolves DNS for validate unless --nameserver is given.
const defaultNameServer = "8.8.8.8"

// validateOptions are the flags of the validate command.
type validateOptions struct {
	nameServer string
	srv        srvOptions
	skipSRV    bool
	skipDNSSEC bool
	skipSSL    bool
}

// validation tallies the outcome of each check.
type validation struct {
	out      *output.Printer
	failures []string
	warnings int
}

// fail records a failed check.
func (v *validation) fail(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	v.failures = append(v.failures, msg)
	v.out.Error("%s", msg)
}

// warn records a check that passed with a caveat.
func (v *validation) warn(format string, args ...any) {
	v.warnings++
	v.out.Warn(format, args...)
}

// newValidateCmd returns `validate <domain|paymail>`.
func newValidateCmd(a *app) *cobra.Command {
	opts := &validateOptions{srv: defaultSRVOptions()}

	cmd := &cobra.Command{
		Use:   "validate <domain|paymail>",
		Short: "Validate a paymail domain or address (SRV, DNSSEC, SSL, capabilities)",
		Long: `
              .__  .__    .___       __
___  _______  |  | |__| __| _/____ _/  |_  ____
\  \/ /\__  \ |  | |  |/ __ |\__  \\   __\/ __ \
 \   /  / __ \|  |_|  / /_/ | / __ \|  | \  ___/
  \_/  (____  /____/__\____ |(____  /__|  \___  >
            \/             \/     \/          \/

Validate a paymail domain, or a specific paymail address (alias@domain.tld),
against the bsvalias specification:

  1. SRV record   _bsvalias._tcp.<domain> (falls back to <domain>:443)
  2. DNSSEC       on the SRV target (recommended, reported as a warning)
  3. SSL          a valid certificate on the SRV target
  4. Capabilities the required pki and paymentDestination capabilities
  5. PKI          for an address: its public key resolves

The command exits non-zero when any required check fails, so it can gate CI.

Read more at: ` + paymailStandard,
		Aliases:    []string{"val", "v"},
		SuggestFor: []string{"valid"},
		Example: a.examples(
			"validate "+exampleDomain,
			"v "+exampleAddress,
			"v "+exampleDomain+" --skip-dnssec --nameserver 1.1.1.1",
		),
		Args: exactArgs(1, "a domain or paymail address"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runValidate(cmd.Context(), args[0], opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.nameServer, "nameserver", "n", defaultNameServer, "DNS name server used for the lookups")
	flags.StringVarP(&opts.srv.service, "service", "s", opts.srv.service, "Expected service name in the SRV record")
	flags.StringVar(&opts.srv.protocol, "protocol", opts.srv.protocol, "Expected protocol in the SRV record")
	flags.Uint16VarP(&opts.srv.port, "port", "p", opts.srv.port, "Expected port in the SRV record")
	flags.Uint16Var(&opts.srv.priority, "priority", opts.srv.priority, "Expected priority in the SRV record")
	flags.Uint16VarP(&opts.srv.weight, "weight", "w", opts.srv.weight, "Expected weight in the SRV record")
	flags.BoolVar(&opts.skipSRV, "skip-srv", false, "Skip the SRV record check")
	flags.BoolVarP(&opts.skipDNSSEC, "skip-dnssec", "d", false, "Skip the DNSSEC check")
	flags.BoolVar(&opts.skipSSL, "skip-ssl", false, "Skip the SSL check")

	return cmd
}

// runValidate runs every check and fails if any required one failed.
func (a *app) runValidate(ctx context.Context, raw string, opts *validateOptions) error {
	addr, domain, err := a.detectValidationTarget(raw)
	if err != nil {
		return err
	}

	s, err := a.newSession(opts.nameServer)
	if err != nil {
		return err
	}

	s.srv = opts.srv
	v := &validation{out: a.out}

	srv := s.checkSRV(ctx, v, domain, opts.skipSRV)
	host := srvHost(srv, domain)

	s.checkDNSSEC(v, host, opts.skipDNSSEC)
	s.checkSSL(v, host, opts.skipSSL)

	if caps := s.checkCapabilities(ctx, v, domain, srv, opts.skipSRV); caps != nil && addr.Address != "" {
		s.checkPKI(v, caps, addr)
	}

	if len(v.failures) > 0 {
		a.out.Header(output.LevelBold, fmt.Sprintf("%s failed %d check(s)", domain, len(v.failures)))

		return fmt.Errorf("%w: %s: %s", errValidationFailed, domain, strings.Join(v.failures, "; "))
	}

	a.out.Header(output.LevelBold, a.out.Good(fmt.Sprintf("%s passed validation (%d warning(s))", domain, v.warnings)))

	return nil
}

// detectValidationTarget reads the argument as a paymail address (or handle)
// or as a domain, and returns the address (empty for a domain) and domain.
func (a *app) detectValidationTarget(raw string) (address, string, error) {
	a.out.Header(output.LevelDefault, "Detecting validation type...")

	if strings.Contains(raw, "@") || strings.HasPrefix(raw, "$") {
		addr, err := parseAddress(raw)
		if err != nil {
			return address{}, "", err
		}

		a.out.Default("Paymail address detected: %s", a.out.Value(addr.Address))

		return addr, addr.Domain, nil
	}

	domain, err := parseDomain(raw)
	if err != nil {
		return address{}, "", err
	}

	a.out.Default("Domain detected: %s", a.out.Value(domain))

	return address{}, domain, nil
}

// srvHost returns the host an SRV record points at, or domain without one.
func srvHost(srv *net.SRV, domain string) string {
	if srv != nil && srv.Target != "" {
		return srv.Target
	}

	return domain
}

// checkSRV validates the SRV record and returns it, so later checks run
// against its target. It returns nil when the check is skipped, which
// makes the later checks run against the domain itself.
func (s *session) checkSRV(ctx context.Context, v *validation, domain string, skip bool) *net.SRV {
	if skip {
		v.warn("Skipping SRV record check for %s", domain)
		return nil
	}

	srv, err := s.srvRecord(ctx, domain, true, false)
	if err != nil {
		v.fail("%s", err)
	}

	return srv
}

// checkDNSSEC reports whether DNSSEC is configured for host. DNSSEC is
// recommended rather than required, so its absence is a warning.
func (s *session) checkDNSSEC(v *validation, host string, skip bool) {
	s.step("Checking %s for DNSSEC...", s.out.Value(host))

	if skip {
		v.warn("Skipping DNSSEC check for %s", host)
		return
	}

	result := s.client.CheckDNSSEC(host)

	switch {
	case result == nil:
		v.warn("DNSSEC check returned no result for %s", host)
	case result.DNSSEC:
		s.out.Success("DNSSEC is valid (%d DS record(s))", result.Answer.DSRecordCount)
	default:
		v.warn("DNSSEC not found or invalid for %s; check manually: https://dnsviz.net/d/%s/dnssec/", host, host)

		if result.ErrorMessage != "" {
			s.out.Dim("DNSSEC detail: %s", result.ErrorMessage)
		}
	}
}

// checkSSL requires a valid certificate on host.
func (s *session) checkSSL(v *validation, host string, skip bool) {
	s.step("Checking %s for SSL...", s.out.Value(host))

	if skip {
		v.warn("Skipping SSL check for %s", host)
		return
	}

	valid, err := s.client.CheckSSL(host)

	switch {
	case err != nil:
		v.fail("SSL check failed for %s: %s", host, err)
	case !valid:
		v.fail("No valid SSL certificate found for %s", host)
	default:
		s.out.Success("SSL certificate is valid for %s", host)
	}
}

// checkCapabilities requires the capabilities every paymail service must
// publish, and returns them for the address checks (nil if unavailable).
func (s *session) checkCapabilities(ctx context.Context, v *validation, domain string, srv *net.SRV, skippedSRV bool) *paymail.CapabilitiesResponse {
	var (
		caps *paymail.CapabilitiesResponse
		err  error
	)

	// Reuse the SRV record just checked; look it up only when its check was
	// skipped. A failed check leaves srv nil, which falls back to the domain.
	if skippedSRV {
		caps, err = s.capabilities(ctx, domain, false)
	} else {
		caps, err = s.capabilitiesAt(domain, srv, false)
	}

	if err != nil {
		v.fail("%s", err)
		return nil
	}

	pkiURL := capString(caps, paymail.BRFCPki, paymail.BRFCPkiAlternate)
	resolveURL := capString(caps, paymail.BRFCPaymentDestination, paymail.BRFCBasicAddressResolution)

	if pkiURL == "" {
		v.fail("Missing required capability: %s", paymail.BRFCPki)
	}

	// Basic address resolution is being retired in favor of P2P payment
	// destinations, so a provider offering only P2P is still reachable.
	switch {
	case resolveURL != "":
	case capString(caps, paymail.BRFCP2PPaymentDestination, "") != "":
		v.warn("No %s capability; the provider offers P2P payment destinations (%s) instead",
			paymail.BRFCPaymentDestination, paymail.BRFCP2PPaymentDestination)
	default:
		v.fail("Missing required capability: %s (or %s)", paymail.BRFCPaymentDestination, paymail.BRFCP2PPaymentDestination)
	}

	if pkiURL != "" && resolveURL != "" {
		s.out.Success("Found required capabilities: [%s] [%s]", paymail.BRFCPki, paymail.BRFCPaymentDestination)
	}

	return caps
}

// checkPKI requires that an address's public key resolves.
func (s *session) checkPKI(v *validation, caps *paymail.CapabilitiesResponse, addr address) {
	pkiURL := capString(caps, paymail.BRFCPki, paymail.BRFCPkiAlternate)
	if pkiURL == "" {
		return // already reported as a missing capability
	}

	pki, err := s.pki(pkiURL, addr, false)
	if err != nil {
		v.fail("%s", err)
		return
	}

	s.out.Header(output.LevelBold, "Paymail information for "+s.out.Value(addr.Address))
	field(s.out, "PubKey", pki.PubKey)
}
