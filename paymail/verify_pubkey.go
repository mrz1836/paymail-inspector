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
  "handle":"somepaymailhandle@domain.tld",
  "match": true,
  "pubkey":"<consulted pubkey>"
}
*/

// VerifyPubKeyResponse is the result returned
type VerifyPubKeyResponse struct {
	BsvAlias string `json:"bsvalias"` // Version of the bsvalias
	Handle   string `json:"handle"`   // The <alias>@<domain>.<tld>
	Match    bool   `json:"match"`    // If the match was successful or not
	PubKey   string `json:"pubkey"`   // The related PubKey
}

// VerifyPubKey will try to match a handle and pubkey
// Specs: https://bsvalias.org/05-verify-public-key-owner.html
func VerifyPubKey(verifyUrl, alias, domain, pubKey string) (response *VerifyPubKeyResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/verifypubkey/{alias}@{domain.tld}/{pubkey}
	reqURL := strings.Replace(strings.Replace(strings.Replace(verifyUrl, "{pubkey}", pubKey, -1), "{alias}", alias, -1), "{domain.tld}", domain, -1)

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
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return
	}

	// Invalid version?
	if len(response.BsvAlias) == 0 {
		err = fmt.Errorf("missing bsvalias version")
		return
	}

	// Check basic requirements (handle)
	if response.Handle != alias+"@"+domain {
		err = fmt.Errorf("verify response handle %s does not match paymail address: %s", response.Handle, alias+"@"+domain)
		return
	}

	// Check the PubKey length
	if len(response.PubKey) == 0 {
		err = fmt.Errorf("verify response is missing a PubKey value")
	}

	return
}
