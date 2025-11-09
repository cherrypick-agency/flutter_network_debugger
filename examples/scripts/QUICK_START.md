# Quick Start Guide - WASM Compilation

## Проверка доступных компиляторов

```bash
curl http://localhost:9092/_api/v1/scripts/compilers | jq .
```

Ответ:
```json
{
  "compilers": ["rust"],
  "all": {
    "assemblyscript": false,
    "go": false,
    "rust": true
  }
}
```

## Создание скрипта с Source Code

### 1. Создать скрипт

```bash
curl -X POST http://localhost:9092/_api/v1/scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Simple Test",
    "runtime": "extism",
    "language": "rust",
    "sourceCode": "use extism_pdk::*;\nuse serde::{Deserialize, Serialize};\nuse std::collections::HashMap;\n\n#[derive(Deserialize, Serialize)]\nstruct HTTPRequest {\n    method: String,\n    url: String,\n    headers: HashMap<String, Vec<String>>,\n    body: Vec<u8>,\n}\n\n#[plugin_fn]\npub fn process(Json(req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {\n    Ok(Json(req))\n}",
    "triggerType": "request",
    "priority": 10,
    "enabled": false
  }'
```

Ответ вернет ID скрипта, например: `{"id":"abc-123-def",...}`

### 2. Скомпилировать скрипт

```bash
curl -X POST http://localhost:9092/_api/v1/scripts/abc-123-def/compile \
  -H "Content-Type: application/json" \
  -d '{"optimize": true}'
```

Ответ:
```json
{
  "status": "success",
  "wasmSize": 145678,
  "compilationTime": "12.5s",
  "logs": ["Compiling script...", "Finished release [optimized] target(s) in 12.34s"]
}
```

### 3. Включить скрипт

```bash
curl -X PATCH http://localhost:9092/_api/v1/scripts/abc-123-def/toggle \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}'
```

### 4. Тестирование

Отправьте HTTP запрос через proxy:

```bash
curl -x http://localhost:9091 http://httpbin.org/get
```

## Использование готового примера

```bash
# Создать из файла
curl -X POST http://localhost:9092/_api/v1/scripts \
  -H "Content-Type: application/json" \
  -d @examples/scripts/rust/add_header_example.json

# Получить ID из ответа и скомпилировать
SCRIPT_ID="<полученный-id>"
curl -X POST http://localhost:9092/_api/v1/scripts/$SCRIPT_ID/compile \
  -d '{"optimize": true}'

# Включить
curl -X PATCH http://localhost:9092/_api/v1/scripts/$SCRIPT_ID/toggle \
  -d '{"enabled": true}'

# Тестировать
curl -x http://localhost:9091 http://httpbin.org/get
# Проверить заголовки X-Script-Processed, X-Compiled-With
```

## Зависимости (Cargo.toml)

Можно указать dependencies:

```json
{
  "name": "With Dependencies",
  "language": "rust",
  "sourceCode": "...",
  "dependencies": {
    "Cargo.toml": "[package]\nname = \"script\"\nversion = \"0.1.0\"\nedition = \"2021\"\n\n[lib]\ncrate-type = [\"cdylib\"]\n\n[dependencies]\nextism-pdk = \"1.0\"\nserde = { version = \"1.0\", features = [\"derive\"] }\nserde_json = \"1.0\"\nregex = \"1.10\""
  }
}
```

## Валидация синтаксиса (без компиляции)

```bash
curl -X POST http://localhost:9092/_api/v1/scripts/validate \
  -H "Content-Type: application/json" \
  -d '{
    "language": "rust",
    "sourceCode": "fn main() { println!(\"test\"); }"
  }'
```

## Troubleshooting

### Rust не доступен

Установить Rust toolchain:
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup target add wasm32-unknown-unknown
```

### TinyGo не доступен

Установить TinyGo:
```bash
# macOS
brew install tinygo

# Linux
wget https://github.com/tinygo-org/tinygo/releases/download/v0.31.0/tinygo_0.31.0_amd64.deb
sudo dpkg -i tinygo_0.31.0_amd64.deb
```

### AssemblyScript не доступен

```bash
npm install -g assemblyscript
```
