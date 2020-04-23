package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/mrz1836/paymail-inspector/roundesk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolves a paymail address",
	Long: chalk.Green.Color(`
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
	SuggestFor: []string{"address", "destination", "payment", "addressing", "whois"},
	Example:    applicationName + " resolve mrz@" + defaultDomainName,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("resolve requires either a paymail address")
		} else if len(args) > 1 {
			return chalker.Error("resolve only supports one address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		var senderDomain string
		var senderHandle string
		senderDomain, senderHandle = paymail.ExtractParts(viper.GetString(flagSenderHandle))
		domain, paymailAddress := paymail.ExtractParts(args[0])

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
		if len(senderHandle) == 0 {
			chalker.Log(chalker.WARN, fmt.Sprintf("The flag --%s is not set, using default: %s", flagSenderHandle, paymailAddress))
			senderHandle = paymailAddress
			senderDomain, senderHandle = paymail.ExtractParts(senderHandle)
		} else { // Sender handle is set (basic validation)

			// Validate the paymail address and domain (error already shown)
			if ok := validatePaymailAndDomain(senderHandle, senderDomain); !ok {
				return
			}
		}

		// Get the capabilities
		capabilities, err := getCapabilities(domain)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", domain))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
			}
			return
		}

		// Set the URL - Does the paymail provider have the capability?
		pkiUrl := capabilities.GetValueString(paymail.BRFCPki, paymail.BRFCPkiAlternate)
		if len(pkiUrl) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", domain, paymail.BRFCPki))
			return
		}

		// Set the URL - Does the paymail provider have the capability?
		resolveUrl := capabilities.GetValueString(paymail.BRFCPaymentDestination, paymail.BRFCBasicAddressResolution)
		if len(resolveUrl) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", domain, paymail.BRFCPaymentDestination))
			return
		}

		// Does this provider require sender validation?
		// https://bsvalias.org/04-02-sender-validation.html
		if capabilities.GetValueBool(paymail.BRFCSenderValidation, "") {
			chalker.Log(chalker.WARN, "Sender validation is ENFORCED")

			// Start the request
			displayHeader(chalker.DEFAULT, fmt.Sprintf("Running sender validations for %s...", chalk.Cyan.Color(senderHandle)))

			// Required if flag is enforced
			if len(signature) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Missing required flag: %s - see the help section: -h", "--signature"))

				// todo: generate a real signature if possible
				chalker.Log(chalker.WARN, fmt.Sprintf("Attempting to fake a signature for: %s...", senderHandle))
				signature, _ = RandomHex(64)
			}

			// Only if it's not the same (set from above ^^)
			if senderHandle != paymailAddress {

				// Get the capabilities
				senderCapabilities, getErr := getCapabilities(senderDomain)
				if getErr != nil {
					if strings.Contains(getErr.Error(), "context deadline exceeded") {
						chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", domain))
					} else {
						chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", getErr.Error()))
					}
					return
				}

				// Set the URL - Does the paymail provider have the capability?
				senderPkiUrl := senderCapabilities.GetValueString(paymail.BRFCPki, paymail.BRFCPkiAlternate)
				if len(senderPkiUrl) == 0 {
					chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", senderDomain, paymail.BRFCPki))
					return
				}

				// Get the alias of the address
				parts := strings.Split(senderHandle, "@")

				// Get the PKI for the given address
				var senderPki *paymail.PKIResponse
				if senderPki, err = getPki(senderPkiUrl, parts[0], parts[1]); err != nil {
					chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
					return
				} else if senderPki != nil {
					chalker.Log(chalker.INFO, fmt.Sprintf("Found --%s %s@%s's pubkey: %s", flagSenderHandle, parts[0], parts[1], chalk.Cyan.Color(senderPki.PubKey)))
				}
			}

			// once completed, the full sender validation will be complete
			chalker.Log(chalker.SUCCESS, `Sender pre-validation: Passed ¯\_(ツ)_/¯`)
		}

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Get the PKI for the given address
		var pki *paymail.PKIResponse
		if pki, err = getPki(pkiUrl, parts[0], domain); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
			return
		}

		// Attempt to resolve the address
		var addressResolution *paymail.AddressResolutionResponse
		if addressResolution, err = resolveAddress(resolveUrl, parts[0], domain, senderHandle, signature, purpose, amount); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Address resolution failed: %s", err.Error()))
			return
		}

		// Attempt to get a public profile if the capability is found
		url := capabilities.GetValueString(paymail.BRFCPublicProfile, "")
		var profile *paymail.PublicProfileResponse
		if len(url) > 0 && !skipPublicProfile {
			if profile, err = getPublicProfile(url, parts[0], domain); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Get public profile failed: %s", err.Error()))
			}
		}

		// Attempt to get a bitpic (if enabled)
		var bitPicURL string
		if !skipPublicProfile {
			if bitPicURL, err = getBitPic(parts[0], domain); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Checking for bitpic failed: %s", err.Error()))
			}
		}

		// Attempt to get a Roundesk profile (if enabled)
		if !skipRoundesk {
			var roundesk *roundesk.Response
			if roundesk, err = getRoundeskProfile(parts[0], domain); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Checking for roundesk profile failed: %s", err.Error()))
			}

			// Display the roundesk profile if found
			if roundesk != nil && roundesk.Profile != nil {

				// Rendering profile information
				displayHeader(chalker.DEFAULT, fmt.Sprintf("Roundesk profile for %s", chalk.Cyan.Color(paymailAddress)))

				if len(roundesk.Profile.Name) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("Name      : %s", chalk.Cyan.Color(roundesk.Profile.Name)))
				}
				if len(roundesk.Profile.Headline) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("Headline  : %s", chalk.Cyan.Color(roundesk.Profile.Headline)))
				}
				if len(roundesk.Profile.Bio) > 0 {
					roundesk.Profile.Bio = strings.TrimSuffix(roundesk.Profile.Bio, "\n")
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("Bio       : %s", chalk.Cyan.Color(roundesk.Profile.Bio)))
				}
				if len(roundesk.Profile.Twetch) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("Twetch    : %s", chalk.Cyan.Color("https://twetch.app/u/"+roundesk.Profile.Twetch)))
				}

				chalker.Log(chalker.DEFAULT, fmt.Sprintf("URL       : %s", chalk.Cyan.Color("https://roundesk.co/u/"+parts[0]+"@"+domain)))

				if len(roundesk.Profile.Nonce) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("Nonce     : %s", chalk.Cyan.Color(roundesk.Profile.Nonce)))
				}
			}
		}

		// Rendering profile information
		displayHeader(chalker.DEFAULT, fmt.Sprintf("Public profile for %s", chalk.Cyan.Color(paymailAddress)))

		// Display the public profile if found
		if profile != nil {
			if len(profile.Name) > 0 {
				chalker.Log(chalker.DEFAULT, fmt.Sprintf("Name         : %s", chalk.Cyan.Color(profile.Name)))
			}
			if len(profile.Avatar) > 0 {
				chalker.Log(chalker.DEFAULT, fmt.Sprintf("Avatar       : %s", chalk.Cyan.Color(profile.Avatar)))
			}
		}

		// Display bitpic if found
		if len(bitPicURL) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Bitpic       : %s", chalk.Cyan.Color(bitPicURL)))
		}

		// Show pubkey
		if pki != nil && len(pki.PubKey) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("PubKey       : %s", chalk.Cyan.Color(pki.PubKey)))
		}

		// Show output script
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Output Script: %s", chalk.Cyan.Color(addressResolution.Output)))

		// Show the resolved address from the output script
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Address      : %s", chalk.Cyan.Color(addressResolution.Address)))

		// If we have a signature
		if len(addressResolution.Signature) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Signature    : %s", chalk.Cyan.Color(addressResolution.Signature)))
		}
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
}
