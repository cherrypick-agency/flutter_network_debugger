# Advanced Python script demonstrating request/response manipulation
# Shows conditional logic, URL parsing, and body modification

import json
from urllib.parse import urlparse, parse_qs

# Parse the URL
parsed_url = urlparse(request.url)
query_params = parse_qs(parsed_url.query)

# Add timing header
request.headers["X-Request-Time"] = [str(time.time())]

# Conditional processing based on URL path
if "/api/" in parsed_url.path:
    # API requests get special headers
    request.headers["X-API-Request"] = ["true"]
    request.headers["X-API-Version"] = ["v1"]

    # Add authentication header if not present
    if "Authorization" not in request.headers:
        request.headers["X-Auth-Warning"] = ["No authentication provided"]

# Process POST/PUT requests with JSON body
if request.method in ["POST", "PUT", "PATCH"]:
    try:
        # Try to parse JSON body
        body_str = request.body.decode('utf-8')
        if body_str:
            data = json.loads(body_str)

            # Add metadata to JSON
            if isinstance(data, dict):
                data["_metadata"] = {
                    "processed_by": "python-wasm",
                    "original_url": request.url
                }

                # Update body
                request.body = json.dumps(data).encode('utf-8')
                request.headers["Content-Length"] = [str(len(request.body))]
    except Exception as e:
        # Add error header if JSON parsing fails
        request.headers["X-Body-Parse-Error"] = [str(e)]

# URL parameter based routing
if "debug" in query_params:
    request.headers["X-Debug-Mode"] = ["enabled"]
    # Log all headers for debugging
    for key, values in request.headers.items():
        print(f"Header: {key} = {values}")

# Rate limiting hint (just adds header, actual limiting would be in separate middleware)
if parsed_url.hostname:
    request.headers["X-Target-Host"] = [parsed_url.hostname]

# Add processing info
request.headers["X-Python-Script"] = ["advanced_processor"]
request.headers["X-Processed-By"] = ["Python + RustPython"]

print(f"[Python] Processed {request.method} request to {request.url}")

# Return modified request
# Note: The 'request' object is automatically returned and converted back to Rust types
