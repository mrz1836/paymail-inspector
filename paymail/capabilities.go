package paymail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
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
	StandardResponse
	BsvAlias     string                 `json:"bsvalias"`     // Version of the bsvalias
	Capabilities map[string]interface{} `json:"capabilities"` // Raw list of the capabilities
}

// Has will check if a BRFC ID (or alternate) is found in the list of capabilities
func (c *CapabilitiesResponse) Has(brfcID, alternate string) bool {
	for key := range c.Capabilities {
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
func GetCapabilities(target string, port int, tracing bool) (response *CapabilitiesResponse, err error) {

	// Set the base url and path
	// https://<host-discovery-target>:<host-discovery-port>/.well-known/bsvalias
	reqURL := fmt.Sprintf("https://%s:%d/.well-known/bsvalias", target, port)

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

	// New struct
	response = new(CapabilitiesResponse)

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
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return
	}

	// Invalid version?
	if len(response.BsvAlias) == 0 {
		err = fmt.Errorf("missing bsvalias version")
	}

	return
}
