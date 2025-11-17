import React, { useCallback, useState } from 'react';

// ❌ VIOLATIONS: Object expressions
export function ObjectProps() {
  return (
    <>
      <Component config={{ theme: 'dark' }} />
      <Settings options={{ enabled: true }} />
      <div style={{ color: 'red', fontSize: 16 }} />
    </>
  );
}

// ❌ VIOLATIONS: Array expressions
export function ArrayProps() {
  return (
    <>
      <List items={[1, 2, 3, 4, 5]} />
      <Tags tags={['react', 'javascript']} />
      <Menu options={['File', 'Edit', 'View']} />
    </>
  );
}

// ❌ VIOLATIONS: Arrow functions
export function FunctionProps() {
  const [count, setCount] = useState(0);
  const handleSubmit = () => {};
  const setValue = (v) => {};

  return (
    <>
      <Button onClick={() => setCount(count + 1)} />
      <Form onSubmit={() => handleSubmit()} />
      <Input onChange={(e) => setValue(e.target.value)} />
    </>
  );
}

// ❌ VIOLATIONS: Function expressions
export function FunctionExpressions() {
  const click = () => {};

  return (
    <>
      <Component onClick={function() { click(); }} />
      <Form onSubmit={function handleSubmit() {}} />
    </>
  );
}

// ❌ VIOLATIONS: Nested inline values
export function NestedInline() {
  const click = () => {};

  return (
    <Component
      data={{
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
  const name = "John";

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
