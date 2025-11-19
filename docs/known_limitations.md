# Known Limitations

This document outlines the current limitations of the react-analyzer tool, particularly for the prop drilling detection feature.

## Prop Drilling Detection Limitations

### 1. Partial Prop Usage Detection

**Issue:** The tool cannot detect when a component both uses a prop locally AND passes it to children.

**Current Behavior:** If a component passes a prop to a child, it's assumed to be a pure passthrough, even if the component also uses the prop.

**Example:**
```tsx
function Parent({ theme }: { theme: string }) {
    // Uses theme for its own styling
    const styles = { background: theme === 'dark' ? '#000' : '#fff' };

    return (
        <div style={styles}>
            <Child theme={theme} />  {/* Also passes it down */}
        </div>
    );
}
```

**Expected:** No violation (Parent uses theme, so not a pure passthrough)
**Actual:** Violation reported if drilling depth >= 3

**Why:** The `componentUsesProp()` heuristic only checks if a prop is passed to children, not whether it's also referenced in the component's code.

**Fix Required:** AST analysis to detect prop references in:
- JSX attributes and content
- Variable declarations and expressions
- Function calls and return statements
- Conditional statements and loops

**Priority:** Medium - Affects accuracy but most real-world cases involve pure passthroughs

---

### 2. Prop Spread Operators

**Issue:** Props passed via spread operators are not tracked.

**Current Behavior:** Spread operators (`{...props}`) are ignored, so drilling through spreads is not detected.

**Example:**
```tsx
function Container(props: Config) {
    return <Panel {...props} />;  // Not tracked
}

function Panel(props: Config) {
    return <Settings {...props} />;  // Not tracked
}

function Settings({ apiUrl, timeout }: Config) {
    return <div>{apiUrl}</div>;  // Uses individual props
}
```

**Expected:** Violation for props drilled through Container → Panel → Settings
**Actual:** No violation detected

**Why:** The edge detection in `processJSXElement()` only handles explicit prop assignments like `<Child prop={value} />`, not spread syntax.

**Fix Required:**
1. Detect `jsx_spread_attribute` nodes in JSX
2. Track which props are included in the spread
3. Create edges for all props in the spread object
4. Handle type information to determine spread contents

**Priority:** High - Common pattern in React codebases

**Planned:** Phase 2.2

---

### 3. Object Property Access

**Issue:** Props accessed as object properties are not tracked as separate prop flows.

**Current Behavior:** When state is stored as an object and individual properties are passed, only the object-level state is tracked, not the individual properties.

**Example:**
```tsx
function App() {
    const [settings, setSettings] = useState({
        locale: 'en-US',
        currency: 'USD',
        dateFormat: 'MM/DD/YYYY'
    });

    return <Layout locale={settings.locale} currency={settings.currency} />;
}
```

**Expected:** Track `locale` and `currency` as separate prop flows
**Actual:** Only tracks `settings` as a single state node; property accesses not detected

**Why:**
1. Edge detection looks for identifier nodes (`{theme}`), not member expressions (`{settings.locale}`)
2. State extraction creates one node per `useState` call, not per property
3. No virtual state nodes for object properties

**Fix Required:**
1. Detect `member_expression` nodes when creating edges (e.g., `settings.locale`)
2. Either:
   - Create virtual state nodes for each property, OR
   - Track property access paths in edges
3. Update prop matching to handle property chains

**Priority:** Medium-High - Common pattern in TypeScript codebases

**Planned:** Phase 2.2 or Phase 3

---

### 4. Cross-File Prop Drilling

**Issue:** Prop drilling detection is limited to components defined in the same file.

**Current Behavior:** The `findComponentID()` function only searches for components in the same file as the parent.

**Example:**
```tsx
// App.tsx
import { Dashboard } from './Dashboard';

function App() {
    const [theme, setTheme] = useState('dark');
    return <Dashboard theme={theme} />;
}

// Dashboard.tsx
export function Dashboard({ theme }: Props) {
    return <Sidebar theme={theme} />;  // Not tracked
}
```

**Expected:** Track prop drilling across file boundaries
**Actual:** Edge not created from Dashboard to Sidebar if they're in different files

**Why:** The `findComponentID()` function has a TODO comment for cross-file resolution but doesn't yet use the module resolver.

**Fix Required:**
1. Track and resolve import statements
2. Search for component definitions in imported modules
3. Build cross-file edges in the graph
4. Handle circular dependencies

**Priority:** High - Multi-file components are standard in React

**Planned:** Phase 2.2

**Code Location:** `internal/graph/builder.go:496-509`

---

### 5. Arrow Function Components

**Issue:** Components defined as arrow functions are not detected.

**Current Behavior:** Only `function_declaration` nodes are processed; arrow function components are ignored.

**Example:**
```tsx
const MyComponent = ({ theme }: Props) => {
    return <div className={theme}>Content</div>;
};
```

**Expected:** Detect and track arrow function components
**Actual:** Component not added to graph

**Why:** The AST walker in `buildComponentNodes()` only handles `function_declaration` node types.

**Fix Required:**
1. Detect `variable_declarator` with arrow function initializers
2. Check if the arrow function returns JSX
3. Validate PascalCase naming convention
4. Extract props from arrow function parameters

**Priority:** High - Very common pattern in modern React

**Planned:** Phase 2.2

**Code Location:** `internal/graph/builder.go:106-107` (TODO comment exists)

---

### 6. Renamed Props

**Issue:** Props renamed during passing are not tracked correctly.

**Current Behavior:** Only tracks props with the same name throughout the chain.

**Example:**
```tsx
function Parent({ appTheme }: Props) {
    return <Child theme={appTheme} />;  // Renamed from appTheme to theme
}
```

**Expected:** Track that `appTheme` flows to `theme`
**Actual:** No edge created (name mismatch)

**Why:** The `isParentVariable()` check matches variable names exactly without considering renames.

**Fix Required:**
1. Track prop name mappings in edges (e.g., "appTheme -> theme")
2. Update path tracing to follow name changes
3. Report original prop name in violations

**Priority:** Low-Medium - Less common pattern

---

### 7. Dynamic Prop Names

**Issue:** Props with computed/dynamic names are not detected.

**Example:**
```tsx
const propName = 'theme';
return <Child {...{ [propName]: value }} />;
```

**Expected:** Detect computed prop names
**Actual:** Not tracked

**Why:** Edge detection only handles static property identifiers.

**Fix Required:** Evaluate constant expressions to determine prop names (complex, potentially unsound)

**Priority:** Low - Uncommon pattern, often an anti-pattern

---

### 8. Conditional Prop Passing

**Issue:** Props passed conditionally are always tracked as if they're passed.

**Example:**
```tsx
function Parent({ theme, showChild }: Props) {
    return showChild ? <Child theme={theme} /> : null;
}
```

**Expected:** Track conditional nature of prop flow
**Actual:** Edge created unconditionally

**Why:** Static analysis doesn't evaluate runtime conditions.

**Fix Required:** Would require data flow analysis; may not be feasible statically.

**Priority:** Low - Acceptable limitation for static analysis

---

## General Limitations

### 9. Re-renders and Performance Detection

**Issue:** The tool performs static analysis only; it cannot detect runtime performance issues or unnecessary re-renders.

**What This Means:**
- Cannot measure actual render frequency
- Cannot detect runtime prop equality checks
- Cannot verify if `React.memo` actually prevents re-renders

**Priority:** Out of scope for static analysis

---

### 10. Context API Detection Gaps

**Issue:** Some Context API patterns may not be recognized as alternatives to prop drilling.

**Current Behavior:** The tool suggests using Context API in violation messages but doesn't verify if Context is already available.

**Fix Required:** Detect existing Context providers and consumers in the codebase.

**Priority:** Low - Recommendations are still valid

---

## Contributing

If you encounter a limitation not listed here, please:
1. Check if it's a bug or expected behavior
2. Create a minimal reproduction test case
3. Open an issue on GitHub with the test case
4. Include expected vs. actual behavior

For limitations listed here, contributions are welcome! See the Priority and Planned sections to understand roadmap priorities.
