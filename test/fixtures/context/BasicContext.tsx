import React, { createContext, useContext, useState } from 'react';

// Create a theme context
const ThemeContext = createContext('light');

function App() {
    const [theme, setTheme] = useState('dark');

    return (
        <ThemeContext.Provider value={theme}>
            <ThemedButton />
        </ThemeContext.Provider>
    );
}

function ThemedButton() {
    const theme = useContext(ThemeContext);

    return <button className={theme}>Click me</button>;
}

export default App;
