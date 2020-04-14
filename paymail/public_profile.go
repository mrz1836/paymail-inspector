package paymail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

/*
Default:

{
    "avatar": "https://<domain><image>",
    "name": "<name>"
}
*/

// PublicProfileResponse is the result returned
type PublicProfileResponse struct {
	Avatar string `json:"avatar"` // A URL that returns a 180x180 image. It can accept an optional parameter `s` to return an image of width and height `s`. The image should be JPEG, PNG, or GIF.
	Name   string `json:"name"`   // A string up to 100 characters long. (name or nickname)
}

// GetPublicProfile will return a valid public profile
// Specs: https://github.com/bitcoin-sv-specs/brfc-paymail/pull/7/files
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

	// Set the client
	client := http.Client{
		Timeout: defaultGetTimeout * time.Second,
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
		err = fmt.Errorf("bad response from paymail provider: %d", resp.StatusCode)
		return
	}

	// Try and decode the response
	err = json.NewDecoder(resp.Body).Decode(&profile)

	return
}
