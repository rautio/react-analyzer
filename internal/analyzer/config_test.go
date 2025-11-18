package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPathAliases_TSConfig(t *testing.T) {
	// Create temporary directory with tsconfig.json
	tmpDir := t.TempDir()

	tsconfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"],
      "~/*": ["./*"]
    }
  }
}`

	tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
	if err := os.WriteFile(tsconfigPath, []byte(tsconfig), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig.json: %v", err)
	}

	// Load aliases
	aliases, err := LoadPathAliases(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load aliases: %v", err)
	}

	// Verify aliases
	expected := map[string]string{
		"@/":           filepath.Join(tmpDir, "src") + string(filepath.Separator),
		"@components/": filepath.Join(tmpDir, "src", "components") + string(filepath.Separator),
		"~/":           tmpDir + string(filepath.Separator),
	}

	if len(aliases) != len(expected) {
		t.Errorf("Expected %d aliases, got %d", len(expected), len(aliases))
	}

	for prefix, expectedTarget := range expected {
		actualTarget, ok := aliases[prefix]
		if !ok {
			t.Errorf("Expected alias %s not found", prefix)
			continue
		}

		// Clean both paths for comparison
		expectedTarget = filepath.Clean(expectedTarget)
		actualTarget = filepath.Clean(actualTarget)

		if actualTarget != expectedTarget {
			t.Errorf("Alias %s: expected %s, got %s", prefix, expectedTarget, actualTarget)
		}
	}
}

func TestLoadPathAliases_ReactAnalyzerConfig(t *testing.T) {
	// Create temporary directory with .reactanalyzer.json
	tmpDir := t.TempDir()

	config := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/": ["app/"],
      "@lib/": ["library/"]
    }
  }
}`

	configPath := filepath.Join(tmpDir, ".reactanalyzer.json")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("Failed to write .reactanalyzer.json: %v", err)
	}

	// Load aliases
	aliases, err := LoadPathAliases(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load aliases: %v", err)
	}

	// Verify aliases
	if len(aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(aliases))
	}

	expectedApp := filepath.Clean(filepath.Join(tmpDir, "app"))
	actualApp := filepath.Clean(aliases["@/"])

	if actualApp != expectedApp {
		t.Errorf("Alias @/: expected %s, got %s", expectedApp, actualApp)
	}
}

func TestLoadPathAliases_PriorityOrder(t *testing.T) {
	// Create temporary directory with both config files
	tmpDir := t.TempDir()

	tsconfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/": ["src/*"],
      "@lib/": ["libraries/*"]
    }
  }
}`

	reactConfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/": ["app/*"]
    }
  }
}`

	tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
	if err := os.WriteFile(tsconfigPath, []byte(tsconfig), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig.json: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".reactanalyzer.json")
	if err := os.WriteFile(configPath, []byte(reactConfig), 0644); err != nil {
		t.Fatalf("Failed to write .reactanalyzer.json: %v", err)
	}

	// Load aliases
	aliases, err := LoadPathAliases(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load aliases: %v", err)
	}

	// .reactanalyzer.json should override tsconfig.json for @/
	expectedApp := filepath.Clean(filepath.Join(tmpDir, "app"))
	actualApp := filepath.Clean(aliases["@/"])

	if actualApp != expectedApp {
		t.Errorf("Expected .reactanalyzer.json to override tsconfig.json: expected %s, got %s", expectedApp, actualApp)
	}

	// @lib/ should still come from tsconfig.json
	expectedLib := filepath.Clean(filepath.Join(tmpDir, "libraries"))
	actualLib := filepath.Clean(aliases["@lib/"])

	if actualLib != expectedLib {
		t.Errorf("Alias @lib/: expected %s, got %s", expectedLib, actualLib)
	}
}

func TestFindLongestMatchingAlias(t *testing.T) {
	aliases := map[string]string{
		"@/":           "/project/src/",
		"@components/": "/project/src/components/",
		"~/":           "/project/",
	}

	tests := []struct {
		importPath     string
		expectedPrefix string
		expectedTarget string
		shouldMatch    bool
	}{
		{"@/utils/helper", "@/", "/project/src/", true},
		{"@components/Button", "@components/", "/project/src/components/", true},
		{"~/config", "~/", "/project/", true},
		{"react", "", "", false},
		{"./relative", "", "", false},
		{"@components/nested/deep/Button", "@components/", "/project/src/components/", true},
	}

	for _, tt := range tests {
		prefix, target, ok := FindLongestMatchingAlias(tt.importPath, aliases)

		if ok != tt.shouldMatch {
			t.Errorf("Import %s: expected match=%v, got %v", tt.importPath, tt.shouldMatch, ok)
			continue
		}

		if !ok {
			continue
		}

		if prefix != tt.expectedPrefix {
			t.Errorf("Import %s: expected prefix %s, got %s", tt.importPath, tt.expectedPrefix, prefix)
		}

		if target != tt.expectedTarget {
			t.Errorf("Import %s: expected target %s, got %s", tt.importPath, tt.expectedTarget, target)
		}
	}
}
