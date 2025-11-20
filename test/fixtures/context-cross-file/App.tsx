import React from 'react';
import { ThemeProvider } from './ThemeProvider';
import { ThemedButton } from './ThemedButton';

export default function App() {
    return (
        <ThemeProvider>
            <div className="app">
                <h1>Theme Demo</h1>
                <ThemedButton />
            </div>
        </ThemeProvider>
    );
}
