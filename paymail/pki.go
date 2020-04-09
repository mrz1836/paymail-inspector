package paymail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	BsvAlias string `json:"bsvalias"`
	Handle   string `json:"handle"`
	PubKey   string `json:"pubkey"`
}

// GetPKI will return a valid PKI response
// Specs: http://bsvalias.org/03-public-key-infrastructure.html
func GetPKI(pkiUrl, alias, domain string) (pki *PKIResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/{alias}@{domain.tld}/id
	reqURL := strings.Replace(strings.Replace(pkiUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

	// Start the request
	var req *http.Request
	if req, err = http.NewRequest("GET", reqURL, nil); err != nil {
		return
	}

	// Set the headers (standard user agent so it cannot be blocked)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36")

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
	if resp.StatusCode != 200 && resp.StatusCode != 304 {
		err = fmt.Errorf("bad response from paymail provider: %d", resp.StatusCode)
		return
	}

	// Try and decode the response
	err = json.NewDecoder(resp.Body).Decode(&pki)

	return
}
