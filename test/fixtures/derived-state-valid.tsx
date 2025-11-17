import React, { useState, useEffect } from 'react';

// ✅ VALID: Initial value only, no useEffect syncing
export function InitialValueOnly({ initialColor }) {
  const [color, setColor] = useState(initialColor);

  // User can modify independently
  return (
    <input
      type="color"
      value={color}
      onChange={(e) => setColor(e.target.value)}
    />
  );
}

// ✅ VALID: State is modified in event handler
export function InteractiveState({ defaultValue }) {
  const [value, setValue] = useState(defaultValue);

  useEffect(() => {
    setValue(defaultValue);
  }, [defaultValue]);

  // State is modified by user - this is intentional!
  const handleChange = (e) => {
    setValue(e.target.value);
  };

  return <input value={value} onChange={handleChange} />;
}

// ✅ VALID: Effect has other side effects
export function ComplexEffect({ user }) {
  const [name, setName] = useState(user.name);

  useEffect(() => {
    setName(user.name);
    // Other side effect - analytics, logging, etc.
    console.log('User changed:', user);
    trackAnalytics(user);
  }, [user]);

  return <div>{name}</div>;
}

// ✅ VALID: Just deriving directly (no useState)
export function DirectDerivation({ user }) {
  const name = user.name; // This is the correct pattern!
  return <div>{name}</div>;
}

function trackAnalytics(user: any) {
  // Mock function
}
