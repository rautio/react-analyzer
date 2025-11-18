import React from 'react';

// Test React.useState with stale state issue
function CounterWithNamespace() {
  const [count, setCount] = React.useState(0);

  const increment = () => {
    setCount(count + 1); // Should detect: stale state
  };

  return <button onClick={increment}>{count}</button>;
}

// Test React.useEffect with derived state issue
function DerivedWithNamespace({ user }: { user: string }) {
  const [name, setName] = React.useState(user);

  React.useEffect(() => {
    setName(user);
  }, [user]);

  return <div>{name}</div>;
}

// Test React.useMemo with object dependencies
function MemoWithNamespace() {
  const config = { theme: 'dark' };

  const computed = React.useMemo(() => {
    return config.theme.toUpperCase();
  }, [config]); // Should detect: object in dependencies

  return <div>{computed}</div>;
}

// Test mixed bare and namespaced hooks
function MixedHooks() {
  const [count, setCount] = React.useState(0);
  const [items, setItems] = React.useState<string[]>([]);

  React.useEffect(() => {
    console.log('Count changed');
  }, [count]);

  const addItem = () => {
    setItems([...items, 'new']); // Should detect: stale state
  };

  return (
    <div>
      <button onClick={() => setCount(count + 1)}>{count}</button>
      <button onClick={addItem}>Add</button>
    </div>
  );
}

// Test inline props with namespaced hooks
function InlinePropsWithNamespace() {
  const [data, setData] = React.useState({});

  return (
    <div>
      <Child config={{ theme: 'dark' }} /> {/* Should detect: inline object */}
      <Child items={[1, 2, 3]} /> {/* Should detect: inline array */}
      <Child onClick={() => console.log('click')} /> {/* Should detect: inline function */}
    </div>
  );
}

function Child(props: any) {
  return <div />;
}
