package paymail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

// CapabilitiesResponse is the result returned (plus some custom features)
type CapabilitiesResponse struct {
	BsvAlias     string                 `json:"bsvalias"`     // Version of the bsvalias
	Capabilities map[string]interface{} `json:"capabilities"` // Raw list of the capabilities
}

// Has will check if a BRFC ID is found in the list of capabilities
func (c *CapabilitiesResponse) Has(brfcID, alternate string) bool {
	for key, _ := range c.Capabilities {
		if key == brfcID || (len(alternate) > 0 && key == alternate) {
			return true
		}
	}
	return false
}

// GetValue will return the value (if found) from the capability (url or bool)
// Alternate is used for IE: pki (it breaks convention of using the BRFC ID)
func (c *CapabilitiesResponse) GetValue(brfcID, alternate string) (bool, interface{}) {
	for key, val := range c.Capabilities {
		if key == brfcID || (len(alternate) > 0 && key == alternate) {
			return true, val
		}
	}
	return false, nil
}

// GetValueString will perform GetValue but cast to a string if found
func (c *CapabilitiesResponse) GetValueString(brfcID, alternate string) string {
	if ok, val := c.GetValue(brfcID, alternate); ok {
		return val.(string)
	}
	return ""
}

// GetValueBool will perform GetValue but cast to a bool if found (not found: false)
func (c *CapabilitiesResponse) GetValueBool(brfcID, alternate string) bool {
	if ok, val := c.GetValue(brfcID, alternate); ok {
		return val.(bool)
	}
	return false
}

// GetCapabilities will return a list of capabilities for a given domain & port
// Specs: http://bsvalias.org/02-02-capability-discovery.html
func GetCapabilities(target string, port int) (capabilities *CapabilitiesResponse, err error) {

	// Set the base url and path
	// https://<host-discovery-target>:<host-discovery-port>/.well-known/bsvalias
	reqURL := fmt.Sprintf("https://%s:%d/.well-known/bsvalias", target, port)

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
	if err = json.NewDecoder(resp.Body).Decode(&capabilities); err != nil {
		return
	}

	// Invalid version?
	if len(capabilities.BsvAlias) == 0 {
		err = fmt.Errorf("missing bsvalias version")
		return
	}

	return
}
