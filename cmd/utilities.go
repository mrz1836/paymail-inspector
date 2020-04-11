package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

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
