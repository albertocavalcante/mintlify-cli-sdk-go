package mintlify

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/albertocavalcante/mintlify-cli-sdk-go/internal/process"
)

// DevServer manages a running `mintlify dev` process.
type DevServer struct {
	proc  *process.Managed
	port  int
	url   string
	ready chan struct{} // closed when server accepts connections
	once  sync.Once
}

// StartDev launches `mintlify dev` as a managed subprocess.
// Cancel the context to stop the server. Use [DevServer.WaitReady] to block
// until the server is accepting HTTP connections.
func (c *Client) StartDev(ctx context.Context, opts DevOptions) (*DevServer, error) {
	port := opts.Port
	if port == 0 {
		port = 3000
	}

	args := c.buildArgs("dev", "--port", fmt.Sprintf("%d", port))

	proc, err := process.Start(ctx, c.runner.Cmd, args, c.dir, nil)
	if err != nil {
		return nil, fmt.Errorf("mintlify dev: %w", err)
	}

	s := &DevServer{
		proc:  proc,
		port:  port,
		url:   fmt.Sprintf("http://localhost:%d", port),
		ready: make(chan struct{}),
	}

	// Background goroutine: poll for readiness.
	go func() {
		if err := process.WaitHTTPReady(ctx, s.url); err == nil {
			s.once.Do(func() { close(s.ready) })
		}
	}()

	return s, nil
}

// URL returns the base URL of the dev server (e.g. "http://localhost:3000").
func (s *DevServer) URL() string { return s.url }

// Port returns the port the dev server is listening on.
func (s *DevServer) Port() int { return s.port }

// Ready returns a channel that is closed when the server is accepting
// HTTP connections. If the server fails to start, the channel is never closed.
func (s *DevServer) Ready() <-chan struct{} { return s.ready }

// WaitReady blocks until the dev server is accepting connections or the
// context is cancelled.
func (s *DevServer) WaitReady(ctx context.Context) error {
	select {
	case <-s.ready:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-s.proc.Done():
		if err := s.proc.Err(); err != nil {
			return fmt.Errorf("dev server exited: %w", err)
		}
		return ErrNotReady
	}
}

// Wait blocks until the dev server process exits and returns the exit error.
func (s *DevServer) Wait() error {
	<-s.proc.Done()
	return s.proc.Err()
}

// Stop gracefully shuts down the dev server. Safe to call multiple times.
func (s *DevServer) Stop() error {
	return s.proc.Stop()
}

// Running reports whether the dev server process is still alive.
func (s *DevServer) Running() bool {
	return s.proc.Running()
}

// Output returns a reader that streams the server's stdout.
func (s *DevServer) Output() io.ReadCloser {
	return s.proc.Stdout()
}

// Errors returns a reader that streams the server's stderr.
func (s *DevServer) Errors() io.ReadCloser {
	return s.proc.Stderr()
}
