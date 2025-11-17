# no-inline-props

Prevents passing inline objects, arrays, and functions as JSX props, which causes unnecessary re-renders.

## Rule Details

This rule detects when JSX props are set to inline objects, arrays, or functions that are created on every render. These inline values have new references each render, causing child components to re-render unnecessarily and breaking memoization.

## The Problem

When you pass an inline object, array, or function as a prop, React creates a **new reference** every render:

```tsx
function Parent() {
  return (
    <Child
      config={{ theme: 'dark' }}      // ❌ New object every render
      items={[1, 2, 3]}               // ❌ New array every render
      onClick={() => handleClick()}   // ❌ New function every render
    />
  );
}

// Child re-renders EVERY time Parent renders, even if nothing changed!
const Child = memo(({ config, items, onClick }) => {
  return <div>...</div>;
});
```

**Why it's a problem:**
- Child component receives new prop references every render
- React.memo is completely bypassed (always re-renders)
- useMemo/useCallback in child components become useless
- Cascading re-renders down the component tree
- Performance degrades exponentially with component depth

## The Solution

Move stable values outside the component, or memoize values that depend on state/props:

```tsx
// ✅ Move constants outside
const CONFIG = { theme: 'dark' };
const ITEMS = [1, 2, 3];

function Parent() {
  // ✅ Memoize callbacks
  const handleClick = useCallback(() => {
    doSomething();
  }, []);

  return (
    <Child
      config={CONFIG}        // ✅ Same reference every render
      items={ITEMS}          // ✅ Same reference every render
      onClick={handleClick}  // ✅ Same reference every render
    />
  );
}

// Now Child only re-renders when props actually change!
const Child = memo(({ config, items, onClick }) => {
  return <div>...</div>;
});
```

## Examples

### ❌ Incorrect

```tsx
function App() {
  const [count, setCount] = useState(0);

  return (
    <>
      {/* Inline objects */}
      <UserProfile user={{ name: 'John', age: 30 }} />
      <Settings config={{ theme: 'dark', locale: 'en' }} />

      {/* Inline arrays */}
      <List items={[1, 2, 3, 4, 5]} />
      <Tags tags={['react', 'javascript']} />

      {/* Inline functions */}
      <Button onClick={() => setCount(count + 1)} />
      <Form onSubmit={() => handleSubmit()} />
      <Input onChange={(e) => setValue(e.target.value)} />

      {/* Inline styles (object) */}
      <div style={{ color: 'red', fontSize: 16 }} />

      {/* Nested inline values */}
      <Component
        data={{
          items: [1, 2, 3],
          handler: () => click()
        }}
      />
    </>
  );
}
```

### ✅ Correct

```tsx
// Move constants outside component
const USER = { name: 'John', age: 30 };
const CONFIG = { theme: 'dark', locale: 'en' };
const ITEMS = [1, 2, 3, 4, 5];
const TAGS = ['react', 'javascript'];
const STYLE = { color: 'red', fontSize: 16 };

function App() {
  const [count, setCount] = useState(0);
  const [value, setValue] = useState('');

  // Memoize callbacks
  const handleIncrement = useCallback(() => {
    setCount(prev => prev + 1);
  }, []);

  const handleSubmit = useCallback(() => {
    // submit logic
  }, []);

  const handleChange = useCallback((e) => {
    setValue(e.target.value);
  }, []);

  return (
    <>
      {/* Stable references */}
      <UserProfile user={USER} />
      <Settings config={CONFIG} />
      <List items={ITEMS} />
      <Tags tags={TAGS} />

      {/* Memoized callbacks */}
      <Button onClick={handleIncrement} />
      <Form onSubmit={handleSubmit} />
      <Input onChange={handleChange} />

      {/* Stable style */}
      <div style={STYLE} />
    </>
  );
}
```

### ✅ Exceptions (Valid Use Cases)

These patterns are acceptable and will NOT be flagged:

```tsx
function App() {
  const theme = useTheme();
  const stable = CONSTANT;

  return (
    <>
      {/* Primitive literals - OK */}
      <Input value="hello" />
      <Counter count={42} />
      <Toggle enabled={true} />
      <Icon size={16} />

      {/* Variables/identifiers - OK (assumed stable) */}
      <Settings theme={theme} />
      <Data value={stable} />

      {/* Special JSX props - OK */}
      <Component key="unique-key" />
      <div className="container" />

      {/* Computed primitives - OK */}
      <Progress value={count * 2} />
      <Text content={name + " Smith"} />
    </>
  );
}
```

## Common Scenarios

### 1. Event Handlers

```tsx
// ❌ BAD: New function every render
<button onClick={() => setCount(count + 1)}>Increment</button>

// ✅ GOOD: Memoized callback
const increment = useCallback(() => setCount(prev => prev + 1), []);
<button onClick={increment}>Increment</button>

// ✅ ALSO GOOD: For simple cases, direct function reference
<button onClick={handleClick}>Click</button>
```

### 2. Configuration Objects

```tsx
// ❌ BAD: New object every render
<Chart options={{ responsive: true, animations: false }} />

// ✅ GOOD: Constant outside component
const CHART_OPTIONS = { responsive: true, animations: false };
<Chart options={CHART_OPTIONS} />

// ✅ GOOD: useMemo for dynamic options
const options = useMemo(() => ({
  responsive: true,
  animations: !isPrinting
}), [isPrinting]);
<Chart options={options} />
```

### 3. Lists and Arrays

```tsx
// ❌ BAD: New array every render
<Dropdown options={['option1', 'option2', 'option3']} />

// ✅ GOOD: Constant outside
const OPTIONS = ['option1', 'option2', 'option3'];
<Dropdown options={OPTIONS} />

// ✅ GOOD: useMemo for filtered/derived arrays
const filteredItems = useMemo(
  () => items.filter(x => x.active),
  [items]
);
<List items={filteredItems} />
```

### 4. Inline Styles

```tsx
// ❌ BAD: New object every render
<div style={{ padding: 20, margin: 10 }} />

// ✅ GOOD: CSS classes (preferred)
<div className="container" />

// ✅ GOOD: Constant for dynamic styles
const CONTAINER_STYLE = { padding: 20, margin: 10 };
<div style={CONTAINER_STYLE} />

// ✅ GOOD: useMemo for truly dynamic styles
const style = useMemo(() => ({
  backgroundColor: isDark ? 'black' : 'white'
}), [isDark]);
<div style={style} />
```

## Performance Impact

**Without this rule:**
```tsx
function Parent() {
  const [count, setCount] = useState(0);

  return <Child onClick={() => setCount(count + 1)} />;
  // ❌ Child re-renders EVERY time Parent re-renders
  // Even if count hasn't changed!
}
```

**With this rule applied:**
```tsx
function Parent() {
  const [count, setCount] = useState(0);

  const handleClick = useCallback(() => setCount(prev => prev + 1), []);

  return <Child onClick={handleClick} />;
  // ✅ Child only re-renders when handleClick reference changes (never)
}
```

**Real-world impact:**
- **Small components:** Minimal impact (few milliseconds)
- **Large component trees:** Exponential impact (seconds of wasted rendering)
- **Lists with inline props:** Each item re-renders unnecessarily
- **Deeply nested components:** Cascading re-renders throughout the tree

## When This Rule Matters Most

### Critical (Always Flag)

1. **Props to memoized components** - Completely defeats React.memo
2. **Props in lists** - Every item re-renders on every parent render
3. **Event handlers** - Especially in frequently updating components
4. **Large objects/arrays** - More expensive to compare and process

### Important (Should Flag)

5. **Any prop in performance-critical paths** - Charts, tables, virtualized lists
6. **Props passed through multiple levels** - Cascading effect amplifies
7. **Inline styles** - Creates new objects for DOM diffing

## Common Mistakes

### "But it's just a small object!"

```tsx
// ❌ "It's just a small object, no big deal"
<Tooltip config={{ placement: 'top' }} />

// The issue isn't the object size, it's the reference change
// Tooltip re-renders EVERY time parent renders, even if placement never changes
```

### "But I need to pass parameters to the handler!"

```tsx
// ❌ WRONG: Creating new function to pass parameter
<Button onClick={() => handleClick(id)} />

// ✅ RIGHT: Use useCallback with dependencies
const handleClick = useCallback(() => handleClick(id), [id]);
<Button onClick={handleClick} />

// ✅ ALTERNATIVE: Use data attributes or currying
<Button onClick={handleClick} data-id={id} />
```

### "But the child isn't memoized anyway!"

Even if the child isn't memoized NOW:
- It might be memoized later for performance
- Its children might be memoized
- You're forming bad habits
- It's a code smell that indicates lack of understanding

## References

- [React Docs: Preserving and Resetting State](https://react.dev/learn/preserving-and-resetting-state)
- [React Docs: memo](https://react.dev/reference/react/memo)
- [Kent C. Dodds: Fix the slow render before you fix the re-render](https://kentcdodds.com/blog/fix-the-slow-render-before-you-fix-the-re-render)
- [Web.dev: Optimize React Re-renders](https://web.dev/optimize-react-re-renders/)

## Severity

**Warning** - While this doesn't break functionality, it significantly impacts performance and is considered bad practice.
