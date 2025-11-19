import { useState } from 'react';
import Dashboard from './Dashboard';

// Test Case: Cross-file prop drilling (3 files deep)
// Expected: 1 violation detecting drilling from App → Dashboard → Sidebar

// Level 0: Origin - creates state
function App() {
    const [theme, setTheme] = useState('dark');
    return <Dashboard theme={theme} />;
}

export default App;
