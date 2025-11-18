import React from 'react';

// Memoized component that will be imported with an alias
export const MemoChild = React.memo(function Child({ config, items }: any) {
  return <div>{config.theme}</div>;
});

// Another memoized component
export const AnotherMemo = React.memo(({ data }: any) => {
  return <span>{data}</span>;
});
