package cli

import (
	"runtime"

	"github.com/spf13/cobra"
)

// newVersionCmd returns `version`, a detailed form of --version.
func newVersionCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Show version, commit, and build information",
		Example: a.examples("version", "--version"),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			field(a.out, "Version", a.build.Version)
			field(a.out, "Commit", a.build.Commit)
			field(a.out, "Built", a.build.Date)
			field(a.out, "Go", runtime.Version())
			field(a.out, "Platform", runtime.GOOS+"/"+runtime.GOARCH)

			return nil
		},
	}
}
