package main

import (
	"context"
	"fmt"
	"os"

	"github.com/samirkoirala/devops-doctor/internal/output"
	"github.com/samirkoirala/devops-doctor/internal/runner"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	jsonOut bool
)

func main() {
	root := &cobra.Command{
		Use:   "devops-doctor",
		Short: "DevOps environment diagnostics (system, Docker, Compose, Kubernetes)",
		Long: `devops-doctor runs quick health checks for local and cluster tooling.
Use "check" for a full pass, or target nginx, Docker, Compose, or Kubernetes only.`,
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show extra command output for successful checks")
	root.PersistentFlags().BoolVar(&jsonOut, "json", false, "Emit results as JSON")

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Run diagnostic checks",
		Long: `Runs system checks, nginx, and Docker. If a compose file exists in the current directory tree,
compose checks run automatically. If ~/.kube/config exists, Kubernetes checks run automatically.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, runner.TargetAll)
		},
	}

	checkDocker := &cobra.Command{
		Use:   "docker",
		Short: "Docker-only checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, runner.TargetDocker)
		},
	}
	checkCompose := &cobra.Command{
		Use:   "compose",
		Short: "Docker Compose checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, runner.TargetCompose)
		},
	}
	checkK8s := &cobra.Command{
		Use:   "k8s",
		Short: "Kubernetes checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, runner.TargetK8s)
		},
	}
	checkNginx := &cobra.Command{
		Use:   "nginx",
		Short: "nginx install, config test, process, and error log scan",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, runner.TargetNginx)
		},
	}

	checkCmd.AddCommand(checkDocker, checkCompose, checkK8s, checkNginx)
	root.AddCommand(checkCmd)

	root.CompletionOptions.DisableDefaultCmd = true
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runCheck(cmd *cobra.Command, t runner.Target) error {
	ctx := context.Background()
	res := runner.Run(ctx, runner.Options{Target: t})
	f := output.NewFormatter(jsonOut, verbose)
	if cmd != nil && cmd.OutOrStdout() != nil {
		f.Out = cmd.OutOrStdout()
	}
	errCount := f.PrintResults(res)
	if errCount > 0 {
		return fmt.Errorf("%d error-level finding(s)", errCount)
	}
	return nil
}
