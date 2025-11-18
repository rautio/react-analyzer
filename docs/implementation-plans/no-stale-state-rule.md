# Implementation Plan: no-stale-state

## Overview

Prevent stale state bugs by detecting when state setters update state based on the previous value without using the functional form, which causes race conditions and stale closures.

**Severity:** Error
**Type:** Single-file analysis
**Estimated Effort:** 4-6 hours (Phase 1), 2-3 hours (Phase 2)

---

## Detection Strategy

### Core Pattern

```tsx
const [state, setState] = useState(initialValue);

// ❌ Detect: setState with expression that references state
setState(state + 1)
setState(!state)
setState([...state, x])
setState({ ...state, x })

// ✅ Allow: setState with new value (not derived)
setState(0)
setState(newValue)
setState(otherState)
```

---

## Implementation Phases

### Phase 1: Basic Detection (4-6 hours)

**Goal:** Detect direct state variable references in setState calls

**What to detect:**
1. Binary expressions: `setState(state + 1)`, `setState(state - 1)`, `setState(state * 2)`
2. Unary expressions: `setState(!state)`, `setState(-state)`
3. Spread expressions: `setState([...state, x])`, `setState({ ...state, x })`
4. Member expressions: `setState(state.concat(x))`, `setState(state.filter(...))`

**Algorithm:**
1. Find all useState declarations in the file
   - Extract state variable name and setter name
   - Store mapping: `{ setter: 'setCount', state: 'count' }`

2. Find all calls to state setters
   - Pattern: `setCount(expression)`
   - Get the argument expression

3. Check if expression references the state variable
   - Walk the expression AST
   - Look for identifier nodes matching the state variable name
   - If found → violation

4. Report violation with suggestion
   - Show the exact expression
   - Suggest functional form: `setState(prev => <expression with state replaced by prev>)`

**AST Patterns:**

```typescript
// useState declaration
variable_declarator {
  name: array_pattern {
    identifier "count"      // state variable
    identifier "setCount"   // setter
  }
  value: call_expression {
    function: identifier "useState"
    arguments: [initialValue]
  }
}

// setState call
call_expression {
  function: identifier "setCount"
  arguments: [
    binary_expression {       // count + 1
      left: identifier "count"   // ← References state!
      operator: "+"
      right: number "1"
    }
  ]
}
```

**Expression Types to Check:**

1. **Binary expressions:** `state + 1`, `state - x`, `state * 2`
   - Node type: `binary_expression`
   - Check left and right operands

2. **Unary expressions:** `!state`, `-state`
   - Node type: `unary_expression`
   - Check argument

3. **Spread expressions:** `[...state, x]`, `{ ...state, x }`
   - Array: `array` with `spread_element`
   - Object: `object` with `spread_element`
   - Check spread argument

4. **Member expressions:** `state.concat(x)`, `state.filter(...)`
   - Node type: `member_expression`
   - Check object

5. **Call expressions:** `state.map(...)`, `state.slice()`
   - Node type: `call_expression` with `member_expression`
   - Check if member expression object is state

**Helper Functions:**

```go
// Find useState declarations
func findUseStateDeclarations(node *parser.Node) []StateMapping

type StateMapping struct {
    StateName  string
    SetterName string
    Line       uint32
}

// Find setState calls
func findSetStateCalls(node *parser.Node, setters map[string]string) []SetStateCall

type SetStateCall struct {
    SetterName string
    StateName  string  // From mapping
    Argument   *parser.Node
    Line       uint32
}

// Check if expression references state variable
func referencesStateVariable(expr *parser.Node, stateName string) bool

// Walk expression tree looking for identifier matching stateName
func walkExpressionForIdentifier(node *parser.Node, targetName string) bool
```

**Test Cases:**

```tsx
// File: functional-state-simple.tsx
import { useState } from 'react';

// ❌ Arithmetic operations
export function Counter() {
  const [count, setCount] = useState(0);

  const increment = () => setCount(count + 1);     // Line 6
  const decrement = () => setCount(count - 1);     // Line 7
  const double = () => setCount(count * 2);        // Line 8
}

// ❌ Unary operations
export function Toggle() {
  const [isOpen, setIsOpen] = useState(false);

  const toggle = () => setIsOpen(!isOpen);         // Line 15
}

// ❌ Array operations
export function ItemList() {
  const [items, setItems] = useState([]);

  const addItem = (item) => setItems([...items, item]);           // Line 22
  const removeFirst = () => setItems(items.slice(1));             // Line 23
  const filtered = () => setItems(items.filter(x => x.active));   // Line 24
}

// ❌ Object operations
export function UserProfile() {
  const [user, setUser] = useState({ name: '', age: 0 });

  const incrementAge = () => setUser({ ...user, age: user.age + 1 });  // Line 31
  const updateName = (name) => setUser({ ...user, name });             // Line 32
}

// ✅ Valid: Setting to new value
export function ValidUpdates() {
  const [count, setCount] = useState(0);
  const [name, setName] = useState('');

  const reset = () => setCount(0);              // OK - literal
  const setMax = () => setCount(100);           // OK - literal
  const updateName = (n) => setName(n);         // OK - new value
}

// ✅ Valid: Functional form
export function CorrectPattern() {
  const [count, setCount] = useState(0);
  const [isOpen, setIsOpen] = useState(false);

  const increment = () => setCount(prev => prev + 1);  // OK
  const toggle = () => setIsOpen(prev => !prev);       // OK
}
```

**Expected Violations (Phase 1):** 10 violations in simple.tsx

**Message Format:**
```
State 'count' is updated using its current value without the functional form, which can cause race conditions. Replace with: setCount(prev => prev + 1)
```

---

### Phase 2: Context-Aware Detection (2-3 hours)

**Goal:** Add context to warnings (high-risk contexts get stronger warnings)

**High-Risk Contexts:**

1. **Async callbacks** (setTimeout, promises)
   ```tsx
   setTimeout(() => setState(state + 1), 1000)
   fetch().then(() => setState(state + 1))
   ```

2. **useEffect** (can run multiple times)
   ```tsx
   useEffect(() => {
     setState(state + 1)
   }, [deps])
   ```

3. **setInterval** (guaranteed stale closure)
   ```tsx
   setInterval(() => setState(state + 1), 1000)
   ```

**Detection:**
- Check if setState call is inside:
  - `call_expression` where function is `setTimeout`, `setInterval`, `requestAnimationFrame`
  - `arrow_function` or `function` passed to `.then()`, `.catch()`, `.finally()`
  - `useEffect` body

**Enhanced Message:**
```
State 'count' is updated in an async callback using its current value, which WILL cause stale closure bugs. Replace with: setCount(prev => prev + 1)
```

**Test Cases:**

```tsx
// File: functional-state-async.tsx

// ❌ setTimeout (high risk!)
export function DelayedCounter() {
  const [count, setCount] = useState(0);

  const delayedIncrement = () => {
    setTimeout(() => {
      setCount(count + 1);  // Line 7 - HIGH RISK
    }, 1000);
  };
}

// ❌ setInterval (guaranteed bug!)
export function Timer() {
  const [seconds, setSeconds] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setSeconds(seconds + 1);  // Line 17 - GUARANTEED BUG
    }, 1000);

    return () => clearInterval(interval);
  }, []);
}

// ❌ Promise callbacks
export function AsyncLoader() {
  const [items, setItems] = useState([]);

  const loadMore = () => {
    fetchItems().then(newItems => {
      setItems([...items, ...newItems]);  // Line 30 - HIGH RISK
    });
  };
}

// ❌ useEffect
export function EffectUpdate() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    setCount(count + 1);  // Line 40 - HIGH RISK
  }, [dependency]);
}
```

**Expected Violations (Phase 2):** 4 high-risk violations

---

### Phase 3: Smart Suggestions (1 hour)

**Goal:** Provide accurate auto-fix suggestions

**Suggestion Algorithm:**

1. Extract the expression from `setState(expression)`
2. Replace all occurrences of state variable with `prev`
3. Wrap in arrow function: `prev => <modified expression>`

**Examples:**

```tsx
// Detected:
setCount(count + 1)
// Suggest:
setCount(prev => prev + 1)

// Detected:
setIsOpen(!isOpen)
// Suggest:
setIsOpen(prev => !prev)

// Detected:
setItems([...items, newItem])
// Suggest:
setItems(prev => [...prev, newItem])

// Detected:
setUser({ ...user, age: user.age + 1 })
// Suggest:
setUser(prev => ({ ...prev, age: prev.age + 1 }))
```

**Implementation:**
```go
func generateFunctionalSuggestion(expr *parser.Node, stateName string) string {
    // Get expression text
    exprText := expr.Text()

    // Replace state variable with 'prev'
    // This is simplified - actual implementation needs AST manipulation
    suggestion := strings.ReplaceAll(exprText, stateName, "prev")

    return fmt.Sprintf("prev => %s", suggestion)
}
```

---

## Edge Cases

### Should NOT Flag

1. **Setting to literal or new value**
   ```tsx
   setState(0)           // OK
   setState(newValue)    // OK
   setState(props.value) // OK
   ```

2. **Different state variable**
   ```tsx
   const [count, setCount] = useState(0);
   const [max, setMax] = useState(100);
   setCount(max);  // OK - not derived from count
   ```

3. **Already using functional form**
   ```tsx
   setState(prev => prev + 1)  // OK - already correct
   ```

### Complex Cases (Future)

1. **Computed property access**
   ```tsx
   setState(state[key])  // Not straightforward to detect
   ```

2. **Function calls with state**
   ```tsx
   setState(transform(state))  // Harder to detect
   ```

3. **Ternary expressions**
   ```tsx
   setState(condition ? state + 1 : state - 1)  // Phase 1 should catch
   ```

---

## Test Structure

```
test/fixtures/
  no-stale-state-simple.tsx      // Phase 1: Basic patterns
  no-stale-state-async.tsx       // Phase 2: Async contexts
  no-stale-state-valid.tsx       // Valid patterns (no violations)
  no-stale-state-edge-cases.tsx  // Edge cases

internal/rules/
  no_stale_state.go
  no_stale_state_test.go
```

---

## Success Criteria

**Phase 1:**
- ✅ Detect all basic patterns (arithmetic, unary, spread, member)
- ✅ All tests passing
- ✅ Zero false positives on valid patterns
- ✅ Accurate suggestions generated

**Phase 2:**
- ✅ Detect async contexts (setTimeout, promises, effects)
- ✅ Enhanced warnings for high-risk contexts
- ✅ All tests passing

**Phase 3:**
- ✅ Auto-fix suggestions are accurate
- ✅ Suggestions work for all expression types

---

## Estimated Timeline

- **Phase 1:** 4-6 hours
  - 2 hours: Core detection logic
  - 2 hours: Test fixtures and tests
  - 1-2 hours: Debugging and edge cases

- **Phase 2:** 2-3 hours
  - 1 hour: Context detection
  - 1 hour: Test cases
  - 1 hour: Integration

- **Phase 3:** 1 hour
  - Suggestion generation and testing

**Total:** 7-10 hours

---

## Rule Configuration (Future)

```json
{
  "no-stale-state": {
    "level": "error",
    "strictness": "all" | "async-only" | "multi-update-only"
  }
}
```

- `"all"` - Flag all updates derived from previous state (recommended)
- `"async-only"` - Only flag updates in async contexts
- `"multi-update-only"` - Only flag when multiple updates detected

---

## Related Anti-Patterns

This rule is related to but distinct from:
- `no-derived-state` - Detects useState mirroring props
- `exhaustive-deps` - Can cause this issue when deps are incomplete
