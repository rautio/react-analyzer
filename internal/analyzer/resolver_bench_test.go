package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkConcurrentParsing benchmarks concurrent file parsing with the parser pool
func BenchmarkConcurrentParsing(b *testing.B) {
	// Create test files
	tmpDir := b.TempDir()

	// Create multiple test files with realistic React components
	testFiles := make([]string, 10)
	componentCode := `import React from 'react';
import { useState } from 'react';

export const TestComponent: React.FC = () => {
	const [count, setCount] = useState(0);
	const [name, setName] = useState('');

	const handleClick = () => {
		setCount(count + 1);
	};

	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		setName(e.target.value);
	};

	return (
		<div>
			<h1>Count: {count}</h1>
			<button onClick={handleClick}>Increment</button>
			<input value={name} onChange={handleChange} />
		</div>
	);
};
`

	for i := 0; i < len(testFiles); i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("Component%d.tsx", i))
		if err := os.WriteFile(filePath, []byte(componentCode), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		testFiles[i] = filePath
	}

	resolver, err := NewModuleResolver(tmpDir)
	if err != nil {
		b.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	b.ResetTimer()

	// Benchmark parsing files concurrently
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Round-robin through test files
			filePath := testFiles[i%len(testFiles)]
			_, err := resolver.GetModule(filePath)
			if err != nil {
				b.Fatalf("Failed to parse file: %v", err)
			}
			i++
		}
	})
}

// BenchmarkSequentialParsing benchmarks sequential file parsing for comparison
func BenchmarkSequentialParsing(b *testing.B) {
	// Create test files
	tmpDir := b.TempDir()

	// Create multiple test files with realistic React components
	testFiles := make([]string, 10)
	componentCode := `import React from 'react';
import { useState } from 'react';

export const TestComponent: React.FC = () => {
	const [count, setCount] = useState(0);
	const [name, setName] = useState('');

	const handleClick = () => {
		setCount(count + 1);
	};

	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		setName(e.target.value);
	};

	return (
		<div>
			<h1>Count: {count}</h1>
			<button onClick={handleClick}>Increment</button>
			<input value={name} onChange={handleChange} />
		</div>
	);
};
`

	for i := 0; i < len(testFiles); i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("Component%d.tsx", i))
		if err := os.WriteFile(filePath, []byte(componentCode), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		testFiles[i] = filePath
	}

	resolver, err := NewModuleResolver(tmpDir)
	if err != nil {
		b.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	b.ResetTimer()

	// Benchmark parsing files sequentially
	for i := 0; i < b.N; i++ {
		filePath := testFiles[i%len(testFiles)]
		_, err := resolver.GetModule(filePath)
		if err != nil {
			b.Fatalf("Failed to parse file: %v", err)
		}
	}
}

// BenchmarkSingleFileParsing benchmarks parsing a single file (measures per-file overhead)
func BenchmarkSingleFileParsing(b *testing.B) {
	tmpDir := b.TempDir()

	componentCode := `import React from 'react';
import { useState } from 'react';

export const TestComponent: React.FC = () => {
	const [count, setCount] = useState(0);
	return <div><h1>Count: {count}</h1></div>;
};
`

	filePath := filepath.Join(tmpDir, "Component.tsx")
	if err := os.WriteFile(filePath, []byte(componentCode), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	resolver, err := NewModuleResolver(tmpDir)
	if err != nil {
		b.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear cache to force re-parsing each time
		resolver.modules = make(map[string]*Module)
		_, err := resolver.GetModule(filePath)
		if err != nil {
			b.Fatalf("Failed to parse file: %v", err)
		}
	}
}
