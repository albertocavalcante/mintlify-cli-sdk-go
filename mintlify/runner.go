package mintlify

import "os/exec"

// Runner describes an npm executor that can invoke the Mintlify CLI.
type Runner struct {
	Name string   // human-friendly name: "mint", "bunx", "pnpm-dlx", "npx"
	Cmd  string   // binary to exec
	Args []string // args prepended before the mintlify subcommand
}

// knownRunners lists runners in priority order.
var knownRunners = []Runner{
	{Name: "mint", Cmd: "mint", Args: nil},
	{Name: "bunx", Cmd: "bunx", Args: []string{"mintlify"}},
	{Name: "pnpm-dlx", Cmd: "pnpm", Args: []string{"dlx", "mintlify"}},
	{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}},
}

// DetectRunner checks PATH for known runners and returns the first one
// found. Returns nil if none are available.
func DetectRunner() *Runner {
	for i := range knownRunners {
		if _, err := exec.LookPath(knownRunners[i].Cmd); err == nil {
			return &knownRunners[i]
		}
	}
	return nil
}

// DetectRunnerByName returns the runner with the given name if its binary
// is available on PATH. Returns nil otherwise.
func DetectRunnerByName(name string) *Runner {
	for i := range knownRunners {
		if knownRunners[i].Name == name {
			if _, err := exec.LookPath(knownRunners[i].Cmd); err == nil {
				return &knownRunners[i]
			}
			return nil
		}
	}
	return nil
}
