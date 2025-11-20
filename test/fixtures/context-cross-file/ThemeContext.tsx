import { createContext } from 'react';

// Define theme context type
interface ThemeContextType {
    theme: 'light' | 'dark';
    setTheme: (theme: 'light' | 'dark') => void;
}

// Create and export the context
export const ThemeContext = createContext<ThemeContextType>({
    theme: 'light',
    setTheme: () => {}
});
