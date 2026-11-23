package cli

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/go-resty/resty/v2"

	"github.com/mrz1836/paymail-inspector/internal/output"
)

// banner is the root command's help header.
const banner = `
__________                             .__.__    .___                                     __
\______   \_____  ___.__. _____ _____  |__|  |   |   | ____   ____________   ____   _____/  |_  ___________
 |     ___/\__  \<   |  |/     \\__  \ |  |  |   |   |/    \ /  ___/\____ \_/ __ \_/ ___\   __\/  _ \_  __ \
 |    |     / __ \\___  |  Y Y  \/ __ \|  |  |__ |   |   |  \\___ \ |  |_> >  ___/\  \___|  | (  <_> )  | \/
 |____|    (____  / ____|__|_|  (____  /__|____/ |___|___|  /____  >|   __/ \___  >\___  >__|  \____/|__|
                \/\/          \/     \/                   \/     \/ |__|        \/     \/
`

// examples renders example invocations, one per line, using the name the
// binary was invoked as.
func (a *app) examples(lines ...string) string {
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, "  "+a.name+" "+line)
	}

	return strings.Join(rendered, "\n")
}

// renderTracing prints the network timing of one request as aligned columns.
func renderTracing(out *output.Printer, info resty.TraceInfo, statusCode int) {
	rows := [][2]string{
		{"DNSLookup", info.DNSLookup.String()},
		{"TTFB", info.ServerTime.String()},
		{"TLSHandshake", info.TLSHandshake.String()},
		{"ConnTime", info.ConnTime.String()},
		{"FB to Close", info.ResponseTime.String()},
		{"TotalTime", fmt.Sprintf("%s [%d]", info.TotalTime, statusCode)},
	}

	if info.IsConnWasIdle {
		rows = append(rows, [2]string{"IsConnWasIdle", "true"}, [2]string{"ConnIdleTime", info.ConnIdleTime.String()})
	}

	if info.IsConnReused {
		rows = append(rows, [2]string{"IsConnReused", "true"}, [2]string{"", ""})
	}

	var buf bytes.Buffer

	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	for i := 0; i+1 < len(rows); i += 2 {
		_, _ = fmt.Fprintf(tw, "%s\t| %s\t| %s\t| %s\n", rows[i][0], rows[i][1], rows[i+1][0], rows[i+1][1])
	}

	_ = tw.Flush()

	for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		out.Line(output.LevelDim, strings.TrimRight(line, " |"))
	}
}

// paymailDetails is everything learned about one paymail address.
type paymailDetails struct {
	Address       address
	Provider      provider
	PKI           *paymail.PKIResponse
	PublicProfile *paymail.PublicProfileResponse
	Resolution    *paymail.ResolutionResponse
}

// render prints the details as a labeled block under a header.
func (d *paymailDetails) render(out *output.Printer) {
	out.Header(output.LevelBold, "Results for "+out.Value(d.Address.Address))

	if d.PublicProfile != nil {
		field(out, "Name", d.PublicProfile.Name)
		field(out, "Avatar", d.PublicProfile.Avatar)
	}

	if d.PKI != nil {
		field(out, "PubKey", d.PKI.PubKey)
	}

	if d.Resolution != nil {
		field(out, "Output Script", d.Resolution.Output)
		field(out, "Address", d.Resolution.Address)
		field(out, "Signature", d.Resolution.Signature)
	}
}

// fieldWidth aligns labels in result blocks.
const fieldWidth = 13

// field prints "Label        : value", skipping empty values.
func field(out *output.Printer, label, value string) {
	if value == "" {
		return
	}

	out.Default("%-*s: %s", fieldWidth, label, out.Value(value))
}
