import React, { useState } from 'react';

/**
 * Test fixture for multiple prop renamings that triggers prop drilling violation
 *
 * This demonstrates prop drilling detection with multiple props being renamed:
 * - Root defines state: user (object with name and email)
 * - Layer1 renames user to userData
 * - Layer2 renames userData to profile
 * - Layer3 renames profile to person
 * - Leaf uses the person prop
 *
 * Expected behavior:
 * - Should detect prop drilling violation (3 passthrough components)
 * - Chain highlighting should work across all renames
 * - Each rename should show in edge detail panel
 */

interface LeafProps {
  person: { name: string; email: string };
}

// Leaf component that uses the state
function Leaf({ person }: LeafProps) {
  return (
    <div>
      <h2>{person.name}</h2>
      <p>{person.email}</p>
    </div>
  );
}

interface Layer3Props {
  profile: { name: string; email: string };
}

// Passthrough layer 3 - renames profile to person
function Layer3({ profile }: Layer3Props) {
  return <Leaf person={profile} />;
}

interface Layer2Props {
  userData: { name: string; email: string };
}

// Passthrough layer 2 - renames userData to profile
function Layer2({ userData }: Layer2Props) {
  return <Layer3 profile={userData} />;
}

interface Layer1Props {
  user: { name: string; email: string };
}

// Passthrough layer 1 - renames user to userData
function Layer1({ user }: Layer1Props) {
  return <Layer2 userData={user} />;
}

// Root component with the actual state
function Root() {
  const [user, setUser] = useState({ name: 'John Doe', email: 'john@example.com' });

  return <Layer1 user={user} />;
}

export default Root;
