---
layout: home
hero:
  name: "Network Debugger"
  text: Dart/Flutter package documentation
  tagline: Open-source toolkit for debugging HTTP, WebSocket and Socket.IO traffic — with UI, CLI and a Go proxy backend
  actions:
    - theme: brand
      text: Quick Start
      link: /guide/
    - theme: alt
      text: API Reference
      link: /api/
    - theme: alt
      text: GitHub Repository
      link: https://github.com/cherrypick-agency/flutter_network_debugger

features:
  - title: What is Network Debugger
    details: A full-featured network inspector for local development and test environments. Works offline, supports Web, Desktop (macOS/Windows/Linux) and CLI.
  - title: Key capabilities
    details: "HTTP/WebSocket/Socket.IO inspection, waterfall timeline, scripting (Go, Rust, JS, Python…), request mapping, throttling, breakpoints, compose/request builder, performance insights and more."
  - title: What's in the docs
    details: Guides and auto-generated API reference for every public package in the dart_packages monorepo, with full-text search.
---

## Packages

| Package | Description |
| --- | --- |
| [`network_debugger`](https://pub.dev/packages/network_debugger) | CLI launcher — starts the proxy and opens the UI |
| [`dio_debugger`](https://pub.dev/packages/dio_debugger) | Attaches the proxy to a Dio HTTP client |
| [`http_debugger`](https://pub.dev/packages/http_debugger) | Global HTTP interception via `HttpOverrides` |
| [`web_socket_debugger`](https://pub.dev/packages/web_socket_debugger) | Integration with `dart:io` WebSocket |
| [`web_socket_channel_debugger`](https://pub.dev/packages/web_socket_channel_debugger) | Integration with `package:web_socket_channel` |
| [`socket_io_debugger`](https://pub.dev/packages/socket_io_debugger) | Integration with the Socket.IO client |
| [`firebase_database_debugger`](https://pub.dev/packages/firebase_database_debugger) | Sends Firebase Realtime Database operations to the ingest API |
| [`hex_viewer`](https://pub.dev/packages/hex_viewer) | Flutter widget for viewing binary payloads in HEX |

## How to use this documentation

1. Start with the **Guide** section for a quick integration walkthrough.
2. Jump to **API Reference** for detailed class and method docs of a specific package.
3. Use the search bar in the top-right corner to find any symbol instantly.
