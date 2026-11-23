// Package output renders the CLI's human-readable terminal output.
//
// Every line goes through a Printer, which owns the destination writer and
// the color decision. Nothing here touches the process-wide color state of
// github.com/fatih/color, so a Printer that writes to a bytes.Buffer in a test
// behaves exactly like one writing to a terminal, minus the escape codes.
package output

import (
	"fmt"
	"io"
	"sync"

	"github.com/fatih/color"
)

// Level selects the style of a whole line.
type Level int

// Line styles, from plain to alarming.
const (
	// LevelDefault is plain body text.
	LevelDefault Level = iota
	// LevelDim is low-priority detail, such as request tracing.
	LevelDim
	// LevelBold is a section title.
	LevelBold
	// LevelInfo is a neutral notice.
	LevelInfo
	// LevelSuccess reports a check or request that succeeded.
	LevelSuccess
	// LevelWarn reports something worth a look that did not fail.
	LevelWarn
	// LevelError reports a failure.
	LevelError
)

// headerPrefix opens every section header.
const headerPrefix = "\n==========| "

// Printer writes styled lines to a single writer. It is safe for concurrent
// use: each line is written under a lock, so output from parallel lookups is
// never interleaved mid-line.
type Printer struct {
	w     io.Writer
	color bool
	mu    sync.Mutex
}

// New returns a Printer writing to w. Colors are emitted only when useColor is
// true; the caller decides (terminal detection, --no-color, NO_COLOR).
func New(w io.Writer, useColor bool) *Printer {
	return &Printer{w: w, color: useColor}
}

// Writer returns the destination writer.
func (p *Printer) Writer() io.Writer {
	return p.w
}

// ColorEnabled reports whether the Printer emits ANSI colors.
func (p *Printer) ColorEnabled() bool {
	return p.color
}

// Line writes text as one line in the given style.
func (p *Printer) Line(level Level, text string) {
	c := p.style(levelAttributes(level)...)

	p.mu.Lock()
	defer p.mu.Unlock()

	_, _ = c.Fprintln(p.w, text)
}

// Linef formats according to format and writes the result as one line.
func (p *Printer) Linef(level Level, format string, args ...any) {
	p.Line(level, fmt.Sprintf(format, args...))
}

// Header writes a section header in the given style.
func (p *Printer) Header(level Level, text string) {
	p.Line(level, headerPrefix+text)
}

// Default writes a formatted line of plain body text.
func (p *Printer) Default(format string, args ...any) {
	p.Linef(LevelDefault, format, args...)
}

// Dim writes a formatted line of low-priority detail.
func (p *Printer) Dim(format string, args ...any) {
	p.Linef(LevelDim, format, args...)
}

// Info writes a formatted neutral notice.
func (p *Printer) Info(format string, args ...any) {
	p.Linef(LevelInfo, format, args...)
}

// Success writes a formatted success line.
func (p *Printer) Success(format string, args ...any) {
	p.Linef(LevelSuccess, format, args...)
}

// Warn writes a formatted warning line.
func (p *Printer) Warn(format string, args ...any) {
	p.Linef(LevelWarn, format, args...)
}

// Error writes a formatted failure line.
func (p *Printer) Error(format string, args ...any) {
	p.Linef(LevelError, format, args...)
}

// Value highlights a value inside a line (an address, URL, or key).
func (p *Printer) Value(s string) string {
	return p.style(color.FgCyan).Sprint(s)
}

// Good highlights a positive word inside a line, such as "Enabled".
func (p *Printer) Good(s string) string {
	return p.style(color.FgGreen).Sprint(s)
}

// Caution highlights a cautionary word inside a line.
func (p *Printer) Caution(s string) string {
	return p.style(color.FgYellow).Sprint(s)
}

// Alert highlights a negative word inside a line, such as "Disabled".
func (p *Printer) Alert(s string) string {
	return p.style(color.FgMagenta).Sprint(s)
}

// Label highlights a field label inside a line.
func (p *Printer) Label(s string) string {
	return p.style(color.FgWhite).Sprint(s)
}

// style returns a color with attrs, honoring the Printer's color decision
// rather than the package-level default of fatih/color.
func (p *Printer) style(attrs ...color.Attribute) *color.Color {
	c := color.New(attrs...)
	if p.color {
		c.EnableColor()
	} else {
		c.DisableColor()
	}

	return c
}

// levelAttributes maps a line style to its terminal attributes.
func levelAttributes(level Level) []color.Attribute {
	switch level {
	case LevelDim:
		return []color.Attribute{color.Faint}
	case LevelBold:
		return []color.Attribute{color.Bold}
	case LevelInfo:
		return []color.Attribute{color.FgCyan}
	case LevelSuccess:
		return []color.Attribute{color.FgGreen}
	case LevelWarn:
		return []color.Attribute{color.FgYellow}
	case LevelError:
		return []color.Attribute{color.FgMagenta}
	case LevelDefault:
		return []color.Attribute{color.Reset}
	}

	return []color.Attribute{color.Reset}
}
