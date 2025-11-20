import { useState } from 'react';

function App() {
    // Test different data types
    const [count, setCount] = useState(0);                    // number
    const [name, setName] = useState('John');                 // string (primitive)
    const [isActive, setIsActive] = useState(true);           // boolean
    const [user, setUser] = useState({ id: 1, name: 'Bob' }); // object
    const [items, setItems] = useState([1, 2, 3]);           // array
    const [callback, setCallback] = useState(() => {});      // function

    return (
        <div>
            <p>Count: {count}</p>
            <p>Name: {name}</p>
            <p>Active: {isActive ? 'yes' : 'no'}</p>
        </div>
    );
}

export default App;
