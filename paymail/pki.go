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
Default:

{
  "bsvalias": "1.0",
  "handle": "<alias>@<domain>.<tld>",
  "pubkey": "..."
}
*/

// PKIResponse is the result returned
type PKIResponse struct {
	StandardResponse
	BsvAlias string `json:"bsvalias"` // Version
	Handle   string `json:"handle"`   // The <alias>@<domain>.<tld>
	PubKey   string `json:"pubkey"`   // The related PubKey
}

// GetPKI will return a valid PKI response
// Specs: http://bsvalias.org/03-public-key-infrastructure.html
func GetPKI(pkiUrl, alias, domain string, tracing bool) (response *PKIResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/{alias}@{domain.tld}/id
	reqURL := strings.Replace(strings.Replace(pkiUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

	// Create a Client and start the request
	client := resty.New().SetTimeout(defaultGetTimeout * time.Second)
	var resp *resty.Response
	req := client.R().SetHeader("User-Agent", defaultUserAgent)
	if tracing {
		req.EnableTrace()
	}
	if resp, err = req.Get(reqURL); err != nil {
		return
	}

	// New struct
	response = new(PKIResponse)

	// Tracing enabled?
	if tracing {
		response.Tracing = resp.Request.TraceInfo()
	}

	// Test the status code
	response.StatusCode = resp.StatusCode()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		err = fmt.Errorf("bad response from paymail provider: %d", response.StatusCode)
		return
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return
	}

	// Invalid version?
	if len(response.BsvAlias) == 0 {
		err = fmt.Errorf("missing bsvalias version")
		return
	}

	// Check basic requirements (handle)
	if response.Handle != alias+"@"+domain {
		err = fmt.Errorf("pki response handle %s does not match paymail address: %s", response.Handle, alias+"@"+domain)
		return
	}

	// Check the PubKey length
	if len(response.PubKey) == 0 {
		err = fmt.Errorf("pki response is missing a PubKey value")
	} else if len(response.PubKey) != PubKeyLength {
		err = fmt.Errorf("returned pubkey is not the required length of %d, got: %d", PubKeyLength, len(response.PubKey))
	}

	return
}
