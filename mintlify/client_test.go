package mintlify

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// mockCmd returns a CommandFunc that returns canned output.
func mockCmd(stdout, stderr string, exitCode int) CommandFunc {
	return func(_ context.Context, _ string, _ string, _ ...string) (string, string, int, error) {
		return stdout, stderr, exitCode, nil
	}
}

// recordingCmd records the command invocation.
type recordingCmd struct {
	name string
	args []string
	dir  string
}

func mockCmdRecording(stdout, stderr string, exitCode int, rec *recordingCmd) CommandFunc {
	return func(_ context.Context, dir string, name string, args ...string) (string, string, int, error) {
		rec.dir = dir
		rec.name = name
		rec.args = args
		return stdout, stderr, exitCode, nil
	}
}

func newTestClient(t *testing.T, runner *Runner, cmd CommandFunc) *Client {
	t.Helper()
	c, err := New("/test/dir",
		WithRunner(runner),
		WithCommandFunc(cmd),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return c
}

// ── Constructor tests ───────────────────────────────────────────────

func TestNew_NoRunnerDetected(t *testing.T) {
	_, err := New("/test", WithCommandFunc(mockCmd("", "", 0)))
	if err != nil && !strings.Contains(err.Error(), "no runner found") {
		t.Fatalf("New() unexpected error: %v", err)
	}
}

func TestNew_WithRunner(t *testing.T) {
	r := &Runner{Name: "test", Cmd: "test-cmd", Args: []string{"arg1"}}
	c := newTestClient(t, r, mockCmd("", "", 0))

	if c.Runner().Name != "test" {
		t.Errorf("Runner().Name = %q, want %q", c.Runner().Name, "test")
	}
	if c.Dir() != "/test/dir" {
		t.Errorf("Dir() = %q, want %q", c.Dir(), "/test/dir")
	}
}

func TestNew_WithTimeout(t *testing.T) {
	r := &Runner{Name: "test", Cmd: "test-cmd"}
	c, err := New("/test",
		WithRunner(r),
		WithCommandFunc(mockCmd("", "", 0)),
		WithTimeout(5*time.Minute),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if c.Timeout() != 5*time.Minute {
		t.Errorf("Timeout() = %v, want %v", c.Timeout(), 5*time.Minute)
	}
}

// ── Arg construction per runner ─────────────────────────────────────

func TestArgConstruction(t *testing.T) {
	tests := []struct {
		name     string
		runner   Runner
		command  []string
		wantArgs []string
	}{
		{
			name:     "mint_system",
			runner:   Runner{Name: "mint", Cmd: "mint", Args: nil},
			command:  []string{"version"},
			wantArgs: []string{"version"},
		},
		{
			name:     "bunx",
			runner:   Runner{Name: "bunx", Cmd: "bunx", Args: []string{"mintlify"}},
			command:  []string{"version"},
			wantArgs: []string{"mintlify", "version"},
		},
		{
			name:     "pnpm_dlx",
			runner:   Runner{Name: "pnpm-dlx", Cmd: "pnpm", Args: []string{"dlx", "mintlify"}},
			command:  []string{"validate", "--strict"},
			wantArgs: []string{"dlx", "mintlify", "validate", "--strict"},
		},
		{
			name:     "npx",
			runner:   Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}},
			command:  []string{"broken-links"},
			wantArgs: []string{"mintlify", "broken-links"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rec recordingCmd
			c := newTestClient(t, &tt.runner, mockCmdRecording("ok", "", 0, &rec))

			_, _, _ = c.run(context.Background(), tt.command...)

			if rec.name != tt.runner.Cmd {
				t.Errorf("command = %q, want %q", rec.name, tt.runner.Cmd)
			}
			got := strings.Join(rec.args, " ")
			want := strings.Join(tt.wantArgs, " ")
			if got != want {
				t.Errorf("args = %q, want %q", got, want)
			}
		})
	}
}

// ── Command method tests ────────────────────────────────────────────

func TestVersion(t *testing.T) {
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd("4.2.33\n", "", 0))

	result, err := c.Version(context.Background())
	if err != nil {
		t.Fatalf("Version() error: %v", err)
	}
	if result.Version != "4.2.33" {
		t.Errorf("Version = %q, want %q", result.Version, "4.2.33")
	}
}

func TestValidate_Clean(t *testing.T) {
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd("All checks passed!\n", "", 0))

	result, err := c.Validate(context.Background(), ValidateOptions{})
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
	if !result.OK {
		t.Error("OK = false, want true")
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %d, want 0", len(result.Errors))
	}
}

func TestValidate_WithErrors(t *testing.T) {
	output := "Error: Invalid frontmatter at docs/page.mdx:5:1\nError: Missing title\n"
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd(output, "", 1))

	result, err := c.Validate(context.Background(), ValidateOptions{})
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
	if result.OK {
		t.Error("OK = true, want false")
	}
	if len(result.Errors) != 2 {
		t.Fatalf("Errors = %d, want 2", len(result.Errors))
	}
	if result.Errors[0].File != "docs/page.mdx" {
		t.Errorf("Errors[0].File = %q, want %q", result.Errors[0].File, "docs/page.mdx")
	}
	if result.Errors[0].Line != 5 {
		t.Errorf("Errors[0].Line = %d, want 5", result.Errors[0].Line)
	}
}

func TestValidate_Strict(t *testing.T) {
	var rec recordingCmd
	r := &Runner{Name: "mint", Cmd: "mint", Args: nil}
	c := newTestClient(t, r, mockCmdRecording("OK\n", "", 0, &rec))

	_, err := c.Validate(context.Background(), ValidateOptions{Strict: true})
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	got := strings.Join(rec.args, " ")
	if !strings.Contains(got, "--strict") {
		t.Errorf("args = %q, want --strict flag", got)
	}
}

func TestBrokenLinks(t *testing.T) {
	output := "✗ https://example.com/dead (404) in docs/links.mdx\n✗ /missing in docs/nav.mdx\n"
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd(output, "", 1))

	result, err := c.BrokenLinks(context.Background())
	if err != nil {
		t.Fatalf("BrokenLinks() error: %v", err)
	}
	if result.OK {
		t.Error("OK = true, want false")
	}
	if len(result.Links) != 2 {
		t.Fatalf("Links = %d, want 2", len(result.Links))
	}
	if result.Links[0].URL != "https://example.com/dead" {
		t.Errorf("Links[0].URL = %q, want %q", result.Links[0].URL, "https://example.com/dead")
	}
	if result.Links[0].Status != 404 {
		t.Errorf("Links[0].Status = %d, want 404", result.Links[0].Status)
	}
}

func TestA11y(t *testing.T) {
	output := "✗ Image missing alt text in docs/page.mdx\n"
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd(output, "", 1))

	result, err := c.A11y(context.Background())
	if err != nil {
		t.Fatalf("A11y() error: %v", err)
	}
	if result.OK {
		t.Error("OK = true, want false")
	}
	if len(result.Issues) != 1 {
		t.Fatalf("Issues = %d, want 1", len(result.Issues))
	}
}

func TestOpenAPICheck(t *testing.T) {
	var rec recordingCmd
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmdRecording("All checks passed\n", "", 0, &rec))

	result, err := c.OpenAPICheck(context.Background(), OpenAPICheckOptions{Target: "api.yaml"})
	if err != nil {
		t.Fatalf("OpenAPICheck() error: %v", err)
	}
	if !result.OK {
		t.Error("OK = false, want true")
	}
	got := strings.Join(rec.args, " ")
	if !strings.Contains(got, "api.yaml") {
		t.Errorf("args = %q, want api.yaml", got)
	}
}

func TestBuild_Success(t *testing.T) {
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd("Build completed successfully\n", "", 0))

	result, err := c.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if !result.OK {
		t.Error("OK = false, want true")
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %d, want 0", len(result.Errors))
	}
}

func TestBuild_WithErrors(t *testing.T) {
	output := "error: Could not compile at docs/broken.mdx:42\nerror: Undefined component\n"
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd(output, "", 1))

	result, err := c.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if result.OK {
		t.Error("OK = true, want false")
	}
	if len(result.Errors) < 1 {
		t.Fatal("expected at least 1 build error")
	}
}

func TestMigrateMDX(t *testing.T) {
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd("Migration complete.\n", "", 0))

	out, err := c.MigrateMDX(context.Background())
	if err != nil {
		t.Fatalf("MigrateMDX() error: %v", err)
	}
	if out != "Migration complete." {
		t.Errorf("output = %q, want %q", out, "Migration complete.")
	}
}

func TestScrape(t *testing.T) {
	var rec recordingCmd
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmdRecording("Scraped 10 pages\n", "", 0, &rec))

	out, err := c.Scrape(context.Background(), ScrapeOptions{
		Mode:   ScrapeSitemap,
		Target: "https://example.com/sitemap.xml",
	})
	if err != nil {
		t.Fatalf("Scrape() error: %v", err)
	}
	if out != "Scraped 10 pages" {
		t.Errorf("output = %q, want %q", out, "Scraped 10 pages")
	}
}

func TestNewProject(t *testing.T) {
	var rec recordingCmd
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmdRecording("Created project\n", "", 0, &rec))

	out, err := c.NewProject(context.Background(), NewProjectOptions{Directory: "my-docs"})
	if err != nil {
		t.Fatalf("NewProject() error: %v", err)
	}
	if out != "Created project" {
		t.Errorf("output = %q, want %q", out, "Created project")
	}
	got := strings.Join(rec.args, " ")
	if !strings.Contains(got, "my-docs") {
		t.Errorf("args = %q, want my-docs", got)
	}
}

// ── Error handling ──────────────────────────────────────────────────

func TestRun_NonZeroExitError(t *testing.T) {
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	c := newTestClient(t, r, mockCmd("", "something failed", 1))

	_, err := c.Version(context.Background())
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
	if !strings.Contains(err.Error(), "exit code 1") {
		t.Errorf("error = %q, want exit code 1 mention", err.Error())
	}
}

func TestRun_CommandError(t *testing.T) {
	r := &Runner{Name: "npx", Cmd: "npx", Args: []string{"mintlify"}}
	cmd := func(_ context.Context, _ string, _ string, _ ...string) (string, string, int, error) {
		return "", "", -1, fmt.Errorf("exec: not found")
	}
	c := newTestClient(t, r, cmd)

	_, err := c.Version(context.Background())
	if err == nil {
		t.Fatal("expected error for exec failure")
	}
	if !strings.Contains(err.Error(), "exec: not found") {
		t.Errorf("error = %q, want exec: not found mention", err.Error())
	}
}

func TestWorkingDir(t *testing.T) {
	var rec recordingCmd
	r := &Runner{Name: "mint", Cmd: "mint", Args: nil}
	c := newTestClient(t, r, mockCmdRecording("", "", 0, &rec))

	_, _, _ = c.run(context.Background(), "version")

	if rec.dir != "/test/dir" {
		t.Errorf("dir = %q, want %q", rec.dir, "/test/dir")
	}
}
