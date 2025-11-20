import React from 'react';

function App() {
    // Pass primitive literal
    return <Child number={42} />;
}

function Child({ number }: any) {
    return <div>{number}</div>;
}

export default App;
