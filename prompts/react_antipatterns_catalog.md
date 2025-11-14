# React Anti-Patterns Catalog

**Date:** 2025-11-13
**Purpose:** Identify tough anti-patterns for MVP scope - focus on patterns hard for linters to catch

---

## Anti-Pattern Classification

| # | Anti-Pattern | Description | Existing ESLint Rule | Requires Cross-File Analysis | Requires Component Tree | Difficulty to Detect | Priority for MVP |
|---|-------------|-------------|---------------------|----------------------------|----------------------|-------------------|-----------------|
| **PERFORMANCE - RE-RENDERS** ||||||||
| 1 | **Inline Object/Array Props** | Creating new object/array literals in JSX props causes child re-renders even with React.memo | ⚠️ Partial (`react/jsx-no-bind` warns about functions only) | ❌ No | ✅ Yes (need to know if child is memoized) | 🟡 Medium | ⭐⭐⭐ HIGH |
| 2 | **Inline Function Props** | Passing inline arrow functions as props breaks memoization | ⚠️ Partial (`react/jsx-no-bind`) | ❌ No | ✅ Yes (need to know if child is memoized) | 🟡 Medium | ⭐⭐⭐ HIGH |
| 3 | **Unstable Dependencies in Hooks** | Objects/arrays as dependencies in useEffect/useMemo cause infinite re-runs | ❌ No comprehensive rule | ❌ No | ❌ No | 🟡 Medium | ⭐⭐⭐ HIGH |
| 4 | **Missing React.memo on Expensive Components** | Components with high render complexity re-render unnecessarily | ❌ No | ❌ No | ⚠️ Helpful (to measure cost) | 🟡 Medium | ⭐⭐ MEDIUM |
| 5 | **Memoized Component with Unstable Props** | React.memo is useless when parent passes unstable props | ❌ No | ✅ **YES** | ✅ **YES** | 🔴 **HARD** | ⭐⭐⭐ **HIGH** |
| 6 | **Creating Components Inside Render** | Defining components inside render functions causes full remount (slowest) | ⚠️ `react/no-unstable-nested-components` | ❌ No | ❌ No | 🟢 Easy | ⭐ LOW (ESLint catches it) |
| 7 | **Context Value Not Memoized** | Context provider value changes on every render, re-rendering all consumers | ❌ No | ❌ No | ⚠️ Helpful (find all consumers) | 🟡 Medium | ⭐⭐ MEDIUM |
| 8 | **Over-Memoization** | Using useMemo/useCallback where cost > benefit | ❌ No | ❌ No | ❌ No | 🔴 Hard (requires profiling) | ⭐ LOW (philosophical) |
| 9 | **Cascading Re-renders** | Parent state change triggers chain of child re-renders | ❌ No | ✅ **YES** | ✅ **YES** | 🔴 **HARD** | ⭐⭐⭐ **HIGH** |
| 10 | **Partial Memoization** | Memoizing component but not all its props (breaks memoization) | ❌ No | ✅ **YES** | ✅ **YES** | 🔴 **HARD** | ⭐⭐⭐ **HIGH** |
| **HOOKS - CORRECTNESS** ||||||||
| 11 | **Exhaustive Dependencies** | Missing dependencies in useEffect/useMemo/useCallback dependency arrays | ✅ `react-hooks/exhaustive-deps` | ❌ No | ❌ No | 🟢 Easy | ⭐ LOW (ESLint solves it) |
| 12 | **Infinite useEffect Loop** | useEffect updates state that triggers itself | ⚠️ Partial (exhaustive-deps helps) | ❌ No | ❌ No | 🟡 Medium | ⭐⭐ MEDIUM |
| 13 | **useEffect for Derived State** | Using useEffect to sync state when it should be computed during render | ❌ No | ❌ No | ❌ No | 🟡 Medium | ⭐⭐ MEDIUM |
| 14 | **Stale Closure** | useCallback/useMemo references outdated values due to missing deps | ⚠️ Caught by exhaustive-deps | ❌ No | ❌ No | 🟢 Easy | ⭐ LOW (ESLint solves it) |
| 15 | **Conditional Hook Calls** | Calling hooks inside conditions/loops (violates Rules of Hooks) | ✅ `react-hooks/rules-of-hooks` | ❌ No | ❌ No | 🟢 Easy | ⭐ LOW (ESLint solves it) |
| 16 | **Non-Primitive Hook Dependencies** | Including objects/arrays in dependency array without memoization | ❌ No | ❌ No | ❌ No | 🟡 Medium | ⭐⭐⭐ HIGH |
| **STATE MANAGEMENT** ||||||||
| 17 | **Props Drilling (3+ levels)** | Passing props through multiple components that don't use them | ❌ No | ✅ **YES** | ✅ **YES** | 🔴 **HARD** | ⭐⭐⭐ **HIGH** |
| 18 | **Direct State Mutation** | Mutating state directly instead of using setState | ⚠️ `no-param-reassign` (generic) | ❌ No | ❌ No | 🟡 Medium | ⭐⭐ MEDIUM |
| 19 | **Derived State Not Computed** | Storing derived state instead of computing from source | ❌ No | ❌ No | ❌ No | 🟡 Medium | ⭐ LOW |
| 20 | **Over-Reliance on useState** | Using state for values that don't need re-renders (should be useRef) | ❌ No | ❌ No | ❌ No | 🟡 Medium | ⭐ LOW |
| **JSX - PERFORMANCE** ||||||||
| 21 | **Inline Styles** | Style objects created on every render | ⚠️ `react/forbid-dom-props` (can ban style) | ❌ No | ❌ No | 🟢 Easy | ⭐⭐ MEDIUM |
| 22 | **Missing Keys in Lists** | Missing or improper keys in .map() | ✅ `react/jsx-key` | ❌ No | ❌ No | 🟢 Easy | ⭐ LOW (ESLint solves it) |
| 23 | **Index as Key** | Using array index as key (causes issues on reorder) | ⚠️ Warned by `react/no-array-index-key` | ❌ No | ❌ No | 🟢 Easy | ⭐ LOW (ESLint warns) |
| 24 | **Large Component Rendering Small Update** | Rendering large lists when only one item changed | ❌ No | ❌ No | ⚠️ Helpful | 🔴 Hard | ⭐ LOW (needs profiling) |
| **ADVANCED - CROSS-FILE** ||||||||
| 25 | **Unstable Import Breaking Child Memo** | Importing non-memoized function and passing to memoized child | ❌ No | ✅ **YES** | ✅ **YES** | 🔴 **HARD** | ⭐⭐⭐ **HIGH** |
| 26 | **Context Provider Too High** | Context provider at root when only small subtree needs it | ❌ No | ✅ YES | ✅ YES | 🔴 Hard | ⭐⭐ MEDIUM |
| 27 | **Shared State Causing Unrelated Re-renders** | Single state object shared by unrelated components | ❌ No | ✅ YES | ✅ YES | 🔴 Hard | ⭐⭐ MEDIUM |

---

## Legend

**Existing ESLint Rule:**
- ✅ Fully covered by ESLint
- ⚠️ Partially covered (limited detection)
- ❌ No ESLint rule

**Requires Cross-File Analysis:**
- ✅ YES - Must analyze multiple files
- ⚠️ Helpful - Single-file analysis works but cross-file improves accuracy
- ❌ No - Can detect in single file

**Requires Component Tree:**
- ✅ YES - Must understand parent-child relationships
- ⚠️ Helpful - Improves detection accuracy
- ❌ No - Can detect without tree

**Difficulty to Detect:**
- 🟢 Easy - Pattern matching on AST
- 🟡 Medium - Requires data flow analysis
- 🔴 Hard - Requires cross-file + semantic analysis

**Priority for MVP:**
- ⭐⭐⭐ HIGH - Core value, hard for linters
- ⭐⭐ MEDIUM - Valuable but less critical
- ⭐ LOW - Already solved or not critical

---

## MVP Recommendations: Tough Anti-Patterns

### Tier 1: Core Value Proposition (Must-Have)

These are **hard for linters** and provide **immediate value**:

| # | Anti-Pattern | Why It's Hard | Why It Matters |
|---|-------------|---------------|----------------|
| **5** | **Memoized Component with Unstable Props** | Requires analyzing parent (prop source) + child (memo status) across files | React.memo is **completely useless** if this happens - huge performance waste |
| **9** | **Cascading Re-renders** | Requires full component tree to trace state flow | Single state change can cause 10+ components to re-render unnecessarily |
| **10** | **Partial Memoization** | Need to verify ALL props are stable, requires cross-file analysis | "Breaks the memoization chain" - one unstable prop ruins everything |
| **17** | **Props Drilling (3+ levels)** | Requires tracing prop path through component tree | Makes code brittle, causes unnecessary re-renders in intermediate components |
| **25** | **Unstable Import Breaking Child Memo** | Requires resolving imports + checking memoization in other file | Subtle bug - child looks memoized but parent passes unstable imported value |

### Tier 2: High-Value Add-Ons (Should-Have)

| # | Anti-Pattern | Why Include |
|---|-------------|-------------|
| **1** | **Inline Object/Array Props** | Easy to detect, very common, high performance impact |
| **2** | **Inline Function Props** | Easy to detect, very common, breaks memoization |
| **3** | **Unstable Dependencies in Hooks** | Not well-covered by ESLint, causes infinite loops |
| **16** | **Non-Primitive Hook Dependencies** | Closely related to #3, common mistake |

### Tier 3: Nice-to-Have (Future)

| # | Anti-Pattern | Why Later |
|---|-------------|-----------|
| **4** | **Missing React.memo on Expensive Components** | Requires complexity scoring, less critical |
| **7** | **Context Value Not Memoized** | Moderate impact, less common |
| **13** | **useEffect for Derived State** | More of a code smell than performance issue |

---

## Detailed Analysis: Top 5 MVP Rules

### Rule 1: Memoized Component with Unstable Props ⭐⭐⭐

**Example:**
```typescript
// Child.tsx
export const Child = React.memo(({ config }) => {
  // This component is memoized...
  return <div>{config.theme}</div>;
});

// Parent.tsx
function Parent() {
  return <Child config={{ theme: 'dark' }} />;  // ← But parent passes new object every render!
}
```

**Why ESLint Can't Catch It:**
- Need to analyze `Parent.tsx` and `Child.tsx` simultaneously
- Must resolve import to know Child is memoized
- Must detect inline object in Parent's JSX

**Detection Algorithm:**
```
For each JSX element:
  1. Resolve component (follow import)
  2. Check if component is wrapped in React.memo
  3. If yes, check each prop in current file
  4. If prop is inline object/array/function → ERROR
```

**Impact:** High - React.memo overhead with zero benefit

---

### Rule 2: Cascading Re-renders ⭐⭐⭐

**Example:**
```typescript
// App.tsx
function App() {
  const [user, setUser] = useState();
  return <Dashboard user={user} />;
}

// Dashboard.tsx
function Dashboard({ user }) {
  return <Sidebar user={user} />;  // Passes user down
}

// Sidebar.tsx
function Sidebar({ user }) {
  return <UserProfile user={user} />;  // Passes user down again
}

// UserProfile.tsx
function UserProfile({ user }) {
  return <div>{user.name}</div>;  // Finally uses it!
}
```

**Issue:** When `user` changes in App:
- App re-renders ✅ (necessary)
- Dashboard re-renders ❌ (unnecessary - just passes prop)
- Sidebar re-renders ❌ (unnecessary - just passes prop)
- UserProfile re-renders ✅ (necessary - uses the data)

**Why ESLint Can't Catch It:**
- Requires building component hierarchy across 4 files
- Must trace data flow through intermediate components
- Must identify components that "just pass through" props

**Detection Algorithm:**
```
Build component tree
For each component:
  For each prop:
    Trace prop from source to final consumer
    If path.length >= 3:
      Identify intermediate components (don't use prop, just pass it)
      → WARN: Props drilling, suggest Context or component composition
```

**Impact:** Critical - Can cause 10x more renders than necessary

---

### Rule 3: Partial Memoization ⭐⭐⭐

**Example:**
```typescript
// Child.tsx
export const ExpensiveChild = React.memo(({ data, onUpdate }) => {
  // Memoized component
  return <div onClick={onUpdate}>{data.value}</div>;
});

// Parent.tsx
function Parent() {
  const data = useMemo(() => ({ value: 'test' }), []);  // ✅ Memoized

  const handleUpdate = () => console.log('update');  // ❌ NOT memoized!

  return <ExpensiveChild data={data} onUpdate={handleUpdate} />;
}
```

**Issue:**
- `data` prop is stable (useMemo)
- `onUpdate` prop is unstable (new function every render)
- React.memo compares **all props** → always returns false → always re-renders
- The useMemo effort is wasted!

**Why ESLint Can't Catch It:**
- Need to analyze Child (it's memoized)
- Need to analyze Parent (prop stability)
- Need to cross-reference which props are stable vs unstable

**Detection Algorithm:**
```
For each React.memo component:
  Find all JSX elements that render it (may be in other files)
  For each render site:
    Check each prop:
      - Inline object/array/function? → UNSTABLE
      - useMemo/useCallback result? → STABLE
      - Constant/import? → STABLE
      - Regular variable? → UNSTABLE

    If ANY prop is unstable:
      → WARN: "Partial memoization - React.memo is useless here"
```

**Impact:** High - Common mistake, wastes memoization effort

---

### Rule 4: Props Drilling (3+ levels) ⭐⭐⭐

**Covered above in Rule 2 example**

**Additional Detection:**
```
For each prop path:
  If depth >= 3:
    Count intermediate components (don't destructure or use the prop)
    → SUGGEST: "Consider Context API or component composition"
    → SHOW: Full path visualization (A → B → C → D)
```

**Auto-fix Suggestion:**
```typescript
// Option 1: Context
const UserContext = createContext();

// Option 2: Component Composition
function App() {
  return (
    <Dashboard>
      <UserProfile user={user} />  {/* Skip intermediate components */}
    </Dashboard>
  );
}
```

---

### Rule 5: Unstable Import Breaking Child Memo ⭐⭐⭐

**Example:**
```typescript
// utils.ts
export const formatUser = (user) => user.name.toUpperCase();  // Not memoized

// UserProfile.tsx
export const UserProfile = React.memo(({ user, formatter }) => {
  return <div>{formatter(user)}</div>;
});

// Dashboard.tsx
import { formatUser } from './utils';

function Dashboard({ user }) {
  return <UserProfile user={user} formatter={formatUser} />;  // ✅ Looks stable
}
```

**Issue:**
- `formatUser` import **looks** stable
- But it's a new reference on every render (function is re-evaluated)
- Breaks UserProfile's memoization

**Why ESLint Can't Catch It:**
- Need to resolve import (`formatUser` from `utils.ts`)
- Check if it's wrapped in `useCallback` or defined outside component
- Cross-reference with child component's memoization

**Detection Algorithm:**
```
For each prop passed to React.memo component:
  If prop is identifier (not inline):
    Trace to definition:
      - If import → resolve to source file
        - If source is regular function → UNSTABLE
        - If source is memoized/constant → STABLE
      - If local variable → check if memoized
      - If parameter → check if stable

  If unstable prop to memoized component:
    → WARN: "Imported function breaks memoization"
```

**Fix Suggestion:**
```typescript
// Option 1: Memoize in parent
const formatter = useCallback(formatUser, []);

// Option 2: Make it a constant
const FORMAT_USER = (user) => user.name.toUpperCase();
```

---

## What NOT to Build (Already Solved)

| Anti-Pattern | Why Skip |
|-------------|----------|
| Exhaustive Dependencies | `react-hooks/exhaustive-deps` already perfect |
| Conditional Hook Calls | `react-hooks/rules-of-hooks` already perfect |
| Missing Keys in Lists | `react/jsx-key` already perfect |
| Nested Components | `react/no-unstable-nested-components` already exists |

---

## Cross-File Analysis Requirements Summary

**Anti-patterns requiring cross-file analysis (your differentiator):**

1. ✅ Memoized component with unstable props (#5)
2. ✅ Cascading re-renders (#9)
3. ✅ Partial memoization (#10)
4. ✅ Props drilling (#17)
5. ✅ Unstable import breaking memo (#25)

**Anti-patterns solvable with single-file analysis (quick wins):**

1. ✅ Inline object/array/function props (#1, #2)
2. ✅ Unstable dependencies in hooks (#3)
3. ✅ Non-primitive hook dependencies (#16)

---

## MVP Scope Recommendation

### Phase 1 (Weeks 1-8): Core Cross-File Rules

**Build these 3 rules first (highest ROI):**

1. **Memoized Component with Unstable Props** (#5)
   - Cross-file: YES
   - Impact: Critical (makes React.memo useless)
   - Difficulty: Medium (import resolution + prop checking)

2. **Partial Memoization** (#10)
   - Cross-file: YES
   - Impact: High (common mistake)
   - Difficulty: Medium (stability analysis per prop)

3. **Props Drilling** (#17)
   - Cross-file: YES
   - Impact: High (code smell + performance)
   - Difficulty: Medium (path tracing)

### Phase 2 (Weeks 9-12): Single-File Quick Wins

**Add these 2 rules (easy + high value):**

4. **Inline Object/Array/Function Props** (#1, #2)
   - Cross-file: NO
   - Impact: High (very common)
   - Difficulty: Easy (AST pattern matching)

5. **Unstable Dependencies in Hooks** (#3, #16)
   - Cross-file: NO
   - Impact: High (causes infinite loops)
   - Difficulty: Medium (dependency analysis)

### Phase 3 (Weeks 13+): Advanced

6. **Cascading Re-renders** (#9)
7. **Unstable Import Breaking Memo** (#25)

---

## Success Metrics

**For each rule, measure:**

1. **Detection Rate:** % of real issues found
2. **False Positive Rate:** % of warnings that are incorrect
3. **Auto-fix Success Rate:** % of fixes that are correct
4. **Performance Impact:** Estimated re-renders saved

**Target:**
- Detection Rate: >80%
- False Positive Rate: <10%
- Auto-fix Success: >90%

---

## Next Steps

1. ✅ Validate these anti-patterns with your target users
2. ✅ Build POC for Rule #5 (Memoized Component with Unstable Props)
3. ✅ Measure false positive rate on real codebases
4. ✅ Refine detection algorithms based on feedback

---

## References

- [React Re-renders Guide](https://www.developerway.com/posts/react-re-renders-guide)
- [Fixing Memoization-Breaking Re-renders](https://blog.sentry.io/fixing-memoization-breaking-re-renders-in-react/)
- [React Hooks Common Mistakes](https://blog.logrocket.com/15-common-useeffect-mistakes-react)
- [React Anti-Patterns](https://reactantipatterns.com/)
