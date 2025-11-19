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
// Priority: .rarc > .reactanalyzerrc.json > .reactanalyzer.json > tsconfig.json
func LoadPathAliases(baseDir string) (map[string]string, error) {
	aliases := make(map[string]string)

	// Try tsconfig.json first (lowest priority)
	tsconfigPath := filepath.Join(baseDir, "tsconfig.json")
	if tsAliases, err := loadConfigFile(tsconfigPath, baseDir); err == nil {
		aliases = tsAliases
	}

	// Try .reactanalyzer.json (low priority, will override)
	reactAnalyzerPath := filepath.Join(baseDir, ".reactanalyzer.json")
	if configAliases, err := loadConfigFile(reactAnalyzerPath, baseDir); err == nil {
		// Merge/override with .reactanalyzer.json aliases
		for k, v := range configAliases {
			aliases[k] = v
		}
	}

	// Try .reactanalyzerrc.json (medium priority, will override)
	reactAnalyzerRcPath := filepath.Join(baseDir, ".reactanalyzerrc.json")
	if configAliases, err := loadConfigFile(reactAnalyzerRcPath, baseDir); err == nil {
		// Merge/override with .reactanalyzerrc.json aliases
		for k, v := range configAliases {
			aliases[k] = v
		}
	}

	// Try .rarc (highest priority, will override)
	rarcPath := filepath.Join(baseDir, ".rarc")
	if configAliases, err := loadConfigFile(rarcPath, baseDir); err == nil {
		// Merge/override with .rarc aliases
		for k, v := range configAliases {
			aliases[k] = v
		}
	}

	return aliases, nil
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
