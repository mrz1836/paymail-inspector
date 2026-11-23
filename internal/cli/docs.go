package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// Defaults of the hidden docs generator.
const (
	defaultDocsDir  = "docs/commands"
	docsDirPerm     = 0o750
	defaultDocsName = aliasName // the docs show the short name every install provides

	// docsConfigHint replaces the machine-specific config path in the docs.
	docsConfigHint = "<user config dir>/" + appName + "/" + configName
)

// newDocsCmd returns the hidden `gen-docs` command, which writes the
// Markdown command reference in docs/commands.
func newDocsCmd(a *app) *cobra.Command {
	var dir, name string

	cmd := &cobra.Command{
		Use:    "gen-docs",
		Short:  "Generate the Markdown command reference",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := os.MkdirAll(dir, docsDirPerm); err != nil {
				return fmt.Errorf("create docs directory: %w", err)
			}

			// Build a fresh tree under the documented name, so examples and
			// command paths read "paymail ..." whatever this binary is called,
			// and with a config hint that is not this machine's path.
			docsApp := a.renamed(name)
			docsApp.configHint = docsConfigHint

			root := newRootCmd(docsApp)
			root.InitDefaultCompletionCmd()
			fenceArt(root)

			if err := doc.GenMarkdownTree(root, dir); err != nil {
				return fmt.Errorf("generate docs: %w", err)
			}

			a.out.Success("Generated the command reference in %s", dir)

			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", defaultDocsDir, "Output directory")
	cmd.Flags().StringVar(&name, "name", defaultDocsName, "Command name used in the generated docs")

	return cmd
}

// fenceArt wraps each command's ASCII-art banner in a code block so it
// survives Markdown rendering. By convention a Long help that opens with a
// newline starts with a banner, which runs to the first blank line.
func fenceArt(cmd *cobra.Command) {
	if strings.HasPrefix(cmd.Long, "\n") {
		if art, text, found := strings.Cut(cmd.Long[1:], "\n\n"); found {
			cmd.Long = "```text\n" + art + "\n```\n\n" + text
		}
	}

	for _, sub := range cmd.Commands() {
		fenceArt(sub)
	}
}
