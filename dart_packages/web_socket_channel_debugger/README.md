# web_socket_channel_debugger

<p align="left">
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_channel_debugger.yml"><img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_channel_debugger.yml/badge.svg?branch=main" alt="CI" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License" /></a>
</p>

Инструмент для быстрого подключения прокси network-debugger к `package:web_socket_channel` (reverse/forward режимы) для локальной отладки и перехвата трафика.

## Установка

Добавьте в `pubspec.yaml` вашего проекта зависимость на `web_socket_channel` и этот пакет.

## Быстрый старт

```dart
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

void main() async {
  const upstream = 'wss://echo.websocket.events';
  final cfg = WebSocketChannelDebugger.attach(
    baseUrl: upstream,
    proxyBaseUrl: 'http://localhost:9091',
    proxyPath: '/wsproxy',
  );

  final channel = WebSocketChannelDebugger.connect(config: cfg);
  channel.stream.listen((m) => print('> $m'));
  channel.sink.add('hello');
}
```

## Режимы

- reverse (по умолчанию): трафик идёт на `http://<proxy>/<proxyPath>?_target=<ws target>`
- forward: использует `HttpClient.findProxy`, не меняет URL подключения

Выбор через `--dart-define=SOCKET_PROXY_MODE=reverse|forward|none` или ENV `SOCKET_PROXY_MODE`.

## Переменные окружения

- `SOCKET_PROXY` — URL прокси (`http://localhost:9091`)
- `SOCKET_PROXY_PATH` — путь WS-прокси (обычно `/wsproxy`)
- `SOCKET_PROXY_ENABLED` — включить/выключить (`true|false`)
- `SOCKET_PROXY_MODE` — `reverse|forward|none`
- `SOCKET_PROXY_ALLOW_BAD_CERTS` — разрешить self-signed в forward
- `SOCKET_UPSTREAM_URL` — исходный WS, если baseUrl указывает на сам прокси
- `SOCKET_UPSTREAM_TARGET` — явный полный target для `_target`

## Ссылки

- web_socket_channel на pub.dev: `https://pub.dev/packages/web_socket_channel`
- web_socket_channel исходники: `https://github.com/dart-lang/http/tree/master/pkgs/web_socket_channel`

## Лицензия

Apache-2.0


