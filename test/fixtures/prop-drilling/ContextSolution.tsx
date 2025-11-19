import { useState, createContext, useContext } from 'react';

// Test Case: CORRECT solution using Context API
// Expected: 0 violations (no prop drilling)
// This demonstrates the recommended solution for the theme drilling problem

// Create context for theme
interface ThemeContextType {
    theme: 'light' | 'dark';
    setTheme: (theme: 'light' | 'dark') => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

// Custom hook for accessing theme
function useTheme() {
    const context = useContext(ThemeContext);
    if (!context) {
        throw new Error('useTheme must be used within ThemeProvider');
    }
    return context;
}

// App provides theme at top level
function App() {
    const [theme, setTheme] = useState<'light' | 'dark'>('dark');

    return (
        <ThemeContext.Provider value={{ theme, setTheme }}>
            <div className={`app ${theme}`}>
                <Dashboard />
            </div>
        </ThemeContext.Provider>
    );
}

// Dashboard doesn't need theme prop - clean!
function Dashboard() {
    return (
        <div className="dashboard">
            <Header>
                <Sidebar />
            </Header>
            <MainContent />
        </div>
    );
}

function Header({ children }: { children: React.ReactNode }) {
    return <header>{children}</header>;
}

function MainContent() {
    return <div>Main content area</div>;
}

// Sidebar doesn't need theme prop either
function Sidebar() {
    return (
        <aside>
            <nav>
                <ul>
                    <li>Home</li>
                    <li>Settings</li>
                </ul>
            </nav>
            <ThemeToggle />
        </aside>
    );
}

// ThemeToggle consumes theme directly from context - no drilling!
function ThemeToggle() {
    const { theme, setTheme } = useTheme();

    return (
        <button onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
            {theme === 'dark' ? '☀️' : '🌙'} Toggle Theme
        </button>
    );
}

export default App;
