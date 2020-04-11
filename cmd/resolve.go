package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

// Flags for the resolve command
var (
	amount            uint64
	purpose           string
	senderHandle      string
	senderName        string
	signature         string
	skipPki           bool
	skipPublicProfile bool
)

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:        "resolve",
	Short:      "Resolves a paymail address",
	Long:       `Resolves a paymail address into a hex-encoded Bitcoin script and address`,
	Aliases:    []string{"r", "resolution"},
	SuggestFor: []string{"address", "destination", "payment", "addressing"},
	Example:    "resolve this@address.com",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("%s requires either a paymail address")
		} else if len(args) > 1 {
			return chalker.Error("resolve only supports one address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		var senderDomain string
		senderDomain, senderHandle = paymail.ExtractParts(senderHandle)
		domain, paymailAddress := paymail.ExtractParts(args[0])

		// Did we get a paymail address?
		if len(paymailAddress) == 0 {
			chalker.Log(chalker.ERROR, "paymail address not found or invalid")
			return
		}

		// Validate the format for the paymail address (paymail addresses follow conventional email requirements)
		if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("paymail address failed format validation: %s", err.Error()))
			return
		} else if !ok {
			chalker.Log(chalker.ERROR, "paymail address failed format validation: unknown reason")
			return
		}

		// No sender handle given? (default: set to the receiver's paymail address)
		if len(senderHandle) == 0 {
			chalker.Log(chalker.WARN, fmt.Sprintf("--sender-handle not set, using: %s", paymailAddress))
			senderHandle = paymailAddress
			senderDomain, senderHandle = paymail.ExtractParts(senderHandle)
		} else { // Sender handle is set (basic validation)

			// Validate the format for the given handle
			if ok, err := validate.IsValidEmail(senderHandle, false); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("--sender-handle failed format validation: %s", err.Error()))
				return
			} else if !ok {
				chalker.Log(chalker.ERROR, "--sender-handle failed format validation: unknown reason")
				return
			}

			// Check for a real domain (require at least one period)
			if !strings.Contains(senderDomain, ".") {
				chalker.Log(chalker.ERROR, fmt.Sprintf("--sender-handle domain name is invalid: %s", senderDomain))
				return
			} else if !validate.IsValidDNSName(senderDomain) { // Basic DNS check (not a REAL domain name check)
				chalker.Log(chalker.ERROR, fmt.Sprintf("--sender-handle domain name failed DNS check: %s", senderDomain))
				return
			}
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

		// Does the paymail provider have the capability?
		if len(capabilities.PaymentDestination) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("missing a required capability: %s", paymail.CapabilityPaymentDestination))
			return
		}

		// Does this provider require sender validation?
		// https://bsvalias.org/04-02-sender-validation.html
		if capabilities.SenderValidation {
			chalker.Log(chalker.WARN, "sender validation is ENFORCED")

			// Required if flag is enforced
			if len(signature) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("missing required flag: %s - see the help section: -h", "--signature"))

				// todo: generate a real signature if possible
				chalker.Log(chalker.WARN, fmt.Sprintf("attempting to fake a signature for: %s...", senderHandle))
				signature, _ = RandomHex(64)
			}

			// Only if it's not the same (set from above ^^)
			if senderHandle != paymailAddress {

				// Get the capabilities
				senderCapabilities, getErr := getCapabilities(senderDomain)
				if getErr != nil {
					if strings.Contains(getErr.Error(), "context deadline exceeded") {
						chalker.Log(chalker.WARN, fmt.Sprintf("no capabilities found for: %s", domain))
					} else {
						chalker.Log(chalker.ERROR, fmt.Sprintf("error: %s", getErr.Error()))
					}
					return
				}
				// Does the paymail provider have the capability?
				if len(senderCapabilities.Pki) == 0 {
					chalker.Log(chalker.ERROR, fmt.Sprintf("--sender-handle missing a required capability: %s", paymail.CapabilityPaymentDestination))
					return
				}

				// Get the alias of the address
				parts := strings.Split(senderHandle, "@")

				// Get the PKI for the given address
				var senderPki *paymail.PKIResponse
				if senderPki, err = getPki(senderCapabilities.Pki, parts[0], parts[1]); err != nil {
					chalker.Log(chalker.ERROR, fmt.Sprintf("error: %s", err.Error()))
					return
				} else if senderPki != nil {
					chalker.Log(chalker.INFO, fmt.Sprintf("--sender-handle %s@%s's pubkey: %s", parts[0], parts[1], chalk.Cyan.Color(senderPki.PubKey)))
				}
			}

			// once completed, the full sender validation will be complete
			chalker.Log(chalker.SUCCESS, "send request pre-validation: passed")
		}

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Get the PKI for the given address
		var pki *paymail.PKIResponse
		if pki, err = getPki(capabilities.Pki, parts[0], domain); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("error: %s", err.Error()))
			return
		}

		// Setup the request body
		senderRequest := &paymail.AddressResolutionRequest{
			Amount:       amount,
			Dt:           time.Now().UTC().Format(time.RFC3339), // UTC is assumed
			Purpose:      purpose,
			SenderHandle: senderHandle,
			SenderName:   senderName,
			Signature:    signature,
		}

		// Resolve the address from a given paymail
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("resolving address: %s...", chalk.Cyan.Color(parts[0]+"@"+domain)))

		var resolutionResponse *paymail.AddressResolutionResponse
		if resolutionResponse, err = paymail.AddressResolution(capabilities.PaymentDestination, parts[0], domain, senderRequest); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("address resolution failed: %s", err.Error()))
			return
		}

		// Success
		chalker.Log(chalker.SUCCESS, "address resolution successful")

		// Attempt to get a public profile if the capability is found
		if len(capabilities.PublicProfile) > 0 && !skipPublicProfile {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("getting public profile for: %s...", chalk.Cyan.Color(parts[0]+"@"+domain)))
			var profile *paymail.PublicProfileResponse
			if profile, err = paymail.GetPublicProfile(capabilities.PublicProfile, parts[0], domain); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("get public profile failed: %s", err.Error()))
				return
			} else if profile != nil {
				if len(profile.Name) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("name: %s", chalk.Cyan.Color(profile.Name)))
				}
				if len(profile.Avatar) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("avatar: %s", chalk.Cyan.Color(profile.Avatar)))
				}
			}
		}

		// Show pubkey
		if pki != nil && len(pki.PubKey) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("pubkey: %s", chalk.Cyan.Color(pki.PubKey)))
		}

		// Show output script
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("output script: %s", chalk.Cyan.Color(resolutionResponse.Output)))

		// Show the resolved address from the output script
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("address: %s", chalk.Cyan.Color(resolutionResponse.Address)))
	},
}

func init() {
	rootCmd.AddCommand(resolveCmd)

	// Set the amount for the sender request
	resolveCmd.Flags().Uint64VarP(&amount, "amount", "a", 0, "Amount in satoshis for the payment request")

	// Set the purpose for the sender request
	resolveCmd.Flags().StringVarP(&purpose, "purpose", "p", "", "Purpose for the transaction")

	// Set the sender's handle for the sender request
	resolveCmd.Flags().StringVar(&senderHandle, "sender-handle", "", "Sender's paymail handle. Required by bsvalias spec. Receiver paymail used if not specified.")

	// Set the sender's name for the sender request
	resolveCmd.Flags().StringVarP(&senderName, "sender-name", "n", "", "The sender's name")

	// Set the signature of the entire request
	resolveCmd.Flags().StringVarP(&signature, "signature", "s", "", "The signature of the entire request")

	// Skip getting the PubKey
	resolveCmd.Flags().BoolVar(&skipPki, "skip-pki", false, "Skip firing pki request and getting the pubkey")

	// Skip getting public profile
	resolveCmd.Flags().BoolVar(&skipPublicProfile, "skip-public-profile", false, "Skip firing public profile request and getting the avatar")
}
