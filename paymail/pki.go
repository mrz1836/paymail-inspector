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
	BsvAlias string `json:"bsvalias"` // Version
	Handle   string `json:"handle"`   // The <alias>@<domain>.<tld>
	PubKey   string `json:"pubkey"`   // The related PubKey
}

// GetPKI will return a valid PKI response
// Specs: http://bsvalias.org/03-public-key-infrastructure.html
func GetPKI(pkiUrl, alias, domain string) (pki *PKIResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/{alias}@{domain.tld}/id
	reqURL := strings.Replace(strings.Replace(pkiUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

	// Start the request
	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, reqURL, nil); err != nil {
		return
	}

	// Set the headers (standard user agent so it cannot be blocked)
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
		err = fmt.Errorf("bad response from paymail provider: %d", resp.StatusCode)
		return
	}

	// Try and decode the response
	if err = json.NewDecoder(resp.Body).Decode(&pki); err != nil {
		return
	}

	// Invalid version?
	if len(pki.BsvAlias) == 0 {
		err = fmt.Errorf("missing bsvalias version")
		return
	}

	// Check basic requirements (handle)
	if pki.Handle != alias+"@"+domain {
		err = fmt.Errorf("pki response handle %s does not match paymail address: %s", pki.Handle, alias+"@"+domain)
		return
	}

	// Check the PubKey length
	if len(pki.PubKey) == 0 {
		err = fmt.Errorf("pki response is missing a PubKey value")
	} else if len(pki.PubKey) != pubKeyLength {
		err = fmt.Errorf("returned pubkey is not the required length of %d, got: %d", pubKeyLength, len(pki.PubKey))
	}

	return
}
