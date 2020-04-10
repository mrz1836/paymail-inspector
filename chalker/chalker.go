package chalker

import (
	"fmt"

	"github.com/ttacon/chalk"
)

const (
	DEFAULT = "default"
	INFO    = "info"
	WARN    = "warn"
	SUCCESS = "success"
	ERROR   = "error"
)

var logPrefix = "paymail-inspector:" // Prefix for the logs in the CLI application output

func setPrefix(prefix string) {
	logPrefix = prefix
}

// Error chalks and returns an error
func Error(body string) error {
	return fmt.Errorf("%s%s %s%s", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), body, chalk.Reset)
}

// Log chalks stuff to console, returns nothing
func Log(level string, body string) {
	switch level {
	case INFO:
		fmt.Printf("%s%s %s%s\n", chalk.Cyan, chalk.Dim.TextStyle(logPrefix), body, chalk.Reset)
	case WARN:
		fmt.Printf("%s%s %s%s\n", chalk.Yellow, chalk.Dim.TextStyle(logPrefix), body, chalk.Reset)
	case ERROR:
		fmt.Printf("%s%s %s%s\n", chalk.Magenta, chalk.Dim.TextStyle(logPrefix), body, chalk.Reset)
	case SUCCESS:
		fmt.Printf("%s%s %s%s\n", chalk.Green, chalk.Dim.TextStyle(logPrefix), body, chalk.Reset)
	case DEFAULT:
		fallthrough
	default:
		fmt.Printf("%s %s\n", chalk.Dim.TextStyle(logPrefix), body)
	}
}
