import { useState, createContext, useContext } from 'react';

// Test Case: Real-world theme drilling scenario
// Expected: 1 violation (theme drilled through Dashboard → Sidebar → ThemeToggle)
// This demonstrates WHY Context API should be used for theme

function App() {
    const [theme, setTheme] = useState<'light' | 'dark'>('dark');

    return (
        <div className={`app ${theme}`}>
            <Dashboard theme={theme} onThemeChange={setTheme} />
        </div>
    );
}

// Dashboard doesn't care about theme, just passes it
interface DashboardProps {
    theme: 'light' | 'dark';
    onThemeChange: (theme: 'light' | 'dark') => void;
}

function Dashboard({ theme, onThemeChange }: DashboardProps) {
    return (
        <div className="dashboard">
            <Header>
                <Sidebar theme={theme} onThemeChange={onThemeChange} />
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

// Sidebar doesn't use theme either, just passes it
interface SidebarProps {
    theme: 'light' | 'dark';
    onThemeChange: (theme: 'light' | 'dark') => void;
}

function Sidebar({ theme, onThemeChange }: SidebarProps) {
    return (
        <aside>
            <nav>
                <ul>
                    <li>Home</li>
                    <li>Settings</li>
                </ul>
            </nav>
            <ThemeToggle theme={theme} onThemeChange={onThemeChange} />
        </aside>
    );
}

// ThemeToggle finally uses theme
interface ThemeToggleProps {
    theme: 'light' | 'dark';
    onThemeChange: (theme: 'light' | 'dark') => void;
}

function ThemeToggle({ theme, onThemeChange }: ThemeToggleProps) {
    return (
        <button onClick={() => onThemeChange(theme === 'dark' ? 'light' : 'dark')}>
            {theme === 'dark' ? '☀️' : '🌙'} Toggle Theme
        </button>
    );
}

export default App;
