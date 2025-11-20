import React, { useState } from 'react';

// Simple test to show edge colors based on stability
function App() {
    const [x, setX] = useState(0);

    // Pass stable identifier to memoized component (gray edge)
    return <Level1 x={x} />;
}

const Level1 = React.memo(({ x }: any) => {
    // Pass member access to memoized component (red edge - breaks memo)
    const obj = { value: x };
    return <Level2 y={obj.value} />;
});

const Level2 = React.memo(({ y }: any) => {
    // Pass identifier to regular component (gray edge)
    return <Level3 z={y} />;
});

function Level3({ z }: any) {
    return <div>{z}</div>;
}

export default App;
