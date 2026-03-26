package mintlify

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/albertocavalcante/mintlify-cli-sdk-go/internal/process"
)

// ContainerRuntime specifies the container engine.
type ContainerRuntime string

const (
	Docker ContainerRuntime = "docker"
	Podman ContainerRuntime = "podman"
)

// ContainerConfig configures container-based command execution.
type ContainerConfig struct {
	Runtime ContainerRuntime // "docker" or "podman" (auto-detected if empty)
	Image   string           // container image (default: "node:22-slim")
	Pull    bool             // pull image before first use
}

func (c *ContainerConfig) runtime() string {
	if c.Runtime != "" {
		return string(c.Runtime)
	}
	return string(DetectRuntime())
}

func (c *ContainerConfig) image() string {
	if c.Image != "" {
		return c.Image
	}
	return "node:22-slim"
}

// DetectRuntime checks PATH for podman (preferred — rootless by default)
// then docker. Returns empty string if neither is found.
func DetectRuntime() ContainerRuntime {
	if _, err := exec.LookPath("podman"); err == nil {
		return Podman
	}
	if _, err := exec.LookPath("docker"); err == nil {
		return Docker
	}
	return ""
}

// ErrNoContainerRuntime is returned when neither docker nor podman is on PATH.
var ErrNoContainerRuntime = errors.New("mintlify: no container runtime found (install docker or podman)")

// ContainerCommandFunc returns a [CommandFunc] that runs commands inside
// a container with the docs dir bind-mounted at /workspace.
//
// Each invocation creates a new ephemeral container (--rm). Suitable for
// one-shot commands like validate, build, broken-links, etc.
func ContainerCommandFunc(cfg ContainerConfig) (CommandFunc, error) {
	rt := cfg.runtime()
	if rt == "" {
		return nil, ErrNoContainerRuntime
	}
	image := cfg.image()

	return func(ctx context.Context, dir string, _ string, args ...string) (string, string, int, error) {
		// Build container args: run --rm -v dir:/workspace -w /workspace image <args>
		containerArgs := []string{
			"run", "--rm",
			"-v", dir + ":/workspace:rw",
			"-w", "/workspace",
			image,
		}
		containerArgs = append(containerArgs, args...)

		cmd := exec.CommandContext(ctx, rt, containerArgs...)
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
	}, nil
}

// ContainerDevServer manages a `mintlify dev` process running inside a
// container with port forwarding.
type ContainerDevServer struct {
	proc          *process.Managed
	containerName string
	runtime       string
	port          int
	url           string
	ready         chan struct{}
	once          sync.Once
}

// StartContainerDev launches `mintlify dev` inside a container.
// The container is created with port forwarding and the docs directory
// bind-mounted at /workspace.
func (c *Client) StartContainerDev(ctx context.Context, opts DevOptions, cfg ContainerConfig) (*ContainerDevServer, error) {
	rt := cfg.runtime()
	if rt == "" {
		return nil, ErrNoContainerRuntime
	}

	port := opts.Port
	if port == 0 {
		port = 3000
	}
	image := cfg.image()
	name := fmt.Sprintf("mintlify-dev-%d", port)

	// Build the full command args for the runner.
	mintArgs := c.buildArgs("dev", "--port", fmt.Sprintf("%d", port))

	containerArgs := []string{
		"run", "--rm",
		"--name", name,
		"-v", c.dir + ":/workspace:rw",
		"-w", "/workspace",
		"-p", fmt.Sprintf("%d:%d", port, port),
		image,
	}
	containerArgs = append(containerArgs, mintArgs...)

	proc, err := process.Start(ctx, rt, containerArgs, "", nil)
	if err != nil {
		return nil, fmt.Errorf("mintlify container dev: %w", err)
	}

	s := &ContainerDevServer{
		proc:          proc,
		containerName: name,
		runtime:       rt,
		port:          port,
		url:           fmt.Sprintf("http://localhost:%d", port),
		ready:         make(chan struct{}),
	}

	go func() {
		if err := process.WaitHTTPReady(ctx, s.url); err == nil {
			s.once.Do(func() { close(s.ready) })
		}
	}()

	return s, nil
}

// URL returns the base URL of the container dev server.
func (s *ContainerDevServer) URL() string { return s.url }

// Port returns the forwarded port.
func (s *ContainerDevServer) Port() int { return s.port }

// Ready returns a channel closed when the server accepts HTTP connections.
func (s *ContainerDevServer) Ready() <-chan struct{} { return s.ready }

// WaitReady blocks until the server is accepting connections or ctx is cancelled.
func (s *ContainerDevServer) WaitReady(ctx context.Context) error {
	select {
	case <-s.ready:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-s.proc.Done():
		if err := s.proc.Err(); err != nil {
			return fmt.Errorf("container dev server exited: %w", err)
		}
		return ErrNotReady
	}
}

// Wait blocks until the container process exits.
func (s *ContainerDevServer) Wait() error {
	<-s.proc.Done()
	return s.proc.Err()
}

// Stop stops the container and waits for it to exit.
func (s *ContainerDevServer) Stop() error {
	// Signal the container to stop.
	stopCtx := context.Background()
	//nolint:gosec // container name is generated, not user input
	_ = exec.CommandContext(stopCtx, s.runtime, "stop", s.containerName).Run()
	return s.proc.Stop()
}

// Running reports whether the container process is still alive.
func (s *ContainerDevServer) Running() bool {
	return s.proc.Running()
}

// Output returns stdout from the container process.
func (s *ContainerDevServer) Output() io.ReadCloser {
	return s.proc.Stdout()
}

// Errors returns stderr from the container process.
func (s *ContainerDevServer) Errors() io.ReadCloser {
	return s.proc.Stderr()
}

// ContainerName returns the name of the running container.
func (s *ContainerDevServer) ContainerName() string {
	return s.containerName
}

// Runtime returns the container runtime being used ("docker" or "podman").
func (s *ContainerDevServer) Runtime() string {
	return s.runtime
}
