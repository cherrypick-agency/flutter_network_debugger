# Privileged Helper Tool для Process Detection

## Что это такое?

Privileged Helper Tool - это фоновый daemon, который работает с правами администратора (root) и позволяет **детектировать ВСЕ процессы** в системе, а не только процессы текущего пользователя.

### Зачем нужен?

**Без Helper Tool (unprivileged режим):**
- ✅ Детектирует процессы вашего пользователя (Safari, Chrome, curl и т.д.)
- ❌ НЕ видит процессы других пользователей
- ❌ НЕ видит системные процессы

**С Helper Tool (privileged режим):**
- ✅ Детектирует **ВСЕ процессы** в системе
- ✅ Полная детекция как в Proxyman
- ✅ Извлечение иконок с правами root

## Установка

### Вариант 1: Через UI (рекомендуется)

1. Откройте приложение
2. Перейдите в **Settings → Process Detection**
3. Нажмите кнопку **"Install Helper Tool"**
4. Введите пароль администратора в диалоговом окне macOS
5. Готово! Helper tool установлен и запущен

### Вариант 2: Через API

```bash
curl -X POST http://localhost:9092/_api/v1/process/helper/install
```

После выполнения команды:
- Откроется диалоговое окно с запросом пароля администратора
- Введите пароль
- Helper tool будет установлен автоматически

### Что происходит при установке?

1. Helper binary копируется в `/Library/PrivilegedHelperTools/com.networkdebugger.helper`
2. LaunchDaemon plist создается в `/Library/LaunchDaemons/com.networkdebugger.helper.plist`
3. Daemon автоматически запускается через `launchctl`
4. Unix socket создается в `/var/run/network-debugger-helper.sock`

## Проверка статуса

### Через API:

```bash
curl http://localhost:9092/_api/v1/process/helper/status
```

**Ответ:**
```json
{
  "installed": true,
  "running": true,
  "version": "1.0.0"
}
```

### Через командную строку:

```bash
# Проверить что daemon загружен
sudo launchctl list | grep networkdebugger

# Проверить что socket существует
ls -la /var/run/network-debugger-helper.sock

# Проверить логи
tail -f /var/log/network-debugger-helper.log
```

## Как это работает?

```
┌─────────────────┐
│ Main App        │
│ (без sudo)      │
└────────┬────────┘
         │ IPC через Unix socket
         │
         ▼
┌─────────────────┐
│ Helper Daemon   │
│ (с root правами)│
└────────┬────────┘
         │ Выполняет lsof, fileicon
         │
         ▼
┌─────────────────┐
│ Детекция ВСЕХ   │
│ процессов       │
└─────────────────┘
```

**Fallback стратегия:**
1. Попытка через helper (если установлен и работает)
2. Fallback на local detector (unprivileged)
3. Fallback на "Unknown Process" (если enabled)

## Удаление

### Через командную строку:

```bash
# 1. Остановить и выгрузить daemon
sudo launchctl unload /Library/LaunchDaemons/com.networkdebugger.helper.plist

# 2. Удалить файлы
sudo rm /Library/LaunchDaemons/com.networkdebugger.helper.plist
sudo rm /Library/PrivilegedHelperTools/com.networkdebugger.helper

# 3. Удалить socket (если остался)
sudo rm /var/run/network-debugger-helper.sock
```

### Через API (будет добавлено позже):

```bash
curl -X POST http://localhost:9092/_api/v1/process/helper/uninstall
```

## Troubleshooting

### Helper не запускается

**Проверить логи:**
```bash
tail -f /var/log/network-debugger-helper.log
tail -f /var/log/network-debugger-helper.err
```

**Проверить что binary существует:**
```bash
ls -la /Library/PrivilegedHelperTools/com.networkdebugger.helper
```

**Перезапустить daemon:**
```bash
sudo launchctl kickstart -k system/com.networkdebugger.helper
```

### Socket не создается

**Проверить права:**
```bash
ls -la /var/run/network-debugger-helper.sock
# Должен быть: srw------- (0600)
```

**Пересоздать socket:**
```bash
# Остановить daemon
sudo launchctl stop com.networkdebugger.helper

# Удалить старый socket
sudo rm /var/run/network-debugger-helper.sock

# Запустить daemon снова
sudo launchctl start com.networkdebugger.helper
```

### Детекция не работает

**Проверить что helper запущен:**
```bash
sudo launchctl list | grep networkdebugger
# Должен показать PID и label
```

**Проверить конфигурацию:**
```bash
curl http://localhost:9092/_api/v1/process/config
# useHelperTool должен быть true
```

**Включить helper tool в настройках:**
```bash
curl -X POST http://localhost:9092/_api/v1/process/config \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "useHelperTool": true,
    "cacheEnabled": true,
    "cacheTtl": 300,
    "fallbackEnabled": true
  }'
```

### Протестировать IPC вручную

Создайте тестовый скрипт:

```bash
# test-helper-ipc.sh
echo '{"id":"test-1","method":"ping"}' | \
  nc -U /var/run/network-debugger-helper.sock

echo '{"id":"test-2","method":"detect","params":{"port":8080}}' | \
  nc -U /var/run/network-debugger-helper.sock
```

Запустите:
```bash
chmod +x test-helper-ipc.sh
./test-helper-ipc.sh
```

**Ожидаемый ответ:**
```json
{"id":"test-1","result":{"pong":"pong"}}
{"id":"test-2","result":{"processInfo":{"PID":12345,"Name":"curl",...}}}
```

## Безопасность

### Что делает helper?

Helper daemon **только читает** информацию о процессах:
- ✅ Выполняет `lsof` для детекции процесса по порту
- ✅ Выполняет `fileicon` и `sips` для извлечения иконок
- ❌ **НЕ модифицирует** систему
- ❌ **НЕ имеет** сетевого доступа (только Unix socket localhost)
- ❌ **НЕ записывает** файлы (кроме логов)

### Permissions

- **Socket:** 0600 (только owner может читать/писать)
- **Binary:** 0755 (выполнение разрешено всем, запись только root)
- **Plist:** 0644 (чтение всем, запись только root)
- **Logs:** только root может читать

### IPC Protocol

- **Транспорт:** Unix socket (не сетевой, только localhost)
- **Формат:** JSON-RPC style
- **Валидация:** все входящие данные проверяются
- **Timeout:** 10 секунд на каждый request

## Продвинутое использование

### Посмотреть IPC трафик

```bash
# Установить socat если еще нет
brew install socat

# Прослушивать Unix socket
sudo socat -v UNIX-LISTEN:/tmp/debug.sock,fork UNIX-CONNECT:/var/run/network-debugger-helper.sock &

# Обновить socket path в конфиге (временно)
# Теперь все IPC будет логироваться
```

### Запустить helper вручную для отладки

```bash
# Остановить launchd daemon
sudo launchctl unload /Library/LaunchDaemons/com.networkdebugger.helper.plist

# Запустить вручную с логами в терминал
sudo /Library/PrivilegedHelperTools/com.networkdebugger.helper

# После отладки вернуть обратно
sudo launchctl load -w /Library/LaunchDaemons/com.networkdebugger.helper.plist
```

### Кастомный helper binary (для разработки)

```bash
# Собрать helper
go build -o bin/process-helper ./cmd/process-helper

# Установить свою версию
sudo cp bin/process-helper /Library/PrivilegedHelperTools/com.networkdebugger.helper
sudo launchctl kickstart -k system/com.networkdebugger.helper
```

## FAQ

**Q: Нужно ли устанавливать helper tool?**
A: Нет, приложение работает и без него. Helper нужен только для детекции процессов других пользователей.

**Q: Безопасно ли давать приложению права администратора?**
A: Да, helper tool только читает процессы и не модифицирует систему. Код открытый, можете проверить.

**Q: Можно ли удалить helper после установки?**
A: Да, следуйте инструкциям в разделе "Удаление".

**Q: Работает ли на Windows/Linux?**
A: Пока только macOS. Планируется поддержка других платформ.

**Q: Helper запускается при загрузке системы?**
A: Да, launchd автоматически запускает helper при загрузке macOS (RunAtLoad=true).

**Q: Сколько памяти потребляет helper?**
A: ~5-10 MB в idle, ~20 MB при активной детекции.

**Q: Могу ли я использовать несколько экземпляров приложения?**
A: Да, но все будут использовать один и тот же helper daemon.

## См. также

- [Process Detection Architecture](./plans/process-detection-implementation.md)
- [Main README](../README.md)
