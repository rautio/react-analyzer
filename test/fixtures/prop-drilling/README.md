# Prop Drilling Test Fixtures

This directory contains test fixtures for the prop drilling detection algorithm. Each fixture tests different scenarios and edge cases.

## Overview

Prop drilling occurs when props are passed through multiple component levels where intermediate components don't use the props - they only pass them down to descendants. Our algorithm detects violations when props are drilled through 3+ levels.

## Test Fixtures

### 1. SimpleDrilling.tsx

**Purpose:** Basic 3-level prop drilling scenario

**Structure:**
```
App (defines count) → Parent → Child → Display (uses count)
```

**Expected Violations:** 1
- `count` drilled through 3 levels
- Passthrough components: Parent, Child

**Complexity:** Simple

---

### 2. NoDrilling.tsx

**Purpose:** No prop drilling (negative test case)

**Structure:**
```
App (defines count) → Parent (uses count directly)
```

**Expected Violations:** 0
- Parent uses count immediately, no drilling

**Complexity:** Simple

---

### 3. MultiplePropsDrilled.tsx

**Purpose:** Multiple props drilled together

**Structure:**
```
App (defines theme, lang) → Parent → Child → Display (uses theme, lang)
```

**Expected Violations:** 2
- `theme` drilled through 3 levels
- `lang` drilled through 3 levels
- Tests algorithm's ability to track multiple props independently

**Complexity:** Simple

---

### 4. PartialUsage.tsx

**Purpose:** Mixed usage - component uses AND passes prop

**Structure:**
```
App (defines theme) → Parent (uses AND passes theme) → Child → Display (uses theme)
```

**Expected Violations:** 0
- Parent uses theme locally (for styling), so it's not pure passthrough
- Only Child is passthrough, which is < 2 intermediate components
- Tests algorithm's ability to distinguish between pure passthrough and actual usage

**Complexity:** Medium

---

### 5. ThemeProvider.tsx

**Purpose:** Real-world theme drilling scenario

**Structure:**
```
App (defines theme, setTheme) → Dashboard → Sidebar → ThemeToggle (uses theme, setTheme)
```

**Expected Violations:** 2
- `theme` drilled through 3 levels
- `onThemeChange` (setTheme) drilled through 3 levels
- Realistic example showing WHY Context API should be used
- Compare with ContextSolution.tsx for the correct approach

**Complexity:** Medium (realistic)

---

### 6. UserProfile.tsx

**Purpose:** User data drilling through profile components

**Structure:**
```
App (defines user data) → ProfilePage → ProfileContent → UserHeader (uses user data)
```

**Expected Violations:** 3
- `userId` drilled through 3 levels
- `userName` drilled through 3 levels
- `userAvatar` drilled through 3 levels
- Common pattern in apps that fetch user data at top level

**Complexity:** Medium (realistic)

---

### 7. Dashboard.tsx

**Purpose:** Complex dashboard with multiple configuration props

**Structure:**
```
App (defines settings) → DashboardLayout → MainArea → MetricsPanel → MetricCards (uses settings)
```

**Expected Violations:** 3
- `locale` drilled through 4 levels
- `currency` drilled through 4 levels
- `dateFormat` drilled through 4 levels
- Realistic e-commerce or analytics dashboard scenario
- Tests deeper nesting (4 levels)

**Complexity:** High (realistic)

---

### 8. DeepNesting.tsx

**Purpose:** Very deep prop drilling (6 levels)

**Structure:**
```
App → PageWrapper → ContentArea → FeatureSection → WidgetContainer → Widget → DataDisplay
```

**Expected Violations:** 1
- `apiKey` drilled through 6 levels
- Tests algorithm's ability to handle deeply nested component hierarchies
- Tests performance with deep traversal

**Complexity:** High (stress test)

---

### 9. MixedScenario.tsx

**Purpose:** Some props drilled, others used at each level

**Structure:**
```
App (defines title, userId, isLoading)
  → Layout (uses title, isLoading; passes userId)
    → Content (passes userId)
      → UserPanel (uses userId)
```

**Expected Violations:** 1
- Only `userId` is drilled (through Layout → Content → UserPanel)
- `title` and `isLoading` are used by Layout, so they're not drilled
- Tests algorithm's ability to distinguish between drilled and locally-used props

**Complexity:** Medium

---

### 10. PropSpread.tsx

**Purpose:** Props passed via spread operator

**Structure:**
```
App (defines config object) → Container {...config} → Panel {...config} → Settings (uses config)
```

**Expected Violations:** Depends on implementation
- Tests prop spreading pattern
- More advanced case - may require Phase 2.2 for full detection
- Individual config properties (apiUrl, timeout, retries) are spread

**Complexity:** High (edge case)

**Note:** This is a Phase 2.2 feature - basic detection may not catch spread patterns initially.

---

### 11. FormWithValidation.tsx

**Purpose:** Form validation config drilled through form components

**Structure:**
```
App (defines validationRules) → UserForm → FormSection → InputField (uses validationRules)
```

**Expected Violations:** 1
- `validationRules` drilled through 3 levels
- Common pattern in form libraries and custom form implementations
- Realistic example showing when Context or custom hooks would be better

**Complexity:** Medium (realistic)

---

### 12. ContextSolution.tsx

**Purpose:** CORRECT solution using Context API (negative test)

**Structure:**
```
App (ThemeContext.Provider)
  → Dashboard
    → Sidebar
      → ThemeToggle (useContext to consume theme)
```

**Expected Violations:** 0
- Theme accessed via Context API instead of prop drilling
- Demonstrates the recommended solution
- Compare with ThemeProvider.tsx to see the improvement
- No props passed through intermediate components

**Complexity:** Medium (educational)

---

## Testing Strategy

### Phase 2.1 (Basic Detection)

**Must Handle:**
- SimpleDrilling.tsx ✓
- NoDrilling.tsx ✓
- MultiplePropsDrilled.tsx ✓
- PartialUsage.tsx ✓
- ThemeProvider.tsx ✓
- UserProfile.tsx ✓
- MixedScenario.tsx ✓
- ContextSolution.tsx ✓

**Should Handle:**
- Dashboard.tsx (deeper nesting)
- DeepNesting.tsx (stress test)
- FormWithValidation.tsx (realistic)

**May Not Handle (Phase 2.2):**
- PropSpread.tsx (requires spread analysis)

### Test Metrics

For each fixture, validate:
1. **Violation Count:** Number of violations detected
2. **Drilling Depth:** Number of levels for each drilled prop
3. **Passthrough Components:** List of components that don't use the prop
4. **Origin Location:** Where the prop is defined
5. **Consumer Location:** Where the prop is actually used

### Example Expected Output

For `SimpleDrilling.tsx`:
```
src/test/fixtures/prop-drilling/SimpleDrilling.tsx:8:5
  [deep-prop-drilling] Prop 'count' is drilled through 3 component levels.
  Consider using Context API instead.

  Path:
    App (defines 'count') → Parent → Child → Display (uses 'count')

  Passthrough components:
    - Parent (SimpleDrilling.tsx:13)
    - Child (SimpleDrilling.tsx:18)
```

## Future Enhancements (Phase 2.2)

Additional test fixtures to create:
1. **PropTransformation.tsx** - Props that are transformed along the path
2. **ConditionalPassing.tsx** - Props passed conditionally
3. **ObjectDestructuring.tsx** - Destructuring patterns
4. **ArrayProps.tsx** - Array props being drilled
5. **CallbackProps.tsx** - Callback functions being drilled
6. **MultipleConsumers.tsx** - Same prop used by multiple descendants

## References

See design documents for detailed algorithm:
- `docs/design/phase2_prop_drilling_detection.md` - Detection algorithm
- `docs/design/phase2_graph_data_structures.md` - Graph data structures
- `docs/design/phase2_vscode_extension_architecture.md` - VS Code integration
