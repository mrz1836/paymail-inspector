package paymail

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvd/txscript"
	"github.com/bitcoinsv/bsvutil"
)

/*
Example:
{
  "satoshis": 1000100
}
*/

// P2PPaymentDestinationRequest is the request body for the P2P payment request
type P2PPaymentDestinationRequest struct {
	Satoshis uint64 `json:"satoshis"` // The amount, in Satoshis, that the sender intends to transfer to the receiver
}

// P2PPaymentDestinationResponse is the response frm the request
type P2PPaymentDestinationResponse struct {
	Outputs   []*Output `json:"outputs"`   // A list of outputs
	Reference string    `json:"reference"` // A reference for the payment, created by the receiver of the transaction
}

// Output is returned inside the payment destination response
type Output struct {
	Address  string `json:"address,omitempty"`  // Hex encoded locking script
	Satoshis uint64 `json:"satoshis,omitempty"` // Number of satoshis for that output
	Script   string `json:"script"`             // Hex encoded locking script
}

// GetP2PPaymentDestination will return list of outputs for the P2P transactions to use
// Specs: https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html
func GetP2PPaymentDestination(p2pUrl, alias, domain string, senderRequest *P2PPaymentDestinationRequest) (response *P2PPaymentDestinationResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/api/rawtx/{alias}@{domain.tld}
	reqURL := strings.Replace(strings.Replace(p2pUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

	// Set post value
	var jsonValue []byte
	if jsonValue, err = json.Marshal(senderRequest); err != nil {
		return
	}

	// Start the request
	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(jsonValue)); err != nil {
		return
	}

	// Set the headers (standard user agent so it cannot be blocked)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent)

	// Set the client
	client := http.Client{
		Timeout: defaultPostTimeout * time.Second,
	}

	// Fire the request
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}

	// Close the body
	defer func() {
		_ = resp.Body.Close()
	}()

	// Test the status code
	// Only 200 and 304 are accepted
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotModified {

		// Paymail address not found?
		if resp.StatusCode == http.StatusNotFound {
			err = fmt.Errorf("paymail address not found")
		} else {
			err = fmt.Errorf("bad response from paymail provider: %d", resp.StatusCode)
		}

		return
	}

	// Try and decode the response
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return
	}

	// Check for a reference number
	if len(response.Reference) == 0 {
		err = fmt.Errorf("missing a returned reference value")
		return
	}

	// No outputs?
	if len(response.Outputs) == 0 {
		err = fmt.Errorf("missing a returned output")
		return
	}

	// Loop all outputs
	for index, output := range response.Outputs {

		// No script returned
		if len(output.Script) == 0 {
			continue
		}

		// Decode the hex string into bytes
		var script []byte
		if script, err = hex.DecodeString(output.Script); err != nil {
			return
		}

		// Extract the components from the script
		var addresses []bsvutil.Address
		if _, addresses, _, err = txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams); err != nil {
			return
		}

		// Missing an address?
		if len(addresses) == 0 {
			err = fmt.Errorf("invalid output script, missing an address")
			return
		}

		// Extract the address from the pubkey hash
		var address *bsvutil.LegacyAddressPubKeyHash
		if address, err = bsvutil.NewLegacyAddressPubKeyHash(addresses[0].ScriptAddress(), &chaincfg.MainNetParams); err != nil {
			return
		} else if address == nil {
			err = fmt.Errorf("failed in NewLegacyAddressPubKeyHash, address was nil")
			return
		}

		// Use the encoded version of the address
		response.Outputs[index].Address = address.EncodeAddress()
	}

	return
}
