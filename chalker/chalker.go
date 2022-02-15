/*
Package chalker is a logging interface for the commands->stdout and uses the chalk package
*/
package chalker

import (
	"errors"

	"github.com/fatih/color"
)

// Logging color/style types
const (
	DEFAULT = "default"
	DIM     = "dim"
	BOLD    = "bold"
	ERROR   = "error"
	INFO    = "info"
	SUCCESS = "success"
	WARN    = "warn"
)

// Error chalks and returns an error
func Error(body string) error {
	return errors.New(color.MagentaString(body))
}

// Log writes chalks to console
func Log(level string, body string) {
	switch level {
	case INFO:
		color.Cyan(body)
		// fmt.Printf("%s%s%s%s\n", chalk.Cyan, spacer, body, chalk.Reset)
	case WARN:
		color.Yellow(body)
		// fmt.Printf("%s%s%s%s\n", chalk.Yellow, spacer, body, chalk.Reset)
	case ERROR:
		color.Magenta(body)
		// fmt.Printf("%s%s%s%s\n", chalk.Magenta, spacer, body, chalk.Reset)
	case SUCCESS:
		color.Green(body)
		// fmt.Printf("%s%s%s%s\n", chalk.Green, spacer, body, chalk.Reset)
	case DEFAULT:
		fallthrough
	default:
		color.White(body)
		// fmt.Printf("%s%s\n", spacer, body)
	}
}
