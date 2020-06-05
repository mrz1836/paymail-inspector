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
	"hex": "01000000012adda020db81f2155ebba69e7c841275517ebf91674268c32ff2f5c7e2853b2c010000006b483045022100872051ef0b6c47714130c12a067db4f38b988bfc22fe270731c2146f5229386b02207abf68bbf092ec03e2c616defcc4c868ad1fc3cdbffb34bcedfab391a1274f3e412102affe8c91d0a61235a3d07b1903476a2e2f7a90451b2ed592fea9937696a07077ffffffff02ed1a0000000000001976a91491b3753cf827f139d2dc654ce36f05331138ddb588acc9670300000000001976a914da036233873cc6489ff65a0185e207d243b5154888ac00000000",
	"metadata": {
		"sender": "someone@example.tld",
		"pubkey": "<somepubkey>",
		"signature": "signature(txid)",
		"note": "Human readeble information related to the tx."
	},
	"reference": "someRefId"
}
*/

// P2PTransactionRequest is the request body for the P2P transaction request
type P2PTransactionRequest struct {
	Hex       string    `json:"hex"`       // The raw transaction, encoded as a hexadecimal string
	MetaData  *MetaData `json:"metadata"`  // An object containing data associated with the transaction
	Reference string    `json:"reference"` // Reference for the payment

}

// MetaData is an object containing data associated with the transaction
type MetaData struct {
	Sender    string `json:"sender,omitempty"`    // The paymail of the person that originated the transaction
	PubKey    string `json:"pubkey,omitempty"`    // Public key to validate the signature
	Signature string `json:"signature,omitempty"` // A signature of the tx id made by the sender
	Note      string `json:"note,omitempty"`      // A human readable information about the payment.
}

// P2PTransactionResponse is the response to the request
type P2PTransactionResponse struct {
	StandardResponse
	TxID string `json:"txid"` // The txid of the broadcasted tx
	Note string `json:"note"` // Some human readable note
}

// SendP2PTransaction will submit a transaction hex string (txhex) to a paymail provider
// Specs: https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html
func SendP2PTransaction(p2pUrl, alias, domain string, senderRequest *P2PTransactionRequest, tracing bool) (response *P2PTransactionResponse, err error) {

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
	response = new(P2PTransactionResponse)

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
	if len(response.TxID) == 0 {
		err = fmt.Errorf("missing a returned txid")
		return
	}

	return
}
