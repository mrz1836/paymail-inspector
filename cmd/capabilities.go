package cmd

import (
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

// capabilitiesCmd represents the capabilities command
var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Get the capabilities of the paymail domain",
	Long: chalk.Green.Color(`
                          ___.   .__.__  .__  __  .__               
  ____ _____  ___________ \_ |__ |__|  | |__|/  |_|__| ____   ______
_/ ___\\__  \ \____ \__  \ | __ \|  |  | |  \   __\  |/ __ \ /  ___/
\  \___ / __ \|  |_> > __ \| \_\ \  |  |_|  ||  | |  \  ___/ \___ \ 
 \___  >____  /   __(____  /___  /__|____/__||__| |__|\___  >____  >
     \/     \/|__|       \/    \/                         \/     \/`) + `
` + chalk.Yellow.Color(`
This command will return the capabilities for a given paymail domain.

Capability Discovery is the process by which a paymail client learns the supported 
features of a paymail service and their respective endpoints and configurations.

Drawing inspiration from RFC 5785 and IANA's Well-Known URIs resource, the Capability Discovery protocol 
dictates that a machine-readable document is placed in a predictable location on a web server.

Read more at: `+chalk.Cyan.Color("http://bsvalias.org/02-02-capability-discovery.html")),
	Aliases: []string{"c", "abilities", "inspect", "lookup"},
	Example: configDefault + " capabilities " + defaultDomainName,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("capabilities requires either a domain or paymail address")
		} else if len(args) > 1 {
			return chalker.Error("capabilities only supports one domain or address at a time")
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

		// Get the capabilities
		capabilities, err := getCapabilities(domain)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				chalker.Log(chalker.WARN, fmt.Sprintf("no capabilities found for: %s", domain))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error: %s", err.Error()))
			}
			return
		}

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

// getCapabilities will check SRV first, then attempt default domain:port check
func getCapabilities(domain string) (capabilities *paymail.CapabilitiesResponse, err error) {

	chalker.Log(chalker.DEFAULT, fmt.Sprintf("getting SRV record for: %s...", chalk.Cyan.Color(domain)))

	// Get the details from the SRV record
	capabilityDomain := ""
	capabilityPort := paymail.DefaultPort
	var srv *net.SRV
	if srv, err = paymail.GetSRVRecord(serviceName, protocol, domain, nameServer); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("get SRV record failed: %s", err.Error()))
		capabilityDomain = domain
	} else if srv != nil {
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("SRV record found, target: %s:%d", srv.Target, srv.Port))
		capabilityDomain = srv.Target
		capabilityPort = int(srv.Port)
	}

	// Get the capabilities for the given target domain
	chalker.Log(chalker.DEFAULT, fmt.Sprintf("getting capabilities from: %s...", chalk.Cyan.Color(fmt.Sprintf("%s:%d", capabilityDomain, capabilityPort))))
	if capabilities, err = paymail.GetCapabilities(capabilityDomain, capabilityPort); err != nil {
		return
	}

	// Check the version
	if capabilities.BsvAlias != viper.GetString(flagBsvAlias) {
		err = fmt.Errorf("capabilities %s version mismatch, expected: %s but got: %s", flagBsvAlias, chalk.Cyan.Color(viper.GetString(flagBsvAlias)), chalk.Magenta.Color(capabilities.BsvAlias))
		return
	}

	// Success
	chalker.Log(chalker.SUCCESS, fmt.Sprintf("capabilities found: %d", len(capabilities.Capabilities)))
	chalker.Log(chalker.DEFAULT, fmt.Sprintf("%s version: %s", flagBsvAlias, chalk.Cyan.Color(capabilities.BsvAlias)))

	return
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
}
