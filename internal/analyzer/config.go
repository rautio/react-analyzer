package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// CompilerOptions represents the compilerOptions from tsconfig.json
type CompilerOptions struct {
	BaseURL string              `json:"baseUrl"`
	Paths   map[string][]string `json:"paths"`
}

// Config represents a configuration file (tsconfig.json or .reactanalyzer.json)
type Config struct {
	CompilerOptions CompilerOptions `json:"compilerOptions"`
}

// LoadPathAliases loads path aliases from configuration files
// Walks up the directory tree to find config files, similar to config.Load()
// Priority: .rarc > .reactanalyzerrc.json > .reactanalyzer.json > tsconfig.json
func LoadPathAliases(startDir string) (map[string]string, error) {
	aliases, _ := LoadPathAliasesWithPath(startDir)
	return aliases, nil
}

// LoadPathAliasesWithPath loads path aliases and returns the config file path
// Returns (aliases, configPath) where configPath is the file that contained the aliases
func LoadPathAliasesWithPath(startDir string) (map[string]string, string) {
	aliases := make(map[string]string)
	configNames := []string{"tsconfig.json", ".reactanalyzer.json", ".reactanalyzerrc.json", ".rarc"}
	foundConfigPath := ""

	// Walk up directory tree looking for config files
	currentDir := startDir
	for {
		// Try each config file in priority order (reverse, so higher priority overrides)
		for _, configName := range configNames {
			configPath := filepath.Join(currentDir, configName)
			if configAliases, err := loadConfigFile(configPath, currentDir); err == nil {
				// Merge aliases (higher priority files override lower priority)
				for k, v := range configAliases {
					aliases[k] = v
				}
				// Track the last (highest priority) config file that had aliases
				if len(configAliases) > 0 {
					foundConfigPath = configPath
				}
			}
		}

		// If we found any aliases, stop searching
		if len(aliases) > 0 {
			break
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root directory
			break
		}
		currentDir = parentDir
	}

	return aliases, foundConfigPath
}

// loadConfigFile loads and parses a config file (tsconfig.json or .reactanalyzer.json)
func loadConfigFile(configPath string, baseDir string) (map[string]string, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Convert paths to aliases
	return parsePathAliases(config.CompilerOptions, baseDir)
}

// parsePathAliases converts tsconfig-style paths to simple alias map
func parsePathAliases(opts CompilerOptions, baseDir string) (map[string]string, error) {
	aliases := make(map[string]string)

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "."
	}

	// Convert baseURL to absolute path
	absBaseURL := filepath.Join(baseDir, baseURL)

	for alias, targets := range opts.Paths {
		if len(targets) == 0 {
			continue
		}

		// tsconfig uses glob patterns: "@/*" -> ["src/*"]
		// Convert to prefix match: "@/" -> "src/"
		aliasPrefix := strings.TrimSuffix(alias, "*")
		targetPath := strings.TrimSuffix(targets[0], "*") // Use first target only

		// Combine with baseURL to get absolute path
		fullTarget := filepath.Join(absBaseURL, targetPath)

		// Clean the path
		fullTarget = filepath.Clean(fullTarget)

		aliases[aliasPrefix] = fullTarget
	}

	return aliases, nil
}

// FindLongestMatchingAlias finds the longest alias prefix that matches the import path
// This handles cases where multiple aliases could match (e.g., "@/" and "@/components/")
func FindLongestMatchingAlias(importPath string, aliases map[string]string) (string, string, bool) {
	var longestPrefix string
	var longestTarget string

	for prefix, target := range aliases {
		if strings.HasPrefix(importPath, prefix) {
			if len(prefix) > len(longestPrefix) {
				longestPrefix = prefix
				longestTarget = target
			}
		}
	}

	if longestPrefix != "" {
		return longestPrefix, longestTarget, true
	}

	return "", "", false
}
