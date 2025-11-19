import { useState } from 'react';

// Test Case: Very deep prop drilling (6 levels)
// Expected: 1 violation with depth of 6
// Tests algorithm's ability to handle deeply nested component hierarchies

function App() {
    const [apiKey, setApiKey] = useState('sk-1234567890');

    return (
        <div>
            <PageWrapper apiKey={apiKey} />
        </div>
    );
}

// Level 1: PageWrapper
function PageWrapper({ apiKey }: { apiKey: string }) {
    return (
        <div className="page-wrapper">
            <ContentArea apiKey={apiKey} />
        </div>
    );
}

// Level 2: ContentArea
function ContentArea({ apiKey }: { apiKey: string }) {
    return (
        <div className="content-area">
            <FeatureSection apiKey={apiKey} />
        </div>
    );
}

// Level 3: FeatureSection
function FeatureSection({ apiKey }: { apiKey: string }) {
    return (
        <section>
            <WidgetContainer apiKey={apiKey} />
        </section>
    );
}

// Level 4: WidgetContainer
function WidgetContainer({ apiKey }: { apiKey: string }) {
    return (
        <div className="widget-container">
            <Widget apiKey={apiKey} />
        </div>
    );
}

// Level 5: Widget
function Widget({ apiKey }: { apiKey: string }) {
    return (
        <div className="widget">
            <DataDisplay apiKey={apiKey} />
        </div>
    );
}

// Level 6: DataDisplay (finally uses apiKey)
function DataDisplay({ apiKey }: { apiKey: string }) {
    return (
        <div>
            <p>Connected with API key: {apiKey.substring(0, 8)}...</p>
        </div>
    );
}

export default App;
