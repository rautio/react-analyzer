import { useState } from 'react';
import Dashboard from './Dashboard';

// Test Case: Cross-file prop drilling (3 files deep)
// Expected: 0 violations (only 1 passthrough component: Dashboard)
// Phase 2.3 partial prop usage: violations require 2+ passthrough components

// Level 0: Origin - creates state
function App() {
    const [theme, setTheme] = useState('dark');
    return <Dashboard theme={theme} />;
}

export default App;
