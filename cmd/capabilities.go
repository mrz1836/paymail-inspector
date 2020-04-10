package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
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
			return chalker.Error("requires either a domain or paymail address")
		} else if len(args) > 1 {
			return chalker.Error("validate only supports one domain or address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		domain, _ := paymail.ExtractParts(args[0])

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			chalker.Log(chalker.ERROR, fmt.Sprintf("domain name is invalid: %s", domain))
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			chalker.Log(chalker.ERROR, fmt.Sprintf("domain name failed DNS check: %s", domain))
			return
		}

		// Get the details from the SRV record
		chalker.Log(chalker.DEFAULT, "getting SRV record...")

		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("get SRV record failed: %s", err.Error()))
			return
		}

		// Get the capabilities for the given domain
		chalker.Log(chalker.DEFAULT, "getting capabilities...")
		var capabilities *paymail.CapabilitiesResponse
		capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port))
		if err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("get capabilities failed: %s", err.Error()))
			return
		}

		// Check the version
		if capabilities.BsvAlias != bsvAliasVersion {
			chalker.Log(chalker.ERROR, fmt.Sprintf("capabilities bsvalias version mismatch, expected: %s but got: %s", bsvAliasVersion, capabilities.BsvAlias))
			return
		}

		// Show basic results
		chalker.Log(chalker.INFO, fmt.Sprintf("bsvalias version: %s", capabilities.BsvAlias))
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("%d capabilities found:", len(capabilities.Capabilities)))

		// Show all the found capabilities
		for key, val := range capabilities.Capabilities {
			valType := reflect.TypeOf(val).String()
			if valType == "string" {
				chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v %s: %s", chalk.White.Color("capability"), chalk.Cyan.Color(key), chalk.White.Color("target"), chalk.Yellow.Color(fmt.Sprintf("%s", val))))
			} else if valType == "bool" { // See: http://bsvalias.org/04-02-sender-validation.html
				if val.(bool) {
					chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v is      %s", chalk.White.Color("capability"), chalk.Cyan.Color(key), chalk.Green.Color("enabled")))
				} else {
					chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v is      %s", chalk.White.Color("capability"), chalk.Cyan.Color(key), chalk.Magenta.Color("disabled")))
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
}
