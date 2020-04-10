package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flags for the resolve command
var (
	amount       uint64
	purpose      string
	senderHandle string
	senderName   string
	signature    string
)

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:        "resolve",
	Short:      "Resolves a paymail address",
	Long:       `Resolves a paymail address into a hex-encoded Bitcoin script and address`,
	Aliases:    []string{"resolution", "address_resolution"},
	SuggestFor: []string{"address", "destination"},
	Example:    "resolve this@address.com",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("%s requires either a paymail address")
		} else if len(args) > 1 {
			return chalker.Error("validate only supports one address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
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

		// No sender handle given? (set to the receiver's paymail address)
		if len(senderHandle) == 0 {
			senderHandle = paymailAddress
			chalker.Log(chalker.WARN, fmt.Sprintf("--sender-handle not set, using: %s", paymailAddress))
		}

		// Validate the format for the given handle
		if ok, err := validate.IsValidEmail(senderHandle, false); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("sender handle failed format validation: %s", err.Error()))
			return
		} else if !ok {
			chalker.Log(chalker.ERROR, "sender handle failed format validation: unknown reason")
			return
		}

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			chalker.Log(chalker.ERROR, fmt.Sprintf("domain name is invalid: %s", domain))
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			chalker.Log(chalker.ERROR, fmt.Sprintf("domain name failed DNS check: %s", domain))
			return
		}

		// Get the details from the SRV record
		chalker.Log(chalker.DEFAULT, "getting SRV record...")
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("get SRV record failed: %s", err.Error()))
			return
		}

		// Get the capabilities for the given domain
		chalker.Log(chalker.DEFAULT, "getting capabilities...")
		var capabilities *paymail.CapabilitiesResponse
		if capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port)); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("get capabilities failed: %s", err.Error()))
			return
		}

		// Check the version
		if capabilities.BsvAlias != viper.GetString(flagBsvAlias) {
			chalker.Log(chalker.ERROR, fmt.Sprintf("capabilities %s version mismatch, expected: %s but got: %s", flagBsvAlias, viper.GetString(flagBsvAlias), capabilities.BsvAlias))
			return
		}

		// Do we have the capability?
		if len(capabilities.PaymentDestination) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("missing a required capability: %s", paymail.CapabilityPaymentDestination))
			return
		}

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Get the PKI for the given address
		chalker.Log(chalker.DEFAULT, "getting PKI...")
		var pki *paymail.PKIResponse
		if pki, err = paymail.GetPKI(capabilities.Pki, parts[0], domain); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("get PKI failed: %s", err.Error()))
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
		chalker.Log(chalker.DEFAULT, "resolving address...")

		var resolution *paymail.AddressResolutionResponse
		if resolution, err = paymail.AddressResolution(capabilities.PaymentDestination, parts[0], domain, senderRequest); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("address resolution failed: %s", err.Error()))
			return
		}

		// Success!
		chalker.Log(chalker.SUCCESS, "address resolution successful")
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("pubkey: %s", pki.PubKey))
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("output hash: %s", resolution.Output))
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("address: %s", resolution.Address))
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
}
