package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Sentinel errors returned by commands.
var (
	errInvalidPaymail      = errors.New("invalid paymail address")
	errInvalidDomain       = errors.New("invalid domain")
	errInvalidHandle       = errors.New("invalid handle")
	errInvalidPubKey       = errors.New("invalid pubkey")
	errMissingCapability   = errors.New("missing required capability")
	errVersionMismatch     = errors.New("bsvalias version mismatch")
	errNoBRFCs             = errors.New("no BRFC specifications available")
	errNoMatches           = errors.New("no BRFC specifications matched")
	errBRFCExists          = errors.New("BRFC already exists")
	errValidationFailed    = errors.New("validation failed")
	errNoProviderResponded = errors.New("no paymail provider responded")
	errPubKeyMismatch      = errors.New("pubkey does not match the paymail address")
)

// usageError marks a command-line mistake (bad arguments, unknown flag or
// command), which exits with ExitUsage instead of ExitError. It carries
// either its own message or a wrapped error from cobra.
type usageError struct {
	msg string
	err error
}

// Error implements error.
func (e usageError) Error() string {
	if e.msg != "" {
		return e.msg
	}

	return e.err.Error()
}

// Unwrap exposes the wrapped error, if any.
func (e usageError) Unwrap() error { return e.err }

// newUsageError formats a usage error.
func newUsageError(format string, args ...any) error {
	return usageError{msg: fmt.Sprintf(format, args...)}
}

// isUsageError reports whether err is a command-line mistake. cobra raises
// "unknown command" while resolving the command, before any hook of ours
// runs, so it is recognized by its stable prefix.
func isUsageError(err error) bool {
	if err == nil {
		return false
	}

	var usage usageError
	if errors.As(err, &usage) {
		return true
	}

	return strings.HasPrefix(err.Error(), "unknown command ")
}

// wireUsageErrors wraps every command's positional-argument validator so
// its failures surface as usage errors. It must run after the whole tree is
// built.
func wireUsageErrors(cmd *cobra.Command) {
	if validate := cmd.Args; validate != nil {
		cmd.Args = func(c *cobra.Command, args []string) error {
			if err := validate(c, args); err != nil {
				if isUsageError(err) {
					return err
				}

				return usageError{err: err}
			}

			return nil
		}
	}

	for _, sub := range cmd.Commands() {
		wireUsageErrors(sub)
	}
}

// unknownSubcommand rejects any argument to a command group as an unknown
// subcommand, with cobra's "Did you mean" suggestions.
func unknownSubcommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	// cobra only defaults the suggestion distance on the root command.
	if cmd.SuggestionsMinimumDistance <= 0 {
		cmd.SuggestionsMinimumDistance = 2
	}

	msg := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())
	if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		msg += "\n\nDid you mean this?\n\t" + strings.Join(suggestions, "\n\t")
	}

	return newUsageError("%s", msg)
}

// exactArgs returns a validator that requires exactly n arguments and
// explains what they are when the count is wrong.
func exactArgs(n int, what string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return newUsageError("%s requires %s", cmd.Name(), what)
		}

		return nil
	}
}
