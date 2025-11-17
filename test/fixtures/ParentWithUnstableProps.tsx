import React from 'react';
import { MemoChild, AnotherMemo } from './MemoChild';

// Parent component passing unstable props to memoized children
function ParentWithUnstableProps() {
  return (
    <div>
      {/* BAD: Passing inline object to memoized component */}
      <MemoChild
        config={{ title: 'Hello' }}
        items={['a', 'b', 'c']}
        onUpdate={() => console.log('update')}
      />

      {/* BAD: Another inline object */}
      <AnotherMemo data={{ value: 42 }} />
    </div>
  );
}

export default ParentWithUnstableProps;
