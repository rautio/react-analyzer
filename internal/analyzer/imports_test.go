package analyzer

import (
	"os"
	"testing"

	"github.com/rautio/react-analyzer/internal/parser"
)

func TestExtractImports(t *testing.T) {
	// Parse a file with various import types
	content := []byte(`
import React from 'react';
import { useState, useEffect } from 'react';
import * as Utils from './utils';
import MyComponent from './MyComponent';
`)

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("test.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	imports := ExtractImports(ast)

	// Should find 4 imports
	if len(imports) != 4 {
		t.Errorf("Expected 4 imports, got %d", len(imports))
	}

	// Check first import: default import
	if imports[0].Source != "react" {
		t.Errorf("Expected source 'react', got '%s'", imports[0].Source)
	}
	if imports[0].Default != "React" {
		t.Errorf("Expected default 'React', got '%s'", imports[0].Default)
	}

	// Check second import: named imports
	if imports[1].Source != "react" {
		t.Errorf("Expected source 'react', got '%s'", imports[1].Source)
	}
	if len(imports[1].Named) != 2 {
		t.Errorf("Expected 2 named imports, got %d", len(imports[1].Named))
	}

	// Check third import: namespace import
	if imports[2].Source != "./utils" {
		t.Errorf("Expected source './utils', got '%s'", imports[2].Source)
	}
	if imports[2].Namespace != "Utils" {
		t.Errorf("Expected namespace 'Utils', got '%s'", imports[2].Namespace)
	}
}

func TestExtractImports_RealFile(t *testing.T) {
	content, err := os.ReadFile("../../test/fixtures/with-hooks.tsx")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("with-hooks.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	imports := ExtractImports(ast)

	// Should find the React import
	if len(imports) == 0 {
		t.Error("Expected to find imports")
	}

	foundReact := false
	for _, imp := range imports {
		if imp.Source == "react" {
			foundReact = true
		}
	}

	if !foundReact {
		t.Error("Expected to find React import")
	}
}

func TestExtractImports_AliasedImports(t *testing.T) {
	code := `
import React from 'react';
import { useState, useEffect as useMount } from 'react';
import { MemoChild as FastChild, AnotherMemo } from './components';
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("test.tsx", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	imports := ExtractImports(ast)

	// Should have 3 imports (2 from react + 1 from ./components)
	if len(imports) != 3 {
		t.Fatalf("Expected 3 imports, got %d", len(imports))
	}

	// Check first import (react default)
	reactImport := imports[0]
	if reactImport.Source != "react" {
		t.Errorf("Expected source 'react', got '%s'", reactImport.Source)
	}
	if reactImport.Default != "React" {
		t.Errorf("Expected default 'React', got '%s'", reactImport.Default)
	}

	// Check second import (react named imports)
	reactNamedImport := imports[1]
	if reactNamedImport.Source != "react" {
		t.Errorf("Expected source 'react', got '%s'", reactNamedImport.Source)
	}
	if len(reactNamedImport.Named) != 2 {
		t.Fatalf("Expected 2 named imports from react, got %d", len(reactNamedImport.Named))
	}

	// Check useState (no alias)
	useState := reactNamedImport.Named[0]
	if useState.ImportedName != "useState" {
		t.Errorf("Expected ImportedName 'useState', got '%s'", useState.ImportedName)
	}
	if useState.LocalName != "useState" {
		t.Errorf("Expected LocalName 'useState', got '%s'", useState.LocalName)
	}

	// Check useEffect (aliased as useMount)
	useEffect := reactNamedImport.Named[1]
	if useEffect.ImportedName != "useEffect" {
		t.Errorf("Expected ImportedName 'useEffect', got '%s'", useEffect.ImportedName)
	}
	if useEffect.LocalName != "useMount" {
		t.Errorf("Expected LocalName 'useMount', got '%s'", useEffect.LocalName)
	}

	// Check third import (./components)
	compImport := imports[2]
	if compImport.Source != "./components" {
		t.Errorf("Expected source './components', got '%s'", compImport.Source)
	}
	if len(compImport.Named) != 2 {
		t.Fatalf("Expected 2 named imports from ./components, got %d", len(compImport.Named))
	}

	// Check MemoChild (aliased as FastChild)
	memoChild := compImport.Named[0]
	if memoChild.ImportedName != "MemoChild" {
		t.Errorf("Expected ImportedName 'MemoChild', got '%s'", memoChild.ImportedName)
	}
	if memoChild.LocalName != "FastChild" {
		t.Errorf("Expected LocalName 'FastChild', got '%s'", memoChild.LocalName)
	}

	// Check AnotherMemo (no alias)
	anotherMemo := compImport.Named[1]
	if anotherMemo.ImportedName != "AnotherMemo" {
		t.Errorf("Expected ImportedName 'AnotherMemo', got '%s'", anotherMemo.ImportedName)
	}
	if anotherMemo.LocalName != "AnotherMemo" {
		t.Errorf("Expected LocalName 'AnotherMemo', got '%s'", anotherMemo.LocalName)
	}
}
