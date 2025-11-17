import React, { useEffect, useMemo, useCallback, useState } from 'react';

// Constants outside component - these are stable references
const GLOBAL_CONFIG = { theme: 'dark', mode: 'normal' };
const ITEMS = ['item1', 'item2', 'item3'];

// This file shows VALID patterns for using objects in dependency arrays
function ComponentWithValidObjectDeps() {
  const [count, setCount] = useState(0);
  const [name, setName] = useState('');

  // GOOD: Using state object (useState provides stable reference)
  const [user, setUser] = useState({ name: 'John', age: 30 });

  useEffect(() => {
    console.log('User changed:', user);
  }, [user]); // Valid - user is from useState

  // GOOD: Using useMemo for object
  const config = useMemo(() => ({
    theme: 'dark',
    mode: 'normal'
  }), []);

  useEffect(() => {
    console.log('Config:', config);
  }, [config]); // Valid - config is memoized

  // GOOD: Using useMemo with dependencies
  const options = useMemo(() => ({
    enabled: count > 0,
    timeout: 5000
  }), [count]);

  const memoized = useMemo(() => {
    return count * 2;
  }, [options]); // Valid - options is memoized

  // GOOD: Using global constant
  useEffect(() => {
    console.log('Global config:', GLOBAL_CONFIG);
  }, [GLOBAL_CONFIG]); // Valid - stable reference from outside component

  // GOOD: Using global constant array
  useEffect(() => {
    console.log('Items:', ITEMS);
  }, [ITEMS]); // Valid - stable reference

  // GOOD: Primitive values only
  useEffect(() => {
    console.log('Count and name:', count, name);
  }, [count, name]); // Valid - primitives

  // GOOD: Empty deps
  useEffect(() => {
    console.log('Mount only');
  }, []);

  // GOOD: useCallback with primitives
  const handleClick = useCallback(() => {
    console.log('Clicked:', count);
  }, [count]);

  return (
    <div onClick={handleClick}>
      Count: {count}, Memoized: {memoized}
    </div>
  );
}

export default ComponentWithValidObjectDeps;
