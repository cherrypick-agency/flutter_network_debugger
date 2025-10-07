# Network Debugger

Простой универсальный прокси для отладки HTTP и WebSocket. Подходит для локальной разработки и тестовых окружений. Есть веб‑интерфейс (открывается в браузере), десктоп и cli.

Что умеет
- Перехват и просмотр HTTP и WebSocket‑трафика
- Waterfall timeline запросов
- группировка по домену/маршруту
- Фильтры: метод, статус, MIME, минимальная длительность, по заголовкам
- Удобный поиск с подсветкой
- Детали HTTP: заголовки (с маскировкой чувствительных), тело (pretty/JSON‑дерево), TTFB/Total
- Подсказки по CORS/Cache, сводка по cookies и TLS
- Детали WebSocket: события/фреймы, пинги/понги, превью payload
- Экспорт HAR
- Искусственная задержка ответа (удобно воспроизводить «медленные сети»)
- Запись/стоп и управление записями
...

Быстрый старт
- Через CLI (автоматически скачает бинарь и откроет UI):
  ```bash
  dart pub global activate network_debugger_cli
  network-debugger
  ```
- Docker:
  ```bash
  docker compose -f deploy/docker-compose.yml up -d
  ```
- Из исходников (Go):
  ```bash
  # сервер/десктопный бинарь
  go build -o ./network-debugger ./cmd/network-debugger
  ./network-debugger

  # веб‑вариант, который сам открывает браузер
  go build -o ./network-debugger-web ./cmd/network-debugger-web
  ./network-debugger-web
  ```

Где открывается UI
- По умолчанию сервер слушает на :9091, UI доступен по адресу:
  - http://localhost:9091/_ui (или корень, если включён авто‑редирект)

Основные настройки (ENV)
- `ADDR` — адрес сервера (по умолчанию :9091)
- `DEV_MODE` — режим разработки (1/true)
- `DEFAULT_TARGET` — целевой upstream по умолчанию
- `CAPTURE_BODIES` — сохранять тела запросов/ответов (1/true)
- `RESPONSE_DELAY_MS` — фикс или диапазон, напр. `1000` или `1000-3000`
- `INSECURE_TLS` — доверять самоподписанным сертификатам (1/true)

Локальная разработка (без GitHub)
- Готовый бинарь/архив в `./dist`:
  ```bash
  network-debugger --local-dir ./dist --no-remote
  ```
- Локальный сервер артефактов:
  ```bash
  network-debugger serve-artifacts --dir ./dist --port 8099
  network-debugger --base-url http://127.0.0.1:8099 --no-remote
  ```

Полезно знать
- Кеш бинарей: macOS `~/Library/Caches/network-debugger/bin-cache`, Linux `~/.cache/network-debugger/bin-cache`, Windows `%LOCALAPPDATA%/network-debugger/bin-cache`
- Имя бинаря: `network-debugger-web` (Windows — `network-debugger-web.exe`)
