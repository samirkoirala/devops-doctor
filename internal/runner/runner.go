package runner

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/samirkoirala/devops-doctor/internal/compose"
	"github.com/samirkoirala/devops-doctor/internal/docker"
	"github.com/samirkoirala/devops-doctor/internal/k8s"
	"github.com/samirkoirala/devops-doctor/internal/nginx"
	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/internal/system"
)

// Target selects which check suites to run.
type Target string

const (
	TargetAll     Target = "all"
	TargetDocker  Target = "docker"
	TargetCompose Target = "compose"
	TargetK8s     Target = "k8s"
	TargetNginx   Target = "nginx"
)

// Options control detection and scoping.
type Options struct {
	Target Target
}

// Run executes checks per target and returns ordered results.
func Run(ctx context.Context, opt Options) []output.Result {
	var parts [][]output.Result
	switch opt.Target {
	case TargetDocker:
		parts = append(parts, docker.Run(ctx))
	case TargetCompose:
		wd, err := os.Getwd()
		if err != nil {
			wd = "."
		}
		dir, name := compose.FindComposeFile(wd)
		parts = append(parts, compose.Run(ctx, dir, name))
	case TargetK8s:
		parts = append(parts, k8s.Run(ctx))
	case TargetNginx:
		parts = append(parts, nginx.Run(ctx))
	default: // all
		parts = append(parts, system.Run(ctx))
		parts = append(parts, nginx.Run(ctx))
		parts = append(parts, docker.Run(ctx))
		wd, err := os.Getwd()
		if err != nil {
			wd = "."
		}
		dir, name := compose.FindComposeFile(wd)
		if dir != "" {
			parts = append(parts, compose.Run(ctx, dir, name))
		}
		if k8s.HasKubeconfig() {
			parts = append(parts, k8s.Run(ctx))
		}
	}
	var merged []output.Result
	for _, p := range parts {
		merged = append(merged, p...)
	}
	sortResults(merged)
	return merged
}

func categoryOrder(c string) int {
	switch strings.ToLower(c) {
	case "system":
		return 0
	case "nginx":
		return 1
	case "docker":
		return 2
	case "compose":
		return 3
	case "k8s":
		return 4
	default:
		return 9
	}
}

func sortResults(r []output.Result) {
	sort.SliceStable(r, func(i, j int) bool {
		oi, oj := categoryOrder(r[i].Category), categoryOrder(r[j].Category)
		if oi != oj {
			return oi < oj
		}
		if r[i].SortOrder != r[j].SortOrder {
			return r[i].SortOrder < r[j].SortOrder
		}
		return r[i].Check < r[j].Check
	})
}
