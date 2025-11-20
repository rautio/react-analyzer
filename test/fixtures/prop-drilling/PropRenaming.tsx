import React, { useState } from 'react';

/**
 * Test fixture for prop renaming during passing
 *
 * This demonstrates the scenario where a prop is renamed at different levels:
 * - Root defines onRemove
 * - Middle renames it to onSelectedFieldRemove
 * - Child renames it back to onRemove
 *
 * Expected behavior:
 * - Should detect this as a single prop drilling chain
 * - Should trace the full chain even with different names
 * - Should show renaming info in edge details
 */

interface ChildProps {
  onRemove: () => void;
}

// Leaf component that uses the prop
function Child({ onRemove }: ChildProps) {
  return (
    <button onClick={onRemove}>
      Remove Item
    </button>
  );
}

interface MiddleProps {
  onSelectedFieldRemove: () => void;
}

// Passthrough component that renames the prop
function Middle({ onSelectedFieldRemove }: MiddleProps) {
  // Just passes through, doesn't use it
  return <Child onRemove={onSelectedFieldRemove} />;
}

interface TopProps {
  onRemove: () => void;
}

// Another passthrough that renames the prop
function Top({ onRemove }: TopProps) {
  // Just passes through, doesn't use it
  return <Middle onSelectedFieldRemove={onRemove} />;
}

// Root component with the actual state
function Root() {
  const [items, setItems] = useState(['item1', 'item2']);

  const handleRemove = () => {
    setItems([]);
  };

  return <Top onRemove={handleRemove} />;
}

export default Root;
