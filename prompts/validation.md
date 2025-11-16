• Problem Reality Check
  - Unnecessary re-renders are a mainstream pain. DeveloperWay’s long-running “React re-renders guide” explains how
    state changes cascade down the tree and why “unnecessary re-renders” quickly make large pages feel laggy even
    though the code passes linting; it lists parent re-renders, context churn, and hook changes as the main triggers
    and links to visual examples of the issue (https://www.developerway.com/posts/react-re-renders-guide). React’s own
    memo docs call out the exact scenario your analyzer targets—memoized children still re-render when parents pass new
    object/array/function references—and dedicate troubleshooting guidance titled “My component re-renders when a prop
    is an object, array, or function” to encourage developers to stabilize props (https://react.dev/reference/react/
    memo#my-component-rerenders-when-a-prop-is-an-object-or-array). Sentry published a 2023 engineering post on “Fixing
    memoization-breaking re-renders,” profiling a real app where a memoized child still re-rendered because the parent
    recreated callbacks; they even hint at building automatic “component re-render performance issues,” which signals
    customer demand beyond anecdote (https://blog.sentry.io/fixing-memoization-breaking-re-renders-in-react/).
  - Hook dependency mistakes remain one of the most asked-about topics. The official useEffect reference dedicates
    an entire troubleshooting section to “My Effect runs after every re-render” and “My Effect keeps re-running in an
    infinite cycle,” explicitly blaming missing or unstable dependencies and recommending tooling to verify dependency
    arrays (https://react.dev/reference/react/useEffect#troubleshooting). Dan Abramov’s widely read “A Complete Guide
    to useEffect” opens with common developer questions—“Do I need to specify functions as effect dependencies?”, “Why
    do I sometimes get an infinite refetching loop?”, “Why do I sometimes get an old state or prop value?”—and stresses
    that even the React team struggled with stale closures and dependency arrays before codifying best practices
    (https://overreacted.io/a-complete-guide-to-useeffect/). That article (and countless StackOverflow threads it
    references) exists precisely because lint rules only cover part of the problem; developers still ship buggy
    dependency arrays, so advanced static analysis has clear audience pull.
  - Prop drilling and cross-file prop flow are recognized refactoring hazards. Kent C. Dodds’ “Prop Drilling” post
    describes how data threaded through intermediate components leads to maintenance headaches as the tree deepens:
    accidental over/under-forwarding, renaming inconsistencies, and the difficulty of knowing which components truly
    need a prop (https://kentcdodds.com/blog/prop-drilling). The article treats prop drilling as both a necessary
    pattern and a source of “real pain” during refactors—exactly the kind of signal your cross-file analyzer aims to
    surface (e.g., identifying when props are passed through three layers unused). Traditional linters operate per file
    and can’t reconstruct those threaded paths, so teams either accept the debt or rely on expensive manual audits.
  - Community demand extends beyond blog posts. Sentry’s call for feedback on automated re-render issue detection, Dan
    Abramov’s mega-guide answering recurring Hook questions, and Kent’s widely circulated prop-drilling piece all exist
    because practitioners repeatedly hit these problems even after adopting ESLint and the official Hooks lint rules.
    Those posts continue to receive updates and translations (e.g., DeveloperWay offers Japanese and Korean versions),
    implying sustained global interest rather than a one-off anecdote.
  Implication for the project
  The pain points motivating React Analyzer—unstable props breaking memoization, incorrect dependency arrays, and deep
  prop threading—are well documented by React core authors, major observability vendors, and influential educators.
  Existing tools help partially (React’s Hook lint rule, ESLint plugins, Perf Fiscal), but the breadth of ongoing
  guidance and the fact that companies like Sentry are considering product features around re-render detection indicate
  the unmet need is real. Building deeper static analysis that tracks prop flow and memoization stability across files
  aligns with active, public developer struggles, so the project is addressing a validated problem space rather than an
  isolated anecdote.

