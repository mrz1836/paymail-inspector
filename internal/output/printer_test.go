package output

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ansiEscape starts every terminal color sequence.
const ansiEscape = "\x1b["

func TestPrinter_PlainLevels(t *testing.T) {
	tests := []struct {
		name  string
		write func(p *Printer)
		want  string
	}{
		{"default", func(p *Printer) { p.Default("hello %s", "world") }, "hello world\n"},
		{"dim", func(p *Printer) { p.Dim("dim %d", 1) }, "dim 1\n"},
		{"info", func(p *Printer) { p.Info("info") }, "info\n"},
		{"success", func(p *Printer) { p.Success("ok") }, "ok\n"},
		{"warn", func(p *Printer) { p.Warn("careful") }, "careful\n"},
		{"error", func(p *Printer) { p.Error("broken") }, "broken\n"},
		{"line", func(p *Printer) { p.Line(LevelBold, "title") }, "title\n"},
		{"linef", func(p *Printer) { p.Linef(LevelInfo, "%s-%s", "a", "b") }, "a-b\n"},
		{"header", func(p *Printer) { p.Header(LevelBold, "Section") }, "\n==========| Section\n"},
		{"unknown level", func(p *Printer) { p.Line(Level(99), "fallback") }, "fallback\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			tt.write(New(&buf, false))
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestPrinter_ColorLevels(t *testing.T) {
	levels := []Level{LevelDefault, LevelDim, LevelBold, LevelInfo, LevelSuccess, LevelWarn, LevelError}

	for _, level := range levels {
		var buf bytes.Buffer

		New(&buf, true).Line(level, "text")
		assert.Contains(t, buf.String(), ansiEscape, "level %d should be colored", level)
		assert.Contains(t, buf.String(), "text")
	}
}

func TestPrinter_Highlights(t *testing.T) {
	highlights := map[string]func(p *Printer, s string) string{
		"value":   (*Printer).Value,
		"good":    (*Printer).Good,
		"caution": (*Printer).Caution,
		"alert":   (*Printer).Alert,
		"label":   (*Printer).Label,
	}

	for name, highlight := range highlights {
		t.Run(name, func(t *testing.T) {
			plain := New(&bytes.Buffer{}, false)
			assert.Equal(t, "x", highlight(plain, "x"))

			colored := New(&bytes.Buffer{}, true)
			got := highlight(colored, "x")
			assert.True(t, strings.HasPrefix(got, ansiEscape), "got %q", got)
			assert.Contains(t, got, "x")
		})
	}
}

func TestPrinter_Accessors(t *testing.T) {
	var buf bytes.Buffer

	p := New(&buf, true)
	assert.Same(t, &buf, p.Writer())
	assert.True(t, p.ColorEnabled())
	assert.False(t, New(&buf, false).ColorEnabled())
}

func TestPrinter_ConcurrentLinesNeverInterleave(t *testing.T) {
	var buf bytes.Buffer

	p := New(&buf, false)
	line := strings.Repeat("x", 200)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			p.Default("%s", line)
		}()
	}

	wg.Wait()

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	require.Len(t, lines, 50)

	for _, got := range lines {
		assert.Equal(t, line, got)
	}
}
