import { useState } from 'react';

// Test Case: Simple prop spread through component chain
// Expected: Detect drilling when props are spread

// Level 0: Origin
function App() {
    const [theme, setTheme] = useState('dark');
    const [lang, setLang] = useState('en');
    return <Container theme={theme} lang={lang} />;
}

// Level 1: Spreads all props
function Container(props: any) {
    return <Panel {...props} />;
}

// Level 2: Also spreads
function Panel(props: any) {
    return <Display {...props} />;
}

// Level 3: Consumer
function Display({ theme, lang }: { theme: string; lang: string }) {
    return <div className={theme}>{lang}</div>;
}

export default App;
