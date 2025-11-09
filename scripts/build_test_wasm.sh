#!/bin/bash
set -euo pipefail

# Build WASM fixtures for E2E tests
# Run this script after modifying script examples or test fixtures

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "📦 Building WASM fixtures for E2E tests..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Rust is installed
if ! command -v cargo &> /dev/null; then
    echo "❌ Rust not installed. Install from https://rustup.rs/"
    exit 1
fi

# Check if wasm32 target is installed
if ! rustup target list --installed | grep -q "wasm32-unknown-unknown"; then
    echo "${YELLOW}⚠️  wasm32-unknown-unknown target not installed, installing...${NC}"
    rustup target add wasm32-unknown-unknown
fi

cd "$PROJECT_ROOT"

# 1. Build add_header.wasm from examples
echo "🦀 Building add_header.wasm..."
cd examples/scripts/rust
cargo build --target wasm32-unknown-unknown --release --quiet
cp target/wasm32-unknown-unknown/release/add_header.wasm \
   "$PROJECT_ROOT/internal/e2e/testdata/scripts/wasm/"
echo "${GREEN}✓${NC} add_header.wasm ($(du -h target/wasm32-unknown-unknown/release/add_header.wasm | cut -f1))"

# 2. Build noop.wasm
echo "🦀 Building noop.wasm..."
cd "$PROJECT_ROOT/internal/e2e/testdata/scripts/wasm-src/noop"
cargo build --target wasm32-unknown-unknown --release --quiet
cp target/wasm32-unknown-unknown/release/noop.wasm ../../wasm/
echo "${GREEN}✓${NC} noop.wasm ($(du -h target/wasm32-unknown-unknown/release/noop.wasm | cut -f1))"

# 3. Create invalid.wasm (corrupt header for negative tests)
echo "🔨 Creating invalid.wasm..."
echo -n "INVALID_WASM_HEADER" > "$PROJECT_ROOT/internal/e2e/testdata/scripts/wasm/invalid.wasm"
echo "${GREEN}✓${NC} invalid.wasm (corrupt for negative tests)"

# Summary
echo ""
echo "✅ All WASM fixtures built successfully:"
ls -lh "$PROJECT_ROOT/internal/e2e/testdata/scripts/wasm/"

echo ""
echo "📝 Note: Dart fixtures are source files, no compilation needed"
ls -lh "$PROJECT_ROOT/internal/e2e/testdata/scripts/dart/"
