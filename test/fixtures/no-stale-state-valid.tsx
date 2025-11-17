import React, { useState } from 'react';

// ✅ VALID: Setting to literal values
export function LiteralValues() {
  const [count, setCount] = useState(0);
  const [isOpen, setIsOpen] = useState(false);

  const reset = () => setCount(0);
  const setToHundred = () => setCount(100);
  const close = () => setIsOpen(false);
  const open = () => setIsOpen(true);
}

// ✅ VALID: Setting from props
export function FromProps({ initialValue, maxValue }) {
  const [count, setCount] = useState(0);

  const resetToInitial = () => setCount(initialValue);
  const setToMax = () => setCount(maxValue);
}

// ✅ VALID: Setting from different state variable
export function DifferentState() {
  const [count, setCount] = useState(0);
  const [max, setMax] = useState(100);
  const [min, setMin] = useState(0);

  const copyMax = () => setCount(max);  // OK - 'max', not 'count'
  const copyMin = () => setCount(min);  // OK - 'min', not 'count'
}

// ✅ VALID: Setting from parameter/variable
export function FromVariable() {
  const [count, setCount] = useState(0);

  const setToValue = (newValue) => setCount(newValue);

  const calculate = () => {
    const result = Math.random() * 100;
    setCount(result);  // OK - local variable, not state
  };
}

// ✅ VALID: Already using functional form
export function FunctionalForm() {
  const [count, setCount] = useState(0);
  const [isOpen, setIsOpen] = useState(false);
  const [items, setItems] = useState([]);
  const [user, setUser] = useState({ name: '', age: 0 });

  const increment = () => setCount(prev => prev + 1);
  const decrement = () => setCount(prev => prev - 1);
  const toggle = () => setIsOpen(prev => !prev);
  const addItem = (item) => setItems(prev => [...prev, item]);
  const updateAge = () => setUser(prev => ({ ...prev, age: prev.age + 1 }));
}

// ✅ VALID: Functional form with function keyword
export function FunctionalFormWithFunction() {
  const [count, setCount] = useState(0);

  const increment = () => setCount(function(prev) {
    return prev + 1;
  });
}

// ✅ VALID: Complex functional form
export function ComplexFunctionalForm() {
  const [count, setCount] = useState(0);

  const complexUpdate = () => {
    setCount(prev => {
      const newValue = prev * 2 + 5;
      console.log('New value:', newValue);
      return newValue;
    });
  };
}
