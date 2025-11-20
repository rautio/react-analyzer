import React, { useState, ReactNode } from 'react';
import { ThemeContext } from './ThemeContext';

interface ThemeProviderProps {
    children: ReactNode;
}

export function ThemeProvider({ children }: ThemeProviderProps) {
    const [theme, setTheme] = useState<'light' | 'dark'>('dark');

    // ISSUE: Inline object creates new reference every render
    return (
        <ThemeContext.Provider value={{ theme, setTheme }}>
            {children}
        </ThemeContext.Provider>
    );
}
