# Known Limitations

**Last Updated:** 2025-11-18 (Arrow Functions + Cross-File ✅ COMPLETED)
**See also:** [ROADMAP.md](ROADMAP.md) for planned fixes

This document outlines current limitations of the react-analyzer tool, organized by priority and impact.

---

## Quick Reference

| Priority | Limitation | Target Phase | ETA |
|----------|-----------|--------------|-----|
| ✅ COMPLETED | [Arrow Function Components](#1-arrow-function-components) | Phase 2.2 | ✅ DONE |
| ✅ COMPLETED | [Cross-File Prop Drilling](#2-cross-file-prop-drilling) | Phase 2.2 | ✅ DONE |
| 🟡 PARTIAL | [Prop Spread Operators](#3-prop-spread-operators) | Phase 2.3 | Q1-Q2 2026 |
| 🟠 HIGH | [Object Property Access](#4-object-property-access) | Phase 2.3 | Q1-Q2 2026 |
| 🟡 MEDIUM | [Partial Prop Usage Detection](#5-partial-prop-usage-detection) | Phase 2.3 | Q1-Q2 2026 |
| 🟡 MEDIUM | [Renamed Props](#6-renamed-props) | Phase 2.4 | Q2 2026 |
| 🟢 LOW | [Dynamic Prop Names](#7-dynamic-prop-names) | TBD | TBD |
| 🟢 LOW | [Conditional Prop Passing](#8-conditional-prop-passing) | N/A | N/A |
| ⚪ OUT OF SCOPE | [Runtime Performance](#9-runtime-performance-detection) | N/A | N/A |
| ⚪ OUT OF SCOPE | [Context API Integration](#10-context-api-integration) | Phase 2.4 | Q2 2026 |

---

## 🔴 CRITICAL Limitations

These gaps block real-world adoption and are highest priority for fixes.

### 1. Arrow Function Components

**Status:** ✅ COMPLETED (Phase 2.2 Priority 1A)
**Completed:** 2025-11-18
**Code Location:** `internal/graph/builder.go`

#### ~~Issue~~ FIXED
Components defined as arrow functions are now fully supported.

#### What Now Works
All arrow function component patterns are detected:
```tsx
// ✅ ALL DETECTED NOW
const MyComponent = ({ theme }: Props) => {
    return <div className={theme}>Content</div>;
};

const MemoComponent = memo(({ theme }: Props) => {
    return <div className={theme}>Content</div>;
});

export const ExportedComponent = () => <div />;

function MyComponent({ theme }: Props) {
    return <div className={theme}>Content</div>;
}
```

#### Implementation Details
- Added `getComponentNameFromArrowFunction()` to detect arrow components from `variable_declarator` nodes
- Added `extractPropsFromArrowFunction()` to extract props from arrow function parameters
- Updated all 4 build phases to handle arrow functions using stack-based component tracking
- Supports both direct arrow functions and `React.memo()` wrapped arrow functions
- Test fixtures added: `ArrowFunctionDrilling.tsx`, `MemoArrowFunctionDrilling.tsx`, `MixedArrowAndFunction.tsx`

#### Workaround
~~No workaround needed - arrow functions are fully supported!~~

---

### 2. Cross-File Prop Drilling

**Status:** ✅ COMPLETED (Phase 2.2 Priority 1B)
**Completed:** 2025-11-18
**Code Location:** `internal/graph/builder.go`

#### ~~Issue~~ FIXED
Prop drilling detection now works across multiple files!

#### What Now Works
All cross-file prop drilling patterns are detected:
```tsx
// App.tsx - ✅ ALL WORK NOW
import { Dashboard } from './Dashboard';

function App() {
    const [theme, setTheme] = useState('dark');
    return <Dashboard theme={theme} />;  // ✅ Edge created
}

// Dashboard.tsx - ✅ NOW WORKS
export function Dashboard({ theme }: Props) {
    return <Sidebar theme={theme} />;  // ✅ Edge created (cross-file!)
}

// Sidebar.tsx - ✅ NOW WORKS
export function Sidebar({ theme }: Props) {
    return <ThemeToggle theme={theme} />;  // ✅ Edge created (cross-file!)
}
```

**Result:** ✅ Detects complete drilling chain across all files: App → Dashboard → Sidebar → ThemeToggle

#### Implementation Details
- Added `findComponentInImports()` to search for components in imported files
- Added `findComponentInFile()` helper to search by component name in specific file
- Uses ModuleResolver to resolve import paths and get imported modules
- Handles both default exports (`import Dashboard from './Dashboard'`)
- Handles named exports (`import { Dashboard } from './Dashboard'`)
- Handles import aliases (`import { Dashboard as DashPanel } from './Dashboard'`)
- Test fixtures added: 3-file and 4-file deep drilling scenarios

#### Workaround
~~No workaround needed - cross-file prop drilling is fully supported!~~

---

## 🟠 HIGH Priority Limitations

Common patterns that significantly limit the tool's usefulness.

### 3. Prop Spread Operators

**Status:** 🟡 PARTIAL - Infrastructure implemented (Phase 2.2C)
**Completed:** Basic detection (2025-11-18)
**Target:** Full support in Phase 2.3
**Code Location:** `internal/graph/builder.go:339-428`

#### What Works (Partial Support)
Basic spread infrastructure is implemented:
```tsx
function Wrapper(props: any) {
    return <Child {...props} />;  // ✅ Spread detected
}
```

**Implemented:**
- ✅ Detects `jsx_spread_attribute` nodes
- ✅ Creates edges for identified spread variables
- ✅ Handles spreading parent's props object
- ✅ Handles spreading parent's state variables

#### What Doesn't Work Yet
Full tracking through non-destructured props:
```tsx
function Container(props: Config) {  // ❌ Props not destructured
    return <Panel {...props} />;     // ❌ Can't infer individual props
}

function Panel(props: Config) {
    return <Settings {...props} />;
}

function Settings({ apiUrl, timeout }: Config) {
    return <div>{apiUrl}</div>;
}
```

**Result:** Can't track `apiUrl` and `timeout` through the chain without TypeScript type analysis.

#### Why Full Support is Complex
Without destructuring, we can't determine which individual props exist in `props: Config`. This requires:
- TypeScript AST type resolution
- Object property tracking
- Advanced data flow analysis

#### Workaround
**Use destructured props:**
```tsx
// ✅ This pattern works
function Container({ apiUrl, timeout }: Config) {
    return <Panel apiUrl={apiUrl} timeout={timeout} />;
}

// Or pass explicitly:
<Panel apiUrl={props.apiUrl} timeout={props.timeout} />
```

#### Next Steps
Full spread support moved to Phase 2.3, combined with object property access detection (similar challenges).

---

### 4. Object Property Access

**Status:** 🟠 HIGH - Common TypeScript pattern
**Target:** Phase 2.3 (Priority 2B)
**ETA:** Q1-Q2 2026 (2 weeks)
**Test Fixture:** `test/fixtures/prop-drilling/Dashboard.tsx` (expects 3 violations, gets 0)

#### Issue
Props accessed as object properties are not tracked as separate prop flows.

#### Current Behavior
When state is stored as an object and individual properties are passed, only the object-level state is tracked, not the individual properties.

#### Example
```tsx
function App() {
    const [settings, setSettings] = useState({
        locale: 'en-US',
        currency: 'USD',
        dateFormat: 'MM/DD/YYYY'
    });

    return <Layout
        locale={settings.locale}      // ❌ Not tracked
        currency={settings.currency}  // ❌ Not tracked
    />;
}
```

**Result:** Only tracks `settings` as single state node; doesn't detect `locale` and `currency` as separate prop flows.

#### Why
1. Edge detection looks for identifier nodes (`{theme}`), not member expressions (`{settings.locale}`)
2. State extraction creates one node per `useState` call, not per property
3. No virtual state nodes for object properties

#### Workaround
**Destructure state into individual variables:**
```tsx
function App() {
    const [settings, setSettings] = useState({ ... });
    const { locale, currency, dateFormat } = settings;

    return <Layout locale={locale} currency={currency} />;  // ✅ Now tracked
}
```

**Impact:** Less idiomatic; creates extra variables.

#### Planned Fix
1. Detect `member_expression` nodes when creating edges
2. Either create virtual state nodes for each property, OR track property access paths in edges
3. Update prop matching to handle property chains
4. Handle deep paths: `config.settings.theme.primary`

**See:** [ROADMAP.md - Phase 2.3, Priority 2B](ROADMAP.md#priority-2b-object-property-access-tracking-2-weeks)

---

## 🟡 MEDIUM Priority Limitations

Impact accuracy but less common or less critical.

### 5. Partial Prop Usage Detection

**Status:** 🟡 MEDIUM - Causes false positives
**Target:** Phase 2.3 (Priority 2A)
**ETA:** Q1-Q2 2026 (1-2 weeks)
**Test Fixture:** `test/fixtures/prop-drilling/PartialUsage.tsx` (currently fails)

#### Issue
The tool cannot detect when a component both uses a prop locally AND passes it to children.

#### Current Behavior
If a component passes a prop to a child, it's assumed to be a pure passthrough, even if the component also uses the prop.

#### Example
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

function Child({ theme }: { theme: string }) {
    return <Display theme={theme} />;
}

function Display({ theme }: { theme: string }) {
    return <div className={theme}>Content</div>;
}
```

**Expected:** No violation (Parent uses theme, so not a pure passthrough)
**Actual:** ❌ Violation reported for theme drilled through 3 levels

#### Why
The `componentUsesProp()` heuristic only checks if a prop is passed to children, not whether it's also referenced in the component's code.

#### Workaround
**None currently.** This is a false positive that users must ignore.

#### Planned Fix
AST analysis to detect prop references in:
- JSX attributes and content (`<div>{theme}</div>`)
- Variable declarations (`const x = theme`)
- Function calls (`doSomething(theme)`)
- Hook dependencies (`useEffect(() => {}, [theme])`)
- Conditional statements (`if (theme === 'dark')`)

**See:** [ROADMAP.md - Phase 2.3, Priority 2A](ROADMAP.md#priority-2a-partial-prop-usage-detection-1-2-weeks)

---

### 6. Renamed Props

**Status:** 🟡 MEDIUM - Less common pattern
**Target:** Phase 2.4
**ETA:** Q2 2026 (1 week)

#### Issue
Props renamed during passing are not tracked correctly.

#### Current Behavior
Only tracks props with the same name throughout the chain.

#### Example
```tsx
function Parent({ appTheme }: Props) {
    return <Child theme={appTheme} />;  // ❌ Renamed from appTheme to theme
}

function Child({ theme }: Props) {
    return <Display theme={theme} />;
}
```

**Result:** No edge created from Parent to Child (name mismatch).

#### Why
The `isParentVariable()` check matches variable names exactly without considering renames.

#### Workaround
**Use consistent prop names:**
```tsx
function Parent({ theme }: Props) {  // Renamed to 'theme'
    return <Child theme={theme} />;  // ✅ Now matches
}
```

**Impact:** Minor; just requires naming consistency.

#### Planned Fix
1. Track prop name mappings in edges (e.g., "appTheme -> theme")
2. Update path tracing to follow name changes
3. Report original prop name in violations

**See:** [ROADMAP.md - Phase 2.4](ROADMAP.md#features)

---

## 🟢 LOW Priority Limitations

Edge cases or anti-patterns that are less critical.

### 7. Dynamic Prop Names

**Status:** 🟢 LOW - Uncommon, often an anti-pattern
**Target:** TBD
**ETA:** TBD

#### Issue
Props with computed/dynamic names are not detected.

#### Example
```tsx
const propName = 'theme';
return <Child {...{ [propName]: value }} />;
```

**Result:** Not tracked.

#### Why
Edge detection only handles static property identifiers.

#### Workaround
**Use static prop names:**
```tsx
return <Child theme={value} />;
```

**Impact:** Minimal; computed prop names are rare in practice.

#### Fix Complexity
Would require constant propagation and expression evaluation, which is complex and potentially unsound.

---

### 8. Conditional Prop Passing

**Status:** 🟢 LOW - Acceptable static analysis limitation
**Target:** N/A (Unlikely to fix)
**ETA:** N/A

#### Issue
Props passed conditionally are always tracked as if they're passed.

#### Example
```tsx
function Parent({ theme, showChild }: Props) {
    return showChild ? <Child theme={theme} /> : null;
}
```

**Expected:** Track conditional nature of prop flow
**Actual:** Edge created unconditionally

#### Why
Static analysis doesn't evaluate runtime conditions.

#### Workaround
None needed; this is an acceptable limitation.

#### Fix Complexity
Would require data flow analysis and may not be feasible with static analysis alone.

---

## ⚪ OUT OF SCOPE Limitations

Inherent to static analysis or intentionally deferred.

### 9. Runtime Performance Detection

**Status:** ⚪ OUT OF SCOPE - Beyond static analysis capabilities
**Target:** N/A
**ETA:** N/A

#### What's Not Possible
- Cannot measure actual render frequency
- Cannot detect runtime prop equality checks
- Cannot verify if `React.memo` actually prevents re-renders
- Cannot measure bundle size impact

#### Why
These require runtime profiling, not static analysis.

#### Alternatives
Use browser DevTools or React Profiler for runtime performance analysis.

---

### 10. Context API Integration

**Status:** 🟡 MEDIUM - Improves UX but not critical
**Target:** Phase 2.4
**ETA:** Q2 2026 (2 weeks)

#### Issue
The tool suggests using Context API in violation messages but doesn't verify if Context is already available.

#### Current Behavior
All recommendations suggest creating a new Context, even if one exists.

#### Planned Fix
1. Detect existing Context providers in the codebase
2. Detect existing Context consumers
3. Suggest using existing Context when available
4. Only suggest creating new Context when none exists

**See:** [ROADMAP.md - Phase 2.4](ROADMAP.md#features)

---

## Contributing

If you encounter a limitation not listed here:

1. **Check if it's a bug or expected behavior**
   - Review this document
   - Search existing GitHub issues

2. **Create a minimal reproduction test case**
   - Single file or small directory
   - Minimal code that demonstrates the issue

3. **Open an issue on GitHub**
   - Include the test case
   - Describe expected vs. actual behavior
   - Link to this document if related

4. **Want to fix it yourself?**
   - See [CONTRIBUTING.md](CONTRIBUTING.md) for how to add features
   - Check [ROADMAP.md](ROADMAP.md) to see if it's already planned
   - Open a discussion to coordinate work

---

**Last updated:** 2025-11-18
**Next review:** 2026-01 (monthly updates)
