import React, { useState, useMemo, useCallback } from 'react';

// Test different stability scenarios
function App() {
    const [count, setCount] = useState(0);

    // Scenario 1: Pass stable identifier
    return <Parent1 count={count} />;
}

function Parent1({ count }: { count: number }) {
    // Scenario 2: Pass primitive (42)
    return <Parent2 count={count} magic={42} />;
}

function Parent2({ count, magic }: any) {
    // Scenario 3: Pass useMemo result
    const config = useMemo(() => ({ threshold: 100 }), []);
    return <Parent3 count={count} config={config} />;
}

// Memoized child
const Parent3 = React.memo(({ count, config }: any) => {
    // Scenario 4: Pass inline object (unstable, breaks memo!)
    return <Child data={{ value: count }} threshold={config.threshold} />;
});

// Memoized child receiving unstable prop
const Child = React.memo(({ data, threshold }: any) => {
    return <div>Value: {data.value}, Threshold: {threshold}</div>;
});

export default App;
