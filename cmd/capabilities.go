package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mrz1836/go-validate"
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
			return fmt.Errorf("%s%s requires either a domain or paymail address%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), chalk.Reset)
		} else if len(args) > 1 {
			return fmt.Errorf("%s%s validate only supports one domain or address at a time%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), chalk.Reset)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		domain, _ := paymail.ExtractParts(args[0])

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			fmt.Printf("%s%s domain name is invalid: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), domain, chalk.Reset)
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			fmt.Printf("%s%s domain name failed DNS check: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), domain, chalk.Reset)
			return
		}

		// Get the details from the SRV record
		fmt.Printf("%s getting SRV record...\n", chalk.Dim.TextStyle(logPrefix))
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			fmt.Printf("%s%s get SRV record failed: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), err.Error(), chalk.Reset)
			return
		}

		// Get the capabilities for the given domain
		fmt.Printf("%s getting capabilities...\n", chalk.Dim.TextStyle(logPrefix))
		var capabilities *paymail.CapabilitiesResponse
		capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port))
		if err != nil {
			fmt.Printf("%s%s get capabilities failed: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), err.Error(), chalk.Reset)
			return
		}

		// Check the version
		if capabilities.BsvAlias != bsvAliasVersion {
			fmt.Printf("%s%s capabilities bsvalias version mismatch, expected: %s but got: %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), bsvAliasVersion, capabilities.BsvAlias, chalk.Reset)
			return
		}

		// Show basic results
		fmt.Printf("%s%s bsvalias version: %s%s\n", chalk.Cyan, chalk.Dim.TextStyle(logPrefix), capabilities.BsvAlias, chalk.Reset)
		fmt.Printf("%s%s %d capabilities found: %s\n", chalk.Green.NewStyle().WithTextStyle(chalk.Bold), chalk.Dim.TextStyle(logPrefix), len(capabilities.Capabilities), chalk.Reset)

		// Show all the found capabilities
		for key, val := range capabilities.Capabilities {
			valType := reflect.TypeOf(val).String()
			if valType == "string" {
				fmt.Printf("%s %s: %-28v %s: %s\n", chalk.Dim.TextStyle(logPrefix), chalk.White.Color("capability"), chalk.Cyan.Color(key), chalk.White.Color("target"), chalk.Yellow.Color(fmt.Sprintf("%s", val)))
			} else if valType == "bool" { // See: http://bsvalias.org/04-02-sender-validation.html
				if val.(bool) {
					fmt.Printf("%s %s: %-28v is      %s\n", chalk.Dim.TextStyle(logPrefix), chalk.White.Color("capability"), chalk.Cyan.Color(key), chalk.Green.Color("enabled"))
				} else {
					fmt.Printf("%s %s: %-28v is      %s\n", chalk.Dim.TextStyle(logPrefix), chalk.White.Color("capability"), chalk.Cyan.Color(key), chalk.Magenta.Color("disabled"))
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
}
