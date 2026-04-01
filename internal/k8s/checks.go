package k8s

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/pkg/utils"
)

const cat = "k8s"

// HasKubeconfig returns true if default kubeconfig path exists and is non-empty.
func HasKubeconfig() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	p := filepath.Join(home, ".kube", "config")
	st, err := os.Stat(p)
	return err == nil && !st.IsDir() && st.Size() > 0
}

// Run executes Kubernetes diagnostics.
func Run(ctx context.Context) []output.Result {
	type job struct {
		fn func(context.Context) []output.Result
	}
	jobs := []job{
		{fn: checkKubectl},
		{fn: checkContext},
		{fn: checkCluster},
		{fn: checkNodes},
		{fn: checkPods},
	}
	ch := make(chan []output.Result, len(jobs))
	for _, j := range jobs {
		j := j
		go func() { ch <- j.fn(ctx) }()
	}
	var all []output.Result
	for range jobs {
		all = append(all, <-ch...)
	}
	return all
}

func checkKubectl(ctx context.Context) []output.Result {
	_, _, err := utils.Run(ctx, "kubectl", "version", "--client", "-o", "yaml")
	if err != nil {
		return []output.Result{output.Err(cat, "kubectl", "kubectl is not installed or not on PATH",
			"Install kubectl from https://kubernetes.io/docs/tasks/tools/ and ensure it is in PATH.")}
	}
	out, _, err := utils.Run(ctx, "kubectl", "version", "--client", "--short")
	if err != nil {
		out, _, _ = utils.Run(ctx, "kubectl", "version", "--client")
	}
	return []output.Result{output.OkDetail(cat, "kubectl", "kubectl client available", out)}
}

func checkContext(ctx context.Context) []output.Result {
	out, stderr, err := utils.Run(ctx, "kubectl", "config", "current-context")
	if err != nil || strings.TrimSpace(out) == "" {
		return []output.Result{output.Warn(cat, "context", "No current kubectl context",
			"Run `kubectl config get-contexts` and `kubectl config use-context <name>`.")}
	}
	detail := strings.TrimSpace(stderr)
	return []output.Result{output.OkDetail(cat, "context", "Current context: "+strings.TrimSpace(out), detail)}
}

func checkCluster(ctx context.Context) []output.Result {
	out, stderr, err := utils.Run(ctx, "kubectl", "cluster-info")
	if err != nil {
		return []output.Result{output.ErrDetail(cat, "cluster", "Cannot reach Kubernetes cluster",
			"Check VPN/network, kubeconfig server URL, and credentials; try `kubectl cluster-info` for details.",
			stderr)}
	}
	return []output.Result{output.OkDetail(cat, "cluster", "Cluster API is reachable", out)}
}

func checkNodes(ctx context.Context) []output.Result {
	out, stderr, err := utils.Run(ctx, "kubectl", "get", "nodes", "-o", "wide", "--no-headers")
	if err != nil {
		return []output.Result{output.ErrDetail(cat, "nodes", "Could not list cluster nodes",
			"Verify RBAC allows node read and that the API server is healthy.",
			stderr)}
	}
	var notReady []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			status := fields[1]
			if strings.EqualFold(status, "NotReady") || strings.EqualFold(status, "Unknown") {
				notReady = append(notReady, line)
			}
		}
	}
	if len(notReady) > 0 {
		return []output.Result{output.WarnDetail(cat, "nodes", "Some nodes are not Ready",
			"Inspect node conditions with `kubectl describe node <name>` and fix kubelet/CNI/disk pressure.",
			strings.Join(notReady, "\n"))}
	}
	return []output.Result{output.OkDetail(cat, "nodes", "Node status", out)}
}

func checkPods(ctx context.Context) []output.Result {
	out, stderr, err := utils.Run(ctx, "kubectl", "get", "pods", "-A", "--no-headers")
	if err != nil {
		return []output.Result{output.Warn(cat, "pods", "Could not list pods in all namespaces", stderr)}
	}
	var bad []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "crashloopbackoff") ||
			strings.Contains(lower, " pending") ||
			strings.Contains(lower, "\tpending\t") ||
			strings.Contains(lower, "imagepullbackoff") ||
			strings.Contains(lower, "errimagepull") {
			bad = append(bad, line)
		}
	}
	if len(bad) > 15 {
		bad = bad[:15]
		bad = append(bad, "... (truncated)")
	}
	if len(bad) > 0 {
		return []output.Result{output.WarnDetail(cat, "pods", "Unhealthy or pending pods detected (sample)",
			"Run `kubectl describe pod` / `kubectl logs` for failing workloads; fix images, resources, or probes.",
			strings.Join(bad, "\n"))}
	}
	return []output.Result{output.Ok(cat, "pods", "No obvious CrashLoopBackOff / Pending / image pull failures in pod list")}
}
