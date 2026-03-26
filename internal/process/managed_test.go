package process

import (
	"context"
	"testing"
	"time"
)

func TestStart_SimpleCommand(t *testing.T) {
	ctx := context.Background()
	m, err := Start(ctx, "echo", []string{"hello"}, "", nil)
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	<-m.Done()
	if err := m.Err(); err != nil {
		t.Fatalf("process exited with error: %v", err)
	}
	if m.Running() {
		t.Error("Running() = true after process exited")
	}
}

func TestStart_Stop(t *testing.T) {
	ctx := context.Background()
	m, err := Start(ctx, "sleep", []string{"60"}, "", nil)
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if !m.Running() {
		t.Error("Running() = false, want true")
	}

	_ = m.Stop()

	if m.Running() {
		t.Error("Running() = true after Stop()")
	}
}

func TestStop_Idempotent(t *testing.T) {
	ctx := context.Background()
	m, err := Start(ctx, "sleep", []string{"60"}, "", nil)
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	_ = m.Stop()
	_ = m.Stop() // must not deadlock
}

func TestStart_InvalidCommand(t *testing.T) {
	ctx := context.Background()
	_, err := Start(ctx, "nonexistent-binary-that-does-not-exist", nil, "", nil)
	if err == nil {
		t.Fatal("expected error for invalid command")
	}
}

func TestWaitHTTPReady_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err := WaitHTTPReady(ctx, "http://127.0.0.1:1") // port 1 should never respond
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
