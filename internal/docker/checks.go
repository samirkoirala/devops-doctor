package docker

import (
	"context"
	"strings"

	"github.com/samirkoirala/devops-doctor/internal/network"
	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/pkg/utils"
)

const cat = "docker"

// Run executes Docker-related checks (parallel sub-groups).
func Run(ctx context.Context) []output.Result {
	type job struct {
		fn func(context.Context) []output.Result
	}
	jobs := []job{
		{fn: checkInstalled},
		{fn: checkDaemon},
		{fn: checkPs},
		{fn: func(c context.Context) []output.Result { return network.CheckListeningPorts(c, network.CommonDevPorts) }},
		{fn: checkDiskUsage},
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

func checkInstalled(ctx context.Context) []output.Result {
	_, _, err := utils.Run(ctx, "docker", "version", "--format", "{{.Client.Version}}")
	if err != nil {
		return []output.Result{output.Err(cat, "install", "Docker CLI not found or not working",
			"Install Docker Desktop (macOS/Windows) or docker.io (Linux) and ensure `docker` is on PATH.")}
	}
	out, _, err := utils.Run(ctx, "docker", "version", "--format", "Client: {{.Client.Version}}")
	if err != nil {
		return []output.Result{output.Ok(cat, "install", "Docker CLI is installed")}
	}
	return []output.Result{output.OkDetail(cat, "install", "Docker CLI is installed", out)}
}

func checkDaemon(ctx context.Context) []output.Result {
	_, stderr, err := utils.Run(ctx, "docker", "info")
	if err != nil {
		msg := "Docker daemon does not respond"
		if strings.Contains(strings.ToLower(stderr), "permission denied") {
			return []output.Result{output.ErrDetail(cat, "daemon", msg+": permission denied",
				"Add your user to the `docker` group or use sudo; on macOS start Docker Desktop.",
				stderr)}
		}
		if strings.Contains(strings.ToLower(stderr), "cannot connect") || strings.Contains(strings.ToLower(stderr), "connection refused") {
			return []output.Result{output.ErrDetail(cat, "daemon", msg+" (cannot connect)",
				"Start Docker Desktop or the docker service: `sudo systemctl start docker` on Linux.",
				stderr)}
		}
		return []output.Result{output.ErrDetail(cat, "daemon", msg,
			"Run `docker info` manually; ensure the daemon is running and your context is correct.",
			stderr)}
	}
	return []output.Result{output.Ok(cat, "daemon", "Docker daemon is reachable")}
}

func checkPs(ctx context.Context) []output.Result {
	out, stderr, err := utils.Run(ctx, "docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	if err != nil {
		return []output.Result{output.ErrDetail(cat, "ps", "Could not list containers",
			"Fix daemon connectivity first, then retry `docker ps -a`.",
			stderr)}
	}
	if strings.TrimSpace(out) == "" {
		return []output.Result{output.Ok(cat, "ps", "No containers reported (empty list)")}
	}
	return []output.Result{output.OkDetail(cat, "ps", "Container list", out)}
}

func checkDiskUsage(ctx context.Context) []output.Result {
	out, stderr, err := utils.Run(ctx, "docker", "system", "df", "-v")
	if err != nil {
		// Older docker without -v
		out, stderr, err = utils.Run(ctx, "docker", "system", "df")
	}
	if err != nil {
		return []output.Result{output.Warn(cat, "disk", "Could not read docker system df",
			"Upgrade Docker or run `docker system df` manually to inspect image/volume usage.")}
	}
	combined := out
	if stderr != "" {
		combined += "\n" + stderr
	}
	return []output.Result{output.OkDetail(cat, "disk", "Docker disk usage", combined)}
}
