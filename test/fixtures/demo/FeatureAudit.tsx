import React, { useState, useContext, createContext, useMemo, useCallback, useReducer, useRef, useEffect } from 'react';

// Test 1: Context (Provider and Consumer)
const ThemeContext = createContext({ theme: 'dark' });

function App() {
    const [count, setCount] = useState(0);

    return (
        <ThemeContext.Provider value={{ theme: 'dark' }}>
            <Parent count={count} />
        </ThemeContext.Provider>
    );
}

// Test 2: useContext consumption
function Parent({ count }: any) {
    const theme = useContext(ThemeContext);

    // Test 3: useReducer
    const [state, dispatch] = useReducer(reducer, { value: 0 });

    // Test 4: useMemo (derived state)
    const doubled = useMemo(() => count * 2, [count]);

    // Test 5: useCallback
    const handleClick = useCallback(() => {
        dispatch({ type: 'increment' });
    }, []);

    // Test 6: useRef
    const inputRef = useRef<HTMLInputElement>(null);

    // Test 7: useEffect with dependencies
    useEffect(() => {
        console.log('Count changed:', count);
    }, [count]);

    // Test 8: Pass inline object to memoized child
    return (
        <Child
            count={count}
            config={{ mode: 'dev' }}
            onClick={handleClick}
            doubled={doubled}
        />
    );
}

// Test 9: Memoized component
const Child = React.memo(({ count, config, onClick, doubled }: any) => {
    // Test 10: Component consuming props
    return (
        <div>
            <p>Count: {count}</p>
            <p>Doubled: {doubled}</p>
            <button onClick={onClick}>Click</button>
        </div>
    );
});

function reducer(state: any, action: any) {
    switch (action.type) {
        case 'increment':
            return { value: state.value + 1 };
        default:
            return state;
    }
}

export default App;
