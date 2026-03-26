// Package mintlify provides a typed Go SDK for the Mintlify CLI.
//
// It wraps the Mintlify CLI as a subprocess with structured result types,
// pluggable npm runners, and first-class dev server lifecycle management.
//
// Basic usage:
//
//	client, err := mintlify.New("./docs")
//	result, err := client.Validate(ctx, mintlify.ValidateOptions{})
//	if !result.OK {
//	    for _, e := range result.Errors {
//	        fmt.Printf("%s:%d — %s\n", e.File, e.Line, e.Message)
//	    }
//	}
package mintlify

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CommandFunc executes a command and returns stdout, stderr, exit code.
// Inject a custom implementation via [WithCommandFunc] for testing.
type CommandFunc func(ctx context.Context, dir string, name string, args ...string) (stdout, stderr string, exitCode int, err error)

// Client wraps the Mintlify CLI with typed methods for each command.
type Client struct {
	dir     string
	runner  *Runner
	cmd     CommandFunc
	timeout time.Duration
}

// Option configures a [Client].
type Option func(*Client)

// WithRunner overrides auto-detected runner selection.
func WithRunner(r *Runner) Option {
	return func(c *Client) { c.runner = r }
}

// WithCommandFunc injects a custom command executor (for testing).
func WithCommandFunc(f CommandFunc) Option {
	return func(c *Client) { c.cmd = f }
}

// WithTimeout overrides the default command timeout (2 minutes).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// New creates a Client for the Mintlify project at dir.
// It auto-detects the best available runner unless overridden via [WithRunner].
func New(dir string, opts ...Option) (*Client, error) {
	c := &Client{
		dir:     dir,
		cmd:     defaultCmd,
		timeout: 2 * time.Minute,
	}
	for _, o := range opts {
		o(c)
	}
	if c.runner == nil {
		c.runner = DetectRunner()
	}
	if c.runner == nil {
		return nil, ErrNoRunner
	}
	return c, nil
}

// Dir returns the working directory of the client.
func (c *Client) Dir() string { return c.dir }

// Runner returns the active runner.
func (c *Client) Runner() *Runner { return c.runner }

// Timeout returns the configured command timeout.
func (c *Client) Timeout() time.Duration { return c.timeout }

// run executes a mintlify command, prepending runner args and applying
// the configured timeout. A non-zero exit code is treated as an error.
func (c *Client) run(ctx context.Context, args ...string) (stdout, stderr string, err error) {
	stdout, stderr, exitCode, err := c.runRaw(ctx, args...)
	if err != nil {
		return "", stderr, err
	}
	if exitCode != 0 {
		return stdout, stderr, fmt.Errorf("mintlify %s: exit code %d: %s",
			strings.Join(args, " "), exitCode, strings.TrimSpace(stderr))
	}
	return stdout, stderr, nil
}

// runRaw executes a mintlify command but returns the exit code instead of
// treating non-zero as an error. Useful for commands like validate where
// non-zero means "issues found" rather than failure.
func (c *Client) runRaw(ctx context.Context, args ...string) (stdout, stderr string, exitCode int, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	fullArgs := make([]string, 0, len(c.runner.Args)+len(args))
	fullArgs = append(fullArgs, c.runner.Args...)
	fullArgs = append(fullArgs, args...)

	return c.cmd(ctx, c.dir, c.runner.Cmd, fullArgs...)
}

// buildArgs prepends runner args to the given command args.
func (c *Client) buildArgs(args ...string) []string {
	fullArgs := make([]string, 0, len(c.runner.Args)+len(args))
	fullArgs = append(fullArgs, c.runner.Args...)
	fullArgs = append(fullArgs, args...)
	return fullArgs
}

// defaultCmd executes commands via os/exec.
func defaultCmd(ctx context.Context, dir string, name string, args ...string) (string, string, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return "", "", -1, err
		}
	}
	return stdoutBuf.String(), stderrBuf.String(), exitCode, nil
}
