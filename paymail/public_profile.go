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
    "name": "<name>",
    "avatar": "https://<domain><image>"
}
*/

// PublicProfileResponse is the result returned
type PublicProfileResponse struct {
	Avatar string `json:"avatar"` // The image url
	Name   string `json:"name"`   // Name associated to paymail
}

// GetPublicProfile will return a valid public profile
// Specs: "unlisted" // todo: add specs once they are found
func GetPublicProfile(publicProfileUrl, alias, domain string) (profile *PublicProfileResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/public-profile/{alias}@{domain.tld}
	reqURL := strings.Replace(strings.Replace(publicProfileUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

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
	err = json.NewDecoder(resp.Body).Decode(&profile)

	return
}
