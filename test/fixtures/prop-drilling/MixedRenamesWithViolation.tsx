import React, { useState } from 'react';

/**
 * Test fixture with both renamed and non-renamed props
 *
 * This demonstrates that renamed and non-renamed props are tracked independently:
 * - Root defines two state variables: count and title
 * - count is renamed at each layer (count → num → value → total)
 * - title is NOT renamed, stays "title" throughout
 * - Both should trigger prop drilling violations independently
 *
 * Expected behavior:
 * - Should detect 2 separate prop drilling violations (one for each prop)
 * - Clicking count chain should highlight only count edges
 * - Clicking title chain should highlight only title edges
 * - Renaming should only show for count edges, not title
 */

interface LeafProps {
  total: number;
  title: string;
}

// Leaf component that uses both props
function Leaf({ total, title }: LeafProps) {
  return (
    <div>
      <h2>{title}</h2>
      <p>Count: {total}</p>
    </div>
  );
}

interface Layer3Props {
  value: number;
  title: string;
}

// Passthrough layer 3 - renames value to total, keeps title
function Layer3({ value, title }: Layer3Props) {
  return <Leaf total={value} title={title} />;
}

interface Layer2Props {
  num: number;
  title: string;
}

// Passthrough layer 2 - renames num to value, keeps title
function Layer2({ num, title }: Layer2Props) {
  return <Layer3 value={num} title={title} />;
}

interface Layer1Props {
  count: number;
  title: string;
}

// Passthrough layer 1 - renames count to num, keeps title
function Layer1({ count, title }: Layer1Props) {
  return <Layer2 num={count} title={title} />;
}

// Root component with the actual state
function Root() {
  const [count, setCount] = useState(0);
  const [title, setTitle] = useState('My Counter');

  return <Layer1 count={count} title={title} />;
}

export default Root;
