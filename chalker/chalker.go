/*
Package chalker is a logging interface for the commands->stdout and uses the chalk package
*/
package chalker

import (
	"fmt"

	"github.com/ttacon/chalk"
)

// Logging types
const (
	DEFAULT = "default"
	ERROR   = "error"
	INFO    = "info"
	SUCCESS = "success"
	WARN    = "warn"
)

var (
	spacer string
)

// Error chalks and returns an error
func Error(body string) error {
	return fmt.Errorf("%s %s%s", chalk.Magenta, body, chalk.Reset)
}

// Log chalks stuff to console, returns nothing
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
	case DEFAULT:
		fallthrough
	default:
		fmt.Printf("%s%s\n", spacer, body)
	}
}
