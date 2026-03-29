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

// ── Export ─────────────────────────────────────────────────────────

// ExportFormat is the output format for export.
type ExportFormat string

const (
	ExportPDF  ExportFormat = "pdf"
	ExportHTML ExportFormat = "html"
)

// ExportOptions controls the export command.
type ExportOptions struct {
	Format    ExportFormat // output format (default: pdf)
	OutputDir string       // output directory
	Pages     []string     // specific pages to export (empty = all)
}

// ExportResult holds the output of `mintlify export`.
type ExportResult struct {
	OutputPath string // path to exported file/directory
	Raw        string // full unparsed output
}

// ── Workflow ───────────────────────────────────────────────────────

// WorkflowAction is the workflow subcommand.
type WorkflowAction string

const (
	WorkflowCreate WorkflowAction = "create"
	WorkflowList   WorkflowAction = "list"
	WorkflowRun    WorkflowAction = "run"
)

// WorkflowOptions controls the workflow command.
type WorkflowOptions struct {
	Action       WorkflowAction
	Name         string // workflow name (for create)
	TriggerType  string // "cron", "push", "manual"
	CronExpr     string // cron expression (for cron trigger)
	Instructions string // workflow instructions
	Dir          string // output directory
}

// WorkflowResult holds the output of `mintlify workflow`.
type WorkflowResult struct {
	FilePath string // created workflow file path
	Raw      string // full unparsed output
}

// ── Skills ─────────────────────────────────────────────────────────

// SkillsAddOptions controls the skills add command.
type SkillsAddOptions struct {
	Name string // skill name to install
}

// SkillsResult holds the output of `mintlify skills add`.
type SkillsResult struct {
	Name string // installed skill name
	Path string // path to skill file
	Raw  string // full unparsed output
}

// ── Search ─────────────────────────────────────────────────────────

// SearchOptions controls a search query against the Mintlify Search API.
type SearchOptions struct {
	Query    string // search query text
	Limit    int    // max results (0 = API default)
	Language string // filter by language code
}

// SearchResult holds the response from the Mintlify Search API.
type SearchResult struct {
	Results []SearchHit // matched documents
	Total   int         // total number of matches
}

// SearchHit represents a single search result.
type SearchHit struct {
	Title   string  // page title
	Path    string  // page path
	Snippet string  // text snippet with match context
	Score   float64 // relevance score
}

// ── LLMs.txt ──────────────────────────────────────────────────────

// LLMsTxtOptions controls llms.txt generation.
type LLMsTxtOptions struct {
	MaxDescLen  int    // max description length (default: 300)
	IncludeAPIs bool   // include OpenAPI/AsyncAPI spec links
	OutputDir   string // output directory (default: docs root)
}

// LLMsTxtResult holds the result of llms.txt generation.
type LLMsTxtResult struct {
	IndexPath string // path to generated llms.txt
	FullPath  string // path to generated llms-full.txt
	PageCount int    // number of pages included
}

// LLMsTxtIssue represents a validation issue in an existing llms.txt file.
type LLMsTxtIssue struct {
	Message string // description of the issue
	Page    string // page path (if applicable)
}

// ── Config ────────────────────────────────────────────────────────

// DocsConfig represents a parsed Mintlify docs.json or mint.json configuration.
type DocsConfig struct {
	Name       string     `json:"name,omitempty"`
	Theme      string     `json:"theme,omitempty"`
	Favicon    string     `json:"favicon,omitempty"`
	Colors     *Colors    `json:"colors,omitempty"`
	Logo       *Logo      `json:"logo,omitempty"`
	Navigation []NavGroup `json:"navigation,omitempty"`
	Tabs       []Tab      `json:"tabs,omitempty"`
	Anchors    []Anchor   `json:"anchors,omitempty"`
	Feedback   *Feedback  `json:"feedback,omitempty"`
	OpenAPI    any        `json:"openapi,omitempty"`    // string or []string
	API        *API       `json:"api,omitempty"`
	ConfigFile string     `json:"-"` // "docs.json" or "mint.json" (not serialized)
}

// Colors holds the Mintlify color configuration.
type Colors struct {
	Primary    string `json:"primary,omitempty"`
	Light      string `json:"light,omitempty"`
	Dark       string `json:"dark,omitempty"`
	Background *struct {
		Light string `json:"light,omitempty"`
		Dark  string `json:"dark,omitempty"`
	} `json:"background,omitempty"`
}

// Logo holds logo configuration.
type Logo struct {
	Light string `json:"light,omitempty"`
	Dark  string `json:"dark,omitempty"`
	Href  string `json:"href,omitempty"`
}

// NavGroup represents a navigation group in the sidebar.
type NavGroup struct {
	Group string `json:"group"`
	Pages []any  `json:"pages"` // string or nested NavGroup
}

// Tab represents a top-level navigation tab.
type Tab struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Anchor represents a sidebar anchor link.
type Anchor struct {
	Name string `json:"name"`
	Icon string `json:"icon,omitempty"`
	URL  string `json:"url"`
}

// Feedback controls the feedback widget.
type Feedback struct {
	ThumbsRating bool `json:"thumbsRating,omitempty"`
}

// API holds API playground configuration.
type API struct {
	BaseURL    string `json:"baseUrl,omitempty"`
	Auth       *Auth  `json:"auth,omitempty"`
	Playground string `json:"playground,omitempty"` // "interactive", "simple", "none", "auth"
}

// Auth holds API authentication configuration.
type Auth struct {
	Method string `json:"method,omitempty"` // "bearer", "basic", "key"
	Name   string `json:"name,omitempty"`
}

// NavigationPages returns a flat list of all page paths from the navigation tree.
func (c *DocsConfig) NavigationPages() []string {
	var pages []string
	for _, group := range c.Navigation {
		pages = append(pages, extractPages(group.Pages)...)
	}
	return pages
}
