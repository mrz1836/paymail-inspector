package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tonicpow/go-paymail"
	"github.com/ttacon/chalk"
)

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolves a paymail address",
	Long: chalk.Green.NewStyle().WithTextStyle(chalk.Bold).Style(`
                            .__               
_______   ____   __________ |  |___  __ ____  
\_  __ \_/ __ \ /  ___/  _ \|  |\  \/ // __ \ 
 |  | \/\  ___/ \___ (  <_> )  |_\   /\  ___/ 
 |__|    \___  >____  >____/|____/\_/  \___  >
             \/     \/                     \/`) + `
` + chalk.Yellow.Color(`
Resolves a paymail address into a hex-encoded Bitcoin script, address and public profile (if found).

Given a sender and a receiver, where the sender knows the receiver's 
paymail handle <alias>@<domain>.<tld>, the sender can perform Service Discovery against 
the receiver and request a payment destination from the receiver's paymail service.

Read more at: `+chalk.Cyan.Color("http://bsvalias.org/04-01-basic-address-resolution.html")),
	Aliases:    []string{"r", "resolution"},
	SuggestFor: []string{"address", "destination", "payment", "addressing"},
	Example: applicationName + " resolve mrz@" + defaultDomainName + `
` + applicationName + " r mrz@" + defaultDomainName + `
` + applicationName + " r 1mrz",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("resolve requires either a paymail address")
		} else if len(args) > 1 {
			return chalker.Error("resolve only supports one address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract sender parts
		senderAlias, senderDomain, senderAddress := paymail.SanitizePaymail(viper.GetString(flagSenderHandle))

		// Extract paymail parts
		handle, domain, paymailAddress := paymail.SanitizePaymail(paymail.ConvertHandle(args[0], false))

		// Did we get a paymail address?
		if len(paymailAddress) == 0 {
			chalker.Log(chalker.ERROR, "Paymail address not found or invalid")
			return
		}

		// Validate the paymail address and domain (error already shown)
		if ok := validatePaymailAndDomain(paymailAddress, domain); !ok {
			return
		}

		// No sender handle given? (default: set to the receiver's paymail address)
		if len(senderAddress) == 0 {
			chalker.Log(chalker.WARN, fmt.Sprintf("The flag --%s is not set, using default: %s", flagSenderHandle, paymailAddress))
			senderAddress = paymailAddress
			senderAlias, senderDomain, senderAddress = paymail.SanitizePaymail(senderAddress)
		} else { // Sender handle is set (basic validation)

			// Validate the paymail address and domain (error already shown)
			if ok := validatePaymailAndDomain(senderAddress, senderDomain); !ok {
				return
			}
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
		pkiURL := capabilities.GetString(paymail.BRFCPki, paymail.BRFCPkiAlternate)
		if len(pkiURL) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", domain, paymail.BRFCPki))
			return
		}

		// Set the URL - Does the paymail provider have the capability?
		resolveURL := capabilities.GetString(paymail.BRFCPaymentDestination, paymail.BRFCBasicAddressResolution)
		if len(resolveURL) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", domain, paymail.BRFCPaymentDestination))
			return
		}

		// Does this provider require sender validation?
		// https://bsvalias.org/04-02-sender-validation.html
		if capabilities.GetBool(paymail.BRFCSenderValidation, "") {
			chalker.Log(chalker.WARN, "Sender validation is ENFORCED")

			// Start the request
			displayHeader(chalker.DEFAULT, fmt.Sprintf("Running sender validations for %s...", chalk.Cyan.Color(senderAddress)))

			// Required if flag is enforced
			if len(signature) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Missing required flag: %s - see the help section: -h", "--signature"))

				// todo: generate a real signature if possible
				chalker.Log(chalker.WARN, fmt.Sprintf("Attempting to fake a signature for: %s...", senderAddress))
				signature, _ = RandomHex(64)
			}

			// Only if it's not the same (set from above ^^)
			if senderAddress != paymailAddress {

				// Get the capabilities
				senderCapabilities, getErr := getCapabilities(senderDomain, true)
				if getErr != nil {
					if strings.Contains(getErr.Error(), "context deadline exceeded") {
						chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", domain))
					} else {
						chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", getErr.Error()))
					}
					return
				}

				// Set the URL - Does the paymail provider have the capability?
				senderPkiURL := senderCapabilities.GetString(paymail.BRFCPki, paymail.BRFCPkiAlternate)
				if len(senderPkiURL) == 0 {
					chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", senderDomain, paymail.BRFCPki))
					return
				}

				// Get the PKI for the given address
				var senderPki *paymail.PKI
				if senderPki, err = getPki(senderPkiURL, senderAlias, senderDomain, true); err != nil {
					chalker.Log(chalker.ERROR, fmt.Sprintf("Find PKI Failed: %s", err.Error()))
					return
				} else if senderPki != nil {
					chalker.Log(chalker.INFO, fmt.Sprintf("Found --%s %s@%s's pubkey: %s", flagSenderHandle, senderAlias, senderDomain, chalk.Cyan.Color(senderPki.PubKey)))
				}
			}

			// once completed, the full sender validation will be complete
			chalker.Log(chalker.SUCCESS, `Sender pre-validation: Passed ¯\_(ツ)_/¯`)
		}

		// Set the provider (known vs new provider)
		provider := getProvider(domain)
		if provider == nil {
			provider = &Provider{Domain: domain, Link: "https://" + domain}
		}

		// Create result
		result := &PaymailDetails{Handle: handle, Provider: provider}

		// Get the PKI for the given address
		if result.PKI, err = getPki(pkiURL, handle, domain, true); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Find PKI Failed: %s", err.Error()))
			return
		}

		// Attempt to resolve the address
		if result.Resolution, err = resolveAddress(resolveURL, handle, domain, senderAddress, signature, purpose, amount); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Address resolution failed: %s", err.Error()))
			return
		}

		// Get all the public info
		if err = result.GetPublicInfo(capabilities); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
			// return
		}

		// Show the results
		result.Display()
	},
}

func init() {
	rootCmd.AddCommand(resolveCmd)

	// Set the amount for the sender request
	resolveCmd.Flags().Uint64VarP(&amount, "amount", "a", 0, "Amount in satoshis for the payment request")

	// Set the purpose for the sender request
	resolveCmd.Flags().StringVarP(&purpose, "purpose", "p", "", "Purpose for the transaction")

	// Set the sender's handle for the sender request
	resolveCmd.PersistentFlags().String(flagSenderHandle, "", "Sender's paymail handle. Required by bsvalias spec. Receiver paymail used if not specified.")
	er(viper.BindPFlag(flagSenderHandle, resolveCmd.PersistentFlags().Lookup(flagSenderHandle)))

	// Set the sender's name for the sender request
	resolveCmd.Flags().String(flagSenderName, "", "The sender's name")
	er(viper.BindPFlag(flagSenderName, resolveCmd.PersistentFlags().Lookup(flagSenderHandle)))

	// Set the signature of the entire request
	resolveCmd.Flags().StringVarP(&signature, "signature", "s", "", "The signature of the entire request")

	// Skip getting the PubKey
	resolveCmd.Flags().BoolVar(&skipPki, "skip-pki", false, "Skip the pki request")

	// Skip getting public profile
	resolveCmd.Flags().BoolVar(&skipPublicProfile, "skip-public-profile", false, "Skip the public profile request")

	// Skip getting Bitpic avatar
	resolveCmd.Flags().BoolVar(&skipBitpic, "skip-bitpic", false, "Skip trying to get an associated Bitpic")

	// Skip getting Roundesk profile
	resolveCmd.Flags().BoolVar(&skipRoundesk, "skip-roundesk", false, "Skip trying to get an associated Roundesk profile")

	// Skip getting Baemail account
	resolveCmd.Flags().BoolVar(&skipBaemail, "skip-baemail", false, "Skip trying to get an associated Baemail account")

	// Skip getting PowPing account
	resolveCmd.Flags().BoolVar(&skipPowPing, "skip-powping", false, "Skip trying to get an associated PowPing account")
}
