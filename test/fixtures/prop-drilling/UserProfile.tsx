import { useState } from 'react';

// Test Case: User data drilling through profile components
// Expected: 2 violations (userId and userName drilled through ProfilePage → ProfileContent → UserAvatar)
// Common pattern in apps that fetch user data at top level

interface User {
    id: string;
    name: string;
    email: string;
    avatar: string;
}

function App() {
    const [user, setUser] = useState<User>({
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatar: 'https://example.com/avatar.jpg',
    });

    return (
        <div className="app">
            <ProfilePage userId={user.id} userName={user.name} userAvatar={user.avatar} />
        </div>
    );
}

// ProfilePage doesn't use user data, just layouts
interface ProfilePageProps {
    userId: string;
    userName: string;
    userAvatar: string;
}

function ProfilePage({ userId, userName, userAvatar }: ProfilePageProps) {
    return (
        <div className="profile-page">
            <h1>User Profile</h1>
            <ProfileContent userId={userId} userName={userName} userAvatar={userAvatar} />
            <Sidebar />
        </div>
    );
}

function Sidebar() {
    return (
        <aside>
            <nav>
                <ul>
                    <li>Overview</li>
                    <li>Settings</li>
                    <li>Activity</li>
                </ul>
            </nav>
        </aside>
    );
}

// ProfileContent still doesn't use the data, just organizes layout
function ProfileContent({ userId, userName, userAvatar }: ProfilePageProps) {
    return (
        <div className="profile-content">
            <UserHeader userId={userId} userName={userName} userAvatar={userAvatar} />
            <UserStats />
        </div>
    );
}

function UserStats() {
    return (
        <div className="stats">
            <div>Posts: 42</div>
            <div>Followers: 1.2k</div>
        </div>
    );
}

// UserHeader finally uses the user data
function UserHeader({ userId, userName, userAvatar }: ProfilePageProps) {
    return (
        <div className="user-header">
            <img src={userAvatar} alt={userName} />
            <h2>{userName}</h2>
            <p>ID: {userId}</p>
        </div>
    );
}

export default App;
