import React, { useState } from 'react';

// Leaf component
function Leaf({ count }: { count: number }) {
  return <div>{count}</div>;
}

// Passthrough 4
function Layer4({ count }: { count: number }) {
  return <Leaf count={count} />;
}

// Passthrough 3
function Layer3({ count }: { count: number }) {
  return <Layer4 count={count} />;
}

// Passthrough 2
function Layer2({ count }: { count: number }) {
  return <Layer3 count={count} />;
}

// Passthrough 1
function Layer1({ count }: { count: number }) {
  return <Layer2 count={count} />;
}

// Root with state
function Root() {
  const [count, setCount] = useState(0);

  return <Layer1 count={count} />;
}

export default Root;
