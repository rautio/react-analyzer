import React, { useContext } from 'react';
import { ThemeContext } from './ThemeContext';

export function ThemedButton() {
    const { theme, setTheme } = useContext(ThemeContext);

    return (
        <button
            className={`btn-${theme}`}
            onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
        >
            {theme === 'dark' ? '☀️' : '🌙'} Toggle Theme
        </button>
    );
}
