package paymail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
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
	StandardResponse
	Outputs   []*Output `json:"outputs"`   // A list of outputs
	Reference string    `json:"reference"` // A reference for the payment, created by the receiver of the transaction
}

// Output is returned inside the payment destination response
type Output struct {
	Satoshis uint64 `json:"satoshis,omitempty"` // Number of satoshis for that output
	Script   string `json:"script"`             // Hex encoded locking script
}

// GetP2PPaymentDestination will return list of outputs for the P2P transactions to use
// Specs: https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html
func GetP2PPaymentDestination(p2pUrl, alias, domain string, senderRequest *P2PPaymentDestinationRequest, tracing bool) (response *P2PPaymentDestinationResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/api/rawtx/{alias}@{domain.tld}
	reqURL := strings.Replace(strings.Replace(p2pUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

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
	response = new(P2PPaymentDestinationResponse)

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

	return
}
