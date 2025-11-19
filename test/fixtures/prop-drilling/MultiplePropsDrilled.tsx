import { useState } from 'react';

// Test Case: Multiple props drilled together
// Expected: 2 violations (one for theme, one for lang)

type Props = {
    theme: string;
    lang: string;
};

function App() {
    const [theme, setTheme] = useState('dark');
    const [lang, setLang] = useState('en');
    return <Parent theme={theme} lang={lang} />;
}

// Level 1: Passthrough (doesn't use theme or lang)
function Parent({ theme, lang }: Props) {
    return <Child theme={theme} lang={lang} />;
}

// Level 2: Passthrough (doesn't use theme or lang)
function Child({ theme, lang }: Props) {
    return <Display theme={theme} lang={lang} />;
}

// Level 3: Consumer (uses both theme and lang)
function Display({ theme, lang }: Props) {
    return <div className={theme}>{lang}</div>;
}

export default App;
