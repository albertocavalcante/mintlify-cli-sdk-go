package mintlify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateLLMsTxt(t *testing.T) {
	dir := t.TempDir()

	// Write config
	config := `{
		"name": "Bazel Docs",
		"navigation": [
			{
				"group": "Getting Started",
				"pages": ["intro", "install"]
			},
			{
				"group": "Reference",
				"pages": ["api"]
			}
		]
	}`
	os.WriteFile(filepath.Join(dir, "docs.json"), []byte(config), 0o644)

	// Write page files
	os.WriteFile(filepath.Join(dir, "intro.mdx"), []byte(`---
title: Introduction
description: Learn how to get started with Bazel
---

Welcome to Bazel!
`), 0o644)

	os.WriteFile(filepath.Join(dir, "install.mdx"), []byte(`---
title: Installation
description: Install Bazel on your platform
---

Follow these steps to install.
`), 0o644)

	os.WriteFile(filepath.Join(dir, "api.mdx"), []byte(`---
title: API Reference
---

The API reference.
`), 0o644)

	result, err := GenerateLLMsTxt(dir, LLMsTxtOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PageCount != 3 {
		t.Errorf("page count: got %d, want 3", result.PageCount)
	}

	// Read generated llms.txt
	index, err := os.ReadFile(result.IndexPath)
	if err != nil {
		t.Fatalf("reading llms.txt: %v", err)
	}
	indexStr := string(index)

	if !strings.Contains(indexStr, "# Bazel Docs") {
		t.Error("llms.txt missing site name header")
	}
	if !strings.Contains(indexStr, "[Introduction](/intro)") {
		t.Error("llms.txt missing intro entry")
	}
	if !strings.Contains(indexStr, "Learn how to get started") {
		t.Error("llms.txt missing intro description")
	}
	if !strings.Contains(indexStr, "[API Reference](/api)") {
		t.Error("llms.txt missing api entry")
	}

	// Read generated llms-full.txt
	full, err := os.ReadFile(result.FullPath)
	if err != nil {
		t.Fatalf("reading llms-full.txt: %v", err)
	}
	fullStr := string(full)

	if !strings.Contains(fullStr, "Welcome to Bazel!") {
		t.Error("llms-full.txt missing intro content")
	}
	if !strings.Contains(fullStr, "The API reference.") {
		t.Error("llms-full.txt missing api content")
	}
}

func TestValidateLLMsTxt_Missing(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "docs.json"), []byte(`{"navigation":[]}`), 0o644)

	issues, err := ValidateLLMsTxt(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 || issues[0].Message != "llms.txt not found" {
		t.Errorf("expected 'not found' issue, got %v", issues)
	}
}

func TestValidateLLMsTxt_StaleEntry(t *testing.T) {
	dir := t.TempDir()

	config := `{"navigation":[{"group":"Docs","pages":["intro"]}]}`
	os.WriteFile(filepath.Join(dir, "docs.json"), []byte(config), 0o644)

	// llms.txt has an extra page not in navigation
	llmsTxt := "# Docs\n- [Intro](/intro)\n- [Removed Page](/removed)\n"
	os.WriteFile(filepath.Join(dir, "llms.txt"), []byte(llmsTxt), 0o644)

	issues, err := ValidateLLMsTxt(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundStale := false
	for _, issue := range issues {
		if issue.Page == "removed" && strings.Contains(issue.Message, "stale") {
			foundStale = true
		}
	}
	if !foundStale {
		t.Errorf("expected stale entry issue, got %v", issues)
	}
}
