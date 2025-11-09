# Simple Python script that adds custom headers to HTTP requests
# This demonstrates basic request modification using Python

# The 'request' object is automatically available in the scope
# It has the following structure:
# - request.method: str (e.g., "GET", "POST")
# - request.url: str
# - request.headers: dict[str, list[str]]
# - request.body: bytes

# Add custom headers
if "X-Python-Processed" not in request.headers:
    request.headers["X-Python-Processed"] = ["true"]

request.headers["X-Script-Language"] = ["Python"]
request.headers["X-Runtime"] = ["RustPython-WASM"]

# Log the request (optional)
print(f"Processing {request.method} {request.url}")

# The modified 'request' object is automatically returned
