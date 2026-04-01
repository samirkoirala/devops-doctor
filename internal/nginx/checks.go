package nginx

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/pkg/utils"
)

const cat = "nginx"

// Run checks nginx binary, config syntax, running processes, and recent error log lines.
func Run(ctx context.Context) []output.Result {
	_, _, err := utils.Run(ctx, "nginx", "-v")
	if err != nil {
		return []output.Result{output.Warn(cat, "install", "nginx CLI not found (not installed or not on PATH)",
			"Install nginx (e.g. `brew install nginx` on macOS, `apt install nginx` on Debian/Ubuntu) or ignore if you do not use it.")}
	}

	var out []output.Result
	verStderr, _, _ := utils.Run(ctx, "nginx", "-v")
	ver := strings.TrimSpace(verStderr)
	if ver == "" {
		ver = "nginx present"
	}
	r := output.OkDetail(cat, "version", "nginx is installed", ver)
	r.SortOrder = 10
	out = append(out, r)

	for _, x := range checkConfig(ctx) {
		x.SortOrder = 20
		out = append(out, x)
	}
	for _, x := range checkProcess(ctx) {
		x.SortOrder = 30
		out = append(out, x)
	}
	for _, x := range scanErrorLogs(ctx) {
		x.SortOrder = 40
		out = append(out, x)
	}
	return out
}

func checkConfig(ctx context.Context) []output.Result {
	_, stderr, err := utils.Run(ctx, "nginx", "-t")
	combined := strings.TrimSpace(stderr)
	if err != nil {
		if combined == "" {
			combined = "nginx -t failed (no stderr)"
		}
		return []output.Result{output.ErrDetail(cat, "config", "nginx configuration test failed (`nginx -t`)",
			"Fix syntax errors in nginx.conf and included files; run `sudo nginx -t` if permission denied without sudo.",
			combined)}
	}
	msg := "nginx configuration syntax is OK"
	if combined != "" {
		msg = strings.Split(combined, "\n")[0]
	}
	return []output.Result{output.OkDetail(cat, "config", msg, combined)}
}

func checkProcess(ctx context.Context) []output.Result {
	_, _, err := utils.Run(ctx, "pgrep", "-x", "nginx")
	if err != nil {
		return []output.Result{output.Warn(cat, "process", "No nginx master/worker processes found (pgrep)",
			"Start nginx: `brew services start nginx`, `sudo systemctl start nginx`, or `sudo nginx`; check for port conflicts.")}
	}
	pids, _, _ := utils.Run(ctx, "pgrep", "-x", "nginx")
	return []output.Result{output.OkDetail(cat, "process", "nginx process(es) running", pids)}
}

var logProblemPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\[emerg\]`),
	regexp.MustCompile(`(?i)\[alert\]`),
	regexp.MustCompile(`(?i)\[crit\]`),
	regexp.MustCompile(`(?i)\[error\]`),
}

func scanErrorLogs(ctx context.Context) []output.Result {
	paths := errorLogCandidates()
	for _, p := range paths {
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			continue
		}
		out, stderr, err := utils.Run(ctx, "tail", "-n", "60", p)
		if err != nil {
			continue
		}
		text := out + "\n" + stderr
		var hits []string
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			matched := false
			for _, re := range logProblemPatterns {
				if re.MatchString(line) {
					matched = true
					break
				}
			}
			if !matched || isBenignNginxLogLine(line) {
				continue
			}
			if len(line) > 220 {
				line = line[:220] + "…"
			}
			hits = append(hits, line)
		}
		if len(hits) > 12 {
			hits = hits[:12]
			hits = append(hits, "... (truncated)")
		}
		if len(hits) > 0 {
			return []output.Result{output.WarnDetail(cat, "error-log",
				"Recent error-level lines in "+filepath.Base(p),
				"Inspect the full log at "+p+"; fix upstream SSL, permissions, or listen addresses.",
				strings.Join(hits, "\n"))}
		}
		return []output.Result{output.OkDetail(cat, "error-log", "No serious errors in recent log tail: "+p, "")}
	}
	return []output.Result{output.Ok(cat, "error-log", "No nginx error.log found at common paths (skipped tail scan)")}
}

func isBenignNginxLogLine(line string) bool {
	lower := strings.ToLower(line)
	if strings.Contains(lower, "no such file or directory") {
		if strings.Contains(lower, "favicon.ico") ||
			strings.Contains(lower, "robots.txt") ||
			strings.Contains(lower, "apple-touch-icon") {
			return true
		}
	}
	return false
}

func errorLogCandidates() []string {
	home, _ := os.UserHomeDir()
	return []string{
		"/var/log/nginx/error.log",
		"/usr/local/var/log/nginx/error.log",
		"/opt/homebrew/var/log/nginx/error.log",
		filepath.Join(home, "var/log/nginx/error.log"),
	}
}
