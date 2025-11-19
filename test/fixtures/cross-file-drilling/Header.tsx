import { UserBadge } from './UserBadge';

// Level 2: Passthrough
export function Header({ user }: { user: any }) {
    return (
        <header>
            <UserBadge user={user} />
        </header>
    );
}
