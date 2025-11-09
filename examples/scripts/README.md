# Script Examples

This directory contains example scripts for the Network Debugger scripting feature. These examples demonstrate common use cases and best practices for writing scripts.

## Available Examples

### 1. [rust-add-header](./rust-add-header/) 🦀
**Language:** Rust | **Trigger:** Before Request | **Difficulty:** Beginner

Adds a custom HTTP header to requests. Perfect starting point for learning Rust scripts.

**Use cases:** Add authentication tokens, tracking IDs, modify user-agent strings

### 2. [js-transform-response](./js-transform-response/) 🟨
**Language:** JavaScript/TypeScript | **Trigger:** After Response | **Difficulty:** Intermediate

Transforms JSON response bodies by adding custom metadata fields.

**Use cases:** Inject custom fields, filter sensitive data, add analytics, transform data formats

### 3. [go-rate-limit](./go-rate-limit/) 🐹
**Language:** Go | **Trigger:** Before Request | **Difficulty:** Advanced

Implements simple rate limiting based on client IP addresses.

**Use cases:** Protect APIs from abuse, fair usage policies, DDoS protection, quota enforcement

---

## How to Use These Examples

### Option 1: Upload via UI

1. Open the Network Debugger UI
2. Navigate to Scripts page → Click "New Script"
3. Fill in name, select language
4. Click "Upload Project" → Select ZIPped example folder
5. Click "Compile" → Enable the script

### Option 2: Copy & Paste

1. Create a new script in the UI
2. Copy source code from example files
3. Copy dependency files (Cargo.toml/package.json/go.mod)
4. Click "Compile"

### Option 3: CLI Tool (Advanced)

```bash
# Upload a project
go-proxy-sync upload --script-id=<id> --dir=./rust-add-header

# Download a project
go-proxy-sync download --script-id=<id> --output=./my-script

# Watch for changes
go-proxy-sync watch --script-id=<id> --dir=./my-script
```

---

## Creating Your Own Scripts

### Script Structure

Every script must export a `transform` function that accepts ScriptContext and returns ScriptResult.

**Input (ScriptContext):**
```json
{
  "request": { "method": "GET", "url": "...", "headers": {}, "body": [] },
  "response": { "status": 200, "headers": {}, "body": [] },
  "session": { "id": "...", "client_addr": "..." }
}
```

**Output (ScriptResult):**
```json
{
  "modified": true,
  "modifiedRequest": {...},
  "modifiedResponse": {...},
  "logs": ["log 1", "log 2"]
}
```

### Best Practices

1. **Always log** - Use logs array for debugging
2. **Handle errors gracefully** - Don't crash on invalid input
3. **Check content types** - Only parse JSON if appropriate
4. **Be efficient** - Scripts run on every matching request
5. **Test thoroughly** - Use Test tab before enabling

---

## Testing Your Scripts

Use the built-in Test tab:
1. Select your script → Go to "Test" tab
2. Enter sample request/response data
3. Click "Run Test" → Review output and logs

---

## Need Help?

- Check the scripting documentation
- Look at existing examples for patterns
- Use the debugger's built-in test feature

## Contributing

PRs welcome! Common ideas: OAuth 2.0 auth, circuit breaker, endpoint mocking, request signing
