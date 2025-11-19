import { useState } from 'react';
import { Layout } from './Layout';

// Test Case: Deep cross-file drilling (4 files deep)
// Expected: 1 violation for drilling through 4+ levels

function DeepApp() {
    const [user, setUser] = useState({ name: 'John', role: 'admin' });
    return <Layout user={user} />;
}

export default DeepApp;
