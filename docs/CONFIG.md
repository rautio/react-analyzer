# Configuration Guide

React Analyzer can be configured using a `.reactanalyzerrc.json` file in your project root.

## Configuration File

Create a `.reactanalyzerrc.json` file in your project root:

```json
{
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "options": {
        "minPassthroughComponents": 2
      }
    }
  }
}
```

## File Discovery

The analyzer searches for configuration files in the following order:
1. `.reactanalyzerrc.json` in the current directory
2. `react-analyzer.json` in the current directory
3. Walks up parent directories until a config is found
4. Uses default configuration if no file is found

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

**Example:**
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
