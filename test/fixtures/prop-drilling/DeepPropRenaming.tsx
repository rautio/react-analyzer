import React, { useState } from 'react';

/**
 * Test fixture for prop renaming with enough layers to trigger violation
 *
 * This has 4 passthrough layers with renaming to exceed the threshold of 3.
 */

// Leaf component that displays the count
function Leaf({ value }: { value: number }) {
  return <div>Count: {value}</div>;
}

// Passthrough 4 - renames to value
function Layer4({ num }: { num: number }) {
  return <Leaf value={num} />;
}

// Passthrough 3 - renames to num
function Layer3({ amount }: { amount: number }) {
  return <Layer4 num={amount} />;
}

// Passthrough 2 - renames to amount
function Layer2({ total }: { total: number }) {
  return <Layer3 amount={total} />;
}

// Passthrough 1 - renames to total
function Layer1({ count }: { count: number }) {
  return <Layer2 total={count} />;
}

// Root with state
function Root() {
  const [count, setCount] = useState(0);

  return <Layer1 count={count} />;
}

export default Root;
