package compose

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/pkg/utils"
)

const cat = "compose"

var composeFileNames = []string{
	"docker-compose.yml",
	"docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
}

// FindComposeFile walks upward from startDir looking for a compose file.
func FindComposeFile(startDir string) (dir, file string) {
	dir = startDir
	for {
		for _, name := range composeFileNames {
			p := filepath.Join(dir, name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return dir, name
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", ""
}

// Run executes Docker Compose checks in the directory that contains the compose file.
func Run(ctx context.Context, projectDir, composeFileName string) []output.Result {
	if projectDir == "" || composeFileName == "" {
		return []output.Result{output.Warn(cat, "file", "No docker-compose.yml / compose.yaml found in current tree",
			"Add a compose file or run checks from the project root; compose-specific checks were skipped.")}
	}

	fullPath := filepath.Join(projectDir, composeFileName)
	var results []output.Result
	results = append(results, output.Ok(cat, "file", "Compose file found: "+fullPath))

	results = append(results, checkComposePs(ctx, projectDir)...)
	results = append(results, checkHealth(ctx, projectDir)...)
	results = append(results, checkExposedPorts(ctx, projectDir)...)
	results = append(results, scanLogs(ctx, projectDir)...)
	return results
}

func checkComposePs(ctx context.Context, projectDir string) []output.Result {
	out, stderr, err := utils.RunInDir(ctx, projectDir, "docker", "compose", "ps", "-a")
	if err != nil {
		if strings.Contains(strings.ToLower(stderr), "no configuration file") {
			return []output.Result{output.Warn(cat, "ps", "docker compose ps failed: no compose file in directory",
				"Ensure the detected directory contains a valid compose file.")}
		}
		return []output.Result{output.ErrDetail(cat, "ps", "docker compose ps failed",
			"Ensure Docker is running and you use Compose V2 (`docker compose`). Install the plugin if missing.",
			stderr)}
	}
	return []output.Result{output.OkDetail(cat, "ps", "docker compose ps -a", out)}
}

func checkHealth(ctx context.Context, projectDir string) []output.Result {
	out, stderr, err := utils.RunInDir(ctx, projectDir, "docker", "compose", "ps", "-a", "--format", "{{.Name}}\t{{.Status}}")
	if err != nil {
		return []output.Result{output.Warn(cat, "health", "Could not inspect compose health/status", stderr)}
	}
	var unhealthy []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "unhealthy") || strings.Contains(lower, "restarting") || strings.Contains(lower, "exited") {
			unhealthy = append(unhealthy, line)
		}
	}
	if len(unhealthy) > 0 {
		return []output.Result{output.WarnDetail(cat, "health", "Some containers look unhealthy or stopped",
			"Inspect with `docker compose logs <service>` and fix failing healthchecks or dependencies.",
			strings.Join(unhealthy, "\n"))}
	}
	return []output.Result{output.Ok(cat, "health", "No obvious unhealthy/restarting/exited states in compose ps")}
}

func checkExposedPorts(ctx context.Context, projectDir string) []output.Result {
	out, _, err := utils.RunInDir(ctx, projectDir, "docker", "compose", "ps", "--format", "{{.Name}}\t{{.Ports}}")
	if err != nil || strings.TrimSpace(out) == "" {
		return []output.Result{output.Ok(cat, "ports", "No published ports listed (services may not be running)")}
	}
	return []output.Result{output.OkDetail(cat, "ports", "Published ports (compose ps)", out)}
}

var logErrorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\berror\b`),
	regexp.MustCompile(`(?i)\bfailed\b`),
	regexp.MustCompile(`(?i)\bcrash\b`),
}

func scanLogs(ctx context.Context, projectDir string) []output.Result {
	out, stderr, err := utils.RunInDir(ctx, projectDir, "docker", "compose", "logs", "--tail", "80")
	if err != nil {
		return []output.Result{output.Warn(cat, "logs", "Could not read compose logs (services may be stopped)",
			"Start services with `docker compose up -d` if you expect logs.")}
	}
	combined := out + "\n" + stderr
	var hits []string
	for _, line := range strings.Split(combined, "\n") {
		for _, re := range logErrorPatterns {
			if re.MatchString(line) {
				if len(line) > 200 {
					line = line[:200] + "…"
				}
				hits = append(hits, line)
				break
			}
		}
	}
	if len(hits) > 10 {
		hits = hits[:10]
		hits = append(hits, "... (truncated)")
	}
	if len(hits) > 0 {
		return []output.Result{output.WarnDetail(cat, "logs", "Recent log lines match error/failed/crash patterns",
			"Review full logs with `docker compose logs -f` and fix the underlying service.",
			strings.Join(hits, "\n"))}
	}
	return []output.Result{output.Ok(cat, "logs", "No obvious error/failed/crash patterns in last 80 log lines")}
}
