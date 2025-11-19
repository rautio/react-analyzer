import { useState } from 'react';

// Test Case: Partial usage (mixed)
// Expected: 0 violations (Parent uses theme, so it's not pure passthrough)

function App() {
    const [theme, setTheme] = useState('dark');
    return <Parent theme={theme} />;
}

// Parent uses theme locally AND passes it down
function Parent({ theme }: { theme: string }) {
    // Uses theme for its own styling
    const styles = { background: theme === 'dark' ? '#000' : '#fff' };
    return (
        <div style={styles}>
            <Child theme={theme} />
        </div>
    );
}

// Child just passes it through
function Child({ theme }: { theme: string }) {
    return <Display theme={theme} />;
}

// Display uses theme
function Display({ theme }: { theme: string }) {
    return <div className={theme}>Content</div>;
}

export default App;
