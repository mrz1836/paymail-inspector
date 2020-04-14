package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a paymail address or domain",
	Long: chalk.Green.Color(`
              .__  .__    .___       __          
___  _______  |  | |__| __| _/____ _/  |_  ____  
\  \/ /\__  \ |  | |  |/ __ |\__  \\   __\/ __ \ 
 \   /  / __ \|  |_|  / /_/ | / __ \|  | \  ___/ 
  \_/  (____  /____/__\____ |(____  /__|  \___  >
            \/             \/     \/          \/`) + `
` + chalk.Yellow.Color(`
Validate a specific paymail address (user@domain.tld) or validate a domain for required paymail capabilities. 

By default, this will check for a SRV record, DNSSEC and SSL for the domain. 

This will also check for required capabilities that all paymail services are required to support.

All these validations are suggestions/requirements from bsvalias spec.

Read more at: `+chalk.Cyan.Color("http://bsvalias.org/index.html")),
	Example:    configDefault + " validate " + defaultDomainName,
	Aliases:    []string{"val", "check"},
	SuggestFor: []string{"valid"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("validate requires either a domain or paymail address")
		} else if len(args) > 1 {
			return chalker.Error("validate only supports one domain or address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		domain, paymailAddress := paymail.ExtractParts(args[0])
		var err error

		// Are we an address?
		if len(paymailAddress) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("paymail detected: %s", chalk.Cyan.Color(paymailAddress)))

			// Validate the paymail address and domain (error already shown)
			if ok := validatePaymailAndDomain(paymailAddress, domain); !ok {
				return
			}

		} else {
			chalker.Log(chalker.INFO, fmt.Sprintf("domain detected: %s", chalk.Cyan.Color(domain)))

			// Check for a real domain (require at least one period)
			if !strings.Contains(domain, ".") {
				chalker.Log(chalker.ERROR, fmt.Sprintf("domain name is invalid: %s", domain))
				return
			} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
				chalker.Log(chalker.ERROR, fmt.Sprintf("domain name failed DNS check: %s", domain))
				return
			}
		}

		// Used for future checks
		checkDomain := domain

		// Get the SRV record
		if !skipSrvCheck {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("getting SRV record for: %s...", chalk.Cyan.Color(domain)))

			var srv *net.SRV
			srv, err = paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
			if err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error getting SRV record: %s", err.Error()))
			} else if srv != nil {
				// Validate the SRV record for the domain name (using all flags or default values)
				if err = paymail.ValidateSRVRecord(srv, nameServer, port, priority, weight); err != nil {
					chalker.Log(chalker.ERROR, fmt.Sprintf("failed validating SRV record: %s", err.Error()))
				} else {
					checkDomain = srv.Target

					// Success message
					chalker.Log(chalker.SUCCESS, "SRV record passed all validations (target, port, priority, weight)")
					chalker.Log(chalker.INFO, fmt.Sprintf("target record found: %s", chalk.Cyan.Color(checkDomain)))
				}
			}

		} else {
			chalker.Log(chalker.WARN, fmt.Sprintf("skipping SRV record check for: %s", chalk.Cyan.Color(checkDomain)))
		}

		// Validate the DNSSEC if the flag is true
		if !skipDnsCheck {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("checking %s for DNSSEC validation...", chalk.Cyan.Color(checkDomain)))

			if result := paymail.CheckDNSSEC(checkDomain, nameServer); result.DNSSEC {
				chalker.Log(chalker.SUCCESS, fmt.Sprintf("DNSSEC found and valid and found %d DS record(s)", result.Answer.DSRecordCount))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("DNSSEC not found or invalid for %s", result.Domain))
				if len(result.ErrorMessage) > 0 {
					chalker.Log(chalker.ERROR, fmt.Sprintf("error checking DNSSEC: %s", result.ErrorMessage))
				}
			}
		} else {
			chalker.Log(chalker.WARN, fmt.Sprintf("skipping DNSSEC check for: %s", chalk.Cyan.Color(checkDomain)))
		}

		// Validate that there is SSL on the target
		if !skipSSLCheck {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("checking %s for SSL validation...", chalk.Cyan.Color(checkDomain)))

			var valid bool
			if valid, err = paymail.CheckSSL(checkDomain, nameServer); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error checking SSL: %s", err.Error()))
			} else if !valid {
				chalker.Log(chalker.ERROR, "zero SSL certificates found (or timed out)")
			}
			chalker.Log(chalker.SUCCESS, fmt.Sprintf("SSL found and valid for: %s", checkDomain))
		} else {
			chalker.Log(chalker.WARN, fmt.Sprintf("skipping SSL check for: %s", chalk.Cyan.Color(checkDomain)))
		}

		// Get the capabilities
		var capabilities *paymail.CapabilitiesResponse
		if capabilities, err = getCapabilities(domain); err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				chalker.Log(chalker.WARN, fmt.Sprintf("no capabilities found for: %s", domain))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error: %s", err.Error()))
			}
			return
		}

		// Missing required capabilities?
		pkiUrl := capabilities.GetValueString(paymail.BRFCPki, paymail.BRFCPkiAlternate)
		resolveUrl := capabilities.GetValueString(paymail.BRFCPaymentDestination, paymail.BRFCBasicAddressResolution)
		if len(pkiUrl) == 0 {
			chalker.Log(chalker.WARN, fmt.Sprintf("missing required capability: %s", paymail.BRFCPki))
		} else if len(resolveUrl) == 0 {
			chalker.Log(chalker.WARN, fmt.Sprintf("missing required capability: %s", paymail.BRFCPaymentDestination))
		} else if len(pkiUrl) > 0 && len(resolveUrl) > 0 {
			chalker.Log(chalker.SUCCESS, fmt.Sprintf("found required capabilities: [%s] [%s]", paymail.BRFCPki, paymail.BRFCPaymentDestination))
		}

		// Only if we have an address (basic validation that the address exists)
		if len(paymailAddress) > 0 && len(pkiUrl) > 0 {

			// Get the alias of the address
			parts := strings.Split(paymailAddress, "@")

			// Get the PKI for the given address
			var pki *paymail.PKIResponse
			if pki, err = getPki(pkiUrl, parts[0], parts[1]); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error: %s", err.Error()))
				return
			} else if pki != nil {
				chalker.Log(chalker.SUCCESS, fmt.Sprintf("found paymail: %s@%s's pubkey: %s", parts[0], parts[1], pki.PubKey))
			}
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

	// Run the SRV check on the domain
	validateCmd.Flags().BoolVar(&skipSrvCheck, "skip-srv", false, "Skip checking SRV record of the main domain")

	// Run the DNSSEC check on the target domain
	validateCmd.Flags().BoolVarP(&skipDnsCheck, "skip-dnssec", "d", false, "Skip checking DNSSEC of the target domain")

	// Run the SSL check on the target domain
	validateCmd.Flags().BoolVar(&skipSSLCheck, "skip-ssl", false, "Skip checking SSL of the target domain")
}
