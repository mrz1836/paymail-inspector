package paymail

import (
	"encoding/json"
	"fmt"
	"net/http"
)

/*
Default:

{
  "bsvalias": "1.0",
  "capabilities": {
	"pki": "https://bsvalias.example.org/{alias}@{domain.tld}/id",
	"paymentDestination": "https://bsvalias.example.org/{alias}@{domain.tld}/payment-destination"
  }
}
*/

// CapabilitiesResponse is the result returned
type CapabilitiesResponse struct {
	BsvAlias     string                 `json:"bsvalias"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// GetCapabilities will return a list of capabilities for a given domain & port
// Specs: http://bsvalias.org/02-02-capability-discovery.html
func GetCapabilities(target string, port int) (capabilities *CapabilitiesResponse, err error) {

	// Set the base url and path
	// https://<host-discovery-target>:<host-discovery-port>/.well-known/bsvalias
	reqURL := fmt.Sprintf("https://%s:%d/.well-known/bsvalias", target, port)

	// Start the request
	var req *http.Request
	if req, err = http.NewRequest("GET", reqURL, nil); err != nil {
		return
	}

	// Set the headers (standard user agent so it cannot be blocked)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36")

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
	if resp.StatusCode != 200 && resp.StatusCode != 304 {
		err = fmt.Errorf("bad response from paymail provider: %d", resp.StatusCode)
		return
	}

	// Try and decode the response
	err = json.NewDecoder(resp.Body).Decode(&capabilities)

	return
}
