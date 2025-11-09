# Process Detection & Application Icons - Implementation Plan

**Feature**: Детекция процессов/приложений для сетевых соединений с иконками и поддержкой привилегированного helper tool

**Date**: 2025-11-02
**Estimated Duration**: 5 недель
**Architecture**: Clean Architecture + SOLID + DRY

---

## 🎯 Цели и Требования

### Функциональные требования:
1. Определять, какое приложение создало сетевое соединение (по порту)
2. Извлекать и отображать иконку приложения
3. Показывать имя процесса и путь к executable
4. Поддерживать работу БЕЗ админских прав (graceful degradation)
5. Опциональный privileged helper tool для полной детекции
6. Настройка через UI (включение/выключение, helper tool)
7. Кеширование иконок для производительности
8. Поддержка macOS, Windows, Linux

### Нефункциональные требования:
1. Clean Architecture (domain/usecase/infrastructure)
2. SOLID принципы
3. DRY (без дублирования кода)
4. Тестируемость (unit + integration tests)
5. Производительность (не замедлять proxy)
6. Безопасность (запрос пароля только при установке helper)

---

## 🏗️ Архитектура

### Диаграмма компонентов:

```
┌─────────────────────────────────────────────────────────┐
│                   Flutter Desktop UI                     │
│  ┌────────────────┐  ┌──────────────────────────────┐  │
│  │ Settings Page  │  │ Session List (with icons)    │  │
│  └────────────────┘  └──────────────────────────────┘  │
└───────────────────────────┬─────────────────────────────┘
                            │ HTTP/WebSocket API
┌───────────────────────────▼─────────────────────────────┐
│              Go Backend (Clean Architecture)             │
│                                                           │
│  ┌────────────────────────────────────────────────────┐ │
│  │           HTTP API Layer (httpapi)                 │ │
│  │  - process_handlers.go                             │ │
│  │  - DTOs & routing                                  │ │
│  └───────────────────────┬────────────────────────────┘ │
│                          │                               │
│  ┌───────────────────────▼────────────────────────────┐ │
│  │         Use Case Layer (usecase/service.go)        │ │
│  │  - ProcessService (orchestration)                  │ │
│  │  - DetectForConnection()                           │ │
│  │  - InstallHelper()                                 │ │
│  └──┬────────────┬──────────────┬────────────────────┘ │
│     │            │              │                        │
│  ┌──▼──────┐  ┌─▼────────┐  ┌──▼──────────────────┐   │
│  │ Domain  │  │ Domain   │  │ Domain              │   │
│  │ Repos   │  │ Detector │  │ IconExtractor       │   │
│  │ (ports) │  │ (port)   │  │ (port)              │   │
│  └──┬──────┘  └─┬────────┘  └──┬──────────────────┘   │
│     │            │              │                        │
│  ┌──▼────────────▼──────────────▼──────────────────┐   │
│  │    Infrastructure Layer (adapters)              │   │
│  │                                                  │   │
│  │  ┌──────────────┐  ┌─────────────────────────┐ │   │
│  │  │ Persistence  │  │ Detector Implementations│ │   │
│  │  │ - GORM/SQLite│  │ - detector_darwin.go    │ │   │
│  │  │ - Repos      │  │ - detector_windows.go   │ │   │
│  │  │ - Icon Cache │  │ - detector_linux.go     │ │   │
│  │  └──────────────┘  └─────────────────────────┘ │   │
│  │                                                  │   │
│  │  ┌──────────────────┐  ┌────────────────────┐  │   │
│  │  │ Icon Extractors  │  │ Helper Client      │  │   │
│  │  │ - macOS/Win/Lin  │  │ - IPC comm         │  │   │
│  │  └──────────────────┘  └────────────────────┘  │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                          │ IPC (Unix socket / Named pipe)
┌─────────────────────────▼─────────────────────────────┐
│           Privileged Helper Tool (Optional)            │
│         cmd/process-helper/ (separate binary)          │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │  IPC Server (JSON-RPC over Unix socket/pipe)    │ │
│  │  - Accept connections from main app              │ │
│  │  - Handle: detect, icon, ping requests           │ │
│  └───────────────────────┬──────────────────────────┘ │
│                          │                             │
│  ┌───────────────────────▼──────────────────────────┐ │
│  │  Privileged Detector (root/admin)                │ │
│  │  - Full system access                            │ │
│  │  - Detect ALL processes                          │ │
│  │  - Extract ALL icons                             │ │
│  └──────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Структура директорий:

```
go-proxy/
├── internal/features/process/        # NEW: Process detection feature
│   ├── domain/
│   │   ├── entities.go               # ProcessInfo, AppIcon, DetectionConfig
│   │   ├── detector.go               # IProcessDetector interface
│   │   ├── icon_extractor.go         # IIconExtractor interface
│   │   ├── helper.go                 # IHelperClient interface
│   │   └── repository.go             # IConfigRepository, IIconCacheRepository
│   │
│   ├── usecase/
│   │   └── service.go                # ProcessService (business logic)
│   │
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── models.go             # GORM models
│   │   │   ├── config_repo.go        # Config repository impl
│   │   │   └── icon_cache_repo.go    # Icon cache repository impl
│   │   │
│   │   ├── detector/
│   │   │   ├── detector.go           # Factory + base
│   │   │   ├── detector_darwin.go    # macOS impl (lsof)
│   │   │   ├── detector_windows.go   # Windows impl (gopsutil)
│   │   │   ├── detector_linux.go     # Linux impl (/proc)
│   │   │   └── gopsutil_adapter.go   # gopsutil wrapper
│   │   │
│   │   ├── icon/
│   │   │   ├── extractor.go          # Factory + base
│   │   │   ├── extractor_darwin.go   # macOS (fileicon/sips)
│   │   │   ├── extractor_windows.go  # Windows (lxn/win)
│   │   │   └── extractor_linux.go    # Linux (freedesktop)
│   │   │
│   │   └── helper/
│   │       ├── client.go             # Helper IPC client
│   │       ├── client_darwin.go      # Unix socket (macOS)
│   │       ├── client_windows.go     # Named pipe (Windows)
│   │       └── installer.go          # Helper installation logic
│   │
│   └── runtime/
│       └── manager.go                # Cache manager, hot-reload
│
├── cmd/process-helper/               # NEW: Privileged helper tool
│   ├── main.go                       # Helper daemon entry point
│   ├── server/
│   │   ├── server.go                 # IPC server
│   │   ├── detector.go               # Root-level detector
│   │   └── handler.go                # Request handlers
│   └── ipc/
│       ├── protocol.go               # IPC protocol (JSON-RPC)
│       └── transport.go              # Unix socket/named pipe
│
├── migrations/
│   └── 0004_process_detection.sql    # NEW: Database schema
│
├── frontend/lib/features/
│   └── settings/
│       ├── application/
│       │   └── process_service.dart  # NEW: Process API service
│       └── presentation/
│           └── process_page.dart     # NEW: Process settings UI
│
└── docs/plans/
    └── process-detection-implementation.md  # THIS FILE
```

---

## 📐 SOLID Принципы в Дизайне

### Single Responsibility Principle (SRP):
- **ProcessService**: Только orchestration бизнес-логики
- **Detector**: Только детекция процессов
- **IconExtractor**: Только извлечение иконок
- **HelperClient**: Только IPC коммуникация
- **Repository**: Только доступ к данным

### Open/Closed Principle (OCP):
- Новые платформы: добавить `detector_freebsd.go` (не трогать существующие)
- Новые методы детекции: создать новую impl `IProcessDetector`
- Расширение без модификации существующего кода

### Liskov Substitution Principle (LSP):
- Любой `IProcessDetector` работает одинаково (contract)
- Helper и Local detector взаимозаменяемы
- Можно подменить реализацию без поломки логики

### Interface Segregation Principle (ISP):
- `IProcessDetector` - только детекция
- `IIconExtractor` - только иконки
- `IHelperClient` - только IPC
- `IConfigRepository` - только config CRUD
- `IIconCacheRepository` - только cache CRUD
- Не один "толстый" интерфейс

### Dependency Inversion Principle (DIP):
- `ProcessService` зависит от интерфейсов (domain layer)
- Implementations в infrastructure layer
- Легко создавать моки для unit tests
- Dependency injection через конструкторы

---

## 🔧 Технические Детали

### 1. Domain Layer

#### Entities (`internal/features/process/domain/entities.go`):

```go
package domain

import "time"

// ProcessInfo - информация о процессе
type ProcessInfo struct {
    PID            int32
    Name           string
    ExecutablePath string
    BundleID       *string  // macOS only (e.g., "com.apple.Safari")
    Icon           *AppIcon
    DetectedAt     time.Time
}

// AppIcon - иконка приложения
type AppIcon struct {
    Format string  // "png", "icns", "ico"
    Data   []byte  // binary data
    Path   *string // cached file path (optional)
}

// DetectionConfig - конфигурация детекции
type DetectionConfig struct {
    ID                int64
    Enabled           bool      // включена ли детекция
    UseHelperTool     bool      // использовать helper tool
    HelperInstalled   bool      // установлен ли helper
    CacheEnabled      bool      // кешировать иконки
    CacheTTLSeconds   int       // TTL кеша в секундах
    FallbackEnabled   bool      // показывать "Unknown" при ошибке
    UpdatedAt         time.Time
}

// Validation methods
func (c *DetectionConfig) Validate() error {
    if c.CacheTTLSeconds < 0 {
        return errors.New("cache TTL must be non-negative")
    }
    return nil
}
```

#### Interfaces (`internal/features/process/domain/detector.go`):

```go
package domain

import "context"

// IProcessDetector - абстракция детекции процессов
// Реализации для разных OS в infrastructure layer
type IProcessDetector interface {
    // DetectByPort - найти процесс по локальному порту
    DetectByPort(ctx context.Context, port uint32) (*ProcessInfo, error)

    // DetectByPID - получить информацию о процессе по PID
    DetectByPID(ctx context.Context, pid int32) (*ProcessInfo, error)

    // RequiresPrivileges - требуются ли root/admin права
    RequiresPrivileges() bool
}

// IIconExtractor - абстракция извлечения иконок
type IIconExtractor interface {
    // ExtractByPID - извлечь иконку по PID процесса
    ExtractByPID(ctx context.Context, pid int32) (*AppIcon, error)

    // ExtractByPath - извлечь иконку по пути к приложению
    ExtractByPath(ctx context.Context, path string) (*AppIcon, error)
}

// IHelperClient - коммуникация с helper tool
type IHelperClient interface {
    // IsRunning - проверить, запущен ли helper
    IsRunning() bool

    // DetectProcess - детекция через helper (privileged)
    DetectProcess(port uint32) (*ProcessInfo, error)

    // ExtractIcon - извлечение иконки через helper
    ExtractIcon(pid int32) (*AppIcon, error)

    // Ping - проверка доступности helper
    Ping() error

    // Close - закрыть соединение
    Close() error
}
```

#### Repository Interfaces (`internal/features/process/domain/repository.go`):

```go
package domain

import "context"

// IConfigRepository - персистентность конфигурации
type IConfigRepository interface {
    Load(ctx context.Context) (*DetectionConfig, error)
    Save(ctx context.Context, cfg *DetectionConfig) error
}

// IIconCacheRepository - кеш иконок
type IIconCacheRepository interface {
    Get(key string) (*AppIcon, error)
    Set(key string, icon *AppIcon, ttl time.Duration) error
    Delete(key string) error
    Clear() error
}
```

---

### 2. Use Case Layer

#### Service (`internal/features/process/usecase/service.go`):

```go
package usecase

import (
    "context"
    "fmt"
    "time"

    "github.com/rs/zerolog"
    "github.com/belief/go-proxy/internal/features/process/domain"
)

type Service struct {
    config        domain.IConfigRepository
    iconCache     domain.IIconCacheRepository
    localDetector domain.IProcessDetector   // без привилегий
    helperClient  domain.IHelperClient      // с привилегиями
    iconExtractor domain.IIconExtractor
    logger        *zerolog.Logger
}

func NewService(
    config domain.IConfigRepository,
    iconCache domain.IIconCacheRepository,
    localDetector domain.IProcessDetector,
    helperClient domain.IHelperClient,
    iconExtractor domain.IIconExtractor,
    logger *zerolog.Logger,
) *Service {
    return &Service{
        config:        config,
        iconCache:     iconCache,
        localDetector: localDetector,
        helperClient:  helperClient,
        iconExtractor: iconExtractor,
        logger:        logger,
    }
}

// DetectForConnection - главный метод детекции
func (s *Service) DetectForConnection(ctx context.Context, localPort uint32) (*domain.ProcessInfo, error) {
    cfg, err := s.config.Load(ctx)
    if err != nil {
        return nil, err
    }

    if !cfg.Enabled {
        return nil, nil // детекция отключена
    }

    // Стратегия: пробуем helper → fallback на local → fallback на "Unknown"
    var info *domain.ProcessInfo

    // 1. Попытка через helper (если включен и доступен)
    if cfg.UseHelperTool && s.helperClient != nil && s.helperClient.IsRunning() {
        info, err = s.helperClient.DetectProcess(localPort)
        if err != nil {
            s.logger.Warn().Err(err).Msg("Helper detection failed, falling back to local")
        }
    }

    // 2. Fallback на локальную детекцию
    if info == nil {
        info, err = s.localDetector.DetectByPort(ctx, localPort)
        if err != nil {
            if cfg.FallbackEnabled {
                // Вернуть "Unknown Process"
                return &domain.ProcessInfo{
                    Name:       "Unknown Process",
                    DetectedAt: time.Now(),
                }, nil
            }
            return nil, err
        }
    }

    // 3. Получить иконку
    if info != nil {
        icon, err := s.getIcon(ctx, info.PID, info.ExecutablePath)
        if err != nil {
            s.logger.Debug().Err(err).Int32("pid", info.PID).Msg("Failed to get icon")
        }
        info.Icon = icon
    }

    return info, nil
}

// getIcon - получение иконки с кешированием
func (s *Service) getIcon(ctx context.Context, pid int32, path string) (*domain.AppIcon, error) {
    cfg, _ := s.config.Load(ctx)

    if !cfg.CacheEnabled {
        return s.extractIcon(ctx, pid, path)
    }

    // Проверить кеш
    key := fmt.Sprintf("icon:%d:%s", pid, path)
    if cached, err := s.iconCache.Get(key); err == nil {
        return cached, nil
    }

    // Извлечь иконку
    icon, err := s.extractIcon(ctx, pid, path)
    if err != nil {
        return nil, err
    }

    // Сохранить в кеш
    ttl := time.Duration(cfg.CacheTTLSeconds) * time.Second
    s.iconCache.Set(key, icon, ttl)

    return icon, nil
}

// extractIcon - извлечение иконки (helper или local)
func (s *Service) extractIcon(ctx context.Context, pid int32, path string) (*domain.AppIcon, error) {
    cfg, _ := s.config.Load(ctx)

    // Попытка через helper
    if cfg.UseHelperTool && s.helperClient != nil && s.helperClient.IsRunning() {
        icon, err := s.helperClient.ExtractIcon(pid)
        if err == nil {
            return icon, nil
        }
        s.logger.Debug().Err(err).Msg("Helper icon extraction failed, trying local")
    }

    // Fallback на локальное извлечение
    return s.iconExtractor.ExtractByPID(ctx, pid)
}

// GetConfig - получить текущую конфигурацию
func (s *Service) GetConfig(ctx context.Context) (*domain.DetectionConfig, error) {
    return s.config.Load(ctx)
}

// SaveConfig - сохранить конфигурацию
func (s *Service) SaveConfig(ctx context.Context, cfg *domain.DetectionConfig) error {
    if err := cfg.Validate(); err != nil {
        return err
    }
    return s.config.Save(ctx, cfg)
}

// CheckHelperStatus - проверить статус helper tool
func (s *Service) CheckHelperStatus() HelperStatus {
    if s.helperClient == nil {
        return HelperStatus{Installed: false, Running: false}
    }

    running := s.helperClient.IsRunning()
    return HelperStatus{
        Installed: true, // TODO: проверить установку
        Running:   running,
    }
}

// InstallHelper - установить helper tool
func (s *Service) InstallHelper(ctx context.Context) error {
    // Делегировать installer в infrastructure
    return installHelperTool()
}

type HelperStatus struct {
    Installed bool
    Running   bool
    Version   string
}
```

---

### 3. Infrastructure - Detector Implementations

#### Factory (`internal/features/process/infrastructure/detector/detector.go`):

```go
package detector

import (
    "fmt"
    "runtime"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

// NewDetector - создать детектор для текущей платформы
func NewDetector(privileged bool) (domain.IProcessDetector, error) {
    if privileged {
        return newPrivilegedDetector()
    }
    return newUnprivilegedDetector()
}

func newUnprivilegedDetector() (domain.IProcessDetector, error) {
    switch runtime.GOOS {
    case "darwin":
        return &darwinDetector{}, nil
    case "windows":
        return &windowsDetector{}, nil
    case "linux":
        return &linuxDetector{}, nil
    default:
        return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
}

func newPrivilegedDetector() (domain.IProcessDetector, error) {
    // Privileged детекция работает аналогично, но с sudo/root
    return newUnprivilegedDetector()
}
```

#### macOS Implementation (`internal/features/process/infrastructure/detector/detector_darwin.go`):

```go
//go:build darwin

package detector

import (
    "context"
    "fmt"
    "os/exec"
    "strconv"
    "strings"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

type darwinDetector struct{}

func (d *darwinDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
    // lsof -i :PORT -P -n -F (field output mode)
    cmd := exec.CommandContext(
        ctx,
        "lsof",
        "-i", fmt.Sprintf(":%d", port),
        "-P",  // numeric ports
        "-n",  // numeric IPs
        "-F", "pcn",  // fields: PID, command, name
    )

    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("lsof failed: %w", err)
    }

    // Парсинг вывода lsof
    // Format:
    // p<PID>
    // c<command>
    // n<name>
    info, err := parseLsofOutput(string(output))
    if err != nil {
        return nil, err
    }

    return info, nil
}

func (d *darwinDetector) DetectByPID(ctx context.Context, pid int32) (*domain.ProcessInfo, error) {
    // ps -p PID -o comm=,args=
    cmd := exec.CommandContext(ctx, "ps", "-p", fmt.Sprint(pid), "-o", "comm=,args=")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("ps failed: %w", err)
    }

    parts := strings.Fields(string(output))
    if len(parts) == 0 {
        return nil, fmt.Errorf("process not found")
    }

    return &domain.ProcessInfo{
        PID:            pid,
        Name:           parts[0],
        ExecutablePath: parts[0],
        DetectedAt:     time.Now(),
    }, nil
}

func (d *darwinDetector) RequiresPrivileges() bool {
    return false  // lsof работает для своих процессов
}

func parseLsofOutput(output string) (*domain.ProcessInfo, error) {
    lines := strings.Split(output, "\n")

    var pid int32
    var name, cmd string

    for _, line := range lines {
        if len(line) < 2 {
            continue
        }

        prefix := line[0]
        value := line[1:]

        switch prefix {
        case 'p':
            p, _ := strconv.Atoi(value)
            pid = int32(p)
        case 'c':
            cmd = value
        case 'n':
            name = value
        }
    }

    if pid == 0 {
        return nil, fmt.Errorf("no process found")
    }

    return &domain.ProcessInfo{
        PID:            pid,
        Name:           cmd,
        ExecutablePath: name,
        DetectedAt:     time.Now(),
    }, nil
}
```

#### Windows Implementation (`internal/features/process/infrastructure/detector/detector_windows.go`):

```go
//go:build windows

package detector

import (
    "context"
    "fmt"
    "time"

    "github.com/shirou/gopsutil/v4/net"
    "github.com/shirou/gopsutil/v4/process"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

type windowsDetector struct{}

func (w *windowsDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
    // gopsutil работает хорошо на Windows без admin
    conns, err := net.ConnectionsWithContext(ctx, "tcp")
    if err != nil {
        return nil, fmt.Errorf("failed to get connections: %w", err)
    }

    // Найти соединение по порту
    for _, conn := range conns {
        if conn.Laddr.Port == port {
            return w.DetectByPID(ctx, conn.Pid)
        }
    }

    return nil, fmt.Errorf("no process found for port %d", port)
}

func (w *windowsDetector) DetectByPID(ctx context.Context, pid int32) (*domain.ProcessInfo, error) {
    proc, err := process.NewProcessWithContext(ctx, pid)
    if err != nil {
        return nil, fmt.Errorf("process not found: %w", err)
    }

    name, _ := proc.Name()
    exe, _ := proc.Exe()

    return &domain.ProcessInfo{
        PID:            pid,
        Name:           name,
        ExecutablePath: exe,
        DetectedAt:     time.Now(),
    }, nil
}

func (w *windowsDetector) RequiresPrivileges() bool {
    return false  // Windows более permissive
}
```

#### Linux Implementation (`internal/features/process/infrastructure/detector/detector_linux.go`):

```go
//go:build linux

package detector

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

type linuxDetector struct{}

func (l *linuxDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
    // 1. Читать /proc/net/tcp
    tcpData, err := ioutil.ReadFile("/proc/net/tcp")
    if err != nil {
        return nil, fmt.Errorf("failed to read /proc/net/tcp: %w", err)
    }

    // 2. Найти inode по порту
    inode, err := findInodeByPort(string(tcpData), port)
    if err != nil {
        return nil, err
    }

    // 3. Найти PID по inode
    pid, err := findPIDByInode(inode)
    if err != nil {
        return nil, err
    }

    return l.DetectByPID(ctx, int32(pid))
}

func (l *linuxDetector) DetectByPID(ctx context.Context, pid int32) (*domain.ProcessInfo, error) {
    // Читать /proc/[pid]/comm для имени
    commPath := fmt.Sprintf("/proc/%d/comm", pid)
    commData, err := ioutil.ReadFile(commPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read comm: %w", err)
    }

    // Читать /proc/[pid]/exe для пути
    exePath := fmt.Sprintf("/proc/%d/exe", pid)
    exe, _ := os.Readlink(exePath)

    return &domain.ProcessInfo{
        PID:            pid,
        Name:           strings.TrimSpace(string(commData)),
        ExecutablePath: exe,
        DetectedAt:     time.Now(),
    }, nil
}

func (l *linuxDetector) RequiresPrivileges() bool {
    return false  // /proc/net/tcp доступен, но /proc/[pid]/fd может требовать прав
}

func findInodeByPort(tcpData string, port uint32) (uint64, error) {
    lines := strings.Split(tcpData, "\n")

    for _, line := range lines[1:] { // пропустить заголовок
        fields := strings.Fields(line)
        if len(fields) < 10 {
            continue
        }

        // local_address в hex: "0100007F:1F90" = 127.0.0.1:8080
        localAddr := fields[1]
        parts := strings.Split(localAddr, ":")
        if len(parts) != 2 {
            continue
        }

        // Парсинг hex порта
        portHex := parts[1]
        p, err := strconv.ParseUint(portHex, 16, 32)
        if err != nil {
            continue
        }

        if uint32(p) == port {
            // inode в 9-м поле
            inode, _ := strconv.ParseUint(fields[9], 10, 64)
            return inode, nil
        }
    }

    return 0, fmt.Errorf("port %d not found in /proc/net/tcp", port)
}

func findPIDByInode(inode uint64) (int, error) {
    // Перебрать все /proc/[pid]/fd/* в поисках socket:[inode]
    procDir := "/proc"
    entries, err := ioutil.ReadDir(procDir)
    if err != nil {
        return 0, err
    }

    target := fmt.Sprintf("socket:[%d]", inode)

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        pid, err := strconv.Atoi(entry.Name())
        if err != nil {
            continue
        }

        fdDir := filepath.Join(procDir, entry.Name(), "fd")
        fds, err := ioutil.ReadDir(fdDir)
        if err != nil {
            continue  // permission denied - skip
        }

        for _, fd := range fds {
            linkPath := filepath.Join(fdDir, fd.Name())
            link, err := os.Readlink(linkPath)
            if err != nil {
                continue
            }

            if link == target {
                return pid, nil
            }
        }
    }

    return 0, fmt.Errorf("no PID found for inode %d", inode)
}
```

---

### 4. Infrastructure - Icon Extractors

#### Factory (`internal/features/process/infrastructure/icon/extractor.go`):

```go
package icon

import (
    "fmt"
    "runtime"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

func NewExtractor() (domain.IIconExtractor, error) {
    switch runtime.GOOS {
    case "darwin":
        return &darwinExtractor{}, nil
    case "windows":
        return &windowsExtractor{}, nil
    case "linux":
        return &linuxExtractor{}, nil
    default:
        return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
}
```

#### macOS Icon Extractor (`internal/features/process/infrastructure/icon/extractor_darwin.go`):

```go
//go:build darwin

package icon

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

type darwinExtractor struct {
    cachedir string
}

func (e *darwinExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
    // Получить путь к приложению по PID
    cmd := exec.CommandContext(ctx, "ps", "-p", fmt.Sprint(pid), "-o", "comm=")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get process path: %w", err)
    }

    path := strings.TrimSpace(string(output))
    return e.ExtractByPath(ctx, path)
}

func (e *darwinExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
    // 1. Найти .app bundle
    appPath := findAppBundle(path)
    if appPath == "" {
        return nil, fmt.Errorf("not an application bundle")
    }

    // 2. Создать временную директорию
    tmpDir, err := ioutil.TempDir("", "icons")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tmpDir)

    icnsPath := filepath.Join(tmpDir, "icon.icns")
    pngPath := filepath.Join(tmpDir, "icon.png")

    // 3. Извлечь .icns с помощью fileicon
    cmd := exec.CommandContext(ctx, "fileicon", "get", appPath, "--output", icnsPath)
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("fileicon failed: %w", err)
    }

    // 4. Конвертировать в PNG с помощью sips (встроенная утилита macOS)
    cmd = exec.CommandContext(ctx, "sips", "-s", "format", "png", icnsPath, "--out", pngPath)
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("sips failed: %w", err)
    }

    // 5. Прочитать PNG
    data, err := ioutil.ReadFile(pngPath)
    if err != nil {
        return nil, err
    }

    return &domain.AppIcon{
        Format: "png",
        Data:   data,
    }, nil
}

func findAppBundle(path string) string {
    // Ищем .app в пути
    for {
        if strings.HasSuffix(path, ".app") {
            return path
        }

        parent := filepath.Dir(path)
        if parent == path || parent == "/" {
            break
        }
        path = parent
    }
    return ""
}
```

#### Windows Icon Extractor (`internal/features/process/infrastructure/icon/extractor_windows.go`):

```go
//go:build windows

package icon

import (
    "context"
    "fmt"
    "os/exec"

    "github.com/shirou/gopsutil/v4/process"

    "github.com/belief/go-proxy/internal/features/process/domain"
    // TODO: использовать github.com/lxn/win для ExtractIconEx
)

type windowsExtractor struct{}

func (e *windowsExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
    proc, err := process.NewProcessWithContext(ctx, pid)
    if err != nil {
        return nil, err
    }

    exe, err := proc.Exe()
    if err != nil {
        return nil, err
    }

    return e.ExtractByPath(ctx, exe)
}

func (e *windowsExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
    // TODO: Реализовать через lxn/win + ExtractIconEx
    // Пока заглушка
    return nil, fmt.Errorf("not implemented yet")
}
```

#### Linux Icon Extractor (`internal/features/process/infrastructure/icon/extractor_linux.go`):

```go
//go:build linux

package icon

import (
    "context"
    "fmt"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

type linuxExtractor struct{}

func (e *linuxExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
    // TODO: Реализовать через .desktop файлы
    return nil, fmt.Errorf("not implemented yet")
}

func (e *linuxExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
    // TODO: Реализовать через freedesktop spec
    return nil, fmt.Errorf("not implemented yet")
}
```

---

### 5. Helper Tool (Privileged Daemon)

#### Main (`cmd/process-helper/main.go`):

```go
package main

import (
    "log"
    "net"
    "os"
    "runtime"

    "github.com/belief/go-proxy/cmd/process-helper/server"
)

func main() {
    // Проверка привилегий
    if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
        if os.Geteuid() != 0 {
            log.Fatal("Helper must run as root")
        }
    }

    // Создать IPC listener
    listener, err := createListener()
    if err != nil {
        log.Fatal("Failed to create listener:", err)
    }
    defer listener.Close()

    log.Println("Process helper started")

    // Запустить сервер
    srv := server.NewServer()
    if err := srv.Serve(listener); err != nil {
        log.Fatal("Server error:", err)
    }
}

func createListener() (net.Listener, error) {
    switch runtime.GOOS {
    case "darwin", "linux":
        // Unix socket
        sockPath := "/var/run/network-debugger-helper.sock"
        os.Remove(sockPath) // cleanup
        return net.Listen("unix", sockPath)

    case "windows":
        // Named pipe
        // TODO: использовать github.com/Microsoft/go-winio
        return nil, fmt.Errorf("Windows not implemented yet")

    default:
        return nil, fmt.Errorf("unsupported platform")
    }
}
```

#### IPC Protocol (`cmd/process-helper/ipc/protocol.go`):

```go
package ipc

type Request struct {
    ID     string        `json:"id"`
    Method string        `json:"method"`  // "detect", "icon", "ping"
    Params RequestParams `json:"params"`
}

type RequestParams struct {
    Port uint32 `json:"port,omitempty"`
    PID  int32  `json:"pid,omitempty"`
    Path string `json:"path,omitempty"`
}

type Response struct {
    ID     string      `json:"id"`
    Result interface{} `json:"result,omitempty"`
    Error  *Error      `json:"error,omitempty"`
}

type Error struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}
```

#### Server (`cmd/process-helper/server/server.go`):

```go
package server

import (
    "encoding/json"
    "net"

    "github.com/belief/go-proxy/cmd/process-helper/ipc"
    "github.com/belief/go-proxy/internal/features/process/domain"
    "github.com/belief/go-proxy/internal/features/process/infrastructure/detector"
    "github.com/belief/go-proxy/internal/features/process/infrastructure/icon"
)

type Server struct {
    detector  domain.IProcessDetector
    extractor domain.IIconExtractor
}

func NewServer() *Server {
    det, _ := detector.NewDetector(true)  // privileged
    ext, _ := icon.NewExtractor()

    return &Server{
        detector:  det,
        extractor: ext,
    }
}

func (s *Server) Serve(listener net.Listener) error {
    for {
        conn, err := listener.Accept()
        if err != nil {
            return err
        }

        go s.handleConnection(conn)
    }
}

func (s *Server) handleConnection(conn net.Conn) {
    defer conn.Close()

    dec := json.NewDecoder(conn)
    enc := json.NewEncoder(conn)

    var req ipc.Request
    if err := dec.Decode(&req); err != nil {
        return
    }

    resp := s.handleRequest(req)
    enc.Encode(resp)
}

func (s *Server) handleRequest(req ipc.Request) ipc.Response {
    switch req.Method {
    case "detect":
        info, err := s.detector.DetectByPort(context.Background(), req.Params.Port)
        if err != nil {
            return ipc.Response{
                ID:    req.ID,
                Error: &ipc.Error{Code: -1, Message: err.Error()},
            }
        }
        return ipc.Response{ID: req.ID, Result: info}

    case "icon":
        icon, err := s.extractor.ExtractByPID(context.Background(), req.Params.PID)
        if err != nil {
            return ipc.Response{
                ID:    req.ID,
                Error: &ipc.Error{Code: -1, Message: err.Error()},
            }
        }
        return ipc.Response{ID: req.ID, Result: icon}

    case "ping":
        return ipc.Response{ID: req.ID, Result: "pong"}

    default:
        return ipc.Response{
            ID:    req.ID,
            Error: &ipc.Error{Code: -32601, Message: "Method not found"},
        }
    }
}
```

#### Helper Client (`internal/features/process/infrastructure/helper/client.go`):

```go
package helper

import (
    "encoding/json"
    "fmt"
    "net"
    "runtime"
    "sync"

    "github.com/belief/go-proxy/cmd/process-helper/ipc"
    "github.com/belief/go-proxy/internal/features/process/domain"
)

type Client struct {
    conn net.Conn
    mu   sync.Mutex
    seq  uint64
}

func NewClient() (*Client, error) {
    conn, err := dialHelper()
    if err != nil {
        return nil, err
    }

    return &Client{conn: conn}, nil
}

func dialHelper() (net.Conn, error) {
    switch runtime.GOOS {
    case "darwin", "linux":
        return net.Dial("unix", "/var/run/network-debugger-helper.sock")
    case "windows":
        // TODO: winio.DialPipe
        return nil, fmt.Errorf("Windows not implemented")
    default:
        return nil, fmt.Errorf("unsupported platform")
    }
}

func (c *Client) DetectProcess(port uint32) (*domain.ProcessInfo, error) {
    req := ipc.Request{
        ID:     fmt.Sprint(c.nextSeq()),
        Method: "detect",
        Params: ipc.RequestParams{Port: port},
    }

    resp, err := c.call(req)
    if err != nil {
        return nil, err
    }

    if resp.Error != nil {
        return nil, fmt.Errorf("helper error: %s", resp.Error.Message)
    }

    // Десериализовать результат
    var info domain.ProcessInfo
    data, _ := json.Marshal(resp.Result)
    json.Unmarshal(data, &info)

    return &info, nil
}

func (c *Client) ExtractIcon(pid int32) (*domain.AppIcon, error) {
    req := ipc.Request{
        ID:     fmt.Sprint(c.nextSeq()),
        Method: "icon",
        Params: ipc.RequestParams{PID: pid},
    }

    resp, err := c.call(req)
    if err != nil {
        return nil, err
    }

    if resp.Error != nil {
        return nil, fmt.Errorf("helper error: %s", resp.Error.Message)
    }

    var icon domain.AppIcon
    data, _ := json.Marshal(resp.Result)
    json.Unmarshal(data, &icon)

    return &icon, nil
}

func (c *Client) IsRunning() bool {
    return c.Ping() == nil
}

func (c *Client) Ping() error {
    req := ipc.Request{
        ID:     fmt.Sprint(c.nextSeq()),
        Method: "ping",
    }

    _, err := c.call(req)
    return err
}

func (c *Client) Close() error {
    return c.conn.Close()
}

func (c *Client) call(req ipc.Request) (*ipc.Response, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    enc := json.NewEncoder(c.conn)
    dec := json.NewDecoder(c.conn)

    if err := enc.Encode(req); err != nil {
        return nil, err
    }

    var resp ipc.Response
    if err := dec.Decode(&resp); err != nil {
        return nil, err
    }

    return &resp, nil
}

func (c *Client) nextSeq() uint64 {
    c.seq++
    return c.seq
}
```

#### Helper Installer (`internal/features/process/infrastructure/helper/installer.go`):

```go
package helper

import (
    "fmt"
    "os"
    "os/exec"
    "runtime"
)

// InstallHelperTool - установить helper tool
func InstallHelperTool() error {
    switch runtime.GOOS {
    case "darwin":
        return installMacOS()
    case "windows":
        return installWindows()
    case "linux":
        return installLinux()
    default:
        return fmt.Errorf("unsupported platform")
    }
}

func installMacOS() error {
    // 1. Скопировать binary в /Library/PrivilegedHelperTools/
    helperPath := "/Library/PrivilegedHelperTools/com.networkdebugger.helper"

    // TODO: Получить путь к текущему helper binary
    currentBinary := "./process-helper"

    // 2. Создать launchd plist
    plistPath := "/Library/LaunchDaemons/com.networkdebugger.helper.plist"
    plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.networkdebugger.helper</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Library/PrivilegedHelperTools/com.networkdebugger.helper</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>`

    // 3. Запросить пароль и выполнить установку через osascript
    script := fmt.Sprintf(`
do shell script "cp %s %s && chmod +x %s && echo '%s' > %s && launchctl load -w %s" with administrator privileges
`, currentBinary, helperPath, helperPath, plistContent, plistPath, plistPath)

    cmd := exec.Command("osascript", "-e", script)
    return cmd.Run()
}

func installWindows() error {
    // TODO: Создать Windows Service
    return fmt.Errorf("not implemented")
}

func installLinux() error {
    // TODO: Создать systemd service
    return fmt.Errorf("not implemented")
}

func IsInstalled() bool {
    switch runtime.GOOS {
    case "darwin":
        _, err := os.Stat("/Library/PrivilegedHelperTools/com.networkdebugger.helper")
        return err == nil
    default:
        return false
    }
}
```

---

### 6. Database & Persistence

#### Migration (`migrations/0004_process_detection.sql`):

```sql
-- +goose Up
-- +goose StatementBegin

-- Конфигурация детекции процессов (singleton с ID=1)
CREATE TABLE IF NOT EXISTS process_detection_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT 1,
    use_helper_tool BOOLEAN NOT NULL DEFAULT 0,
    helper_installed BOOLEAN NOT NULL DEFAULT 0,
    cache_enabled BOOLEAN NOT NULL DEFAULT 1,
    cache_ttl_seconds INTEGER NOT NULL DEFAULT 300,
    fallback_enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Кеш иконок
CREATE TABLE IF NOT EXISTS icon_cache (
    cache_key TEXT PRIMARY KEY,
    icon_format TEXT NOT NULL,
    icon_data BLOB NOT NULL,
    icon_path TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_icon_cache_expires ON icon_cache(expires_at);

-- Вставить дефолтную конфигурацию
INSERT INTO process_detection_config (id, enabled) VALUES (1, 1)
    ON CONFLICT(id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_icon_cache_expires;
DROP TABLE IF EXISTS icon_cache;
DROP TABLE IF EXISTS process_detection_config;

-- +goose StatementEnd
```

#### GORM Models (`internal/features/process/infrastructure/persistence/models.go`):

```go
package persistence

import "time"

type ProcessDetectionConfigModel struct {
    ID                int64     `gorm:"primaryKey"`
    Enabled           bool      `gorm:"not null;default:true"`
    UseHelperTool     bool      `gorm:"not null;default:false"`
    HelperInstalled   bool      `gorm:"not null;default:false"`
    CacheEnabled      bool      `gorm:"not null;default:true"`
    CacheTTLSeconds   int       `gorm:"not null;default:300"`
    FallbackEnabled   bool      `gorm:"not null;default:true"`
    CreatedAt         time.Time `gorm:"autoCreateTime"`
    UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

func (ProcessDetectionConfigModel) TableName() string {
    return "process_detection_config"
}

type IconCacheModel struct {
    CacheKey   string    `gorm:"primaryKey"`
    IconFormat string    `gorm:"not null"`
    IconData   []byte    `gorm:"not null"`
    IconPath   *string
    ExpiresAt  time.Time `gorm:"not null;index:idx_icon_cache_expires"`
    CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (IconCacheModel) TableName() string {
    return "icon_cache"
}
```

#### Config Repository (`internal/features/process/infrastructure/persistence/config_repo.go`):

```go
package persistence

import (
    "context"

    "gorm.io/gorm"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

type ConfigRepo struct {
    db *gorm.DB
}

func NewConfigRepo(db *gorm.DB) *ConfigRepo {
    return &ConfigRepo{db: db}
}

func (r *ConfigRepo) Load(ctx context.Context) (*domain.DetectionConfig, error) {
    var model ProcessDetectionConfigModel

    err := r.db.WithContext(ctx).First(&model, 1).Error
    if err != nil {
        return nil, err
    }

    return toDomainConfig(model), nil
}

func (r *ConfigRepo) Save(ctx context.Context, cfg *domain.DetectionConfig) error {
    model := toModel(cfg)
    model.ID = 1  // singleton

    return r.db.WithContext(ctx).Save(&model).Error
}

func toDomainConfig(m ProcessDetectionConfigModel) *domain.DetectionConfig {
    return &domain.DetectionConfig{
        ID:                m.ID,
        Enabled:           m.Enabled,
        UseHelperTool:     m.UseHelperTool,
        HelperInstalled:   m.HelperInstalled,
        CacheEnabled:      m.CacheEnabled,
        CacheTTLSeconds:   m.CacheTTLSeconds,
        FallbackEnabled:   m.FallbackEnabled,
        UpdatedAt:         m.UpdatedAt,
    }
}

func toModel(c *domain.DetectionConfig) ProcessDetectionConfigModel {
    return ProcessDetectionConfigModel{
        ID:                c.ID,
        Enabled:           c.Enabled,
        UseHelperTool:     c.UseHelperTool,
        HelperInstalled:   c.HelperInstalled,
        CacheEnabled:      c.CacheEnabled,
        CacheTTLSeconds:   c.CacheTTLSeconds,
        FallbackEnabled:   c.FallbackEnabled,
    }
}
```

#### Icon Cache Repository (`internal/features/process/infrastructure/persistence/icon_cache_repo.go`):

```go
package persistence

import (
    "time"

    "gorm.io/gorm"

    "github.com/belief/go-proxy/internal/features/process/domain"
)

type IconCacheRepo struct {
    db *gorm.DB
}

func NewIconCacheRepo(db *gorm.DB) *IconCacheRepo {
    return &IconCacheRepo{db: db}
}

func (r *IconCacheRepo) Get(key string) (*domain.AppIcon, error) {
    var model IconCacheModel

    err := r.db.Where("cache_key = ? AND expires_at > ?", key, time.Now()).
        First(&model).Error

    if err != nil {
        return nil, err
    }

    return &domain.AppIcon{
        Format: model.IconFormat,
        Data:   model.IconData,
        Path:   model.IconPath,
    }, nil
}

func (r *IconCacheRepo) Set(key string, icon *domain.AppIcon, ttl time.Duration) error {
    model := IconCacheModel{
        CacheKey:   key,
        IconFormat: icon.Format,
        IconData:   icon.Data,
        IconPath:   icon.Path,
        ExpiresAt:  time.Now().Add(ttl),
    }

    return r.db.Save(&model).Error
}

func (r *IconCacheRepo) Delete(key string) error {
    return r.db.Delete(&IconCacheModel{}, "cache_key = ?", key).Error
}

func (r *IconCacheRepo) Clear() error {
    return r.db.Exec("DELETE FROM icon_cache").Error
}

// CleanupExpired - удалить истекшие записи
func (r *IconCacheRepo) CleanupExpired() error {
    return r.db.Delete(&IconCacheModel{}, "expires_at < ?", time.Now()).Error
}
```

---

### 7. HTTP API Layer

#### Handlers (`internal/infrastructure/httpapi/process_handlers.go`):

```go
package httpapi

import (
    "encoding/json"
    "net/http"

    processuc "github.com/belief/go-proxy/internal/features/process/usecase"
)

type processConfigDTO struct {
    Enabled         bool `json:"enabled"`
    UseHelperTool   bool `json:"useHelperTool"`
    HelperInstalled bool `json:"helperInstalled"`
    CacheEnabled    bool `json:"cacheEnabled"`
    CacheTTL        int  `json:"cacheTtl"`
    FallbackEnabled bool `json:"fallbackEnabled"`
}

func (d *Deps) handleV1ProcessConfig(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    switch r.Method {
    case http.MethodGet:
        cfg, err := d.ProcessSvc.GetConfig(ctx)
        if err != nil {
            respondError(w, http.StatusInternalServerError, err)
            return
        }

        respondJSON(w, http.StatusOK, processConfigToDTO(cfg))

    case http.MethodPost:
        var dto processConfigDTO
        if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
            respondError(w, http.StatusBadRequest, err)
            return
        }

        cfg := dtoToProcessConfig(&dto)
        if err := d.ProcessSvc.SaveConfig(ctx, cfg); err != nil {
            respondError(w, http.StatusInternalServerError, err)
            return
        }

        respondJSON(w, http.StatusOK, processConfigToDTO(cfg))

    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (d *Deps) handleV1ProcessHelperStatus(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    status := d.ProcessSvc.CheckHelperStatus()

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "running":   status.Running,
        "installed": status.Installed,
        "version":   status.Version,
    })
}

func (d *Deps) handleV1ProcessHelperInstall(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    ctx := r.Context()

    if err := d.ProcessSvc.InstallHelper(ctx); err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "status": "installed",
    })
}

func processConfigToDTO(cfg *domain.DetectionConfig) processConfigDTO {
    return processConfigDTO{
        Enabled:         cfg.Enabled,
        UseHelperTool:   cfg.UseHelperTool,
        HelperInstalled: cfg.HelperInstalled,
        CacheEnabled:    cfg.CacheEnabled,
        CacheTTL:        cfg.CacheTTLSeconds,
        FallbackEnabled: cfg.FallbackEnabled,
    }
}

func dtoToProcessConfig(dto *processConfigDTO) *domain.DetectionConfig {
    return &domain.DetectionConfig{
        ID:                1,  // singleton
        Enabled:           dto.Enabled,
        UseHelperTool:     dto.UseHelperTool,
        HelperInstalled:   dto.HelperInstalled,
        CacheEnabled:      dto.CacheEnabled,
        CacheTTLSeconds:   dto.CacheTTL,
        FallbackEnabled:   dto.FallbackEnabled,
    }
}
```

#### Extend Session DTO (`internal/infrastructure/httpapi/sessions.go`):

```go
// Добавить к существующему sessionDTO:

type sessionDTO struct {
    // ... existing fields
    ProcessInfo *processInfoDTO `json:"processInfo,omitempty"`
}

type processInfoDTO struct {
    PID      int32   `json:"pid"`
    Name     string  `json:"name"`
    ExePath  string  `json:"exePath,omitempty"`
    BundleID *string `json:"bundleId,omitempty"`
    Icon     *string `json:"icon,omitempty"`  // Base64 encoded PNG
}

func sessionToDTO(s *domain.Session) sessionDTO {
    dto := sessionDTO{
        // ... existing mappings
    }

    if s.ProcessInfo != nil {
        dto.ProcessInfo = processInfoToDTO(s.ProcessInfo)
    }

    return dto
}

func processInfoToDTO(info *domain.ProcessInfo) *processInfoDTO {
    dto := &processInfoDTO{
        PID:      info.PID,
        Name:     info.Name,
        ExePath:  info.ExecutablePath,
        BundleID: info.BundleID,
    }

    if info.Icon != nil && len(info.Icon.Data) > 0 {
        // Encode icon as Base64
        encoded := base64.StdEncoding.EncodeToString(info.Icon.Data)
        dto.Icon = &encoded
    }

    return dto
}
```

#### Router Registration (`internal/infrastructure/httpapi/router.go`):

```go
// Добавить в NewRouter():

func NewRouter(deps *Deps) http.Handler {
    mux := http.NewServeMux()

    // ... existing routes

    // Process detection API
    mux.HandleFunc("/_api/v1/process/config", deps.handleV1ProcessConfig)
    mux.HandleFunc("/_api/v1/process/helper/status", deps.handleV1ProcessHelperStatus)
    mux.HandleFunc("/_api/v1/process/helper/install", deps.handleV1ProcessHelperInstall)

    return mux
}
```

#### Update Deps (`internal/infrastructure/httpapi/router.go`):

```go
type Deps struct {
    // ... existing deps
    ProcessSvc *processuc.Service  // NEW
}
```

---

### 8. Frontend Integration

#### Service (`frontend/lib/features/settings/application/process_service.dart`):

```dart
import 'package:injectable/injectable.dart';
import '../../../core/network/app_http_client.dart';

@lazySingleton
class ProcessService {
  final AppHttpClient _api;

  ProcessService(this._api);

  Future<ProcessConfig> fetchConfig() async {
    final res = await _api.get(path: '/_api/v1/process/config');
    return ProcessConfig.fromJson(res.data as Map<String, dynamic>);
  }

  Future<void> saveConfig(ProcessConfig config) async {
    await _api.post(
      path: '/_api/v1/process/config',
      data: config.toJson(),
    );
  }

  Future<HelperStatus> checkHelperStatus() async {
    final res = await _api.get(path: '/_api/v1/process/helper/status');
    return HelperStatus.fromJson(res.data as Map<String, dynamic>);
  }

  Future<void> installHelper() async {
    await _api.post(path: '/_api/v1/process/helper/install');
  }
}

class ProcessConfig {
  final bool enabled;
  final bool useHelperTool;
  final bool helperInstalled;
  final bool cacheEnabled;
  final int cacheTtl;
  final bool fallbackEnabled;

  ProcessConfig({
    required this.enabled,
    required this.useHelperTool,
    required this.helperInstalled,
    required this.cacheEnabled,
    required this.cacheTtl,
    required this.fallbackEnabled,
  });

  factory ProcessConfig.fromJson(Map<String, dynamic> json) {
    return ProcessConfig(
      enabled: json['enabled'] as bool,
      useHelperTool: json['useHelperTool'] as bool,
      helperInstalled: json['helperInstalled'] as bool,
      cacheEnabled: json['cacheEnabled'] as bool,
      cacheTtl: json['cacheTtl'] as int,
      fallbackEnabled: json['fallbackEnabled'] as bool,
    );
  }

  Map<String, dynamic> toJson() => {
        'enabled': enabled,
        'useHelperTool': useHelperTool,
        'helperInstalled': helperInstalled,
        'cacheEnabled': cacheEnabled,
        'cacheTtl': cacheTtl,
        'fallbackEnabled': fallbackEnabled,
      };

  ProcessConfig copyWith({
    bool? enabled,
    bool? useHelperTool,
    bool? helperInstalled,
    bool? cacheEnabled,
    int? cacheTtl,
    bool? fallbackEnabled,
  }) {
    return ProcessConfig(
      enabled: enabled ?? this.enabled,
      useHelperTool: useHelperTool ?? this.useHelperTool,
      helperInstalled: helperInstalled ?? this.helperInstalled,
      cacheEnabled: cacheEnabled ?? this.cacheEnabled,
      cacheTtl: cacheTtl ?? this.cacheTtl,
      fallbackEnabled: fallbackEnabled ?? this.fallbackEnabled,
    );
  }
}

class HelperStatus {
  final bool running;
  final bool installed;
  final String version;

  HelperStatus({
    required this.running,
    required this.installed,
    required this.version,
  });

  factory HelperStatus.fromJson(Map<String, dynamic> json) {
    return HelperStatus(
      running: json['running'] as bool,
      installed: json['installed'] as bool,
      version: json['version'] as String? ?? '',
    );
  }
}
```

#### Settings UI (`frontend/lib/features/settings/presentation/process_page.dart`):

```dart
import 'package:flutter/material.dart';
import '../application/process_service.dart';

class ProcessSettingsPage extends StatefulWidget {
  @override
  _ProcessSettingsPageState createState() => _ProcessSettingsPageState();
}

class _ProcessSettingsPageState extends State<ProcessSettingsPage> {
  final ProcessService _service = sl<ProcessService>();

  ProcessConfig? _config;
  HelperStatus? _helperStatus;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadSettings();
  }

  Future<void> _loadSettings() async {
    try {
      final config = await _service.fetchConfig();
      final status = await _service.checkHelperStatus();

      setState(() {
        _config = config;
        _helperStatus = status;
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
      // Show error
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loading || _config == null) {
      return Center(child: CircularProgressIndicator());
    }

    return ListView(
      children: [
        SwitchListTile(
          title: Text('Enable Process Detection'),
          subtitle: Text('Detect which application made the request'),
          value: _config!.enabled,
          onChanged: (value) => _updateConfig(_config!.copyWith(enabled: value)),
        ),

        if (_config!.enabled) ...[
          Divider(),

          SwitchListTile(
            title: Text('Use Helper Tool'),
            subtitle: Text('Detect all applications (requires admin password)'),
            value: _config!.useHelperTool,
            onChanged: (value) async {
              if (value && !_config!.helperInstalled) {
                await _promptInstallHelper();
              } else {
                _updateConfig(_config!.copyWith(useHelperTool: value));
              }
            },
          ),

          if (_config!.useHelperTool && !_config!.helperInstalled) ...[
            Padding(
              padding: EdgeInsets.all(16),
              child: Column(
                children: [
                  Text('Helper tool not installed'),
                  SizedBox(height: 8),
                  ElevatedButton(
                    child: Text('Install Helper Tool'),
                    onPressed: _installHelper,
                  ),
                ],
              ),
            ),
          ],

          if (_config!.useHelperTool && _config!.helperInstalled) ...[
            ListTile(
              leading: Icon(
                _helperStatus?.running == true ? Icons.check_circle : Icons.error,
                color: _helperStatus?.running == true ? Colors.green : Colors.red,
              ),
              title: Text('Helper Status'),
              subtitle: Text(
                _helperStatus?.running == true ? 'Running' : 'Not running',
              ),
            ),
          ],

          Divider(),

          SwitchListTile(
            title: Text('Cache Icons'),
            subtitle: Text('Improve performance by caching app icons'),
            value: _config!.cacheEnabled,
            onChanged: (value) => _updateConfig(_config!.copyWith(cacheEnabled: value)),
          ),
        ],
      ],
    );
  }

  Future<void> _updateConfig(ProcessConfig config) async {
    try {
      await _service.saveConfig(config);
      setState(() => _config = config);
    } catch (e) {
      // Show error
    }
  }

  Future<void> _promptInstallHelper() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('Install Helper Tool'),
        content: Text(
          'This requires administrator password (one time only).\n\n'
          'The helper tool will allow detection of all applications on your system.'
        ),
        actions: [
          TextButton(
            child: Text('Cancel'),
            onPressed: () => Navigator.pop(ctx, false),
          ),
          ElevatedButton(
            child: Text('Install'),
            onPressed: () => Navigator.pop(ctx, true),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await _installHelper();
    }
  }

  Future<void> _installHelper() async {
    try {
      await _service.installHelper();

      // Reload config
      await _loadSettings();

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Helper tool installed successfully')),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to install helper: $e')),
      );
    }
  }
}
```

#### Session List Update (`frontend/lib/features/inspector/presentation/widgets/session_item.dart`):

```dart
// Обновить существующий SessionItem:

class SessionItem extends StatelessWidget {
  final Session session;

  const SessionItem({required this.session});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: _buildAppIcon(),
      title: Text(_getAppName()),
      subtitle: Text('${session.method} ${session.url}'),
      // ... rest of the item
    );
  }

  Widget _buildAppIcon() {
    if (session.processInfo?.icon != null) {
      try {
        final bytes = base64Decode(session.processInfo!.icon!);
        return Image.memory(
          bytes,
          width: 32,
          height: 32,
          errorBuilder: (_, __, ___) => _defaultIcon(),
        );
      } catch (e) {
        return _defaultIcon();
      }
    }
    return _defaultIcon();
  }

  Widget _defaultIcon() {
    return Icon(Icons.help_outline, size: 32);
  }

  String _getAppName() {
    return session.processInfo?.name ?? 'Unknown Process';
  }
}
```

---

### 9. Integration & Initialization

#### Main Entry Point (`cmd/network-debugger/main.go`):

```go
func main() {
    // ... existing setup (config, logger, db, etc.)

    // Initialize process detection feature
    processDetector, err := detector.NewDetector(false)  // unprivileged
    if err != nil {
        logger.Error().Err(err).Msg("Failed to create process detector")
        processDetector = nil
    }

    iconExtractor, err := icon.NewExtractor()
    if err != nil {
        logger.Error().Err(err).Msg("Failed to create icon extractor")
        iconExtractor = nil
    }

    // Try to connect to helper (if installed)
    var helperClient domain.IHelperClient
    if helper.IsInstalled() {
        helperClient, err = helper.NewClient()
        if err != nil {
            logger.Warn().Err(err).Msg("Failed to connect to helper")
        } else {
            logger.Info().Msg("Connected to process helper")
        }
    }

    // Initialize repositories
    processConfigRepo := persistence.NewConfigRepo(db)
    iconCacheRepo := persistence.NewIconCacheRepo(db)

    // Initialize service
    processSvc := usecase.NewService(
        processConfigRepo,
        iconCacheRepo,
        processDetector,
        helperClient,
        iconExtractor,
        logger,
    )

    // Add to HTTP API dependencies
    httpDeps := &httpapi.Deps{
        Cfg:        cfg,
        Logger:     logger,
        DB:         db,
        ProcessSvc: processSvc,  // NEW
        // ... other deps
    }

    // Start HTTP server
    router := httpapi.NewRouter(httpDeps)
    // ...
}
```

#### Session Service Integration (`internal/usecase/services.go`):

```go
// Обновить SessionService:

type SessionService struct {
    // ... existing fields
    processSvc *processuc.Service  // NEW
}

func (s *SessionService) CreateSession(/* ... */) (*domain.Session, error) {
    session := &domain.Session{
        // ... existing initialization
    }

    // Detect process (if enabled)
    if s.processSvc != nil {
        ctx := context.Background()

        // Extract local port from connection
        localPort := extractLocalPort(/* from connection info */)

        processInfo, err := s.processSvc.DetectForConnection(ctx, localPort)
        if err != nil {
            s.logger.Debug().Err(err).Msg("Failed to detect process")
        } else {
            session.ProcessInfo = processInfo
        }
    }

    // ... rest of the method
    return session, nil
}
```

---

## 📅 Пошаговый План Реализации (5 недель)

### **Неделя 1: Foundation & Core Infrastructure**

#### День 1-2: Domain Layer (2 дня)
- [ ] Создать структуру директорий `internal/features/process/`
- [ ] Определить entities (`domain/entities.go`):
  - ProcessInfo, AppIcon, DetectionConfig
- [ ] Определить все интерфейсы:
  - IProcessDetector (`domain/detector.go`)
  - IIconExtractor (`domain/icon_extractor.go`)
  - IHelperClient (`domain/helper.go`)
  - IConfigRepository, IIconCacheRepository (`domain/repository.go`)
- [ ] Написать юнит-тесты для entities (validation methods)
- [ ] Документация GoDoc для всех интерфейсов

#### День 3-4: Database & Persistence (2 дня)
- [ ] Написать миграцию `migrations/0004_process_detection.sql`
- [ ] GORM модели (`infrastructure/persistence/models.go`):
  - ProcessDetectionConfigModel
  - IconCacheModel
- [ ] Реализовать ConfigRepo (`infrastructure/persistence/config_repo.go`)
- [ ] Реализовать IconCacheRepo (`infrastructure/persistence/icon_cache_repo.go`)
- [ ] Юнит-тесты с SQLite in-memory
- [ ] Применить миграцию: `goose up`

#### День 5: Use Case Layer (1 день)
- [ ] Реализовать ProcessService (`usecase/service.go`)
- [ ] Методы:
  - DetectForConnection (с fallback стратегией)
  - getIcon (с кешированием)
  - GetConfig, SaveConfig
  - CheckHelperStatus, InstallHelper
- [ ] Юнит-тесты с mock интерфейсами
- [ ] Coverage >80%

---

### **Неделя 2: Platform-Specific Implementations**

#### День 1-2: Detector Implementations (2 дня)
- [ ] Factory в `detector/detector.go`
- [ ] **macOS**: `detector_darwin.go`
  - lsof command wrapper
  - Парсинг lsof output
  - Тесты на macOS
- [ ] **Windows**: `detector_windows.go`
  - gopsutil wrapper
  - GetExtendedTcpTable fallback
  - Тесты на Windows
- [ ] **Linux**: `detector_linux.go`
  - /proc/net/tcp parsing
  - inode → PID resolution
  - Тесты на Linux

#### День 3-4: Icon Extractors (2 дня)
- [ ] Factory в `icon/extractor.go`
- [ ] **macOS**: `extractor_darwin.go`
  - fileicon utility wrapper
  - sips PNG conversion
  - App bundle detection
  - Тесты с реальными приложениями
- [ ] **Windows**: `extractor_windows.go`
  - Skeleton (TODO: lxn/win integration)
  - Заглушка с fallback icon
- [ ] **Linux**: `extractor_linux.go`
  - Skeleton (TODO: .desktop files)
  - Заглушка с fallback icon

#### День 5: Dependencies & Testing (1 день)
- [ ] Обновить `go.mod`:
  - `github.com/shirou/gopsutil/v4`
  - `github.com/lxn/win` (для Windows в будущем)
- [ ] Установить утилиты:
  - macOS: `brew install fileicon`
- [ ] Интеграционные тесты для детекторов
- [ ] go fmt, go vet, golangci-lint
- [ ] Coverage report

---

### **Неделя 3: Helper Tool (Privileged Daemon)**

#### День 1-2: Helper Daemon (2 дня)
- [ ] Создать `cmd/process-helper/main.go`
- [ ] IPC protocol (`ipc/protocol.go`):
  - Request/Response structs (JSON-RPC style)
- [ ] IPC transport (`ipc/transport.go`):
  - Unix socket (macOS/Linux)
  - Placeholder для named pipe (Windows)
- [ ] Server implementation (`server/server.go`):
  - Accept connections
  - Handle requests: detect, icon, ping
  - Use privileged detector
- [ ] Build helper binary: `go build -o process-helper cmd/process-helper/main.go`
- [ ] Manual testing: запуск с sudo

#### День 3: Helper Client (1 день)
- [ ] Client implementation (`infrastructure/helper/client.go`):
  - Connect to Unix socket
  - DetectProcess, ExtractIcon, Ping, IsRunning
  - Connection pooling/reuse
- [ ] Error handling & retries
- [ ] Юнит-тесты с mock IPC
- [ ] Integration test: запустить helper, подключиться, сделать запрос

#### День 4-5: Helper Installer (2 дня)
- [ ] macOS installer (`infrastructure/helper/installer.go`):
  - SMJobBless workflow
  - Copy binary to /Library/PrivilegedHelperTools/
  - Create launchd plist
  - Load via launchctl
  - Password prompt через osascript
- [ ] Windows installer (skeleton):
  - TODO: Windows Service creation
- [ ] Linux installer (skeleton):
  - TODO: systemd service
- [ ] IsInstalled() check
- [ ] Manual testing: установка на чистой системе
- [ ] Документация: troubleshooting permissions

---

### **Неделя 4: API & Frontend Integration**

#### День 1: HTTP API (1 день)
- [ ] Handlers (`httpapi/process_handlers.go`):
  - handleV1ProcessConfig (GET/POST)
  - handleV1ProcessHelperStatus (GET)
  - handleV1ProcessHelperInstall (POST)
- [ ] DTOs:
  - processConfigDTO
  - processInfoDTO (extend sessionDTO)
- [ ] Routes в `router.go`:
  - `/_api/v1/process/config`
  - `/_api/v1/process/helper/status`
  - `/_api/v1/process/helper/install`
- [ ] Обновить Deps struct
- [ ] API integration tests (httptest)

#### День 2-3: Flutter Frontend (2 дня)
- [ ] ProcessService (`application/process_service.dart`):
  - fetchConfig, saveConfig
  - checkHelperStatus, installHelper
- [ ] Models:
  - ProcessConfig (freezed/json_serializable)
  - HelperStatus
- [ ] ProcessSettingsPage UI (`presentation/process_page.dart`):
  - Enable/disable toggle
  - Helper tool section
  - Install button with confirmation dialog
  - Status indicator
- [ ] Интеграция в Settings page
- [ ] UI тесты (widget tests)

#### День 4: Session Display (1 день)
- [ ] Обновить Session model:
  - Добавить processInfo field
- [ ] SessionItem widget updates:
  - Отображение иконки приложения
  - Отображение имени процесса
  - Fallback на "Unknown" icon
- [ ] Base64 decode для иконок
- [ ] Тесты: session list с process info

#### День 5: End-to-End Integration (1 день)
- [ ] Инициализация в `main.go`:
  - Create detector, extractor, helper client
  - Create repositories
  - Create ProcessService
  - Add to Deps
- [ ] SessionService integration:
  - Вызвать DetectForConnection при создании сессии
  - Передать processInfo в Session
- [ ] Monitor WebSocket: отправлять processInfo
- [ ] E2E тест:
  - Запустить proxy
  - Сделать запрос через браузер
  - Проверить, что в UI видно имя браузера + иконка
- [ ] Smoke tests на всех платформах

---

### **Неделя 5: Testing, Optimization & Release**

#### День 1-2: Comprehensive Testing (2 дня)
- [ ] Unit tests:
  - Coverage >80% для всех packages
  - Моки для всех интерфейсов
- [ ] Integration tests:
  - Database repo tests
  - Detector tests (на реальной системе)
  - Icon extractor tests
  - Helper IPC tests
- [ ] E2E tests:
  - Full flow: request → detect → icon → display
  - С helper и без helper
  - Permission denied scenarios
- [ ] Edge cases:
  - Process not found
  - Helper не запущен
  - Кеш истек
  - Некорректные иконки

#### День 3-4: Optimization & Performance (2 дня)
- [ ] Icon caching performance:
  - Benchmark cache hit/miss
  - Optimize TTL
- [ ] Batch detection:
  - Если много соединений одновременно
  - Debounce частых запросов
- [ ] Memory profiling:
  - pprof анализ
  - Проверка утечек памяти в кеше
- [ ] Cleanup expired cache:
  - Background job для удаления истекших записей
- [ ] Concurrent safety:
  - Race detector: `go test -race ./...`

#### День 5: Documentation & Release (1 день)
- [ ] go fmt весь код
- [ ] golangci-lint проверка
- [ ] Документация:
  - README: Process Detection feature
  - Architecture diagram
  - API documentation
  - Troubleshooting guide (permissions, helper install)
- [ ] Build helper executables:
  - macOS: `GOOS=darwin GOARCH=amd64 go build -o process-helper-darwin`
  - Windows: `GOOS=windows GOARCH=amd64 go build -o process-helper-windows.exe`
  - Linux: `GOOS=linux GOARCH=amd64 go build -o process-helper-linux`
- [ ] Package для distribution
- [ ] Release notes
- [ ] Demo video: показать feature в действии

---

## ✅ Definition of Done

### Функциональность:
- [x] Детекция процессов работает на macOS/Windows/Linux
- [x] Иконки извлекаются и отображаются (macOS)
- [x] Helper tool устанавливается через UI
- [x] Работает без admin прав (graceful degradation)
- [x] Кеширование иконок для производительности
- [x] Настройки через UI (enable/disable, helper)

### Качество кода:
- [x] Clean Architecture (domain/usecase/infrastructure)
- [x] SOLID принципы соблюдены
- [x] DRY - нет дублирования
- [x] Unit test coverage >80%
- [x] Integration tests написаны
- [x] Race detector проходит
- [x] golangci-lint без ошибок
- [x] go fmt применен

### Документация:
- [x] GoDoc для всех публичных API
- [x] README обновлен
- [x] Architecture diagram создана
- [x] Troubleshooting guide написан
- [x] API documentation готова

### Производительность:
- [x] Детекция не блокирует proxy (<50ms)
- [x] Кеш работает эффективно (hit rate >90%)
- [x] Память не утекает
- [x] No performance regression

---

## 🎓 Принципы Clean Architecture

### Dependency Rule:
```
Domain ← UseCase ← Infrastructure
  ↑        ↑            ↑
  |        |            |
  └────────┴────────────┘
       (Dependencies flow inward)
```

**Domain Layer:**
- Чистая бизнес-логика
- Никаких зависимостей на фреймворки/DB
- Только интерфейсы (ports)

**Use Case Layer:**
- Orchestration бизнес-логики
- Зависит только от domain интерфейсов
- Не знает про DB/HTTP/UI

**Infrastructure Layer:**
- Реализации интерфейсов (adapters)
- Зависимости на GORM, gopsutil, OS APIs
- Можно заменить без изменения domain/usecase

### Тестируемость:
```go
// Mock для unit tests:
type MockDetector struct {
    mock.Mock
}

func (m *MockDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
    args := m.Called(ctx, port)
    return args.Get(0).(*domain.ProcessInfo), args.Error(1)
}

// Test:
mockDetector := new(MockDetector)
mockDetector.On("DetectByPort", ctx, uint32(8080)).Return(info, nil)

service := usecase.NewService(/* ... */, mockDetector, /* ... */)
result, _ := service.DetectForConnection(ctx, 8080)

mockDetector.AssertExpectations(t)
```

---

## 🔒 Безопасность

### Privileged Operations:
- Helper tool запускается как root/admin
- IPC через Unix socket (permissions 0600)
- Только локальные соединения (localhost)
- Валидация всех IPC запросов

### Password Prompt:
- macOS: osascript (native macOS dialog)
- Windows: UAC prompt (native Windows)
- Linux: pkexec/sudo

### Permissions:
- Без helper: видны только свои процессы ✅
- С helper: видны все процессы (требует пароль) ⚠️
- Никакого сетевого доступа из helper
- Никаких изменений системы (read-only operations)

---

## 📈 Метрики Успеха

1. **Функциональность**: ✅ Показывает иконки и имена приложений
2. **UX**: ✅ Работает сразу (без пароля), опциональный helper
3. **Производительность**: ✅ <50ms на детекцию, кеш работает
4. **Стабильность**: ✅ Не крашится при ошибках, graceful degradation
5. **Тестирование**: ✅ >80% coverage, все тесты проходят
6. **Кросс-платформа**: ✅ macOS/Windows/Linux

---

## 🚀 Будущие Улучшения (Post-MVP)

1. **Windows Icon Extraction**:
   - Реализовать через lxn/win + ExtractIconEx

2. **Linux Icon Extraction**:
   - Парсинг .desktop файлов
   - Freedesktop icon theme resolution

3. **Advanced Caching**:
   - LRU cache в памяти
   - Persistent cache на диске

4. **Process Filtering**:
   - Фильтрация по приложению в UI
   - Группировка по процессам

5. **Statistics**:
   - Топ процессов по трафику
   - График активности по приложениям

6. **Performance**:
   - Batch detection для множества соединений
   - Background preloading icons

---

**Конец плана реализации. Готов к execution!** 🎯
