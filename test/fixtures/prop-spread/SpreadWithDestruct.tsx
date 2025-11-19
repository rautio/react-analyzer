import { useState } from 'react';

// Test Case: Prop spread with final destructuring
// Expected: Detect drilling for theme prop through spread chain

function App() {
    const [theme, setTheme] = useState('dark');
    const [lang, setLang] = useState('en');
    return <Wrapper theme={theme} lang={lang} />;
}

// Wrapper spreads all props
function Wrapper(props: any) {
    return <Container {...props} />;
}

// Container also spreads
function Container(props: any) {
    return <Display {...props} />;
}

// Display finally destructures and uses the props
function Display({ theme, lang }: { theme: string; lang: string }) {
    return <div className={theme}>{lang}</div>;
}

export default App;
