/*
Package chalker is a logging interface for the commands->stdout and uses the chalk package
*/
package chalker

import (
	"fmt"

	"github.com/ttacon/chalk"
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

var (
	bold   = chalk.White.NewStyle().WithTextStyle(chalk.Bold)
	dim    = chalk.White.NewStyle().WithTextStyle(chalk.Dim)
	spacer string
)

// Error chalks and returns an error
func Error(body string) error {
	return fmt.Errorf("%s %s%s", chalk.Magenta, body, chalk.Reset)
}

// Log writes chalks to console
func Log(level string, body string) {
	switch level {
	case INFO:
		fmt.Printf("%s%s%s%s\n", chalk.Cyan, spacer, body, chalk.Reset)
	case WARN:
		fmt.Printf("%s%s%s%s\n", chalk.Yellow, spacer, body, chalk.Reset)
	case ERROR:
		fmt.Printf("%s%s%s%s\n", chalk.Magenta, spacer, body, chalk.Reset)
	case SUCCESS:
		fmt.Printf("%s%s%s%s\n", chalk.Green, spacer, body, chalk.Reset)
	case DIM:
		fmt.Printf("%s%s\n", dim.Style(body), chalk.Reset)
	case BOLD:
		fmt.Printf("%s%s\n", bold.Style(body), chalk.Reset)
	case DEFAULT:
		fallthrough
	default:
		fmt.Printf("%s%s\n", spacer, body)
	}
}
