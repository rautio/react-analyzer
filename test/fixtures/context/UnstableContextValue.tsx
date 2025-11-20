import React, { createContext, useContext, useState, memo } from 'react';

const ThemeContext = createContext({ theme: 'light' });

function App() {
    const [theme, setTheme] = useState('dark');

    // ISSUE: Inline object creates new reference every render
    return (
        <ThemeContext.Provider value={{ theme, setTheme }}>
            <MemoizedChild />
        </ThemeContext.Provider>
    );
}

// Memoized component that consumes context
const MemoizedChild = memo(function MemoizedChild() {
    const { theme } = useContext(ThemeContext);

    return <div>Current theme: {theme}</div>;
});

export default App;
