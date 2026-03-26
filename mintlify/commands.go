package mintlify

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// combineOutput merges stdout and stderr into a single string.
func combineOutput(stdout, stderr string) string {
	if stderr == "" {
		return stdout
	}
	return stdout + "\n" + stderr
}

// Version runs `mintlify version` and returns the parsed result.
func (c *Client) Version(ctx context.Context) (*VersionResult, error) {
	stdout, _, err := c.run(ctx, "version")
	if err != nil {
		return nil, err
	}
	return parseVersion(stdout), nil
}

// Validate runs `mintlify validate` and returns parsed validation issues.
// A non-zero exit code is not treated as an error — it means issues were found.
func (c *Client) Validate(ctx context.Context, opts ValidateOptions) (*ValidateResult, error) {
	args := []string{"validate"}
	if opts.Strict {
		args = append(args, "--strict")
	}
	stdout, stderr, exitCode, err := c.runRaw(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("mintlify validate: %w", err)
	}

	result := parseValidate(combineOutput(stdout, stderr))
	result.OK = exitCode == 0
	return result, nil
}

// BrokenLinks runs `mintlify broken-links` and returns parsed results.
// A non-zero exit code means broken links were found (not an error).
func (c *Client) BrokenLinks(ctx context.Context) (*BrokenLinksResult, error) {
	stdout, stderr, exitCode, err := c.runRaw(ctx, "broken-links")
	if err != nil {
		return nil, fmt.Errorf("mintlify broken-links: %w", err)
	}

	result := parseBrokenLinks(combineOutput(stdout, stderr))
	result.OK = exitCode == 0
	return result, nil
}

// A11y runs `mintlify a11y` and returns parsed accessibility issues.
// A non-zero exit code means issues were found (not an error).
func (c *Client) A11y(ctx context.Context) (*A11yResult, error) {
	stdout, stderr, exitCode, err := c.runRaw(ctx, "a11y")
	if err != nil {
		return nil, fmt.Errorf("mintlify a11y: %w", err)
	}

	result := parseA11y(combineOutput(stdout, stderr))
	result.OK = exitCode == 0
	return result, nil
}

// OpenAPICheck runs `mintlify openapi-check` and returns parsed results.
func (c *Client) OpenAPICheck(ctx context.Context, opts OpenAPICheckOptions) (*OpenAPICheckResult, error) {
	args := []string{"openapi-check"}
	if opts.Target != "" {
		args = append(args, opts.Target)
	}
	stdout, stderr, exitCode, err := c.runRaw(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("mintlify openapi-check: %w", err)
	}

	raw := strings.TrimSpace(combineOutput(stdout, stderr))
	issues := ""
	if exitCode != 0 {
		issues = raw
	}

	return &OpenAPICheckResult{
		OK:     exitCode == 0,
		Issues: issues,
		Raw:    raw,
	}, nil
}

// Build runs `mintlify build` and returns a structured result.
// A non-zero exit code means the build failed; errors are parsed from output.
func (c *Client) Build(ctx context.Context, _ BuildOptions) (*BuildResult, error) {
	start := time.Now()
	stdout, stderr, exitCode, err := c.runRaw(ctx, "build")
	elapsed := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("mintlify build: %w", err)
	}

	result := parseBuild(combineOutput(stdout, stderr))
	result.OK = exitCode == 0
	result.Duration = elapsed
	return result, nil
}

// MigrateMDX runs `mintlify migrate-mdx` and returns the raw output.
func (c *Client) MigrateMDX(ctx context.Context) (string, error) {
	stdout, _, err := c.run(ctx, "migrate-mdx")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// Scrape runs `mintlify scrape <mode> [target]` and returns the raw output.
func (c *Client) Scrape(ctx context.Context, opts ScrapeOptions) (string, error) {
	args := []string{"scrape"}
	if opts.Mode != "" {
		args = append(args, string(opts.Mode))
	}
	if opts.Target != "" {
		args = append(args, opts.Target)
	}
	stdout, _, err := c.run(ctx, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// NewProject runs `mintlify new [dir]` and returns the raw output.
func (c *Client) NewProject(ctx context.Context, opts NewProjectOptions) (string, error) {
	args := []string{"new"}
	if opts.Directory != "" {
		args = append(args, opts.Directory)
	}
	stdout, _, err := c.run(ctx, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// Rename runs `mintlify rename` and returns the raw output.
func (c *Client) Rename(ctx context.Context) (string, error) {
	stdout, _, err := c.run(ctx, "rename")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// Upgrade runs `mintlify upgrade` and returns the raw output.
func (c *Client) Upgrade(ctx context.Context) (string, error) {
	stdout, _, err := c.run(ctx, "upgrade")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}
