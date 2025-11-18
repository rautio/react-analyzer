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

func TestModuleResolver_MonorepoSupport(t *testing.T) {
	// Create temporary monorepo structure:
	// /
	//   packages/
	//     package-a/
	//       tsconfig.json (aliases: @pkgA/* -> src/*)
	//       src/
	//         utils.ts
	//         App.tsx (imports @pkgA/utils)
	//         nested/
	//           Component.tsx (imports @pkgA/utils)
	//     package-b/
	//       .reactanalyzer.json (aliases: @pkgB/* -> lib/*)
	//       lib/
	//         helpers.ts
	//       Main.tsx (imports @pkgB/helpers)

	tmpDir := t.TempDir()

	// Create package-a structure
	pkgADir := filepath.Join(tmpDir, "packages", "package-a")
	pkgASrcDir := filepath.Join(pkgADir, "src")
	pkgANestedDir := filepath.Join(pkgASrcDir, "nested")
	if err := os.MkdirAll(pkgANestedDir, 0755); err != nil {
		t.Fatalf("Failed to create package-a directories: %v", err)
	}

	// Create package-a tsconfig.json
	pkgATsconfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@pkgA/*": ["src/*"]
    }
  }
}`
	pkgATsconfigPath := filepath.Join(pkgADir, "tsconfig.json")
	if err := os.WriteFile(pkgATsconfigPath, []byte(pkgATsconfig), 0644); err != nil {
		t.Fatalf("Failed to write package-a tsconfig.json: %v", err)
	}

	// Create package-a files
	pkgAUtilsContent := `export const helperA = () => 'A';`
	pkgAUtilsPath := filepath.Join(pkgASrcDir, "utils.ts")
	if err := os.WriteFile(pkgAUtilsPath, []byte(pkgAUtilsContent), 0644); err != nil {
		t.Fatalf("Failed to write package-a utils.ts: %v", err)
	}

	pkgAAppContent := `import { helperA } from '@pkgA/utils';`
	pkgAAppPath := filepath.Join(pkgASrcDir, "App.tsx")
	if err := os.WriteFile(pkgAAppPath, []byte(pkgAAppContent), 0644); err != nil {
		t.Fatalf("Failed to write package-a App.tsx: %v", err)
	}

	pkgAComponentContent := `import { helperA } from '@pkgA/utils';`
	pkgAComponentPath := filepath.Join(pkgANestedDir, "Component.tsx")
	if err := os.WriteFile(pkgAComponentPath, []byte(pkgAComponentContent), 0644); err != nil {
		t.Fatalf("Failed to write package-a Component.tsx: %v", err)
	}

	// Create package-b structure
	pkgBDir := filepath.Join(tmpDir, "packages", "package-b")
	pkgBLibDir := filepath.Join(pkgBDir, "lib")
	if err := os.MkdirAll(pkgBLibDir, 0755); err != nil {
		t.Fatalf("Failed to create package-b directories: %v", err)
	}

	// Create package-b .reactanalyzer.json
	pkgBConfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@pkgB/*": ["lib/*"]
    }
  }
}`
	pkgBConfigPath := filepath.Join(pkgBDir, ".reactanalyzer.json")
	if err := os.WriteFile(pkgBConfigPath, []byte(pkgBConfig), 0644); err != nil {
		t.Fatalf("Failed to write package-b .reactanalyzer.json: %v", err)
	}

	// Create package-b files
	pkgBHelpersContent := `export const helperB = () => 'B';`
	pkgBHelpersPath := filepath.Join(pkgBLibDir, "helpers.ts")
	if err := os.WriteFile(pkgBHelpersPath, []byte(pkgBHelpersContent), 0644); err != nil {
		t.Fatalf("Failed to write package-b helpers.ts: %v", err)
	}

	pkgBMainContent := `import { helperB } from '@pkgB/helpers';`
	pkgBMainPath := filepath.Join(pkgBDir, "Main.tsx")
	if err := os.WriteFile(pkgBMainPath, []byte(pkgBMainContent), 0644); err != nil {
		t.Fatalf("Failed to write package-b Main.tsx: %v", err)
	}

	// Create resolver with monorepo root
	resolver, err := NewModuleResolver(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Test 1: Resolve @pkgA/utils from package-a App.tsx
	t.Log("Test 1: Resolve @pkgA/utils from package-a/src/App.tsx")
	resolved, err := resolver.Resolve(pkgAAppPath, "@pkgA/utils")
	if err != nil {
		t.Errorf("Failed to resolve @pkgA/utils from package-a: %v", err)
	} else {
		expectedPath, _ := filepath.Abs(pkgAUtilsPath)
		if resolved != expectedPath {
			t.Errorf("Package-a: Expected %s, got %s", expectedPath, resolved)
		} else {
			t.Logf("✓ Package-a resolved @pkgA/utils -> %s", resolved)
		}
	}

	// Test 2: Resolve @pkgA/utils from package-a nested Component.tsx (test caching)
	t.Log("Test 2: Resolve @pkgA/utils from package-a/src/nested/Component.tsx (caching test)")
	resolved, err = resolver.Resolve(pkgAComponentPath, "@pkgA/utils")
	if err != nil {
		t.Errorf("Failed to resolve @pkgA/utils from package-a nested: %v", err)
	} else {
		expectedPath, _ := filepath.Abs(pkgAUtilsPath)
		if resolved != expectedPath {
			t.Errorf("Package-a nested: Expected %s, got %s", expectedPath, resolved)
		} else {
			t.Logf("✓ Package-a nested resolved @pkgA/utils -> %s", resolved)
		}
	}

	// Test 3: Resolve @pkgB/helpers from package-b Main.tsx
	t.Log("Test 3: Resolve @pkgB/helpers from package-b/Main.tsx")
	resolved, err = resolver.Resolve(pkgBMainPath, "@pkgB/helpers")
	if err != nil {
		t.Errorf("Failed to resolve @pkgB/helpers from package-b: %v", err)
	} else {
		expectedPath, _ := filepath.Abs(pkgBHelpersPath)
		if resolved != expectedPath {
			t.Errorf("Package-b: Expected %s, got %s", expectedPath, resolved)
		} else {
			t.Logf("✓ Package-b resolved @pkgB/helpers -> %s", resolved)
		}
	}

	// Test 4: Verify package-a can't resolve package-b aliases
	t.Log("Test 4: Verify package-a cannot resolve package-b aliases")
	_, err = resolver.Resolve(pkgAAppPath, "@pkgB/helpers")
	if err == nil {
		t.Error("Expected package-a to fail resolving package-b alias @pkgB/helpers")
	} else {
		t.Logf("✓ Correctly rejected cross-package alias: %v", err)
	}

	// Test 5: Verify package-b can't resolve package-a aliases
	t.Log("Test 5: Verify package-b cannot resolve package-a aliases")
	_, err = resolver.Resolve(pkgBMainPath, "@pkgA/utils")
	if err == nil {
		t.Error("Expected package-b to fail resolving package-a alias @pkgA/utils")
	} else {
		t.Logf("✓ Correctly rejected cross-package alias: %v", err)
	}

	// Test 6: Verify cache is being used (check internal state)
	t.Log("Test 6: Verify alias cache has entries for all accessed directories")
	resolver.aliasCacheMu.RLock()
	cacheSize := len(resolver.aliasCache)
	resolver.aliasCacheMu.RUnlock()

	if cacheSize < 3 {
		t.Errorf("Expected at least 3 cache entries (package-a dir, package-a src, package-b dir), got %d", cacheSize)
	} else {
		t.Logf("✓ Alias cache has %d entries (proper caching)", cacheSize)
	}
}
