package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/go-resty/resty/v2"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

// Cache key prefixes, one per kind of cached response.
const (
	keyCapabilities  = "capabilities:"
	keyPKI           = "pki:"
	keySRV           = "srv:"
	keyPublicProfile = "public-profile:"
	keyPrefixLength  = 10 // characters of a key or address shown in progress lines
)

// srvOptions are the expected values of a domain's SRV record.
type srvOptions struct {
	service  string
	protocol string
	port     uint16
	priority uint16
	weight   uint16
}

// defaultSRVOptions are the values the bsvalias specification requires.
func defaultSRVOptions() srvOptions {
	return srvOptions{
		service:  paymail.DefaultServiceName,
		protocol: paymail.DefaultProtocol,
		port:     paymail.DefaultPort,
		priority: paymail.DefaultPriority,
		weight:   paymail.DefaultWeight,
	}
}

// session performs the paymail requests of one command, with progress
// output, request tracing, and response caching.
type session struct {
	*app

	client   paymailClient
	srv      srvOptions
	bsvAlias string // expected bsvalias version, read once so lookups can run concurrently
	quiet    bool   // suppress progress and tracing (used by whois)
}

// newSession creates a paymail client for one command. An empty nameServer
// uses the system resolver.
func (a *app) newSession(nameServer string) (*session, error) {
	client, err := a.newClient(clientOptions{
		userAgent:  a.userAgent(),
		nameServer: nameServer,
		tracing:    !a.flags.skipTracing,
	})
	if err != nil {
		return nil, err
	}

	return &session{
		app:      a,
		client:   client,
		srv:      defaultSRVOptions(),
		bsvAlias: a.cfg.GetString(keyBsvAlias),
	}, nil
}

// step announces the request about to run.
func (s *session) step(format string, args ...any) {
	if s.quiet {
		return
	}

	s.out.Header(output.LevelDefault, fmt.Sprintf(format, args...))
}

// progress reports the outcome of a request.
func (s *session) progress(format string, args ...any) {
	if s.quiet {
		return
	}

	s.out.Success(format, args...)
}

// advise reports a non-fatal observation.
func (s *session) advise(format string, args ...any) {
	if s.quiet {
		return
	}

	s.out.Warn(format, args...)
}

// trace prints the timing of a request unless tracing is off.
func (s *session) trace(info resty.TraceInfo, statusCode int) {
	if s.quiet || s.flags.skipTracing {
		return
	}

	renderTracing(s.out, info, statusCode)
}

// cached loads key into dst when caching is allowed and a live entry exists.
func (s *session) cached(key string, allowCache bool, dst any) bool {
	if !allowCache || s.flags.noCache {
		return false
	}

	found, err := s.cacheStore().Get(key, dst)
	if err != nil {
		s.errOut.Dim("Ignoring unreadable cache entry: %s", err)
		return false
	}

	return found
}

// remember stores value under key. Cache failures never fail a command.
func (s *session) remember(key string, value any) {
	if err := s.cacheStore().Set(key, value, cacheTTL); err != nil {
		s.errOut.Dim("Could not cache response: %s", err)
	}
}

// srvRecord looks up the bsvalias SRV record of domain and, when validate
// is set, checks it against s.srv.
func (s *session) srvRecord(ctx context.Context, domain string, validate, allowCache bool) (*net.SRV, error) {
	s.step("Retrieving SRV record for %s...", s.out.Value(domain))

	var srv *net.SRV
	if !validate && s.cached(keySRV+domain, allowCache, &srv) && srv != nil {
		s.progress("SRV target: %s (from cache)", formatSRV(srv))
		return srv, nil
	}

	srv, err := s.client.GetSRVRecord(s.srv.service, s.srv.protocol, domain)
	if err != nil {
		return nil, fmt.Errorf("lookup SRV record for %s: %w", domain, err)
	}

	if srv == nil {
		return nil, fmt.Errorf("lookup SRV record for %s: %w", domain, paymail.ErrSRVMissing)
	}

	if validate {
		if err = s.client.ValidateSRVRecord(ctx, srv, s.srv.port, s.srv.priority, s.srv.weight); err != nil {
			return srv, fmt.Errorf("SRV record for %s is invalid: %w", domain, err)
		}

		s.progress("SRV record passed all validations (target, port, priority, weight)")
	}

	s.progress("SRV target: %s", formatSRV(srv))

	// go-paymail answers a failed or empty SRV lookup with a synthetic
	// <domain>:443 record, indistinguishable from a real one that points at
	// the domain itself. Caching it would pin a transient DNS failure for an
	// hour, so only records that point elsewhere are cached.
	if srv.Target != domain {
		s.remember(keySRV+domain, srv)
	}

	return srv, nil
}

// capabilities discovers the capabilities of domain: the SRV record picks
// the host and port, falling back to the domain itself on port 443.
func (s *session) capabilities(ctx context.Context, domain string, allowCache bool) (*paymail.CapabilitiesResponse, error) {
	srv, err := s.srvRecord(ctx, domain, false, allowCache)
	if err != nil {
		s.advise("No usable SRV record (%s); using %s:%d", errors.Unwrap(err), domain, paymail.DefaultPort)
	}

	return s.capabilitiesAt(domain, srv, allowCache)
}

// capabilitiesAt fetches the capabilities of domain from the host and port
// of srv, or from the domain itself on port 443 when srv is nil.
func (s *session) capabilitiesAt(domain string, srv *net.SRV, allowCache bool) (*paymail.CapabilitiesResponse, error) {
	target, port := domain, int(paymail.DefaultPort)
	if srv != nil && srv.Target != "" {
		target, port = srv.Target, int(srv.Port)
	}

	s.step("Retrieving capabilities for %s...", s.out.Value(fmt.Sprintf("%s:%d", target, port)))

	var (
		caps *paymail.CapabilitiesResponse
		err  error
	)

	fromCache := s.cached(keyCapabilities+domain, allowCache, &caps) && caps != nil
	if !fromCache {
		if caps, err = s.client.GetCapabilities(target, port); err != nil {
			return nil, fmt.Errorf("get capabilities for %s: %w", domain, err)
		}

		s.trace(caps.Tracing, caps.StatusCode)
	}

	// Checked on cached responses too: --bsvalias may differ between runs.
	if caps.BsvAlias != s.bsvAlias {
		return nil, fmt.Errorf("%w for %s: expected %q but got %q", errVersionMismatch, domain, s.bsvAlias, caps.BsvAlias)
	}

	if fromCache {
		s.progress("Found %d capabilities (from cache)", len(caps.Capabilities))
		return caps, nil
	}

	s.progress("Found %d capabilities", len(caps.Capabilities))
	s.remember(keyCapabilities+domain, caps)

	return caps, nil
}

// pki retrieves the public key of an address.
func (s *session) pki(pkiURL string, addr address, allowCache bool) (*paymail.PKIResponse, error) {
	s.step("Retrieving public key information for %s...", s.out.Value(addr.Address))

	var pki *paymail.PKIResponse
	if s.cached(keyPKI+addr.Address, allowCache, &pki) && pki != nil {
		s.progress("Found pubkey %s (from cache)", truncate(pki.PubKey, keyPrefixLength))
		return pki, nil
	}

	pki, err := s.client.GetPKI(pkiURL, addr.Alias, addr.Domain)
	if err != nil {
		return pki, fmt.Errorf("get pki for %s: %w", addr.Address, err)
	}

	s.trace(pki.Tracing, pki.StatusCode)
	s.progress("Found pubkey %s", truncate(pki.PubKey, keyPrefixLength))
	s.remember(keyPKI+addr.Address, pki)

	return pki, nil
}

// publicProfile retrieves the public profile (name and avatar) of an address.
func (s *session) publicProfile(profileURL string, addr address, allowCache bool) (*paymail.PublicProfileResponse, error) {
	s.step("Retrieving public profile for %s...", s.out.Value(addr.Address))

	var profile *paymail.PublicProfileResponse
	if s.cached(keyPublicProfile+addr.Address, allowCache, &profile) && profile != nil {
		s.progress("Found public profile (from cache)")
		return profile, nil
	}

	profile, err := s.client.GetPublicProfile(profileURL, addr.Alias, addr.Domain)
	if err != nil {
		return nil, fmt.Errorf("get public profile for %s: %w", addr.Address, err)
	}

	s.trace(profile.Tracing, profile.StatusCode)

	if profile.Name == "" && profile.Avatar == "" {
		s.advise("Public profile is empty")
		return profile, nil
	}

	s.progress("Found public profile [name, avatar]")
	s.remember(keyPublicProfile+addr.Address, profile)

	return profile, nil
}

// optionalProfile fetches the public profile of addr when the provider
// publishes the capability and skip is false. The profile is decoration, so a
// failure is reported and otherwise ignored.
func (s *session) optionalProfile(caps *paymail.CapabilitiesResponse, addr address, skip bool) *paymail.PublicProfileResponse {
	profileURL := capString(caps, paymail.BRFCPublicProfile, "")
	if profileURL == "" || skip {
		return nil
	}

	profile, err := s.publicProfile(profileURL, addr, true)
	if err != nil {
		s.advise("Public profile unavailable: %s", err)
		return nil
	}

	return profile
}

// resolveAddress runs basic address resolution for an address.
func (s *session) resolveAddress(resolveURL string, addr address, sender *paymail.SenderRequest) (*paymail.ResolutionResponse, error) {
	s.step("Resolving address for %s...", s.out.Value(addr.Address))

	sender.Dt = s.now().UTC().Format(time.RFC3339)

	resolution, err := s.client.ResolveAddress(resolveURL, addr.Alias, addr.Domain, sender)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", addr.Address, err)
	}

	s.trace(resolution.Tracing, resolution.StatusCode)
	s.progress("Resolved output script %s", truncate(resolution.Output, keyPrefixLength))

	return resolution, nil
}

// p2pDestination requests P2P payment outputs for an address.
func (s *session) p2pDestination(destinationURL string, addr address, satoshis uint64) (*paymail.PaymentDestinationResponse, error) {
	s.step("Starting a P2P payment request for %s...", s.out.Value(addr.Address))

	destination, err := s.client.GetP2PPaymentDestination(destinationURL, addr.Alias, addr.Domain, &paymail.PaymentRequest{Satoshis: satoshis})
	if err != nil {
		return nil, fmt.Errorf("p2p payment destination for %s: %w", addr.Address, err)
	}

	s.trace(destination.Tracing, destination.StatusCode)
	s.progress("Found %d payment output(s)", len(destination.Outputs))

	return destination, nil
}

// verifyPubKey asks the provider whether pubKey belongs to an address.
func (s *session) verifyPubKey(verifyURL string, addr address, pubKey string) (*paymail.VerificationResponse, error) {
	s.step("Verifying pubkey for %s...", s.out.Value(addr.Address))

	verification, err := s.client.VerifyPubKey(verifyURL, addr.Alias, addr.Domain, pubKey)
	if err != nil {
		return nil, fmt.Errorf("verify pubkey for %s: %w", addr.Address, err)
	}

	s.trace(verification.Tracing, verification.StatusCode)

	return verification, nil
}

// isNotFound reports whether a provider answered a PKI lookup with "this
// handle is not registered": a 404, or the 400 "Paymail not found" that
// HandCash sends.
func isNotFound(pki *paymail.PKIResponse, err error) bool {
	if err == nil || pki == nil {
		return false
	}

	switch pki.StatusCode {
	case http.StatusNotFound:
		return true
	case http.StatusBadRequest:
		return strings.Contains(strings.ToLower(err.Error()), "not found")
	}

	return false
}

// capString returns the string value of a capability (or its alternate),
// or "" when it is missing or not a string. Unlike the go-paymail helpers it
// never panics on a provider that publishes an unexpected type.
func capString(caps *paymail.CapabilitiesResponse, id, alternate string) string {
	for _, key := range []string{id, alternate} {
		if key == "" {
			continue
		}

		if value, ok := caps.Capabilities[key].(string); ok {
			return value
		}
	}

	return ""
}

// capBool returns the boolean value of a capability, or false when it is
// missing or not a boolean.
func capBool(caps *paymail.CapabilitiesResponse, id string) bool {
	value, ok := caps.Capabilities[id].(bool)
	return ok && value
}

// requireCapability returns the URL of a required capability, or an error
// naming it when the provider does not publish it.
func requireCapability(caps *paymail.CapabilitiesResponse, domain, id, alternate string) (string, error) {
	if url := capString(caps, id, alternate); url != "" {
		return url, nil
	}

	return "", fmt.Errorf("%w: %s does not publish %s", errMissingCapability, domain, id)
}

// formatSRV renders an SRV record for progress lines.
func formatSRV(srv *net.SRV) string {
	return fmt.Sprintf("%s:%d --weight %d --priority %d", srv.Target, srv.Port, srv.Weight, srv.Priority)
}
