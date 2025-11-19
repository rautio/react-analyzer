import { useState } from 'react';
import { Container } from './Container';

// Test Case: Cross-file with named exports (3 files deep)
// Expected: 1 violation detecting drilling from AppNamed → Container → Panel

function AppNamed() {
    const [config, setConfig] = useState({ color: 'blue' });
    return <Container config={config} />;
}

export default AppNamed;
