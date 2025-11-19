package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Check that all expected rules are present
	expectedRules := []string{
		"deep-prop-drilling",
		"no-object-deps",
		"unstable-props-to-memo",
		"no-derived-state",
		"no-stale-state",
		"no-inline-props",
	}

	for _, ruleName := range expectedRules {
		if _, exists := cfg.Rules[ruleName]; !exists {
			t.Errorf("Expected rule %s not found in default config", ruleName)
		}
	}

	// Verify deep-prop-drilling default options
	deepPropDrillingConfig := cfg.Rules["deep-prop-drilling"]
	if !deepPropDrillingConfig.Enabled {
		t.Error("deep-prop-drilling should be enabled by default")
	}

	maxDepth, ok := deepPropDrillingConfig.Options["maxDepth"].(int)
	if !ok {
		t.Error("maxDepth option should be an int")
	}
	if maxDepth != 3 {
		t.Errorf("Expected maxDepth=3, got %d", maxDepth)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	// Create temp directory with no config file
	tempDir := t.TempDir()

	cfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load() should not error when no config file exists: %v", err)
	}

	// Should return default config
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Verify it's the default config
	if _, exists := cfg.Rules["deep-prop-drilling"]; !exists {
		t.Error("Expected default config to have deep-prop-drilling rule")
	}
}

func TestLoadConfig_RarcFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create .rarc config file
	configContent := map[string]interface{}{
		"rules": map[string]interface{}{
			"deep-prop-drilling": map[string]interface{}{
				"enabled": false,
				"options": map[string]interface{}{
					"maxDepth": 5,
				},
			},
		},
	}

	configPath := filepath.Join(tempDir, ".rarc")
	writeJSONFile(t, configPath, configContent)

	cfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify custom config was loaded
	deepPropDrilling := cfg.Rules["deep-prop-drilling"]
	if deepPropDrilling.Enabled {
		t.Error("Expected deep-prop-drilling to be disabled per config")
	}

	maxDepth, ok := deepPropDrilling.Options["maxDepth"].(float64) // JSON unmarshals numbers as float64
	if !ok {
		t.Fatalf("maxDepth should be a number")
	}
	if int(maxDepth) != 5 {
		t.Errorf("Expected maxDepth=5, got %d", int(maxDepth))
	}
}

func TestLoadConfig_ReactAnalyzerRcFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create .reactanalyzerrc.json config file
	configContent := map[string]interface{}{
		"rules": map[string]interface{}{
			"no-object-deps": map[string]interface{}{
				"enabled": false,
			},
		},
	}

	configPath := filepath.Join(tempDir, ".reactanalyzerrc.json")
	writeJSONFile(t, configPath, configContent)

	cfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify custom config was loaded
	noObjectDeps := cfg.Rules["no-object-deps"]
	if noObjectDeps.Enabled {
		t.Error("Expected no-object-deps to be disabled per config")
	}
}

func TestLoadConfig_Priority(t *testing.T) {
	tempDir := t.TempDir()

	// Create both .rarc and .reactanalyzerrc.json
	// .rarc should take priority

	rarcContent := map[string]interface{}{
		"rules": map[string]interface{}{
			"deep-prop-drilling": map[string]interface{}{
				"enabled": false, // Disabled in .rarc
			},
		},
	}

	reactAnalyzerRcContent := map[string]interface{}{
		"rules": map[string]interface{}{
			"deep-prop-drilling": map[string]interface{}{
				"enabled": true, // Enabled in .reactanalyzerrc.json
			},
		},
	}

	writeJSONFile(t, filepath.Join(tempDir, ".rarc"), rarcContent)
	writeJSONFile(t, filepath.Join(tempDir, ".reactanalyzerrc.json"), reactAnalyzerRcContent)

	cfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// .rarc should take priority
	deepPropDrilling := cfg.Rules["deep-prop-drilling"]
	if deepPropDrilling.Enabled {
		t.Error("Expected .rarc to take priority and disable deep-prop-drilling")
	}
}

func TestLoadConfig_WalkUpDirectories(t *testing.T) {
	// Create nested directory structure
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "dir")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	// Put config in parent directory
	configContent := map[string]interface{}{
		"rules": map[string]interface{}{
			"no-inline-props": map[string]interface{}{
				"enabled": false,
			},
		},
	}
	writeJSONFile(t, filepath.Join(tempDir, ".rarc"), configContent)

	// Load from nested directory - should walk up and find config
	cfg, err := Load(nestedDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify config was found
	noInlineProps := cfg.Rules["no-inline-props"]
	if noInlineProps.Enabled {
		t.Error("Expected config from parent directory to be loaded")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Write invalid JSON
	configPath := filepath.Join(tempDir, ".rarc")
	if err := os.WriteFile(configPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	_, err := Load(tempDir)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetRuleConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test existing rule
	ruleConfig := cfg.GetRuleConfig("deep-prop-drilling")
	if !ruleConfig.Enabled {
		t.Error("Expected deep-prop-drilling to be enabled")
	}

	// Test non-existent rule - should return default enabled config
	unknownRule := cfg.GetRuleConfig("unknown-rule")
	if !unknownRule.Enabled {
		t.Error("Expected unknown rule to default to enabled")
	}
}

func TestGetDeepPropDrillingOptions(t *testing.T) {
	cfg := DefaultConfig()

	options := cfg.GetDeepPropDrillingOptions()
	if options.MaxDepth != 3 {
		t.Errorf("Expected default maxDepth=3, got %d", options.MaxDepth)
	}

	// Test with custom config
	cfg.Rules["deep-prop-drilling"] = RuleConfig{
		Enabled: true,
		Options: map[string]interface{}{
			"maxDepth": float64(5), // JSON unmarshals as float64
		},
	}

	options = cfg.GetDeepPropDrillingOptions()
	if options.MaxDepth != 5 {
		t.Errorf("Expected custom maxDepth=5, got %d", options.MaxDepth)
	}
}

func TestGetDeepPropDrillingOptions_MissingOption(t *testing.T) {
	cfg := &Config{
		Rules: map[string]RuleConfig{
			"deep-prop-drilling": {
				Enabled: true,
				Options: map[string]interface{}{},
			},
		},
	}

	options := cfg.GetDeepPropDrillingOptions()
	// Should use default when option is missing
	if options.MaxDepth != 3 {
		t.Errorf("Expected default maxDepth=3 when option missing, got %d", options.MaxDepth)
	}
}

func TestMergeConfig(t *testing.T) {
	base := DefaultConfig()
	user := &Config{
		Rules: map[string]RuleConfig{
			"deep-prop-drilling": {
				Enabled: false,
				Options: map[string]interface{}{
					"maxDepth": 10,
				},
			},
			"new-rule": {
				Enabled: true,
				Options: map[string]interface{}{},
			},
		},
	}

	mergeConfig(base, user)

	// Check that base rules were updated
	deepPropDrilling := base.Rules["deep-prop-drilling"]
	if deepPropDrilling.Enabled {
		t.Error("Expected deep-prop-drilling to be disabled after merge")
	}

	maxDepth, ok := deepPropDrilling.Options["maxDepth"].(int)
	if !ok {
		t.Fatal("maxDepth should be an int")
	}
	if maxDepth != 10 {
		t.Errorf("Expected maxDepth=10 after merge, got %d", maxDepth)
	}

	// Check that new rule was added
	if _, exists := base.Rules["new-rule"]; !exists {
		t.Error("Expected new-rule to be added to base config")
	}
}

func TestCompilerOptions(t *testing.T) {
	tempDir := t.TempDir()

	// Create config with compilerOptions
	configContent := map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			"baseUrl": "src",
			"paths": map[string]interface{}{
				"@/*": []string{"src/*"},
			},
		},
		"rules": map[string]interface{}{},
	}

	configPath := filepath.Join(tempDir, ".rarc")
	writeJSONFile(t, configPath, configContent)

	cfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify compilerOptions were loaded
	if cfg.CompilerOptions.BaseURL != "src" {
		t.Errorf("Expected baseUrl='src', got '%s'", cfg.CompilerOptions.BaseURL)
	}

	if cfg.CompilerOptions.Paths == nil {
		t.Fatal("Expected paths to be loaded")
	}

	paths, exists := cfg.CompilerOptions.Paths["@/*"]
	if !exists {
		t.Error("Expected @/* path mapping to exist")
	}

	if len(paths) != 1 || paths[0] != "src/*" {
		t.Errorf("Expected @/* -> src/*, got %v", paths)
	}
}

// Helper function to write JSON to file
func writeJSONFile(t *testing.T, path string, content interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
}
