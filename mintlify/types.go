package mintlify

import "time"

// ── Result types ────────────────────────────────────────────────────

// VersionResult holds the parsed output of `mintlify version`.
type VersionResult struct {
	Version string // semver string, e.g. "4.2.33"
	Raw     string // full unparsed output
}

// ValidationIssue represents a single error from `mintlify validate`.
type ValidationIssue struct {
	Message string
	File    string
	Line    int
	Column  int
}

// ValidateResult holds the parsed output of `mintlify validate`.
type ValidateResult struct {
	OK     bool              // true when no issues found
	Errors []ValidationIssue // parsed issues
	Raw    string            // full unparsed output
}

// BrokenLink represents a single broken link from `mintlify broken-links`.
type BrokenLink struct {
	URL    string
	Source string
	Status int
}

// BrokenLinksResult holds the parsed output of `mintlify broken-links`.
type BrokenLinksResult struct {
	OK    bool         // true when no broken links found
	Links []BrokenLink // parsed broken links
	Raw   string       // full unparsed output
}

// A11yIssue represents an accessibility issue.
type A11yIssue struct {
	Message string
	File    string
}

// A11yResult holds the parsed output of `mintlify a11y`.
type A11yResult struct {
	OK     bool        // true when no issues found
	Issues []A11yIssue // parsed issues
	Raw    string      // full unparsed output
}

// OpenAPICheckResult holds the parsed output of `mintlify openapi-check`.
type OpenAPICheckResult struct {
	OK     bool   // true when check passed
	Issues string // raw issue text if any
	Raw    string // full unparsed output
}

// BuildError represents a single error from `mintlify build`.
type BuildError struct {
	Message string
	File    string
	Line    int
}

// BuildResult holds the parsed output of `mintlify build`.
type BuildResult struct {
	OK       bool          // true when build succeeded
	Errors   []BuildError  // parsed build errors
	Duration time.Duration // build duration
	Raw      string        // full unparsed output
}

// ── Option types ────────────────────────────────────────────────────

// ValidateOptions controls the validate command.
type ValidateOptions struct {
	Strict bool // pass --strict flag
}

// OpenAPICheckOptions controls the openapi-check command.
type OpenAPICheckOptions struct {
	Target string // file or URL to check
}

// DevOptions controls the dev server command.
type DevOptions struct {
	Port int // --port flag (0 = default 3000)
}

// BuildOptions controls the build command.
type BuildOptions struct {
	// Currently no flags; reserved for future use.
}

// ScrapeMode is the scraping mode.
type ScrapeMode string

const (
	ScrapeSitemap ScrapeMode = "sitemap"
	ScrapeURLs    ScrapeMode = "urls"
)

// ScrapeOptions controls the scrape command.
type ScrapeOptions struct {
	Mode   ScrapeMode
	Target string // sitemap URL or file path
}

// NewProjectOptions controls the new command.
type NewProjectOptions struct {
	Directory string // target directory
}
