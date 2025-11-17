import React, { useCallback, useMemo, useState } from 'react';

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
