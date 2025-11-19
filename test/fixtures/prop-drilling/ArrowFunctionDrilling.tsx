import { useState } from 'react';

// Test Case: Arrow function components with 3-level prop drilling
// Expected: 1 violation (depth: 3, passthrough: Parent, Child)

// Level 0: Origin (arrow function)
const App = () => {
    const [count, setCount] = useState(0);
    return <Parent count={count} />;
};

// Level 1: Passthrough (arrow function, doesn't use count, just passes it)
const Parent = ({ count }: { count: number }) => {
    return <Child count={count} />;
};

// Level 2: Passthrough (arrow function, doesn't use count, just passes it)
const Child = ({ count }: { count: number }) => {
    return <Display count={count} />;
};

// Level 3: Consumer (arrow function, actually uses count)
const Display = ({ count }: { count: number }) => {
    return <div>{count}</div>;
};

export default App;
