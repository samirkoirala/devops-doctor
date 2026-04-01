package network

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/pkg/utils"
)

// CommonDevPorts are frequently conflicting in local dev.
var CommonDevPorts = []string{"80", "443", "3000", "4200", "5000", "5432", "6379", "8000", "8080", "8443", "9000", "27017"}

// CheckListeningPorts reports ports from the list that appear in use (best-effort).
func CheckListeningPorts(ctx context.Context, ports []string) []output.Result {
	var ssSnapshot string
	if runtime.GOOS == "linux" {
		ssSnapshot, _, _ = utils.Run(ctx, "ss", "-tln")
	}
	var results []output.Result
	for _, p := range ports {
		inUse, detail := portInUse(ctx, p, ssSnapshot)
		if inUse {
			results = append(results, output.ErrDetail("docker", "port",
				fmt.Sprintf("Port %s appears to be in use", p),
				"Kill the process bound to this port (`lsof -i :PORT` / `ss -tlnp`) or change your service port in compose/Kubernetes.",
				detail))
		}
	}
	if len(results) == 0 {
		results = append(results, output.Ok("docker", "port",
			fmt.Sprintf("Scanned %d common dev ports — none obviously conflicting", len(ports))))
	}
	return results
}

func portInUse(ctx context.Context, port string, ssSnapshot string) (bool, string) {
	// lsof works on macOS/BSD and many Linux installs
	out, _, err := utils.Run(ctx, "lsof", "-nP", "-iTCP:"+port, "-sTCP:LISTEN")
	if err == nil && strings.TrimSpace(out) != "" {
		return true, out
	}
	if runtime.GOOS == "linux" && ssSnapshot != "" {
		// Match *:8080, 0.0.0.0:8080, [::]:8080
		re := regexp.MustCompile(`:` + regexp.QuoteMeta(port) + `\b`)
		if re.MatchString(ssSnapshot) {
			return true, ssSnapshot
		}
	}
	return false, ""
}
