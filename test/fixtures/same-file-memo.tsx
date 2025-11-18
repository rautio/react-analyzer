import React from 'react';

// Memoized child component defined in the same file
const MemoChild = React.memo(function Child({ config, items, onUpdate }: any) {
  return <div>{config.theme}</div>;
});

// Parent component passing unstable props to same-file memoized component
function ParentWithSameFileMemo() {
  return (
    <div>
      <MemoChild config={{ theme: 'dark' }} /> {/* Should detect: unstable object breaks memo */}
      <MemoChild items={[1, 2, 3]} /> {/* Should detect: unstable array breaks memo */}
      <MemoChild onUpdate={() => console.log('update')} /> {/* Should detect: unstable function breaks memo */}
    </div>
  );
}

// Another memoized component
const AnotherMemo = React.memo(({ data }: { data: any }) => {
  return <span>{data}</span>;
});

// Using the second memoized component
function AnotherParent() {
  return <AnotherMemo data={{ value: 42 }} />; {/* Should detect: unstable object */}
}
