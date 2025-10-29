#!/bin/bash
# Install git hooks for automatic code formatting
# Run this script after cloning the repository: ./scripts/install-git-hooks.sh

set -e

# Get the root directory of the git repository
REPO_ROOT=$(git rev-parse --show-toplevel)
HOOKS_DIR="$REPO_ROOT/.git/hooks"

echo "Installing git hooks..."

# Create pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash
# Pre-commit hook for automatic code formatting
# This hook formats Go and Dart code before committing

set -e

echo "Running pre-commit formatting..."

# Get the root directory of the git repository
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT"

# Format Go files
echo "Formatting Go files..."
if command -v gofmt >/dev/null 2>&1; then
    # Get list of staged Go files
    GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

    if [ -n "$GO_FILES" ]; then
        echo "  Found Go files to format:"
        echo "$GO_FILES" | sed 's/^/    /'

        # Format each file
        echo "$GO_FILES" | xargs gofmt -w

        # Add formatted files back to staging
        echo "$GO_FILES" | xargs git add

        echo "  ✓ Go files formatted"
    else
        echo "  No Go files to format"
    fi
else
    echo "  ⚠ gofmt not found, skipping Go formatting"
fi

# Format Dart files
echo "Formatting Dart files..."
if command -v dart >/dev/null 2>&1; then
    # Get list of staged Dart files
    DART_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.dart$' || true)

    if [ -n "$DART_FILES" ]; then
        echo "  Found Dart files to format:"
        echo "$DART_FILES" | sed 's/^/    /'

        # Format each file
        echo "$DART_FILES" | xargs dart format

        # Add formatted files back to staging
        echo "$DART_FILES" | xargs git add

        echo "  ✓ Dart files formatted"
    else
        echo "  No Dart files to format"
    fi
else
    echo "  ⚠ dart not found, skipping Dart formatting"
fi

echo "✓ Pre-commit formatting complete!"
exit 0
EOF

# Make the hook executable
chmod +x "$HOOKS_DIR/pre-commit"

echo "✓ Git hooks installed successfully!"
echo ""
echo "The following hooks are now active:"
echo "  - pre-commit: Automatically formats Go and Dart files"
echo ""
echo "To disable automatic formatting, remove: .git/hooks/pre-commit"
