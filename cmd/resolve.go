package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
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
			return fmt.Errorf("%s%s requires either a paymail address%s\n", chalk.Red, logPrefix, chalk.Reset)
		} else if len(args) > 1 {
			return fmt.Errorf("%s%s validate only supports one address at a time%s\n", chalk.Red, logPrefix, chalk.Reset)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		domain, paymailAddress := paymail.ExtractParts(args[0])

		// Did we get a paymail address?
		if len(paymailAddress) == 0 {
			fmt.Printf("%s%s paymail address not found or invalid%s\n", chalk.Red, logPrefix, chalk.Reset)
			return
		}

		// Validate the email format for the paymail address (paymail addresses follow conventional email requirements)
		if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
			fmt.Printf("%s%s paymail address failed email format validation: %s%s\n", chalk.Red, logPrefix, err.Error(), chalk.Reset)
			return
		} else if !ok {
			fmt.Printf("%s%s paymail address failed email format validation: %s%s\n", chalk.Red, logPrefix, "unknown reason", chalk.Reset)
			return
		}

		// Validate the email format for the given handle
		if ok, err := validate.IsValidEmail(senderHandle, false); err != nil {
			fmt.Printf("%s%s sender handle failed email format validation: %s%s\n", chalk.Red, logPrefix, err.Error(), chalk.Reset)
			return
		} else if !ok {
			fmt.Printf("%s%s sender handle failed email format validation: %s%s\n", chalk.Red, logPrefix, "unknown reason", chalk.Reset)
			return
		}

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			fmt.Printf("%s%s domain name is invalid: %s%s\n", chalk.Red, logPrefix, domain, chalk.Reset)
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			fmt.Printf("%s%s domain name failed DNS check: %s%s\n", chalk.Red, logPrefix, domain, chalk.Reset)
			return
		}

		// Get the details from the SRV record
		fmt.Printf("%s getting SRV record...\n", logPrefix)
		srv, err := paymail.GetSRVRecord(serviceName, protocol, domain, nameServer)
		if err != nil {
			fmt.Printf("%s%s get SRV record failed: %s%s\n", chalk.Red, logPrefix, err.Error(), chalk.Reset)
			return
		}

		// Get the capabilities for the given domain
		fmt.Printf("%s getting capabilities...\n", logPrefix)
		var capabilities *paymail.CapabilitiesResponse
		if capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port)); err != nil {
			fmt.Printf("%s%s get capabilities failed: %s%s\n", chalk.Red, logPrefix, err.Error(), chalk.Reset)
			return
		}

		// Check the version
		if capabilities.BsvAlias != bsvAliasVersion {
			fmt.Printf("%s%s capabilities bsvalias version mismatch, expected: %s but got: %s%s\n", chalk.Red, logPrefix, bsvAliasVersion, capabilities.BsvAlias, chalk.Reset)
			return
		}

		// Do we have the capability?
		if len(capabilities.PaymentDestination) == 0 {
			fmt.Printf("%s%s missing a required capability: %s%s\n", chalk.Red, logPrefix, paymail.CapabilityPaymentDestination, chalk.Reset)
			return
		}

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Get the PKI for the given address
		fmt.Printf("%s getting PKI...\n", logPrefix)
		var pki *paymail.PKIResponse
		if pki, err = paymail.GetPKI(capabilities.Pki, parts[0], domain); err != nil {
			fmt.Printf("%s%s get PKI failed: %s%s\n", chalk.Red, logPrefix, err.Error(), chalk.Reset)
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
		fmt.Printf("%s resolving address...\n", logPrefix)
		var resolution *paymail.AddressResolutionResponse
		if resolution, err = paymail.AddressResolution(capabilities.PaymentDestination, parts[0], domain, senderRequest); err != nil {
			fmt.Printf("%s%s address resolution failed: %s%s\n", chalk.Red, logPrefix, err.Error(), chalk.Reset)
			return
		}

		// Success!
		fmt.Printf("%s%s address resolution successful%s\n", chalk.Green, logPrefix, chalk.Reset)
		fmt.Printf("%s%s pubkey: %s\n", chalk.Cyan, logPrefix, pki.PubKey)
		fmt.Printf("%s output hash: %s\n", logPrefix, resolution.Output)
		fmt.Printf("%s address: %s%s\n", logPrefix, resolution.Address, chalk.Reset)
	},
}

func init() {
	rootCmd.AddCommand(resolveCmd)

	// Set the amount for the sender request
	resolveCmd.Flags().Uint64VarP(&amount, "amount", "a", 0, chalk.Cyan.Color("Amount in satoshis for the payment request"))

	// Set the purpose for the sender request
	resolveCmd.Flags().StringVarP(&purpose, "purpose", "p", "", chalk.Cyan.Color("Purpose for the transaction"))

	// Set the sender's handle for the sender request
	resolveCmd.Flags().StringVar(&senderHandle, "sender-handle", "", chalk.Cyan.Color("The sender's paymail handle")+" "+chalk.Red.Color("Required"))
	_ = resolveCmd.MarkFlagRequired("sender-handle")

	// Set the sender's name for the sender request
	resolveCmd.Flags().StringVarP(&senderName, "sender-name", "n", "", chalk.Cyan.Color("The sender's name"))

	// Set the signature of the entire request
	resolveCmd.Flags().StringVarP(&signature, "signature", "s", "", chalk.Cyan.Color("The signature of the entire request"))
	fmt.Printf("%+v", resolveCmd.Flags())
}
