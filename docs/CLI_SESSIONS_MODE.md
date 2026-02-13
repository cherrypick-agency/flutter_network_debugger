# CLI sessions mode (colored console output)

Run the server with CLI output enabled (no browser auto-open):

```bash
network_debugger --cli --cli-preset basic
# or fine‑tune fields:
network_debugger --cli \
  --cli-fields line,sizes,timings,req-headers,resp-headers \
  --cli-color auto \
  --cli-filter "/api/" \
  --cli-body-bytes 50000
```

Presets: `minimal | basic | advanced | full` (fields listed below).
`--cli-fields` overrides preset entirely.

Color modes: `auto | always | never`.

## Flags

- `--cli`: enable CLI mode (disables auto-open browser)
- `--open-browser`: open browser on start (opt-in for `network-debugger`,
  enabled by default for `network-debugger-web`)
- `--cli-preset`: one of `minimal|basic|advanced|full`
- `--cli-fields`: comma-separated list of sections to show (overrides preset)
- `--cli-body-bytes`: body preview limit (bytes); `0` = use `PREVIEW_MAX_BYTES`
- `--cli-color`: `auto|always|never`
- `--cli-filter`: substring filter (matches URL/method/status)

## Fields (sections)

- `line`: single-line summary (time, METHOD, URL, STATUS, totalMs, sizes)
- `sizes`: request/response byte sizes (also included in `line` summary)
- `timings`: HTTP timings (DNS/Connect/TLS/TTFB/Total) when available
- `req-headers`, `resp-headers`: selected headers with masking of sensitive
  values
- `req-body`, `resp-body`: pretty JSON or raw preview, trimmed by
  `--cli-body-bytes`
- `tls`: TLS handshake timing + HTTP protocol hints when available
- `cookies`: Set-Cookie flags summary (counts of Secure/HttpOnly/SameSite)
- `ids`: internal IDs (session/tx) for cross-referencing

## Presets mapping

- minimal: `line`
- basic: `line,sizes`
- advanced: `line,sizes,timings,req-headers,resp-headers`
- full: `advanced + req-body,resp-body,tls,cookies,ids`

## Notes

- HTTP request/response bodies shown are previews; they may be truncated.
- Sensitive headers are masked by default; enable raw exposure via server
  config if needed.
- Colors are enabled automatically for TTY; force with `--cli-color always`.

## Examples

```bash
# Show only single-line summaries, always with colors
network_debugger --cli --cli-preset minimal --cli-color always

# Full details but limit body previews to 16KB
network_debugger --cli --cli-preset full --cli-body-bytes 16384

# Custom selection with filtering for API routes
network_debugger --cli \
  --cli-fields line,timings,req-headers,resp-headers \
  --cli-filter "/v1/"
```

## Where the UI opens

By default, the server listens on:
- UI: `http://localhost:9092/`
- Proxy base (HTTP/WebSocket forward): `http://localhost:9091`

