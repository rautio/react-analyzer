// Level 2: Consumer - actually uses theme
function Sidebar({ theme }: { theme: string }) {
    return <div className={theme}>Sidebar Content</div>;
}

export default Sidebar;
