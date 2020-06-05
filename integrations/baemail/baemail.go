/*
Package baemail interfaces with Deggen's Baemail
Read more at: https://baemail.me/
*/
package baemail

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Defaults for baemail package
const (
	defaultGetTimeout = 15                   // In seconds
	defaultUserAgent  = "go:baemail"         // Default user agent
	baemailURL        = "https://baemail.me" // Network to use
)

// Override the package defaults
var (
	Network   = baemailURL       // override the default network
	UserAgent = defaultUserAgent // override the default user agent
)

// Response is the standard fields returned on all responses
type Response struct {
	ComposeURL string          `json:"compose_url"` // Compose email url
	Found      bool            `json:"found"`       // Flag if the profile was found
	StatusCode int             `json:"status_code"` // Status code returned on the request
	Tracing    resty.TraceInfo `json:"tracing"`     // Trace information if enabled on the request
}

// HasProfile will check if a profile exists for the given paymail address
// Specs: (no docs)
func HasProfile(alias, domain string, tracing bool) (response *Response, err error) {

	// Set the url for the request
	reqURL := fmt.Sprintf("%s/api/exists/%s@%s", Network, alias, domain)

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

	// Check for a successful status code
	response.StatusCode = resp.StatusCode()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		err = fmt.Errorf("bad response from baemail provider: %d", response.StatusCode)
		return
	}

	// Test the response
	response.Found = string(resp.Body()) == "1"
	if response.Found {
		response.ComposeURL = Compose(alias, domain)
	}

	return
}

// Compose will return a url for composing a baemail
// Specs: https://baemail.me/compose?to=user%40domain
func Compose(alias, domain string) (url string) {
	return fmt.Sprintf("%s/compose?to=%s@%s", Network, alias, domain)
}
