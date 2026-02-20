---
layout: home
hero:
  name: "Network Debugger"
  text: See Every Byte Your App Sends
  tagline: HTTP, WebSocket & Socket.IO inspector for Dart and Flutter — with modern UI, waterfall timeline, breakpoints, and more
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
  - icon: <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 351" width="48" height="48"><path d="M1.253 280.732l1.605-3.131 99.353-188.518-44.15-83.475C54.392-1.283 45.074.474 43.87 8.188z" fill="#ffa000"/><path d="m134.417 148.974 32.039-32.812-32.039-61.518c-3.042-5.791-10.433-6.398-13.443-.59l-19.14 36.322-.2.394z" fill="#f57c00"/><path d="m0 282.998 2.123-2.416L134.417 148.974l-32.205-57.891z" fill="#ffca28"/><path d="M139.121 347.551l116.165-64.404-32.674-201.29a7.27 7.27 0 0 0-12.472-3.727L0 282.998l115.608 64.553a24.126 24.126 0 0 0 23.513 0" fill="#ffa000"/><path d="M254.354 282.16 221.402 79.218c-1.29-7.972-11.204-10.66-15.673-4.253L0 282.998l115.608 64.553a24.126 24.126 0 0 0 23.513 0l115.233-64.404z" fill="#f57c00"/><path d="M139.121 345.64a24.126 24.126 0 0 1-23.513 0L1.253 282.732l-.88.266 115.608 64.553a24.126 24.126 0 0 0 23.513 0l115.233-64.404-.926-5.753z" fill="#f57c00" opacity=".2"/></svg>
    title: Firebase RTDB
    details: Monitor Firebase Realtime Database reads, writes and listeners — see exactly what data flows through your app.
---

## Quick Setup

<script setup>
import { ref } from 'vue'
const tab = ref('install')
</script>

<div class="setup-tabs">
<div class="setup-tabs-nav">
<button v-for="t in [
  {id:'install',label:'Install'},
  {id:'dio',label:'Dio'},
  {id:'http',label:'HttpOverrides'},
  {id:'ws',label:'WebSocket'},
  {id:'wsc',label:'WebSocketChannel'},
  {id:'sio',label:'Socket.IO'},
  {id:'firebase',label:'Firebase RTDB'},
]" :key="t.id" :class="['setup-tab-btn', {active: tab === t.id}]" @click="tab = t.id">{{ t.label }}</button>
</div>
</div>

<div v-show="tab === 'install'">

```bash
dart pub global activate network_debugger
network_debugger
```

<a class="setup-guide-link" href="/guide/network_debugger_workspace/quick-start">📚 Quick Start Guide →</a>

</div>
<div v-show="tab === 'dio'">

```dart
import 'package:dio_debugger/dio_debugger.dart';

final dio = Dio()..interceptors.add(DioDebugger());
```

<a class="setup-guide-link" href="/guide/network_debugger_workspace/packages/dio_debugger">📚 Dio Debugger documentation →</a>

</div>
<div v-show="tab === 'http'">

```dart
import 'package:http_debugger/http_debugger.dart';

HttpOverrides.global = DebuggerHttpOverrides();
```

<a class="setup-guide-link" href="/guide/network_debugger_workspace/packages/http_debugger">📚 HTTP Debugger documentation →</a>

</div>
<div v-show="tab === 'ws'">

```dart
import 'package:web_socket_debugger/web_socket_debugger.dart';

final cfg = WebSocketDebugger.attach(baseUrl: 'wss://echo.websocket.events');
final socket = await WebSocketDebugger.connect(config: cfg);
```

<a class="setup-guide-link" href="/guide/network_debugger_workspace/packages/web_socket_debugger">📚 WebSocket Debugger documentation →</a>

</div>
<div v-show="tab === 'wsc'">

```dart
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

final cfg = WebSocketChannelDebugger.attach(baseUrl: 'wss://echo.websocket.events');
final channel = WebSocketChannelDebugger.connect(config: cfg);
```

<a class="setup-guide-link" href="/guide/network_debugger_workspace/packages/web_socket_channel_debugger">📚 WebSocketChannel Debugger documentation →</a>

</div>
<div v-show="tab === 'sio'">

```dart
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';

final cfg = SocketIoDebugger.attach(baseUrl: 'https://example.com');
final socket = io.io(cfg.effectiveBaseUrl,
  io.OptionBuilder()
    .setPath(cfg.effectivePath)
    .setQuery(cfg.query)
    .build());
```

<a class="setup-guide-link" href="/guide/network_debugger_workspace/packages/socket_io_debugger">📚 Socket.IO Debugger documentation →</a>

</div>
<div v-show="tab === 'firebase'">

```dart
import 'package:firebase_database_debugger/firebase_database_debugger.dart';

final debugger = FirebaseDatabaseDebugger();
final ref = debugger.ref(FirebaseDatabase.instance.ref('users'));
```

<a class="setup-guide-link" href="/guide/firebase-database">📚 Firebase RTDB Debugger documentation →</a>

</div>

<style>
.setup-tabs { margin-top: 16px; }
.setup-tabs-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
  border-bottom: 1px solid var(--vp-c-divider);
  padding: 0 4px;
}
.setup-tab-btn {
  padding: 8px 16px;
  border: none;
  background: none;
  color: var(--vp-c-text-2);
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  border-bottom: 2px solid transparent;
  transition: color 0.2s, border-color 0.2s;
}
.setup-tab-btn:hover { color: var(--vp-c-text-1); }
.setup-tab-btn.active {
  color: var(--vp-c-brand-1);
  border-bottom-color: var(--vp-c-brand-1);
}
.setup-guide-link {
  display: inline-block;
  margin-top: -8px;
  margin-bottom: 8px;
  padding: 6px 14px;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  color: var(--vp-c-brand-1);
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  transition: background 0.2s, border-color 0.2s;
}
.setup-guide-link:hover {
  background: var(--vp-c-bg-elv);
  border-color: var(--vp-c-brand-1);
}
</style>

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
