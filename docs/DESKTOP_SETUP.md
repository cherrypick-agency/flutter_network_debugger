# Network Debugger - Desktop Application Setup

## Обзор

Network Debugger поддерживает native desktop приложения для:
- **macOS** (Intel x86_64 и Apple Silicon arm64) - DMG installer
- **Windows** (64-bit) - ZIP archive с install.bat
- **Linux** (64-bit) - tar.gz и deb packages

Desktop приложение запускает как Flutter UI, так и Go proxy server в одном процессе.

## Архитектура

```
┌─────────────────────────────────────┐
│   Flutter Desktop App (UI)          │
│                                      │
│  ┌────────────────────────────────┐ │
│  │  BootstrapApp                  │ │
│  │  - Startup Dialog              │ │
│  │  - Port Configuration          │ │
│  │  - Auto-Update Check           │ │
│  └────────┬───────────────────────┘ │
│           │                          │
│           ▼                          │
│  ┌────────────────────────────────┐ │
│  │  GoServerManager               │ │
│  │  - Launch subprocess           │ │
│  │  - Health monitoring           │ │
│  │  - Graceful shutdown           │ │
│  └────────┬───────────────────────┘ │
│           │                          │
└───────────┼──────────────────────────┘
            │ Process.start()
            ▼
┌─────────────────────────────────────┐
│   Go Server (subprocess)             │
│   - Forward Proxy                    │
│   - HTTP API                         │
│   - WebSocket connections            │
│   - Session storage                  │
└─────────────────────────────────────┘
```

## Требования для разработки

### Общие
- Flutter SDK 3.x или выше
- Go 1.22.x или выше
- Git

### macOS
- macOS 11 (Big Sur) или новее
- Xcode Command Line Tools
- `brew install create-dmg` (опционально, для красивых DMG)

### Windows
- Windows 10 или новее
- Visual Studio 2022 или Visual Studio Build Tools
- PowerShell 5.1 или выше

### Linux
- Ubuntu 20.04+ или аналогичный дистрибутив
- GTK 3.0 development headers
- Required packages:
  ```bash
  sudo apt-get install \
    clang cmake ninja-build pkg-config \
    libgtk-3-dev liblzma-dev libstdc++-12-dev
  ```

## Локальная сборка

### macOS

```bash
# Включить desktop support (первый раз)
cd frontend
flutter create --platforms=macos .

# Собрать DMG
cd ..
chmod +x scripts/package-macos.sh
VERSION=1.0.0 ./scripts/package-macos.sh

# Результат: dist/NetworkDebugger-1.0.0-macos-{arch}.dmg
```

### Windows

```powershell
# Включить desktop support (первый раз)
cd frontend
flutter create --platforms=windows .

# Собрать ZIP
cd ..
.\scripts\package-windows.ps1 -Version "1.0.0"

# Результат: dist\NetworkDebugger-1.0.0-windows-amd64.zip
```

### Linux

```bash
# Включить desktop support (первый раз)
cd frontend
flutter create --platforms=linux .

# Собрать tar.gz и deb
cd ..
chmod +x scripts/package-linux.sh
VERSION=1.0.0 ARCH=amd64 ./scripts/package-linux.sh

# Результат:
# - dist/NetworkDebugger-1.0.0-linux-amd64.tar.gz
# - dist/network-debugger_1.0.0_amd64.deb
```

## CI/CD Pipeline

### GitHub Actions Workflow

Workflow `.github/workflows/build-desktop.yml` автоматически:

1. **Триггеры:**
   - Push в main
   - Pull requests
   - Version tags (v*.*.*)
   - Manual dispatch

2. **Jobs:**
   - `build-macos`: Собирает DMG для macOS (x86_64 и arm64)
   - `build-windows`: Собирает ZIP для Windows (amd64)
   - `build-linux`: Собирает tar.gz и deb для Linux (amd64)
   - `release`: При push тега создает GitHub Release с artifacts
   - `summary`: Показывает статус всех builds

3. **Артефакты:**
   - Хранятся 7 дней для dev builds
   - Прикрепляются к GitHub Release для version tags

### Создание релиза

```bash
# 1. Обновите версию в pubspec.yaml
cd frontend
# version: 1.0.1+2

# 2. Обновите currentVersion в main.dart
# currentVersion: '1.0.1',

# 3. Закоммитьте изменения
git add frontend/pubspec.yaml frontend/lib/main.dart
git commit -m "chore: bump version to 1.0.1"

# 4. Создайте и push tag
git tag v1.0.1
git push origin main
git push origin v1.0.1

# 5. GitHub Actions автоматически:
# - Соберет для всех платформ
# - Создаст GitHub Release
# - Загрузит все installers
```

## Установка

### macOS

1. Скачайте DMG для вашей архитектуры:
   - Intel: `NetworkDebugger-*-macos-x86_64.dmg`
   - Apple Silicon: `NetworkDebugger-*-macos-arm64.dmg`

2. Откройте DMG и перетащите приложение в Applications

3. При первом запуске: System Preferences → Security & Privacy → "Open Anyway"

### Windows

1. Скачайте `NetworkDebugger-*-windows-amd64.zip`

2. Распакуйте ZIP в любую папку

3. Запустите `install.bat` (создаст ярлыки на Desktop и Start Menu)

4. Приложение установится в `%LOCALAPPDATA%\NetworkDebugger`

### Linux

#### Через .deb (Ubuntu/Debian):
```bash
sudo dpkg -i network-debugger_*_amd64.deb
network-debugger
```

#### Через tar.gz (любой дистрибутив):
```bash
# Извлечь архив
tar -xzf NetworkDebugger-*-linux-amd64.tar.gz
cd NetworkDebugger-*

# Установить
./install.sh

# Запустить
network-debugger
```

## Использование

### Первый запуск

При запуске приложения появится **Startup Dialog** с настройками:

```
┌─────────────────────────────────────┐
│ Network Debugger - Configuration    │
├─────────────────────────────────────┤
│                                     │
│  API Server Port:    [9092]         │
│  Forward Proxy Port: [9093]         │
│                                     │
│  ℹ️ These ports must be available   │
│  and different from each other      │
│                                     │
├─────────────────────────────────────┤
│           [Cancel]  [Start]         │
└─────────────────────────────────────┘
```

**Настройки:**
- **API Server Port**: Порт для UI и REST API (по умолчанию: 9092)
- **Forward Proxy Port**: Порт для forward proxy (по умолчанию: 9093)

**Валидация:**
- Порты должны быть в диапазоне 1024-65535
- Порты должны быть разными
- Порты должны быть свободны

После нажатия **Start**:
1. Go server запускается с указанными портами
2. Проверяется health endpoint (`/_health`)
3. Flutter UI подключается к серверу
4. Приложение готово к использованию

### Настройки сохраняются

Выбранные порты сохраняются в SharedPreferences:
- macOS: `~/Library/Preferences/com.belieflab.networkDebugger`
- Windows: Registry `HKCU\Software\belieflab\network-debugger`
- Linux: `~/.local/share/network-debugger/shared_preferences.json`

При следующем запуске используются сохраненные значения.

### Изменение портов

Чтобы изменить порты после установки:
1. Settings → Server Settings → Restart with different ports
2. Или удалить saved preferences и перезапустить

## Автообновления

Desktop приложение автоматически проверяет обновления через GitHub Releases API.

### Как это работает

1. **При старте** приложения (не чаще раза в час)
2. Проверяется latest release через API
3. Сравнивается версия с текущей (semantic versioning)
4. Если есть новая версия → показывается диалог

### Update Dialog

```
┌─────────────────────────────────────┐
│ 🔄 Update Available                 │
├─────────────────────────────────────┤
│                                     │
│  New Version: v1.0.1  Size: 45 MB   │
│                                     │
│  What's New:                        │
│  • Fixed critical bug               │
│  • Added new feature                │
│  • Performance improvements         │
│                                     │
├─────────────────────────────────────┤
│  [Skip]  [Later]  [Download] ⬇️     │
└─────────────────────────────────────┘
```

**Действия:**
- **Download Update**: Открывает страницу GitHub Release в браузере
- **Skip This Version**: Больше не показывать этот релиз
- **Remind Me Later**: Показать при следующем запуске

См. [AUTO_UPDATE.md](AUTO_UPDATE.md) для деталей.

## Troubleshooting

### macOS: "App is damaged and can't be opened"

```bash
xattr -cr "/Applications/Network Debugger.app"
```

### Windows: "Windows protected your PC"

1. Click "More info"
2. Click "Run anyway"

Или запустите установщик из-под Administrator.

### Linux: Port permission denied (< 1024)

```bash
# Используйте порты >= 1024 (например, 9092/9093)
# Или дайте capability:
sudo setcap 'cap_net_bind_service=+ep' ~/.local/share/network-debugger/resources/server_linux_amd64
```

### Go server не запускается

**Проверьте:**
1. Порты не заняты: `lsof -i :9092` (macOS/Linux) или `netstat -ano | findstr 9092` (Windows)
2. Binary существует в Resources:
   - macOS: `Network Debugger.app/Contents/Resources/server_darwin_*`
   - Windows: `resources\server_windows_amd64.exe`
   - Linux: `resources/server_linux_amd64`
3. Binary executable: `chmod +x <binary>`

**Логи:**
- macOS: Console.app → Filter: "network-debugger"
- Windows: Event Viewer → Application logs
- Linux: `journalctl -f | grep network-debugger`

### Flutter web assets не загружаются

Go server должен иметь доступ к embedded Flutter web assets.

**Проверка:**
```bash
# После build должна быть папка:
frontend/build/macos/Build/Products/Release/Network Debugger.app/Contents/Frameworks/App.framework/Resources/flutter_assets/

# Она должна содержать:
- assets/
- fonts/
- packages/
```

## Разработка

### Локальный запуск для разработки

#### Вариант 1: Отдельные процессы (рекомендуется для dev)

Terminal 1 - Go server:
```bash
cd cmd/network-debugger
go run . --api-port 9092 --proxy-port 9093
```

Terminal 2 - Flutter desktop:
```bash
cd frontend
flutter run -d macos  # или -d windows / -d linux
```

#### Вариант 2: Полная desktop интеграция

```bash
# 1. Соберите Go binary
cd cmd/network-debugger
go build -o ../../frontend/macos/Runner/Resources/server_darwin_arm64 .

# 2. Запустите Flutter desktop
cd ../../frontend
flutter run -d macos
```

### Отладка

#### Flutter DevTools

```bash
cd frontend
flutter run -d macos --observatory-port=9090
# Затем откройте DevTools в браузере
```

#### Go delve debugger

```bash
cd cmd/network-debugger
dlv debug . -- --api-port 9092 --proxy-port 9093
```

### Тестирование packaging scripts

```bash
# macOS
VERSION=dev ./scripts/package-macos.sh
open dist/*.dmg

# Windows
.\scripts\package-windows.ps1 -Version "dev"
Expand-Archive dist\*.zip -DestinationPath dist\test

# Linux
VERSION=dev ./scripts/package-linux.sh
tar -xzf dist/*.tar.gz -C dist/
```

## Структура проекта

```
go-proxy/
├── cmd/network-debugger/         # Go server entry point
│   └── main.go                    # CLI flags: --api-port, --proxy-port, --data-dir
├── frontend/
│   ├── lib/
│   │   ├── core/
│   │   │   ├── desktop/
│   │   │   │   └── desktop_bootstrap.dart  # Desktop initialization
│   │   │   ├── go_server/
│   │   │   │   ├── go_server_manager.dart  # Process management
│   │   │   │   ├── go_server_path_io.dart  # Binary path resolution
│   │   │   │   └── go_server_path_stub.dart
│   │   │   └── update/
│   │   │       ├── update_service.dart      # GitHub Releases API
│   │   │       ├── update_dialog.dart       # Update UI
│   │   │       └── update_info.dart         # Version comparison
│   │   ├── features/startup/
│   │   │   └── startup_dialog.dart          # Port configuration dialog
│   │   └── main.dart                         # Entry point with BootstrapApp
│   ├── macos/                     # macOS specific
│   ├── windows/                   # Windows specific
│   └── linux/                     # Linux specific
├── scripts/
│   ├── package-macos.sh           # macOS DMG builder
│   ├── package-windows.ps1        # Windows ZIP builder
│   └── package-linux.sh           # Linux tar.gz/deb builder
├── .github/workflows/
│   └── build-desktop.yml          # CI/CD for desktop builds
└── docs/
    ├── DESKTOP_SETUP.md           # This file
    └── AUTO_UPDATE.md             # Auto-update documentation
```

## Best Practices

### Версионирование

Всегда синхронизируйте версии:
1. `frontend/pubspec.yaml`: `version: 1.0.1+2`
2. `frontend/lib/main.dart`: `currentVersion: '1.0.1'`
3. Git tag: `v1.0.1`

### Changelog

Используйте conventional commits:
- `feat:` - новые фичи
- `fix:` - исправления багов
- `chore:` - технические изменения
- `docs:` - документация

### Тестирование перед релизом

```bash
# 1. Соберите локально для всех платформ
make desktop-all  # (или используйте packaging scripts)

# 2. Протестируйте установку
# - macOS: откройте DMG, установите, запустите
# - Windows: распакуйте ZIP, install.bat, запустите
# - Linux: установите deb/tar.gz, запустите

# 3. Проверьте автообновление
# - Установите старую версию
# - Создайте draft release с новой версией
# - Запустите приложение - должен показаться update dialog

# 4. Создайте настоящий release
git tag v1.0.1 && git push origin v1.0.1
```

## FAQ

**Q: Можно ли cross-compile для всех платформ с одной машины?**

A: Частично. Go поддерживает cross-compilation, но Flutter desktop требует native toolchain для каждой платформы. Используйте GitHub Actions для multi-platform builds.

**Q: Как добавить code signing?**

A:
- macOS: `codesign --deep --force --verify --verbose --sign "Developer ID" "Network Debugger.app"`
- Windows: Используйте `signtool.exe` с certificate
- Linux: Code signing обычно не требуется

**Q: Можно ли создать .app/.exe installer вместо DMG/ZIP?**

A: Да:
- macOS: Используйте `pkgbuild` для .pkg installer
- Windows: Используйте Inno Setup или NSIS для .exe installer
- Linux: Snap, Flatpak, или AppImage

**Q: Как обновить Go dependencies в desktop app?**

A:
```bash
# Пересоберите Go binary
cd cmd/network-debugger
go mod tidy
go build -o <destination> .

# Запустите packaging script
cd ../..
./scripts/package-<platform>.sh
```

**Q: Можно ли запустить несколько инстансов приложения?**

A: Да, но убедитесь что они используют разные порты. Startup dialog предотвращает конфликты.

## Ссылки

- [Flutter Desktop](https://docs.flutter.dev/desktop)
- [Go Cross-compilation](https://go.dev/doc/install/source#environment)
- [GitHub Releases API](https://docs.github.com/en/rest/releases/releases)
- [Semantic Versioning](https://semver.org/)
- [macOS App Distribution](https://developer.apple.com/documentation/xcode/distributing-your-app-for-beta-testing-and-releases)
- [Windows App Packaging](https://docs.microsoft.com/en-us/windows/msix/desktop/desktop-to-uwp-packaging-dot-net)
