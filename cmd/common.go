package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/ryanuber/columnize"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

// RandomHex returns a random hex string and error
func RandomHex(n int) (hexString string, err error) {
	b := make([]byte, n)
	if _, err = rand.Read(b); err != nil {
		return
	}
	return hex.EncodeToString(b), nil
}

// getPki will get a pki response (logging and basic error handling)
func getPki(url, alias, domain string) (pki *paymail.PKIResponse, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Retrieving public key information for %s...", chalk.Cyan.Color(alias+"@"+domain)))

	// Get the PKI for the given address
	if pki, err = paymail.GetPKI(url, alias, domain, !skipTracing); err != nil {
		return
	}

	// No pubkey found
	if len(pki.PubKey) == 0 {
		err = fmt.Errorf("failed getting pubkey for: %s@%s", alias, domain)
		return
	}

	// Possible invalid pubkey
	if len(pki.PubKey) != paymail.PubKeyLength {
		chalker.Log(chalker.WARN, fmt.Sprintf("PubKey length is: %d, expected: %d", len(pki.PubKey), paymail.PubKeyLength))
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(pki.Tracing, pki.StatusCode)
	}

	// Success
	if len(pki.PubKey) > 0 {
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("Found pubkey %s...", pki.PubKey[:10]))
	}

	return
}

// getSrvRecord will return an srv record, and optional validation
func getSrvRecord(domain string, validate bool) (srv *net.SRV, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Retrieving SRV record for %s...", chalk.Cyan.Color(domain)))

	// Get the record
	if srv, err = paymail.GetSRVRecord(serviceName, protocol, domain, nameServer); err != nil {
		return
	}

	// Run validation on the SRV record?
	if validate {
		if srv == nil {
			err = fmt.Errorf("missing SRV record for: %s", domain)
			return
		}

		// Validate the SRV record for the domain name (using all flags or default values)
		if err = paymail.ValidateSRVRecord(srv, nameServer, port, priority, weight); err != nil {
			err = fmt.Errorf("validation error: %s", err.Error())
			return
		}

		// Validation good
		chalker.Log(chalker.SUCCESS, "SRV record passed all validations (target, port, priority, weight)")
	}

	if srv != nil {
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("SRV target: %s:%d --weight %d --priority %d", srv.Target, srv.Port, srv.Weight, srv.Priority))
	}

	return
}

// getCapabilities will check SRV first, then attempt default domain:port check (logging and basic error handling)
func getCapabilities(domain string) (capabilities *paymail.CapabilitiesResponse, err error) {

	capabilityDomain := ""
	capabilityPort := paymail.DefaultPort

	// Get the details from the SRV record
	var srv *net.SRV
	if srv, err = getSrvRecord(domain, false); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("retrieving SRV record failed: %s", err.Error()))
		capabilityDomain = domain
	} else if srv != nil {
		capabilityDomain = srv.Target
		capabilityPort = int(srv.Port)
	}

	// Get the capabilities for the given target domain
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Retrieving available capabilities for %s...", chalk.Cyan.Color(fmt.Sprintf("%s:%d", capabilityDomain, capabilityPort))))
	if capabilities, err = paymail.GetCapabilities(capabilityDomain, capabilityPort, !skipTracing); err != nil {
		return
	}

	// Check the version
	if capabilities.BsvAlias != viper.GetString(flagBsvAlias) {
		err = fmt.Errorf("capabilities %s version mismatch, expected: %s but got: %s", flagBsvAlias, chalk.Cyan.Color(viper.GetString(flagBsvAlias)), chalk.Magenta.Color(capabilities.BsvAlias))
		return
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(capabilities.Tracing, capabilities.StatusCode)
	}

	// Success
	chalker.Log(chalker.SUCCESS, fmt.Sprintf("Found [%d] capabilities", len(capabilities.Capabilities)))

	return
}

// resolveAddress will resolve an address (logging and basic error handling)
func resolveAddress(url, alias, domain, senderHandle, signature, purpose string, amount uint64) (response *paymail.AddressResolutionResponse, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Resolving address for %s...", chalk.Cyan.Color(alias+"@"+domain)))

	// Create the address resolution request
	if response, err = paymail.AddressResolution(
		url,
		alias,
		domain,
		&paymail.AddressResolutionRequest{
			Amount:       amount,
			Dt:           time.Now().UTC().Format(time.RFC3339), // UTC is assumed
			Purpose:      purpose,
			SenderHandle: senderHandle,
			SenderName:   viper.GetString(flagSenderName),
			Signature:    signature,
		},
		!skipTracing,
	); err != nil {
		return
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(response.Tracing, response.StatusCode)
	}

	// Success
	if len(response.Address) > 0 {
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("Found address %s...", response.Address[:10]))
	}

	return
}

// getP2PPaymentDestination will start a new p2p transaction request (logging and basic error handling)
func getP2PPaymentDestination(url, alias, domain string, satoshis uint64) (response *paymail.P2PPaymentDestinationResponse, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Starting new P2P payment request for %s...", chalk.Cyan.Color(alias+"@"+domain)))

	// Create the address resolution request
	if response, err = paymail.GetP2PPaymentDestination(
		url,
		alias,
		domain,
		&paymail.P2PPaymentDestinationRequest{Satoshis: satoshis},
		!skipTracing,
	); err != nil {
		return
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(response.Tracing, response.StatusCode)
	}

	// Success
	if len(response.Outputs) > 0 {
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("Found [%d] payment output(s)", len(response.Outputs)))
	}

	return
}

// getPublicProfile will get a public profile (logging and basic error handling)
func getPublicProfile(url, alias, domain string) (profile *paymail.PublicProfileResponse, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Retrieving public profile for %s...", chalk.Cyan.Color(alias+"@"+domain)))

	// Get the profile
	if profile, err = paymail.GetPublicProfile(url, alias, domain, !skipTracing); err != nil {
		return
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(profile.Tracing, profile.StatusCode)
	}

	// Success
	if len(profile.Name) > 0 && len(profile.Avatar) > 0 {
		chalker.Log(chalker.SUCCESS, "Valid profile found [name, avatar]")
	}

	return
}

// verifyPubKey will verify a given pubkey against a paymail address (logging and basic error handling)
func verifyPubKey(url, alias, domain, pubKey string) (response *paymail.VerifyPubKeyResponse, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("verifing pubkey for %s...", chalk.Cyan.Color(alias+"@"+domain)))

	// Verify the given pubkey
	if response, err = paymail.VerifyPubKey(url, alias, domain, pubKey, !skipTracing); err != nil {
		return
	} else if response == nil {
		err = fmt.Errorf("failed getting verification response")
		return
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(response.Tracing, response.StatusCode)
	}

	return
}

// validatePaymailAndDomain will do a basic validation on the paymail format
func validatePaymailAndDomain(paymailAddress, domain string) (valid bool) {

	// Validate the format for the paymail address (paymail addresses follow conventional email requirements)
	if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("Paymail address failed format validation: %s", err.Error()))
		return
	} else if !ok {
		chalker.Log(chalker.ERROR, "Paymail address failed format validation: unknown reason")
		return
	}

	// Check for a real domain (require at least one period)
	if !strings.Contains(domain, ".") {
		chalker.Log(chalker.ERROR, fmt.Sprintf("Domain name is invalid: %s", domain))
		return
	} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
		chalker.Log(chalker.ERROR, fmt.Sprintf("Domain name failed DNS check: %s", domain))
		return
	}

	valid = true
	return
}

// displayTracingResults displays the tracing results into the terminal per request
func displayTracingResults(tracing resty.TraceInfo, statusCode int) {

	// Add the network time columns
	output := []string{
		fmt.Sprintf(`DNSLookup | %s | TTFB | %s`, tracing.DNSLookup.String(), tracing.ServerTime.String()),
		fmt.Sprintf(`TLSHandshake | %s | ConnTime | %s`, tracing.TLSHandshake.String(), tracing.ConnTime.String()),
		fmt.Sprintf(`FB to Close  | %s | %s | %s [%d]`, tracing.ResponseTime.String(), "TotalTime", tracing.TotalTime.String(), statusCode),
	}

	// Connection was idle?
	if tracing.IsConnWasIdle {
		output = append(output,
			fmt.Sprintf(`IsConnWasIdle | %s | ConnIdleTime | %s`,
				chalk.Magenta.Color(fmt.Sprintf("%v", tracing.IsConnWasIdle)),
				chalk.Magenta.Color(tracing.ConnIdleTime.String()),
			))
	}

	// Connection reused?
	if tracing.IsConnReused {
		output = append(output, fmt.Sprintf(`IsConnReused | %s `, chalk.Magenta.Color(fmt.Sprintf("%v", tracing.IsConnReused))))
	}

	// Render the data
	fmt.Println(columnize.SimpleFormat(output))
}

// displayHeader will display a standard header
func displayHeader(level, text string) {
	chalker.Log(level, "\n==========| "+text)
}
