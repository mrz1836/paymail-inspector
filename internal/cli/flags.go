package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// annotationRetired marks a flag kept only for compatibility; its value is
// the reason shown when someone still passes it.
const annotationRetired = "paymail-inspector/retired"

// Reasons shown when a retired flag is used.
const (
	reasonServiceShutDown = "the service it queried has shut down; the flag is ignored"
	reasonAlwaysRequired  = "the pki lookup is required to resolve an address; the flag is ignored"
	reasonUseGenDocs      = "generate the command reference with the gen-docs command; the flag is ignored"
)

// retiredIntegrationFlags are the skip flags of third-party integrations whose
// services have shut down (Bitpic, Roundesk, Baemail, PowPing).
func retiredIntegrationFlags() []string {
	return []string{"skip-bitpic", "skip-roundesk", "skip-baemail", "skip-powping"}
}

// deprecateFlags registers flags of retired features as hidden no-ops, so
// scripts that still pass them keep working. warnRetiredFlags tells the user
// why the flag no longer does anything.
func deprecateFlags(cmd *cobra.Command, reason string, names ...string) {
	for _, name := range names {
		cmd.Flags().Bool(name, false, "")
		_ = cmd.Flags().MarkHidden(name)
		_ = cmd.Flags().SetAnnotation(name, annotationRetired, []string{reason})
	}
}

// warnRetiredFlags prints a notice on stderr, never stdout, for each retired
// flag passed to cmd, so redirected output stays clean.
func (a *app) warnRetiredFlags(cmd *cobra.Command) {
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if reason := f.Annotations[annotationRetired]; len(reason) > 0 {
			a.errOut.Warn("Flag --%s is retired: %s", f.Name, reason[0])
		}
	})
}
