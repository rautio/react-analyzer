import React, { useEffect, useMemo, useCallback, useState } from 'react';

// This file contains the REAL anti-pattern:
// Objects/arrays created in component body and used in dependency arrays
// These are recreated on every render, causing unnecessary re-runs

function ComponentWithObjectInRender() {
  const [count, setCount] = useState(0);
  const [user, setUser] = useState({ name: 'John' });

  // BAD: Object created in render, used in deps
  const config = { theme: 'dark', mode: 'normal' };

  useEffect(() => {
    console.log('Config changed:', config);
  }, [config]); // Line 15: violation - config is recreated every render

  // BAD: Array created in render, used in deps
  const items = ['item1', 'item2', 'item3'];

  useEffect(() => {
    console.log('Items:', items);
  }, [items]); // Line 22: violation - items recreated every render

  // BAD: Object created inline as variable
  const options = { enabled: true, timeout: 5000 };

  const memoized = useMemo(() => {
    return count * 2;
  }, [count, options]); // Line 29: violation - options recreated every render

  // BAD: Callback that creates object in render
  const handleClick = useCallback(() => {
    console.log('Clicked');
  }, [{ isEnabled: true }]); // Line 34: violation - inline object

  // BAD: Multiple objects in render
  const settings = { volume: 50 };
  const preferences = { autoplay: false };

  useEffect(() => {
    console.log('Settings changed');
  }, [settings, preferences]); // Line 42: violations - both recreated every render

  // BAD: Nested object in render
  const theme = {
    colors: { primary: 'blue', secondary: 'green' },
    spacing: { small: 8, large: 16 }
  };

  useEffect(() => {
    console.log('Theme:', theme);
  }, [theme]); // Line 51: violation

  return <div onClick={handleClick}>{memoized}</div>;
}

export default ComponentWithObjectInRender;
