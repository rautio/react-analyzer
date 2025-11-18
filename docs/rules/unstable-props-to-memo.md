# unstable-props-to-memo

**Category:** Performance - Broken Memoization
**Severity:** Warning
**Cross-file Analysis:** ✅ Required
**Auto-fixable:** No (requires architectural decisions)

## Summary

Detects when unstable props (inline objects, arrays, or functions) are passed to memoized components, completely breaking memoization and causing unnecessary re-renders. **This rule requires cross-file analysis** to detect violations that ESLint cannot catch.

## Why This Matters

React's memoization features (`React.memo`, `useMemo`, `useCallback`) prevent unnecessary re-renders by comparing props/dependencies by **reference**. When you pass inline objects, arrays, or functions, you create a **new reference every render**, breaking memoization entirely.

### The Cost

Broken memoization means:
- ❌ Memoized component re-renders on **every parent render**
- ❌ Performance is **worse** than without memoization (overhead without benefit)
- ❌ The entire component subtree re-renders unnecessarily
- ❌ Wasted CPU cycles, memory allocations, and DOM updates

## Rule Details

This rule currently detects:

### ✅ Detection 1: React.memo with Unstable Props (Cross-file)

Detects when a parent component passes unstable props to a `React.memo` component defined in another file.

**❌ Violation:**

```tsx
// MemoChild.tsx
export const MemoChild = React.memo(({ config, items, onUpdate }) => {
  return (
    <div>
      {config.theme} - {items.length} items
      <button onClick={onUpdate}>Update</button>
    </div>
  );
});

// Parent.tsx
function Parent() {
  const [count, setCount] = useState(0);

  return (
    <div>
      {/* ❌ VIOLATION: Inline object - new reference every render */}
      <MemoChild config={{ theme: 'dark' }} />

      {/* ❌ VIOLATION: Inline array - new reference every render */}
      <MemoChild items={[1, 2, 3]} />

      {/* ❌ VIOLATION: Inline function - new reference every render */}
      <MemoChild onUpdate={() => console.log('update')} />

      {/* All three violations cause MemoChild to re-render on EVERY Parent render */}
    </div>
  );
}
```

**Why ESLint Can't Detect This:**
- Requires analyzing `MemoChild.tsx` to know it's memoized
- Requires analyzing `Parent.tsx` to find unstable props
- Must follow imports across files
- Needs semantic understanding of `React.memo` wrapping

**✅ Correct Patterns:**

```tsx
// Option 1: Extract to module-level constants
const CONFIG = { theme: 'dark' };
const ITEMS = [1, 2, 3];

function Parent() {
  const handleUpdate = useCallback(() => {
    console.log('update');
  }, []);

  return (
    <div>
      <MemoChild config={CONFIG} />        {/* ✅ Same reference every render */}
      <MemoChild items={ITEMS} />          {/* ✅ Same reference every render */}
      <MemoChild onUpdate={handleUpdate} /> {/* ✅ Stable with useCallback */}
    </div>
  );
}
```

```tsx
// Option 2: Use useMemo for objects/arrays
function Parent() {
  const config = useMemo(() => ({ theme: 'dark' }), []);
  const items = useMemo(() => [1, 2, 3], []);

  return <MemoChild config={config} items={items} />;
}
```

```tsx
// Option 3: If props change based on state, include in dependencies
function Parent() {
  const [theme, setTheme] = useState('dark');

  const config = useMemo(() => ({ theme }), [theme]);

  return <MemoChild config={config} />;
}
```

### ⏳ Detection 2: useMemo/useCallback with Unstable Prop Dependencies

**Status:** Partially implemented, currently disabled

This detection will catch when components use `useMemo` or `useCallback` with prop dependencies that are unstable objects from parent components.

```tsx
// TODO: Not yet implemented
function Child({ config }) {
  // If parent passes inline object to config,
  // this useMemo will re-run every render
  const computed = useMemo(() => {
    return expensiveCalculation(config);
  }, [config]); // Will be detected once implemented

  return <div>{computed}</div>;
}
```

## Detection Algorithm

### Cross-file Analysis Flow

1. **Find JSX elements** in current file (e.g., `<MemoChild />`)
2. **Extract component name** from the element
3. **Check if imported** - look through import statements
4. **Resolve import path** - follow the import to source file
5. **Load target module** - parse and analyze the source file
6. **Check if memoized** - detect `React.memo()` wrapping
7. **Analyze props** - check for inline objects/arrays/functions
8. **Report violations** with precise line numbers

### What Counts as Unstable?

**Unstable (detected):**
- Inline object literals: `{ theme: 'dark' }`
- Inline array literals: `[1, 2, 3]`
- Inline arrow functions: `() => console.log()`
- Inline function expressions: `function() { ... }`

**Stable (allowed):**
- Variables defined outside component: `const CONFIG = { ... }`
- Props wrapped in `useMemo`: `useMemo(() => ({ ... }), [deps])`
- Props wrapped in `useCallback`: `useCallback(() => { ... }, [deps])`
- Primitive values: strings, numbers, booleans, null, undefined
- References to stable variables

## When to Use React.memo

React.memo is beneficial when:

✅ **Use React.memo when:**
- Component is expensive to render
- Component receives the same props frequently
- Component is in a list that re-renders often
- You can provide stable props (or use useMemo/useCallback)

❌ **Don't use React.memo when:**
- Props are always unstable (inline objects/arrays/functions)
- Component is cheap to render
- Props change frequently
- You're not prepared to stabilize props with useMemo/useCallback

## Performance Impact

### Broken Memoization Performance

```tsx
// ❌ Broken memoization is WORSE than no memoization
const MemoChild = React.memo(Child);

function Parent() {
  const [count, setCount] = useState(0);

  return (
    <>
      {/* Every count change causes:
          1. Parent renders
          2. New config object created
          3. React.memo compares props (overhead)
          4. Props are different (new reference)
          5. MemoChild renders anyway
          Result: Memoization overhead + full render = slower than no memo */}
      <MemoChild config={{ theme: 'dark' }} />
    </>
  );
}
```

### Working Memoization Performance

```tsx
// ✅ Working memoization skips child renders
const MemoChild = React.memo(Child);
const CONFIG = { theme: 'dark' };

function Parent() {
  const [count, setCount] = useState(0);

  return (
    <>
      {/* Every count change:
          1. Parent renders
          2. CONFIG reference unchanged
          3. React.memo compares props (overhead)
          4. Props are same (same reference)
          5. MemoChild render skipped ✅
          Result: Only memoization overhead, no child render */}
      <MemoChild config={CONFIG} />
    </>
  );
}
```

## Common Patterns and Solutions

### Pattern 1: Configuration Objects

**❌ Problem:**
```tsx
<Component config={{ apiUrl: '/api', timeout: 5000 }} />
```

**✅ Solutions:**
```tsx
// Option A: Module constant
const API_CONFIG = { apiUrl: '/api', timeout: 5000 };
<Component config={API_CONFIG} />

// Option B: useMemo
const config = useMemo(() => ({ apiUrl: '/api', timeout: 5000 }), []);
<Component config={config} />
```

### Pattern 2: Dynamic Data Arrays

**❌ Problem:**
```tsx
<List items={items.filter(x => x.active)} />
```

**✅ Solution:**
```tsx
const activeItems = useMemo(
  () => items.filter(x => x.active),
  [items]
);
<List items={activeItems} />
```

### Pattern 3: Event Handlers

**❌ Problem:**
```tsx
<Button onClick={() => handleClick(id)} />
```

**✅ Solutions:**
```tsx
// Option A: useCallback with dependency
const handleButtonClick = useCallback(() => {
  handleClick(id);
}, [id]);
<Button onClick={handleButtonClick} />

// Option B: Pass data through props
<Button onClick={handleClick} data={id} />
// Then in Button: onClick={() => props.onClick(props.data)}
```

### Pattern 4: Styling Objects

**❌ Problem:**
```tsx
<Component style={{ padding: 20, margin: 10 }} />
```

**✅ Solutions:**
```tsx
// Option A: Module constant
const STYLE = { padding: 20, margin: 10 };
<Component style={STYLE} />

// Option B: CSS classes (preferred)
<Component className="padded-box" />
```

## Limitations

### Currently Not Detected

- **useMemo/useCallback with unstable prop dependencies** - Partially implemented, disabled
- **Unstable props from state/context** - Only detects inline literals currently
- **Props constructed from functions** - `items={getData()}` not tracked yet
- **Conditional unstable props** - `config={show ? {...} : null}` may not be caught

### By Design

- **Same-file memoization** - Detected (implemented in Issue 3)
- **Inline primitives** - Not flagged (primitives are compared by value)
- **Spread operators** - `{...config}` not detected as spreading may be intentional

## Related Rules

- **no-object-deps** - Detects unstable objects in hook dependency arrays
- **no-inline-props** - Would detect ALL inline props (not just to memo components)

## Examples

### Real-world Example: Data Table

**❌ Before (broken memoization):**
```tsx
// Table.tsx
export const Table = React.memo(({ columns, data, onSort }) => {
  // Expensive rendering logic
  return <table>...</table>;
});

// Dashboard.tsx
function Dashboard() {
  const [sortKey, setSortKey] = useState('name');

  return (
    <Table
      columns={[
        { key: 'name', label: 'Name' },
        { key: 'email', label: 'Email' }
      ]}
      data={userData}
      onSort={(key) => setSortKey(key)}
    />
  );
  // Table re-renders on EVERY Dashboard render
  // Memoization is completely broken
}
```

**✅ After (working memoization):**
```tsx
// Table.tsx - unchanged
export const Table = React.memo(({ columns, data, onSort }) => {
  return <table>...</table>;
});

// Dashboard.tsx
const COLUMNS = [
  { key: 'name', label: 'Name' },
  { key: 'email', label: 'Email' }
];

function Dashboard() {
  const [sortKey, setSortKey] = useState('name');

  const handleSort = useCallback((key) => {
    setSortKey(key);
  }, []);

  return (
    <Table
      columns={COLUMNS}
      data={userData}
      onSort={handleSort}
    />
  );
  // Table only re-renders when userData changes
  // Memoization working perfectly
}
```

**Performance Improvement:**
- Before: Table re-renders ~10-20 times per second during interactions
- After: Table re-renders only when data actually changes
- Result: **Smooth 60fps** instead of janky interactions

## Configuration

This rule has no configuration options. It always checks for unstable props to memoized components.

## When to Disable

You might want to disable this rule if:
- You're not using `React.memo` in your codebase
- You prefer a different memoization strategy (e.g., Recoil selectors)
- You're in a code section where performance is not a concern

## Further Reading

- [React Docs: React.memo](https://react.dev/reference/react/memo)
- [React Docs: useMemo](https://react.dev/reference/react/useMemo)
- [React Docs: useCallback](https://react.dev/reference/react/useCallback)
- [Before You memo()](https://overreacted.io/before-you-memo/)
- [Why React Re-Renders](https://www.joshwcomeau.com/react/why-react-re-renders/)

## Technical Notes

### Cross-file Analysis Implementation

This rule uses the `ModuleResolver` to:
1. Parse and cache imported modules
2. Follow import statements across files
3. Detect `React.memo` wrapping in target modules
4. Track aliased imports correctly (e.g., `import { MemoChild as FastChild }`)

### AST Patterns Detected

- `jsx_element` and `jsx_self_closing_element` for component usage
- `call_expression` with `React.memo` or `memo` function
- `object` and `array` node types for inline literals
- `arrow_function` and `function` for inline functions
- `member_expression` for import resolution with aliases
