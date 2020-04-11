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
	logPrefix string
	spacer    string
)

// SetPrefix for the logs in the CLI application output
func SetPrefix(prefix string) {
	logPrefix = prefix
	if len(logPrefix) > 0 {
		spacer = " "
	}
}

// Error chalks and returns an error
func Error(body string) error {
	return fmt.Errorf("%s%s %s%s", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), body, chalk.Reset)
}

// Log chalks stuff to console, returns nothing
func Log(level string, body string) {
	switch level {
	case INFO:
		fmt.Printf("%s%s%s%s%s\n", chalk.Cyan, chalk.Dim.TextStyle(logPrefix), spacer, body, chalk.Reset)
	case WARN:
		fmt.Printf("%s%s%s%s%s\n", chalk.Yellow, chalk.Dim.TextStyle(logPrefix), spacer, body, chalk.Reset)
	case ERROR:
		fmt.Printf("%s%s%s%s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), spacer, body, chalk.Reset)
	case SUCCESS:
		fmt.Printf("%s%s%s%s%s\n", chalk.Green, chalk.Dim.TextStyle(logPrefix), spacer, body, chalk.Reset)
	case DEFAULT:
		fallthrough
	default:
		fmt.Printf("%s%s%s\n", chalk.Dim.TextStyle(logPrefix), spacer, body)
	}
}
