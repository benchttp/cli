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

func Report(w io.Writer, rep *runner.Report) (int, error) {
	return w.Write([]byte(ReportString(rep)))
}

// String returns a default summary of the Report as a string.
func ReportString(rep *runner.Report) string {
	var b strings.Builder
	writeDefaultSummary(&b, rep)
	writeTestsResult(&b, rep)
	return b.String()
}

func writeDefaultSummary(w io.StringWriter, rep *runner.Report) {
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

	w.WriteString(ansi.Bold("→ Summary"))
	w.WriteString("\n")
	w.WriteString(line("Endpoint", cfg.Request.URL))
	w.WriteString(line("Requests", formatRequests(len(m.Records), cfg.Runner.Requests)))
	w.WriteString(line("Errors", len(m.RequestFailures)))
	w.WriteString(line("Min response time", msString(m.ResponseTimes.Min)))
	w.WriteString(line("Max response time", msString(m.ResponseTimes.Max)))
	w.WriteString(line("Mean response time", msString(m.ResponseTimes.Mean)))
	w.WriteString(line("Total duration", msString(rep.Metadata.TotalDuration)))
}

func writeTestsResult(w io.StringWriter, rep *runner.Report) {
	sr := rep.Tests
	if len(sr.Results) == 0 {
		return
	}

	w.WriteString("\n")
	w.WriteString(ansi.Bold("→ Test suite"))
	w.WriteString("\n")

	writeResultString(w, sr.Pass)
	w.WriteString("\n")

	for _, tr := range sr.Results {
		writeIndent(w, 1)
		writeResultString(w, tr.Pass)
		w.WriteString(" ")
		w.WriteString(tr.Input.Name)

		if !tr.Pass {
			w.WriteString("\n ")
			writeIndent(w, 3)
			w.WriteString(ansi.Bold("→ "))
			w.WriteString(tr.Summary)
		}

		w.WriteString("\n")
	}
}

func writeResultString(w io.StringWriter, pass bool) {
	if pass {
		w.WriteString(ansi.Green("PASS"))
	} else {
		w.WriteString(ansi.Red("FAIL"))
	}
}

func writeIndent(w io.StringWriter, count int) {
	if count <= 0 {
		return
	}
	const baseIndent = "  "
	w.WriteString(strings.Repeat(baseIndent, count))
}
