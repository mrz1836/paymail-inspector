package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/spf13/cobra"
	"github.com/tonicpow/go-paymail"
	"github.com/ttacon/chalk"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies if a paymail is associated to a pubkey",
	Long: chalk.Green.NewStyle().WithTextStyle(chalk.Bold).Style(`
                   .__  _____       
___  __ ___________|__|/ ____\__.__.
\  \/ // __ \_  __ \  \   __<   |  |
 \   /\  ___/|  | \/  ||  |  \___  |
  \_/  \___  >__|  |__||__|  / ____|
           \/                \/`) + `
` + chalk.Yellow.Color(`
Verify will check the paymail address against a given pubkey using the provider domain (if capability is supported).

This capability allows clients to verify if a given public key is a valid identity key for a given paymail handle.

The public key returned by pki flow for a given paymail handle may change over time. 
This situation may produce troubles to verify data signed using old keys, because even having the keys, 
the verifier doesn't know if the public key actually belongs to the right user.

Read more at: `+chalk.Cyan.Color("http://bsvalias.org/05-verify-public-key-owner.html")),
	Aliases:    []string{"verification"},
	SuggestFor: []string{"pubkey"},
	Example: applicationName + " verify mrz@" + defaultDomainName + " 02ead23149a1e33df17325ec7a7ba9e0b20c674c57c630f527d69b866aa9b65b10" +
		"\n" + applicationName + " verify 1mrz 0352530c305378fd9dfd99f8c8c44e9092efa7c1674b61d4e9be65f92aa7a77bbe",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return chalker.Error("verify requires a paymail address AND pubkey")
		} else if len(args) > 2 {
			return chalker.Error("verify only supports one address and one pubkey at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		var paymailAddress, alias, domain, pubKey string

		// Convert handle if detected
		if len(args[0]) < 25 {
			args[0] = paymail.ConvertHandle(args[0], false)
		} else if len(args[1]) < 25 {
			args[1] = paymail.ConvertHandle(args[1], false)
		}

		// Check for paymail in both args
		if strings.Contains(args[0], "@") {
			alias, domain, paymailAddress = paymail.SanitizePaymail(args[0])
			pubKey = args[1]
		} else if strings.Contains(args[1], "@") {
			pubKey = args[0]
			alias, domain, paymailAddress = paymail.SanitizePaymail(args[1])
		}

		// Require a paymail address
		if len(paymailAddress) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("One argument must be a paymail address [%s] [%s]", args[0], args[1]))
			return
		}

		// Require a pubkey
		if len(pubKey) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("One argument must be a pubkey [%s] [%s]", args[0], args[1]))
			return
		}

		// Validate the paymail address and domain (error already shown)
		if ok := validatePaymailAndDomain(paymailAddress, domain); !ok {
			return
		}

		// Validate pubkey
		if len(pubKey) != paymail.PubKeyLength {
			chalker.Log(chalker.ERROR, fmt.Sprintf("PubKey is an invalid length, expected: %d but got: %d", paymail.PubKeyLength, len(pubKey)))
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
		verifyURL := capabilities.GetString(paymail.BRFCVerifyPublicKeyOwner, "")
		if len(verifyURL) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", domain, paymail.BRFCVerifyPublicKeyOwner))
			return
		}

		// Fire the verify request
		var verify *paymail.Verification
		if verify, err = verifyPubKey(verifyURL, alias, domain, pubKey); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("verify pubkey request failed: %s", err.Error()))
			return
		}

		// Rendering profile information
		displayHeader(chalker.BOLD, fmt.Sprintf("Rendering verify response for %s...", chalk.Cyan.Color(paymailAddress)))

		// Show the results
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Paymail: %s", chalk.Cyan.Color(paymailAddress)))
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("PubKey : %s", chalk.Cyan.Color(pubKey)))

		if verify.Match {
			chalker.Log(chalker.SUCCESS, "Paymail & PubKey Match! (service responded: match=true)")
		} else {
			chalker.Log(chalker.ERROR, "DO NOT MATCH! (service responded: match=false)")
		}
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
