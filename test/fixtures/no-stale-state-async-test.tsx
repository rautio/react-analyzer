import React, { useState, useEffect } from 'react';

// All of these SHOULD be caught by Phase 1

// setTimeout
export function WithTimeout() {
  const [count, setCount] = useState(0);

  const delayed = () => {
    setTimeout(() => {
      setCount(count + 1);  // Line 10 - Should be caught!
    }, 1000);
  };
}

// setInterval
export function WithInterval() {
  const [seconds, setSeconds] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setSeconds(seconds + 1);  // Line 21 - Should be caught!
    }, 1000);

    return () => clearInterval(interval);
  }, []);
}

// Promise callbacks
export function WithPromise() {
  const [items, setItems] = useState([]);

  const loadMore = () => {
    fetch('/api/items').then(data => {
      setItems([...items, data]);  // Line 35 - Should be caught!
    });
  };
}

// Async/await
export function WithAsync() {
  const [items, setItems] = useState([]);

  const loadMore = async () => {
    const data = await fetch('/api/items');
    setItems([...items, data]);  // Line 47 - Should be caught!
  };
}

// useEffect
export function WithEffect() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    setCount(count + 1);  // Line 56 - Should be caught!
  }, []);
}
