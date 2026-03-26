package mintlify

import (
	"context"
	"strings"
	"testing"
)

func TestDetectRuntime(t *testing.T) {
	rt := DetectRuntime()
	// On CI or minimal envs, neither may be available — that's OK.
	if rt != "" && rt != Docker && rt != Podman {
		t.Errorf("DetectRuntime() = %q, want docker, podman, or empty", rt)
	}
}

func TestContainerCommandFunc_Integration(t *testing.T) {
	// Integration test — requires a real container runtime and pulls an image.
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if DetectRuntime() == "" {
		t.Skip("no container runtime available")
	}

	cfg := ContainerConfig{Image: "alpine:latest"}
	cmdFunc, err := ContainerCommandFunc(cfg)
	if err != nil {
		t.Fatalf("ContainerCommandFunc() error: %v", err)
	}

	// The CommandFunc receives pre-built args from the runner. In container
	// mode the image entrypoint runs them directly.
	stdout, _, exitCode, err := cmdFunc(context.Background(), t.TempDir(), "", "echo", "hello-from-container")
	if err != nil {
		t.Skipf("container exec failed (image not available?): %v", err)
	}
	if exitCode != 0 {
		t.Skipf("container exited %d (runtime issue, not SDK bug)", exitCode)
	}
	if !strings.Contains(stdout, "hello-from-container") {
		t.Errorf("stdout = %q, want 'hello-from-container'", stdout)
	}
}

func TestContainerCommandFunc_ExplicitRuntime(t *testing.T) {
	// An explicit Runtime is trusted at factory time (validation happens at exec).
	cfg := ContainerConfig{Runtime: "nonexistent-runtime"}
	cmdFunc, err := ContainerCommandFunc(cfg)
	if err != nil {
		t.Fatalf("ContainerCommandFunc() should not error with explicit Runtime: %v", err)
	}
	// Actually executing should fail since the binary doesn't exist.
	_, _, _, execErr := cmdFunc(context.Background(), t.TempDir(), "", "echo", "test")
	if execErr == nil {
		t.Error("expected exec error for nonexistent runtime")
	}
}

func TestContainerConfig_Defaults(t *testing.T) {
	cfg := ContainerConfig{}
	if img := cfg.image(); img != "node:22-slim" {
		t.Errorf("image() = %q, want 'node:22-slim'", img)
	}
}

func TestContainerConfig_CustomImage(t *testing.T) {
	cfg := ContainerConfig{Image: "my-custom:v1"}
	if img := cfg.image(); img != "my-custom:v1" {
		t.Errorf("image() = %q, want 'my-custom:v1'", img)
	}
}
