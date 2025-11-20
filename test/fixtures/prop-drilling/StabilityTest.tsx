import React, { useState, useMemo } from 'react';

function Parent() {
    const [count, setCount] = useState(0);

    // Test: pass identifier (stable)
    return <Child1 count={count} />;
}

const Child1 = React.memo(({ count }: any) => {
    // Test: pass useMemo result (stable, optimized)
    const doubled = useMemo(() => count * 2, [count]);
    return <Child2 value={doubled} />;
});

const Child2 = React.memo(({ value }: any) => {
    // Test: pass inline object (unstable, breaks memo)
    return <Child3 data={{ x: value }} />;
});

const Child3 = React.memo(({ data }: any) => {
    // Test: pass primitive (stable)
    return <Child4 number={42} />;
});

const Child4 = React.memo(({ number }: any) => {
    return <div>{number}</div>;
});

export default Parent;
