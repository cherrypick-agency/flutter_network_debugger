# Contributing to Network Debugger

## Development Setup

### Prerequisites

- Go 1.22+
- Dart/Flutter SDK 3.x
- Git

### Getting Started

1. Clone the repository:
```bash
git clone https://github.com/cherrypick-agency/flutter_network_debugger.git
cd flutter_network_debugger
```

2. Install dependencies:
```bash
go mod download
cd frontend && flutter pub get
```

3. **Install git hooks for automatic code formatting:**
```bash
make install-hooks
```
or manually:
```bash
./scripts/install-git-hooks.sh
```

This will set up a pre-commit hook that automatically formats your code before each commit.

## Code Formatting

We use automatic code formatting to ensure consistency across the codebase.

### Automatic Formatting (Recommended)

The pre-commit hook will automatically format your code when you commit. Just run:
```bash
git add .
git commit -m "your message"
```

The hook will:
- Format all staged Go files with `gofmt`
- Format all staged Dart files with `dart format`
- Re-add formatted files to the commit

### Manual Formatting

If you need to format code manually:

```bash
# Format all files (Go + Dart)
make fmt

# Format only Go files
make fmt-go

# Format only Dart files
make fmt-dart
```

### CI/CD Checks

All pull requests and commits must pass:
- Go formatting check (`go fmt`)
- Dart formatting check (`dart format`)
- All tests (Go and Dart)
- Code coverage

**Important:** The CI will reject commits with formatting issues, so please ensure you have the pre-commit hook installed.

## Testing

```bash
# Run all Go tests
make test

# Run integration tests
make itest

# Run with coverage
go test -coverprofile=coverage.out ./...
```

## Building

```bash
# Build for development
make build

# Build with web UI
make build-app

# Cross-compile for all platforms
make build-cross
```

## Development Workflow

1. Create a feature branch:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes

3. The pre-commit hook will automatically format your code on commit

4. Run tests:
```bash
make test
```

5. Push your changes:
```bash
git push origin feature/your-feature-name
```

6. Create a pull request

## Troubleshooting

### Pre-commit hook not working?

Ensure the hook is executable:
```bash
chmod +x .git/hooks/pre-commit
```

Or reinstall:
```bash
make install-hooks
```

### Want to skip the hook temporarily?

Use `--no-verify`:
```bash
git commit --no-verify -m "message"
```

**Note:** Only use this when absolutely necessary, as it will bypass formatting checks.

### Formatting tools not found?

Ensure you have the required tools installed:
- Go: `gofmt` comes with Go installation
- Dart: Install Flutter/Dart SDK from https://flutter.dev

## Questions?

Feel free to open an issue or reach out to the maintainers!
