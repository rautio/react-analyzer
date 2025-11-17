import React, { useState, useEffect } from 'react';

// ❌ VIOLATION: Simple prop mirroring
export function SimpleMirror({ user }) {
  const [name, setName] = useState(user);

  useEffect(() => {
    setName(user);
  }, [user]);

  return <div>{name}</div>;
}

// ❌ VIOLATION: Multiple prop mirrors
export function MultipleMirrors({ firstName, lastName }) {
  const [first, setFirst] = useState(firstName);
  const [last, setLast] = useState(lastName);

  useEffect(() => {
    setFirst(firstName);
  }, [firstName]);

  useEffect(() => {
    setLast(lastName);
  }, [lastName]);

  return <div>{first} {last}</div>;
}

// ❌ VIOLATION: Inline effect
export function InlineEffect({ count }) {
  const [value, setValue] = useState(count);

  useEffect(() => setValue(count), [count]);

  return <div>{value}</div>;
}
