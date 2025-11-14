# no-object-deps

**Prevent unstable object/array dependencies in React hooks**

## Rule Details

This rule catches a common but dangerous pattern: using objects or arrays as dependencies in React hooks (useEffect, useMemo, useCallback) without proper memoization. This causes infinite re-render loops and severe performance issues.

### Why This Matters

In JavaScript, objects and arrays are compared by **reference**, not by value:

```javascript
{} === {}           // false - different references!
[1, 2] === [1, 2]   // false - different references!
```

When you create an object or array inside a component, React creates a **new reference every render**:

```tsx
function Component() {
  const config = { theme: 'dark' };  // ← New object EVERY render

  useEffect(() => {
    applyConfig(config);
  }, [config]);  // ← Dependency changes EVERY render = infinite loop!
}
```

**Result:** Your effect runs on every single render, potentially causing:
- Infinite re-render loops
- Excessive API calls
- Performance degradation
- Browser crashes

## ❌ Incorrect Code

### Example 1: Object dependency causing infinite loop

```tsx
function UserProfile({ userId }) {
  const [user, setUser] = useState(null);

  // ❌ BAD: config is a new object every render
  const config = {
    userId,
    includeAvatar: true
  };

  useEffect(() => {
    fetchUser(config).then(setUser);
  }, [config]);  // ← Runs every render!

  return <div>{user?.name}</div>;
}
```

**What happens:**
1. Component renders → creates new `config` object
2. useEffect sees `config` changed → runs effect
3. Effect calls `setUser` → triggers re-render
4. Go to step 1 → **INFINITE LOOP** 🔥

---

### Example 2: Array dependency in useMemo

```tsx
function ProductList({ products }) {
  // ❌ BAD: filters is a new array every render
  const filters = ['inStock', 'onSale'];

  const filtered = useMemo(() => {
    return products.filter(p =>
      filters.some(f => p.tags.includes(f))
    );
  }, [products, filters]);  // ← filters changes every render!

  return <List items={filtered} />;
}
```

**What happens:**
- `useMemo` is supposed to cache the result
- But `filters` is a new array every render
- So `useMemo` re-computes every time
- **No memoization benefit** - wasted effort!

---

### Example 3: Nested object in useCallback

```tsx
function Form() {
  const [name, setName] = useState('');

  // ❌ BAD: validation rules object recreated every render
  const validationRules = {
    minLength: 3,
    maxLength: 50,
    pattern: /^[a-zA-Z\s]+$/
  };

  const handleSubmit = useCallback((e) => {
    e.preventDefault();
    if (validate(name, validationRules)) {
      submit(name);
    }
  }, [name, validationRules]);  // ← validationRules changes every render!

  return <form onSubmit={handleSubmit}>...</form>;
}
```

**What happens:**
- `useCallback` is supposed to return the same function reference
- But `validationRules` changes every render
- So a new function is created every render
- Child components re-render unnecessarily

---

## ✅ Correct Code

### Solution 1: Extract to module-level constant

**Best for:** Static configuration that never changes

```tsx
// ✅ GOOD: Constant outside component
const CONFIG = {
  includeAvatar: true
};

function UserProfile({ userId }) {
  const [user, setUser] = useState(null);

  useEffect(() => {
    fetchUser({ ...CONFIG, userId }).then(setUser);
  }, [userId]);  // ← Only depends on userId (stable)

  return <div>{user?.name}</div>;
}
```

**Why it works:** `CONFIG` is created once and never changes.

---

### Solution 2: Use useMemo for dynamic objects

**Best for:** Objects that depend on props/state

```tsx
function UserProfile({ userId, includeAvatar }) {
  const [user, setUser] = useState(null);

  // ✅ GOOD: Memoized object, only recreated when dependencies change
  const config = useMemo(() => ({
    userId,
    includeAvatar
  }), [userId, includeAvatar]);

  useEffect(() => {
    fetchUser(config).then(setUser);
  }, [config]);  // ← config is stable (only changes when deps change)

  return <div>{user?.name}</div>;
}
```

**Why it works:** `useMemo` returns the same object reference until `userId` or `includeAvatar` changes.

---

### Solution 3: Inline the object (avoid the dependency)

**Best for:** Simple objects used only in the effect

```tsx
function UserProfile({ userId }) {
  const [user, setUser] = useState(null);

  // ✅ GOOD: No dependency at all!
  useEffect(() => {
    const config = { userId, includeAvatar: true };
    fetchUser(config).then(setUser);
  }, [userId]);  // ← Only depends on primitive value

  return <div>{user?.name}</div>;
}
```

**Why it works:** The object is created inside the effect, so it doesn't need to be a dependency.

---

### Solution 4: Extract array to constant

```tsx
// ✅ GOOD: Static array as constant
const FILTERS = ['inStock', 'onSale'];

function ProductList({ products }) {
  const filtered = useMemo(() => {
    return products.filter(p =>
      FILTERS.some(f => p.tags.includes(f))
    );
  }, [products]);  // ← No need to include FILTERS

  return <List items={filtered} />;
}
```

---

### Solution 5: Use primitive dependencies

```tsx
function Form() {
  const [name, setName] = useState('');

  // ✅ GOOD: All dependencies are primitives
  const handleSubmit = useCallback((e) => {
    e.preventDefault();
    const rules = {
      minLength: 3,
      maxLength: 50,
      pattern: /^[a-zA-Z\s]+$/
    };
    if (validate(name, rules)) {
      submit(name);
    }
  }, [name]);  // ← Only depends on primitive string

  return <form onSubmit={handleSubmit}>...</form>;
}
```

---

## Decision Tree: How to Fix

```
Is the object/array static (never changes)?
├─ YES → Extract to module constant
│         const CONFIG = { ... };
│
└─ NO → Does it depend on props/state?
    ├─ YES → Use useMemo
    │         const config = useMemo(() => ({ ... }), [deps]);
    │
    └─ NO → Is it only used inside the hook?
        ├─ YES → Move inside the hook body
        │         useEffect(() => {
        │           const config = { ... };
        │         }, [primitives]);
        │
        └─ NO → Consider if you really need it
```

---

## Common Mistakes

### Mistake 1: Memoizing with wrong dependencies

```tsx
// ❌ WRONG: useMemo but missing dependencies
const config = useMemo(() => ({
  userId,
  theme: selectedTheme  // ← Uses selectedTheme
}), [userId]);  // ← But doesn't include selectedTheme!

// ✅ RIGHT: Include all dependencies
const config = useMemo(() => ({
  userId,
  theme: selectedTheme
}), [userId, selectedTheme]);
```

### Mistake 2: Over-memoizing simple values

```tsx
// ❌ UNNECESSARY: Don't memoize simple primitives
const count = useMemo(() => items.length, [items]);

// ✅ BETTER: Just compute directly (it's fast!)
const count = items.length;
```

### Mistake 3: Forgetting spread operators

```tsx
// ❌ WRONG: Spreading inside useMemo
const merged = useMemo(() => ({
  ...baseConfig,  // ← If baseConfig is unstable, this breaks!
  userId
}), [userId]);

// ✅ RIGHT: Include baseConfig as dependency
const merged = useMemo(() => ({
  ...baseConfig,
  userId
}), [baseConfig, userId]);
```

---

## Performance Impact

**Severity:** 🔴 **Critical**

- **Infinite loops:** Can crash the browser
- **Excessive re-renders:** 10-100x more renders than necessary
- **Wasted API calls:** Same fetch executed repeatedly
- **Poor UX:** Janky, unresponsive UI

**Real-world example:**
```tsx
// This caused 300+ re-renders per second in production:
useEffect(() => {
  analytics.track('pageView', { page: location.pathname });
}, [{ page: location.pathname }]);  // ← Object dependency!

// Fix: Use primitive
useEffect(() => {
  analytics.track('pageView', { page: location.pathname });
}, [location.pathname]);  // ← Only 1 re-render per page change
```

---

## When to Disable

This rule should **almost never** be disabled. However, rare exceptions:

```tsx
// If you WANT the effect to run every render (rare!)
useEffect(() => {
  // Some logging or analytics
  log({ timestamp: Date.now(), data });
}, [{ data }]);  // ← Intentionally runs every render

// Better: Just omit the dependency array
useEffect(() => {
  log({ timestamp: Date.now(), data });
});  // ← Clearer intent: "run every render"
```

To disable for a specific line:
```tsx
// react-analyzer-disable-next-line no-object-deps
}, [config]);
```

---

## References

- [React Docs: useEffect](https://react.dev/reference/react/useEffect)
- [React Docs: useMemo](https://react.dev/reference/react/useMemo)
- [A Complete Guide to useEffect](https://overreacted.io/a-complete-guide-to-useeffect/)
- [Why React Re-Renders](https://www.joshwcomeau.com/react/why-react-re-renders/)

---

## Related Rules

- **exhaustive-deps** - Ensures all referenced values are in dependency array
- **no-unstable-props** - Detects inline objects passed as props
- **memo-unstable-props** - Validates React.memo effectiveness (cross-file)

---

**Rule ID:** `no-object-deps`
**Category:** Performance
**Severity:** Error
**Auto-fixable:** Partial (suggests fixes)
