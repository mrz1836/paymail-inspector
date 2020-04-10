package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
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
			return fmt.Errorf("%s requires either a domain or paymail address\n", chalk.Dim.TextStyle(logPrefix))
		} else if len(args) > 1 {
			return fmt.Errorf("%s validate only supports one domain or address at a time\n", chalk.Dim.TextStyle(logPrefix))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("%s starting validation... found args: %s\n", chalk.Dim.TextStyle(logPrefix), args)

		// Extract the parts given
		domain, paymailAddress := paymail.ExtractParts(args[0])

		// Are we an address?
		if len(paymailAddress) > 0 {
			fmt.Printf("%s paymail address detected: %s\n", chalk.Dim.TextStyle(logPrefix), paymailAddress)

			// Validate the format for the paymail address (paymail addresses follow conventional email requirements)
			if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
				fmt.Printf("%s paymail address failed format validation: %s\n", chalk.Dim.TextStyle(logPrefix), err.Error())
				return
			} else if !ok {
				fmt.Printf("%s paymail address failed format validation: %s\n", chalk.Dim.TextStyle(logPrefix), "unknown reason")
				return
			}

		} else {
			fmt.Printf("%s domain detected: %s\n", chalk.Dim.TextStyle(logPrefix), domain)
		}

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			fmt.Printf("%s%s domain name is invalid: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), domain, chalk.Reset)
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			fmt.Printf("%s%s domain name failed DNS check: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), domain, chalk.Reset)
			return
		}

		// Get the SRV record
		fmt.Printf("%s getting SRV record...\n", chalk.Dim.TextStyle(logPrefix))
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			fmt.Printf("%s%s error getting SRV record: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), err.Error(), chalk.Reset)
			return
		}

		// Validate the SRV record for the domain name (using all flags or default values)
		if err = paymail.ValidateSRVRecord(srv, nameServer, port, priority, weight); err != nil {
			fmt.Printf("%s%s failed validating SRV record: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), err.Error(), chalk.Reset)
			return
		}

		// Success message
		fmt.Printf("%s%s SRV record passed all validations (target, port, priority, weight)\n", chalk.Green, chalk.Dim.TextStyle(logPrefix))
		fmt.Printf("%s target record found: %s%s\n", chalk.Dim.TextStyle(logPrefix), srv.Target, chalk.Reset)

		// Validate the DNSSEC if the flag is true
		if !skipDnsCheck {
			fmt.Printf("%s checking %s for DNSSEC validation...\n", chalk.Dim.TextStyle(logPrefix), srv.Target)

			if result := paymail.CheckDNSSEC(srv.Target, nameServer); result.DNSSEC {
				fmt.Printf("%s%s DNSSEC found and valid and found %d DS record(s)%s\n", chalk.Green, chalk.Dim.TextStyle(logPrefix), result.Answer.DSRecordCount, chalk.Reset)
			} else {
				fmt.Printf("%s%s DNSSEC not found or invalid for %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), result.Domain, chalk.Reset)
				if len(result.ErrorMessage) > 0 {
					fmt.Printf("%s%s error checking DNSSEC: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), result.ErrorMessage, chalk.Reset)
				}
				return
			}
		} else {
			fmt.Printf("%s skipping DNSSEC check for %s\n", chalk.Dim.TextStyle(logPrefix), srv.Target)
		}

		// Validate that there is SSL on the target
		if !skipSSLCheck {
			fmt.Printf("%s checking %s for SSL validation...\n", chalk.Dim.TextStyle(logPrefix), srv.Target)

			var valid bool
			if valid, err = paymail.CheckSSL(srv.Target, nameServer); err != nil {
				fmt.Printf("%s%s error checking SSL: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), err.Error(), chalk.Reset)
				return
			} else if !valid {
				fmt.Printf("%s%s SSL is not valid or not found%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), chalk.Reset)
				return
			}
			fmt.Printf("%s%s SSL found and valid%s\n", chalk.Green, chalk.Dim.TextStyle(logPrefix), chalk.Reset)
		} else {
			fmt.Printf("%s skipping SSL check for %s\n", chalk.Dim.TextStyle(logPrefix), srv.Target)
		}

		// Now lookup the capabilities
		fmt.Printf("%s getting capabilities...\n", chalk.Dim.TextStyle(logPrefix))
		var capabilities *paymail.CapabilitiesResponse
		if capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port)); err != nil {
			fmt.Printf("%s%s get capabilities failed: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), err.Error(), chalk.Reset)
			return
		}

		// Check the version
		if capabilities.BsvAlias != bsvAliasVersion {
			fmt.Printf("%s%s capabilities bsvalias version mismatch, expected: %s but got: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), bsvAliasVersion, capabilities.BsvAlias, chalk.Reset)
			return
		}

		// Show some basic results
		fmt.Printf("%s%s bsvalias version: %s%s\n", chalk.Cyan, chalk.Dim.TextStyle(logPrefix), capabilities.BsvAlias, chalk.Reset)
		fmt.Printf("%s total capabilities found: %d\n", chalk.Dim.TextStyle(logPrefix), len(capabilities.Capabilities))

		// Missing required capabilities?
		if len(capabilities.Pki) == 0 {
			fmt.Printf("%s%s missing required capability: %s%s\n", chalk.Yellow, chalk.Dim.TextStyle(logPrefix), paymail.CapabilityPki, chalk.Reset)
			return
		} else if len(capabilities.PaymentDestination) == 0 {
			fmt.Printf("%s%s missing required capability: %s%s\n", chalk.Yellow, chalk.Dim.TextStyle(logPrefix), paymail.CapabilityPaymentDestination, chalk.Reset)
			return
		}

		// Passed the capabilities check
		fmt.Printf("%s%s found required %s and %s capabilities%s\n", chalk.Cyan, chalk.Dim.TextStyle(logPrefix), paymail.CapabilityPki, paymail.CapabilityPaymentDestination, chalk.Reset)

		// Only if we have an address (extra validations)
		if len(paymailAddress) > 0 {

			// Get the alias of the address
			parts := strings.Split(paymailAddress, "@")

			// Get the PKI for the given address
			fmt.Printf("%s getting PKI...\n", chalk.Dim.TextStyle(logPrefix))
			var pki *paymail.PKIResponse
			if pki, err = paymail.GetPKI(capabilities.Pki, parts[0], domain); err != nil {
				fmt.Printf("%s%s get PKI failed: %s%s\n", chalk.Magenta, logPrefix, err.Error(), chalk.Reset)
				return
			}

			// Check the version
			if pki.BsvAlias != bsvAliasVersion {
				fmt.Printf("%s%s pki bsvalias version mismatch, expected: %s but got: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), bsvAliasVersion, capabilities.BsvAlias, chalk.Reset)
				return
			}

			// Found the paymail address
			fmt.Printf("%s%s fetching PKI was successful - found PubKey: %s%s\n", chalk.Green, chalk.Dim.TextStyle(logPrefix), pki.PubKey, chalk.Reset)
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
