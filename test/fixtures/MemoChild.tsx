import React, { memo } from 'react';

// Memoized component expecting stable props
export const MemoChild = memo(({ config, items, onUpdate }) => {
  return (
    <div>
      <h1>{config.title}</h1>
      <ul>
        {items.map(item => <li key={item}>{item}</li>)}
      </ul>
      <button onClick={onUpdate}>Update</button>
    </div>
  );
});

// Another memoized component
export const AnotherMemo = React.memo(({ data }) => {
  return <span>{data.value}</span>;
});
