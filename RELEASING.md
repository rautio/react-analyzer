# Release Process

This document describes how to create a new release of react-analyzer.

## Prerequisites

- Push access to the repository
- All changes merged to `main` branch
- All tests passing
- Changelog is up to date (or will be auto-generated)

## Creating a Release

The release process is fully automated using GoReleaser and GitHub Actions.

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

1. GitHub Actions workflow (`.github/workflows/release.yml`) is triggered
2. Tests are run to ensure quality
3. GoReleaser builds binaries for:
   - **macOS**: Intel (amd64) and Apple Silicon (arm64)
   - **Linux**: amd64 and arm64
   - **Windows**: amd64
4. Archives are created (`.tar.gz` for Unix, `.zip` for Windows)
5. Checksums are generated
6. GitHub Release is created with:
   - Release notes (auto-generated from commits)
   - All binaries attached
   - Installation instructions

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

# Full release build (all platforms, snapshot mode)
goreleaser build --snapshot --clean
```

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

**Solution:** The GitHub Actions runner has the necessary build tools. If testing locally, ensure you have a C compiler installed:
- **macOS**: `xcode-select --install`
- **Linux**: `apt-get install build-essential`
- **Windows**: Install MinGW or TDM-GCC

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
