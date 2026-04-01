package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

const (
	symOK      = "✔"
	symErr     = "✖"
	symWarn    = "⚠"
	symSuggest = "💡"
)

// Formatter prints Results for humans or as JSON.
type Formatter struct {
	JSON    bool
	Verbose bool
	Out     io.Writer
}

// NewFormatter writes to stdout by default.
func NewFormatter(json, verbose bool) *Formatter {
	return &Formatter{JSON: json, Verbose: verbose, Out: os.Stdout}
}

// PrintResults renders all results; returns non-zero exit hint (count of errors).
func (f *Formatter) PrintResults(results []Result) int {
	if f.JSON {
		return f.printJSON(results)
	}
	return f.printHuman(results)
}

func (f *Formatter) printJSON(results []Result) int {
	enc := json.NewEncoder(f.Out)
	enc.SetIndent("", "  ")
	type payload struct {
		Results []Result `json:"results"`
		Summary struct {
			Success int `json:"success"`
			Warning int `json:"warning"`
			Error   int `json:"error"`
		} `json:"summary"`
	}
	var p payload
	p.Results = results
	for _, r := range results {
		switch r.Status {
		case Success:
			p.Summary.Success++
		case Warning:
			p.Summary.Warning++
		case Error:
			p.Summary.Error++
		}
	}
	_ = enc.Encode(p)
	return p.Summary.Error
}

func (f *Formatter) printHuman(results []Result) int {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.FgHiBlack).SprintFunc()

	errCount := 0
	var lastCat string
	for _, r := range results {
		if r.Category != lastCat {
			if lastCat != "" {
				_, _ = fmt.Fprintln(f.Out)
			}
			_, _ = fmt.Fprintf(f.Out, "%s\n", cyan(strings.ToUpper(r.Category)))
			lastCat = r.Category
		}

		var sym string
		switch r.Status {
		case Success:
			sym = green(symOK)
		case Warning:
			sym = yellow(symWarn)
		case Error:
			sym = red(symErr)
			errCount++
		}

		_, _ = fmt.Fprintf(f.Out, "  %s %s\n", sym, r.Message)
		if r.Suggestion != "" {
			_, _ = fmt.Fprintf(f.Out, "    %s Suggestion: %s\n", symSuggest, r.Suggestion)
		}
		showDetail := r.Detail != "" && (f.Verbose || r.Status != Success)
		if showDetail {
			for _, line := range strings.Split(strings.TrimSpace(r.Detail), "\n") {
				_, _ = fmt.Fprintf(f.Out, "    %s\n", dim(line))
			}
		}
	}
	_, _ = fmt.Fprintln(f.Out)
	return errCount
}

// Result helpers for check packages.

func Ok(category, check, msg string) Result {
	return Result{Category: category, Check: check, Status: Success, Message: msg}
}

func OkDetail(category, check, msg, detail string) Result {
	r := Ok(category, check, msg)
	r.Detail = detail
	return r
}

func Warn(category, check, msg, suggestion string) Result {
	return Result{Category: category, Check: check, Status: Warning, Message: msg, Suggestion: suggestion}
}

func WarnDetail(category, check, msg, suggestion, detail string) Result {
	r := Warn(category, check, msg, suggestion)
	r.Detail = detail
	return r
}

func Err(category, check, msg, suggestion string) Result {
	return Result{Category: category, Check: check, Status: Error, Message: msg, Suggestion: suggestion}
}

func ErrDetail(category, check, msg, suggestion, detail string) Result {
	r := Err(category, check, msg, suggestion)
	r.Detail = detail
	return r
}
