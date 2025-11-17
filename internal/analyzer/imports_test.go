package analyzer

import (
	"os"
	"testing"

	"github.com/oskari/react-analyzer/internal/parser"
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
