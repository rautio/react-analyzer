import React, { useState } from 'react';

/**
 * Test fixture for multiple prop renamings in a chain
 *
 * This demonstrates a more complex scenario with multiple props being renamed:
 * - Root defines onEdit and onDelete
 * - Layer1 renames onEdit to onItemEdit
 * - Layer2 renames onItemEdit to handleEdit
 * - Layer3 renames handleEdit back to onEdit
 * - Delete handler follows a different renaming path
 *
 * Expected behavior:
 * - Each prop chain should be traced independently
 * - Clicking on any edge in the chain should highlight the full path
 * - Renaming info should show at each step
 */

interface LeafProps {
  onEdit: () => void;
  onDelete: () => void;
}

// Leaf component that uses both props
function Leaf({ onEdit, onDelete }: LeafProps) {
  return (
    <div>
      <button onClick={onEdit}>Edit</button>
      <button onClick={onDelete}>Delete</button>
    </div>
  );
}

interface Layer3Props {
  handleEdit: () => void;
  removeItem: () => void;
}

// Passthrough layer 3 - renames back to onEdit and to onDelete
function Layer3({ handleEdit, removeItem }: Layer3Props) {
  return <Leaf onEdit={handleEdit} onDelete={removeItem} />;
}

interface Layer2Props {
  onItemEdit: () => void;
  removeItem: () => void;
}

// Passthrough layer 2 - renames to handleEdit, keeps removeItem
function Layer2({ onItemEdit, removeItem }: Layer2Props) {
  return <Layer3 handleEdit={onItemEdit} removeItem={removeItem} />;
}

interface Layer1Props {
  onEdit: () => void;
  onDelete: () => void;
}

// Passthrough layer 1 - renames both props
function Layer1({ onEdit, onDelete }: Layer1Props) {
  return <Layer2 onItemEdit={onEdit} removeItem={onDelete} />;
}

// Root component with the actual state
function Root() {
  const [item, setItem] = useState({ name: 'Test Item', id: 1 });

  const handleEdit = () => {
    setItem({ ...item, name: 'Edited Item' });
  };

  const handleDelete = () => {
    setItem(null);
  };

  return <Layer1 onEdit={handleEdit} onDelete={handleDelete} />;
}

export default Root;
