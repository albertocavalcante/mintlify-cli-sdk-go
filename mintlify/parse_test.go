package mintlify

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"4.2.33", "4.2.33"},
		{"mintlify v4.2.33", "4.2.33"},
		{"v1.0.0-beta", "1.0.0"},
		{"no version here", ""},
	}
	for _, tt := range tests {
		result := parseVersion(tt.input)
		if result.Version != tt.want {
			t.Errorf("parseVersion(%q).Version = %q, want %q", tt.input, result.Version, tt.want)
		}
	}
}

func TestParseValidate(t *testing.T) {
	input := "Error: Invalid frontmatter at docs/page.mdx:5:1\nError: Missing title\n"
	result := parseValidate(input)

	if len(result.Errors) != 2 {
		t.Fatalf("Errors = %d, want 2", len(result.Errors))
	}
	if result.Errors[0].Message != "Invalid frontmatter" {
		t.Errorf("Errors[0].Message = %q, want %q", result.Errors[0].Message, "Invalid frontmatter")
	}
	if result.Errors[0].File != "docs/page.mdx" {
		t.Errorf("Errors[0].File = %q, want %q", result.Errors[0].File, "docs/page.mdx")
	}
	if result.Errors[0].Line != 5 {
		t.Errorf("Errors[0].Line = %d, want 5", result.Errors[0].Line)
	}
	if result.Errors[0].Column != 1 {
		t.Errorf("Errors[0].Column = %d, want 1", result.Errors[0].Column)
	}
	if result.Errors[1].File != "" {
		t.Errorf("Errors[1].File = %q, want empty", result.Errors[1].File)
	}
}

func TestParseValidate_Empty(t *testing.T) {
	result := parseValidate("All checks passed!")
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %d, want 0", len(result.Errors))
	}
}

func TestParseBrokenLinks(t *testing.T) {
	input := "✗ https://example.com/dead (404) in docs/links.mdx\n✗ /missing in docs/nav.mdx\n"
	result := parseBrokenLinks(input)

	if len(result.Links) != 2 {
		t.Fatalf("Links = %d, want 2", len(result.Links))
	}
	if result.Links[0].URL != "https://example.com/dead" {
		t.Errorf("Links[0].URL = %q", result.Links[0].URL)
	}
	if result.Links[0].Status != 404 {
		t.Errorf("Links[0].Status = %d, want 404", result.Links[0].Status)
	}
	if result.Links[0].Source != "docs/links.mdx" {
		t.Errorf("Links[0].Source = %q", result.Links[0].Source)
	}
	if result.Links[1].Status != 0 {
		t.Errorf("Links[1].Status = %d, want 0", result.Links[1].Status)
	}
}

func TestParseA11y(t *testing.T) {
	input := "✗ Image missing alt text in docs/page.mdx\n"
	result := parseA11y(input)

	if len(result.Issues) != 1 {
		t.Fatalf("Issues = %d, want 1", len(result.Issues))
	}
	if result.Issues[0].Message != "Image missing alt text" {
		t.Errorf("Message = %q", result.Issues[0].Message)
	}
	if result.Issues[0].File != "docs/page.mdx" {
		t.Errorf("File = %q", result.Issues[0].File)
	}
}

func TestParseBuild_Success(t *testing.T) {
	result := parseBuild("Build completed successfully")
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %d, want 0", len(result.Errors))
	}
}

func TestParseBuild_WithErrors(t *testing.T) {
	input := "error: Could not compile at docs/broken.mdx:42\nerror: Undefined component\n"
	result := parseBuild(input)

	if len(result.Errors) < 1 {
		t.Fatal("expected at least 1 error")
	}
	if result.Errors[0].File != "docs/broken.mdx" {
		t.Errorf("Errors[0].File = %q, want docs/broken.mdx", result.Errors[0].File)
	}
	if result.Errors[0].Line != 42 {
		t.Errorf("Errors[0].Line = %d, want 42", result.Errors[0].Line)
	}
}

func TestParseBuild_Empty(t *testing.T) {
	result := parseBuild("")
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %d, want 0", len(result.Errors))
	}
}
