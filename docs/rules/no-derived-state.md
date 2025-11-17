# no-derived-state

**Category:** Performance - Unnecessary Re-renders
**Severity:** Warning
**Auto-fixable:** Potential (suggest useMemo)

## Summary

Detects when `useState` is used to mirror or derive a value from props, causing unnecessary re-renders and complexity. This is one of the most common React anti-patterns and directly violates React's "Don't mirror props in state" principle.

## The Problem

When you store a derived value in state and sync it via `useEffect`, you create multiple problems:

### 1. **Extra Re-render** (Performance Issue)
The component renders twice when the prop changes:
- **Render 1:** With stale state (before useEffect runs)
- **Render 2:** After useEffect sets the new state

### 2. **Stale State During First Render** (Correctness Issue)
```tsx
function UserCard({ user }) {
  const [name, setName] = useState(user.name);

  useEffect(() => {
    setName(user.name); // Runs AFTER render
  }, [user.name]);

  // ⚠️ First render shows OLD name if user prop changes!
  return <div>{name}</div>;
}
```

### 3. **Increased Complexity** (Maintainability)
- More code to write and maintain
- More state to track and debug
- Easy to forget to sync all derived values
- Harder to reason about data flow

### 4. **Memory Overhead** (Resource Usage)
Each state variable consumes memory and increases React's workload.

## Rule Details

This rule detects the following anti-pattern:

**❌ Bad Pattern:**
```tsx
function Component({ user, isAdmin }) {
  // Mirroring props to state
  const [userName, setUserName] = useState(user.name);
  const [admin, setAdmin] = useState(isAdmin);

  // Syncing state when props change
  useEffect(() => {
    setUserName(user.name);
  }, [user.name]);

  useEffect(() => {
    setAdmin(isAdmin);
  }, [isAdmin]);

  return <div>{userName} - {admin ? 'Admin' : 'User'}</div>;
}
```

**Problems:**
- Two extra re-renders every time `user` or `isAdmin` changes
- Stale values during first render after prop change
- 10+ lines of code for something that should be 1 line

**✅ Good Pattern:**
```tsx
function Component({ user, isAdmin }) {
  // Derive during render - no state, no effects, no extra renders!
  const userName = user.name;
  const admin = isAdmin;

  return <div>{userName} - {admin ? 'Admin' : 'User'}</div>;
}

// Or even simpler:
function Component({ user, isAdmin }) {
  return <div>{user.name} - {isAdmin ? 'Admin' : 'User'}</div>;
}
```

**Benefits:**
- Zero extra re-renders
- Always shows current value
- 1 line instead of 10
- Easier to understand and maintain

## Examples

### Example 1: Simple Prop Mirroring

❌ **Incorrect:**
```tsx
function ProductCard({ product }) {
  const [title, setTitle] = useState(product.title);
  const [price, setPrice] = useState(product.price);

  useEffect(() => {
    setTitle(product.title);
    setPrice(product.price);
  }, [product.title, product.price]);

  return (
    <div>
      <h2>{title}</h2>
      <p>${price}</p>
    </div>
  );
}
```

✅ **Correct:**
```tsx
function ProductCard({ product }) {
  // Just use the props directly!
  return (
    <div>
      <h2>{product.title}</h2>
      <p>${product.price}</p>
    </div>
  );
}
```

### Example 2: Derived Computation

❌ **Incorrect:**
```tsx
function OrderSummary({ items }) {
  const [total, setTotal] = useState(calculateTotal(items));

  useEffect(() => {
    setTotal(calculateTotal(items));
  }, [items]);

  return <div>Total: ${total}</div>;
}
```

✅ **Correct (Simple):**
```tsx
function OrderSummary({ items }) {
  // Derive during render
  const total = calculateTotal(items);
  return <div>Total: ${total}</div>;
}
```

✅ **Correct (Expensive Computation):**
```tsx
function OrderSummary({ items }) {
  // Use useMemo for expensive calculations
  const total = useMemo(() => calculateTotal(items), [items]);
  return <div>Total: ${total}</div>;
}
```

### Example 3: Formatted Value

❌ **Incorrect:**
```tsx
function DateDisplay({ timestamp }) {
  const [formatted, setFormatted] = useState(formatDate(timestamp));

  useEffect(() => {
    setFormatted(formatDate(timestamp));
  }, [timestamp]);

  return <div>{formatted}</div>;
}
```

✅ **Correct:**
```tsx
function DateDisplay({ timestamp }) {
  const formatted = formatDate(timestamp); // Or useMemo if expensive
  return <div>{formatted}</div>;
}
```

## When State IS Appropriate

There are valid cases for initializing state from props:

### ✅ Valid: Seed Value (Use Once, Then Independent)

```tsx
function ColorPicker({ initialColor }) {
  // User can change color independently - state is appropriate!
  const [color, setColor] = useState(initialColor);

  // No useEffect needed - we only use initialColor ONCE

  return (
    <input
      type="color"
      value={color}
      onChange={(e) => setColor(e.target.value)}
    />
  );
}
```

**Why this is OK:**
- `initialColor` is used only for the initial value
- User can modify `color` independently
- No useEffect syncing prop changes
- State represents user input, not derived data

### ✅ Valid: Resettable State with Key

```tsx
function EditableField({ value, key }) {
  const [draft, setDraft] = useState(value);

  // Reset when key changes (e.g., different item selected)
  useEffect(() => {
    setDraft(value);
  }, [key, value]); // OK because we're resetting, not syncing

  return (
    <input
      value={draft}
      onChange={(e) => setDraft(e.target.value)}
    />
  );
}
```

**Why this is OK:**
- Draft is user input, not just derived
- Effect resets on explicit signal (key change)
- State is needed for editing experience

## Detection Strategy

The rule detects this pattern:

1. **useState** initialized with a value derived from props
2. **useEffect** that updates that state when the prop changes
3. **No user interaction** that modifies the state independently

### Detected Pattern:
```tsx
const [stateVar, setStateVar] = useState(prop); // or useState(derive(prop))

useEffect(() => {
  setStateVar(prop); // or setStateVar(derive(prop))
}, [prop]);
```

### Heuristics to Avoid False Positives:

✅ **Allow** if:
- State is modified in event handlers (onClick, onChange, etc.)
- Effect does more than just set state (has other side effects)
- Component uses the state setter in callbacks passed to children

❌ **Report** if:
- State is ONLY set in useEffect
- Effect ONLY sets the state (no other side effects)
- The derived value could be computed during render

## Configuration

```json
{
  "rules": {
    "no-derived-state": "warn"
  }
}
```

## Recommended Fixes

### Fix 1: Direct Derivation (Simplest)
```tsx
// Before
const [value, setValue] = useState(prop);
useEffect(() => setValue(prop), [prop]);

// After
const value = prop;
```

### Fix 2: useMemo (If Expensive)
```tsx
// Before
const [computed, setComputed] = useState(expensiveCalc(data));
useEffect(() => setComputed(expensiveCalc(data)), [data]);

// After
const computed = useMemo(() => expensiveCalc(data), [data]);
```

### Fix 3: Keep State (If Interactive)
```tsx
// If state is modified by user interaction, keep it but document why:
const [editable, setEditable] = useState(initialValue);
// Note: State is needed for user editing

useEffect(() => {
  setEditable(initialValue); // Reset on initialValue change
}, [initialValue]);
```

## Performance Impact

### Measured Impact:
- **Extra renders:** 1 additional render per prop change
- **Render time:** 0-50ms per extra render (depends on component complexity)
- **At scale:** 10 instances × 10 prop changes/sec = 100 unnecessary renders/sec
- **Memory:** ~100 bytes per state variable

### Real-World Example:
A list of 100 items, each unnecessarily deriving 2 values from props:
- **200 extra state variables** (200 × ~100 bytes = 20KB)
- **200 extra re-renders** when parent updates
- **Combined delay:** 200 × 5ms = 1000ms (1 second) of wasted render time

## Related Rules

- `no-object-deps` - Detects unstable objects in dependency arrays
- `unstable-props-to-memo` - Detects unstable props to memoized components
- `missing-memo` (future) - Suggests memoization for expensive components

## References

- [React Docs: You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect#adjusting-some-state-when-a-prop-changes)
- [React Docs: Deriving State](https://react.dev/learn/you-might-not-need-an-effect#updating-state-based-on-props-or-state)
- [React Docs: Choosing the State Structure](https://react.dev/learn/choosing-the-state-structure#avoid-duplication-in-state)

## Version

Added in: v0.2.0
Last updated: 2025-11-17
