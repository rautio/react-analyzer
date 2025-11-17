import React, { useMemo, useCallback } from 'react';
import { MemoChild, AnotherMemo } from './MemoChild';

// Parent component properly using memoized props
function ParentWithStableProps() {
  // GOOD: Memoized object
  const config = useMemo(() => ({ title: 'Hello' }), []);

  // GOOD: Memoized array
  const items = useMemo(() => ['a', 'b', 'c'], []);

  // GOOD: Memoized callback
  const handleUpdate = useCallback(() => console.log('update'), []);

  // GOOD: Memoized data
  const data = useMemo(() => ({ value: 42 }), []);

  return (
    <div>
      <MemoChild
        config={config}
        items={items}
        onUpdate={handleUpdate}
      />

      <AnotherMemo data={data} />
    </div>
  );
}

export default ParentWithStableProps;
