# web_socket_debugger

Пакет для подключения прокси network-debugger к `package:web_socket` (reverse/forward) для локальной отладки и перехвата WebSocket-трафика.

## Установка

Добавьте в `pubspec.yaml`:

```yaml
dependencies:
  web_socket: ^1.0.1
  web_socket_debugger:
    path: ../web_socket_debugger
```

## Пример

```dart
import 'package:web_socket_debugger/web_socket_debugger.dart';

Future<void> main() async {
  const upstream = 'wss://ws.postman-echo.com/raw';
  final cfg = WebSocketDebugger.attach(
    baseUrl: upstream,
    proxyBaseUrl: 'http://localhost:9091',
    proxyPath: '/wsproxy',
  );

  final socket = await WebSocketDebugger.connect(config: cfg);
  socket.events.listen((e) {
    print(e);
  });
  socket.sendText('hello');
}
```

## Режимы
- reverse (по умолчанию): коннект к `http://<proxy>/<path>` + `_target=<ws(s)://...>`
- forward: `HttpClient.findProxy`, URL без изменений

Переключение через `--dart-define=SOCKET_PROXY_MODE=reverse|forward|none` или ENV `SOCKET_PROXY_MODE`.

## ENV/defines
- `SOCKET_PROXY`, `SOCKET_PROXY_PATH`, `SOCKET_PROXY_MODE`, `SOCKET_PROXY_ENABLED`
- `SOCKET_PROXY_ALLOW_BAD_CERTS` — для forward
- `SOCKET_UPSTREAM_URL`, `SOCKET_UPSTREAM_TARGET`

## Ссылки
- web_socket на pub.dev: `https://pub.dev/packages/web_socket`

Лицензия: Apache-2.0
