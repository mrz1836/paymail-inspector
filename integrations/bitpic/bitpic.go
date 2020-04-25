/*
Package bitpic interfaces with unwriter's bitpic.network
Read more at: https://bitpic.network/about
*/
package bitpic

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Defaults for bitpic package
const (
	bitPicURL         = "https://bitpic.network" // Network to use
	defaultGetTimeout = 15                       // In seconds
	defaultUserAgent  = "go:bitpic"              // Default user agent
)

// Override the package defaults
var (
	DefaultImage string             // custom default image (if no image is found)
	Network      = bitPicURL        // override the default network
	UserAgent    = defaultUserAgent // override the default user agent
)

// Response is the standard fields returned on all responses
type Response struct {
	Found      bool            `json:"found"`       // Flag if the bitpic was found
	StatusCode int             `json:"status_code"` // Status code returned on the request
	Tracing    resty.TraceInfo `json:"tracing"`     // Trace information if enabled on the request
	URL        string          `json:"url"`         // The bitpic url for the image
}

// GetPic will check if a bitpic exists for the given paymail address and fetch the url if found
// Specs: https://bitpic.network/about
func GetPic(alias, domain string, tracing bool) (response *Response, err error) {

	// Set the url for the request
	reqURL := fmt.Sprintf("%s/exists/%s@%s", Network, alias, domain)

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
		err = fmt.Errorf("bad response from bitpic provider: %d", response.StatusCode)
		return
	}

	// Test the response
	response.Found = string(resp.Body()) == "1"
	if response.Found {
		response.URL = Url(alias, domain)
	}

	return
}

// Url will return a url for the bitpic avatar
// Specs: https://bitpic.network/about
func Url(alias, domain string) string {
	if len(DefaultImage) > 0 {
		return fmt.Sprintf("%s/u/%s@%s?d=%s", Network, alias, domain, DefaultImage)
	}
	return fmt.Sprintf("%s/u/%s@%s", Network, alias, domain)
}
