import Sidebar from './Sidebar';

// Level 1: Passthrough - doesn't use theme, just passes it
function Dashboard({ theme }: { theme: string }) {
    return <Sidebar theme={theme} />;
}

export default Dashboard;
