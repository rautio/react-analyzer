# Contributing to React Analyzer

Thanks for your interest in contributing! This document provides guidelines and information for developers working on React Analyzer.

## Development Status

**Current Phase:** MVP Development (v0.1.0)

- [x] CLI interface and argument parsing
- [x] File validation and error handling
- [x] Help and version commands
- [ ] Tree-sitter parser integration (in progress)
- [ ] AST traversal and semantic analysis
- [ ] `no-object-deps` rule implementation
- [ ] Output formatting with diagnostics

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, for convenience commands)

### Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/rautio/react-analyzer
   cd react-analyzer
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the project:**
   ```bash
   go build -o react-analyzer ./cmd/react-analyzer
   ```

4. **Run tests:**
   ```bash
   go test ./...
   ```

5. **Test the CLI:**
   ```bash
   ./react-analyzer test/fixtures/simple.tsx
   ```

## Project Structure

```
react-analyzer/
├── cmd/
│   └── react-analyzer/         # CLI entry point
│       └── main.go             # Argument parsing, orchestration
├── internal/
│   ├── cli/                    # CLI implementation
│   │   ├── help.go             # Help text
│   │   └── runner.go           # Analysis orchestration
│   ├── parser/                 # Tree-sitter wrapper (planned)
│   │   ├── parser.go           # Parser interface
│   │   └── node.go             # AST node wrapper
│   ├── analyzer/               # Analysis logic (planned)
│   │   ├── analyzer.go         # Main analyzer
│   │   └── scope.go            # Scope tracking
│   └── rules/                  # Rule implementations (planned)
│       ├── rule.go             # Rule interface
│       └── no_object_deps.go   # First MVP rule
├── test/
│   └── fixtures/               # Test React files
├── prompts/                    # Design documents
│   ├── technical_design.md     # Architecture overview
│   ├── mvp_scope.md            # MVP scope definition
│   └── interfaces.md           # CLI interface spec
├── go.mod
├── go.sum
└── README.md
```

## Architecture

### Technology Stack

- **Go 1.21+** - Performance, concurrency, easy distribution
- **tree-sitter** - Fast, incremental parsing of TypeScript/JSX
- **Standard library** - Minimal dependencies for MVP

### Design Principles

1. **Incremental Development** - Build working features iteratively
2. **Simple First** - Avoid premature abstraction
3. **Test-Driven** - Write tests alongside implementation
4. **User-Focused** - Optimize for clear error messages and fast feedback

### Key Components

**1. CLI Layer** (`cmd/react-analyzer/`, `internal/cli/`)
- Argument parsing using standard library `flag` package
- File validation
- Output formatting
- Exit code handling

**2. Parser Layer** (`internal/parser/`) - *In Progress*
- tree-sitter wrapper
- AST node abstraction
- Semantic helpers (IsComponent, IsHookCall, etc.)

**3. Analyzer Layer** (`internal/analyzer/`) - *Planned*
- Scope analysis
- Symbol tracking
- Variable stability detection

**4. Rule Engine** (`internal/rules/`) - *Planned*
- Rule interface
- Rule registration
- Diagnostic generation

## Development Workflow

### Making Changes

1. **Create a branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write code
   - Add tests
   - Update documentation

3. **Test your changes:**
   ```bash
   # Run tests
   go test ./...

   # Build and test CLI
   go build -o react-analyzer ./cmd/react-analyzer
   ./react-analyzer test/fixtures/simple.tsx
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "Add feature: description"
   ```

### Testing

**Run all tests:**
```bash
go test ./...
```

**Run with coverage:**
```bash
go test -cover ./...
```

**Run with race detection:**
```bash
go test -race ./...
```

**Test CLI manually:**
```bash
# Build first
go build -o react-analyzer ./cmd/react-analyzer

# Test various scenarios
./react-analyzer --help
./react-analyzer --version
./react-analyzer test/fixtures/simple.tsx
./react-analyzer --verbose test/fixtures/simple.tsx
./react-analyzer missing.tsx  # Should error
```

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Write clear, self-documenting code
- Add comments for complex logic
- Keep functions small and focused

### Adding a Test Fixture

Create test React files in `test/fixtures/`:

```bash
cat > test/fixtures/my-test.tsx << 'EOF'
import React from 'react';

function MyComponent() {
  return <div>Test</div>;
}

export default MyComponent;
EOF
```

## Roadmap

### Phase 1: MVP (Current - Weeks 1-2)

**Goal:** Working CLI with one rule

- [x] CLI interface (Week 1, Days 1-2)
- [ ] Parser integration (Week 1, Days 3-4)
- [ ] Scope analysis (Week 1, Day 5)
- [ ] `no-object-deps` rule (Week 2, Days 1-2)
- [ ] Output formatter (Week 2, Days 3-4)
- [ ] Testing & polish (Week 2, Day 5)

**Deliverable:** CLI tool that detects inline objects in hook dependencies

### Phase 2: Additional Rules (Weeks 3-4)

- [ ] `no-unstable-props` rule
- [ ] Configuration file support
- [ ] Multiple file analysis
- [ ] Enhanced error messages

### Phase 3: IDE Integration (Weeks 5-8)

- [ ] Language Server Protocol (LSP) implementation
- [ ] VS Code extension
- [ ] Real-time analysis
- [ ] Auto-fix suggestions

### Phase 4: Advanced Features (Future)

- [ ] Cross-file analysis (`memo-unstable-props`)
- [ ] Import resolution
- [ ] Component graph visualization
- [ ] Performance profiling integration

## Performance Targets

Based on POC validation:

| Operation | Target | Current Status |
|-----------|--------|----------------|
| Parse 30-line file | <1ms | ✅ 80μs (POC) |
| Parse 233-line file | <5ms | ✅ 1.85ms (POC) |
| Single file analysis | <100ms | 🎯 On track |
| Full project (10k LOC) | <10s | 🎯 On track |

See `tree-sitter-poc/POC_RESULTS.md` for detailed benchmarks.

## Design Documents

Key design documents in `prompts/`:

- **`technical_design.md`** - Overall architecture and technology decisions
- **`mvp_scope.md`** - MVP scope and simplifications
- **`interfaces.md`** - CLI interface specification
- **`implementation_plan.md`** - Detailed implementation phases

## Release Process

(To be defined as we approach first release)

1. Update version in `cmd/react-analyzer/main.go`
2. Update CHANGELOG.md
3. Tag release: `git tag v0.1.0`
4. Build binaries for all platforms
5. Create GitHub release

## Getting Help

- **Questions?** Open a GitHub issue
- **Bugs?** Open a GitHub issue with reproduction steps
- **Ideas?** Open a GitHub discussion

## Code of Conduct

Be respectful, collaborative, and constructive. We're all here to build something useful together.

## License

TBD
