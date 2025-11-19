import { Header } from './Header';

// Level 1: Passthrough
export function Layout({ user }: { user: any }) {
    return (
        <div>
            <Header user={user} />
        </div>
    );
}
