/*
Package bitpic interfaces with unwriter's bitpic.network
Read more at: https://bitpic.network/about
*/
package bitpic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Defaults for bitpic package
const (
	bitPicURL         = "bitpic.network" // Choose the provider
	defaultGetTimeout = 15               // In seconds
	defaultUserAgent  = "go:bitpic"      // Default user agent
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

// SearchResponse is the response from /search
type SearchResponse struct {
	Result     *SearchResult   `json:"result"`      // Result from BitPics
	StatusCode int             `json:"status_code"` // Status code returned on the request
	Tracing    resty.TraceInfo `json:"tracing"`     // Trace information if enabled on the request
}

// SearchResult is the child of the response
type SearchResult struct {
	Posts []*Post `json:"posts"`
	Query string  `json:"query"`
}

// Post is a BitPic post
type Post struct {
	CreatedAt string      `json:"created_at"`
	Data      *Data       `json:"data"`
	Message   string      `json:"message"`
	Meta      *Meta       `json:"meta"`
	Name      string      `json:"name"`
	Tags      interface{} `json:"tags"`
	TxID      string      `json:"tx_id"`
}

// Data is the required bitpic information
type Data struct {
	BitFs   string `json:"bitfs"`
	Paymail string `json:"paymail"`
	URI     string `json:"uri"`
}

// Meta is additional information
type Meta struct {
	Content     string `json:"content"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Title       string `json:"title"`
	URL         string `json:"url"`
}

// GetPic will check if a bitpic exists for the given paymail address and fetch the url if found
// Specs: https://bitpic.network/about
func GetPic(alias, domain string, tracing bool) (response *Response, err error) {

	// Set the url for the request
	reqURL := fmt.Sprintf("https://%s/exists/%s@%s", Network, alias, domain)

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
		response.URL = URL(alias, domain)
	}

	return
}

// URL will return a url for the bitpic avatar using alias and domain
// Specs: https://bitpic.network/about
func URL(alias, domain string) string {
	if len(DefaultImage) > 0 {
		return fmt.Sprintf("https://%s/u/%s@%s?d=%s", Network, alias, domain, DefaultImage)
	}
	return fmt.Sprintf("https://%s/u/%s@%s", Network, alias, domain)
}

// URLFromPaymail will return a url for the bitpic avatar using a paymail
// Specs: https://bitpic.network/about
func URLFromPaymail(paymail string) string {
	if len(DefaultImage) > 0 {
		return fmt.Sprintf("https://%s/u/%s?d=%s", Network, paymail, DefaultImage)
	}
	return fmt.Sprintf("https://%s/u/%s", Network, paymail)
}

// Search will perform a search on the BitPic network
// https://txt.bitpic.network/search/json?text=alias@domain
func Search(alias, domain string, tracing bool) (response *SearchResponse, err error) {

	// Set the url for the request
	reqURL := fmt.Sprintf("https://txt.%s/search/json?text=%s@%s", bitPicURL, alias, domain)

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
	response = new(SearchResponse)

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

	// No profile result?
	if string(resp.Body()) == "{}" {
		return
	}

	// Decode the body of the response
	err = json.Unmarshal(resp.Body(), &response)

	return
}
