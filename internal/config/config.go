package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CompilerOptions represents TypeScript compiler options for path resolution
type CompilerOptions struct {
	BaseURL string              `json:"baseUrl"`
	Paths   map[string][]string `json:"paths"`
}

// Config represents the complete configuration for the analyzer
type Config struct {
	CompilerOptions CompilerOptions       `json:"compilerOptions,omitempty"`
	Rules           map[string]RuleConfig `json:"rules"`
	Ignore          []string              `json:"ignore,omitempty"` // Glob patterns for files to ignore
}

// RuleConfig represents configuration for a specific rule
type RuleConfig struct {
	Enabled bool                   `json:"enabled"`
	Options map[string]interface{} `json:"options"`
}

// DeepPropDrillingOptions represents options for the deep-prop-drilling rule
type DeepPropDrillingOptions struct {
	MaxDepth int `json:"maxDepth"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Ignore: []string{},
		Rules: map[string]RuleConfig{
			"deep-prop-drilling": {
				Enabled: true,
				Options: map[string]interface{}{
					"maxDepth": 3,
				},
			},
			"no-object-deps": {
				Enabled: true,
				Options: map[string]interface{}{},
			},
			"unstable-props-to-memo": {
				Enabled: true,
				Options: map[string]interface{}{},
			},
			"no-derived-state": {
				Enabled: true,
				Options: map[string]interface{}{},
			},
			"no-stale-state": {
				Enabled: true,
				Options: map[string]interface{}{},
			},
			"no-inline-props": {
				Enabled: true,
				Options: map[string]interface{}{},
			},
		},
	}
}

// Load loads configuration from a file
// Searches for .reactanalyzerrc.json starting from the given directory
// and walking up to the root, returning default config if not found
func Load(startDir string) (*Config, error) {
	cfg, _, err := LoadWithPath(startDir)
	return cfg, err
}

// LoadWithPath loads configuration and returns the path to the config file found
// Returns (config, configPath, error) where configPath is empty string if using defaults
func LoadWithPath(startDir string) (*Config, string, error) {
	// Start with default config
	config := DefaultConfig()

	// Search for config file starting from startDir and walking up
	configPath, err := findConfigFile(startDir)
	if err != nil {
		// No config file found, use defaults
		return config, "", nil
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config file: %v", err)
	}

	var userConfig Config
	if err := json.Unmarshal(data, &userConfig); err != nil {
		return nil, "", fmt.Errorf("failed to parse config file: %v", err)
	}

	// Merge user config with defaults
	mergeConfig(config, &userConfig)

	return config, configPath, nil
}

// findConfigFile searches for config files starting from dir
// and walking up to the root directory
func findConfigFile(dir string) (string, error) {
	configNames := []string{".rarc", ".reactanalyzerrc.json", "react-analyzer.json"}

	currentDir := dir
	for {
		for _, name := range configNames {
			configPath := filepath.Join(currentDir, name)
			if _, err := os.Stat(configPath); err == nil {
				return configPath, nil
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root directory
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("no config file found")
}

// mergeConfig merges user config into base config
func mergeConfig(base *Config, user *Config) {
	// Merge compiler options
	if user.CompilerOptions.BaseURL != "" {
		base.CompilerOptions.BaseURL = user.CompilerOptions.BaseURL
	}
	if user.CompilerOptions.Paths != nil {
		base.CompilerOptions.Paths = user.CompilerOptions.Paths
	}

	// Merge ignore patterns (user patterns replace defaults)
	if user.Ignore != nil {
		base.Ignore = user.Ignore
	}

	// Merge rules
	for ruleName, userRuleConfig := range user.Rules {
		baseRuleConfig, exists := base.Rules[ruleName]
		if !exists {
			// New rule in user config
			base.Rules[ruleName] = userRuleConfig
			continue
		}

		// Merge enabled status
		baseRuleConfig.Enabled = userRuleConfig.Enabled

		// Merge options
		if userRuleConfig.Options != nil {
			if baseRuleConfig.Options == nil {
				baseRuleConfig.Options = make(map[string]interface{})
			}
			for key, value := range userRuleConfig.Options {
				baseRuleConfig.Options[key] = value
			}
		}

		base.Rules[ruleName] = baseRuleConfig
	}
}

// GetRuleConfig returns the configuration for a specific rule
func (c *Config) GetRuleConfig(ruleName string) RuleConfig {
	if ruleConfig, exists := c.Rules[ruleName]; exists {
		return ruleConfig
	}
	// Return default enabled config
	return RuleConfig{
		Enabled: true,
		Options: map[string]interface{}{},
	}
}

// GetDeepPropDrillingOptions extracts and returns deep-prop-drilling options
func (c *Config) GetDeepPropDrillingOptions() DeepPropDrillingOptions {
	ruleConfig := c.GetRuleConfig("deep-prop-drilling")

	options := DeepPropDrillingOptions{
		MaxDepth: 3, // default: allow up to 3 components in chain
	}

	if val, ok := ruleConfig.Options["maxDepth"].(float64); ok {
		options.MaxDepth = int(val)
	}

	return options
}

// ShouldIgnore checks if a file path matches any of the ignore patterns
func (c *Config) ShouldIgnore(filePath string) bool {
	// Normalize path to use forward slashes for consistent matching
	normalizedPath := filepath.ToSlash(filePath)

	for _, pattern := range c.Ignore {
		if matchGlobPattern(normalizedPath, pattern) {
			return true
		}
	}

	return false
}

// matchGlobPattern implements simple glob pattern matching
// Supports: *, **, and negation with !
func matchGlobPattern(path, pattern string) bool {
	// Handle negation patterns (e.g., !src/important.tsx)
	if strings.HasPrefix(pattern, "!") {
		return !matchGlobPattern(path, pattern[1:])
	}

	// Normalize both to forward slashes
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Handle ** (match any number of directories)
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]

			// Remove trailing slash from prefix
			prefix = strings.TrimSuffix(prefix, "/")
			// Remove leading slash from suffix
			suffix = strings.TrimPrefix(suffix, "/")

			// Check prefix matches
			if prefix != "" {
				if !strings.HasPrefix(path, prefix+"/") && path != prefix {
					return false
				}
			}

			// Check suffix matches
			if suffix != "" {
				// For patterns like **/*.test.tsx, check if path ends with .test.tsx
				if strings.HasPrefix(suffix, "*") {
					// Use simple glob matching for the suffix
					return simpleGlobMatch(path, "*"+suffix)
				}
				// For patterns like **/__tests__/**, check if path contains the substring
				return strings.Contains(path, "/"+suffix+"/") ||
					strings.HasSuffix(path, "/"+suffix) ||
					strings.HasPrefix(path, suffix+"/")
			}

			return true
		}
	}

	// Handle * (match within a single directory level)
	if strings.Contains(pattern, "*") {
		return simpleGlobMatch(path, pattern)
	}

	// Exact match or substring match
	return path == pattern || strings.Contains(path, pattern) || strings.HasSuffix(path, "/"+pattern)
}

// simpleGlobMatch implements basic glob matching with * and ?
func simpleGlobMatch(path, pattern string) bool {
	// Convert glob pattern to simple matching logic
	// This is a simplified implementation - for production use filepath.Match or a glob library

	patternParts := strings.Split(pattern, "*")
	if len(patternParts) == 1 {
		// No wildcards, exact match
		return path == pattern
	}

	// Check if path contains all parts in order
	searchPath := path
	for i, part := range patternParts {
		if part == "" {
			continue
		}

		index := strings.Index(searchPath, part)
		if index == -1 {
			return false
		}

		// For first part, must be at the beginning (unless pattern starts with *)
		if i == 0 && !strings.HasPrefix(pattern, "*") && index != 0 {
			return false
		}

		// For last part, must be at the end (unless pattern ends with *)
		if i == len(patternParts)-1 && !strings.HasSuffix(pattern, "*") {
			return strings.HasSuffix(searchPath, part)
		}

		// Move search position forward
		searchPath = searchPath[index+len(part):]
	}

	return true
}
