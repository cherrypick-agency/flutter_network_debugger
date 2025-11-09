# Simple Rate Limiting (Go)

This example demonstrates how to implement basic rate limiting using Go and TinyGo.

## What it does

- Tracks request counts per client IP address
- Blocks requests that exceed the rate limit (10 requests per minute)
- Returns a 429 Too Many Requests error when limit is exceeded
- Useful for protecting APIs from abuse, implementing fair usage policies

## Files

- `main.go` - Main script logic
- `counter.go` - Simple in-memory counter implementation
- `go.mod` - Go module dependencies

## Limitations

**Note:** This is a simplified example for demonstration purposes. In production:
- Rate limit state is per-script-instance (not shared across multiple debugger instances)
- State is lost on script reload/restart
- For production use, consider Redis or similar for distributed rate limiting

## Usage

1. Create a new script in the debugger UI
2. Select "Go" as the language
3. Upload this project folder as a ZIP
4. Compile the script
5. Enable the script
6. Configure trigger type: "Before Request"
7. Configure match rules for the API you want to protect

## Testing

Use the Test tab and send multiple rapid requests with the same client IP. After 10 requests within a minute, you should see a 429 error response.

## Extending

You can modify this script to:
- Implement token bucket or sliding window algorithms
- Add different rate limits per endpoint or user
- Read rate limit from request headers (API keys)
- Store state in external storage (with Extism host functions)
- Add rate limit headers (X-RateLimit-Remaining, etc.)
