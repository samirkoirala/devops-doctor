package system

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/pkg/utils"
)

const cat = "system"

// Run executes all system diagnostics (parallel where independent).
func Run(ctx context.Context) []output.Result {
	type job struct {
		fn func(context.Context) []output.Result
	}
	jobs := []job{
		{fn: checkLoad},
		{fn: checkMemory},
		{fn: checkDiskAll},
		{fn: checkInternet},
		{fn: checkDNS},
	}
	var all []output.Result
	ch := make(chan []output.Result, len(jobs))
	for _, j := range jobs {
		j := j
		go func() {
			ch <- j.fn(ctx)
		}()
	}
	for range jobs {
		all = append(all, <-ch...)
	}
	return all
}

func checkLoad(ctx context.Context) []output.Result {
	var msg string
	switch runtime.GOOS {
	case "darwin":
		out, _, err := utils.Run(ctx, "sysctl", "-n", "vm.loadavg")
		if err != nil {
			return []output.Result{output.Err(cat, "cpu", "Could not read CPU load average",
				"Ensure sysctl is available; on Linux use procps.")}
		}
		msg = fmt.Sprintf("Load average: %s", strings.TrimSpace(out))
	case "linux":
		data, err := os.ReadFile("/proc/loadavg")
		if err != nil {
			return []output.Result{output.Err(cat, "cpu", "Could not read /proc/loadavg",
				"Check permissions and that you are on Linux.")}
		}
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			msg = fmt.Sprintf("Load average (1/5/15m): %s %s %s", fields[0], fields[1], fields[2])
		} else {
			msg = string(data)
		}
	default:
		return []output.Result{output.Warn(cat, "cpu", "CPU load check not implemented for "+runtime.GOOS,
			"Inspect Task Manager / Activity Monitor manually.")}
	}
	return []output.Result{output.OkDetail(cat, "cpu", msg,
		"If load stays above CPU count for long periods, reduce workload or scale out.")}
}

func checkMemory(ctx context.Context) []output.Result {
	switch runtime.GOOS {
	case "linux":
		out, _, err := utils.Run(ctx, "free", "-h")
		if err != nil {
			return []output.Result{output.Warn(cat, "memory", "Could not run free -h", "Install procps.")}
		}
		return []output.Result{output.OkDetail(cat, "memory", "Memory summary (free -h)", out)}
	case "darwin":
		out, _, err := utils.Run(ctx, "vm_stat")
		if err != nil {
			return []output.Result{output.Warn(cat, "memory", "Could not run vm_stat", "Check macOS system tools.")}
		}
		lines := strings.Split(out, "\n")
		n := min(8, len(lines))
		summary := strings.Join(lines[:n], "\n")
		return []output.Result{output.OkDetail(cat, "memory", "Memory (vm_stat excerpt)", summary)}
	default:
		return []output.Result{output.Warn(cat, "memory", "Memory check not implemented for "+runtime.GOOS,
			"Use system monitoring tools for this OS.")}
	}
}

func checkInternet(ctx context.Context) []output.Result {
	urls := []string{
		"https://1.1.1.1",
		"https://www.cloudflare.com",
		"https://example.com",
	}
	var lastErr string
	for _, u := range urls {
		out, stderr, err := utils.Run(ctx, "curl", "-sS", "-L", "-o", "/dev/null", "-w", "%{http_code}", "--max-time", "10", u)
		if err == nil && out != "" && out != "000" {
			return []output.Result{output.OkDetail(cat, "internet", "Outbound HTTPS connectivity OK ("+u+")", "HTTP "+out)}
		}
		if stderr != "" {
			lastErr = stderr
		} else if err != nil {
			lastErr = err.Error()
		}
	}
	// ICMP fallback (Unix)
	if runtime.GOOS != "windows" {
		var pingErr error
		if runtime.GOOS == "darwin" {
			_, _, pingErr = utils.Run(ctx, "ping", "-c", "1", "1.1.1.1")
		} else {
			_, _, pingErr = utils.Run(ctx, "ping", "-c", "1", "-W", "3", "1.1.1.1")
		}
		if pingErr == nil {
			return []output.Result{output.Ok(cat, "internet", "ICMP reachability OK (1.1.1.1)")}
		}
	}
	return []output.Result{output.ErrDetail(cat, "internet",
		"No reliable outbound connectivity (HTTPS and ICMP failed)",
		"Check Wi‑Fi/VPN, proxy settings, firewall, and corporate egress rules.",
		lastErr)}
}

func checkDNS(ctx context.Context) []output.Result {
	hosts := []string{"google.com", "cloudflare.com"}
	var ok bool
	var lastOut string
	for _, h := range hosts {
		out, _, err := utils.Run(ctx, "getent", "hosts", h)
		if err == nil && out != "" {
			ok = true
			lastOut = out
			break
		}
		// macOS often has no getent
		out2, _, err2 := utils.Run(ctx, "dig", "+short", h)
		if err2 == nil && strings.TrimSpace(out2) != "" {
			ok = true
			lastOut = out2
			break
		}
		out3, _, err3 := utils.Run(ctx, "nslookup", h)
		if err3 == nil && strings.Contains(strings.ToLower(out3), "address") {
			ok = true
			lastOut = out3
			break
		}
	}
	if ok {
		return []output.Result{output.OkDetail(cat, "dns", "DNS resolution working", lastOut)}
	}
	return []output.Result{output.Err(cat, "dns", "DNS resolution failed for test hosts",
		"Fix /etc/resolv.conf, router DNS, or VPN split-DNS; try dig/nslookup manually.")}
}
