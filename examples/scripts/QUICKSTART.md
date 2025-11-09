# Quick Start Guide - 5 минут до первого скрипта

Создадим простой скрипт который добавляет заголовок "X-Hello: World" ко всем HTTP запросам.

## Вариант 1: JavaScript (самый простой) 🚀

### Шаг 1: Установите зависимости

```bash
# Создайте директорию для скрипта
mkdir my-first-script && cd my-first-script

# Инициализируйте npm проект
npm init -y

# Установите AssemblyScript и Extism PDK
npm install -D assemblyscript @extism/assemblyscript-pdk

# Инициализируйте AssemblyScript
npx asinit .
```

### Шаг 2: Напишите скрипт

Создайте файл `assembly/index.ts`:

```typescript
import { JSON } from "assemblyscript-json/assembly";
import { Host } from "@extism/as-pdk";

export function process(): i32 {
  // Читаем request
  const input = Host.inputString();
  const req = <JSON.Obj>JSON.parse(input);

  // Получаем текущие headers
  const headers = <JSON.Obj>req.getObj("headers")!;

  // Добавляем новый заголовок
  const helloHeader = new JSON.Arr();
  helloHeader.push(new JSON.Str("World"));
  headers.set("X-Hello", helloHeader);

  // Логируем
  Host.log("Added X-Hello header!");

  // Возвращаем модифицированный request
  Host.outputString(req.stringify());
  return 0;
}
```

### Шаг 3: Скомпилируйте

```bash
# Добавьте в package.json:
# "scripts": {
#   "asbuild": "asc assembly/index.ts --target release"
# }

npm run asbuild
# Результат: build/release.wasm
```

### Шаг 4: Загрузите в Network Debugger

```bash
# Конвертируем в base64
WASM_BASE64=$(base64 -i build/release.wasm)

# Создаём скрипт
curl -X POST http://localhost:9092/_api/v1/scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Add Hello Header",
    "description": "Adds X-Hello: World to all requests",
    "runtime": "extism",
    "code": "'$WASM_BASE64'",
    "language": "javascript",
    "triggerType": "request",
    "enabled": true,
    "priority": 10
  }'
```

### Шаг 5: Тестируйте!

```bash
# Настройте proxy
export http_proxy=http://localhost:9092
export https_proxy=http://localhost:9092

# Сделайте запрос
curl -v http://httpbin.org/headers

# В ответе вы увидите:
# "X-Hello": "World"
```

🎉 **Готово!** Ваш первый скрипт работает!

---

## Вариант 2: Rust (production-ready) 🦀

### Шаг 1: Создайте проект

```bash
cargo new --lib hello-header
cd hello-header
```

### Шаг 2: Настройте Cargo.toml

```toml
[package]
name = "hello-header"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[dependencies]
extism-pdk = "1.0"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
```

### Шаг 3: Напишите скрипт (src/lib.rs)

```rust
use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize, Serialize)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
    body: Vec<u8>,
}

#[plugin_fn]
pub fn process(Json(mut req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {
    // Добавляем заголовок
    req.headers.insert(
        "X-Hello".to_string(),
        vec!["World".to_string()]
    );

    // Логируем
    extism_pdk::log!(
        extism_pdk::LogLevel::Info,
        "Added X-Hello header to {}",
        req.url
    );

    Ok(Json(req))
}
```

### Шаг 4: Компиляция и загрузка

```bash
# Компилируем
cargo build --target wasm32-unknown-unknown --release

# Загружаем
WASM=$(base64 -i target/wasm32-unknown-unknown/release/hello_header.wasm)

curl -X POST http://localhost:9092/_api/v1/scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Hello Header (Rust)",
    "runtime": "extism",
    "code": "'$WASM'",
    "language": "rust",
    "triggerType": "request",
    "enabled": true
  }'
```

---

## Вариант 3: Dart (для Flutter разработчиков) 🎯

### Шаг 1: Создайте скрипт

Создайте файл `hello_header.dart`:

```dart
import 'dart:convert';
import 'dart:io';

void main() {
  stdin
      .transform(utf8.decoder)
      .transform(const LineSplitter())
      .listen((line) {
    try {
      final rpcRequest = jsonDecode(line);
      final params = rpcRequest['params'];

      if (params['request'] != null) {
        final req = params['request'];

        // Добавляем заголовок
        req['headers'] ??= {};
        req['headers']['X-Hello'] = ['World'];

        stderr.writeln('[Dart] Added X-Hello header');

        // Возвращаем response
        final rpcResponse = {
          'jsonrpc': '2.0',
          'id': rpcRequest['id'],
          'result': req,
        };

        stdout.writeln(jsonEncode(rpcResponse));
      }
    } catch (e) {
      stderr.writeln('[Error] $e');
    }
  });
}
```

### Шаг 2: Загрузите

```bash
# Конвертируем в base64
CODE=$(base64 -i hello_header.dart)

curl -X POST http://localhost:9092/_api/v1/scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Hello Header (Dart)",
    "runtime": "dart",
    "code": "'$CODE'",
    "language": "dart",
    "triggerType": "request",
    "enabled": true
  }'
```

---

## Управление скриптами

### Посмотреть все скрипты

```bash
curl http://localhost:9092/_api/v1/scripts
```

### Получить конкретный скрипт

```bash
curl http://localhost:9092/_api/v1/scripts/{id}
```

### Включить/выключить скрипт

```bash
curl -X PATCH http://localhost:9092/_api/v1/scripts/{id}/toggle \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

### Удалить скрипт

```bash
curl -X DELETE http://localhost:9092/_api/v1/scripts/{id}
```

---

## Advanced: Pattern Matching

Запускать скрипт только для определенных запросов:

```bash
curl -X POST http://localhost:9092/_api/v1/scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Only Script",
    "runtime": "extism",
    "code": "'$WASM_BASE64'",
    "language": "rust",
    "triggerType": "request",
    "enabled": true,
    "matchRules": {
      "methods": ["GET", "POST"],
      "hostPattern": "api.*.com",
      "pathPattern": "/v1/*",
      "patternType": "wildcard"
    }
  }'
```

**Match Rules**:
- `methods`: Список HTTP методов
- `hostPattern`: Wildcard или regex для хоста
- `pathPattern`: Wildcard или regex для пути
- `patternType`: `"wildcard"` или `"regex"`

---

## Debugging

### Просмотр логов

Скрипты логируют через `extism_pdk::log!` (Rust), `Host.log()` (JS), или `stderr` (Dart).

Логи видны в консоли Network Debugger:

```bash
# Запустите с DEV_MODE для подробных логов
DEV_MODE=1 ./bin/network-debugger
```

### Проверка ошибок выполнения

```bash
# Получите детали скрипта
curl http://localhost:9092/_api/v1/scripts/{id}

# Проверьте поле "lastError" (если есть ошибка)
```

---

## Best Practices

### 1. Начинайте с простого
Сначала сделайте скрипт который просто логирует request, затем добавляйте логику.

### 2. Тестируйте локально
Используйте unit тесты для вашего языка перед компиляцией в WASM.

### 3. Используйте type safety
Определяйте структуры для HTTPRequest/HTTPResponse.

### 4. Обрабатывайте ошибки
Всегда проверяйте результаты парсинга JSON и других операций.

### 5. Оптимизируйте размер
```bash
# Rust: используйте release mode
cargo build --target wasm32-unknown-unknown --release

# Опционально: оптимизируйте WASM
wasm-opt -Oz input.wasm -o output.wasm
```

---

## Troubleshooting

### "Script failed to load"
- Проверьте что WASM файл валидный
- Убедитесь что экспортирован `process` function

### "Timeout error"
- Скрипт работает слишком долго (>5000ms по умолчанию)
- Увеличьте `config.timeoutMs` в скрипте

### "Memory limit exceeded"
- Скрипт использует >10MB памяти
- Увеличьте `config.memoryLimitMB`

### "Dart executor not available"
- Dart SDK не установлен
- Установите: `brew install dart` (macOS) или см. https://dart.dev/get-dart

---

## Next Steps

1. Изучите [примеры](./README.md) для других use cases
2. Прочитайте [сравнение языков](./LANGUAGES.md)
3. Посмотрите [документацию Extism](https://extism.org/docs)
4. Экспериментируйте с разными trigger types: `request`, `response`, `both`

**Happy Scripting!** 🚀
