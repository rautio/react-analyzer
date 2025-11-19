import { useState } from 'react';

// Test Case: Destructured props with spread
// Expected: Detect drilling with destructured spread

function App() {
    const [config, setConfig] = useState({ theme: 'dark', locale: 'en' });
    return <Wrapper config={config} />;
}

// Destructures one prop, spreads the rest
function Wrapper({ config, ...rest }: any) {
    return (
        <div>
            <Inner config={config} {...rest} />
        </div>
    );
}

function Inner({ config }: any) {
    return <div>{config.theme}</div>;
}

export default App;
