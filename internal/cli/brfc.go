package cli

import (
	"fmt"
	"strings"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/mrz1836/go-sanitize"
	"github.com/spf13/cobra"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

// brfcOptions are the flags of the brfc commands.
type brfcOptions struct {
	skipValidation bool
	title          string
	author         string
	version        string
}

// newBrfcCmd returns the brfc command group: list, search and generate.
func newBrfcCmd(a *app) *cobra.Command {
	opts := &brfcOptions{}

	cmd := &cobra.Command{
		Use:     "brfc",
		Short:   "List, search, or generate BRFC specification IDs",
		Aliases: []string{"spec", "b"},
		Long: `
___.           _____
\_ |__________/ ____\____
 | __ \_  __ \   __\/ ___\
 | \_\ \  | \/|  | \  \___
 |___  /__|   |__|  \___  >
     \/                 \/

BRFC (Bitcoin SV Request-For-Comments) specifications describe functionality
across the ecosystem. The bsvalias protocols behind paymail are defined across
a series of BRFC documents, and BRFC IDs are how a provider advertises the
extensions it supports during capability discovery.

  list      show every known BRFC specification
  search    find specifications by ID, title, or author
  generate  compute the BRFC ID for a new specification

Read more at: ` + specMirror + "01-brfc-specifications.html",
		SuggestFor: []string{"specs", "specifications"},
		Example: a.examples(
			"brfc list",
			"brfc search nChain",
			`brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1`,
		),
		Args: unknownSubcommand,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.skipValidation, "skip-validation", false, "Skip validating the IDs of existing BRFC specifications")

	cmd.AddCommand(
		newBrfcListCmd(a, opts),
		newBrfcSearchCmd(a, opts),
		newBrfcGenerateCmd(a, opts),
	)

	return cmd
}

// newBrfcListCmd returns `brfc list`.
func newBrfcListCmd(a *app, opts *brfcOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all known BRFC specifications",
		Aliases: []string{"ls"},
		Example: a.examples("brfc list", "brfc list --skip-validation"),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			specs, err := a.brfcSpecs()
			if err != nil {
				return err
			}

			shown := 0

			for _, spec := range specs {
				// Skip malformed entries; the embedded list should have none.
				if spec.ID == "" || spec.Title == "" || spec.Version == "" {
					continue
				}

				a.renderBRFC(spec, opts.skipValidation)

				shown++
			}

			a.out.Success("\nTotal BRFC specifications found: %d", shown)

			return nil
		},
	}
}

// newBrfcSearchCmd returns `brfc search <term>`.
func newBrfcSearchCmd(a *app, opts *brfcOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "search <term>",
		Short:   "Search BRFC specifications by ID, title, or author",
		Aliases: []string{"find"},
		Example: a.examples("brfc search nChain", "brfc search 759684b1a19a"),
		Args:    exactArgs(1, "a search term"),
		RunE: func(_ *cobra.Command, args []string) error {
			specs, err := a.brfcSpecs()
			if err != nil {
				return err
			}

			term := strings.TrimSpace(sanitize.SingleLine(args[0]))
			found := 0

			for _, spec := range specs {
				if containsFold(spec.ID, term) || containsFold(spec.Title, term) || containsFold(spec.Author, term) {
					a.renderBRFC(spec, opts.skipValidation)

					found++
				}
			}

			if found == 0 {
				return fmt.Errorf("%w: %q", errNoMatches, term)
			}

			a.out.Success("\nTotal BRFC specifications found: %d searching: %s", found, term)

			return nil
		},
	}
}

// newBrfcGenerateCmd returns `brfc generate`.
func newBrfcGenerateCmd(a *app, opts *brfcOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate the BRFC ID for a new specification",
		Aliases: []string{"gen", "new"},
		Example: a.examples(`brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1`),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return a.runBrfcGenerate(opts)
		},
	}

	cmd.Flags().StringVar(&opts.title, "title", "", "Title of the new BRFC specification (required)")
	cmd.Flags().StringVar(&opts.author, "author", "", "Author(s) of the new BRFC specification")
	cmd.Flags().StringVar(&opts.version, "version", "", "Version of the new BRFC specification")

	return cmd
}

// runBrfcGenerate computes a new BRFC ID and refuses one that already exists.
func (a *app) runBrfcGenerate(opts *brfcOptions) error {
	spec := &paymail.BRFCSpec{
		Author:  strings.TrimSpace(opts.author),
		Title:   strings.TrimSpace(opts.title),
		Version: strings.TrimSpace(opts.version),
	}

	if spec.Title == "" {
		return newUsageError("brfc generate requires --title")
	}

	if err := spec.Generate(); err != nil {
		return fmt.Errorf("generate BRFC ID: %w", err)
	}

	specs, err := a.brfcSpecs()
	if err != nil {
		return err
	}

	for _, existing := range specs {
		if existing.ID == spec.ID {
			return fmt.Errorf("%w: %s (%s)", errBRFCExists, spec.ID, existing.Title)
		}
	}

	a.out.Header(output.LevelBold, "Generated BRFC ID")
	field(a.out, "ID", spec.ID)
	field(a.out, "Title", spec.Title)
	field(a.out, "Author", spec.Author)
	field(a.out, "Version", spec.Version)

	return nil
}

// brfcSpecs returns the BRFC specifications bundled with go-paymail.
func (a *app) brfcSpecs() ([]*paymail.BRFCSpec, error) {
	client, err := a.newClient(clientOptions{userAgent: a.userAgent()})
	if err != nil {
		return nil, err
	}

	specs := client.GetBRFCs()
	if len(specs) == 0 {
		return nil, errNoBRFCs
	}

	return specs, nil
}

// renderBRFC prints one specification, flagging an ID that does not match
// the one its title, author, and version generate.
func (a *app) renderBRFC(spec *paymail.BRFCSpec, skipValidation bool) {
	a.out.Header(output.LevelBold, spec.Title+" v"+spec.Version)

	status := a.out.Good("(valid)")

	if !skipValidation {
		valid, generated, err := spec.Validate()

		switch {
		case err != nil:
			status = a.out.Alert("(validation error: " + err.Error() + ")")
		case !valid:
			status = a.out.Caution("(invalid ID, expected " + generated + ")")
		}
	}

	a.out.Default("%-10s: %s %s", "ID", a.out.Value(spec.ID), status)
	field10(a, "Author(s)", spec.Author)
	field10(a, "Alias", spec.Alias)
	field10(a, "URL", specURL(spec.URL))
}

// field10 prints a BRFC field with the narrower label column of brfc output.
func field10(a *app, label, value string) {
	if value == "" {
		return
	}

	a.out.Default("%-10s: %s", label, a.out.Value(value))
}

// containsFold reports whether substr is within s, ignoring case.
func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
