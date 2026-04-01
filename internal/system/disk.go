package system

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/pkg/utils"
)

// checkDiskAll scans mounted filesystems from df and flags any that are nearly full (includes /).
func checkDiskAll(ctx context.Context) []output.Result {
	out, stderr, err := utils.Run(ctx, "df", "-h")
	if err != nil {
		return []output.Result{output.Warn(cat, "disk", "Could not list filesystems (df -h)",
			"Run `df -h` manually; fix permissions or PATH if this fails.")}
	}
	combined := strings.TrimSpace(out)
	if stderr != "" {
		combined += "\n" + stderr
	}
	var results []output.Result
	var scanned int
	for _, line := range strings.Split(combined, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(strings.ToLower(line), "filesystem") {
			continue
		}
		fs, mount, pct, ok := parseDFLine(line)
		if !ok || shouldSkipFS(fs, mount) {
			continue
		}
		scanned++
		switch {
		case pct >= 95:
			results = append(results, output.ErrDetail(cat, "disk",
				fmt.Sprintf("Mount %q (%s) is %d%% full — critically low space", mount, fs, pct),
				"Free space on that volume: archive or delete data, resize the disk, or move workloads.",
				line))
		case pct >= 85:
			results = append(results, output.Warn(cat, "disk",
				fmt.Sprintf("Mount %q (%s) is %d%% full", mount, fs, pct),
				"Plan cleanup or capacity increase before the filesystem fills completely."))
		}
	}
	if scanned == 0 {
		return []output.Result{output.Warn(cat, "disk", "No local disk mounts parsed from df output",
			"Verify `df -h` output on this OS; report a bug if mounts are missing.")}
	}
	if len(results) == 0 {
		return []output.Result{output.OkDetail(cat, "disk",
			fmt.Sprintf("Scanned %d mount(s) — none are above 85%% capacity", scanned),
			combined)}
	}
	return results
}

func parseDFLine(line string) (fs, mount string, pct int, ok bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return "", "", 0, false
	}
	fs = fields[0]
	pctIdx := -1
	for i, f := range fields {
		if !strings.HasSuffix(f, "%") {
			continue
		}
		pctStr := strings.TrimSuffix(f, "%")
		n, err := strconv.Atoi(pctStr)
		if err != nil || n < 0 || n > 100 {
			continue
		}
		pctIdx = i
		pct = n
		break
	}
	if pctIdx < 0 {
		return "", "", 0, false
	}
	mount = fields[len(fields)-1]
	return fs, mount, pct, true
}

func shouldSkipFS(fs, mount string) bool {
	switch {
	case strings.HasPrefix(fs, "tmpfs"), strings.HasPrefix(fs, "devtmpfs"):
		return true
	case fs == "proc", fs == "sysfs", fs == "devfs":
		return true
	case strings.HasPrefix(fs, "map "):
		return true
	case mount == "/dev", mount == "/run", mount == "/sys", mount == "/proc":
		return true
	case strings.HasPrefix(fs, "sunrpc"), strings.HasPrefix(fs, "rpc_pipefs"):
		return true
	}
	return false
}
