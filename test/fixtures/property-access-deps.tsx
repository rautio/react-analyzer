import React, { useEffect } from 'react';

// Test property access in dependency arrays
function ComponentWithPropertyAccess() {
  // Object created in render
  const config = { theme: 'dark', mode: 'normal' };
  const user = { name: 'John', age: 30 };

  // Should detect: accessing property of object created in render
  useEffect(() => {
    console.log('Theme:', config.theme);
  }, [config.theme]); // Should be flagged

  // Should detect: optional chaining
  useEffect(() => {
    console.log('User:', user?.name);
  }, [user?.name]); // Should be flagged

  // Should detect: nested property access
  const data = { settings: { theme: { color: 'blue' } } };
  useEffect(() => {
    console.log('Color:', data.settings.theme.color);
  }, [data.settings.theme.color]); // Should be flagged

  return <div>Test</div>;
}

export default ComponentWithPropertyAccess;
