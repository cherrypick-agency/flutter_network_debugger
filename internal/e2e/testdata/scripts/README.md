# E2E Test Fixtures for Scripting API

Test fixtures (WASM modules and Dart scripts) for testing the Scripting API feature.

## Directory Structure

```
testdata/scripts/
├── wasm/                   # Compiled WASM modules (fixtures)
│   ├── add_header.wasm    # Rust: adds X-Script-Processed header
│   ├── noop.wasm          # Minimal valid WASM (passthrough)
│   └── invalid.wasm       # Corrupt WASM for negative tests
├── wasm-src/              # Source code for WASM fixtures
│   └── noop/              # Minimal Rust project
└── dart/                  # Dart scripts (source, no compilation)
    └── simple_logger.dart # Logs requests, adds X-Dart-Processed header
```

## Rebuilding Fixtures

If you modify script examples or test fixtures, rebuild them:

```bash
# From project root
./scripts/build_test_wasm.sh
```

### Requirements

- **Rust** + `wasm32-unknown-unknown` target
  ```bash
  rustup target add wasm32-unknown-unknown
  ```

- **Dart** (optional, for Dart executor tests)
  ```bash
  # macOS
  brew install dart

  # Linux
  sudo apt install dart
  ```

## Manual Build

### WASM Fixtures

```bash
# add_header.wasm
cd examples/scripts/rust
cargo build --target wasm32-unknown-unknown --release
cp target/wasm32-unknown-unknown/release/add_header.wasm \
   internal/e2e/testdata/scripts/wasm/

# noop.wasm
cd internal/e2e/testdata/scripts/wasm-src/noop
cargo build --target wasm32-unknown-unknown --release
cp target/wasm32-unknown-unknown/release/noop.wasm ../../wasm/
```

### Dart Scripts

Dart scripts are executed as source files, no compilation needed.

## Fixture Descriptions

### `add_header.wasm` (Rust)
- **Purpose**: Tests request modification
- **Behavior**: Adds headers `X-Script-Processed: Rust` and `X-Test-Header: E2E-Test`
- **Use case**: Verify script execution, header injection
- **Size**: ~204KB

### `noop.wasm` (Rust)
- **Purpose**: Tests minimal valid WASM
- **Behavior**: Returns input unchanged (passthrough)
- **Use case**: WASM loading/validation, baseline performance
- **Size**: ~95KB

### `invalid.wasm`
- **Purpose**: Tests WASM validation rejection
- **Behavior**: Corrupt header, fails to load
- **Use case**: Negative testing (should return 400 Bad Request)
- **Size**: 19B

### `simple_logger.dart` (Dart)
- **Purpose**: Tests Dart subprocess executor
- **Behavior**:
  - Logs requests to stderr
  - Adds headers `X-Dart-Processed: true` and `X-Dart-Test: E2E`
- **Use case**: Verify Dart executor, JSON-RPC communication
- **Note**: Requires Dart SDK installed

## Adding New Fixtures

1. Create source in `wasm-src/<name>/` or `dart/`
2. Build WASM (if applicable)
3. Copy to `wasm/` or `dart/`
4. Update this README
5. Update E2E tests to use new fixture

## CI/CD

In GitHub Actions, fixtures are cached based on hash of `examples/scripts/**`:

```yaml
- name: Cache WASM fixtures
  uses: actions/cache@v3
  with:
    path: internal/e2e/testdata/scripts/wasm
    key: wasm-fixtures-${{ hashFiles('examples/scripts/**', 'internal/e2e/testdata/scripts/wasm-src/**') }}
```

On cache miss, CI runs `./scripts/build_test_wasm.sh` to rebuild all fixtures.

## Troubleshooting

### "Rust not installed"
Install Rust toolchain: https://rustup.rs/

### "wasm32-unknown-unknown not found"
```bash
rustup target add wasm32-unknown-unknown
```

### "Dart not found" (in E2E tests)
Dart executor tests are automatically skipped if Dart SDK is not installed.
Install Dart: https://dart.dev/get-dart

### Fixture size too large
WASM fixtures are optimized in release mode (`--release`). Typical sizes:
- Minimal (noop): ~95KB
- With dependencies (add_header): ~200KB

If needed, use `wasm-opt`:
```bash
wasm-opt -Oz input.wasm -o output.wasm
```
