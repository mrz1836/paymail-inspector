package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/database"
	twopaymail "github.com/mrz1836/paymail-inspector/integrations/2paymail"
	"github.com/mrz1836/paymail-inspector/integrations/bitpic"
	"github.com/mrz1836/paymail-inspector/integrations/roundesk"
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
func getPki(url, alias, domain string, allowCache bool) (pki *paymail.PKIResponse, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Retrieving public key information for %s...", chalk.Cyan.Color(alias+"@"+domain)))

	// Cache key
	keyName := "model-pki-" + alias + "@" + domain

	// Do we have cache and db?
	if !disableCache && databaseEnabled && allowCache {
		var jsonStr string
		if jsonStr, err = database.Get(keyName); err != nil {
			return
		}
		if len(jsonStr) > 0 {
			if err = json.Unmarshal([]byte(jsonStr), &pki); err != nil {
				return
			}
			chalker.Log(chalker.SUCCESS, fmt.Sprintf("Found pubkey %s... (from cache)", pki.PubKey[:10]))
			return
		}
	}

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

		// Store in db?
		if databaseEnabled {
			var jsonStr []byte
			if jsonStr, err = json.Marshal(pki); err != nil {
				return
			}
			if err = database.Set(keyName, string(jsonStr), 1*time.Hour); err != nil {
				return
			}
		}
	}

	return
}

// getSrvRecord will return an srv record, and optional validation
func getSrvRecord(domain string, validate bool, allowCache bool) (srv *net.SRV, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Retrieving SRV record for %s...", chalk.Cyan.Color(domain)))

	// Cache key
	keyName := "model-srv-" + domain

	// Do we have cache and db?
	if !disableCache && databaseEnabled && allowCache {
		var jsonStr string
		if jsonStr, err = database.Get(keyName); err != nil {
			return
		}
		if len(jsonStr) > 0 {
			if err = json.Unmarshal([]byte(jsonStr), &srv); err != nil {
				return
			}
			chalker.Log(chalker.SUCCESS, fmt.Sprintf("SRV target: %s:%d --weight %d --priority %d (from cache)", srv.Target, srv.Port, srv.Weight, srv.Priority))
			return
		}
	}

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

		// Store in db?
		if databaseEnabled {
			var jsonStr []byte
			if jsonStr, err = json.Marshal(srv); err != nil {
				return
			}
			if err = database.Set(keyName, string(jsonStr), 1*time.Hour); err != nil {
				return
			}
		}
	}

	return
}

// getCapabilities will check SRV first, then attempt default domain:port check (logging and basic error handling)
func getCapabilities(domain string, allowCache bool) (capabilities *paymail.CapabilitiesResponse, err error) {

	capabilityDomain := ""
	capabilityPort := paymail.DefaultPort

	// Get the details from the SRV record
	var srv *net.SRV
	if srv, err = getSrvRecord(domain, false, allowCache); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("retrieving SRV record failed: %s", err.Error()))
		capabilityDomain = domain
	} else if srv != nil {
		capabilityDomain = srv.Target
		capabilityPort = int(srv.Port)
	}

	// Get the capabilities for the given target domain
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Retrieving available capabilities for %s...", chalk.Cyan.Color(fmt.Sprintf("%s:%d", capabilityDomain, capabilityPort))))

	// Cache key
	keyName := "model-capabilities-" + domain

	// Do we have cache and db?
	if !disableCache && databaseEnabled && allowCache {
		var jsonStr string
		if jsonStr, err = database.Get(keyName); err != nil {
			return
		}
		if len(jsonStr) > 0 {
			if err = json.Unmarshal([]byte(jsonStr), &capabilities); err != nil {
				return
			}
			chalker.Log(chalker.SUCCESS, fmt.Sprintf("Found [%d] capabilities (from cache)", len(capabilities.Capabilities)))
			return
		}
	}

	// Look up the capabilities
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

	// Store in db?
	if databaseEnabled {
		var jsonStr []byte
		if jsonStr, err = json.Marshal(capabilities); err != nil {
			return
		}
		if err = database.Set(keyName, string(jsonStr), 1*time.Hour); err != nil {
			return
		}
	}

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
func getPublicProfile(url, alias, domain string, allowCache bool) (profile *paymail.PublicProfileResponse, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Retrieving public profile for %s...", chalk.Cyan.Color(alias+"@"+domain)))

	// Cache key
	keyName := "model-public-profile-" + alias + "@" + domain

	// Do we have cache and db?
	if !disableCache && databaseEnabled && allowCache {
		var jsonStr string
		if jsonStr, err = database.Get(keyName); err != nil {
			return
		}
		if len(jsonStr) > 0 {
			if err = json.Unmarshal([]byte(jsonStr), &profile); err != nil {
				return
			}
			chalker.Log(chalker.SUCCESS, "Valid profile found [name, avatar] (from cache)")
			return
		}
	}

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

		// Store in db?
		if databaseEnabled {
			var jsonStr []byte
			if jsonStr, err = json.Marshal(profile); err != nil {
				return
			}
			if err = database.Set(keyName, string(jsonStr), 1*time.Hour); err != nil {
				return
			}
		}
	}

	return
}

// getBitPic will get a bitpic if the pic exists
func getBitPic(alias, domain string, allowCache bool) (url string, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Checking %s for a Bitpic...", chalk.Cyan.Color(alias+"@"+domain)))

	// Cache key
	keyName := "app-bitpic-" + alias + "@" + domain

	// Do we have caching and db?
	if !disableCache && databaseEnabled && allowCache {
		if url, err = database.Get(keyName); err != nil {
			return
		}
		if len(url) > 0 {
			chalker.Log(chalker.SUCCESS, "Bitpic was found for "+alias+"@"+domain+" (from cache)")
			return
		}
	}

	// Does this paymail have a bitpic profile?
	var resp *bitpic.Response
	if resp, err = bitpic.GetPic(alias, domain, !skipTracing); err != nil {
		return
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(resp.Tracing, resp.StatusCode)
	}

	// Checks if the response was good
	if resp != nil && resp.Found {
		url = resp.URL
		chalker.Log(chalker.SUCCESS, "Bitpic was found for "+alias+"@"+domain)

		// Store in db?
		if databaseEnabled {
			if err = database.Set(keyName, url, 1*time.Hour); err != nil {
				return
			}
		}
	} else {
		chalker.Log(chalker.DEFAULT, "Bitpic was not found")
	}

	return
}

// getRoundeskProfile will get a Roundesk profile if it exists
func getRoundeskProfile(alias, domain string, allowCache bool) (profile *roundesk.Response, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Checking %s for a Roundesk profile...", chalk.Cyan.Color(alias+"@"+domain)))

	// Cache key
	keyName := "app-roundesk-" + alias + "@" + domain

	// Do we have caching and db?
	if !disableCache && databaseEnabled && allowCache {
		var jsonStr string
		if jsonStr, err = database.Get(keyName); err != nil {
			return
		}
		if len(jsonStr) > 0 {
			if err = json.Unmarshal([]byte(jsonStr), &profile); err != nil {
				return
			}
			chalker.Log(chalker.SUCCESS, "Roundesk profile was found (from cache)")
			return
		}
	}

	// Find a roundesk profile
	if profile, err = roundesk.GetProfile(alias, domain, !skipTracing); err != nil {
		return
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(profile.Tracing, profile.StatusCode)
	}

	// Success or failure
	if profile != nil && profile.Profile != nil && len(profile.Profile.ID) > 0 {
		chalker.Log(chalker.SUCCESS, "Roundesk profile was found")

		// Store in db?
		if databaseEnabled {
			var jsonStr []byte
			if jsonStr, err = json.Marshal(profile); err != nil {
				return
			}
			if err = database.Set(keyName, string(jsonStr), 1*time.Hour); err != nil {
				return
			}
		}

	} else {
		chalker.Log(chalker.DEFAULT, "Roundesk profile was not found")
	}

	return
}

// get2paymail will get a 2paymail account if it exists
func get2paymail(alias, domain string, allowCache bool) (profile *twopaymail.Response, err error) {

	// Start the request
	displayHeader(chalker.DEFAULT, fmt.Sprintf("Checking %s for a 2paymail...", chalk.Cyan.Color(alias+"@"+domain)))

	// Cache key
	keyName := "app-2paymail-" + alias + "@" + domain

	// Do we have caching and db?
	if !disableCache && databaseEnabled && allowCache {
		var jsonStr string
		if jsonStr, err = database.Get(keyName); err != nil {
			return
		}
		if len(jsonStr) > 0 {
			if err = json.Unmarshal([]byte(jsonStr), &profile); err != nil {
				return
			}
			chalker.Log(chalker.SUCCESS, "2paymail was found for "+alias+"@"+domain+" (from cache)")
			return
		}
	}

	// Does this paymail have a profile?
	if profile, err = twopaymail.GetAccount(alias, domain, !skipTracing); err != nil {
		return
	}

	// Display the tracing results
	if !skipTracing {
		displayTracingResults(profile.Tracing, profile.StatusCode)
	}

	// Checks if the response was good
	if profile != nil && profile.Found {
		chalker.Log(chalker.SUCCESS, "2paymail was found for "+alias+"@"+domain)

		// Store in db?
		if databaseEnabled {
			var jsonStr []byte
			if jsonStr, err = json.Marshal(profile); err != nil {
				return
			}
			if err = database.Set(keyName, string(jsonStr), 1*time.Hour); err != nil {
				return
			}
		}
	} else {
		chalker.Log(chalker.DEFAULT, "2paymail was not found")
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
	chalker.Log(chalker.DIM, columnize.SimpleFormat(output))
}

// displayHeader will display a standard header
func displayHeader(level, text string) {
	chalker.Log(level, "\n==========| "+text)
}

// GetPublicInfo will get all the public info for a given paymail
func (p *PaymailDetails) GetPublicInfo(capabilities *paymail.CapabilitiesResponse) (err error) {

	// Requirements
	if len(p.Handle) == 0 {
		err = fmt.Errorf("missing required field: %s", "Handle")
		return
	} else if p.Provider == nil {
		err = fmt.Errorf("missing required field: %s", "Provider")
		return
	}

	// Attempt to get a public profile if the capability is found
	url := capabilities.GetValueString(paymail.BRFCPublicProfile, "")
	if len(url) > 0 && !skipPublicProfile && p.PKI != nil && len(p.PKI.Handle) > 0 {
		if p.PublicProfile, err = getPublicProfile(url, p.Handle, p.Provider.Domain, true); err != nil {
			err = fmt.Errorf("get public profile failed: %s", err.Error())
		}
	}

	// Attempt to get a bitpic (if enabled)
	if !skipBitpic && p.PKI != nil && len(p.PKI.Handle) > 0 {
		if p.Bitpic, err = getBitPic(p.Handle, p.Provider.Domain, true); err != nil {
			err = fmt.Errorf("checking for bitpic failed: %s", err.Error())
		}
	}

	// Attempt to get a 2paymail (if enabled)
	if !skip2paymail && p.PKI != nil && len(p.PKI.Handle) > 0 {
		if p.TwoPaymail, err = get2paymail(p.Handle, p.Provider.Domain, true); err != nil {
			err = fmt.Errorf("checking for 2paymail failed: %s", err.Error())
		}
	}

	// Attempt to get a Roundesk profile (if enabled)
	if !skipRoundesk && p.PKI != nil && len(p.PKI.Handle) > 0 {
		if p.Roundesk, err = getRoundeskProfile(p.Handle, p.Provider.Domain, true); err != nil {
			err = fmt.Errorf("checking for roundesk profile failed: %s", err.Error())
		}
	}

	return
}

// Paymail returns the paymail address from the details struct
func (p *PaymailDetails) Paymail() string {
	if p.PKI != nil && len(p.PKI.Handle) > 0 {
		return p.PKI.Handle
	} else if p.Provider != nil && len(p.Handle) > 0 {
		return p.Handle + "@" + p.Provider.Domain
	}
	return ""
}

// Display all the paymail results for a given paymail search/resolution
func (p *PaymailDetails) Display() {

	displayPaymail := p.Paymail()

	// Rendering profile information
	displayHeader(chalker.BOLD, fmt.Sprintf("Results for %s", chalk.Cyan.Color(displayPaymail)))

	// No PKI - then we don't have a paymail
	if p.PKI == nil || len(p.PKI.PubKey) == 0 {
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("The handle: %s might be available! Reserve it now: %s", p.Handle, p.Provider.Link))
		return
	}

	// Display the public profile if found
	if p.PublicProfile != nil {
		if len(p.PublicProfile.Name) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Name         : %s", chalk.Cyan.Color(p.PublicProfile.Name)))
		}
		if len(p.PublicProfile.Avatar) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Avatar       : %s", chalk.Cyan.Color(p.PublicProfile.Avatar)))
		}
	}

	// Display bitpic if found
	if len(p.Bitpic) > 0 {
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Bitpic       : %s", chalk.Cyan.Color(p.Bitpic)))
	}

	// Display 2paymail if found
	if p.TwoPaymail != nil && len(p.TwoPaymail.URL) > 0 {
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("2paymail     : %s", chalk.Cyan.Color(p.TwoPaymail.URL)))
		if len(p.TwoPaymail.TX) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("2paymail TX  : %s", chalk.Cyan.Color(p.TwoPaymail.TX)))
		}
		if len(p.TwoPaymail.Twitter) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Twitter      : %s", chalk.Cyan.Color(p.TwoPaymail.Twitter)))
		}
	}

	// Show pubkey
	if p.PKI != nil && len(p.PKI.PubKey) > 0 {
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("PubKey       : %s", chalk.Cyan.Color(p.PKI.PubKey)))
	}

	// Show address resolution details
	if p.Resolution != nil && len(p.Resolution.Address) > 0 {
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Output Script: %s", chalk.Cyan.Color(p.Resolution.Output)))
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Address      : %s", chalk.Cyan.Color(p.Resolution.Address)))

		// If we have a signature
		if len(p.Resolution.Signature) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Signature    : %s", chalk.Cyan.Color(p.Resolution.Signature)))
		}
	}

	// Display the roundesk profile if found
	if p.Roundesk != nil && p.Roundesk.Profile != nil {

		// Rendering profile information
		displayHeader(chalker.DEFAULT, fmt.Sprintf("Roundesk profile for %s", chalk.Cyan.Color(displayPaymail)))

		if len(p.Roundesk.Profile.Name) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Name      : %s", chalk.Cyan.Color(p.Roundesk.Profile.Name)))
		}
		if len(p.Roundesk.Profile.Headline) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Headline  : %s", chalk.Cyan.Color(p.Roundesk.Profile.Headline)))
		}
		if len(p.Roundesk.Profile.Bio) > 0 {
			p.Roundesk.Profile.Bio = strings.TrimSuffix(p.Roundesk.Profile.Bio, "\n")
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Bio       : %s", chalk.Cyan.Color(p.Roundesk.Profile.Bio)))
		}
		if len(p.Roundesk.Profile.Twetch) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Twetch    : %s", chalk.Cyan.Color("https://twetch.app/u/"+p.Roundesk.Profile.Twetch)))
		}

		chalker.Log(chalker.DEFAULT, fmt.Sprintf("URL       : %s", chalk.Cyan.Color("https://roundesk.co/u/"+displayPaymail)))

		if len(p.Roundesk.Profile.Nonce) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Nonce     : %s", chalk.Cyan.Color(p.Roundesk.Profile.Nonce)))
		}
	}
}
