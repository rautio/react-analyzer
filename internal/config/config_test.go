package config

import (
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		expected bool
	}{
		{
			name:     "match test file with .test.tsx",
			patterns: []string{"**/*.test.tsx"},
			path:     "src/components/Button.test.tsx",
			expected: true,
		},
		{
			name:     "match test file in __tests__ directory",
			patterns: []string{"**/__tests__/**"},
			path:     "src/components/__tests__/Button.tsx",
			expected: true,
		},
		{
			name:     "match stories file",
			patterns: []string{"**/*.stories.tsx"},
			path:     "src/components/Button.stories.tsx",
			expected: true,
		},
		{
			name:     "do not match regular file",
			patterns: []string{"**/*.test.tsx", "**/__tests__/**"},
			path:     "src/components/Button.tsx",
			expected: false,
		},
		{
			name:     "match spec file",
			patterns: []string{"**/*.spec.ts"},
			path:     "src/utils/helper.spec.ts",
			expected: true,
		},
		{
			name:     "match deep nested test file",
			patterns: []string{"**/*.test.tsx"},
			path:     "src/features/auth/components/LoginForm.test.tsx",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Ignore: tt.patterns,
			}

			result := cfg.ShouldIgnore(tt.path)
			if result != tt.expected {
				t.Errorf("ShouldIgnore(%q) with patterns %v = %v, want %v",
					tt.path, tt.patterns, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test that default config has empty ignore patterns
	if len(cfg.Ignore) != 0 {
		t.Errorf("Default config should have empty ignore patterns, got %d patterns", len(cfg.Ignore))
	}

	// Test that default config does NOT ignore any files (including test files)
	testFiles := []string{
		"src/components/Button.test.tsx",
		"src/components/Button.spec.ts",
		"src/__tests__/helper.ts",
		"src/__mocks__/api.ts",
		"src/components/Button.stories.tsx",
		"src/components/Button.tsx",
		"src/utils/helper.ts",
		"src/App.tsx",
	}

	for _, path := range testFiles {
		if cfg.ShouldIgnore(path) {
			t.Errorf("Default config should NOT ignore %q but does", path)
		}
	}
}
