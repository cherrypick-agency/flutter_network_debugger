---
layout: home
hero:
  name: "Network Debugger"
  text: See Every Byte Your App Sends
  tagline: HTTP, WebSocket & Socket.IO inspector for Dart and Flutter — with waterfall timeline, scripting, breakpoints, and more
  actions:
    - theme: brand
      text: Quick Start
      link: /guide/
    - theme: alt
      text: API Reference
      link: /api/
    - theme: alt
      text: GitHub
      link: https://github.com/cherrypick-agency/flutter_network_debugger
features:
  - icon: 🔍
    title: HTTP Inspection
    details: See headers, body, timing and status for every request. Filter by method, status code, or URL pattern.
  - icon: 🌊
    title: Waterfall Timeline
    details: Visual breakdown of DNS, TLS, waiting and transfer time for each request — spot bottlenecks instantly.
  - icon: 🔌
    title: WebSocket & Socket.IO
    details: Full message history for persistent connections. Inspect frames, events, and binary payloads in real time.
  - icon: ⏱️
    title: Throttling & Latency
    details: Simulate slow networks, high latency, or dropped packets to test how your app handles poor connectivity.
  - icon: 🛑
    title: Breakpoints
    details: Pause any request mid-flight, inspect or modify it, then let it continue. Debug tricky API issues on the fly.
  - icon: 📝
    title: Scripting
    details: Write rules in Dart, Go, Rust, JS, or Python to rewrite requests, mock responses, or inject headers automatically.
  - icon: 🖥️
    title: Cross-Platform
    details: Desktop app for macOS, Windows, Linux. CLI mode for CI pipelines and headless environments. Web UI included.
  - icon: 🔥
    title: Firebase RTDB
    details: Monitor Firebase Realtime Database reads, writes and listeners — see exactly what data flows through your app.
---

## Packages

| Package | What it does |
|---|---|
| [`network_debugger`](https://pub.dev/packages/network_debugger) | CLI launcher — starts the proxy and opens the UI |
| [`dio_debugger`](https://pub.dev/packages/dio_debugger) | Attaches the proxy to a Dio HTTP client |
| [`http_debugger`](https://pub.dev/packages/http_debugger) | Global HTTP interception via `HttpOverrides` |
| [`web_socket_debugger`](https://pub.dev/packages/web_socket_debugger) | Intercepts `dart:io` WebSocket connections |
| [`web_socket_channel_debugger`](https://pub.dev/packages/web_socket_channel_debugger) | Intercepts `package:web_socket_channel` |
| [`socket_io_debugger`](https://pub.dev/packages/socket_io_debugger) | Captures Socket.IO events and payloads |
| [`firebase_database_debugger`](https://pub.dev/packages/firebase_database_debugger) | Tracks Firebase Realtime Database operations |
| [`hex_viewer`](https://pub.dev/packages/hex_viewer) | Flutter widget for viewing binary data in HEX |

## Quick Setup

::: code-group
```bash [Install]
dart pub global activate network_debugger
network_debugger
```
```dart [Dio]
import 'package:dio_debugger/dio_debugger.dart';

final dio = Dio()..interceptors.add(DioDebugger());
```
```dart [HttpOverrides]
import 'package:http_debugger/http_debugger.dart';

HttpOverrides.global = DebuggerHttpOverrides();
```
:::

## Architecture

```mermaid
flowchart LR
    App["Your App<br/>(Flutter / Dart)"] -->|HTTP / WS| Proxy["Go Proxy<br/>(localhost:9091)"]
    Proxy -->|forwards| API["Target API<br/>(remote server)"]
    API -->|response| Proxy
    Proxy -->|response| App
    Proxy -->|streams traffic| UI["Desktop / Web UI<br/>(localhost:9092)"]
```

Your app sends traffic through a local Go proxy. The proxy records everything and forwards it to the real server. The UI connects to the proxy and displays traffic in real time.
