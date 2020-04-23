/*
Package bitpic interfaces with unwriter's bitpic.network
Read more at: https://bitpic.network/about
*/
package bitpic

import (
	"fmt"
	"time"

	"gopkg.in/resty.v1"
)

// Defaults for bitpic functions
const (
	bitPicURL         = "https://bitpic.network" // Network to use
	defaultGetTimeout = 15                       // In seconds
	defaultUserAgent  = "go:bitpic"              // Default user agent
)

// Override defaults
var (
	DefaultImage string             // custom default image (if no image is found)
	Network      = bitPicURL        // override the default network
	UserAgent    = defaultUserAgent // override the default user agent
)

// HasPic will check if a bitpic exists for the given paymail address
// Specs: https://bitpic.network/about
func HasPic(alias, domain string) (found bool, err error) {

	// Set the url for the request
	reqURL := fmt.Sprintf("%s/exists/%s@%s", Network, alias, domain)

	// Create a Client and start the request
	client := resty.New().SetTimeout(defaultGetTimeout * time.Second)
	var resp *resty.Response
	req := client.R().SetHeader("User-Agent", UserAgent)
	if resp, err = req.Get(reqURL); err != nil {
		return
	}

	// Test the response
	if string(resp.Body()) == "1" {
		found = true
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
