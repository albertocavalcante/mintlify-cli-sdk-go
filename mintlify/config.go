package mintlify

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// configFileNames lists supported config filenames in priority order.
// docs.json is the newer format; mint.json is legacy.
var configFileNames = []string{"docs.json", "mint.json"}

// ErrNoConfig is returned when neither docs.json nor mint.json is found.
var ErrNoConfig = errors.New("mintlify: no docs.json or mint.json found")

// DetectConfigFile finds the Mintlify config file in the given directory.
// It checks for docs.json first (newer format), then mint.json.
// Returns the filename (not full path) and nil error, or ErrNoConfig.
func DetectConfigFile(dir string) (string, error) {
	for _, name := range configFileNames {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("%w in %s", ErrNoConfig, dir)
}

// LoadDocsConfig parses the Mintlify configuration from the given directory.
// It auto-detects docs.json vs mint.json, preferring docs.json.
func LoadDocsConfig(dir string) (*DocsConfig, error) {
	name, err := DetectConfigFile(dir)
	if err != nil {
		return nil, err
	}
	return loadConfigFile(filepath.Join(dir, name), name)
}

// LoadDocsConfigFrom parses a specific configuration file.
func LoadDocsConfigFrom(path string) (*DocsConfig, error) {
	return loadConfigFile(path, filepath.Base(path))
}

func loadConfigFile(path, name string) (*DocsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg DocsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	cfg.ConfigFile = name
	return &cfg, nil
}

// extractPages recursively extracts page paths from a navigation pages list.
// Pages can be strings (page paths) or nested NavGroup objects.
func extractPages(pages []any) []string {
	var result []string
	for _, p := range pages {
		switch v := p.(type) {
		case string:
			result = append(result, v)
		case map[string]any:
			if ps, ok := v["pages"]; ok {
				if arr, ok := ps.([]any); ok {
					result = append(result, extractPages(arr)...)
				}
			}
		}
	}
	return result
}
