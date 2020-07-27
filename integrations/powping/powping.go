/*
Package powping interfaces with Unwriter's PowPing.com
Read more at: https://powping.com/about
*/
package powping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Defaults for powping package
const (
	defaultGetTimeout = 15                     // In seconds
	defaultUserAgent  = "go:powping"           // Default user agent
	powPingURL        = "https://powping.com/" // Network to use
)

// Override the package defaults
var (
	Network   = powPingURL       // override the default network
	UserAgent = defaultUserAgent // override the default user agent
)

// Response is the response from fetching a profile
type Response struct {
	Profile    *Profile        `json:"profile"`     // The roundesk profile data
	StatusCode int             `json:"status_code"` // Status code returned on the request
	Tracing    resty.TraceInfo `json:"tracing"`     // Trace information if enabled on the request
}

// Profile is the public profile information for a given paymail
type Profile struct {
	Username string `json:"username"`
}

// GetProfile will get a powping profile if it exists for the given paymail address
// Specs: https://powping.com/about
func GetProfile(alias, domain string, tracing bool) (response *Response, err error) {

	// Set the url for the request
	reqURL := fmt.Sprintf("%su?paymail=%s@%s", Network, alias, domain)

	// Create a Client and start the request
	client := resty.New().SetTimeout(defaultGetTimeout * time.Second)
	var resp *resty.Response
	req := client.R().SetHeader("User-Agent", UserAgent)
	if tracing {
		req.EnableTrace()
	}
	if resp, err = req.Get(reqURL); err != nil {
		return
	}

	// Start the response
	response = new(Response)

	// Tracing enabled?
	if tracing {
		response.Tracing = resp.Request.TraceInfo()
	}

	// Test for a successful status code
	response.StatusCode = resp.StatusCode()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		if response.StatusCode != http.StatusNotFound {
			err = fmt.Errorf("bad response from powping: %d", response.StatusCode)
		}

		return
	}

	// No result
	if string(resp.Body()) == "null" {
		return
	}

	// Decode the body of the response
	err = json.Unmarshal(resp.Body(), &response.Profile)

	return
}
