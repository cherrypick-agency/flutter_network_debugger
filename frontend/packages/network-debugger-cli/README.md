# network_debugger_cli

CLI для запуска бинарника `network-debugger` (web-версия):
- автодетект ОС/архитектуры
- скачивание нужного архива с GitHub Pages/релизов
- кеширование в `~/.cache/network-debugger/bin-cache` (macOS: `~/Library/Caches/network-debugger/bin-cache`, Windows: `%LOCALAPPDATA%/network-debugger/bin-cache`)
- запуск бинарника с пробросом логов в консоль и авто-открытием в браузере (реализуется самим бинарём)

## Установка

В корне фронтенда:

```bash
# однократно
dart pub global activate --source path frontend/packages/network-debugger-cli
```

Убедитесь, что `~/.pub-cache/bin` в PATH (Windows: `%APPDATA%\\Pub\\Cache\\bin`).

## Запуск (одна команда)

```bash
network-debugger
```

Опции:
- `-v, --version` — форсировать конкретный тег релиза (по умолчанию берётся latest через GitHub API)
- `-f, --force` — принудительно перекачать и распаковать

Переменные окружения:
- `NETWORK_DEBUGGER_GITHUB_REPO` — репозиторий с релизами (по умолчанию `belief/network-debugger`)
- `NETWORK_DEBUGGER_DOWNLOAD_BASE` — базовый URL для прямой загрузки артефактов (по умолчанию `https://belief.github.io/network-debugger/assets/downloads`)

## Примечания
- На Windows распаковка ZIP выполняется через PowerShell `Expand-Archive`.
- На Unix требуется доступный `tar` (и `unzip` для .zip).

## Локальная разработка

Вариант A: готовый бинарь или архив локально

```bash
# бинарь или архив лежит в ./dist
network-debugger --local-dir ./dist --no-remote
```

Структура:
- если найден `network-debugger-web` (`.exe` на Windows) — он будет запущен напрямую
- иначе ищется архив `network-debugger-web_<os>_<arch>.(zip|tar.gz)` → распаковка → бинарь находится рекурсивно

Сборка бинаря (пример):
- macOS/Linux:
  - `go build -o ./dist/network-debugger-web ./cmd/network-debugger-web`
- Windows (PowerShell):
  - `go build -o .\dist\network-debugger-web.exe .\cmd\network-debugger-web`

Вариант B: локальный статический сервер

```bash
network-debugger serve-artifacts --dir ./dist --port 8099
# затем во втором терминале
network-debugger --base-url http://127.0.0.1:8099
```

Также поддерживается `file://`:

```bash
network-debugger --base-url file:///absolute/path/to/dist
```

Доп. флаги/ENV:
- `--base-url`, ENV `NETWORK_DEBUGGER_BASE_URL` — http(s) или file://
- `--local-dir`, ENV `NETWORK_DEBUGGER_LOCAL_DIR` — путь к папке с бинарём/архивом
- `--artifact-name`, ENV `NETWORK_DEBUGGER_ARTIFACT_NAME` — кастомное имя файла
- `--no-remote` (по умолчанию отключён) — запретить фолбэк в удалённые источники
- `--allow-remote` — разрешить фолбэк при наличии локального источника
- `--no-cache` — игнорировать кэш

### Кеш
- macOS: `~/Library/Caches/network-debugger/bin-cache`
- Linux: `~/.cache/network-debugger/bin-cache`
- Windows: `%LOCALAPPDATA%/network-debugger/bin-cache`

Очистка кеша:

```bash
rm -rf ~/.cache/network-debugger/bin-cache            # Linux
rm -rf ~/Library/Caches/network-debugger/bin-cache    # macOS
rd /s /q %LOCALAPPDATA%\network-debugger\bin-cache   # Windows (cmd)
```

### Типичные ошибки
- `tar: not found` или `unzip: not found` (Linux/macOS)
  - Установите `tar`/`unzip` через пакетный менеджер
- `Access is denied` (Windows)
  - Закройте предыдущий запущенный бинарь/терминал, проверьте права файла
- `Binary not found inside local archive`
  - Убедитесь, что архив содержит бинарь с именем `network-debugger-web(.exe)` внутри какой-либо подпапки
