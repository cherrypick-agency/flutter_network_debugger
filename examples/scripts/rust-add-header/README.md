# Add Custom Header (Rust)

This example demonstrates how to add a custom HTTP header to requests using Rust and the Extism PDK.

## What it does

- Intercepts HTTP requests before they are sent upstream
- Adds a custom `X-Custom-Header` with value `My-Value`
- Useful for API authentication, tracking, or custom metadata

## Files

- `src/lib.rs` - Main script logic
- `Cargo.toml` - Rust dependencies

## Usage

1. Create a new script in the debugger UI
2. Upload this project folder as a ZIP
3. Compile the script
4. Enable the script
5. Configure match rules (e.g., match all requests to `api.example.com`)

## Testing

Use the Test tab with a sample request:

```json
{
  "method": "GET",
  "url": "https://api.example.com/users",
  "headers": {},
  "body": ""
}
```

You should see the `X-Custom-Header` added to the modified request.

## Extending

You can modify this script to:
- Add authentication tokens (Bearer, API keys)
- Add correlation IDs for distributed tracing
- Add custom user-agent strings
- Conditionally add headers based on request properties
