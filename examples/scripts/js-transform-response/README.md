# Transform JSON Response (JavaScript)

This example demonstrates how to modify HTTP response bodies using JavaScript and the Extism PDK.

## What it does

- Intercepts HTTP responses before they are sent to the client
- Parses JSON response bodies
- Adds a custom `_metadata` field with timestamp and processing info
- Useful for adding tracking data, modifying API responses, or injecting custom fields

## Files

- `index.ts` - Main script logic (TypeScript/JavaScript)
- `package.json` - npm dependencies

## Usage

1. Create a new script in the debugger UI
2. Select "JavaScript" or "TypeScript" as the language
3. Upload this project folder as a ZIP
4. Compile the script
5. Enable the script
6. Configure trigger type: "After Response" or "Both"
7. Configure match rules (e.g., match JSON API endpoints)

## Testing

Use the Test tab with a sample response:

```json
{
  "testRequest": {
    "method": "GET",
    "url": "https://api.example.com/users/123"
  },
  "testResponse": {
    "status": 200,
    "headers": {
      "Content-Type": ["application/json"]
    },
    "body": "{\"id\": 123, \"name\": \"John Doe\"}"
  }
}
```

You should see the `_metadata` field added to the JSON response.

## Extending

You can modify this script to:
- Filter sensitive fields from responses
- Transform data formats (snake_case ↔ camelCase)
- Add CORS headers
- Inject analytics or tracking pixels
- Redact PII (personally identifiable information)
