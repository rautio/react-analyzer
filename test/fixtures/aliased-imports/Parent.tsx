import React from 'react';
// Import memoized components with aliases
import { MemoChild as FastChild, AnotherMemo as QuickMemo } from './MemoComponent';

// Parent component using aliased imports
function ParentWithAliasedImports() {
  return (
    <div>
      {/* Should detect: unstable object to FastChild (aliased from MemoChild) */}
      <FastChild config={{ theme: 'dark' }} />

      {/* Should detect: unstable array to FastChild */}
      <FastChild items={[1, 2, 3]} />

      {/* Should detect: unstable object to QuickMemo (aliased from AnotherMemo) */}
      <QuickMemo data={{ value: 42 }} />
    </div>
  );
}

export default ParentWithAliasedImports;
