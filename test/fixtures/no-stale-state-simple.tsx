import React, { useState } from 'react';

// ❌ VIOLATION: Arithmetic operations
export function Counter() {
  const [count, setCount] = useState(0);

  const increment = () => setCount(count + 1);     // Line 6
  const decrement = () => setCount(count - 1);     // Line 7
  const double = () => setCount(count * 2);        // Line 8
  const divide = () => setCount(count / 2);        // Line 9
}

// ❌ VIOLATION: Unary operations
export function Toggle() {
  const [isOpen, setIsOpen] = useState(false);

  const toggle = () => setIsOpen(!isOpen);         // Line 16
}

// ❌ VIOLATION: Array operations
export function ItemList() {
  const [items, setItems] = useState([]);

  const addItem = (item) => setItems([...items, item]);           // Line 23
  const removeFirst = () => setItems(items.slice(1));             // Line 24
  const filtered = () => setItems(items.filter(x => x.active));   // Line 25
}

// ❌ VIOLATION: Object operations
export function UserProfile() {
  const [user, setUser] = useState({ name: '', age: 0 });

  const incrementAge = () => setUser({ ...user, age: user.age + 1 });  // Line 32
  const updateName = (name) => setUser({ ...user, name });             // Line 33
}

// ❌ VIOLATION: Multiple updates in same function
export function MultipleUpdates() {
  const [count, setCount] = useState(0);

  const incrementTwice = () => {
    setCount(count + 1);  // Line 41
    setCount(count + 1);  // Line 42 - will use same stale value!
  };
}

// ❌ VIOLATION: Multi-line function body
export function ComplexHandler() {
  const [count, setCount] = useState(0);
  const [total, setTotal] = useState(0);

  const handleClick = () => {
    // Some complex logic before the setState
    const multiplier = 2;
    const offset = 5;
    const calculated = (count * multiplier) + offset;

    console.log('Updating count');
    setCount(count + 1);  // Line 56 - Still a violation!

    // More logic after
    setTotal(total + calculated);  // Line 59 - Another violation!
  };
}

// ✅ VALID: Setting to new value (not derived from previous state)
export function ValidUpdates() {
  const [count, setCount] = useState(0);
  const [name, setName] = useState('');
  const [max, setMax] = useState(100);

  const reset = () => setCount(0);              // OK - literal
  const setMax100 = () => setCount(100);        // OK - literal
  const updateName = (n) => setName(n);         // OK - parameter, not state
  const copyMax = () => setCount(max);          // OK - different state variable
}

// ✅ VALID: Already using functional form
export function CorrectPattern() {
  const [count, setCount] = useState(0);
  const [isOpen, setIsOpen] = useState(false);
  const [items, setItems] = useState([]);

  const increment = () => setCount(prev => prev + 1);      // OK
  const toggle = () => setIsOpen(prev => !prev);           // OK
  const addItem = (item) => setItems(prev => [...prev, item]);  // OK
}

// ✅ VALID: Multi-line function with functional form
export function ComplexHandlerCorrect() {
  const [count, setCount] = useState(0);
  const [total, setTotal] = useState(0);

  const handleClick = () => {
    const multiplier = 2;
    const offset = 5;

    console.log('Updating count');
    setCount(prev => prev + 1);  // OK - functional form

    setTotal(prev => {
      const calculated = (prev * multiplier) + offset;
      return prev + calculated;
    });  // OK - functional form with block
  };
}
