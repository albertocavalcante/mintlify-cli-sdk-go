package mintlify

import "errors"

// Sentinel errors.
var (
	// ErrNoRunner is returned when no npm executor is found on PATH.
	ErrNoRunner = errors.New("mintlify: no runner found (install mint, bunx, pnpm, or npx)")

	// ErrNotReady is returned when the dev server has not yet started accepting connections.
	ErrNotReady = errors.New("mintlify: dev server is not ready")

	// ErrBuildFailed is returned when `mintlify build` exits with a non-zero code.
	ErrBuildFailed = errors.New("mintlify: build failed")

	// ErrAlreadyRunning is returned when StartDev is called while a server is already running.
	ErrAlreadyRunning = errors.New("mintlify: dev server is already running")

	// ErrNotRunning is returned when Stop is called on a server that isn't running.
	ErrNotRunning = errors.New("mintlify: dev server is not running")
)
