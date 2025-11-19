import { useState } from 'react';

// Test Case: Simple 3-level prop drilling
// Expected: 1 violation (depth: 3, passthrough: Parent, Child)

// Level 0: Origin
function App() {
    const [count, setCount] = useState(0);
    return <Parent count={count} />;
}

// Level 1: Passthrough (doesn't use count, just passes it)
function Parent({ count }: { count: number }) {
    return <Child count={count} />;
}

// Level 2: Passthrough (doesn't use count, just passes it)
function Child({ count }: { count: number }) {
    return <Display count={count} />;
}

// Level 3: Consumer (actually uses count)
function Display({ count }: { count: number }) {
    return <div>{count}</div>;
}

export default App;
