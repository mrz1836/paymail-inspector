package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-validate"
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
			return fmt.Errorf("%s requires either a domain or paymail address\n", logPrefix)
		} else if len(args) > 1 {
			return fmt.Errorf("%s validate only supports one domain or address at a time\n", logPrefix)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("%s starting validation... found args: %s\n", logPrefix, args)

		// Extract the parts given
		domain, paymailAddress := paymail.ExtractParts(args[0])

		// Are we an address?
		if len(paymailAddress) > 0 {
			fmt.Printf("%s paymail address detected: %s\n", logPrefix, paymailAddress)

			// Validate the email format for the paymail address (paymail addresses follow conventional email requirements)
			if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
				fmt.Printf("%s paymail address failed email format validation: %s\n", logPrefix, err.Error())
				return
			} else if !ok {
				fmt.Printf("%s paymail address failed email format validation: %s\n", logPrefix, "unknown reason")
				return
			}

		} else {
			fmt.Printf("%s domain detected: %s\n", logPrefix, domain)
		}

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			fmt.Printf("%s domain name is invalid: %s\n", logPrefix, domain)
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			fmt.Printf("%s domain name failed DNS check: %s\n", logPrefix, domain)
			return
		}

		// Get the SRV record
		fmt.Printf("%s getting SRV record...\n", logPrefix)
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			fmt.Printf("%s error getting SRV record: %s\n", logPrefix, err.Error())
			return
		}

		// Validate the SRV record for the domain name (using all flags or default values)
		if err = paymail.ValidateSRVRecord(srv, nameServer, port, priority, weight); err != nil {
			fmt.Printf("%s failed validating SRV record: %s\n", logPrefix, err.Error())
			return
		}

		// Remove last character if period (comes from DNS records)
		target := strings.TrimSuffix(srv.Target, ".")

		// Success message
		fmt.Printf("%s SRV record passed all validations (target, port, priority, weight)\n", logPrefix)
		fmt.Printf("%s target record found: %s\n", logPrefix, target)

		// Validate the DNSSEC if the flag is true
		if !skipDnsCheck {
			fmt.Printf("%s checking %s for DNSSEC validation...\n", logPrefix, target)

			if result := paymail.CheckDNSSEC(target, nameServer); result.DNSSEC {
				fmt.Printf("%s DNSSEC found and valid and found %d DS record(s)\n", logPrefix, result.Answer.DSRecordCount)
			} else {
				fmt.Printf("%s DNSSEC not found or invalid for %s\n", logPrefix, result.Domain)
				if len(result.ErrorMessage) > 0 {
					fmt.Printf("%s error checking DNSSEC: %s\n", logPrefix, result.ErrorMessage)
				}
				return
			}
		} else {
			fmt.Printf("%s skipping DNSSEC check for %s\n", logPrefix, target)
		}

		// Validate that there is SSL on the target
		if !skipSSLCheck {
			fmt.Printf("%s checking %s for SSL validation...\n", logPrefix, target)

			var valid bool
			if valid, err = paymail.CheckSSL(target, nameServer); err != nil {
				fmt.Printf("%s error checking SSL: %s\n", logPrefix, err.Error())
				return
			} else if !valid {
				fmt.Printf("%s SSL is not valid or not found\n", logPrefix)
				return
			}
			fmt.Printf("%s SSL found and valid\n", logPrefix)
		} else {
			fmt.Printf("%s skipping SSL check for %s\n", logPrefix, target)
		}

		// Now lookup the capabilities
		fmt.Printf("%s getting capabilities...\n", logPrefix)
		var capabilities *paymail.CapabilitiesResponse
		if capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port)); err != nil {
			fmt.Printf("%s get capabilities failed: %s\n", logPrefix, err.Error())
			return
		}

		// Check the version
		if capabilities.BsvAlias != bsvAliasVersion {
			fmt.Printf("%s capabilities bsvalias version mismatch, expected: %s but got: %s\n", logPrefix, bsvAliasVersion, capabilities.BsvAlias)
			return
		}

		// Show some basic results
		fmt.Printf("%s bsvalias version: %s\n", logPrefix, capabilities.BsvAlias)
		fmt.Printf("%s total capabilities found: %d\n", logPrefix, len(capabilities.Capabilities))

		// Missing required capabilities?
		if len(capabilities.Pki) == 0 {
			fmt.Printf("%s missing required capability: %s\n", logPrefix, paymail.CapabilityPki)
			return
		} else if len(capabilities.PaymentDestination) == 0 {
			fmt.Printf("%s missing required capability: %s\n", logPrefix, paymail.CapabilityPaymentDestination)
			return
		}

		// Passed the capabilities check
		fmt.Printf("%s found required %s and %s capabilities\n", logPrefix, paymail.CapabilityPki, paymail.CapabilityPaymentDestination)

		// Only if we have an address (extra validations)
		if len(paymailAddress) > 0 {

			// Get the alias of the address
			parts := strings.Split(paymailAddress, "@")

			// Get the PKI for the given address
			fmt.Printf("%s getting PKI...\n", logPrefix)
			var pki *paymail.PKIResponse
			if pki, err = paymail.GetPKI(capabilities.Pki, parts[0], domain); err != nil {
				fmt.Printf("%s get PKI failed: %s\n", logPrefix, err.Error())
				return
			}

			// Check the version
			if pki.BsvAlias != bsvAliasVersion {
				fmt.Printf("%s pki bsvalias version mismatch, expected: %s but got: %s\n", logPrefix, bsvAliasVersion, capabilities.BsvAlias)
				return
			}

			// Found the paymail address
			fmt.Printf("%s fetching PKI was successful - found PubKey: %s\n", logPrefix, pki.PubKey)
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
