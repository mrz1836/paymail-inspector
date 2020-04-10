package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
)

// Default flag values
var (
	skipDnsCheck bool
	nameServer   string
	port         int
	priority     int
	protocol     string
	serviceName  string
	skipSSLCheck bool
	weight       int
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a paymail address or domain",
	Long: `Validate a specific paymail address (user@domain.tld) or validate a domain for required paymail capabilities. 
				By default, this will check for a SRV record, DNSSEC and SSL for the domain. 
				Finally, it will list the capabilities for the target and resolve any address given as well.`,
	Example:    "validate " + defaultDomainName,
	Aliases:    []string{"check", "inspect"},
	SuggestFor: []string{"valid", "check", "lookup"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("requires either a domain or paymail address")
		} else if len(args) > 1 {
			return chalker.Error("validate only supports one domain or address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		chalker.Log(chalker.DEFAULT, fmt.Sprintf("starting validation... found args: %s", args))

		// Extract the parts given
		domain, paymailAddress := paymail.ExtractParts(args[0])

		// Are we an address?
		if len(paymailAddress) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("paymail address detected: %s", paymailAddress))

			// Validate the format for the paymail address (paymail addresses follow conventional email requirements)
			if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("paymail address failed format validation: %s", err.Error()))
				return
			} else if !ok {
				chalker.Log(chalker.ERROR, "paymail address failed format validation: unknown reason")
				return
			}

		} else {
			chalker.Log(chalker.INFO, fmt.Sprintf("domain detected: %s", domain))
		}

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			chalker.Log(chalker.ERROR, fmt.Sprintf("domain name is invalid: %s", domain))
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			chalker.Log(chalker.ERROR, fmt.Sprintf("domain name failed DNS check: %", domain))
			return
		}

		// Get the SRV record
		chalker.Log(chalker.DEFAULT, "getting SRV record...")
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("error getting SRV record: %s", err.Error()))
			return
		}

		// Validate the SRV record for the domain name (using all flags or default values)
		if err = paymail.ValidateSRVRecord(srv, nameServer, port, priority, weight); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("%s failed validating SRV record: %s", err.Error()))
			return
		}

		// Success message
		chalker.Log(chalker.SUCCESS, "SRV record passed all validations (target, port, priority, weight)")
		chalker.Log(chalker.INFO, fmt.Sprintf("target record found: %s", srv.Target))

		// Validate the DNSSEC if the flag is true
		if !skipDnsCheck {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("checking %s for DNSSEC validation...", srv.Target))

			if result := paymail.CheckDNSSEC(srv.Target, nameServer); result.DNSSEC {
				chalker.Log(chalker.SUCCESS, fmt.Sprintf("DNSSEC found and valid and found %d DS record(s)", result.Answer.DSRecordCount))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("DNSSEC not found or invalid for %s", result.Domain))
				if len(result.ErrorMessage) > 0 {
					chalker.Log(chalker.ERROR, fmt.Sprintf("error checking DNSSEC: %s", result.ErrorMessage))
				}
				return
			}
		} else {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("skipping DNSSEC check for %s", srv.Target))
		}

		// Validate that there is SSL on the target
		if !skipSSLCheck {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("checking %s for SSL validation...", srv.Target))

			var valid bool
			if valid, err = paymail.CheckSSL(srv.Target, nameServer); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error checking SSL: %s", err.Error()))
				return
			} else if !valid {
				chalker.Log(chalker.ERROR, "SSL is not valid or not found")
				return
			}
			chalker.Log(chalker.SUCCESS, "SSL found and valid")
		} else {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("skipping SSL check for %s", srv.Target))
		}

		// Now lookup the capabilities
		chalker.Log(chalker.DEFAULT, "getting capabilities...")
		var capabilities *paymail.CapabilitiesResponse
		if capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port)); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("get capabilities failed: %s", err.Error()))
			return
		}

		// Check the version
		if capabilities.BsvAlias != bsvAliasVersion {
			chalker.Log(chalker.ERROR, fmt.Sprintf("capabilities bsvalias version mismatch, expected: %s but got: %s", bsvAliasVersion, capabilities.BsvAlias))
			return
		}

		// Show some basic results
		chalker.Log(chalker.INFO, fmt.Sprintf("bsvalias version: %s", capabilities.BsvAlias))
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("total capabilities found: %d", len(capabilities.Capabilities)))

		// Missing required capabilities?
		if len(capabilities.Pki) == 0 {
			chalker.Log(chalker.WARN, fmt.Sprintf("missing required capability: %s", paymail.CapabilityPki))
			return
		} else if len(capabilities.PaymentDestination) == 0 {
			chalker.Log(chalker.WARN, fmt.Sprintf("missing required capability: %s", paymail.CapabilityPaymentDestination))
			return
		}

		// Passed the capabilities check
		chalker.Log(chalker.INFO, fmt.Sprintf("found required %s and %s capabilities%s", paymail.CapabilityPki, paymail.CapabilityPaymentDestination))

		// Only if we have an address (extra validations)
		if len(paymailAddress) > 0 {

			// Get the alias of the address
			parts := strings.Split(paymailAddress, "@")

			// Get the PKI for the given address
			chalker.Log(chalker.DEFAULT, "getting PKI...")

			var pki *paymail.PKIResponse
			if pki, err = paymail.GetPKI(capabilities.Pki, parts[0], domain); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("get PKI failed: %s", err.Error()))
				return
			}

			// Check the version
			if pki.BsvAlias != bsvAliasVersion {
				chalker.Log(chalker.ERROR, fmt.Sprintf("pki bsvalias version mismatch, expected: %s but got: %s", bsvAliasVersion, capabilities.BsvAlias))
				return
			}

			// Found the paymail address
			chalker.Log(chalker.SUCCESS, fmt.Sprintf("fetching PKI was successful - found PubKey: %s", pki.PubKey))
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	// Custom name server for DNS resolution (looking for the SRV record)
	validateCmd.Flags().StringVarP(&nameServer, "nameserver", "n", defaultNameServer, "DNS name server for resolving records")

	// Custom service name for the SRV record
	validateCmd.Flags().StringVarP(&serviceName, "service", "s", paymail.DefaultServiceName, "Service name in the SRV record")

	// Custom protocol for the SRV record
	validateCmd.Flags().StringVar(&protocol, "protocol", paymail.DefaultProtocol, "Protocol in the SRV record")

	// Custom port for the SRV record (target address)
	validateCmd.Flags().IntVarP(&port, "port", "p", paymail.DefaultPort, "Port that is found in the SRV record")

	// Custom priority for the SRV record
	validateCmd.Flags().IntVar(&priority, "priority", paymail.DefaultPriority, "Priority value that is found in the SRV record")

	// Custom weight for the SRV record
	validateCmd.Flags().IntVarP(&weight, "weight", "w", paymail.DefaultWeight, "Weight value that is found in the SRV record")

	// Run the DNSSEC check on the target domain
	validateCmd.Flags().BoolVarP(&skipDnsCheck, "skip-dnssec", "d", false, "Skip checking DNSSEC of the target")

	// Run the SSL check on the target domain
	validateCmd.Flags().BoolVar(&skipSSLCheck, "skip-ssl", false, "Skip checking SSL of the target")
}
