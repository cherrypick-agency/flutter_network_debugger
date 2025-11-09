# Система автообновлений Network Debugger

## Обзор

Network Debugger использует **custom легковесное решение** для автообновлений через GitHub Releases API.

**Преимущества нашего подхода:**
- ✅ Поддержка всех платформ: macOS, Windows, Linux
- ✅ Нет конфликтов зависимостей
- ✅ Простая интеграция с GitHub Releases
- ✅ Полный контроль над процессом
- ✅ Минимальные требования (только http + url_launcher)

## Архитектура

```
┌─────────────────┐
│  BootstrapApp   │
│   (main.dart)   │
└────────┬────────┘
         │
         │ 1. При старте проверяет обновления
         ▼
┌─────────────────┐
│  UpdateService  │──────► GitHub Releases API
└────────┬────────┘
         │
         │ 2. Если есть новая версия
         ▼
┌─────────────────┐
│  UpdateDialog   │
│   показывает    │
│  информацию о   │
│   обновлении    │
└────────┬────────┘
         │
         │ 3. Пользователь выбирает действие
         ▼
    ┌────┴────────────────┐
    │                     │
Download             Skip/Later
    │                     │
    ▼                     ▼
Открывает         Сохраняет в
GitHub Release    SharedPreferences
```

## Настройка

### 1. Обновите GitHub repo в main.dart

Откройте `frontend/lib/main.dart` и замените:

```dart
_updateService = UpdateService(
  githubOwner: 'cherrypick-agency',     // ← GitHub organization
  githubRepo: 'flutter_network_debugger', // ← Название репозитория
  currentVersion: '1.0.0',               // ← Из pubspec.yaml
);
```

**Текущая конфигурация:**
```dart
_updateService = UpdateService(
  githubOwner: 'cherrypick-agency',
  githubRepo: 'flutter_network_debugger',
  currentVersion: '1.0.0',
);
```

### 2. Обновите версию в pubspec.yaml

При каждом релизе обновляйте версию в `frontend/pubspec.yaml`:

```yaml
version: 1.0.1+2  # major.minor.patch+build
```

И в `main.dart`:

```dart
currentVersion: '1.0.1',
```

## Как работает проверка обновлений

### 1. Автоматическая проверка

Приложение автоматически проверяет обновления:
- **При старте приложения**
- **Не чаще раза в час** (кэширование)

### 2. GitHub Releases API

UpdateService делает HTTP запрос:

```
GET https://api.github.com/repos/OWNER/REPO/releases/latest
```

Ответ содержит:
- `tag_name`: версия релиза (например, "v1.0.1")
- `body`: changelog
- `assets`: список файлов для скачивания

### 3. Определение платформы

UpdateService автоматически определяет нужный файл:

| Платформа | Приоритет файлов                |
|-----------|---------------------------------|
| macOS     | `*.dmg`                         |
| Windows   | `*.msi`                         |
| Linux     | `*.AppImage` → `*.deb` → `*.tar.gz` |

### 4. Сравнение версий

Версии сравниваются по семантическому версионированию:
- `1.0.1` > `1.0.0` ✅
- `1.1.0` > `1.0.9` ✅
- `2.0.0` > `1.9.9` ✅

## Пользовательский опыт

### Диалог обновления

Когда доступна новая версия, показывается красивый диалог:

```
┌─────────────────────────────────────────┐
│ 🔄 Update Available                     │
├─────────────────────────────────────────┤
│                                         │
│  New Version: v1.0.1     Size: 45.2 MB │
│                                         │
│  What's New:                            │
│  • Fixed critical bug                   │
│  • Added new feature                    │
│  • Performance improvements             │
│                                         │
│  ℹ️ Download from GitHub                 │
│                                         │
├─────────────────────────────────────────┤
│  [Skip] [Later] [Download Update] ⬇️    │
└─────────────────────────────────────────┘
```

### Действия пользователя

1. **Download Update** → Открывает страницу GitHub Release в браузере
2. **Skip This Version** → Больше не показывать этот релиз
3. **Remind Me Later** → Показать при следующем запуске

## Настройка Skip Version

Пропущенные версии сохраняются в SharedPreferences:

```dart
// Пользователь нажал "Skip This Version"
await _updateService.skipVersion('1.0.1');

// Сброс (для тестирования)
await _updateService.clearSkippedVersion();
```

## Кэширование

Результаты проверки кэшируются на 1 час:

```dart
// Принудительная проверка (игнорирует кэш)
await _updateService.checkForUpdates(forceCheck: true);
```

## Требования к GitHub Releases

### Именование файлов

Файлы в Release должны иметь правильные расширения:

```
NetworkDebugger-1.0.1-macos-arm64.dmg       ✅
NetworkDebugger-1.0.1-macos-amd64.dmg       ✅
NetworkDebugger-1.0.1-windows-amd64.msi     ✅
NetworkDebugger-1.0.1-linux-amd64.AppImage  ✅
NetworkDebugger-1.0.1-linux-amd64.deb       ✅
```

### Версии (Tags)

GitHub Release tag должен соответствовать формату:

```
v1.0.0   ✅
v1.0.1   ✅
1.0.0    ✅ (тоже работает)
release-1.0.0  ❌
```

### Changelog

Тело Release (description) будет показано в диалоге как changelog:

```markdown
## What's New

### Features
- Added auto-update functionality
- Improved startup dialog

### Bug Fixes
- Fixed memory leak in session monitor
- Corrected port validation

### Performance
- Optimized websocket handling
```

## Тестирование

### Локальное тестирование

1. Создайте тестовый Release в GitHub
2. Установите более старую версию в `main.dart`:
   ```dart
   currentVersion: '0.9.0',
   ```
3. Запустите приложение - должен появиться диалог обновления

### Проверка API вручную

```bash
curl -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/OWNER/REPO/releases/latest
```

## Отключение автообновлений

Для разработки можно временно отключить:

```dart
// В _checkForUpdates() добавьте в начало:
if (kDebugMode) {
  return; // Skip update check in debug mode
}
```

## Troubleshooting

### Обновления не проверяются

**Проблема:** Нет интернета или GitHub API недоступен

**Решение:** Приложение тихо игнорирует ошибки. Проверьте логи:

```dart
Logger.root.level = Level.FINE;
Logger.root.onRecord.listen((record) {
  print('${record.level.name}: ${record.time}: ${record.message}');
});
```

### Неправильная версия определяется

**Проблема:** currentVersion не совпадает с pubspec.yaml

**Решение:** Обновите оба места при каждом релизе

### Asset не находится

**Проблема:** Файл не имеет правильного расширения

**Решение:** Убедитесь что в Release есть файлы с `.dmg`, `.msi`, `.AppImage` расширениями

## Best Practices

1. **Всегда синхронизируйте версии:**
   - `pubspec.yaml` version
   - `main.dart` currentVersion
   - GitHub Release tag

2. **Пишите подробные changelogs:**
   - Группируйте изменения (Features, Bug Fixes, etc.)
   - Будьте конкретны
   - Упоминайте breaking changes

3. **Тестируйте перед релизом:**
   - Создайте draft release
   - Проверьте автообновление
   - Только потом публикуйте

4. **Правильная нумерация:**
   - Patch (0.0.x): Bug fixes
   - Minor (0.x.0): New features
   - Major (x.0.0): Breaking changes

## Будущие улучшения

Потенциальные улучшения системы:

- [ ] **Автоматическая установка** (сейчас только открывает браузер)
- [ ] **Delta updates** (загружать только изменения)
- [ ] **In-app download** с progress bar
- [ ] **Background updates** (скачивать в фоне)
- [ ] **Automatic restart** после установки
- [ ] **Rollback mechanism** (откат на предыдущую версию)

## См. также

- [GitHub Releases API Documentation](https://docs.github.com/en/rest/releases/releases)
- [Semantic Versioning](https://semver.org/)
- [Flutter Desktop Documentation](https://docs.flutter.dev/desktop)
