package mintlify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectConfigFile_DocsJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "docs.json"), []byte(`{"name":"test"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	name, err := DetectConfigFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "docs.json" {
		t.Errorf("got %q, want %q", name, "docs.json")
	}
}

func TestDetectConfigFile_MintJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mint.json"), []byte(`{"name":"test"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	name, err := DetectConfigFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "mint.json" {
		t.Errorf("got %q, want %q", name, "mint.json")
	}
}

func TestDetectConfigFile_PrefersDocsJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "docs.json"), []byte(`{"name":"new"}`), 0o644)
	os.WriteFile(filepath.Join(dir, "mint.json"), []byte(`{"name":"old"}`), 0o644)

	name, err := DetectConfigFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "docs.json" {
		t.Errorf("got %q, want %q", name, "docs.json")
	}
}

func TestDetectConfigFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := DetectConfigFile(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadDocsConfig(t *testing.T) {
	dir := t.TempDir()
	config := `{
		"name": "My Docs",
		"theme": "maple",
		"navigation": [
			{
				"group": "Getting Started",
				"pages": ["intro", "quickstart"]
			}
		]
	}`
	os.WriteFile(filepath.Join(dir, "mint.json"), []byte(config), 0o644)

	cfg, err := LoadDocsConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "My Docs" {
		t.Errorf("name: got %q, want %q", cfg.Name, "My Docs")
	}
	if cfg.Theme != "maple" {
		t.Errorf("theme: got %q, want %q", cfg.Theme, "maple")
	}
	if cfg.ConfigFile != "mint.json" {
		t.Errorf("configFile: got %q, want %q", cfg.ConfigFile, "mint.json")
	}

	pages := cfg.NavigationPages()
	if len(pages) != 2 {
		t.Fatalf("pages: got %d, want 2", len(pages))
	}
	if pages[0] != "intro" || pages[1] != "quickstart" {
		t.Errorf("pages: got %v, want [intro quickstart]", pages)
	}
}
