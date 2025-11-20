# Quick Start Guide - Graph Visualization

## Build and Run (5 minutes)

### 1. Build CLI Binary
```bash
cd /Users/oskari/repos/react-analyzer
go build -o react-analyzer ./cmd/react-analyzer
```

### 2. Compile Extension
```bash
cd vscode-extension
npm run compile
```

### 3. Launch Extension
1. Open VS Code in `vscode-extension` directory
2. Press **F5** to launch Extension Development Host
3. In the new window, open `/Users/oskari/repos/react-analyzer`

### 4. Test the Feature
1. Open: `test/fixtures/prop-drilling/SimpleDrilling.tsx`
2. Command Palette: **Cmd+Shift+P** (or Ctrl+Shift+P)
3. Type: "React Analyzer: Show Dependency Graph"
4. Press Enter

## What You Should See

### Graph View (Right Panel)
```
┌──────────────────────────────────────┐
│ Search: [        ] [Reset] [Clear]   │  ← Toolbar
├──────────────────────────────────────┤
│                                       │
│      ┌─────┐                         │
│      │ App │ (green - state origin)  │
│      └──┬──┘                         │
│         │ count                       │
│      ┌──▼────┐                       │
│      │Parent │ (yellow - passthrough)│
│      └──┬────┘                       │
│         │ count                       │
│      ┌──▼───┐                        │
│      │Child │ (yellow - passthrough) │
│      └──┬───┘                        │
│         │ count                       │
│      ┌──▼─────┐                      │
│      │Display │ (gray - regular)     │
│      └────────┘                      │
│                                       │
└──────────────────────────────────────┘
```

### Try These Interactions

1. **Click "App" node**
   - ✅ Jumps to line 7 in SimpleDrilling.tsx
   - ✅ Detail panel slides in from right
   - ✅ Shows: File, Line, Type: "origin", Memoized: No

2. **Click "Parent" node**
   - ✅ Jumps to line 13
   - ✅ Highlights "App" and "Child" nodes (connected)
   - ✅ Detail panel updates

3. **Type "parent" in search**
   - ✅ "Parent" node stays visible
   - ✅ Other nodes fade to 30% opacity

4. **Click "Clear Highlights"**
   - ✅ All nodes return to normal

## Quick Verification Checklist

- [ ] Graph renders without errors
- [ ] Nodes are color-coded correctly
- [ ] Clicking a node jumps to source code
- [ ] Detail panel shows metadata
- [ ] Search filters nodes
- [ ] No console errors (F12 → Console)

## Troubleshooting

### Graph doesn't show
```bash
# Verify CLI works
./react-analyzer -mermaid test/fixtures/prop-drilling/SimpleDrilling.tsx

# Should output flowchart syntax
```

### "CLI not found" error
```bash
# Check CLI exists
ls -la /Users/oskari/repos/react-analyzer/react-analyzer

# If missing, rebuild
go build -o react-analyzer ./cmd/react-analyzer
```

### TypeScript errors
```bash
# Recompile
cd vscode-extension
npm run compile

# Check for errors in output
```

### Webview is blank
1. Open Developer Tools in Extension Host (Help > Toggle Developer Tools)
2. Check Console tab for errors
3. Look for Mermaid.js loading errors
4. Check CSP violations

## Demo Script (2 minutes)

**Intro**: "I built a graph visualization for the React Analyzer that shows component and state relationships"

**Step 1**: Open SimpleDrilling.tsx
- "Here's a simple prop drilling example with App → Parent → Child → Display"

**Step 2**: Run "Show Dependency Graph"
- "The graph shows the component hierarchy"
- "Green = creates state, Yellow = passes props through, Gray = consumes"

**Step 3**: Click "App" node
- "Click a node to jump to its definition and see details"
- "Notice the detail panel on the right"

**Step 4**: Click "Parent" node
- "Connected nodes are highlighted - App (parent) and Child (child)"

**Step 5**: Search "display"
- "Search filters the graph"

**Conclusion**: "The graph uses Mermaid format, so it's universal - works in GitHub, Notion, etc. But we layer custom interactions on top for the IDE experience."

## File Locations

- CLI binary: `/Users/oskari/repos/react-analyzer/react-analyzer`
- Extension code: `/Users/oskari/repos/react-analyzer/vscode-extension/`
- Test fixtures: `/Users/oskari/repos/react-analyzer/test/fixtures/prop-drilling/`

## Next Steps After Testing

1. Test with larger component trees
2. Add more test fixtures
3. Document in main README
4. Consider packaging the extension
5. Add animated demo GIF to docs

## Questions?

See full documentation:
- `TESTING.md` - Comprehensive test cases
- `IMPLEMENTATION_SUMMARY.md` - Technical details
