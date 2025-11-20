import React, { useState, useMemo, useCallback } from 'react';

// Parent showing different stability patterns
function Parent() {
    const [count, setCount] = useState(0);
    const [name, setName] = useState('John');

    // Stable: useMemo
    const config = useMemo(() => ({ theme: 'dark', size: 'large' }), []);

    // Stable: useCallback
    const handleClick = useCallback(() => setCount(c => c + 1), []);

    return (
        <div>
            {/* Stable: identifier */}
            <StableChild count={count} />

            {/* Stable: useMemo */}
            <OptimizedChild config={config} />

            {/* Stable: useCallback */}
            <CallbackChild onClick={handleClick} />

            {/* Stable: primitive */}
            <PrimitiveChild value={42} name="test" />

            {/* Unstable: inline object - breaks memo! */}
            <UnstableChild data={{ x: 1, y: 2 }} />
        </div>
    );
}

const StableChild = React.memo(({ count }: { count: number }) => {
    return <div>Count: {count}</div>;
});

const OptimizedChild = React.memo(({ config }: any) => {
    return <div>Theme: {config.theme}</div>;
});

const CallbackChild = React.memo(({ onClick }: any) => {
    return <button onClick={onClick}>Click me</button>;
});

const PrimitiveChild = React.memo(({ value, name }: any) => {
    return <div>{name}: {value}</div>;
});

const UnstableChild = React.memo(({ data }: any) => {
    return <div>X: {data.x}, Y: {data.y}</div>;
});

export default Parent;
