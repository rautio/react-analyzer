import React, { useEffect, useState } from 'react';

function ComponentWithHooks() {
  const [count, setCount] = useState(0);

  const config = { theme: 'dark' };

  useEffect(() => {
    console.log('Count changed:', count);
  }, [count, config]);

  return <div>Count: {count}</div>;
}

export default ComponentWithHooks;
