# Release Process

This document describes how to create a new release of react-analyzer.

## Release Architecture

**Why This Approach?**

react-analyzer uses `go-tree-sitter`, which requires CGO (C bindings). This makes cross-compilation complex because it needs platform-specific C compilers. Our release strategy uses **native platform builds**, which is the standard pattern for CGO projects:

- **Build**: Use `goreleaser build --single-target` on native runners (macOS/Linux/Windows)
- **Package**: Manually create archives and checksums
- **Release**: Use GitHub Actions to publish

This is the recommended approach for open-source CGO projects using free GoReleaser. To use GoReleaser for packaging, you'd need GoReleaser Pro's `builder: prebuilt` feature (~$300-600/year).

## Prerequisites

- Push access to the repository
- All changes merged to `main` branch
- All tests passing
- Changelog is up to date (or will be auto-generated)

## Creating a Release

The release process is automated using GitHub Actions with native platform builds.

### 1. Decide on Version Number

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backward-compatible manner
- **PATCH** version for backward-compatible bug fixes

### 2. Update Version in Code (Optional)

The version in `cmd/react-analyzer/main.go` can be updated to match the release:

```go
const Version = "X.Y.Z"
```

This is optional as the version will also be set via ldflags during build.

### 3. Create and Push a Git Tag

```bash
# Ensure you're on main and up to date
git checkout main
git pull

# Create an annotated tag
git tag -a vX.Y.Z -m "Release vX.Y.Z"

# Push the tag to GitHub
git push origin vX.Y.Z
```

**Important:** The tag MUST start with `v` (e.g., `v1.0.0`, not `1.0.0`)

### 4. Automated Build Process

Once the tag is pushed:

1. **Test Job**: Tests run once on Ubuntu to verify code quality
2. **Build Jobs**: Binaries are built on native platforms in parallel (matrix strategy):
   - **macOS builds** run on `macos-latest` (for darwin/amd64 and darwin/arm64)
   - **Linux builds** run on `ubuntu-latest` (for linux/amd64 and linux/arm64)
   - **Windows build** runs on `windows-latest` (for windows/amd64)
   - Each job uses `goreleaser build --single-target` for native compilation with CGO
3. **Release Job**: After all builds complete:
   - Downloads all build artifacts
   - Creates archives (`.tar.gz` for Unix, `.zip` for Windows)
   - Generates SHA256 checksums
   - Creates GitHub Release with:
     - Release notes (auto-generated from commits)
     - All platform binaries attached
     - Installation instructions

**Note on GoReleaser:** While GoReleaser is used for building (`goreleaser build`), the final packaging is done manually in GitHub Actions. This is the standard pattern for CGO projects without GoReleaser Pro. The `.goreleaser.yml` file is primarily for local development testing.

### 5. Verify the Release

1. Go to https://github.com/rautio/react-analyzer/releases
2. Find your new release (e.g., `v1.0.0`)
3. Check that all expected binaries are attached:
   - `react-analyzer_X.Y.Z_darwin_amd64.tar.gz`
   - `react-analyzer_X.Y.Z_darwin_arm64.tar.gz`
   - `react-analyzer_X.Y.Z_linux_amd64.tar.gz`
   - `react-analyzer_X.Y.Z_linux_arm64.tar.gz`
   - `react-analyzer_X.Y.Z_windows_amd64.zip`
   - `checksums.txt`
4. Download and test a binary for your platform

## Testing a Release Locally

You can test the release process locally before pushing a tag:

```bash
# Install goreleaser (if not already installed)
brew install goreleaser

# Validate configuration
goreleaser check

# Build for your current platform only (snapshot mode)
goreleaser build --snapshot --clean --single-target

# Test the binary
./dist/react-analyzer_*/react-analyzer --version
```

**Note:** Local testing only builds for your current platform due to CGO requirements. The full multi-platform build happens automatically in GitHub Actions using native runners for each platform.

## Commit Message Format

To get nice auto-generated changelogs, follow conventional commit format:

- `feat: add new rule for xyz` → "New Features" section
- `fix: resolve issue with abc` → "Bug Fixes" section
- `perf: improve parsing speed` → "Performance Improvements" section
- `docs: update README` → (filtered out)
- `test: add tests for xyz` → (filtered out)
- `chore: update dependencies` → (filtered out)

## Troubleshooting

### Build fails with "no tags found"

**Solution:** Ensure you've pushed the tag: `git push origin vX.Y.Z`

### Build fails with CGO errors

**Solution:** The GitHub Actions workflow builds on native runners which have the necessary build tools for each platform. If testing locally, ensure you have a C compiler installed:
- **macOS**: `xcode-select --install`
- **Linux**: `apt-get install build-essential`
- **Windows**: Install MinGW or TDM-GCC

### GitHub Actions build matrix fails

**Solution:** Check the Actions logs to see which platform failed:
1. Go to https://github.com/rautio/react-analyzer/actions
2. Click on the failed workflow run
3. Check the specific platform job that failed
4. Common issues:
   - Missing dependencies on the runner
   - Platform-specific build flags needed
   - CGO compilation issues (ensure `CGO_ENABLED=1` is set correctly)

### Release is created but binaries are missing

**Solution:** Check the GitHub Actions logs at https://github.com/rautio/react-analyzer/actions to see what failed.

### Need to fix a release

1. Delete the release and tag on GitHub
2. Delete the local tag: `git tag -d vX.Y.Z`
3. Make your fixes
4. Create a new tag (same version or bump patch)
5. Push the new tag

## Manual Release (Fallback)

If automated releases fail, you can release manually:

```bash
# Build locally
goreleaser release --clean

# This will:
# - Run tests
# - Build all binaries
# - Create GitHub release
# - Upload all artifacts
```

**Note:** Manual releases require the `GITHUB_TOKEN` environment variable to be set with a GitHub personal access token that has `repo` scope.

## Future Enhancements

Consider adding:
- Homebrew tap for easier macOS installation: `brew install rautio/tap/react-analyzer`
- npm wrapper for Node.js projects: `npm install -g react-analyzer`
- Docker images for containerized environments
- Automated changelog generation from GitHub issues/PRs
