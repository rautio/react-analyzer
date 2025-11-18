package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleResolver_ResolveWithAliases(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()

	// Create directory structure
	srcDir := filepath.Join(tmpDir, "src")
	componentsDir := filepath.Join(srcDir, "components")
	if err := os.MkdirAll(componentsDir, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Create tsconfig.json with aliases
	tsconfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"]
    }
  }
}`
	tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
	if err := os.WriteFile(tsconfigPath, []byte(tsconfig), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig.json: %v", err)
	}

	// Create test files
	buttonContent := `export const Button = () => <button>Click</button>;`
	buttonPath := filepath.Join(componentsDir, "Button.tsx")
	if err := os.WriteFile(buttonPath, []byte(buttonContent), 0644); err != nil {
		t.Fatalf("Failed to write Button.tsx: %v", err)
	}

	utilsContent := `export const formatDate = (date) => date.toString();`
	utilsPath := filepath.Join(srcDir, "utils.ts")
	if err := os.WriteFile(utilsPath, []byte(utilsContent), 0644); err != nil {
		t.Fatalf("Failed to write utils.ts: %v", err)
	}

	// Create App.tsx that imports using aliases
	appContent := `
import { Button } from '@components/Button';
import { formatDate } from '@/utils';

function App() {
  return <Button />;
}
`
	appPath := filepath.Join(srcDir, "App.tsx")
	if err := os.WriteFile(appPath, []byte(appContent), 0644); err != nil {
		t.Fatalf("Failed to write App.tsx: %v", err)
	}

	// Create resolver
	resolver, err := NewModuleResolver(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Verify aliases were loaded
	if len(resolver.aliases) == 0 {
		t.Fatal("Expected aliases to be loaded from tsconfig.json")
	}

	t.Logf("Loaded %d aliases", len(resolver.aliases))
	for prefix, target := range resolver.aliases {
		t.Logf("  %s -> %s", prefix, target)
	}

	// Test resolving @components/Button
	resolved, err := resolver.Resolve(appPath, "@components/Button")
	if err != nil {
		t.Errorf("Failed to resolve @components/Button: %v", err)
	} else {
		expectedPath, _ := filepath.Abs(buttonPath)
		if resolved != expectedPath {
			t.Errorf("Expected %s, got %s", expectedPath, resolved)
		} else {
			t.Logf("✓ Resolved @components/Button -> %s", resolved)
		}
	}

	// Test resolving @/utils
	resolved, err = resolver.Resolve(appPath, "@/utils")
	if err != nil {
		t.Errorf("Failed to resolve @/utils: %v", err)
	} else {
		expectedPath, _ := filepath.Abs(utilsPath)
		if resolved != expectedPath {
			t.Errorf("Expected %s, got %s", expectedPath, resolved)
		} else {
			t.Logf("✓ Resolved @/utils -> %s", resolved)
		}
	}

	// Test that external packages are still rejected
	_, err = resolver.Resolve(appPath, "react")
	if err == nil {
		t.Error("Expected error for external package 'react', but resolution succeeded")
	} else {
		t.Logf("✓ Correctly rejected external package: %v", err)
	}

	// Test relative imports still work
	resolved, err = resolver.Resolve(appPath, "./utils")
	if err != nil {
		t.Errorf("Failed to resolve relative import ./utils: %v", err)
	} else {
		expectedPath, _ := filepath.Abs(utilsPath)
		if resolved != expectedPath {
			t.Errorf("Expected %s, got %s", expectedPath, resolved)
		} else {
			t.Logf("✓ Resolved relative import ./utils -> %s", resolved)
		}
	}
}

func TestModuleResolver_ReactAnalyzerConfigOverride(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()

	// Create directories
	srcDir := filepath.Join(tmpDir, "src")
	appDir := filepath.Join(tmpDir, "app")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("Failed to create app directory: %v", err)
	}

	// Create tsconfig.json pointing to src
	tsconfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/": ["src/"]
    }
  }
}`
	tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
	if err := os.WriteFile(tsconfigPath, []byte(tsconfig), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig.json: %v", err)
	}

	// Create .reactanalyzer.json pointing to app (should override)
	reactConfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/": ["app/"]
    }
  }
}`
	configPath := filepath.Join(tmpDir, ".reactanalyzer.json")
	if err := os.WriteFile(configPath, []byte(reactConfig), 0644); err != nil {
		t.Fatalf("Failed to write .reactanalyzer.json: %v", err)
	}

	// Create test file in app directory
	utilsContent := `export const helper = () => {};`
	utilsPath := filepath.Join(appDir, "utils.ts")
	if err := os.WriteFile(utilsPath, []byte(utilsContent), 0644); err != nil {
		t.Fatalf("Failed to write utils.ts: %v", err)
	}

	// Create test file that imports using @/
	testContent := `import { helper } from '@/utils';`
	testPath := filepath.Join(tmpDir, "test.tsx")
	if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test.tsx: %v", err)
	}

	// Create resolver
	resolver, err := NewModuleResolver(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Test that @/ resolves to app/ (from .reactanalyzer.json) not src/ (from tsconfig.json)
	resolved, err := resolver.Resolve(testPath, "@/utils")
	if err != nil {
		t.Fatalf("Failed to resolve @/utils: %v", err)
	}

	expectedPath, _ := filepath.Abs(utilsPath)
	if resolved != expectedPath {
		t.Errorf("Expected .reactanalyzer.json to override tsconfig.json")
		t.Errorf("Expected: %s", expectedPath)
		t.Errorf("Got:      %s", resolved)
	} else {
		t.Logf("✓ .reactanalyzer.json correctly overrode tsconfig.json")
		t.Logf("  Resolved @/utils -> %s", resolved)
	}
}
