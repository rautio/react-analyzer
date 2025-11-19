import { useState } from 'react';

// Test Case: No prop drilling (direct usage)
// Expected: 0 violations

function App() {
    const [count, setCount] = useState(0);
    return <Parent count={count} />;
}

// Parent uses count directly - no drilling
function Parent({ count }: { count: number }) {
    return <div>{count}</div>;
}

export default App;
