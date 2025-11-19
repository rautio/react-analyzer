import { useState } from 'react';

// Test Case: Mixed arrow functions and function declarations
// Expected: 1 violation (depth: 3, passthrough: Parent, Child)

// Level 0: Origin (arrow function)
const App = () => {
    const [user, setUser] = useState({ name: 'John' });
    return <Parent user={user} />;
};

// Level 1: Passthrough (function declaration)
function Parent({ user }: { user: any }) {
    return <Child user={user} />;
}

// Level 2: Passthrough (arrow function)
const Child = ({ user }: { user: any }) => {
    return <Display user={user} />;
};

// Level 3: Consumer (function declaration)
function Display({ user }: { user: any }) {
    return <div>{user.name}</div>;
}

export default App;
