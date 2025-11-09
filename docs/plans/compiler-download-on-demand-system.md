# Compiler Download-on-Demand System - Implementation Plan

## 🎯 Vision

Создать профессиональную систему автоматического управления компиляторами:
- **Нет установки** - юзер просто выбирает язык
- **Auto-download** - компилятор скачивается автоматически при первом использовании
- **Cache** - хранится в `~/.cache/network-debugger/compilers/`
- **Progress tracking** - WebSocket с real-time прогрессом
- **Scalable** - легко добавить новые языки

## 🏗️ Architecture (Clean Architecture + SOLID + DDD)

### Layer 1: Domain (PORT - бизнес-логика)

```
internal/features/scripting/domain/
├── compiler_manager.go          ← CompilerManager interface
├── compiler_downloader.go       ← CompilerDownloader interface
└── compiler_cache.go            ← CacheManager interface
```

**Key Interfaces:**

#### CompilerManager (управление lifecycle)
```go
type CompilerStatus string
const (
    StatusNotInstalled CompilerStatus = "not_installed"
    StatusInstalling   CompilerStatus = "installing"
    StatusInstalled    CompilerStatus = "installed"
    StatusError        CompilerStatus = "error"
)

type CompilerInfo struct {
    Language      string
    Version       string
    Status        CompilerStatus
    InstalledPath string
    Size          int64
    DownloadSize  int64  // Size to download
    Error         string
}

type CompilerManager interface {
    GetStatus(language string) (*CompilerInfo, error)
    IsInstalled(language string) bool
    GetInstalledPath(language string) (string, error)
    ListAll() ([]*CompilerInfo, error)
}
```

#### CompilerDownloader (скачивание и установка)
```go
type DownloadProgress struct {
    Stage           string  // "downloading", "extracting", "verifying"
    BytesDownloaded int64
    TotalBytes      int64
    Percentage      float64
    Speed           int64  // bytes/sec
    Message         string
}

type DownloadRequest struct {
    Language    string
    Version     string // "latest" or specific version like "0.14.0"
    Platform    string // runtime.GOOS
    Arch        string // runtime.GOARCH
    TargetDir   string
}

type CompilerMetadata struct {
    Language    string
    Version     string
    DownloadURL string
    Size        int64
    Checksum    string
    ReleaseDate time.Time
}

type CompilerDownloader interface {
    // Get download URL for platform/arch
    GetDownloadURL(req DownloadRequest) (string, error)

    // Download with progress callback
    Download(ctx context.Context, req DownloadRequest, progressCb func(DownloadProgress)) error

    // Extract archive to target directory
    Extract(archivePath, targetDir string) error

    // Verify installation works
    Verify(installPath string) error

    // Get metadata (size, checksum, version)
    GetMetadata(req DownloadRequest) (*CompilerMetadata, error)
}
```

#### CacheManager (управление кешем)
```go
type CacheManager interface {
    GetCacheDir() string
    GetCompilerPath(language string) (string, error)
    EnsureCacheDir() error
    Clear(language string) error
    ClearAll() error
    GetCacheSize() (int64, error)
}
```

### Layer 2: UseCase (оркестрация)

```
internal/features/scripting/usecase/
└── compiler_installation_service.go
```

**CompilerInstallationService:**
```go
type CompilerInstallationService struct {
    downloaders map[string]domain.CompilerDownloader  // language -> downloader
    cache       domain.CacheManager
    compilers   map[string]domain.Compiler
}

// Main methods:
func (s *CompilerInstallationService) CheckAvailability(language string) (*domain.CompilerInfo, error)
func (s *CompilerInstallationService) InstallCompiler(ctx context.Context, language, version string, progressCb func(domain.DownloadProgress)) error
func (s *CompilerInstallationService) UninstallCompiler(language string) error
func (s *CompilerInstallationService) GetMetadata(language string) (*domain.CompilerMetadata, error)
func (s *CompilerInstallationService) ListCompilers() ([]*domain.CompilerInfo, error)
```

**Installation Flow:**
1. Check if already installed (system or cache)
2. Get metadata (size, URL, version)
3. Download to temp file with progress tracking
4. Verify checksum
5. Extract to cache directory
6. Verify installation works
7. Update compiler path
8. Cleanup temp files

### Layer 3: Infrastructure (ADAPTERS - реализации)

```
internal/features/scripting/infrastructure/
├── downloaders/
│   ├── base.go                  ← BaseDownloader (DRY)
│   ├── zig_downloader.go        ← ZigDownloader
│   ├── kotlin_downloader.go     ← KotlinDownloader
│   └── swift_downloader.go      ← SwiftDownloader
├── cache/
│   └── filesystem_cache.go      ← FileSystemCache
└── compilers/
    ├── zig.go                   ← ZigCompiler (NEW)
    ├── kotlin.go                ← KotlinCompiler (NEW)
    └── swift.go                 ← SwiftCompiler (NEW)
```

#### BaseDownloader (DRY - shared logic)

**Important:** All common download/extract logic here to avoid duplication!

```go
type BaseDownloader struct {
    httpClient *http.Client
}

// Download file with progress tracking
func (b *BaseDownloader) DownloadFile(
    ctx context.Context,
    url string,
    dest string,
    progressCb func(domain.DownloadProgress),
) error {
    // Create HTTP request
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

    // Execute request
    resp, err := b.httpClient.Do(req)
    defer resp.Body.Close()

    // Create destination file
    out, err := os.Create(dest)
    defer out.Close()

    // Track progress
    totalBytes := resp.ContentLength
    var downloaded int64
    startTime := time.Now()

    // Copy with progress
    buf := make([]byte, 32*1024) // 32 KB buffer
    for {
        nr, err := resp.Body.Read(buf)
        if nr > 0 {
            nw, err := out.Write(buf[:nr])
            downloaded += int64(nw)

            // Calculate progress
            elapsed := time.Since(startTime).Seconds()
            speed := int64(float64(downloaded) / elapsed)
            percentage := float64(downloaded) / float64(totalBytes) * 100

            // Callback
            if progressCb != nil {
                progressCb(domain.DownloadProgress{
                    Stage:           "downloading",
                    BytesDownloaded: downloaded,
                    TotalBytes:      totalBytes,
                    Percentage:      percentage,
                    Speed:           speed,
                    Message:         fmt.Sprintf("Downloaded %s / %s", formatBytes(downloaded), formatBytes(totalBytes)),
                })
            }
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
    }

    return nil
}

// Extract tar.xz archive
func (b *BaseDownloader) ExtractTarXz(archivePath, targetDir string) error {
    // Open archive file
    file, err := os.Open(archivePath)
    defer file.Close()

    // Create xz reader
    xzReader, err := xz.NewReader(file)

    // Create tar reader
    tarReader := tar.NewReader(xzReader)

    // Extract files
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }

        target := filepath.Join(targetDir, header.Name)

        switch header.Typeflag {
        case tar.TypeDir:
            os.MkdirAll(target, 0755)
        case tar.TypeReg:
            outFile, err := os.Create(target)
            io.Copy(outFile, tarReader)
            outFile.Close()
            os.Chmod(target, os.FileMode(header.Mode))
        }
    }

    return nil
}

// Extract zip archive
func (b *BaseDownloader) ExtractZip(archivePath, targetDir string) error {
    // Similar logic for zip
}

// Verify checksum (SHA256)
func (b *BaseDownloader) VerifyChecksum(filePath, expectedChecksum string) error {
    file, err := os.Open(filePath)
    defer file.Close()

    hash := sha256.New()
    io.Copy(hash, file)

    actualChecksum := hex.EncodeToString(hash.Sum(nil))

    if actualChecksum != expectedChecksum {
        return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
    }

    return nil
}
```

#### ZigDownloader (ADAPTER for Zig)

**Important nuances:**

1. **Version Detection:**
   - Fetch from https://ziglang.org/download/index.json
   - Parse JSON to get latest version, URLs, checksums

2. **Platform Mapping:**
   - Go `darwin` → Zig `macos`
   - Go `linux` → Zig `linux`
   - Go `windows` → Zig `windows`
   - Arch: `amd64` → `x86_64`, `arm64` → `aarch64`

3. **Archive Format:**
   - Linux/macOS: `.tar.xz` (use ExtractTarXz)
   - Windows: `.zip` (use ExtractZip)

4. **Installation Structure:**
   ```
   ~/.cache/network-debugger/compilers/zig/
   └── zig-macos-x86_64-0.14.0/
       ├── zig         ← binary
       ├── lib/        ← standard library
       └── ...
   ```

5. **Binary Path:**
   - After extraction: `{cacheDir}/zig/zig-{platform}-{arch}-{version}/zig`
   - Symlink to: `{cacheDir}/zig/current/zig` for easy access

```go
type ZigDownloader struct {
    BaseDownloader
}

func (d *ZigDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
    // Fetch index.json
    resp, err := http.Get("https://ziglang.org/download/index.json")
    defer resp.Body.Close()

    var index map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&index)

    // Get version
    version := req.Version
    if version == "latest" {
        version = index["master"].(map[string]interface{})["version"].(string)
    }

    // Map platform
    platform := req.Platform
    if platform == "darwin" {
        platform = "macos"
    }

    // Map arch
    arch := req.Arch
    if arch == "amd64" {
        arch = "x86_64"
    } else if arch == "arm64" {
        arch = "aarch64"
    }

    // Determine extension
    ext := "tar.xz"
    if req.Platform == "windows" {
        ext = "zip"
    }

    // Build URL
    url := fmt.Sprintf(
        "https://ziglang.org/download/%s/zig-%s-%s-%s.%s",
        version, platform, arch, version, ext,
    )

    return url, nil
}

func (d *ZigDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
    // Parse index.json to get size, checksum
}

func (d *ZigDownloader) Verify(installPath string) error {
    // Run: {installPath}/zig version
    cmd := exec.Command(filepath.Join(installPath, "zig"), "version")
    output, err := cmd.Output()

    if err != nil {
        return fmt.Errorf("zig verification failed: %v", err)
    }

    // Check output contains version
    if !strings.Contains(string(output), "0.") {
        return fmt.Errorf("unexpected zig version output: %s", string(output))
    }

    return nil
}
```

#### KotlinDownloader (ADAPTER for Kotlin)

**Important nuances:**

1. **Two-Step Download:**
   - First: Kotlin compiler (~75 MB)
   - Second: Minimal JRE (~100-150 MB)
   - **OR** detect system JRE and skip download

2. **JRE Detection:**
   ```go
   func (d *KotlinDownloader) DetectSystemJRE() (string, error) {
       // Check JAVA_HOME
       if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
           return javaHome, nil
       }

       // Check `java` command
       cmd := exec.Command("java", "-version")
       output, err := cmd.CombinedOutput()
       if err == nil {
           // Parse version from output
           // Return path if version >= 11
       }

       return "", fmt.Errorf("no JRE found")
   }
   ```

3. **Kotlin Download URL:**
   - https://github.com/JetBrains/kotlin/releases/download/v2.2.20/kotlin-compiler-2.2.20.zip

4. **JRE Download (if needed):**
   - Adoptium: https://api.adoptium.net/v3/binary/latest/17/ga/{os}/{arch}/jre/hotspot/normal/eclipse
   - Or Amazon Corretto

5. **Installation Structure:**
   ```
   ~/.cache/network-debugger/compilers/kotlin/
   ├── kotlinc/
   │   └── bin/
   │       └── kotlinc-jvm
   └── jre/               ← optional if system JRE exists
       └── bin/
           └── java
   ```

6. **Environment Setup:**
   - Set `JAVA_HOME` to JRE path when invoking kotlinc
   - Add kotlinc/bin to PATH

```go
type KotlinDownloader struct {
    BaseDownloader
}

func (d *KotlinDownloader) Download(ctx context.Context, req domain.DownloadRequest, progressCb func(domain.DownloadProgress)) error {
    // Step 1: Check system JRE
    jrePath, err := d.DetectSystemJRE()
    needsJRE := (err != nil)

    // Step 2: Download Kotlin compiler
    kotlinURL, _ := d.GetDownloadURL(req)
    kotlinDest := filepath.Join(req.TargetDir, "kotlin-compiler.zip")

    err = d.DownloadFile(ctx, kotlinURL, kotlinDest, func(progress domain.DownloadProgress) {
        progress.Message = "Downloading Kotlin compiler..."
        progressCb(progress)
    })

    // Step 3: Extract Kotlin
    d.ExtractZip(kotlinDest, req.TargetDir)

    // Step 4: Download JRE if needed
    if needsJRE {
        jreURL := d.getJREDownloadURL(req.Platform, req.Arch)
        jreDest := filepath.Join(req.TargetDir, "jre.tar.gz")

        err = d.DownloadFile(ctx, jreURL, jreDest, func(progress domain.DownloadProgress) {
            progress.Message = "Downloading JRE..."
            progressCb(progress)
        })

        d.ExtractTarGz(jreDest, filepath.Join(req.TargetDir, "jre"))
    }

    return nil
}
```

#### SwiftDownloader (ADAPTER for Swift)

**Important nuances:**

1. **Swift WASM SDK:**
   - Download artifact bundle from https://github.com/swiftwasm/swift/releases
   - NOT full Swift toolchain (too large)
   - Just the WASM SDK addon (~50-100 MB)

2. **Requires System Swift:**
   - SwiftWasm SDK is addon, not standalone
   - Check if `swift` command exists (version >= 6.1)
   - If not found, show error with installation instructions

3. **SDK Installation:**
   - Download artifact bundle: `swift-wasm-6.1-RELEASE-wasm32-unknown-wasi.artifactbundle.zip`
   - Extract to `~/.cache/network-debugger/compilers/swift/`
   - SDK will be used with: `swift build --swift-sdk wasm32-unknown-wasi`

```go
type SwiftDownloader struct {
    BaseDownloader
}

func (d *SwiftDownloader) Verify(installPath string) error {
    // Check 1: System Swift exists
    cmd := exec.Command("swift", "--version")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("Swift not found. Please install Swift 6.1+ from swift.org")
    }

    // Check 2: Version >= 6.1
    // Parse version from output

    // Check 3: SDK is available
    cmd = exec.Command("swift", "sdk", "list")
    output, err = cmd.Output()
    if !strings.Contains(string(output), "wasm32-unknown-wasi") {
        return fmt.Errorf("SwiftWasm SDK not properly installed")
    }

    return nil
}
```

#### FileSystemCache (ADAPTER for cache management)

```go
type FileSystemCache struct {
    baseDir string
}

func NewFileSystemCache() *FileSystemCache {
    // Determine cache directory
    homeDir, _ := os.UserHomeDir()
    cacheDir := filepath.Join(homeDir, ".cache", "network-debugger")

    return &FileSystemCache{
        baseDir: cacheDir,
    }
}

func (c *FileSystemCache) GetCompilerPath(language string) (string, error) {
    path := filepath.Join(c.baseDir, "compilers", language)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return "", fmt.Errorf("compiler not installed: %s", language)
    }
    return path, nil
}

func (c *FileSystemCache) EnsureCacheDir() error {
    compilersDir := filepath.Join(c.baseDir, "compilers")
    return os.MkdirAll(compilersDir, 0755)
}

func (c *FileSystemCache) Clear(language string) error {
    path := filepath.Join(c.baseDir, "compilers", language)
    return os.RemoveAll(path)
}

func (c *FileSystemCache) GetCacheSize() (int64, error) {
    var size int64
    compilersDir := filepath.Join(c.baseDir, "compilers")

    err := filepath.Walk(compilersDir, func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() {
            size += info.Size()
        }
        return nil
    })

    return size, err
}
```

### Layer 4: API (HTTP + WebSocket)

```
internal/infrastructure/httpapi/
├── compiler_installation_handlers.go
└── router.go (update)
```

**Endpoints:**

```go
// GET /_api/v1/compilers/status
// Response:
{
  "compilers": [
    {
      "language": "zig",
      "version": "0.14.0",
      "status": "installed",
      "installedPath": "/Users/user/.cache/network-debugger/compilers/zig/current/zig",
      "size": 268435456,
      "downloadSize": 0
    },
    {
      "language": "kotlin",
      "version": "latest",
      "status": "not_installed",
      "installedPath": "",
      "size": 0,
      "downloadSize": 175000000
    }
  ]
}

// POST /_api/v1/compilers/{lang}/install
// Request:
{
  "version": "latest"  // or specific version like "0.14.0"
}
// Response: 202 Accepted
{
  "message": "Installation started",
  "websocket_url": "ws://localhost:9092/_api/v1/compilers/zig/install/progress"
}

// WebSocket: ws://localhost:9092/_api/v1/compilers/{lang}/install/progress
// Messages:
{
  "stage": "downloading",
  "bytesDownloaded": 12582912,
  "totalBytes": 52428800,
  "percentage": 24.0,
  "speed": 1048576,
  "message": "Downloaded 12 MB / 50 MB (1 MB/s)"
}
{
  "stage": "extracting",
  "percentage": 50.0,
  "message": "Extracting archive..."
}
{
  "stage": "verifying",
  "percentage": 90.0,
  "message": "Verifying installation..."
}
{
  "stage": "complete",
  "percentage": 100.0,
  "message": "Installation complete!"
}

// DELETE /_api/v1/compilers/{lang}/uninstall
// Response: 200 OK

// GET /_api/v1/compilers/{lang}/metadata
// Response:
{
  "language": "zig",
  "version": "0.14.0",
  "downloadURL": "https://ziglang.org/download/0.14.0/zig-macos-x86_64-0.14.0.tar.xz",
  "size": 52428800,
  "checksum": "abc123...",
  "releaseDate": "2024-11-01T00:00:00Z"
}
```

**WebSocket Implementation:**

```go
type ProgressBroadcaster struct {
    mu          sync.RWMutex
    subscribers map[string][]chan domain.DownloadProgress  // language -> channels
}

func (b *ProgressBroadcaster) Subscribe(language string) <-chan domain.DownloadProgress {
    b.mu.Lock()
    defer b.mu.Unlock()

    ch := make(chan domain.DownloadProgress, 10)
    b.subscribers[language] = append(b.subscribers[language], ch)
    return ch
}

func (b *ProgressBroadcaster) Broadcast(language string, progress domain.DownloadProgress) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    for _, ch := range b.subscribers[language] {
        select {
        case ch <- progress:
        default:
            // Skip if channel is full
        }
    }
}
```

## 🔧 Compilers Implementation

### ZigCompiler (NEW)

```go
type ZigCompiler struct {
    zigPath    string
    cache      domain.CacheManager
    downloader domain.CompilerDownloader
}

func (c *ZigCompiler) IsAvailable() bool {
    // Priority 1: Check system installation
    if cmd := exec.Command("zig", "version"); cmd.Run() == nil {
        c.zigPath = "zig"
        return true
    }

    // Priority 2: Check cache
    if cachePath, err := c.cache.GetCompilerPath("zig"); err == nil {
        // Look for zig binary in cache
        // Check both versioned directory and "current" symlink

        currentPath := filepath.Join(cachePath, "current", "zig")
        if _, err := os.Stat(currentPath); err == nil {
            c.zigPath = currentPath
            return true
        }
    }

    return false
}

func (c *ZigCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
    // Ensure zig is available
    if !c.IsAvailable() {
        return nil, fmt.Errorf("zig compiler not available. Please install via API or system package manager")
    }

    // Create workspace
    ws, err := NewWorkspace("zig-compile")
    defer ws.Cleanup()

    // Write source
    ws.WriteFile("main.zig", []byte(req.SourceCode))

    // Compile
    args := []string{
        "build-lib",
        "-target", "wasm32-freestanding",
        "-dynamic",
        "-rdynamic",
    }

    if req.Optimize {
        args = append(args, "-O", "ReleaseSmall")
    }

    args = append(args, "main.zig")

    output, err := ws.ExecuteCommand(ctx, c.zigPath, args...)

    // Read WASM
    wasm, err := ws.ReadFile("main.wasm")

    return &domain.CompileResult{
        WASMBinary: wasm,
        Logs:       strings.Split(string(output), "\n"),
        Duration:   time.Since(start),
        WASMSize:   int64(len(wasm)),
    }, nil
}
```

### KotlinCompiler (NEW)

**Important:** Needs Gradle wrapper for dependency management

```go
type KotlinCompiler struct {
    kotlinPath string
    javaHome   string
    cache      domain.CacheManager
    downloader domain.CompilerDownloader
}

func (c *KotlinCompiler) IsAvailable() bool {
    // Check 1: System Kotlin + JRE
    if _, err := exec.LookPath("kotlinc-jvm"); err == nil {
        if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
            c.kotlinPath = "kotlinc-jvm"
            c.javaHome = javaHome
            return true
        }
    }

    // Check 2: Cache installation
    if cachePath, err := c.cache.GetCompilerPath("kotlin"); err == nil {
        kotlinBin := filepath.Join(cachePath, "kotlinc", "bin", "kotlinc-jvm")
        jreDir := filepath.Join(cachePath, "jre")

        if _, err := os.Stat(kotlinBin); err == nil {
            c.kotlinPath = kotlinBin

            // Check if we have bundled JRE or use system
            if _, err := os.Stat(jreDir); err == nil {
                c.javaHome = jreDir
            } else {
                c.javaHome = os.Getenv("JAVA_HOME")
            }

            return c.javaHome != ""
        }
    }

    return false
}

func (c *KotlinCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
    // Note: Kotlin/Wasm typically uses Gradle
    // For simple scripts, we can use kotlinc-wasm directly
    // But for dependencies, need build.gradle.kts

    ws, err := NewWorkspace("kotlin-compile")
    defer ws.Cleanup()

    // Write source
    ws.WriteFile("Main.kt", []byte(req.SourceCode))

    // For now, use simple kotlinc-wasm compilation
    // TODO: Add Gradle wrapper support for dependencies

    args := []string{
        "-Xwasm",
        "-d", "output.wasm",
        "Main.kt",
    }

    // Set JAVA_HOME for kotlinc
    cmd := exec.CommandContext(ctx, c.kotlinPath, args...)
    cmd.Env = append(os.Environ(), "JAVA_HOME="+c.javaHome)
    cmd.Dir = ws.Path

    output, err := cmd.CombinedOutput()

    // Read WASM
    wasm, err := ws.ReadFile("output.wasm")

    return &domain.CompileResult{
        WASMBinary: wasm,
        Logs:       strings.Split(string(output), "\n"),
        Duration:   time.Since(start),
        WASMSize:   int64(len(wasm)),
    }, nil
}
```

### SwiftCompiler (NEW)

```go
type SwiftCompiler struct {
    swiftPath string
    sdkPath   string
    cache     domain.CacheManager
    downloader domain.CompilerDownloader
}

func (c *SwiftCompiler) IsAvailable() bool {
    // Swift MUST be system-installed (too large to bundle)
    cmd := exec.Command("swift", "--version")
    if err := cmd.Run(); err != nil {
        return false
    }

    // Check if WASM SDK is installed (system or cache)
    cmd = exec.Command("swift", "sdk", "list")
    output, err := cmd.Output()

    if strings.Contains(string(output), "wasm32-unknown-wasi") {
        // SDK found in system
        c.swiftPath = "swift"
        return true
    }

    // Check cache for SDK
    if cachePath, err := c.cache.GetCompilerPath("swift"); err == nil {
        c.sdkPath = filepath.Join(cachePath, "wasm32-unknown-wasi.artifactbundle")
        if _, err := os.Stat(c.sdkPath); err == nil {
            c.swiftPath = "swift"
            return true
        }
    }

    return false
}
```

## 📋 Implementation Checklist

### Phase 1: Domain Layer
- [ ] Create `compiler_manager.go` with interfaces
- [ ] Create `compiler_downloader.go` with interfaces
- [ ] Create `compiler_cache.go` with interfaces
- [ ] Add error types for installation failures

### Phase 2: Infrastructure - Downloaders
- [ ] Create `base.go` with BaseDownloader
- [ ] Implement `DownloadFile` with progress tracking
- [ ] Implement `ExtractTarXz`
- [ ] Implement `ExtractZip`
- [ ] Implement `VerifyChecksum`
- [ ] Create `zig_downloader.go`
- [ ] Create `kotlin_downloader.go` with JRE detection
- [ ] Create `swift_downloader.go`

### Phase 3: Infrastructure - Cache
- [ ] Create `filesystem_cache.go`
- [ ] Implement cache directory management
- [ ] Implement cache size calculation
- [ ] Implement cache cleanup

### Phase 4: Infrastructure - Compilers
- [ ] Create `zig.go` with cache integration
- [ ] Create `kotlin.go` with JRE management
- [ ] Create `swift.go` with SDK support
- [ ] Update compiler detection logic

### Phase 5: UseCase Layer
- [ ] Create `compiler_installation_service.go`
- [ ] Implement `InstallCompiler` with orchestration
- [ ] Implement `UninstallCompiler`
- [ ] Implement `CheckAvailability`
- [ ] Implement `GetMetadata`

### Phase 6: API Layer
- [ ] Create `compiler_installation_handlers.go`
- [ ] Implement `/compilers/status` endpoint
- [ ] Implement `/compilers/{lang}/install` endpoint
- [ ] Implement WebSocket for progress
- [ ] Implement `/compilers/{lang}/uninstall` endpoint
- [ ] Update router with new routes

### Phase 7: Main Integration
- [ ] Register installation service in main.go
- [ ] Register downloaders for Zig, Kotlin, Swift
- [ ] Initialize cache manager
- [ ] Connect to compilation service

### Phase 8: Examples & Documentation
- [ ] Create Zig examples
- [ ] Create Kotlin examples
- [ ] Create Swift examples
- [ ] Update main documentation
- [ ] Create installation guides

## 🚨 Important Nuances & Gotchas

### 1. Platform-Specific Issues

**macOS Gatekeeper:**
- Downloaded binaries might be quarantined
- Need to run: `xattr -dr com.apple.quarantine {compiler_path}`
- Do this automatically after extraction

**Windows:**
- Extract to path without spaces
- Handle `.exe` extensions
- May need to set execution policy

**Linux:**
- Check for `libc` compatibility
- Some distros need additional dependencies

### 2. Concurrent Downloads

**Problem:** Multiple users install same compiler simultaneously

**Solution:**
```go
var installLocks = make(map[string]*sync.Mutex)
var installLocksMu sync.Mutex

func (s *CompilerInstallationService) InstallCompiler(...) error {
    // Get lock for this language
    installLocksMu.Lock()
    if _, exists := installLocks[language]; !exists {
        installLocks[language] = &sync.Mutex{}
    }
    lock := installLocks[language]
    installLocksMu.Unlock()

    // Acquire lock
    lock.Lock()
    defer lock.Unlock()

    // Check again if installed (might have been installed by concurrent request)
    if s.IsInstalled(language) {
        return nil
    }

    // Proceed with installation...
}
```

### 3. Partial Download Recovery

**Problem:** Download fails midway, leaves partial file

**Solution:**
- Download to `.tmp` file first
- Only rename to final name after successful download + verification
- Always cleanup `.tmp` files on startup

```go
func (b *BaseDownloader) DownloadFile(...) error {
    tmpDest := dest + ".tmp"
    defer os.Remove(tmpDest)  // Cleanup on error

    // Download to tmpDest
    // ...

    // Verify checksum
    if err := b.VerifyChecksum(tmpDest, checksum); err != nil {
        return err
    }

    // Rename to final destination
    return os.Rename(tmpDest, dest)
}
```

### 4. Disk Space Check

**Problem:** User has insufficient disk space

**Solution:**
```go
func (s *CompilerInstallationService) InstallCompiler(...) error {
    // Get metadata
    metadata, err := s.downloaders[language].GetMetadata(...)

    // Check available disk space
    var stat unix.Statfs_t
    unix.Statfs(cacheDir, &stat)
    available := stat.Bavail * uint64(stat.Bsize)

    // Need 2x size (compressed + extracted)
    required := metadata.Size * 2

    if available < uint64(required) {
        return fmt.Errorf("insufficient disk space: need %s, have %s",
            formatBytes(required), formatBytes(int64(available)))
    }

    // Proceed...
}
```

### 5. Network Timeout & Retry

**Problem:** Large downloads may timeout on slow connections

**Solution:**
```go
func (b *BaseDownloader) DownloadFile(...) error {
    maxRetries := 3

    for attempt := 0; attempt < maxRetries; attempt++ {
        err := b.downloadWithTimeout(ctx, url, dest, 10*time.Minute, progressCb)
        if err == nil {
            return nil
        }

        if attempt < maxRetries-1 {
            time.Sleep(time.Second * time.Duration(attempt+1))
        }
    }

    return fmt.Errorf("download failed after %d attempts", maxRetries)
}
```

### 6. Version Management

**Problem:** Multiple versions of same compiler

**Current approach:**
- Keep only one version at a time
- Always install to `{cache}/compilers/{lang}/current/`
- If user wants different version, replace existing

**Future improvement:**
- Support multiple versions: `{cache}/compilers/{lang}/{version}/`
- Symlink `current` to active version
- Allow switching versions

## 🎯 Success Criteria

- [ ] User can install Zig with one click
- [ ] Download progress shows in real-time via WebSocket
- [ ] Zig compiles code successfully after installation
- [ ] Kotlin installs with JRE management
- [ ] Swift SDK installs and integrates with system Swift
- [ ] Cache size is displayed correctly
- [ ] Uninstall removes all files
- [ ] Concurrent installs don't corrupt cache
- [ ] Network errors are handled gracefully
- [ ] Binary size impact is minimal (downloaders are lightweight)

## 📚 Documentation to Create

1. **User Guide:** How to use compiler installation API
2. **Admin Guide:** Cache management, troubleshooting
3. **Developer Guide:** How to add new compiler downloaders
4. **API Reference:** All endpoints with examples

## 🔄 Future Enhancements

1. **Version Pinning:** Allow users to specify exact versions
2. **Auto-Update:** Check for new compiler versions
3. **Bandwidth Limiting:** Configurable download speed limit
4. **Proxy Support:** Download through HTTP proxy
5. **Offline Mode:** Pre-download compilers for offline use
6. **CDN Fallback:** Multiple download sources
7. **Compression:** Use Brotli/zstd for smaller downloads

---

## 🔄 PHASE 2: Migration of Existing Embedded Compilers

### Current Embedded Compilers to Migrate

#### 🦀 1. Rust (Cargo + rustc + wasm32-unknown-unknown)
**Current Status:** ❌ Embedded (requires system installation)
**Size:** ~600 MB (minimal) to 2.9 GB (full)

**Migration Plan:**
- Create `RustDownloader` adapter
- Download from: https://static.rust-lang.org/dist/
- Components needed:
  - rustc compiler
  - cargo build tool
  - wasm32-unknown-unknown target
- Minimal installation: ~600 MB
- Full toolchain: ~1.6 GB

**Downloader Strategy:**
```go
type RustDownloader struct {
    *BaseDownloader
    rustupURL string  // https://static.rust-lang.org/rustup/dist/
}

func (r *RustDownloader) Download(...) error {
    // 1. Download rustup-init
    // 2. Install minimal profile (no docs, no rustfmt)
    // 3. Add wasm32-unknown-unknown target
    // 4. Verify with: rustc --version
}
```

**Cache Structure:**
```
~/.cache/network-debugger/compilers/rust/
├── bin/
│   ├── rustc
│   ├── cargo
│   └── rustup
├── lib/
│   └── rustlib/
│       └── wasm32-unknown-unknown/
└── version.txt
```

---

#### 🐹 2. TinyGo

**Current Status:** ❌ Embedded (requires system installation)
**Size:** 180-350 MB depending on platform

**Migration Plan:**
- Create `TinyGoDownloader` adapter
- Download from: https://github.com/tinygo-org/tinygo/releases
- Pre-built binaries available for all platforms
- Includes LLVM toolchain + WASI libraries

**Downloader Strategy:**
```go
type TinyGoDownloader struct {
    *BaseDownloader
}

func (t *TinyGoDownloader) GetDownloadURL(req DownloadRequest) (string, error) {
    version := "0.33.0"  // Latest stable
    // Format: https://github.com/tinygo-org/tinygo/releases/download/v{version}/tinygo{version}.{os}-{arch}.tar.gz

    osName := map[string]string{
        "darwin": "darwin",
        "linux":  "linux",
        "windows": "windows",
    }[req.Platform]

    archName := map[string]string{
        "amd64": "amd64",
        "arm64": "arm64",
    }[req.Arch]

    return fmt.Sprintf("https://github.com/tinygo-org/tinygo/releases/download/v%s/tinygo%s.%s-%s.tar.gz",
        version, version, osName, archName), nil
}
```

**Cache Structure:**
```
~/.cache/network-debugger/compilers/tinygo/
├── bin/
│   └── tinygo
├── lib/
├── targets/
└── pkg/
```

---

#### 🌐 3. AssemblyScript (asc compiler)

**Current Status:** ❌ Embedded (requires Node.js + npm)
**Size:** ~40-60 MB with dependencies

**Migration Plan:**
- Create `AssemblyScriptDownloader` adapter
- **Challenge:** AssemblyScript is an npm package
- **Solution:** Bundle standalone asc binary OR download Node.js + npm package

**Option A: Pre-built Standalone Binary (Preferred)**
```go
type AssemblyScriptDownloader struct {
    *BaseDownloader
}

// Use @assemblyscript/asc compiled to standalone binary
// Or bundle with pkg/nexe
```

**Option B: Download Node.js + npm install**
```go
func (a *AssemblyScriptDownloader) Download(...) error {
    // 1. Download Node.js portable (~30 MB)
    // 2. npm install assemblyscript (~25 MB)
    // 3. Create wrapper script to call asc
}
```

**Cache Structure:**
```
~/.cache/network-debugger/compilers/assemblyscript/
├── node/              # Portable Node.js
│   └── bin/node
├── node_modules/
│   └── assemblyscript/
│       └── bin/asc
└── asc.sh            # Wrapper script
```

---

#### 🔧 4. C/C++ (WASI-SDK / clang)

**Current Status:** ⚠️ Partially embedded (fallback to system clang)
**Size:** 130-500 MB depending on platform

**Migration Plan:**
- Create `WASISDKDownloader` adapter
- Download official WASI-SDK releases
- Include clang + wasi-libc + compiler-rt

**Downloader Strategy:**
```go
type WASISDKDownloader struct {
    *BaseDownloader
}

func (w *WASISDKDownloader) GetDownloadURL(req DownloadRequest) (string, error) {
    version := "24"  // Latest stable WASI-SDK

    // Format: https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-{version}/wasi-sdk-{version}.0-{os}.tar.gz

    osArch := map[string]string{
        "darwin-amd64": "macos",
        "darwin-arm64": "macos",
        "linux-amd64":  "linux",
        "windows-amd64": "mingw",
    }[req.Platform + "-" + req.Arch]

    return fmt.Sprintf("https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-%s/wasi-sdk-%s.0-%s.tar.gz",
        version, version, osArch), nil
}
```

**Cache Structure:**
```
~/.cache/network-debugger/compilers/wasi-sdk/
├── bin/
│   ├── clang
│   ├── clang++
│   └── wasm-ld
├── lib/
│   └── clang/
├── share/
│   └── wasi-sysroot/
```

---

#### 🐍 5. Python (via RustPython)

**Current Status:** ⚠️ Hybrid (requires Rust toolchain)
**Size:** Depends on Rust (~600 MB minimal)

**Migration Plan:**
- **Option A:** Keep as wrapper around Rust (current approach)
- **Option B:** Pre-compile RustPython to WASM module, embed in binary
- **Recommendation:** Depends on Rust migration, keep current approach

**After Rust Migration:**
```go
type PythonCompiler struct {
    rustCompiler *RustCompiler  // Now uses cached Rust
}

func (p *PythonCompiler) IsAvailable() bool {
    // Check if Rust is available in cache
    return p.rustCompiler.IsAvailable()
}
```

---

### Migration Checklist

#### Phase 2.1: Rust Migration
- [ ] Create `RustDownloader` in `internal/features/scripting/infrastructure/downloaders/rust_downloader.go`
- [ ] Implement rustup-init download and installation
- [ ] Add wasm32-unknown-unknown target installation
- [ ] Update `RustCompiler` to use cache
- [ ] Test compilation after cache installation
- [ ] Size: ~600 MB

#### Phase 2.2: TinyGo Migration
- [ ] Create `TinyGoDownloader` in `downloaders/tinygo_downloader.go`
- [ ] Parse GitHub releases API for latest version
- [ ] Implement platform-specific URL generation
- [ ] Update `TinyGoCompiler` to use cache
- [ ] Test Go→WASM compilation
- [ ] Size: ~200-350 MB

#### Phase 2.3: AssemblyScript Migration
- [ ] Decide on standalone binary vs Node.js approach
- [ ] Create `AssemblyScriptDownloader` in `downloaders/assemblyscript_downloader.go`
- [ ] If Node.js: download portable Node + npm install
- [ ] If standalone: research pkg/nexe bundling
- [ ] Update `AssemblyScriptCompiler` to use cache
- [ ] Test TypeScript→WASM compilation
- [ ] Size: ~40-60 MB

#### Phase 2.4: WASI-SDK Migration
- [ ] Create `WASISDKDownloader` in `downloaders/wasisdk_downloader.go`
- [ ] Download from GitHub releases
- [ ] Extract and verify clang binary
- [ ] Update `CCPPCompiler` to prefer cached WASI-SDK
- [ ] Remove system clang fallback (or keep as backup)
- [ ] Size: ~130-500 MB

#### Phase 2.5: Python Dependency Update
- [ ] No separate downloader needed
- [ ] Update `PythonCompiler` to check Rust cache availability
- [ ] Add clear error message if Rust not installed
- [ ] Consider: Pre-built RustPython WASM module

---

## 🎨 PHASE 3: Frontend UI (Flutter/Dart)

### UI Requirements

**Core Features:**
1. **Compiler List View**
   - Show all 8 compilers
   - Status indicator (installed/not installed/installing)
   - Download size
   - Installed size
   - Version

2. **Installation Progress**
   - Real-time progress bar
   - Download speed
   - Stage indicator (downloading/extracting/verifying)
   - Estimated time remaining

3. **Compiler Details**
   - Description
   - Language features
   - Example code
   - Documentation links

4. **Cache Management**
   - Total cache size
   - Per-compiler size
   - Clear cache button
   - Storage location

### UI Architecture (Clean Architecture in Flutter)

```
frontend/lib/features/compiler_management/
├── domain/
│   ├── entities/
│   │   ├── compiler_info.dart
│   │   ├── compiler_status.dart
│   │   └── download_progress.dart
│   ├── repositories/
│   │   └── compiler_repository.dart          # PORT (interface)
│   └── usecases/
│       ├── install_compiler.dart
│       ├── uninstall_compiler.dart
│       ├── get_compilers_status.dart
│       └── watch_installation_progress.dart
├── data/
│   ├── datasources/
│   │   ├── compiler_remote_datasource.dart   # API calls
│   │   └── compiler_local_datasource.dart    # Local cache
│   ├── models/
│   │   ├── compiler_info_model.dart
│   │   └── download_progress_model.dart
│   └── repositories/
│       └── compiler_repository_impl.dart     # ADAPTER
├── presentation/
│   ├── stores/                                # MobX stores
│   │   ├── compiler_list_store.dart
│   │   └── compiler_details_store.dart
│   ├── widgets/
│   │   ├── compiler_card.dart
│   │   ├── installation_progress.dart
│   │   ├── compiler_details_panel.dart
│   │   └── cache_stats.dart
│   └── pages/
│       └── compilers_page.dart
```

### UI Components

#### 1. CompilerCard Widget
```dart
class CompilerCard extends StatelessWidget {
  final CompilerInfo compiler;
  final VoidCallback onInstall;
  final VoidCallback onUninstall;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Column(
        children: [
          // Icon + Name + Status badge
          ListTile(
            leading: CompilerIcon(language: compiler.language),
            title: Text(compiler.name),
            subtitle: Text(compiler.version),
            trailing: StatusBadge(status: compiler.status),
          ),

          // Size info
          if (compiler.status == CompilerStatus.notInstalled)
            Text('Download size: ${formatBytes(compiler.downloadSize)}'),

          if (compiler.status == CompilerStatus.installed)
            Text('Installed: ${formatBytes(compiler.size)}'),

          // Progress bar (if installing)
          if (compiler.status == CompilerStatus.installing)
            InstallationProgress(compiler: compiler),

          // Actions
          ButtonBar(
            children: [
              if (compiler.status == CompilerStatus.notInstalled)
                ElevatedButton(
                  onPressed: onInstall,
                  child: Text('Install'),
                ),

              if (compiler.status == CompilerStatus.installed)
                OutlinedButton(
                  onPressed: onUninstall,
                  child: Text('Uninstall'),
                ),
            ],
          ),
        ],
      ),
    );
  }
}
```

#### 2. InstallationProgress Widget (SSE)
```dart
class InstallationProgress extends StatefulWidget {
  final CompilerInfo compiler;

  @override
  _InstallationProgressState createState() => _InstallationProgressState();
}

class _InstallationProgressState extends State<InstallationProgress> {
  late Stream<DownloadProgress> _progressStream;

  @override
  void initState() {
    super.initState();
    _progressStream = _watchProgress();
  }

  Stream<DownloadProgress> _watchProgress() {
    // Connect to SSE endpoint
    return EventSource(
      'http://localhost:9092/_api/v1/compilers/${widget.compiler.language}/progress'
    ).listen();
  }

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<DownloadProgress>(
      stream: _progressStream,
      builder: (context, snapshot) {
        if (!snapshot.hasData) return CircularProgressIndicator();

        final progress = snapshot.data!;

        return Column(
          children: [
            LinearProgressIndicator(value: progress.percentage / 100),

            Text('${progress.stage}: ${progress.percentage.toStringAsFixed(1)}%'),

            if (progress.speed > 0)
              Text('Speed: ${formatBytes(progress.speed)}/s'),

            Text(progress.message),
          ],
        );
      },
    );
  }
}
```

#### 3. CompilersPage
```dart
class CompilersPage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final store = context.read<CompilerListStore>();

    return Scaffold(
      appBar: AppBar(
        title: Text('Compilers'),
        actions: [
          // Cache stats
          Observer(
            builder: (_) => Text('Cache: ${formatBytes(store.totalCacheSize)}'),
          ),

          // Clear cache button
          IconButton(
            icon: Icon(Icons.delete_outline),
            onPressed: store.clearCache,
          ),
        ],
      ),

      body: Observer(
        builder: (_) {
          if (store.isLoading) return CircularProgressIndicator();

          return GridView.builder(
            gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              childAspectRatio: 1.5,
            ),
            itemCount: store.compilers.length,
            itemBuilder: (context, index) {
              final compiler = store.compilers[index];

              return CompilerCard(
                compiler: compiler,
                onInstall: () => store.install(compiler.language),
                onUninstall: () => store.uninstall(compiler.language),
              );
            },
          );
        },
      ),
    );
  }
}
```

#### 4. CompilerListStore (MobX)
```dart
@MakeObservable()
class CompilerListStore {
  final CompilerRepository _repository;

  @observable
  ObservableList<CompilerInfo> compilers = ObservableList();

  @observable
  bool isLoading = false;

  @observable
  int totalCacheSize = 0;

  @action
  Future<void> loadCompilers() async {
    isLoading = true;

    try {
      final response = await _repository.getCompilersStatus();
      compilers = ObservableList.of(response.compilers);
      totalCacheSize = response.cacheSize;
    } finally {
      isLoading = false;
    }
  }

  @action
  Future<void> install(String language) async {
    // Find compiler and update status
    final compiler = compilers.firstWhere((c) => c.language == language);
    compiler.status = CompilerStatus.installing;

    try {
      await _repository.installCompiler(language);
      await loadCompilers(); // Refresh list
    } catch (e) {
      compiler.status = CompilerStatus.error;
      compiler.error = e.toString();
    }
  }

  @action
  Future<void> uninstall(String language) async {
    await _repository.uninstallCompiler(language);
    await loadCompilers();
  }

  @action
  Future<void> clearCache() async {
    await _repository.clearAllCache();
    await loadCompilers();
  }
}
```

### UI Layout

```
┌────────────────────────────────────────────────────────────┐
│  Compilers                        Cache: 1.2 GB  🗑️       │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │   🦀 Rust   │  │  🐹 TinyGo  │  │ 🌐 AssemblyS│      │
│  │             │  │             │  │   cript      │      │
│  │ ✅ Installed│  │ ❌ Not Inst.│  │ ⚙️ Installing│      │
│  │ v1.83.0     │  │ v0.33.0     │  │ v0.27.0      │      │
│  │             │  │             │  │              │      │
│  │ 1.2 GB      │  │ Download:   │  │ ▓▓▓▓▓▓░░░░   │      │
│  │             │  │ 250 MB      │  │ 65%          │      │
│  │             │  │             │  │ 15 MB/s      │      │
│  │  Uninstall  │  │   Install   │  │ Extracting...│      │
│  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │   🔧 Zig    │  │  ☕ Kotlin  │  │  🍎 Swift    │      │
│  │             │  │             │  │              │      │
│  │ ✅ Installed│  │ ❌ Not Inst.│  │ ✅ Installed │      │
│  │ ...         │  │ ...         │  │ ...          │      │
│  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Frontend Implementation Checklist

- [ ] Create `compiler_management` feature folder structure
- [ ] Domain layer: entities, repositories (interfaces), use cases
- [ ] Data layer: models, datasources (HTTP + SSE), repository implementation
- [ ] Presentation layer: stores (MobX), widgets, pages
- [ ] API client with SSE support for progress
- [ ] Error handling and retry logic
- [ ] Caching of compiler list
- [ ] Beautiful UI with animations
- [ ] Dark mode support
- [ ] Responsive design (mobile + desktop)

---

## 📋 Updated Complete Checklist

### ✅ PHASE 1: New Compilers (Zig, Kotlin, Swift) - COMPLETED
- [x] Domain Layer (3 interfaces)
- [x] Infrastructure Downloaders (4 files)
- [x] Infrastructure Cache (1 file)
- [x] Infrastructure Compilers (3 files)
- [x] UseCase Layer (1 service)
- [x] API Layer (6 endpoints)
- [x] Main Integration
- [x] Documentation

### 🔄 PHASE 2: Migrate Existing Compilers - IN PROGRESS
- [ ] Rust Migration (~600 MB)
  - [ ] RustDownloader implementation
  - [ ] Update RustCompiler to use cache
  - [ ] Test compilation
- [ ] TinyGo Migration (~250 MB)
  - [ ] TinyGoDownloader implementation
  - [ ] Update TinyGoCompiler to use cache
  - [ ] Test compilation
- [ ] AssemblyScript Migration (~50 MB)
  - [ ] AssemblyScriptDownloader implementation
  - [ ] Update AssemblyScriptCompiler to use cache
  - [ ] Test compilation
- [ ] WASI-SDK Migration (~150-500 MB)
  - [ ] WASISDKDownloader implementation
  - [ ] Update CCPPCompiler to use cache
  - [ ] Test compilation
- [ ] Python (depends on Rust)
  - [ ] Update PythonCompiler to check Rust cache
  - [ ] Test compilation

### 🎨 PHASE 3: Frontend UI - TODO
- [ ] Setup feature folder structure
- [ ] Domain layer (entities, repositories, use cases)
- [ ] Data layer (models, datasources, repository impl)
- [ ] Presentation layer (stores, widgets, pages)
- [ ] SSE integration for real-time progress
- [ ] Beautiful responsive UI
- [ ] Dark mode support
- [ ] Error handling
- [ ] Testing

---

## 🎯 Total System Overview After All Phases

**Compilers:**
- 8 languages total (5 migrated + 3 new)
- All use download-on-demand
- Cache in `~/.cache/network-debugger/compilers/`

**Total Cache Sizes:**
- Rust: ~600 MB
- TinyGo: ~250 MB
- AssemblyScript: ~50 MB
- C/C++ (WASI-SDK): ~180 MB
- Zig: ~80 MB
- Kotlin: ~200 MB (without JRE) / ~400 MB (with JRE)
- Swift: ~75 MB (SDK only)
- **Total Maximum:** ~1.8 GB (all installed)
- **Typical Usage:** ~500 MB (2-3 compilers)

**UI Features:**
- Visual compiler management
- One-click installation
- Real-time progress tracking
- Cache management
- Beautiful, responsive design

---

This comprehensive plan covers migration + UI! 🚀
