# Configuration Guide

React Analyzer can be configured using a `.rarc` or `.reactanalyzerrc.json` file in your project root.

## Configuration File

Create a `.rarc` (shorthand) or `.reactanalyzerrc.json` file in your project root:

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"]
    }
  },
  "ignore": [
    "**/*.test.tsx",
    "**/__tests__/**",
    "**/*.stories.tsx"
  ],
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "options": {
        "maxDepth": 3
      }
    }
  }
}
```

The configuration file supports three main sections:
- **`compilerOptions`** - TypeScript path aliases for module resolution (optional)
- **`ignore`** - File/directory patterns to exclude from analysis (optional, **coming in Phase 4**)
- **`rules`** - Rule-specific configuration (optional)

## File Discovery

### Rule Configuration
The analyzer searches for rule configuration files in the following order:
1. `.rarc` in the current directory (shorthand)
2. `.reactanalyzerrc.json` in the current directory
3. `react-analyzer.json` in the current directory
4. Walks up parent directories until a config is found
5. Uses default configuration if no file is found

### Path Aliases (Module Resolution)
The analyzer searches for path aliases in the following order (highest to lowest priority):
1. `.rarc` - `compilerOptions.paths` section (shorthand)
2. `.reactanalyzerrc.json` - `compilerOptions.paths` section
3. `.reactanalyzer.json` - `compilerOptions.paths` section (legacy)
4. `tsconfig.json` - Falls back to TypeScript config if no analyzer config exists

**Recommended:** Use `.rarc` for a shorter filename, or `.reactanalyzerrc.json` for clarity.

## Compiler Options (Path Aliases)

The `compilerOptions` section allows you to configure TypeScript-style path aliases for module resolution. This helps the analyzer understand your import paths when detecting issues across files.

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"],
      "@utils/*": ["src/utils/*"],
      "~/*": ["./*"]
    }
  }
}
```

**Options:**
- `baseUrl` - Base directory for resolving non-relative module names (default: ".")
- `paths` - Path mapping entries (supports glob patterns with `*`)

**Example mappings:**
- `@/*` → `src/*` - Maps `@/components/Button` to `src/components/Button`
- `@components/*` → `src/components/*` - Maps `@components/Button` to `src/components/Button`

**Note:** If you already have a `tsconfig.json`, the analyzer will automatically read path aliases from it. You only need to add `compilerOptions` to `.reactanalyzerrc.json` if you want to override or add additional aliases.

## Ignore Patterns (Coming in Phase 4)

The `ignore` section allows you to exclude specific files and directories from analysis. This is useful for skipping test files, story files, or legacy code.

```json
{
  "ignore": [
    "**/*.test.tsx",
    "**/*.test.ts",
    "**/__tests__/**",
    "**/*.spec.tsx",
    "**/*.stories.tsx",
    "**/storybook/**",
    "src/legacy/**",
    "!src/legacy/important.tsx"
  ]
}
```

**Pattern Syntax:**
- Supports glob patterns (`**/*.test.tsx` matches all test files recursively)
- Supports directory exclusions (`**/stories/**` excludes all story directories)
- Supports negation (`!src/legacy/important.tsx` includes a specific file)
- Similar to ESLint's `ignorePatterns` and Jest's `testPathIgnorePatterns`

**Default Ignores:**
The analyzer always ignores these directories (hardcoded):
- `node_modules/`
- `dist/`
- `build/`
- `.git/`

**CLI Override:**
```bash
# Use custom ignore file
react-analyzer src/ --ignore-path .analyzerignore

# Add patterns via CLI
react-analyzer src/ --ignore "**/*.test.tsx"
```

**Note:** This feature is planned for Phase 4 (see [ROADMAP.md](ROADMAP.md#4d-ignore-pattern-support-1-week)).

## Rule Configuration

Each rule can be configured with:
- `enabled`: Boolean to enable/disable the rule
- `options`: Object with rule-specific options

### Available Rules

#### `deep-prop-drilling`

Detects props passed through multiple component levels without being used.

**Options:**
- `maxDepth` (number, default: 3)
  - Maximum depth of component chain allowed before warning
  - Depth is the total number of components in the chain (origin → consumer)
  - Set to `2` for stricter checking (only direct parent-child)
  - Set to `3` for balanced approach (one intermediate layer, default)
  - Set to `4+` for lenient checking (multiple intermediate layers)

**Behavior:**
- `maxDepth: 2` - Only allows: App → Sidebar (direct)
  - Warns about: App → Dashboard → Sidebar
- `maxDepth: 3` - Allows: App → Dashboard → Sidebar (default)
  - Warns about: App → Dashboard → Sidebar → Footer
- `maxDepth: 4` - Allows: App → A → B → C
  - Warns about: App → A → B → C → D

#### `no-object-deps`

Detects inline objects/arrays in hook dependency arrays.

**Options:** None

#### `unstable-props-to-memo`

Detects unstable props passed to memoized components.

**Options:** None

#### `no-derived-state`

Detects useState mirroring props via useEffect.

**Options:** None

#### `no-stale-state`

Detects state updates without functional form.

**Options:** None

#### `no-inline-props`

Detects inline objects/arrays/functions in JSX props.

**Options:** None

## Example Configurations

### Strict Mode (Catch Everything)

```json
{
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "options": {
        "maxDepth": 2
      }
    },
    "no-object-deps": { "enabled": true, "options": {} },
    "unstable-props-to-memo": { "enabled": true, "options": {} },
    "no-derived-state": { "enabled": true, "options": {} },
    "no-stale-state": { "enabled": true, "options": {} },
    "no-inline-props": { "enabled": true, "options": {} }
  }
}
```

### Balanced Mode (Default)

```json
{
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "options": {
        "maxDepth": 3
      }
    }
  }
}
```

### Lenient Mode (Only Critical Issues)

```json
{
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "options": {
        "maxDepth": 4
      }
    },
    "no-inline-props": { "enabled": false, "options": {} }
  }
}
```

### Disable Specific Rules

```json
{
  "rules": {
    "deep-prop-drilling": {
      "enabled": false,
      "options": {}
    }
  }
}
```

## VS Code Integration

The VS Code extension automatically picks up `.reactanalyzerrc.json` from your workspace.

No additional configuration needed!

## Future Rule Options

As more rules are added, they will support configuration through the same pattern:

```json
{
  "rules": {
    "rule-name": {
      "enabled": true,
      "options": {
        "option1": "value1",
        "option2": 123
      }
    }
  }
}
```

## Default Configuration

If no config file is found, the following defaults are used:

```json
{
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "options": {
        "maxDepth": 3
      }
    },
    "no-object-deps": { "enabled": true, "options": {} },
    "unstable-props-to-memo": { "enabled": true, "options": {} },
    "no-derived-state": { "enabled": true, "options": {} },
    "no-stale-state": { "enabled": true, "options": {} },
    "no-inline-props": { "enabled": true, "options": {} }
  }
}
```

## Future Enhancements

### JSON Schema Support

We plan to add a JSON Schema file (`config-schema.json`) to provide IDE autocomplete and validation for `.reactanalyzerrc.json` files. This will enable:

- **Autocomplete** - IntelliSense for available rules and options in VS Code and other editors
- **Validation** - Real-time error checking for configuration syntax
- **Documentation** - Inline descriptions and examples while editing config files

Once available, you'll be able to reference it in your config:

```json
{
  "$schema": "https://raw.githubusercontent.com/rautio/react-analyzer/main/config-schema.json",
  "rules": { ... }
}
```
