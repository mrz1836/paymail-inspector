package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

// whoisCmd represents the whois command
var whoisCmd = &cobra.Command{
	Use:        "whois",
	Short:      "Find a paymail handle across several providers",
	Aliases:    []string{"who", "w"},
	SuggestFor: []string{"lookup"},
	Example: applicationName + ` whois mrz
` + applicationName + ` w mrz`,
	Long: chalk.Green.Color(`
        .__           .__        
__  _  _|  |__   ____ |__| ______
\ \/ \/ /  |  \ /  _ \|  |/  ___/
 \     /|   Y  (  <_> )  |\___ \ 
  \/\_/ |___|  /\____/|__/____  >
             \/               \/`) + `
` + chalk.Yellow.Color(`

Search `+strconv.Itoa(len(providers))+` public paymail providers for a handle.`),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("whois requires a handle")
		} else if len(args) > 1 {
			return chalker.Error("whois only supports 1 handle at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Handle to search
		handle := ""

		// Detect if handler or not
		if strings.Contains(args[0], "@") {
			parts := strings.Split(strings.TrimSpace(args[0]), "@")
			handle = parts[0]
		} else {
			handle = strings.TrimSpace(args[0])
		}

		// Invalid handle?
		if len(handle) == 0 || len(handle) > 255 {
			chalker.Log(chalker.ERROR, "Handle is invalid")
			return
		}

		// List of paymails found
		var paymails []*PaymailDetails

		// Loop each provider
		for _, provider := range providers {

			// Get the capabilities
			capabilities, err := getCapabilities(provider.Domain, true)
			if err != nil {
				if strings.Contains(err.Error(), "context deadline exceeded") {
					chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", provider.Domain))
				} else {
					chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
				}
				continue
			}

			// Set the URL - Does the paymail provider have the capability?
			pkiUrl := capabilities.GetValueString(paymail.BRFCPki, paymail.BRFCPkiAlternate)
			if len(pkiUrl) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", provider.Domain, paymail.BRFCPki))
				continue
			}

			// Create result
			result := &PaymailDetails{
				Handle:   handle,
				Provider: provider,
			}

			// Get the PKI for the given address
			if result.PKI, err = getPki(pkiUrl, handle, provider.Domain, true); err != nil || result.PKI == nil {
				if err != nil {
					chalker.Log(chalker.ERROR, fmt.Sprintf("Search response: %s", err.Error()))
				}
				paymails = append(paymails, result)
				continue
			}

			// Get all the public info
			if err = result.GetPublicInfo(capabilities); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
			}

			// Add to list
			paymails = append(paymails, result)
		}

		// If we don't have results
		if len(paymails) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", "no paymail results returned"))
			return
		}

		// Show the results header
		displayHeader(chalker.BOLD, fmt.Sprintf("Whois results from %d providers...", len(providers)))

		// Loop results
		for _, result := range paymails {
			result.Display()
		}
	},
}

func init() {
	rootCmd.AddCommand(whoisCmd)

	// todo: flag for custom provider (not in the list)
}
