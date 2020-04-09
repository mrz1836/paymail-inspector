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
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			fmt.Printf("%s get SRV record failed: %s\n", logPrefix, err.Error())
			return
		}

		// Get the capabilities for the given domain
		var resp *paymail.CapabilitiesResponse
		resp, err = paymail.GetCapabilities(srv.Target, int(srv.Port))
		if err != nil {
			fmt.Printf("%s get capabilities failed: %s\n", logPrefix, err.Error())
			return
		}

		// Show basic results
		fmt.Printf("%s BsvAlias version: %s\n", logPrefix, resp.BsvAlias)
		fmt.Printf("%s capabilities found: %d\n", logPrefix, len(resp.Capabilities))

		// Show all the found capabilities
		fmt.Printf("%s capabilities activated:\n", logPrefix)
		for key, val := range resp.Capabilities {
			if reflect.TypeOf(val).String() == "string" {
				fmt.Printf("%s capability: %s target: %s\n", logPrefix, key, val)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)

	// Custom port for the SRV record (target address)
	capabilitiesCmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Port that is found in the SRV record")
}
