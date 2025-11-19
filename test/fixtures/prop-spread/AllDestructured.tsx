import { useState } from 'react';

// Test Case: All components destructure - simpler spread case
// Expected: Should detect prop drilling with spreads

function App() {
    const [theme, setTheme] = useState('dark');
    return <Wrapper theme={theme} />;
}

// Wrapper destructures and spreads
function Wrapper({ theme }: { theme: string }) {
    return <Container theme={theme} />;
}

// Container also destructures and spreads
function Container({ theme }: { theme: string }) {
    return <Display theme={theme} />;
}

// Display uses it
function Display({ theme }: { theme: string }) {
    return <div className={theme}>Content</div>;
}

export default App;
