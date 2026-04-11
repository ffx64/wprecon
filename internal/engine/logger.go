package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	verbose bool
	mu      sync.Mutex
	start   time.Time
}

func NewLogger(verbose bool) *Logger {
	return &Logger{
		verbose: verbose,
		start:   time.Now(),
	}
}

func (l *Logger) IsVerbose() bool {
	return l.verbose
}

func (l *Logger) elapsed() string {
	d := time.Since(l.start)
	return fmt.Sprintf("%6.1fs", d.Seconds())
}

func (l *Logger) log(prefix, color, msg string) {
	if !l.verbose {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	reset := "\033[0m"
	dim := "\033[2m"
	fmt.Printf("%s%s%s %s%s%s %s\n", dim, l.elapsed(), reset, color, prefix, reset, msg)
}

func (l *Logger) TemplateLoaded(id, name, severity string, tags []string) {
	l.log("[TPL]", "\033[36m", fmt.Sprintf("Loaded %-45s [%s] tags=[%s]", id, severity, strings.Join(tags, ",")))
}

func (l *Logger) TemplateFiltered(id, reason string) {
	l.log("[SKP]", "\033[33m", fmt.Sprintf("Skipped %-44s reason=%s", id, reason))
}

func (l *Logger) TemplateStart(id, target string) {
	l.log("[RUN]", "\033[35m", fmt.Sprintf("Executing %-42s → %s", id, target))
}

func (l *Logger) RequestStart(method, url string) {
	l.log("[REQ]", "\033[34m", fmt.Sprintf("%s %s", method, url))
}

func (l *Logger) RequestDone(method, url string, status int, timeMs float64, bodyLen int) {
	statusColor := "\033[32m"
	if status >= 400 {
		statusColor = "\033[31m"
	} else if status >= 300 {
		statusColor = "\033[33m"
	}
	reset := "\033[0m"
	l.log("[RES]", "\033[34m", fmt.Sprintf("%s %s → %s%d%s (%.0fms, %d bytes)", method, url, statusColor, status, reset, timeMs, bodyLen))
}

func (l *Logger) RequestError(method, url string, err error) {
	l.log("[ERR]", "\033[31m", fmt.Sprintf("%s %s → %v", method, url, err))
}

func (l *Logger) RequestRetry(method, url string, attempt int, err error) {
	l.log("[RTY]", "\033[33m", fmt.Sprintf("%s %s attempt=%d err=%v", method, url, attempt, err))
}

func (l *Logger) MatcherResult(matcherType string, matched bool, evidence map[string]string) {
	icon := "✗"
	color := "\033[31m"
	if matched {
		icon = "✓"
		color = "\033[32m"
	}

	extra := ""
	if len(evidence) > 0 {
		parts := make([]string, 0, len(evidence))
		for k, v := range evidence {
			if len(v) > 60 {
				v = v[:60] + "..."
			}
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		extra = " " + strings.Join(parts, ", ")
	}

	l.log("[MAT]", color, fmt.Sprintf("%s %-10s%s", icon, matcherType, extra))
}

func (l *Logger) ExtractorResult(name string, values []string) {
	if len(values) == 0 {
		return
	}
	display := make([]string, 0, len(values))
	for _, v := range values {
		if len(v) > 80 {
			v = v[:80] + "..."
		}
		display = append(display, v)
	}
	l.log("[EXT]", "\033[36m", fmt.Sprintf("%-15s → [%s]", name, strings.Join(display, ", ")))
}

func (l *Logger) FindingDetected(templateID, severity, url string, status int) {
	sevColor := "\033[37m"
	switch severity {
	case "critical":
		sevColor = "\033[31;1m"
	case "high":
		sevColor = "\033[31m"
	case "medium":
		sevColor = "\033[33m"
	case "low":
		sevColor = "\033[34m"
	case "info":
		sevColor = "\033[36m"
	}
	reset := "\033[0m"
	l.log("[FND]", "\033[32;1m", fmt.Sprintf("%s%s%s %s [%d] %s", sevColor, severity, reset, templateID, status, url))
}

func (l *Logger) NoMatch(templateID, url string) {
	l.log("[---]", "\033[2m", fmt.Sprintf("No match %-43s %s", templateID, url))
}

func (l *Logger) Info(msg string) {
	l.log("[INF]", "\033[37m", msg)
}

func (l *Logger) Summary(totalTemplates, totalFindings int, duration time.Duration) {
	l.log("[SUM]", "\033[37;1m", fmt.Sprintf("Done: %d templates, %d findings, %.1fs elapsed", totalTemplates, totalFindings, duration.Seconds()))
}
