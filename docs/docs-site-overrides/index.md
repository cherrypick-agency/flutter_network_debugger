---
layout: home
hero:
  name: "Network Debugger"
  text: Документация по Dart/Flutter пакетам
  tagline: Open-source набор инструментов для отладки HTTP, WebSocket и Socket.IO трафика с UI, CLI и прокси-бэкендом
  actions:
    - theme: brand
      text: Быстрый старт
      link: /guide/
    - theme: alt
      text: API Reference
      link: /api/
    - theme: alt
      text: GitHub Repository
      link: https://github.com/cherrypick-agency/flutter_network_debugger

features:
  - title: Что это за проект
    details: Полноценный network-inspector для локальной разработки и тестовых окружений. Работает офлайн и подходит для WEB, desktop и CLI сценариев.
  - title: Что внутри документации
    details: Гайды + API reference по пакетам из monorepo `dart_packages` с локальным поиском и автогенерацией документации из кода.
  - title: Для кого
    details: Для Flutter/Dart разработчиков, которым нужен быстрый и удобный способ смотреть HTTP/WS трафик приложения в одном месте.
---

## Пакеты в этом workspace

| Package | Назначение |
| --- | --- |
| `network_debugger` | CLI launcher для локального прокси и UI |
| `dio_debugger` | Подключение прокси к клиенту Dio |
| `http_debugger` | Глобальный HTTP interception через `HttpOverrides` |
| `web_socket_debugger` | Интеграция с `dart:io` WebSocket |
| `web_socket_channel_debugger` | Интеграция с `package:web_socket_channel` |
| `socket_io_debugger` | Интеграция с Socket.IO клиентом |
| `firebase_database_debugger` | Интеграция Firebase Realtime Database операций в ingest API |
| `hex_viewer` | Flutter widget для просмотра бинарных payload в HEX |

## Как читать документацию

1. Начни с раздела `Guide`, чтобы быстро понять общий сценарий подключения.
2. Перейди в `API Reference` для деталей классов и методов конкретного пакета.
3. Используй поиск в правом верхнем углу, чтобы мгновенно находить символы.
