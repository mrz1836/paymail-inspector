package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/spf13/cobra"
)

// Exit codes returned by Execute.
const (
	ExitOK    = 0 // success
	ExitError = 1 // the command ran and failed
	ExitUsage = 2 // the command line itself was invalid
)

// Execute runs the CLI with os.Args and returns the process exit code. It
// never calls os.Exit itself, so main stays a one-line shim.
func Execute(ctx context.Context, info BuildInfo) int {
	a := newApp(invokedName(os.Args[0]), info, os.Stdout, os.Stderr)

	return a.run(ctx, os.Args[1:])
}

// run executes one command line against a and maps the outcome to an exit
// code. Errors are reported here, exactly once.
func (a *app) run(ctx context.Context, args []string) int {
	root := newRootCmd(a) //nolint:contextcheck // ctx reaches every command through ExecuteContextC below
	root.SetArgs(args)

	cmd, err := root.ExecuteContextC(ctx)

	if closeErr := a.closeCache(); closeErr != nil {
		a.errOut.Warn("%s", closeErr)
	}

	if err == nil {
		return ExitOK
	}

	if cmd == nil {
		cmd = root
	}

	a.reportError(cmd, err)

	if isUsageError(err) {
		return ExitUsage
	}

	return ExitError
}

// newRootCmd builds the complete command tree for a.
func newRootCmd(a *app) *cobra.Command {
	root := &cobra.Command{
		Use:   a.name,
		Short: "Inspect, validate and resolve paymail addresses and domains",
		Long: banner + `
Inspect, validate and resolve paymail addresses, discover a provider's
capabilities, and explore the BRFC specifications behind them.

Docs: https://github.com/mrz1836/paymail-inspector`,
		Example: a.examples(
			"resolve "+exampleAddress,
			"capabilities "+exampleDomain,
			"validate "+exampleDomain,
			"whois mrz",
		),
		Version:           a.build.String(),
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
		// No Args validator: cobra's default for a root with subcommands
		// rejects stray arguments as "unknown command" *with* "Did you mean"
		// suggestions, which a custom validator would lose.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return a.prepare(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.flags.flushCache {
				return nil // flushed in prepare
			}

			return cmd.Help()
		},
	}

	root.SetOut(a.stdout)
	root.SetErr(a.stderr)
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return usageError{err: err}
	})

	flags := root.PersistentFlags()
	flags.StringVar(&a.flags.configFile, "config", "", "Config file (default: "+a.defaultConfigHint()+")")
	flags.BoolVar(&a.flags.noCache, "no-cache", false, "Skip cached responses for this command (fresh results are still cached)")
	flags.BoolVar(&a.flags.flushCache, "flush-cache", false, "Empty the local response cache before running")
	flags.BoolVarP(&a.flags.skipTracing, "skip-tracing", "t", false, "Hide per-request network timing")
	flags.BoolVar(&a.flags.noColor, "no-color", false, "Disable colored output (NO_COLOR is honored too)")
	flags.String(keyBsvAlias, paymail.DefaultBsvAliasVersion, "The bsvalias version to expect from providers")
	_ = a.cfg.BindPFlag(keyBsvAlias, flags.Lookup(keyBsvAlias))

	deprecateFlags(root, reasonUseGenDocs, "docs")

	root.AddCommand(
		newBrfcCmd(a),
		newCapabilitiesCmd(a),
		newP2PCmd(a),
		newResolveCmd(a),
		newValidateCmd(a),
		newVerifyCmd(a),
		newWhoisCmd(a),
		newVersionCmd(a),
		newDocsCmd(a),
	)

	// Registered after the persistent hooks above: the update notice chains
	// onto them, so they must exist first.
	attachUpdateCommand(root, a)
	skipHooksForCompletion(root)
	wireUsageErrors(root)

	return root
}

// skipHooksForCompletion keeps shell tab-completion fast and side-effect
// free: the hidden __complete commands run without loading config, touching
// the cache, or starting an update check.
func skipHooksForCompletion(root *cobra.Command) {
	pre, post := root.PersistentPreRunE, root.PersistentPostRunE

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if isCompletionRequest(cmd) || pre == nil {
			return nil
		}

		return pre(cmd, args)
	}

	root.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if isCompletionRequest(cmd) || post == nil {
			return nil
		}

		return post(cmd, args)
	}
}

// isCompletionRequest reports whether cmd is one of cobra's hidden
// tab-completion commands.
func isCompletionRequest(cmd *cobra.Command) bool {
	return cmd.Name() == cobra.ShellCompRequestCmd || cmd.Name() == cobra.ShellCompNoDescRequestCmd
}

// prepare runs before every command: it applies --no-color, warns about
// retired flags, loads the config file, and handles --flush-cache.
func (a *app) prepare(cmd *cobra.Command) error {
	if a.flags.noColor {
		a.setColor(false)
	}

	a.warnRetiredFlags(cmd)

	if err := a.loadConfig(); err != nil {
		return err
	}

	if a.flags.flushCache {
		if err := a.flushCache(); err != nil {
			return err
		}

		a.out.Success("Flushed the local response cache")
	}

	return nil
}

// loadConfig reads the config file. A file named by --config must load; one
// found on the default search path is optional, so a broken one is reported
// and skipped rather than taking every command (and tab completion) down.
func (a *app) loadConfig() error {
	if a.flags.configFile != "" {
		return a.readConfig(a.flags.configFile)
	}

	path, err := a.findConfig()
	if err == nil && path != "" {
		err = a.readConfig(path)
	}

	if err != nil {
		a.errOut.Warn("Ignoring config: %s", err)
	}

	return nil
}

// readConfig loads one config file into the registry.
func (a *app) readConfig(path string) error {
	a.cfg.SetConfigFile(path)

	if err := a.cfg.ReadInConfig(); err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}

	return nil
}

// flushCache empties the response cache. When the cache could not even be
// opened (an unreadable file, say), the file is removed directly, so
// --flush-cache still repairs it.
func (a *app) flushCache() error {
	store := a.cacheStore()

	if a.storeErr == nil {
		if err := store.Flush(); err != nil {
			return fmt.Errorf("flush response cache: %w", err)
		}

		return nil
	}

	if err := os.Remove(a.cachePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("flush response cache: %w", err)
	}

	return nil
}

// findConfig returns the first config file on the search path, or "".
func (a *app) findConfig() (string, error) {
	for _, dir := range a.configDirs {
		path := filepath.Join(dir, configName)

		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("inspect config %s: %w", path, err)
		}
	}

	return "", nil
}

// defaultConfigHint describes where the config file is looked for.
func (a *app) defaultConfigHint() string {
	if a.configHint != "" {
		return a.configHint
	}

	if len(a.configDirs) == 0 {
		return configName
	}

	return filepath.Join(a.configDirs[0], configName)
}

// reportError prints a failed command's error to stderr, with a pointer to
// that command's --help for usage mistakes.
func (a *app) reportError(cmd *cobra.Command, err error) {
	a.errLine("Error: %s", err)

	if isUsageError(err) {
		a.errLine("Run '%s --help' for usage.", cmd.CommandPath())
	}
}
