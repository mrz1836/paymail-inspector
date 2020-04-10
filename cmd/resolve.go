package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
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
	Example:    "resolve this@address.com --sender-handle you@yourdomain.com",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("%s requires either a paymail address\n", logPrefix)
		} else if len(args) > 1 {
			return fmt.Errorf("%s validate only supports one address at a time\n", logPrefix)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		domain, paymailAddress := paymail.ExtractParts(args[0])

		// Did we get a paymail address?
		if len(paymailAddress) == 0 {
			fmt.Printf("%s paymail address not found or invalid\n", logPrefix)
			return
		}

		// Validate the email format for the paymail address (paymail addresses follow conventional email requirements)
		if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
			fmt.Printf("%s paymail address failed email format validation: %s\n", logPrefix, err.Error())
			return
		} else if !ok {
			fmt.Printf("%s paymail address failed email format validation: %s\n", logPrefix, "unknown reason")
			return
		}

		// Validate the email format for the given handle
		if ok, err := validate.IsValidEmail(senderHandle, false); err != nil {
			fmt.Printf("%s sender handle failed email format validation: %s\n", logPrefix, err.Error())
			return
		} else if !ok {
			fmt.Printf("%s sender handle failed email format validation: %s\n", logPrefix, "unknown reason")
			return
		}

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			fmt.Printf("%s domain name is invalid: %s\n", logPrefix, domain)
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			fmt.Printf("%s domain name failed DNS check: %s\n", logPrefix, domain)
			return
		}

		// Get the details from the SRV record
		fmt.Printf("%s getting SRV record...\n", logPrefix)
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			fmt.Printf("%s get SRV record failed: %s\n", logPrefix, err.Error())
			return
		}

		// Get the capabilities for the given domain
		fmt.Printf("%s getting capabilities...\n", logPrefix)
		var capabilities *paymail.CapabilitiesResponse
		if capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port)); err != nil {
			fmt.Printf("%s get capabilities failed: %s\n", logPrefix, err.Error())
			return
		}

		// Check the version
		if capabilities.BsvAlias != bsvAliasVersion {
			fmt.Printf("%s capabilities bsvalias version mismatch, expected: %s but got: %s\n", logPrefix, bsvAliasVersion, capabilities.BsvAlias)
			return
		}

		// Do we have the capability?
		if len(capabilities.PaymentDestination) == 0 {
			fmt.Printf("%s missing a required capabilitiy: %s\n", logPrefix, paymail.CapabilityPaymentDestination)
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

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Resolve the address from a given paymail
		fmt.Printf("%s resolving address...\n", logPrefix)
		var resolution *paymail.AddressResolutionResponse
		if resolution, err = paymail.AddressResolution(capabilities.PaymentDestination, parts[0], domain, senderRequest); err != nil {
			fmt.Printf("%s address resolution failed: %s\n", logPrefix, err.Error())
			return
		}

		// Success!
		fmt.Printf("%s address resolution successful\n", logPrefix)
		fmt.Printf("%s output hash: %s\n", logPrefix, resolution.Output)
		fmt.Printf("%s address: %s\n", logPrefix, resolution.Address)
	},
}

func init() {
	rootCmd.AddCommand(resolveCmd)

	// Set the amount for the sender request
	resolveCmd.Flags().Uint64VarP(&amount, "amount", "a", 0, "Amount in satoshis for the payment request")

	// Set the purpose for the sender request
	resolveCmd.Flags().StringVarP(&purpose, "purpose", "p", "", "Purpose for the transaction")

	// Set the sender's handle for the sender request
	resolveCmd.Flags().StringVar(&senderHandle, "sender-handle", "", "(Required) The sender's paymail handle")
	_ = resolveCmd.MarkFlagRequired("sender-handle")

	// Set the sender's name for the sender request
	resolveCmd.Flags().StringVarP(&senderName, "sender-name", "n", "", "The sender's name")

	// Set the signature of the entire request
	resolveCmd.Flags().StringVarP(&signature, "signature", "s", "", "The signature of the entire request")
}
