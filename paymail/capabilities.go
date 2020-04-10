package paymail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
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

// Known capabilities for detecting functionality
const (
	CapabilityBasicAddressResolution = "759684b1a19a"       // Core capability brfc: 759684b1a19a (required from specs)
	CapabilityP2PPaymentDestination  = "2a40af698840"       // Optional - brfc: 2a40af698840 (moneybutton spec)
	CapabilityP2PTransactions        = "5f1323cddf31"       // Optional - brfc: 5f1323cddf31 (moneybutton spec)
	CapabilityPaymentDestination     = "paymentDestination" // Core capability brfc: 759684b1a19a (required from specs)
	CapabilityPayToProtocolPrefix    = "7bd25e5a1fc6"       // Optional - brfc: 7bd25e5a1fc6
	CapabilityPki                    = "pki"                // Core capability brfc: 0c4339ef99c2 (required from specs)
	CapabilityPkiAlternate           = "0c4339ef99c2"       // Core capability brfc: 0c4339ef99c2 (required from specs)
	CapabilityPublicProfile          = "f12f968c92d6"       // Optional - brfc: f12f968c92d6
	CapabilityReceiverApprovals      = "3d7c2ca83a46"       // Optional - brfc: 3d7c2ca83a46
	CapabilitySenderValidation       = "6745385c3fc0"       // Optional - brfc: 6745385c3fc0
	CapabilityVerifyPublicKeyOwner   = "a9f510c16bde"       // Optional - brfc: a9f510c16bde
)

// CapabilitiesResponse is the result returned (plus some custom features)
type CapabilitiesResponse struct {
	BsvAlias              string                 `json:"bsvalias"`                // Version of the bsvalias
	Capabilities          map[string]interface{} `json:"capabilities"`            // Raw list of the capabilities
	P2PPaymentDestination string                 `json:"p2p_payment_destination"` // This is the target url if found
	P2PTransactions       string                 `json:"p2p_transactions"`        // This is the target url if found
	PaymentDestination    string                 `json:"payment_destination"`     // This is the target url if found
	PayToProtocolPrefix   bool                   `json:"pay_to_protocol_prefix"`  // This is the flag if the feature is enabled (client side only)
	Pki                   string                 `json:"pki"`                     // This is the target url if found
	PublicProfile         string                 `json:"public_profile"`          // This is the target url if found
	ReceiverApprovals     string                 `json:"receiver_approvals"`      // This is the target url if found
	SenderValidation      bool                   `json:"sender_validation"`       // This is the flag if the feature is enforced
	VerifyPublicKeyOwner  string                 `json:"verify_public_key_owner"` // This is the target url if found
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

	// Loop the capabilities and set the flags/urls for each detected feature
	for key, val := range capabilities.Capabilities {
		valType := reflect.TypeOf(val).String()
		if (key == CapabilityPki || key == CapabilityPkiAlternate) && valType == typeString {
			capabilities.Pki = val.(string)
		} else if (key == CapabilityPaymentDestination || key == CapabilityBasicAddressResolution) && valType == typeString {
			capabilities.PaymentDestination = val.(string)
		} else if key == CapabilitySenderValidation && valType == typeBool {
			capabilities.SenderValidation = val.(bool)
		} else if key == CapabilityReceiverApprovals && valType == typeString {
			capabilities.ReceiverApprovals = val.(string)
		} else if key == CapabilityVerifyPublicKeyOwner && valType == typeString {
			capabilities.VerifyPublicKeyOwner = val.(string)
		} else if key == CapabilityPublicProfile && valType == typeString {
			capabilities.PublicProfile = val.(string)
		} else if key == CapabilityP2PTransactions && valType == typeString {
			capabilities.P2PTransactions = val.(string)
		} else if key == CapabilityP2PPaymentDestination && valType == typeString {
			capabilities.P2PPaymentDestination = val.(string)
		} else if key == CapabilityPayToProtocolPrefix {
			capabilities.PayToProtocolPrefix = true
		}
	}

	return
}
