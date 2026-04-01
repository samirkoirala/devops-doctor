# devops-doctor

A small **Go** CLI that inspects your machine and toolchain: **system** health, **nginx**, **Docker**, **Docker Compose**, and **Kubernetes**. It prints clear pass/warn/fail lines with actionable suggestions.

## Requirements

- Go **1.22+** (to build)
- Optional runtime tools checked by the CLI: `nginx`, `docker`, `docker compose`, `kubectl`, `curl`, `ping`, `dig`/`nslookup`/`getent`, `df`, `lsof`/`ss`, `pgrep`, `tail`

## Install

From the repository root:

```bash
go build -o devops-doctor ./cmd/devops-doctor
sudo mv devops-doctor /usr/local/bin/   # optional
```

Or install directly with Go:

```bash
go install github.com/samirkoirala/devops-doctor/cmd/devops-doctor@latest
```

(Adjust the module path if you publish under a different import path.)

## Usage

```text
devops-doctor check              # system + nginx + docker; compose if compose file exists; k8s if ~/.kube/config exists
devops-doctor check nginx        # nginx only (version, nginx -t, process, error.log tail)
devops-doctor check docker       # Docker only
devops-doctor check compose      # Compose only (walks up dirs for compose file)
devops-doctor check k8s          # Kubernetes only

Global flags:
  --verbose, -v   Show extra command output for successful checks
  --json          Machine-readable JSON (summary + results)
```

Exit code **1** if any check returned **error** status (warnings do not fail the command).

## What it checks

| Area | Checks |
|------|--------|
| **System** | Load average, memory summary, **all local mounts** from `df -h` including `/` (warn ≥85%, error ≥95%), HTTPS (and ICMP fallback) connectivity, DNS |
| **nginx** | Binary on PATH, **`nginx -t`**, running processes (`pgrep`), tail of **error.log** at common paths (`[emerg]`/`[alert]`/`[crit]`/`[error]`, skipping common missing-file noise like `favicon.ico`) |
| **Docker** | CLI installed, daemon reachable, `docker ps -a`, common dev **port conflicts**, `docker system df` |
| **Compose** | Compose file discovery, `docker compose ps`, unhealthy/restarting/exited containers, published ports, log scan for `error` / `failed` / `crash` |
| **Kubernetes** | `kubectl` client, current context, cluster reachability, nodes, problematic pods (e.g. CrashLoopBackOff, Pending, image pull errors) |

Commands use **timeouts** (default **25s** per invocation via `context`) and run independent check groups **in parallel** where practical.

## Example output

```text
$ devops-doctor check --verbose

SYSTEM
  ✔ Load average: { 1.42 1.38 1.35 }
  ✔ Memory (vm_stat excerpt)
    Pages free: 123456.
    ...
  ✔ Scanned 4 mount(s) — none are above 85% capacity
    (full `df -h` output shown with --verbose)
  ✔ Outbound HTTPS connectivity OK (https://www.cloudflare.com)
    HTTP 200
  ✔ DNS resolution working
    104.16.132.229  cloudflare.com

NGINX
  ✔ nginx is installed
  ✔ nginx configuration syntax is OK
  ✔ nginx process(es) running
  ✔ No serious errors in recent log tail: error.log

DOCKER
  ✔ Docker CLI is installed
    Client: 27.3.1
  ✔ Docker daemon is reachable
  ✔ Container list
    NAMES     STATUS    PORTS
    web-1     Up 2h     0.0.0.0:3000->3000/tcp
  ✖ Port 5432 appears to be in use
    💡 Suggestion: Kill the process bound to this port (`lsof -i :PORT` / `ss -tlnp`) or change your service port in compose/Kubernetes.
    COMMAND  PID USER   FD   TYPE ...
  ✔ Docker disk usage
    TYPE   TOTAL   ACTIVE   SIZE    RECLAIMABLE
    ...

COMPOSE
  ✔ Compose file found: /Users/me/app/docker-compose.yml
  ✔ docker compose ps -a
  ✔ No obvious unhealthy/restarting/exited states in compose ps
  ...

K8S
  ✔ kubectl client available
  ✔ Current context: docker-desktop
  ✔ Cluster API is reachable
  ...
```

JSON (excerpt):

```json
{
  "results": [
    {
      "category": "system",
      "check": "disk",
      "status": "success",
      "message": "Scanned 3 mount(s) — none are above 85% capacity",
      "detail": "Filesystem      Size  Used Avail Use% Mounted on\n/dev/sda1        99G   45G   50G  48% /"
    }
  ],
  "summary": {
    "success": 12,
    "warning": 1,
    "error": 1
  }
}
```

## Project layout

```text
cmd/devops-doctor/     # Cobra CLI entrypoint
internal/
  system/              # CPU/load, memory, disk (all mounts via df), network/DNS
  nginx/               # nginx -t, process, error log tail
  docker/              # Docker daemon, ps, disk, delegates port scan
  compose/             # Compose file discovery and compose checks
  k8s/                 # kubectl / cluster / workloads
  network/             # Listening-port heuristics
  output/              # Result model + human/JSON formatter
  runner/              # Orchestration and ordering
pkg/utils/             # Command execution with timeouts
```

## License

MIT (add a `LICENSE` file if you publish the project).
