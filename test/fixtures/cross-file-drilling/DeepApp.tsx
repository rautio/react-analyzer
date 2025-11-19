import { useState } from 'react';
import { Layout } from './Layout';

// Test Case: Deep cross-file drilling (4 files deep)
// Expected: 1 violation (2 passthrough components: Layout → Header)
// Phase 2.3 partial prop usage: violations require 2+ passthrough components

function DeepApp() {
    const [user, setUser] = useState({ name: 'John', role: 'admin' });
    return <Layout user={user} />;
}

export default DeepApp;
