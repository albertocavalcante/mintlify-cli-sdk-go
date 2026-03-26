package mintlify

import (
	"regexp"
	"strconv"
	"strings"
)

// ── Regexes ─────────────────────────────────────────────────────────

var (
	// versionRe matches output like "4.2.33" or "mintlify v4.2.33".
	versionRe = regexp.MustCompile(`(?:^|v)(\d+\.\d+\.\d+)`)

	// validateErrorRe matches lines like:
	//   Error: some message at path/to/file.mdx:10:5
	//   Error: some message
	validateErrorRe = regexp.MustCompile(`(?m)^Error:\s+(.+?)(?:\s+at\s+(.+?):(\d+):(\d+))?$`)

	// brokenLinkRe matches lines like:
	//   ✗ https://example.com (404) in docs/page.mdx
	//   ✗ /broken/path in docs/page.mdx
	brokenLinkRe = regexp.MustCompile(`(?m)^[✗✘×xX]\s+(\S+)(?:\s+\((\d+)\))?\s+in\s+(.+)$`)

	// a11yIssueRe matches lines like:
	//   ✗ Image missing alt text in docs/page.mdx
	a11yIssueRe = regexp.MustCompile(`(?m)^[✗✘×xX]\s+(.+?)\s+in\s+(\S+)$`)

	// buildErrorRe matches build error lines like:
	//   Error: Could not compile docs/page.mdx
	//   Error: Syntax error at docs/page.mdx:10
	buildErrorRe = regexp.MustCompile(`(?m)^(?:Error|error):\s+(.+?)(?:\s+(?:at|in)\s+(.+?)(?::(\d+))?)?$`)
)

// ── Parsers ─────────────────────────────────────────────────────────

// parseVersion extracts a semver version from CLI output.
func parseVersion(raw string) *VersionResult {
	raw = strings.TrimSpace(raw)
	result := &VersionResult{Raw: raw}
	if m := versionRe.FindStringSubmatch(raw); m != nil {
		result.Version = m[1]
	}
	return result
}

// parseValidate extracts structured validation errors from CLI output.
func parseValidate(raw string) *ValidateResult {
	raw = strings.TrimSpace(raw)
	result := &ValidateResult{Raw: raw}

	for _, match := range validateErrorRe.FindAllStringSubmatch(raw, -1) {
		issue := ValidationIssue{
			Message: match[1],
			File:    match[2],
		}
		if match[3] != "" {
			issue.Line, _ = strconv.Atoi(match[3])
		}
		if match[4] != "" {
			issue.Column, _ = strconv.Atoi(match[4])
		}
		result.Errors = append(result.Errors, issue)
	}

	return result
}

// parseBrokenLinks extracts broken link entries from CLI output.
func parseBrokenLinks(raw string) *BrokenLinksResult {
	raw = strings.TrimSpace(raw)
	result := &BrokenLinksResult{Raw: raw}

	for _, match := range brokenLinkRe.FindAllStringSubmatch(raw, -1) {
		link := BrokenLink{
			URL:    match[1],
			Source: match[3],
		}
		if match[2] != "" {
			link.Status, _ = strconv.Atoi(match[2])
		}
		result.Links = append(result.Links, link)
	}

	return result
}

// parseA11y extracts accessibility issues from CLI output.
func parseA11y(raw string) *A11yResult {
	raw = strings.TrimSpace(raw)
	result := &A11yResult{Raw: raw}

	for _, match := range a11yIssueRe.FindAllStringSubmatch(raw, -1) {
		result.Issues = append(result.Issues, A11yIssue{
			Message: match[1],
			File:    match[2],
		})
	}

	return result
}

// parseBuild extracts structured build errors from CLI output.
func parseBuild(raw string) *BuildResult {
	raw = strings.TrimSpace(raw)
	result := &BuildResult{Raw: raw}

	for _, match := range buildErrorRe.FindAllStringSubmatch(raw, -1) {
		be := BuildError{
			Message: match[1],
		}
		if match[2] != "" {
			be.File = match[2]
		}
		if match[3] != "" {
			be.Line, _ = strconv.Atoi(match[3])
		}
		result.Errors = append(result.Errors, be)
	}

	return result
}
