package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
)

// capabilitiesCmd represents the capabilities command
var capabilitiesCmd = &cobra.Command{
	Use:     "capabilities",
	Short:   "Get the capabilities of the paymail domain",
	Long:    `This command will return the capabilities for a given paymail domain`,
	Aliases: []string{"abilities"},
	Example: "capabilities " + defaultDomainName,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("%s requires either a domain or paymail address\n", logPrefix)
		} else if len(args) > 1 {
			return fmt.Errorf("%s validate only supports one domain or address at a time\n", logPrefix)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		domain, _ := paymail.ExtractParts(args[0])

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			fmt.Printf("%s domain name is invalid: %s\n", logPrefix, domain)
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			fmt.Printf("%s domain name failed DNS check: %s\n", logPrefix, domain)
			return
		}

		// Get the details from the SRV record
		fmt.Printf("%s getting SRV record...\n", logPrefix)
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			fmt.Printf("%s get SRV record failed: %s\n", logPrefix, err.Error())
			return
		}

		// Get the capabilities for the given domain
		fmt.Printf("%s getting capabilities...\n", logPrefix)
		var capabilities *paymail.CapabilitiesResponse
		capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port))
		if err != nil {
			fmt.Printf("%s get capabilities failed: %s\n", logPrefix, err.Error())
			return
		}

		// Check the version
		if capabilities.BsvAlias != bsvAliasVersion {
			fmt.Printf("%s capabilities bsvalias version mismatch, expected: %s but got: %s\n", logPrefix, bsvAliasVersion, capabilities.BsvAlias)
			return
		}

		// Show basic results
		fmt.Printf("%s bsvalias version: %s\n", logPrefix, capabilities.BsvAlias)
		fmt.Printf("%s capabilities found: %d\n", logPrefix, len(capabilities.Capabilities))

		// Show all the found capabilities
		fmt.Printf("%s capabilities activated:\n", logPrefix)
		for key, val := range capabilities.Capabilities {
			valType := reflect.TypeOf(val).String()
			if valType == "string" {
				fmt.Printf("%s capability: %s target: %s\n", logPrefix, key, val)
			} else if valType == "bool" { // See: http://bsvalias.org/04-02-sender-validation.html
				if val.(bool) {
					fmt.Printf("%s capability: %s is enabled\n", logPrefix, key)
				} else {
					fmt.Printf("%s capability: %s is disabled\n", logPrefix, key)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)

	// Custom port for the SRV record (target address)
	capabilitiesCmd.Flags().IntVarP(&port, "port", "p", paymail.DefaultPort, "Port that is found in the SRV record")
}
