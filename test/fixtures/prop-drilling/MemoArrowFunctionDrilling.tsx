import { useState, memo } from 'react';

// Test Case: React.memo wrapped arrow function components with prop drilling
// Expected: 1 violation (depth: 3, passthrough: Parent, Child)

// Level 0: Origin (memo wrapped arrow function)
const App = memo(() => {
    const [theme, setTheme] = useState('dark');
    return <Parent theme={theme} />;
});

// Level 1: Passthrough (memo wrapped arrow function)
const Parent = memo(({ theme }: { theme: string }) => {
    return <Child theme={theme} />;
});

// Level 2: Passthrough (memo wrapped arrow function)
const Child = memo(({ theme }: { theme: string }) => {
    return <Display theme={theme} />;
});

// Level 3: Consumer (memo wrapped arrow function)
const Display = memo(({ theme }: { theme: string }) => {
    return <div className={theme}>Content</div>;
});

export default App;
