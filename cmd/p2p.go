package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/spf13/cobra"
	"github.com/tonicpow/go-paymail"
)

const (
	defaultSatoshiValue = 1000
)

// p2pCmd represents the p2p command
var p2pCmd = &cobra.Command{
	Use:   "p2p",
	Short: "Starts a new P2P payment request",
	Long: color.GreenString(`
       ________         
______ \_____  \______  
\____ \ /  ____/\____ \ 
|  |_> >       \|  |_> >
|   __/\_______ \   __/ 
|__|           \/__|`) + `
` + color.YellowString(`
This command will start a new P2P request with the receiver and optional amount expected (in Satoshis).

This protocol is an alternative protocol to basic address resolution. 
Instead of returning one address, it returns a list of outputs with a reference number. 
It is only intended to be used with P2P Transactions and will continue to function even 
after basic address resolution is deprecated.

Read more at: `+color.CyanString("https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html")),
	Aliases:    []string{"send"},
	SuggestFor: []string{"sending"},
	Example: applicationName + " p2p mrz@" + defaultDomainName + `
` + applicationName + ` p2p \$mr-z`,
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("p2p requires a paymail address")
		} else if len(args) > 1 {
			return chalker.Error("p2p only supports one address at a time")
		}
		return nil
	},
	Run: func(_ *cobra.Command, args []string) {

		// Set the domain and paymail
		_, domain, paymailAddress := paymail.SanitizePaymail(paymail.ConvertHandle(args[0], false))

		// Did we get a paymail address?
		if len(paymailAddress) == 0 {
			chalker.Log(chalker.ERROR, "Paymail address not found or invalid")
			return
		}

		// Validate the paymail address and domain (error already shown)
		if ok := validatePaymailAndDomain(paymailAddress, domain); !ok {
			return
		}

		// Get the capabilities
		capabilities, err := getCapabilities(domain, true)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", domain))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
			}
			return
		}

		// Set the URL - Does the paymail provider have the capability?
		destinationURL := capabilities.GetString(paymail.BRFCP2PPaymentDestination, "")
		if len(destinationURL) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", domain, paymail.BRFCP2PPaymentDestination))
			return
		}

		// Set the satoshis
		if satoshis <= 0 {
			satoshis = defaultSatoshiValue
		}

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Fire the P2P request
		var p2pResponse *paymail.PaymentDestinationResponse
		if p2pResponse, err = getP2PPaymentDestination(destinationURL, parts[0], domain, satoshis); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("P2P payment destination request failed: %s", err.Error()))
			return
		}

		// Attempt to get a public profile if the capability is found
		profileURL := capabilities.GetString(paymail.BRFCPublicProfile, "")
		var profile *paymail.PublicProfileResponse
		if len(profileURL) > 0 && !skipPublicProfile {
			if profile, err = getPublicProfile(profileURL, parts[0], domain, true); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Get public profile failed: %s", err.Error()))
			}
		}

		// Attempt to get a bitpic (if enabled)
		var bitPicURL string
		if !skipBitpic {
			if bitPicURL, err = getBitPic(parts[0], domain, true); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Checking for bitpic failed: %s", err.Error()))
			}
		}

		// Rendering profile information
		displayHeader(chalker.BOLD, fmt.Sprintf("P2P information for %s", color.CyanString(paymailAddress)))

		// Display the public profile if found
		if profile != nil {
			if len(profile.Name) > 0 {
				chalker.Log(chalker.DEFAULT, fmt.Sprintf("Name      : %s", color.CyanString(profile.Name)))
			}
			if len(profile.Avatar) > 0 {
				chalker.Log(chalker.DEFAULT, fmt.Sprintf("Avatar    : %s", color.CyanString(profile.Avatar)))
			}
		}

		// Display bitpic if found
		if len(bitPicURL) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Bitpic    : %s", color.CyanString(bitPicURL)))
		}

		// If there is a reference
		if len(p2pResponse.Reference) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Reference : %s", color.CyanString(p2pResponse.Reference)))
		}

		// Output the results
		for index, output := range p2pResponse.Outputs {

			// Show output script & amount
			displayHeader(chalker.DEFAULT, fmt.Sprintf("Output #%d", index+1))
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Script    : %s", color.CyanString(output.Script)))
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Satoshis  : %s", color.CyanString(fmt.Sprintf("%d", output.Satoshis))))
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Address   : %s", color.CyanString(output.Address)))
		}
	},
}

func init() {
	rootCmd.AddCommand(p2pCmd)

	// Set the amount for the sender request
	p2pCmd.Flags().Uint64Var(&satoshis, "satoshis", 0, "Amount in satoshis for the the incoming transaction(s)")
}
