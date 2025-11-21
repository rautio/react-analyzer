# deep-prop-drilling

**Detect when props are passed through multiple component levels without being used**

## Rule Details

This rule identifies "prop drilling" - a common React anti-pattern where props are passed through multiple intermediate components that don't actually use them. These intermediate components exist only to relay props from a parent to a deeply nested child.

By default, this rule flags chains where props pass through **2 or more intermediate components** (a depth of 3+ total components).

### Why This Matters

Prop drilling creates several problems:

1. **Tight Coupling**: Every component in the chain becomes coupled to props it doesn't use
2. **Maintenance Burden**: Adding/removing/renaming props requires changes across multiple files
3. **Poor Readability**: Component signatures become cluttered with "passthrough" props
4. **Brittle Refactoring**: Moving components requires threading props through new intermediaries
5. **Performance**: Unnecessary re-renders when props change, even in components that don't use them

### How It Works

The rule analyzes your component graph to trace prop chains:

```
App (defines state)
  ↓ passes 'user'
Parent (doesn't use 'user', just passes it)
  ↓ passes 'user'
Child (doesn't use 'user', just passes it)
  ↓ passes 'user'
Display (finally uses 'user')
```

In this example:
- **Depth**: 4 components
- **Passthrough Components**: 2 (Parent, Child)
- **Violation**: Yes (≥2 passthroughs)

## ❌ Incorrect Code

### Example 1: Simple prop drilling

```tsx
// ❌ BAD: 'theme' is drilled through multiple levels
function App() {
  const [theme, setTheme] = useState('dark');

  return (
    <Dashboard theme={theme} />
  );
}

function Dashboard({ theme }: { theme: string }) {
  // Dashboard doesn't use 'theme', just passes it along
  return (
    <div>
      <Sidebar theme={theme} />
    </div>
  );
}

function Sidebar({ theme }: { theme: string }) {
  // Sidebar doesn't use 'theme' either, just passes it along
  return (
    <nav>
      <ThemeToggle theme={theme} />
    </nav>
  );
}

function ThemeToggle({ theme }: { theme: string }) {
  // Finally! Someone actually uses the prop
  return <button className={theme}>Toggle</button>;
}
```

**Problems:**
- `theme` prop appears in 4 component signatures
- Dashboard and Sidebar are coupled to a prop they don't use
- Refactoring ThemeToggle or Sidebar requires changes in Dashboard
- All 3 components re-render when `theme` changes

---

### Example 2: Multiple drilled props

```tsx
// ❌ BAD: Multiple props drilled through the same chain
function App() {
  const [user, setUser] = useState({ name: 'Alice', id: 123 });
  const [settings, setSettings] = useState({ fontSize: 14 });

  return <Layout user={user} settings={settings} />;
}

function Layout({ user, settings }: Props) {
  // Doesn't use user or settings, just arranges children
  return (
    <div>
      <Header user={user} settings={settings} />
      <Content user={user} settings={settings} />
    </div>
  );
}

function Header({ user, settings }: Props) {
  // Still just passing through
  return (
    <header>
      <UserProfile user={user} settings={settings} />
    </header>
  );
}

function UserProfile({ user, settings }: Props) {
  // Finally used here
  return (
    <div style={{ fontSize: settings.fontSize }}>
      Hello, {user.name}!
    </div>
  );
}
```

**Problems:**
- Each intermediate component must accept and forward 2 props
- Adding a 3rd prop (e.g., `theme`) requires changes in 3 files
- Type changes to `user` require updates in 4 places

---

### Example 3: Cross-file prop drilling

```tsx
// ❌ BAD: Props drilled across file boundaries
// App.tsx
export function App() {
  const apiClient = useApiClient();
  return <Dashboard apiClient={apiClient} />;
}

// Dashboard.tsx
export function Dashboard({ apiClient }: Props) {
  return <UserPanel apiClient={apiClient} />;
}

// UserPanel.tsx
export function UserPanel({ apiClient }: Props) {
  return <ProfileCard apiClient={apiClient} />;
}

// ProfileCard.tsx
export function ProfileCard({ apiClient }: Props) {
  const { data } = apiClient.getUser();
  return <div>{data.name}</div>;
}
```

**Problems:**
- 4 files all import and define types for `apiClient`
- If API client interface changes, 4 files need updates
- Testing each component requires mocking `apiClient` even when unused

---

## ✅ Correct Code

### Solution 1: React Context (Recommended)

**Best for:** App-wide state like theme, user, auth

```tsx
// ✅ GOOD: Use Context to provide values deep in the tree
const ThemeContext = createContext('light');

function App() {
  const [theme, setTheme] = useState('dark');

  return (
    <ThemeContext.Provider value={theme}>
      <Dashboard />
    </ThemeContext.Provider>
  );
}

function Dashboard() {
  // No theme prop needed!
  return (
    <div>
      <Sidebar />
    </div>
  );
}

function Sidebar() {
  // Still no theme prop!
  return (
    <nav>
      <ThemeToggle />
    </nav>
  );
}

function ThemeToggle() {
  // Consume directly where needed
  const theme = useContext(ThemeContext);
  return <button className={theme}>Toggle</button>;
}
```

**Why it works:**
- Intermediate components don't know about `theme`
- Adding new theme consumers is trivial
- Refactoring is safer - no prop threading needed

---

### Solution 2: Component Composition

**Best for:** Layout components, flexible UI structure

```tsx
// ✅ GOOD: Pass components as children instead of props
function App() {
  const [user, setUser] = useState({ name: 'Alice' });

  return (
    <Layout>
      <Header>
        <UserProfile user={user} />
      </Header>
    </Layout>
  );
}

function Layout({ children }: { children: React.ReactNode }) {
  // Layout doesn't need to know about user prop
  return <div className="layout">{children}</div>;
}

function Header({ children }: { children: React.ReactNode }) {
  // Header doesn't need to know about user prop
  return <header>{children}</header>;
}

function UserProfile({ user }: { user: User }) {
  return <div>Hello, {user.name}!</div>;
}
```

**Why it works:**
- Layout and Header are decoupled from `user` prop
- More flexible: easy to change what renders in the header
- Follows "Inversion of Control" principle

---

### Solution 3: State Management Library

**Best for:** Complex apps, shared state, multiple prop drilling paths

```tsx
// ✅ GOOD: Use Zustand/Redux/Jotai for global state
import { create } from 'zustand';

// Define store
const useThemeStore = create((set) => ({
  theme: 'dark',
  setTheme: (theme: string) => set({ theme })
}));

function App() {
  return <Dashboard />;
}

function Dashboard() {
  // No props at all!
  return <div><Sidebar /></div>;
}

function Sidebar() {
  return <nav><ThemeToggle /></nav>;
}

function ThemeToggle() {
  // Access state directly from store
  const theme = useThemeStore((state) => state.theme);
  return <button className={theme}>Toggle</button>;
}
```

**Why it works:**
- Zero prop drilling
- State accessible anywhere in the tree
- Better DevTools and debugging

---

### Solution 4: Flatten Component Hierarchy

**Best for:** Over-engineered component trees

```tsx
// ❌ BEFORE: Too many levels
function App() {
  const [count, setCount] = useState(0);
  return <Layout count={count} />;
}

function Layout({ count }: Props) {
  return <Container count={count} />;
}

function Container({ count }: Props) {
  return <Display count={count} />;
}

// ✅ AFTER: Flattened
function App() {
  const [count, setCount] = useState(0);
  return <Display count={count} />;
}
```

**Why it works:**
- Eliminates unnecessary abstraction
- Direct prop passing - no drilling
- Sometimes you just don't need those intermediate components

---

## Configuration

You can configure the maximum allowed depth:

```json
{
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "maxDepth": 3
    }
  }
}
```

**Options:**
- `maxDepth` (default: `3`): Maximum component chain depth before flagging
  - `3` = Allow chains like `A → B → C` (1 passthrough)
  - `4` = Allow chains like `A → B → C → D` (2 passthroughs)
  - `2` = Only allow direct passing `A → B` (0 passthroughs)

---

## Decision Tree: How to Fix

```
Is this app-wide state (theme, auth, i18n)?
├─ YES → Use React Context
│
└─ NO → Is it data fetching/API clients?
    ├─ YES → Use React Query/SWR with Context
    │         or dependency injection
    │
    └─ NO → Is it UI composition/layout?
        ├─ YES → Use component composition
        │         (children/render props)
        │
        └─ NO → Do multiple components need it?
            ├─ YES → Consider state management
            │         (Zustand/Redux/Jotai)
            │
            └─ NO → Can you flatten the hierarchy?
                ├─ YES → Remove unnecessary wrappers
                └─ NO → Accept the drilling (it's rare)
```

---

## Common Patterns

### Pattern 1: The "Page Context" pattern

```tsx
// For page-level data that many components need
const DashboardContext = createContext(null);

function DashboardPage() {
  const data = useDashboardData();

  return (
    <DashboardContext.Provider value={data}>
      <DashboardLayout />
    </DashboardContext.Provider>
  );
}

// Any descendant can access it
function SomeNestedWidget() {
  const data = useContext(DashboardContext);
  return <div>{data.stats}</div>;
}
```

### Pattern 2: The "Compound Component" pattern

```tsx
// For complex components with shared state
function Tabs({ children }: Props) {
  const [activeTab, setActiveTab] = useState(0);

  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      {children}
    </TabsContext.Provider>
  );
}

Tabs.TabList = function TabList({ children }: Props) {
  return <div role="tablist">{children}</div>;
};

Tabs.Tab = function Tab({ index, children }: Props) {
  const { activeTab, setActiveTab } = useContext(TabsContext);
  return (
    <button
      onClick={() => setActiveTab(index)}
      aria-selected={activeTab === index}
    >
      {children}
    </button>
  );
};

// Usage: No prop drilling!
<Tabs>
  <Tabs.TabList>
    <Tabs.Tab index={0}>First</Tabs.Tab>
    <Tabs.Tab index={1}>Second</Tabs.Tab>
  </Tabs.TabList>
  <Tabs.Panel index={0}>Content 1</Tabs.Panel>
  <Tabs.Panel index={1}>Content 2</Tabs.Panel>
</Tabs>
```

---

## Performance Impact

**Severity:** 🟡 **Moderate to High**

Prop drilling itself isn't slow, but it causes:

1. **Unnecessary Re-renders**: All intermediate components re-render when props change
   - If Parent, Child don't use `count`, they shouldn't re-render when it changes
   - Context or memo can prevent this

2. **Maintenance Overhead**: More time spent refactoring
   - Studies show 30-40% of refactoring time is prop threading
   - Context reduces this significantly

3. **Bundle Size**: More prop definitions = more TypeScript types
   - Marginal, but adds up in large apps

**Real-world example:**
```tsx
// Before: 3 components re-render on every user change
App → Layout → Header → UserMenu

// After: Only UserMenu re-renders
App → Layout → Header → UserMenu (via Context)
```

---

## When to Disable

Prop drilling isn't always bad. Consider disabling for:

### 1. Direct parent-child relationships
```tsx
// This is fine - it's just one level
function TodoList({ items }: Props) {
  return items.map(item => <TodoItem item={item} />);
}
```

### 2. Props that truly belong at each level
```tsx
// If intermediate components DO use the prop, it's not drilling
function App() {
  const theme = useTheme();
  return <Layout theme={theme} />; // Layout uses theme
}

function Layout({ theme }: Props) {
  return (
    <div className={theme}> {/* Uses it! */}
      <Header theme={theme} /> {/* Header also uses it */}
    </div>
  );
}
```

### 3. Very stable, simple apps
If you have a small app (< 10 components) and props rarely change, the cost of Context might outweigh the benefits.

To disable:
```tsx
// react-analyzer-disable-next-line deep-prop-drilling
<Component prop={value} />
```

Or in config:
```json
{
  "rules": {
    "deep-prop-drilling": {
      "enabled": false
    }
  }
}
```

---

## Migration Guide

### Step 1: Identify the most drilled props
Run the analyzer and sort by depth. Fix the worst offenders first.

### Step 2: Choose your solution
- **App-wide?** → Context
- **Page-specific?** → Page Context or composition
- **Complex shared state?** → State management library

### Step 3: Refactor incrementally
```tsx
// Week 1: Add Context provider (doesn't break anything)
<ThemeContext.Provider value={theme}>
  <Dashboard theme={theme} /> {/* Keep prop for now */}
</ThemeContext.Provider>

// Week 2: Update consumers to use Context
function ThemeToggle() {
  const theme = useContext(ThemeContext); // New!
  // const theme = props.theme; // Old - remove later
}

// Week 3: Remove prop from intermediate components
function Dashboard() {
  return <Sidebar />; // No theme prop!
}
```

---

## References

- [React Docs: Context](https://react.dev/learn/passing-data-deeply-with-context)
- [Kent C. Dodds: Prop Drilling](https://kentcdodds.com/blog/prop-drilling)
- [React Patterns: Compound Components](https://www.patterns.dev/posts/compound-pattern/)
- [Zustand](https://github.com/pmndrs/zustand)
- [Jotai](https://jotai.org/)

---

## Related Rules

- **unstable-props-to-memo** - Detects unstable props passed to memoized components
- **no-inline-props** - Flags inline object/array/function props

---

**Rule ID:** `deep-prop-drilling`
**Category:** Architecture / Maintainability
**Severity:** Warning
**Auto-fixable:** No (requires architectural changes)
