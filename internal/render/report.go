package render

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/engine/runner"

	"github.com/benchttp/cli/internal/render/ansi"
)

func ReportSummary(w io.Writer, rep *runner.Report) (int, error) {
	return w.Write([]byte(ReportSummaryString(rep)))
}

// String returns a default summary of the Report as a string.
func ReportSummaryString(rep *runner.Report) string {
	var b strings.Builder

	line := func(name string, value interface{}) string {
		const template = "%-18s %v\n"
		return fmt.Sprintf(template, name, value)
	}

	msString := func(d time.Duration) string {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	formatRequests := func(n, max int) string {
		maxString := strconv.Itoa(max)
		if maxString == "-1" {
			maxString = "∞"
		}
		return fmt.Sprintf("%d/%s", n, maxString)
	}

	var (
		m   = rep.Metrics
		cfg = rep.Metadata.Config
	)

	b.WriteString(ansi.Bold("→ Summary"))
	b.WriteString("\n")
	b.WriteString(line("Endpoint", cfg.Request.URL))
	b.WriteString(line("Requests", formatRequests(len(m.Records), cfg.Runner.Requests)))
	b.WriteString(line("Errors", len(m.RequestFailures)))
	b.WriteString(line("Min response time", msString(m.ResponseTimes.Min)))
	b.WriteString(line("Max response time", msString(m.ResponseTimes.Max)))
	b.WriteString(line("Mean response time", msString(m.ResponseTimes.Mean)))
	b.WriteString(line("Total duration", msString(rep.Metadata.TotalDuration)))
	b.WriteString("\n")

	return b.String()
}
