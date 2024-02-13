package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/mrz1836/go-sanitize"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/spf13/cobra"
	"github.com/tonicpow/go-paymail"
)

// capabilitiesCmd represents the capabilities command
var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Get the capabilities of the paymail domain",
	Long: color.GreenString(`
                          ___.   .__.__  .__  __  .__               
  ____ _____  ___________ \_ |__ |__|  | |__|/  |_|__| ____   ______
_/ ___\\__  \ \____ \__  \ | __ \|  |  | |  \   __\  |/ __ \ /  ___/
\  \___ / __ \|  |_> > __ \| \_\ \  |  |_|  ||  | |  \  ___/ \___ \ 
 \___  >____  /   __(____  /___  /__|____/__||__| |__|\___  >____  >
     \/     \/|__|       \/    \/                         \/     \/`) + `
` + color.YellowString(`
This command will return the capabilities for a given paymail domain.

Capability Discovery is the process by which a paymail client learns the supported 
features of a paymail service and their respective endpoints and configurations.

Drawing inspiration from RFC 5785 and IANA's Well-Known URIs resource, the Capability Discovery protocol 
dictates that a machine-readable document is placed in a predictable location on a web server.

Read more at: `+color.CyanString("http://bsvalias.org/02-02-capability-discovery.html")),
	Aliases: []string{"c", "inspect"},
	Example: applicationName + " capabilities " + defaultDomainName + `
` + applicationName + " c " + defaultDomainName,
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("capabilities requires either a domain or paymail address")
		} else if len(args) > 1 {
			return chalker.Error("capabilities only supports one domain or address at a time")
		}
		return nil
	},
	Run: func(_ *cobra.Command, args []string) {

		// Sanitize the domain
		domain, _ := sanitize.Domain(args[0], false, true)

		// Validate the domain
		err := paymail.ValidateDomain(args[0])
		if err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Domain name %s is invalid: %s", domain, err.Error()))
			return
		}

		// Get the capabilities
		var capabilities *paymail.CapabilitiesResponse
		capabilities, err = getCapabilities(domain, false)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", domain))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
			}
			return
		}

		// Rendering profile information
		displayHeader(chalker.BOLD, fmt.Sprintf("Listing %d capabilities...", len(capabilities.Capabilities)))

		// Show all the found capabilities
		// todo: loop known BRFCs and display "more" info in this display for all detected BRFCs
		for key, val := range capabilities.Capabilities {
			valType := reflect.TypeOf(val).String()
			if valType == "string" {
				chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v %s: %s", color.WhiteString("Capability"), color.CyanString(key), color.WhiteString("Target"), color.YellowString(fmt.Sprintf("%s", val))))
			} else if valType == "bool" { // See: http://bsvalias.org/04-02-sender-validation.html
				if val.(bool) {
					chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v Is    : %s", color.WhiteString("Capability"), color.CyanString(key), color.GreenString("Enabled")))
				} else {
					chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v Is    : %s", color.WhiteString("Capability"), color.CyanString(key), color.MagentaString("Disabled")))
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
}
