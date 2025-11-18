# Implementation Plan: no-derived-state

**Rule Name:** `no-derived-state` ✅
**Status:** Planning
**Complexity:** Medium
**Estimated Time:** 11-16 hours

## Overview

Detect when `useState` is used to store a value derived from props, then synced via `useEffect`. This causes unnecessary re-renders and complexity.

## Detection Algorithm

### Pattern to Detect
```tsx
const [state, setState] = useState(prop); // or useState(prop.field)
useEffect(() => {
  setState(prop); // or setState(prop.field)
}, [prop]);
```

### High-Level Flow
1. Find all components (function declarations/expressions)
2. For each component:
   - Extract props (function parameters)
   - Find all useState calls
   - Find all useEffect calls
   - Match useState + useEffect pairs that form the anti-pattern
   - Verify no other state mutations exist
3. Report violations with suggested fixes

## Implementation Phases

### Phase 1: Basic Detection (MVP) - 4-6 hours
Detect simple prop mirroring:
```tsx
const [value, setValue] = useState(prop);
useEffect(() => setValue(prop), [prop]);
```

**Tasks:**
- Extract props from component parameters
- Find useState with prop in initializer
- Find useEffect that sets the same state
- Verify effect depends on the same prop
- Report violation

### Phase 2: Advanced Detection - 3-4 hours
Detect derived values:
```tsx
const [name, setName] = useState(user.name);
useEffect(() => setName(user.name), [user.name]);
```

**Tasks:**
- Handle member expressions (user.name)
- Handle call expressions (calc(items))
- Match complex initializers with effect body
- Extract dependency from nested prop access

### Phase 3: False Positive Reduction - 2-3 hours
Avoid reporting when state is legitimately needed:

**Tasks:**
- Check if setter is called in event handlers
- Check if effect has other side effects
- Check if state is used in callbacks
- Detect "initial value" pattern (no useEffect)

### Phase 4: Smart Suggestions - 2-3 hours
Provide actionable fix suggestions:

**Tasks:**
- Suggest direct derivation for simple cases
- Suggest useMemo for expensive computations
- Show before/after code
- Explain why the pattern is problematic

## Test Cases

### Should Detect ✅

**Simple mirroring:**
```tsx
const [name, setName] = useState(user);
useEffect(() => setName(user), [user]);
```

**Derived property:**
```tsx
const [name, setName] = useState(user.name);
useEffect(() => setName(user.name), [user.name]);
```

**Computed value:**
```tsx
const [total, setTotal] = useState(calc(items));
useEffect(() => setTotal(calc(items)), [items]);
```

### Should NOT Detect ❌

**User interaction (no useEffect):**
```tsx
const [color, setColor] = useState(initialColor);
return <input value={color} onChange={(e) => setColor(e.target.value)} />;
```

**Complex effect with side effects:**
```tsx
useEffect(() => {
  setName(user.name);
  logAnalytics(user); // Other side effect
}, [user]);
```

**State modified elsewhere:**
```tsx
const [value, setValue] = useState(prop);
useEffect(() => setValue(prop), [prop]);
const handleClick = () => setValue(newValue); // Modified by user
```

## File Structure

```
internal/rules/
├─ [rule-name].go             # Main implementation
├─ [rule-name]_test.go        # Tests
└─ state_analysis_helpers.go  # Shared helpers (if needed)

test/fixtures/
├─ derived-state-simple.tsx
├─ derived-state-computed.tsx
└─ valid-state-interactive.tsx
```

## Expected Output

```bash
src/Component.tsx
  5:9 - [rule-name] State 'userName' is derived from prop 'user.name' and causes unnecessary re-renders
    Suggestion: Replace with: const userName = user.name;

✖ Found 1 issue in 1 file
```

## Dependencies

- ✅ Parser with hook detection
- ✅ AST walking utilities
- ✅ Prop identification helpers (IsPropIdentifier from stability_helpers.go)
- ⚠️ Need: State usage analysis
- ⚠️ Need: Effect body analysis

## Success Criteria

1. Detects simple prop mirroring
2. Detects derived property access
3. Detects computed derived values
4. Avoids false positives for interactive state
5. Avoids false positives for complex effects
6. Provides clear, actionable suggestions
7. Comprehensive test coverage
8. Documentation in README and rule docs

## Next Steps

1. ✅ Finalize rule name
2. Create test fixtures
3. Implement Phase 1 (basic detection)
4. Add tests and verify
5. Implement Phases 2-4 incrementally
6. Update README and registry
7. Ship!

---

See `docs/rules/no-derived-state.md` for full rule documentation.
