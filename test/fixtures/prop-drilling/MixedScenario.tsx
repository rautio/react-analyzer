import { useState } from 'react';

// Test Case: Mixed scenario - some props drilled, others used at each level
// Expected: 1 violation (only 'userId' is drilled, 'title' and 'isLoading' are used locally)
// Tests algorithm's ability to distinguish between drilled and locally-used props

interface LayoutProps {
    title: string;
    userId: string;
    isLoading: boolean;
}

function App() {
    const [title] = useState('My Dashboard');
    const [userId] = useState('user-123');
    const [isLoading] = useState(false);

    return (
        <div>
            <Layout title={title} userId={userId} isLoading={isLoading} />
        </div>
    );
}

// Layout uses title and isLoading, but just passes userId
function Layout({ title, userId, isLoading }: LayoutProps) {
    return (
        <div className="layout">
            <h1>{title}</h1>
            {isLoading && <div className="spinner">Loading...</div>}
            <Content userId={userId} />
        </div>
    );
}

// Content doesn't use userId, just passes it
function Content({ userId }: { userId: string }) {
    return (
        <div className="content">
            <p>Welcome to the content area</p>
            <UserPanel userId={userId} />
        </div>
    );
}

// UserPanel finally uses userId
function UserPanel({ userId }: { userId: string }) {
    return (
        <div className="user-panel">
            <p>User ID: {userId}</p>
            <button>View Profile</button>
        </div>
    );
}

export default App;
