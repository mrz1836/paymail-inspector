package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies if a given paymail is associated to the pubkey",
	Long: `Verify will check the paymail address against a given pubkey 
using the provider domain (if capability is supported)`,
	Aliases:    []string{"verification"},
	SuggestFor: []string{"pubkey", "veri"},
	Example:    "verify mrz@" + defaultDomainName + " 02ead23149a1e33df17325ec7a7ba9e0b20c674c57c630f527d69b866aa9b65b10",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return chalker.Error("%s requires a paymail address AND pubkey")
		} else if len(args) > 2 {
			return chalker.Error("verify only supports one address and one pubkey at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		var paymailAddress string
		var pubKey string

		// Check for paymail in both args
		if strings.Contains(args[0], "@") {
			paymailAddress = args[0]
			pubKey = args[1]
		} else if strings.Contains(args[1], "@") {
			pubKey = args[0]
			paymailAddress = args[1]
		}

		// Require a paymail address
		if len(paymailAddress) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("one argument must be a paymail address [%s] [%s]", args[0], args[1]))
			return
		}

		// Require a pubkey
		if len(pubKey) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("one argument must be a pubkey [%s] [%s]", args[0], args[1]))
			return
		}

		// Extract the parts given
		domain, _ := paymail.ExtractParts(paymailAddress)

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			chalker.Log(chalker.ERROR, fmt.Sprintf("domain name is invalid: %s", domain))
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			chalker.Log(chalker.ERROR, fmt.Sprintf("domain name failed DNS check: %s", domain))
			return
		}

		// Validate pubkey
		if len(pubKey) != paymail.PubKeyLength {
			chalker.Log(chalker.ERROR, fmt.Sprintf("pubkey is an invalid length, expected: %d but got: %d", paymail.PubKeyLength, len(pubKey)))
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
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("getting capabilities from %s...", srv.Target))
		var capabilities *paymail.CapabilitiesResponse
		if capabilities, err = paymail.GetCapabilities(srv.Target, int(srv.Port)); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("get capabilities failed: %s", err.Error()))
			return
		}

		// Does the paymail provider have the capability?
		if len(capabilities.VerifyPublicKeyOwner) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("missing a required capability: %s", paymail.CapabilityVerifyPublicKeyOwner))
			return
		}

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Fire the verify request
		var verify *paymail.VerifyPubKeyResponse
		if verify, err = paymail.VerifyPubKey(capabilities.VerifyPublicKeyOwner, parts[0], domain, pubKey); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("get capabilities failed: %s", err.Error()))
			return
		} else if verify == nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("failed getting/parsing verify response for: %s", paymailAddress))
			return
		}

		// Show the results
		chalker.Log(chalker.INFO, fmt.Sprintf("paymail: %s", paymailAddress))
		chalker.Log(chalker.INFO, fmt.Sprintf("pubkey: %s", pubKey))

		if verify.Match {
			chalker.Log(chalker.SUCCESS, "Paymail & PubKey Match!")
		} else {
			chalker.Log(chalker.ERROR, "DO NOT match!")
		}
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
