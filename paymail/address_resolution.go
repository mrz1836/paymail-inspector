package paymail

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvd/txscript"
	"github.com/bitcoinsv/bsvutil"
)

/*
Example:
{
    "senderName": "FirstName LastName",
    "senderHandle": "<alias>@<domain.tld>",
    "dt": "2013-10-21T13:28:06.419Z",
    "amount": 550,
    "purpose": "message to receiver",
    "signature": "<compact Bitcoin message signature>"
}
*/

// AddressResolutionRequest is the request body for the basic address resolution
type AddressResolutionRequest struct {
	Amount       uint64 `json:"amount,omitempty"`     // The amount, in Satoshis, that the sender intends to transfer to the receiver
	Dt           string `json:"dt"`                   // (required) ISO-8601 formatted timestamp; see notes
	Purpose      string `json:"purpose,omitempty"`    // Human-readable description of the purpose of the payment
	SenderHandle string `json:"senderHandle"`         // (required) Sender's paymail handle
	SenderName   string `json:"senderName,omitempty"` // Human-readable sender display name
	Signature    string `json:"signature,omitempty"`  // Compact Bitcoin message signature; see notes
}

// AddressResolutionResponse is the response frm the request
type AddressResolutionResponse struct {
	Address   string `json:"address"`             // Legacy BSV address derived from the output hash
	Output    string `json:"output"`              // hex-encoded Bitcoin script, which the sender MUST use during the construction of a payment transaction
	Signature string `json:"signature,omitempty"` // This is used if SenderValidation is enforced
}

// AddressResolution will return a hex-encoded Bitcoin script if successful
// Specs: http://bsvalias.org/04-01-basic-address-resolution.html
func AddressResolution(resolutionUrl, alias, domain string, senderRequest *AddressResolutionRequest) (response *AddressResolutionResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/{alias}@{domain.tld}/payment-destination
	reqURL := strings.Replace(strings.Replace(resolutionUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

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

	// Fire the request
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
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

	// Check for an output
	if len(response.Output) == 0 {
		err = fmt.Errorf("missing an output value")
		return
	}

	// Decode the hex string into bytes
	var script []byte
	if script, err = hex.DecodeString(response.Output); err != nil {
		return
	}

	// Extract the components from the script
	var addresses []bsvutil.Address
	if _, addresses, _, err = txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams); err != nil {
		return
	}

	// Missing an address?
	if len(addresses) == 0 {
		err = fmt.Errorf("invalid output hash, missing an address")
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
	response.Address = address.EncodeAddress()

	return
}
