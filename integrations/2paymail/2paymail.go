/*
Package twopaymail interfaces with 2paymail.com
Read more at: https://2paymail.com/login
*/
package twopaymail

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-resty/resty/v2"
)

// Defaults for twopaymail (2paymail) package
const (
	defaultUrl        = "https://2paymail.com" // Network to use
	defaultGetTimeout = 15                     // In seconds
	defaultUserAgent  = "go:2paymail"          // Default user agent
)

// Override the package defaults
var (
	Network   = defaultUrl       // override the default network
	UserAgent = defaultUserAgent // override the default user agent
)

// Response is the standard fields returned on all responses
type Response struct {
	Found      bool            `json:"found"`       // Flag if the bitpic was found
	StatusCode int             `json:"status_code"` // Status code returned on the request
	Tracing    resty.TraceInfo `json:"tracing"`     // Trace information if enabled on the request
	Twitter    string          `json:"twitter"`     // The associated twitter account if found
	TX         string          `json:"tx"`          // The associated tx to validate the association
	URL        string          `json:"url"`         // The bitpic url for the image
}

// GetAccount will check if an account exists for the given paymail
// Specs: https://2paymail.com/profiles
func GetAccount(alias, domain string, tracing bool) (response *Response, err error) {

	// Set the url for the request
	reqURL := fmt.Sprintf("%s/me/%s@%s", Network, alias, domain)

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
		if response.StatusCode == http.StatusNotFound {
			return
		}

		err = fmt.Errorf("bad response from 2paymail provider: %d", response.StatusCode)
		return
	}

	// todo: better detection if the account is present (api req? html parse?)

	// Try to get the twitter account (hack for now)
	var reg *regexp.Regexp
	if reg, err = regexp.Compile(`https://twitter.com/(.*?)/status/`); err != nil {
		return
	}

	// Extract the twitter user name
	if found := reg.FindStringSubmatch(string(resp.Body())); len(found) > 1 {
		response.Twitter = found[1]
	}

	// Try to get the TX for authentication (hack for now)
	if reg, err = regexp.Compile(`https://whatsonchain.com/tx/(.*?)">View Transaction`); err != nil {
		return
	}

	// Extract the TX
	if found := reg.FindStringSubmatch(string(resp.Body())); len(found) > 1 {
		response.TX = found[1]
	}

	// Set the url
	response.Found = true
	response.URL = Url(alias, domain)

	return
}

// Url will return a url for the 2paymail profile
// Specs: https://2paymail.com/me/
func Url(alias, domain string) string {
	return fmt.Sprintf("%s/me/%s@%s", Network, alias, domain)
}
