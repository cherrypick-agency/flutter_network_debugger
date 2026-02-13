# Network Debugger Web UI

Web interface for the Network Debugger proxy server. Built with Flutter Web, this UI provides a powerful and intuitive way to inspect HTTP and WebSocket traffic in real-time.

## Features

- **HTTP Traffic Inspection**: View requests/responses with detailed headers, body, timing (TTFB, total)
- **WebSocket Monitoring**: Track WebSocket connections, frames, events, ping/pong messages
- **Waterfall Timeline**: Visualize request/response timing and dependencies
- **Smart Filtering**: Filter by method, status, MIME type, duration, headers, domain/route
- **Search**: Fast full-text search with highlighting
- **Grouping**: Organize traffic by domain or route patterns
- **HAR Export**: Export captured traffic to HAR format
- **Sessions & Captures**: Manage multiple recording sessions
- **Response Preview**: View HTML, JSON (tree view), images, form data
- **Security Insights**: CORS/Cache hints, TLS summary, sensitive data masking
- **Hotkeys**: Keyboard shortcuts for fast navigation

## Project Structure

```
lib/
├── core/              # Core utilities, models, API client
├── features/          # Feature modules
│   ├── http_inspector/    # HTTP traffic viewer
│   ├── ws_inspector/      # WebSocket traffic viewer
│   ├── landing/           # Landing page & integrations
│   ├── filters/           # Filter controls
│   ├── settings/          # Settings panel
│   ├── inspector/         # Shared inspector logic
│   └── hotkeys/           # Keyboard shortcuts
├── services/          # Business logic services
├── theme/             # App theme & styles
├── widgets/           # Reusable UI components
└── main.dart          # App entry point
```

## Tech Stack

- **Flutter** 3.7.2+ (Web)
- **State Management**: flutter_riverpod
- **HTTP Client**: dio
- **Routing**: go_router
- **UI Components**: Custom widgets + Material Design

## Getting Started

### Prerequisites

- Flutter SDK 3.7.2 or later
- Dart SDK (bundled with Flutter)

### Development

1. **Install dependencies**:
   ```bash
   flutter pub get
   ```

2. **Run in development mode**:
   ```bash
   flutter run -d chrome
   ```
   or for hot-reload server:
   ```bash
   flutter run -d web-server --web-port 8080
   ```

3. **Backend connection**: The UI expects the backend running at `http://localhost:9092` by default (UI). Forward proxy is on `http://localhost:9091`. Start the Go backend first:
   ```bash
   # From project root
   go run ./cmd/network-debugger-web
   ```

### Capturing traffic from your Flutter app (important)

The Web UI only shows what reaches the proxy. If you don't route your app's
requests through the proxy, you'll see an empty sessions list.

You have two options:

1) **Flutter integration packages (recommended)** — route requests through the
proxy from inside the app:

- **Dio**: use `dio_debugger` (`DioDebugger.attach(dio)`)
- **package:http**: use `http_debugger`
- **WebSocket**:
  - `web_socket_debugger` (dart:io WebSocket)
  - `web_socket_channel_debugger` (package:web_socket_channel)
  - `socket_io_debugger` (Socket.IO client)

Example with `dio_debugger`:

```dart
import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';
import 'package:flutter/foundation.dart';

final dio = Dio(
  BaseOptions(baseUrl: 'https://api.example.com'),
);

if (kDebugMode) {
  DioDebugger.attach(
    dio,
    proxyBaseUrl: 'http://localhost:9091',
  );
}
```

Notes:
- On Android emulator use `http://10.0.2.2:9091` instead of `localhost`.
- On a real device use your machine's LAN IP (e.g. `http://192.168.1.10:9091`).

2) **OS Forward Proxy mode** — enable forward proxy in the app settings and set
it as a system-wide proxy. Useful when you want to capture traffic from any
application (not only Flutter).

Ports and runtime proxy settings are described in `../docs/SETTINGS.md`.

### Building for Production

**Web build**:
```bash
flutter build web --release --web-renderer canvaskit
```

Output will be in `build/web/` directory.

**Desktop builds** (macOS/Windows):
```bash
# macOS
flutter build macos --release

# Windows
flutter build windows --release
```

## Environment Variables

The UI adapts to the backend configuration automatically via the `/api/v1/settings` endpoint. No environment variables are required for the frontend.

## Development Tips

- **Hot Reload**: Use `r` in terminal to hot reload, `R` for hot restart
- **DevTools**: Access Flutter DevTools at the URL shown in console output
- **API Base URL**: Modify `lib/core/api/api_client.dart` if backend runs on a different port
- **Theme**: Customize colors in `lib/theme/app_theme.dart`

## Testing

Run Flutter tests:
```bash
flutter test
```

Run with coverage:
```bash
flutter test --coverage
```

## Browser Support

- Chrome/Edge (recommended)
- Firefox
- Safari (limited WebGL support)

## License

See [LICENSE](../LICENSE) in the project root.
