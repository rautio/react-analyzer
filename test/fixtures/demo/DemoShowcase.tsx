import React, { useState, useEffect } from 'react';

// Parent component with multiple violations
function Parent() {
    const [count, setCount] = useState(0);
    const [config, setConfig] = useState({ theme: 'dark' });

    // Violation: Object in useEffect deps
    useEffect(() => {
        console.log('Effect running');
    }, [config]);

    // Pass inline object (violation: no-inline-props)
    return (
        <Child
            count={count}
            settings={{ theme: 'dark', size: 'large' }}
            onClick={() => setCount(count + 1)}
        />
    );
}

// Memoized child receiving unstable props
const Child = React.memo(({ count, settings, onClick }: any) => {
    return (
        <div>
            <p>Count: {count}</p>
            <button onClick={onClick}>Increment</button>
            <Display count={count} />
        </div>
    );
});

// Leaf component
function Display({ count }: { count: number }) {
    return <span>{count}</span>;
}

export default Parent;
