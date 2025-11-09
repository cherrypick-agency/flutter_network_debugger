# Python Scripts for WASM Compilation

## Overview

Python scripts are compiled to WASM using a **hybrid approach**:

```
Python Code → Rust Wrapper (with RustPython) → WASM
```

This allows you to write HTTP request/response handlers in Python while getting the performance and security benefits of WASM.

## Architecture

### How It Works

1. **Your Python code** - Write normal Python code that modifies the `request` object
2. **Rust wrapper** - Automatically generated wrapper that embeds your Python code
3. **RustPython runtime** - Python interpreter compiled into the WASM module
4. **Extism PDK** - WASM plugin interface for HTTP handling

### Available API

Your Python script has access to a `request` object with the following structure:

```python
class HTTPRequest:
    method: str  # "GET", "POST", etc.
    url: str     # Full URL
    headers: dict[str, list[str]]  # Headers (multi-value)
    body: bytes  # Request body as bytes
```

**Important:** Modify the `request` object in-place. It will be automatically returned.

## Examples

### Simple Example: Add Headers

```python
# Add custom headers
request.headers["X-Python-Script"] = ["active"]
request.headers["X-Processed-By"] = ["RustPython-WASM"]

print(f"Processing {request.method} {request.url}")
```

### Advanced Example: JSON Body Manipulation

```python
import json
from urllib.parse import urlparse

# Parse URL
parsed_url = urlparse(request.url)

# Add API version header for API endpoints
if "/api/" in parsed_url.path:
    request.headers["X-API-Version"] = ["v1"]

# Modify JSON body
if request.method == "POST" and request.body:
    try:
        data = json.loads(request.body.decode('utf-8'))
        data["_metadata"] = {
            "processed_by": "python",
            "timestamp": time.time()
        }
        request.body = json.dumps(data).encode('utf-8')
        request.headers["Content-Length"] = [str(len(request.body))]
    except Exception as e:
        print(f"Error: {e}")
```

## Available Standard Library

RustPython includes a subset of Python's standard library:

✅ **Available:**
- `json` - JSON encoding/decoding
- `urllib.parse` - URL parsing
- `time` - Time functions
- `re` - Regular expressions (basic)
- `base64` - Base64 encoding
- `hashlib` - Hashing functions
- `collections` - Data structures

❌ **Not Available:**
- File I/O (`open`, `file`)
- Network I/O (`socket`, `requests`)
- OS operations (`os`, `sys`)
- Threading (`threading`, `multiprocessing`)

## Limitations

1. **Sandboxed Environment:** No file system or network access
2. **No External Packages:** Can't use `pip install` packages (only stdlib)
3. **Memory Limits:** WASM modules have memory constraints
4. **Compilation Time:** Slower than native Python (compiles to WASM)

## Usage

### 1. Create Script via API

```bash
curl -X POST http://localhost:9092/_api/v1/scripts \
  -H "Content-Type: application/json" \
  -d @simple_header_example.json
```

### 2. Compile to WASM

```bash
curl -X POST http://localhost:9092/_api/v1/scripts/{id}/compile \
  -d '{"optimize": true}'
```

**Note:** Python compilation requires Rust toolchain because it generates a Rust wrapper.

### 3. Enable and Test

```bash
curl -X PATCH http://localhost:9092/_api/v1/scripts/{id}/toggle \
  -d '{"enabled": true}'

# Test through proxy
curl -x http://localhost:9091 http://httpbin.org/get
```

## Performance Considerations

**Compilation Time:**
- First compilation: ~30-60 seconds (downloads RustPython dependencies)
- Subsequent compilations: ~10-20 seconds (uses Cargo cache)

**Runtime Performance:**
- Python execution via RustPython is slower than native Rust
- Good for: Logic-heavy processing, JSON manipulation
- Not ideal for: High-frequency simple operations

## Best Practices

1. **Keep Scripts Small:** Smaller scripts compile faster
2. **Error Handling:** Always use try-except blocks
3. **Type Safety:** Python dict keys should be strings
4. **Logging:** Use `print()` for debugging (visible in logs)
5. **Idempotency:** Ensure your script can run multiple times safely

## Debugging

### Enable Debug Logging

```python
# Add debug headers
request.headers["X-Debug"] = ["true"]

# Log everything
print(f"Method: {request.method}")
print(f"URL: {request.url}")
print(f"Headers: {request.headers}")
if request.body:
    print(f"Body: {request.body[:100]}")  # First 100 bytes
```

### Common Issues

**Issue:** `KeyError` when accessing headers

**Solution:** Check if key exists first:
```python
if "Authorization" in request.headers:
    auth = request.headers["Authorization"][0]
```

**Issue:** JSON decode error

**Solution:** Always handle exceptions:
```python
try:
    data = json.loads(request.body.decode('utf-8'))
except Exception as e:
    print(f"JSON parse error: {e}")
    # Don't modify body on error
```

## Files in This Directory

- `simple_header.py` - Basic example: add headers
- `advanced_processor.py` - Advanced: URL parsing, JSON manipulation
- `simple_header_example.json` - API request example

## Requirements

To compile Python scripts, you need:

✅ **Rust toolchain** (rustup + cargo)
✅ **wasm32-unknown-unknown target** (`rustup target add wasm32-unknown-unknown`)

Python itself is **not required** - RustPython is embedded!

## Technical Details

### Compilation Pipeline

```
┌─────────────┐
│ Python Code │
└──────┬──────┘
       │
       ▼
┌──────────────────────────┐
│ Generate Rust Wrapper    │
│ with embedded Python     │
└──────┬───────────────────┘
       │
       ▼
┌──────────────────────────┐
│ Add RustPython to        │
│ Cargo.toml dependencies  │
└──────┬───────────────────┘
       │
       ▼
┌──────────────────────────┐
│ cargo build              │
│ --target wasm32-unknown  │
└──────┬───────────────────┘
       │
       ▼
┌──────────────────────────┐
│ WASM Binary Ready!       │
└──────────────────────────┘
```

### Why RustPython?

- **Sandboxed:** Safe execution in WASM
- **Portable:** Works everywhere WASM works
- **Fast Startup:** No interpreter initialization overhead
- **Small Size:** Reasonable WASM binary size (~2-5 MB)

## See Also

- [Main README](../README.md) - All language examples
- [Quick Start](../QUICK_START.md) - Getting started guide
- [Rust Examples](../rust/) - Native WASM examples
