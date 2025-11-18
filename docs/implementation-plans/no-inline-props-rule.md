# Implementation Plan: no-inline-props

## Overview

Detect when JSX props are set to inline objects, arrays, or functions that are created on every render, causing unnecessary re-renders and breaking memoization.

**Severity:** Warning
**Type:** Single-file analysis
**Estimated Effort:** 3-4 hours (Phase 1)

---

## Detection Strategy

### Core Pattern

```tsx
// ❌ Detect: Inline values as props
<Component
  config={{ x: 1 }}              // object_expression
  items={[1, 2, 3]}              // array
  onClick={() => click()}        // arrow_function
  onSubmit={function() {}}       // function
/>

// ✅ Allow: Identifiers, literals, expressions
<Component
  config={CONFIG}                // identifier (assumed stable)
  count={42}                     // number literal
  name="test"                    // string literal
  enabled={true}                 // boolean literal
  total={count + 1}              // binary_expression (primitive)
/>
```

---

## Implementation Phases

### Phase 1: Basic Detection (3-4 hours)

**Goal:** Detect inline objects, arrays, and functions in JSX props

**What to detect:**
1. **Object expressions:** `prop={{ key: value }}`
2. **Array expressions:** `prop={[1, 2, 3]}`
3. **Arrow functions:** `prop={() => func()}`
4. **Function expressions:** `prop={function() {}}`

**Algorithm:**

```
FUNCTION detectInlineProps(ast):
    issues = []

    // Find all JSX elements
    FOR EACH jsx_element in ast:
        // Get all attributes (props)
        FOR EACH jsx_attribute in jsx_element:
            propName = jsx_attribute.name
            propValue = jsx_attribute.value

            // Skip if no value (boolean prop like <Input disabled />)
            IF propValue is null:
                CONTINUE

            // Get the expression inside {}
            IF propValue is jsx_expression_container:
                expression = propValue.expression

                // Check if expression is inline object/array/function
                IF isInlineValue(expression):
                    issues.add({
                        propName: propName,
                        valueType: expression.type,
                        line: jsx_attribute.line
                    })

    RETURN issues


FUNCTION isInlineValue(expression):
    expressionType = expression.type

    // Inline object literal
    IF expressionType == "object_expression":
        RETURN true

    // Inline array literal
    IF expressionType == "array":
        RETURN true

    // Arrow function
    IF expressionType == "arrow_function":
        RETURN true

    // Function expression
    IF expressionType == "function":
        RETURN true

    RETURN false
```

---

## AST Patterns

### JSX Element Structure

```tsx
<Component prop={value} />
```

**AST:**
```
jsx_element {
  opening_element: jsx_opening_element {
    name: jsx_identifier "Component"
    attribute: jsx_attribute {           // ← The prop
      name: property_identifier "prop"
      value: jsx_expression_container {  // ← The value
        expression: [EXPRESSION]         // ← Check this!
      }
    }
  }
}
```

### Expression Types to Detect

**1. Object Expression:**
```tsx
<Component config={{ theme: 'dark' }} />
```
AST: `jsx_expression_container { expression: object_expression }`

**2. Array:**
```tsx
<Component items={[1, 2, 3]} />
```
AST: `jsx_expression_container { expression: array }`

**3. Arrow Function:**
```tsx
<Component onClick={() => click()} />
```
AST: `jsx_expression_container { expression: arrow_function }`

**4. Function Expression:**
```tsx
<Component onSubmit={function() {}} />
```
AST: `jsx_expression_container { expression: function }`

### What NOT to Flag

**1. Identifiers (variables):**
```tsx
<Component config={CONFIG} />
```
AST: `jsx_expression_container { expression: identifier "CONFIG" }`
→ SKIP (could be stable constant)

**2. Literals:**
```tsx
<Component count={42} name="test" enabled={true} />
```
AST:
- String: `jsx_attribute { value: string "test" }` (no expression container)
- Number: `jsx_expression_container { expression: number "42" }`
- Boolean: `jsx_expression_container { expression: true/false }`
→ SKIP (primitives are always stable)

**3. Expressions (arithmetic, etc):**
```tsx
<Component total={count + 1} />
```
AST: `jsx_expression_container { expression: binary_expression }`
→ SKIP (primitive value, not reference type)

**4. Special props:**
```tsx
<Component key="id" className="container" />
```
→ SKIP (key is special, className is string)

---

## Helper Functions

```go
// Find all JSX elements in the AST
func findJSXElements(node *parser.Node) []*parser.Node

// Get attributes from JSX opening element
func getJSXAttributes(openingElement *parser.Node) []*parser.Node

// Check if attribute value is inline object/array/function
func isInlineValue(expression *parser.Node) bool {
    expressionType := expression.Type()

    switch expressionType {
    case "object_expression":
        return true
    case "array":
        return true
    case "arrow_function":
        return true
    case "function":
        return true
    default:
        return false
    }
}

// Get a friendly name for the inline value type
func getValueTypeName(expressionType string) string {
    switch expressionType {
    case "object_expression":
        return "object"
    case "array":
        return "array"
    case "arrow_function", "function":
        return "function"
    default:
        return "value"
    }
}
```

---

## Test Cases

### Test Fixture: no-inline-props-simple.tsx

```tsx
import React, { useCallback, useState } from 'react';

// ❌ VIOLATIONS: Object expressions
export function ObjectProps() {
  return (
    <>
      <Component config={{ theme: 'dark' }} />              // Line 7
      <Settings options={{ enabled: true }} />              // Line 8
      <div style={{ color: 'red', fontSize: 16 }} />        // Line 9
    </>
  );
}

// ❌ VIOLATIONS: Array expressions
export function ArrayProps() {
  return (
    <>
      <List items={[1, 2, 3, 4, 5]} />                      // Line 18
      <Tags tags={['react', 'javascript']} />                // Line 19
      <Menu options={['File', 'Edit', 'View']} />           // Line 20
    </>
  );
}

// ❌ VIOLATIONS: Arrow functions
export function FunctionProps() {
  const [count, setCount] = useState(0);

  return (
    <>
      <Button onClick={() => setCount(count + 1)} />        // Line 30
      <Form onSubmit={() => handleSubmit()} />              // Line 31
      <Input onChange={(e) => setValue(e.target.value)} />  // Line 32
    </>
  );
}

// ❌ VIOLATIONS: Function expressions
export function FunctionExpressions() {
  return (
    <>
      <Component onClick={function() { click(); }} />       // Line 41
      <Form onSubmit={function handleSubmit() {}} />        // Line 42
    </>
  );
}

// ❌ VIOLATIONS: Nested inline values
export function NestedInline() {
  return (
    <Component
      data={{                                                // Line 51
        items: [1, 2, 3],
        handler: () => click()
      }}
    />
  );
}

// ✅ VALID: Identifiers and literals
const CONFIG = { theme: 'dark' };
const ITEMS = [1, 2, 3];

export function ValidProps() {
  const [count, setCount] = useState(0);
  const handleClick = useCallback(() => {}, []);

  return (
    <>
      {/* Identifiers - OK */}
      <Component config={CONFIG} />
      <List items={ITEMS} />
      <Button onClick={handleClick} />

      {/* Primitives - OK */}
      <Input value="hello" />
      <Counter count={42} />
      <Toggle enabled={true} />

      {/* Expressions (primitives) - OK */}
      <Progress value={count * 2} />
      <Text content={name + " Smith"} />
    </>
  );
}
```

**Expected violations:** 11
- ObjectProps: 3 violations
- ArrayProps: 3 violations
- FunctionProps: 3 violations
- FunctionExpressions: 2 violations
- NestedInline: 1 violation (outer object, might catch nested too)

---

### Test Fixture: no-inline-props-valid.tsx

```tsx
import React, { useCallback, useMemo } from 'react';

// ✅ Constants outside component
const CONFIG = { theme: 'dark' };
const ITEMS = [1, 2, 3, 4, 5];
const STYLE = { padding: 20 };

export function StableProps() {
  return (
    <>
      <Component config={CONFIG} />
      <List items={ITEMS} />
      <div style={STYLE} />
    </>
  );
}

export function MemoizedProps() {
  const [filter, setFilter] = useState('');

  const handleChange = useCallback((e) => {
    setFilter(e.target.value);
  }, []);

  const filteredItems = useMemo(() => {
    return ITEMS.filter(x => x.includes(filter));
  }, [filter]);

  return (
    <>
      <Input onChange={handleChange} />
      <List items={filteredItems} />
    </>
  );
}

export function PrimitiveLiterals() {
  return (
    <>
      <Input value="hello" />
      <Counter count={42} />
      <Toggle enabled={true} />
      <Icon size={16} />
    </>
  );
}

export function Identifiers() {
  const theme = useTheme();
  const items = useItems();

  return (
    <>
      <Settings theme={theme} />
      <List items={items} />
    </>
  );
}
```

**Expected violations:** 0

---

## Message Format

```
Prop '{propName}' receives an inline {valueType}, creating a new reference every render. Extract to a constant or use {suggestion}.

Examples:
- Prop 'config' receives an inline object, creating a new reference every render. Extract to a constant or use useMemo.
- Prop 'onClick' receives an inline function, creating a new reference every render. Extract to a constant or use useCallback.
- Prop 'items' receives an inline array, creating a new reference every render. Extract to a constant or use useMemo.
```

---

## Edge Cases

### Should Flag

1. **Nested inline values**
   ```tsx
   <Component data={{ items: [1, 2], fn: () => {} }} />
   // Flag the outer object (contains inline array and function)
   ```

2. **Inline styles**
   ```tsx
   <div style={{ color: 'red' }} />
   // Flag - new object every render
   ```

3. **Template literals with expressions (objects/functions inside)**
   - For Phase 1, skip template literals
   - They're usually strings, which are primitives

### Should NOT Flag

1. **String literals**
   ```tsx
   <Component name="test" />
   // Don't flag - strings are primitives
   ```

2. **Number/Boolean literals**
   ```tsx
   <Component count={42} enabled={true} />
   // Don't flag - primitives
   ```

3. **Spread attributes**
   ```tsx
   <Component {...props} />
   // Don't flag for Phase 1 (complex to analyze)
   ```

4. **Special props**
   ```tsx
   <Component key="unique" ref={ref} />
   // Don't flag key/ref
   ```

---

## Success Criteria

**Phase 1:**
- ✅ Detect all inline objects, arrays, and functions in JSX props
- ✅ Zero false positives on identifiers and literals
- ✅ All tests passing
- ✅ Clear, actionable error messages

---

## Future Enhancements (Post-MVP)

### Phase 2: Smart Suggestions
- Detect if `useCallback` or `useMemo` is imported
- Suggest specific fix based on value type
- Show example of correct pattern

### Phase 3: Exception Rules
- Allow inline values for certain props (e.g., `style` for small components)
- Configuration for allowlist/denylist
- Detect if child is memoized (higher priority if so)

---

## Estimated Timeline

**Phase 1:** 3-4 hours
- 1 hour: Core detection logic
- 1 hour: Test fixtures and tests
- 1 hour: Message formatting and edge cases
- 0.5-1 hour: Testing and debugging

**Total:** 3-4 hours

---

## Related Rules

- `unstable-props-to-memo` - Detects unstable props to memoized components (cross-file)
- This rule is simpler: just checks if props ARE inline, regardless of receiver
