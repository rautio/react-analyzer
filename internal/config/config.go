package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	// Start with default config
	config := DefaultConfig()

	// Search for config file starting from startDir and walking up
	configPath, err := findConfigFile(startDir)
	if err != nil {
		// No config file found, use defaults
		return config, nil
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var userConfig Config
	if err := json.Unmarshal(data, &userConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Merge user config with defaults
	mergeConfig(config, &userConfig)

	return config, nil
}

// findConfigFile searches for .reactanalyzerrc.json starting from dir
// and walking up to the root directory
func findConfigFile(dir string) (string, error) {
	configNames := []string{".reactanalyzerrc.json", "react-analyzer.json"}

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
