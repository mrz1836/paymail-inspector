/*
Package roundesk interfaces with Deggen's Roundesk.co
Read more at: https://roundesk.co/
*/
package roundesk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Defaults for roundesk package
const (
	defaultGetTimeout = 15                         // In seconds
	defaultUserAgent  = "go:roundesk"              // Default user agent
	roundeskURL       = "https://roundesk.co/api/" // Network to use
)

// Override the package defaults
var (
	Network   = roundeskURL      // override the default network
	UserAgent = defaultUserAgent // override the default user agent
)

// Response is the response from fetching a profile
type Response struct {
	Profile    *Profile        `json:"profile"`     // The roundesk profile data
	StatusCode int             `json:"status_code"` // Status code returned on the request
	Tracing    resty.TraceInfo `json:"tracing"`     // Trace information if enabled on the request
}

// Profile is the roundesk public profile
type Profile struct {
	Bio      string  `json:"bio"`
	Dev      float64 `json:"dev"`
	Ent      float64 `json:"ent"`
	Headline string  `json:"headline"`
	Inv      float64 `json:"inv"`
	Mar      float64 `json:"mar"`
	Name     string  `json:"name"`
	Paymail  string  `json:"paymail"`
	Twetch   string  `json:"twetch"`
	Uxd      float64 `json:"uxd"`
}

// GetProfile will get a roundesk profile if it exists for the given paymail address
// Specs: https://roundesk.co/
func GetProfile(alias, domain string, tracing bool) (response *Response, err error) {

	// Set the url for the request
	reqURL := fmt.Sprintf("%su/%s@%s", Network, alias, domain)

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
			err = fmt.Errorf("bad response from roundesk: %d", response.StatusCode)
		}

		return
	}

	// No profile result?
	if string(resp.Body()) == "{}" || string(resp.Body()) == `{"granted":false}` {
		return
	}

	// Decode the body of the response
	err = json.Unmarshal(resp.Body(), &response.Profile)

	// Handle new way of detecting user is not known (Clear out the user data)
	if response.Profile.Name == "Unknown" {
		response.Profile.Paymail = ""
	}

	return
}
