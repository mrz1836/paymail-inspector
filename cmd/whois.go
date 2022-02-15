package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/mrz1836/go-sanitize"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/spf13/cobra"
	"github.com/tonicpow/go-paymail"
)

// whoisCmd represents the whois command
var whoisCmd = &cobra.Command{
	Use:        "whois",
	Short:      "Find a paymail handle across several providers",
	Aliases:    []string{"who", "w"},
	SuggestFor: []string{"lookup"},
	Example: applicationName + ` whois mrz
` + applicationName + ` w mrz
` + applicationName + ` w \$mr-z
` + applicationName + ` w 1mrz`,
	Long: color.GreenString(`
        .__           .__        
__  _  _|  |__   ____ |__| ______
\ \/ \/ /  |  \ /  _ \|  |/  ___/
 \     /|   Y  (  <_> )  |\___ \ 
  \/\_/ |___|  /\____/|__/____  >
             \/               \/`) + `
` + color.YellowString(`

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
		var handle string

		// Are we using a paymail address?
		if strings.Contains(args[0], "@") {
			handle, _, _ = paymail.SanitizePaymail(args[0])
		} else { // Using an alias or $handle
			handle, _, _ = paymail.SanitizePaymail(paymail.ConvertHandle(args[0], false))
		}

		// Sanitize
		handle = sanitize.Custom(handle, `[^a-zA-Z0-9-_.+]`)

		// Invalid handle?
		if len(handle) == 0 || len(handle) > 255 {
			chalker.Log(chalker.ERROR, "Handle is invalid")
			return
		}

		// List of paymails found
		var paymails []*PaymailDetails

		// Loop each provider (break into a Go routine for each provider)
		var wg sync.WaitGroup
		for _, provider := range providers {
			wg.Add(1)
			go fetchPaymailInfo(&wg, handle, provider, &paymails)
		}

		// Waiting for all providers to finish
		wg.Wait()

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

// fetchPaymailInfo will get the paymail information from the provider
func fetchPaymailInfo(wg *sync.WaitGroup, handle string, provider *Provider, results *[]*PaymailDetails) {
	defer wg.Done()

	// Get the capabilities
	capabilities, err := getCapabilities(provider.Domain, true)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", provider.Domain))
		} else {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
		}
		return
	}

	// Set the URL - Does the paymail provider have the capability?
	pkiURL := capabilities.GetString(paymail.BRFCPki, paymail.BRFCPkiAlternate)
	if len(pkiURL) == 0 {
		chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", provider.Domain, paymail.BRFCPki))
		return
	}

	// Create result
	result := &PaymailDetails{
		Handle:   handle,
		Provider: provider,
	}

	// Get the PKI for the given address
	if result.PKI, err = getPki(pkiURL, handle, provider.Domain, true); err != nil || result.PKI == nil {
		if err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Search response: %s", err.Error()))
		}
		*results = append(*results, result)
		return
	}

	// Get all the public info
	if err = result.GetPublicInfo(capabilities); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
	}

	// Add to list
	*results = append(*results, result)
}

func init() {
	rootCmd.AddCommand(whoisCmd)

	// todo: flag for custom provider (not in the list)
}
