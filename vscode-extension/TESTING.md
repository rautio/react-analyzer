# VS Code Extension Testing Guide

## Prerequisites

1. Build the CLI binary:
   ```bash
   cd /Users/oskari/repos/react-analyzer
   go build -o react-analyzer ./cmd/react-analyzer
   ```

2. Compile the TypeScript extension:
   ```bash
   cd vscode-extension
   npm run compile
   ```

## Running the Extension

1. Open VS Code in the `vscode-extension` directory
2. Press F5 to launch the Extension Development Host
3. In the new window, open the parent directory: `/Users/oskari/repos/react-analyzer`

## Test Cases

### Test 1: Show Dependency Graph Command

**Steps:**
1. Open a React file: `test/fixtures/prop-drilling/SimpleDrilling.tsx`
2. Open Command Palette (Cmd+Shift+P / Ctrl+Shift+P)
3. Run: "React Analyzer: Show Dependency Graph"

**Expected Results:**
- ✅ Webview panel opens in column 2
- ✅ Mermaid diagram renders showing component hierarchy
- ✅ Nodes are color-coded:
  - App (green) - state origin
  - Parent (yellow) - passthrough
  - Child (yellow) - passthrough
  - Display (gray) - regular consumer
- ✅ Edges show prop flow with labels ("count")

### Test 2: Click Node to Jump to Source

**Steps:**
1. With the graph open, click on a component node (e.g., "Parent")

**Expected Results:**
- ✅ Editor switches to column 1
- ✅ File opens: `SimpleDrilling.tsx`
- ✅ Cursor jumps to the component definition line
- ✅ Line is centered in the editor view

### Test 3: Detail Panel

**Steps:**
1. Click on a component node in the graph

**Expected Results:**
- ✅ Detail panel slides in from the right
- ✅ Shows component name
- ✅ Shows file name
- ✅ Shows line number
- ✅ Shows node type (origin/passthrough/consumer/regular)
- ✅ Shows memoization status
- ✅ Shows full file path
- ✅ Close button (×) works

### Test 4: Highlight Connected Nodes

**Steps:**
1. Click on a component node (e.g., "Parent")

**Expected Results:**
- ✅ Connected nodes get highlighted with thicker red border
- ✅ For "Parent": Should highlight "App" (incoming) and "Child" (outgoing)

### Test 5: Search Functionality

**Steps:**
1. In the toolbar search box, type "parent"

**Expected Results:**
- ✅ Matching nodes remain at full opacity
- ✅ Non-matching nodes fade to 30% opacity
- ✅ Clearing search restores all nodes to full opacity

### Test 6: Toolbar Buttons

**Steps:**
1. Click "Reset View" button

**Expected Results:**
- ✅ Graph scroll position resets to top-left

**Steps:**
2. Click a node to highlight connections, then click "Clear Highlights"

**Expected Results:**
- ✅ All highlighted nodes return to normal styling

### Test 7: Multiple Files

**Steps:**
1. Open a different React file (e.g., `test/fixtures/prop-drilling/NestedDrilling.tsx`)
2. Run "React Analyzer: Show Dependency Graph" again

**Expected Results:**
- ✅ Same webview panel is reused (not duplicated)
- ✅ Graph updates to show the new file's component tree

### Test 8: Non-React Files

**Steps:**
1. Open a non-React file (e.g., `README.md` or `.go` file)
2. Run "React Analyzer: Show Dependency Graph"

**Expected Results:**
- ✅ Error message: "Please open a React file (.tsx, .jsx, .ts, .js)"

### Test 9: No Active Editor

**Steps:**
1. Close all open editors
2. Run "React Analyzer: Show Dependency Graph"

**Expected Results:**
- ✅ Error message: "No active editor"

## Visual Verification

### Styling Checks
- ✅ Dark mode colors match VS Code theme
- ✅ Toolbar has proper background and border
- ✅ Search input is styled correctly
- ✅ Buttons have hover effects
- ✅ Detail panel animation is smooth (slides in/out)

### Interaction Checks
- ✅ Node hover shows brightness effect and drop shadow
- ✅ Cursor changes to pointer on node hover
- ✅ Pan/drag works on empty space in graph container
- ✅ Memoized nodes have dashed purple border (when applicable)

## Known Issues to Check

1. **Mermaid rendering errors**: Check browser console for any Mermaid.js errors
2. **CSP violations**: Check for Content Security Policy warnings
3. **CLI path**: Verify CLI is found at `../react-analyzer` relative to extension
4. **Performance**: Large graphs should render without freezing

## Debugging

If the webview doesn't show:
1. Open Developer Tools in the Extension Host (Help > Toggle Developer Tools)
2. Check Console for errors
3. Verify CLI binary exists: `ls -la /Users/oskari/repos/react-analyzer/react-analyzer`
4. Check CLI output manually: `./react-analyzer -mermaid <file>`

If metadata is missing:
1. Check the metadata regex in `parseMermaidMetadata()`
2. Verify Mermaid comments are correctly formatted in CLI output
3. Console log the `metadata` object after parsing

## Success Criteria

The MVP is ready for demo when:
- ✅ All test cases pass
- ✅ Graph renders for SimpleDrilling.tsx
- ✅ Click to jump to source works
- ✅ Detail panel shows correct information
- ✅ Visual styling matches VS Code theme
- ✅ No console errors
