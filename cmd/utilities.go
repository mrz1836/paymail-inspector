package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
)

// RandomHex returns a random hex string and error
func RandomHex(n int) (hexString string, err error) {
	b := make([]byte, n)
	if _, err = rand.Read(b); err != nil {
		return
	}
	return hex.EncodeToString(b), nil
}

// getPki will get a pki response
func getPki(url, alias, domain string) (pki *paymail.PKIResponse, err error) {

	chalker.Log(chalker.DEFAULT, fmt.Sprintf("getting PKI for %s@%s...", alias, domain))

	// Get the PKI for the given address
	if pki, err = paymail.GetPKI(url, alias, domain); err != nil {
		return
	}

	// No pubkey found
	if len(pki.PubKey) == 0 {
		err = fmt.Errorf("failed getting pubkey for: %s@%s", alias, domain)
		return
	}

	// Possible invalid pubkey
	if len(pki.PubKey) != paymail.PubKeyLength {
		chalker.Log(chalker.WARN, fmt.Sprintf("pubkey length is: %d, expected: %d", len(pki.PubKey), paymail.PubKeyLength))
	}

	return
}

// validatePaymailAndDomain will do a basic validation on the paymail format
func validatePaymailAndDomain(paymailAddress, domain string) (valid bool) {

	// Validate the format for the paymail address (paymail addresses follow conventional email requirements)
	if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("paymail address failed format validation: %s", err.Error()))
		return
	} else if !ok {
		chalker.Log(chalker.ERROR, "paymail address failed format validation: unknown reason")
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

	valid = true
	return
}
