import { useState } from 'react';

// Test Case: Props passed via spread operator
// Expected: This is a more advanced case - may require Phase 2.2 for full detection
// Tests algorithm's handling of prop spreading patterns

interface Config {
    apiUrl: string;
    timeout: number;
    retries: number;
}

function App() {
    const [config] = useState<Config>({
        apiUrl: 'https://api.example.com',
        timeout: 5000,
        retries: 3,
    });

    return (
        <div>
            <Container {...config} />
        </div>
    );
}

// Container spreads all props to Panel
function Container(props: Config) {
    return (
        <div className="container">
            <Panel {...props} />
        </div>
    );
}

// Panel spreads all props to Settings
function Panel(props: Config) {
    return (
        <div className="panel">
            <Settings {...props} />
        </div>
    );
}

// Settings finally uses the config
function Settings({ apiUrl, timeout, retries }: Config) {
    return (
        <div className="settings">
            <div>API URL: {apiUrl}</div>
            <div>Timeout: {timeout}ms</div>
            <div>Retries: {retries}</div>
        </div>
    );
}

export default App;
