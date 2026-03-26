// Package process provides a managed subprocess abstraction with health polling.
package process

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Managed wraps an [exec.Cmd] with lifecycle management: start, health
// polling, graceful shutdown, and output streaming.
type Managed struct {
	cmd       *exec.Cmd
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	exitErr   error          // set once by the wait goroutine
	done      chan struct{}   // closed when process exits
	mu        sync.Mutex
	running   bool
	cancelCtx context.CancelFunc
}

// Start launches the command. The process is killed when ctx is cancelled.
func Start(ctx context.Context, name string, args []string, dir string, env []string) (*Managed, error) {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start: %w", err)
	}

	m := &Managed{
		cmd:       cmd,
		stdout:    stdout,
		stderr:    stderr,
		done:      make(chan struct{}),
		running:   true,
		cancelCtx: cancel,
	}

	go func() {
		m.exitErr = cmd.Wait()
		m.mu.Lock()
		m.running = false
		m.mu.Unlock()
		close(m.done)
	}()

	return m, nil
}

// Stdout returns the process stdout for streaming.
func (m *Managed) Stdout() io.ReadCloser { return m.stdout }

// Stderr returns the process stderr for streaming.
func (m *Managed) Stderr() io.ReadCloser { return m.stderr }

// Done returns a channel that is closed when the process exits.
// Use [Managed.Err] to retrieve the exit error afterward.
func (m *Managed) Done() <-chan struct{} { return m.done }

// Err returns the process exit error. Only valid after [Managed.Done] is closed.
func (m *Managed) Err() error { return m.exitErr }

// Running reports whether the process is still alive.
func (m *Managed) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// Stop cancels the context (sends SIGKILL on most platforms) and waits
// for the process to exit. Safe to call multiple times.
func (m *Managed) Stop() error {
	m.cancelCtx()
	<-m.done
	return m.exitErr
}

// WaitHTTPReady polls the given URL until it returns a non-5xx status or
// the context is cancelled. Uses exponential backoff starting at 200ms.
func WaitHTTPReady(ctx context.Context, url string) error {
	backoff := 200 * time.Millisecond
	client := &http.Client{Timeout: 2 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := client.Get(url) //nolint:noctx // short-lived health probe
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		if backoff < 2*time.Second {
			backoff = backoff * 3 / 2
		}
	}
}
