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
    "avatar": "https://<domain><image>",
    "name": "<name>"
}
*/

// PublicProfileResponse is the result returned
type PublicProfileResponse struct {
	StandardResponse
	Avatar string `json:"avatar"` // A URL that returns a 180x180 image. It can accept an optional parameter `s` to return an image of width and height `s`. The image should be JPEG, PNG, or GIF.
	Name   string `json:"name"`   // A string up to 100 characters long. (name or nickname)
}

// GetPublicProfile will return a valid public profile
// Specs: https://github.com/bitcoin-sv-specs/brfc-paymail/pull/7/files
func GetPublicProfile(publicProfileUrl, alias, domain string, tracing bool) (response *PublicProfileResponse, err error) {

	// Set the base url and path (assuming the url is from the GetCapabilities request)
	// https://<host-discovery-target>/public-profile/{alias}@{domain.tld}
	reqURL := strings.Replace(strings.Replace(publicProfileUrl, "{alias}", alias, -1), "{domain.tld}", domain, -1)

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
	response = new(PublicProfileResponse)

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
	err = json.Unmarshal(resp.Body(), &response)

	return
}
