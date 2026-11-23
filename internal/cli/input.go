package cli

import (
	"fmt"
	"strings"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/mrz1836/go-sanitize"
)

// maxHandleLength bounds a handle searched by whois.
const maxHandleLength = 255

// address is a validated paymail address split into its parts.
type address struct {
	Alias   string
	Domain  string
	Address string
}

// parseAddress accepts a paymail address or a provider handle shortcut
// ($handle for HandCash, 1handle for RelayX) and validates it.
func parseAddress(raw string) (address, error) {
	alias, domain, addr := paymail.SanitizePaymail(paymail.ConvertHandle(strings.TrimSpace(raw), false))
	if alias == "" || domain == "" || addr == "" {
		return address{}, fmt.Errorf("%w: %q", errInvalidPaymail, raw)
	}

	if err := paymail.ValidatePaymail(addr); err != nil {
		return address{}, fmt.Errorf("%w: %q: %w", errInvalidPaymail, raw, err)
	}

	if err := paymail.ValidateDomain(domain); err != nil {
		return address{}, fmt.Errorf("%w: %q: %w", errInvalidDomain, domain, err)
	}

	return address{Alias: alias, Domain: domain, Address: addr}, nil
}

// parseDomain accepts a domain or a paymail address and returns the
// validated domain.
func parseDomain(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "@") {
		addr, err := parseAddress(raw)
		if err != nil {
			return "", err
		}

		return addr.Domain, nil
	}

	domain, err := sanitize.Domain(raw, false, true)
	if err != nil || domain == "" {
		return "", fmt.Errorf("%w: %q", errInvalidDomain, raw)
	}

	if err = paymail.ValidateDomain(domain); err != nil {
		return "", fmt.Errorf("%w: %q: %w", errInvalidDomain, raw, err)
	}

	return domain, nil
}

// parseHandle extracts the alias to search for from a bare handle, a
// provider shortcut, or a full paymail address.
func parseHandle(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "@") {
		raw = paymail.ConvertHandle(raw, false)
	}

	handle := raw
	if strings.Contains(raw, "@") {
		handle, _, _ = paymail.SanitizePaymail(raw)
	}

	handle = sanitize.Custom(strings.ToLower(handle), `[^a-z0-9-_.+]`)
	if handle == "" || len(handle) > maxHandleLength {
		return "", fmt.Errorf("%w: %q", errInvalidHandle, raw)
	}

	return handle, nil
}

// truncate shortens s to at most n characters followed by an ellipsis.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n] + "..."
}
