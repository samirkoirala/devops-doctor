package utils

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// DefaultCommandTimeout is applied when ctx has no deadline.
const DefaultCommandTimeout = 25 * time.Second

// Run executes name with args. Uses ctx deadline or DefaultCommandTimeout.
func Run(ctx context.Context, name string, args ...string) (stdout, stderr string, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, DefaultCommandTimeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}

// RunSimple is Run without custom context (default timeout).
func RunSimple(name string, args ...string) (stdout, stderr string, err error) {
	return Run(context.Background(), name, args...)
}

// RunInDir runs a command with Dir set to workDir. Uses same timeout rules as Run.
func RunInDir(ctx context.Context, workDir, name string, args ...string) (stdout, stderr string, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, DefaultCommandTimeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}

// ExitError extracts exit code when err is *exec.ExitError.
func ExitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}
