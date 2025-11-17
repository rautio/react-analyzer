# no-stale-state

Prevents stale state bugs by requiring state updates that depend on previous state to use the functional form.

## Rule Details

This rule detects when state setters update state based on the previous value without using the functional form, which can cause race conditions and unexpected behavior.

## The Problem

When you call a state setter with a value that depends on the current state, using the direct form can lead to bugs:

```tsx
// ❌ BAD - Uses stale value, causes race conditions
const [count, setCount] = useState(0);

function increment() {
  setCount(count + 1);  // References count from this render
  setCount(count + 1);  // Still uses same stale count!
  // Result: count becomes 1, not 2 ❌
}
```

**Why it fails:**
- `count` is captured from the current render (closure)
- Multiple updates in the same cycle use the same stale value
- React batches updates, but closures don't update
- Async updates use values from when the closure was created

## The Solution

Use the functional form of state updates, which React guarantees will use the latest value:

```tsx
// ✅ GOOD - Always uses latest value
const [count, setCount] = useState(0);

function increment() {
  setCount(prev => prev + 1);  // Uses actual current value
  setCount(prev => prev + 1);  // Uses value after first update
  // Result: count becomes 2 ✅
}
```

**Why it works:**
- React guarantees `prev` is always the most up-to-date value
- Works correctly with batched updates
- Safe in async callbacks and effects
- No stale closure issues

## Examples

### ❌ Incorrect

```tsx
// Arithmetic operations
const [count, setCount] = useState(0);
setCount(count + 1);  // ❌
setCount(count - 1);  // ❌
setValue(value * 2);  // ❌

// Toggle boolean
const [isOpen, setIsOpen] = useState(false);
setIsOpen(!isOpen);  // ❌

// Array operations
const [items, setItems] = useState([]);
setItems([...items, newItem]);  // ❌

// Object updates
const [user, setUser] = useState({ name: '', age: 0 });
setUser({ ...user, age: user.age + 1 });  // ❌

// Async callbacks (especially dangerous!)
setTimeout(() => {
  setCount(count + 1);  // ❌ Stale closure!
}, 1000);

// Event handlers (can be called rapidly)
<button onClick={() => setCount(count + 1)}>  // ❌
  Increment
</button>

// Effects (can run multiple times)
useEffect(() => {
  setCount(count + 1);  // ❌
}, [dependency]);
```

### ✅ Correct

```tsx
// Arithmetic operations
const [count, setCount] = useState(0);
setCount(prev => prev + 1);  // ✅
setCount(prev => prev - 1);  // ✅
setValue(prev => prev * 2);  // ✅

// Toggle boolean
const [isOpen, setIsOpen] = useState(false);
setIsOpen(prev => !prev);  // ✅

// Array operations
const [items, setItems] = useState([]);
setItems(prev => [...prev, newItem]);  // ✅

// Object updates
const [user, setUser] = useState({ name: '', age: 0 });
setUser(prev => ({ ...prev, age: prev.age + 1 }));  // ✅

// Async callbacks (safe)
setTimeout(() => {
  setCount(prev => prev + 1);  // ✅
}, 1000);

// Event handlers
<button onClick={() => setCount(prev => prev + 1)}>  // ✅
  Increment
</button>

// Effects
useEffect(() => {
  setCount(prev => prev + 1);  // ✅
}, [dependency]);
```

### ✅ Exceptions (Not Derived from Previous State)

These are safe because they're setting to a new value, not deriving from the previous state:

```tsx
// Setting to specific value - OK
setCount(0);          // ✅
setCount(42);         // ✅
setValue(newValue);   // ✅

// Setting from props - OK
setName(user.name);   // ✅

// Setting from another state - OK
const [count, setCount] = useState(0);
const [max, setMax] = useState(100);
setCount(max);  // ✅ Not derived from count

// Conditional set to literal - OK
setIsOpen(false);  // ✅
setIsOpen(true);   // ✅
```

## Common Bug Scenarios

### 1. Multiple Rapid Updates

```tsx
// ❌ BUG: User clicks 3 times, count becomes 1 (not 3!)
function Counter() {
  const [count, setCount] = useState(0);

  return (
    <button onClick={() => setCount(count + 1)}>
      Count: {count}
    </button>
  );
}

// ✅ FIX: Count correctly becomes 3
function Counter() {
  const [count, setCount] = useState(0);

  return (
    <button onClick={() => setCount(prev => prev + 1)}>
      Count: {count}
    </button>
  );
}
```

### 2. Timer/Interval Stale Closures

```tsx
// ❌ BUG: Timer gets stuck at 1!
function Timer() {
  const [seconds, setSeconds] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setSeconds(seconds + 1);  // Always adds 1 to 0!
    }, 1000);

    return () => clearInterval(interval);
  }, []); // Empty deps - seconds is stale!

  return <div>Seconds: {seconds}</div>;
}

// ✅ FIX: Timer works correctly
function Timer() {
  const [seconds, setSeconds] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setSeconds(prev => prev + 1);  // Always uses latest!
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  return <div>Seconds: {seconds}</div>;
}
```

### 3. Toggle Race Condition

```tsx
// ❌ BUG: Rapid toggles can get out of sync
function Modal() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <button onClick={() => setIsOpen(!isOpen)}>
      Toggle
    </button>
  );
}

// ✅ FIX: Always toggles correctly
function Modal() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <button onClick={() => setIsOpen(prev => !prev)}>
      Toggle
    </button>
  );
}
```

### 4. API Response Race Condition

```tsx
// ❌ BUG: Multiple API calls can overwrite each other
function ItemList() {
  const [items, setItems] = useState([]);

  const loadMore = async () => {
    const newItems = await fetchItems();
    setItems([...items, ...newItems]);  // Stale items!
  };

  return <button onClick={loadMore}>Load More</button>;
}

// ✅ FIX: Updates always accumulate correctly
function ItemList() {
  const [items, setItems] = useState([]);

  const loadMore = async () => {
    const newItems = await fetchItems();
    setItems(prev => [...prev, ...newItems]);
  };

  return <button onClick={loadMore}>Load More</button>;
}
```

## When This Rule Matters Most

### Critical (Always a Bug)

1. **Multiple updates in sequence** - Will definitely lose updates
2. **Async callbacks** (setTimeout, promises) - Guaranteed stale closure
3. **Intervals/timers** - Will get stuck
4. **Toggle operations** - Can get out of sync with rapid clicks

### Important (Likely a Bug)

5. **Event handlers** - Can be called multiple times rapidly
6. **useEffect updates** - Can run multiple times
7. **Array/object spreads** - Can lose concurrent updates

## Performance Impact

Beyond correctness, this pattern can cause:
- Lost user actions (clicks that don't register)
- Incorrect application state
- UI getting out of sync with data
- Frustrating user experience ("Why doesn't this work?")

## Typical Symptoms

Users report:
- "Button works when I click slowly but fails when I click fast"
- "Timer gets stuck at 1"
- "Sometimes items disappear from my list"
- "Toggle switch gets out of sync"
- "Counter shows wrong number"

## References

- [React Docs: Queueing a Series of State Updates](https://react.dev/learn/queueing-a-series-of-state-updates)
- [React Docs: Updating State Based on Previous State](https://react.dev/reference/react/useState#updating-state-based-on-the-previous-state)
- [Common React Mistakes: Stale Closures](https://dmitripavlutin.com/react-hooks-stale-closures/)

## Severity

**Error** - This almost always indicates a bug that will cause issues in production.
