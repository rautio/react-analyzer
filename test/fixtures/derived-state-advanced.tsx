import React, { useState, useEffect } from 'react';

// ❌ VIOLATION: Nested property mirroring
export function NestedProperty({ user }) {
  const [name, setName] = useState(user.name);

  useEffect(() => {
    setName(user.name);
  }, [user.name]);

  return <div>{name}</div>;
}

// ❌ VIOLATION: Deep nested property
export function DeepNested({ data }) {
  const [value, setValue] = useState(data.settings.theme.color);

  useEffect(() => {
    setValue(data.settings.theme.color);
  }, [data.settings.theme.color]);

  return <div>{value}</div>;
}

// ❌ VIOLATION: Computed value
export function ComputedValue({ items }) {
  const [count, setCount] = useState(items.length);

  useEffect(() => {
    setCount(items.length);
  }, [items.length]);

  return <div>Count: {count}</div>;
}

// ❌ VIOLATION: Function call
export function FunctionCall({ items }) {
  const [total, setTotal] = useState(calculateTotal(items));

  useEffect(() => {
    setTotal(calculateTotal(items));
  }, [items]);

  return <div>Total: {total}</div>;
}

// ❌ VIOLATION: Template literal with prop
export function TemplateLiteral({ name }) {
  const [greeting, setGreeting] = useState(`Hello, ${name}`);

  useEffect(() => {
    setGreeting(`Hello, ${name}`);
  }, [name]);

  return <div>{greeting}</div>;
}

// ✅ VALID: Complex expression that's not just prop mirroring
export function ComplexComputation({ items, multiplier }) {
  const [result, setResult] = useState(items.length * multiplier);

  useEffect(() => {
    // This is more complex - involves multiple props and computation
    // Phase 2 might still flag this, but Phase 3 will handle better
    setResult(items.length * multiplier);
  }, [items.length, multiplier]);

  return <div>{result}</div>;
}

function calculateTotal(items: any[]) {
  return items.reduce((sum, item) => sum + item.price, 0);
}
