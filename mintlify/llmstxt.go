package mintlify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var frontmatterRe = regexp.MustCompile(`(?s)^---\n(.+?)\n---`)
var titleRe = regexp.MustCompile(`(?m)^title:\s*["']?(.+?)["']?\s*$`)
var descriptionRe = regexp.MustCompile(`(?m)^description:\s*["']?(.+?)["']?\s*$`)

// GenerateLLMsTxt generates llms.txt and llms-full.txt files from the docs
// directory. It reads the navigation tree from docs.json/mint.json, extracts
// frontmatter (title, description) from each page, and produces:
//   - llms.txt: index with page titles and truncated descriptions
//   - llms-full.txt: full content of all pages concatenated
func GenerateLLMsTxt(docsDir string, opts LLMsTxtOptions) (*LLMsTxtResult, error) {
	if opts.MaxDescLen <= 0 {
		opts.MaxDescLen = 300
	}
	if opts.OutputDir == "" {
		opts.OutputDir = docsDir
	}

	cfg, err := LoadDocsConfig(docsDir)
	if err != nil {
		return nil, fmt.Errorf("loading config for llms.txt: %w", err)
	}

	pages := cfg.NavigationPages()
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found in navigation")
	}

	var indexLines []string
	var fullContent []string
	pageCount := 0

	// Header
	siteName := cfg.Name
	if siteName == "" {
		siteName = "Documentation"
	}
	indexLines = append(indexLines, fmt.Sprintf("# %s", siteName))
	indexLines = append(indexLines, "")

	// Group pages by navigation group
	for _, group := range cfg.Navigation {
		groupPages := extractPages(group.Pages)
		if len(groupPages) == 0 {
			continue
		}

		indexLines = append(indexLines, fmt.Sprintf("## %s", group.Group))

		for _, page := range groupPages {
			title, desc, content, err := readPageMeta(docsDir, page)
			if err != nil {
				continue // skip unreadable pages
			}

			// Truncate description
			if len(desc) > opts.MaxDescLen {
				desc = desc[:opts.MaxDescLen-3] + "..."
			}

			if title == "" {
				title = page
			}

			entry := fmt.Sprintf("- [%s](/%s)", title, page)
			if desc != "" {
				entry += fmt.Sprintf(": %s", desc)
			}
			indexLines = append(indexLines, entry)

			// Full content
			fullContent = append(fullContent, fmt.Sprintf("# %s\n\nSource: /%s\n\n%s", title, page, content))

			pageCount++
		}

		indexLines = append(indexLines, "")
	}

	// Add API spec links if requested
	if opts.IncludeAPIs && cfg.OpenAPI != nil {
		indexLines = append(indexLines, "## API Specifications")
		switch v := cfg.OpenAPI.(type) {
		case string:
			indexLines = append(indexLines, fmt.Sprintf("- [OpenAPI Spec](%s)", v))
		case []any:
			for _, spec := range v {
				if s, ok := spec.(string); ok {
					indexLines = append(indexLines, fmt.Sprintf("- [OpenAPI Spec](%s)", s))
				}
			}
		}
		indexLines = append(indexLines, "")
	}

	// Write files
	indexPath := filepath.Join(opts.OutputDir, "llms.txt")
	if err := os.WriteFile(indexPath, []byte(strings.Join(indexLines, "\n")+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("writing llms.txt: %w", err)
	}

	fullPath := filepath.Join(opts.OutputDir, "llms-full.txt")
	fullSep := "\n\n---\n\n"
	if err := os.WriteFile(fullPath, []byte(strings.Join(fullContent, fullSep)+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("writing llms-full.txt: %w", err)
	}

	return &LLMsTxtResult{
		IndexPath: indexPath,
		FullPath:  fullPath,
		PageCount: pageCount,
	}, nil
}

// ValidateLLMsTxt checks that an existing llms.txt is in sync with the
// docs navigation. Returns a list of issues found.
func ValidateLLMsTxt(docsDir string) ([]LLMsTxtIssue, error) {
	cfg, err := LoadDocsConfig(docsDir)
	if err != nil {
		return nil, fmt.Errorf("loading config for validation: %w", err)
	}

	indexPath := filepath.Join(docsDir, "llms.txt")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return []LLMsTxtIssue{{Message: "llms.txt not found"}}, nil
	}

	indexStr := string(content)
	pages := cfg.NavigationPages()
	var issues []LLMsTxtIssue

	for _, page := range pages {
		if !strings.Contains(indexStr, "/"+page) {
			issues = append(issues, LLMsTxtIssue{
				Message: fmt.Sprintf("page missing from llms.txt"),
				Page:    page,
			})
		}
	}

	// Check for stale entries: pages in llms.txt that are no longer in navigation
	// (heuristic: look for markdown links)
	linkRe := regexp.MustCompile(`\[.+?\]\((/[^)]+)\)`)
	for _, match := range linkRe.FindAllStringSubmatch(indexStr, -1) {
		linkedPage := strings.TrimPrefix(match[1], "/")
		found := false
		for _, page := range pages {
			if page == linkedPage {
				found = true
				break
			}
		}
		if !found {
			issues = append(issues, LLMsTxtIssue{
				Message: fmt.Sprintf("stale entry in llms.txt (not in navigation)"),
				Page:    linkedPage,
			})
		}
	}

	return issues, nil
}

// readPageMeta reads an MDX file and extracts its frontmatter title,
// description, and full content (without frontmatter).
func readPageMeta(docsDir, page string) (title, description, content string, err error) {
	// Try .mdx first, then .md
	var data []byte
	for _, ext := range []string{".mdx", ".md"} {
		path := filepath.Join(docsDir, page+ext)
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", "", "", err
	}

	raw := string(data)
	content = raw

	// Extract frontmatter
	if m := frontmatterRe.FindStringSubmatch(raw); m != nil {
		fm := m[1]
		content = strings.TrimSpace(raw[len(m[0]):])

		if tm := titleRe.FindStringSubmatch(fm); tm != nil {
			title = tm[1]
		}
		if dm := descriptionRe.FindStringSubmatch(fm); dm != nil {
			description = dm[1]
		}
	}

	return title, description, content, nil
}
