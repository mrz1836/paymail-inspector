package paymail

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvd/txscript"
	"github.com/bitcoinsv/bsvutil"
	"github.com/go-resty/resty/v2"
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
	StandardResponse
	Address   string `json:"address"`             // Legacy BSV address derived from the output script
	Output    string `json:"output"`              // hex-encoded Bitcoin script, which the sender MUST use during the construction of a payment transaction
	Signature string `json:"signature,omitempty"` // This is used if SenderValidation is enforced
}

// AddressResolution will return a hex-encoded Bitcoin script if successful
// Specs: http://bsvalias.org/04-01-basic-address-resolution.html
func AddressResolution(resolutionUrl, alias, domain string, senderRequest *AddressResolutionRequest, tracing bool) (response *AddressResolutionResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/{alias}@{domain.tld}/payment-destination
	reqURL := strings.Replace(strings.Replace(resolutionUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

	// Create a Client and start the request
	client := resty.New().SetTimeout(defaultPostTimeout * time.Second)
	var resp *resty.Response
	req := client.R().SetBody(senderRequest).SetHeader("User-Agent", UserAgent)
	if tracing {
		req.EnableTrace()
	}
	if resp, err = req.Post(reqURL); err != nil {
		return
	}

	// New struct
	response = new(AddressResolutionResponse)

	// Tracing enabled?
	if tracing {
		response.Tracing = resp.Request.TraceInfo()
	}

	// Test the status code
	response.StatusCode = resp.StatusCode()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		// Paymail address not found?
		if response.StatusCode == http.StatusNotFound {
			err = fmt.Errorf("paymail address not found")
		} else {
			err = fmt.Errorf("bad response from paymail provider: %d", response.StatusCode)
		}

		return
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
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
	response.Address = address.EncodeAddress()

	return
}
