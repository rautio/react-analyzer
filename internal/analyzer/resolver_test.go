package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleResolver_Resolve(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "resolver-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)

	appFile := filepath.Join(srcDir, "App.tsx")
	componentFile := filepath.Join(srcDir, "Component.tsx")

	os.WriteFile(appFile, []byte("export const App = () => <div />;"), 0644)
	os.WriteFile(componentFile, []byte("export const Component = () => <div />;"), 0644)

	// Create resolver
	resolver, err := NewModuleResolver(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Test resolving relative import
	resolved, err := resolver.Resolve(appFile, "./Component")
	if err != nil {
		t.Errorf("Failed to resolve ./Component: %v", err)
	}

	expectedPath, _ := filepath.Abs(componentFile)
	if resolved != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, resolved)
	}
}

func TestModuleResolver_GetModule(t *testing.T) {
	resolver, err := NewModuleResolver("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Get a test module
	fixturePath, _ := filepath.Abs("../../test/fixtures/with-hooks.tsx")
	module, err := resolver.GetModule(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get module: %v", err)
	}

	if module == nil {
		t.Fatal("Expected module, got nil")
	}

	if module.AST == nil {
		t.Error("Expected AST in module")
	}

	// Should have found imports
	if len(module.Imports) == 0 {
		t.Error("Expected to find imports in with-hooks.tsx")
	}

	// Test caching - second call should return same module
	module2, err := resolver.GetModule(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get cached module: %v", err)
	}

	if module != module2 {
		t.Error("Expected cached module to be same instance")
	}
}

func TestModuleResolver_ResolveExternal(t *testing.T) {
	resolver, err := NewModuleResolver(".")
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Should skip external packages
	_, err = resolver.Resolve("/some/file.tsx", "react")
	if err == nil {
		t.Error("Expected error for external package")
	}
}
