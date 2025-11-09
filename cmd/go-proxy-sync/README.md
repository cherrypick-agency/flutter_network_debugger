# go-proxy-sync

CLI tool for syncing script projects with Network Debugger server. Designed for IDE plugins and automated workflows.

## Installation

```bash
# Build from source
cd cmd/go-proxy-sync
go build -o go-proxy-sync .

# Or install globally
go install .
```

## Usage

### Upload a project

Upload a local project directory to the server:

```bash
go-proxy-sync upload --script-id=<script-id> --dir=./my-rust-project
```

This will:
1. Create a ZIP archive of all source files
2. Upload to `/_api/v1/scripts/{id}/upload-project`
3. Server extracts and stores files in Dependencies map

### Download a project

Download a project from the server:

```bash
go-proxy-sync download --script-id=<script-id> --dir=./downloaded-project
```

This will:
1. Download ZIP from `/_api/v1/scripts/{id}/download-project`
2. Extract all files to the specified directory
3. Preserve folder structure

### Watch for changes (Auto-sync)

Watch a directory and automatically upload on changes:

```bash
go-proxy-sync watch --script-id=<script-id> --dir=./my-project
```

Features:
- Watches all source files (`.rs`, `.go`, `.ts`, `.toml`, etc.)
- Debounces uploads (2 second delay)
- Skips build artifacts (`target/`, `node_modules/`)
- Skips hidden files and directories

### List project files

List all files stored in the server:

```bash
go-proxy-sync files --script-id=<script-id>
```

## Options

- `--server` - Server URL (default: `http://localhost:9092`)
- `--script-id` - Script ID (required)
- `--dir` - Project directory (default: current directory)
- `-v` - Verbose output

## IDE Plugin Integration

This tool is designed to be called by IDE plugins:

### VSCode Extension Example

```typescript
import { exec } from 'child_process';

async function uploadProject(scriptId: string, projectPath: string) {
  return new Promise((resolve, reject) => {
    exec(`go-proxy-sync upload --script-id=${scriptId} --dir=${projectPath}`,
      (error, stdout, stderr) => {
        if (error) reject(error);
        else resolve(stdout);
      }
    );
  });
}
```

### IntelliJ Plugin Example

```kotlin
import java.io.BufferedReader

fun uploadProject(scriptId: String, projectPath: String): String {
    val process = ProcessBuilder(
        "go-proxy-sync", "upload",
        "--script-id=$scriptId",
        "--dir=$projectPath"
    ).start()

    return process.inputStream.bufferedReader().use(BufferedReader::readText)
}
```

## Supported File Types

The tool automatically includes:
- **Rust:** `.rs`, `Cargo.toml`, `Cargo.lock`
- **Go:** `.go`, `go.mod`, `go.sum`
- **TypeScript/JavaScript:** `.ts`, `.js`, `package.json`, `package-lock.json`
- **Dart:** `.dart`, `pubspec.yaml`, `pubspec.lock`
- **Python:** `.py`, `requirements.txt`
- **C/C++:** `.c`, `.cpp`, `.h`, `.hpp`
- **Zig:** `.zig`
- **Kotlin:** `.kt`
- **Swift:** `.swift`

## Exclusions

The following are automatically excluded:
- Hidden files (`.git`, `.vscode`, etc.)
- Build artifacts (`target/`, `build/`, `dist/`)
- Dependencies (`node_modules/`, `vendor/`)
- Binary files

## Workflow Example

```bash
# 1. Create a new script in the UI and get the script ID
SCRIPT_ID="abc123"

# 2. Develop your project locally
cd ~/projects/my-rust-script
cargo init --lib

# 3. Upload to debugger
go-proxy-sync upload --script-id=$SCRIPT_ID --dir=.

# 4. Compile via API or UI
curl -X POST http://localhost:9092/_api/v1/scripts/$SCRIPT_ID/compile

# 5. Test and iterate
# Edit files locally, then upload again
```

## Watch Mode Workflow

Perfect for rapid development:

```bash
# Start watch mode
go-proxy-sync watch --script-id=abc123 --dir=./my-project

# Now edit files in your IDE
# Tool automatically uploads changes after 2 second debounce
# Compile via UI and see results immediately
```

## Error Handling

The tool returns non-zero exit codes on errors:
- Exit code 1: General error (network, file I/O, etc.)
- Check stderr for error messages

## Requirements

- Go 1.22+ (for building)
- Network Debugger server running
- Valid script ID from the server

## Contributing

This tool is part of the Network Debugger project. Report issues or contribute at the main repository.
