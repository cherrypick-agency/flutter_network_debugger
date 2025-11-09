# Supported Languages Overview

## Quick Comparison Table

| Язык | Runtime | Производительность | Сложность | Размер WASM | Лучше всего для |
|------|---------|-------------------|-----------|-------------|-----------------|
| **Rust** | Extism (WASM) | ⭐⭐⭐⭐⭐ | Средняя | ~500KB | Высокопроизводительная обработка, безопасность |
| **Go (TinyGo)** | Extism (WASM) | ⭐⭐⭐⭐ | Низкая | ~300KB | Быстрая разработка, Go экосистема |
| **JavaScript** | Extism (WASM) | ⭐⭐⭐ | Очень низкая | ~200KB | Прототипирование, JSON манипуляции |
| **Dart** | Subprocess | ⭐⭐ | Низкая | N/A | Flutter интеграция, stateful logic |
| **Python** | Extism (WASM) | ⭐⭐ | Средняя | ~5MB | Быстрые скрипты, data science |
| **C/C++** | Extism (WASM) | ⭐⭐⭐⭐⭐ | Высокая | ~100KB | Системное программирование, legacy code |

---

## 1. Rust 🦀

### Когда использовать:
- Максимальная производительность критична
- Нужна type safety и надежность
- Обработка binary данных
- Криптография, валидация

### Преимущества:
✅ Самая высокая производительность
✅ Zero-cost abstractions
✅ Отличная экосистема для WASM (extism-pdk)
✅ Compile-time гарантии безопасности

### Недостатки:
❌ Более крутая кривая обучения
❌ Более длительная компиляция

### Примеры:
```rust
// request modification
examples/scripts/rust/add_header.rs

// response sanitization
examples/scripts/rust/modify_response.rs
```

### Установка:
```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Add WASM target
rustup target add wasm32-unknown-unknown
```

---

## 2. Go (TinyGo) 🐹

### Когда использовать:
- Знаете Go и хотите использовать знакомый синтаксис
- Нужна быстрая разработка
- Простая логика без сложных зависимостей

### Преимущества:
✅ Простой синтаксис
✅ Быстрая разработка
✅ Хорошая производительность
✅ Маленький размер WASM (через TinyGo)

### Недостатки:
❌ TinyGo имеет ограничения (не все stdlib доступен)
❌ Некоторые features Go недоступны

### Примеры:
```go
// mock responses
examples/scripts/go/mock_response.go
```

### Установка:
```bash
# Install TinyGo
# macOS
brew tap tinygo-org/tools
brew install tinygo

# Linux
wget https://github.com/tinygo-org/tinygo/releases/download/v0.30.0/tinygo_0.30.0_amd64.deb
sudo dpkg -i tinygo_0.30.0_amd64.deb
```

---

## 3. JavaScript (AssemblyScript) 🟨

### Когда использовать:
- JavaScript/TypeScript разработчик
- Нужно быстро создать прототип
- JSON manipulation
- Простые трансформации

### Преимущества:
✅ Знакомый TypeScript-like синтаксис
✅ Очень низкий порог входа
✅ Хорошо для JSON обработки
✅ Быстрая разработка

### Недостатки:
❌ Не стандартный JavaScript (AssemblyScript имеет отличия)
❌ Меньше производительность чем Rust/Go
❌ Меньшая экосистема библиотек

### Примеры:
```typescript
// JSON transformation
examples/scripts/javascript/transform_json.ts
```

### Установка:
```bash
npm install -D assemblyscript @extism/assemblyscript-pdk
```

---

## 4. Dart 🎯

### Когда использовать:
- Flutter разработчик
- Нужен доступ к Dart packages
- Stateful логика (rate limiting, sessions)
- Dart-специфичные задачи

### Преимущества:
✅ Полный доступ к Dart ecosystem
✅ Stateful скрипты (in-memory state)
✅ Отличная type safety
✅ Знакомо Flutter разработчикам

### Недостатки:
❌ Требует установленный Dart SDK на сервере
❌ Subprocess overhead (slower startup)
❌ Больше памяти

### Примеры:
```dart
// rate limiting
examples/scripts/dart/rate_limiter.dart
```

### Установка:
```bash
# Install Dart SDK
# macOS
brew tap dart-lang/dart
brew install dart

# Linux
sudo apt update
sudo apt install dart
```

---

## 5. Python 🐍 (Experimental)

### Когда использовать:
- Быстрые скрипты
- Data processing
- Используете Python библиотеки

### Преимущества:
✅ Огромная экосистема
✅ Простой синтаксис
✅ Отлично для data science

### Недостатки:
❌ Большой размер WASM (~5MB+)
❌ Экспериментальная поддержка
❌ Медленнее других языков

### Компиляция:
```bash
# Через PyScript/Pyodide (experimental)
# См. https://github.com/extism/py-pdk
```

---

## 6. C/C++ ⚙️

### Когда использовать:
- Максимальная производительность + минимальный размер
- Legacy C/C++ код
- Системное программирование

### Преимущества:
✅ Максимальная производительность
✅ Минимальный размер WASM
✅ Переиспользование C/C++ кода

### Недостатки:
❌ Высокая сложность
❌ Manual memory management
❌ Больше возможностей для ошибок

### Компиляция:
```bash
# Через Emscripten
emcc script.c -o script.wasm \
  -s STANDALONE_WASM \
  -s EXPORTED_FUNCTIONS='["_process"]'
```

---

## Выбор языка: Decision Tree

```
Нужна максимальная производительность?
├─ Да → Rust или C/C++
└─ Нет ↓

Знаете Flutter/Dart?
├─ Да → Dart (subprocess)
└─ Нет ↓

Нужна простота разработки?
├─ Да → JavaScript (AssemblyScript)
└─ Нет ↓

Знаете Go?
├─ Да → Go (TinyGo)
└─ Нет → Rust (лучший баланс)
```

---

## Runtime Comparison

### Extism (WASM) - Рекомендуется
**Языки**: Rust, Go, JavaScript, Python, C/C++

**Плюсы**:
- Песочница (sandbox) - безопасно
- Быстрый старт (no process spawning)
- Низкая latency
- Контроль памяти и CPU

**Минусы**:
- Ограничения WASM (no threads, limited syscalls)
- Статичный код (нужна перекомпиляция)

### Subprocess (Dart) - Специальные случаи
**Языки**: Dart (можно расширить на другие)

**Плюсы**:
- Полный доступ к runtime возможностям
- Stateful scripts
- Можно использовать любые packages

**Минусы**:
- Медленнее startup
- Больше памяти
- Требует установленный runtime на сервере

---

## Рекомендации по выбору

### Для начинающих:
1. **JavaScript (AssemblyScript)** - самый простой старт
2. **Go (TinyGo)** - если знакомы с Go

### Для production:
1. **Rust** - лучший баланс производительности и безопасности
2. **Go (TinyGo)** - для простых задач

### Для Flutter разработчиков:
1. **Dart** - знакомый язык и ecosystem

### Для максимальной производительности:
1. **Rust** - modern и безопасный
2. **C/C++** - если нужен минимальный размер

---

## Примеры использования

Все примеры находятся в:
```
examples/scripts/
├── rust/           # Rust примеры
├── go/             # Go (TinyGo) примеры
├── javascript/     # AssemblyScript примеры
├── dart/           # Dart примеры
└── README.md       # Подробная документация
```

Смотрите [README.md](./README.md) для детальных примеров и инструкций.
