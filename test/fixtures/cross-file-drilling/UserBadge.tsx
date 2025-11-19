// Level 3: Consumer - finally uses the user prop
export function UserBadge({ user }: { user: any }) {
    return <div>Welcome, {user.name} ({user.role})</div>;
}
