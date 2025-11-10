# Scripts Feature

## Status: ✅ Ready to Use

The Scripts management UI feature is fully implemented and ready for production use.

## Features

### Complete Clean Architecture Implementation
- ✅ Domain Layer: Entities with Freezed, Repository interface, Use Cases
- ✅ Data Layer: API Service, Repository Implementation
- ✅ Application Layer: MobX Stores (ScriptsStore, ScriptEditorStore)
- ✅ Presentation Layer: Full UI with pages and widgets

### User Interface
- ✅ Scripts list page with search and filters
- ✅ Script editor dialog with 4 tabs:
  - **Code Tab**: Monaco editor with syntax highlighting for 10+ languages
  - **Settings Tab**: Configure runtime, trigger type, priority, timeouts
  - **Match Rules Tab**: Visual builder for filtering HTTP requests
  - **Test Tab**: Request builder with live testing
- ✅ WASM file upload with drag-n-drop support
- ✅ Enable/disable toggle for each script
- ✅ Real-time search by script name/description
- ✅ Filters by runtime, trigger type, and status

### Technical Stack
- **State Management**: MobX with reactive stores
- **Code Generation**: Freezed for immutable models
- **Code Editor**: flutter_monaco with multi-language support
- **Architecture**: Clean Architecture + Domain-Driven Design
- **Dependency Injection**: GetIt service locator

## Usage

### Navigation
Access the Scripts page via route: `/scripts`

### Creating a Script
1. Click "Create Script" button
2. Choose runtime (Extism WASM or Dart)
3. Select creation mode (see below)
4. Configure settings (trigger type, priority, etc.)
5. Optionally add match rules to filter requests
6. Test the script with sample requests
7. Save

### Script Creation Modes

For **Extism WASM** runtime, you have two creation modes:

#### 1. Write Source Code (`writeSource`)
- Write source code directly in multi-file editor
- Supports multiple languages: Rust, Go, TypeScript, Python, etc.
- Create multiple files and organize your project structure
- Click green **Compile** button to generate WASM
- Requires compiler to be installed (see Compiler Management)
- **Alternative**: Use "Upload WASM" button in toolbar to upload pre-compiled WASM file

#### 2. Import ZIP Project (`importZip`)
- Upload a ZIP file containing your project source files
- Supports multi-file projects with dependencies
- Files are automatically extracted into the multi-file editor
- Click green **Compile** button to build the project
- Requires compiler to be installed

For **Dart** runtime, you always use the multi-file editor to write your script code.

### Compilation Process

When using `writeSource` or `importZip` modes:

1. **Green Compile Button**: Always visible and active in the editor toolbar
2. **Without Compiler**: Clicking shows dialog with link to Compiler Management page
3. **With Compiler**: Compiles source code to WASM on backend
4. **Keyboard Shortcut**: Press `Ctrl+Shift+C` (or `Cmd+Shift+C` on Mac) to compile
5. **Warning Banner**: Shows "Source code detected" or "Project files detected" until compilation

### Compiler Management

Navigate to `/compilers` page to:
- View available compilers (Rust, Go, Python, etc.)
- Download and install compilers
- Manage compiler versions
- Check installation status

Access via:
- Settings form in script editor when compiler missing
- Dialog shown when clicking Compile without compiler installed

### Testing Scripts
Use the Test tab to:
- Build sample HTTP requests
- Execute scripts against test data
- View modifications and logs
- Debug before deployment

## Implementation Details

### Domain Entities (with Freezed)
- `Script` - Main script entity
- `MatchRules` - HTTP request filtering rules
- `ScriptConfig` - Execution configuration
- `TestRequest` - Test request structure
- `ScriptTestResult` - Test execution results

### Use Cases
- `GetScriptsUseCase` - Retrieve all scripts
- `CreateScriptUseCase` - Create new script
- `UpdateScriptUseCase` - Update existing script
- `DeleteScriptUseCase` - Delete script
- `ToggleScriptUseCase` - Enable/disable script
- `TestScriptUseCase` - Test script execution

### Stores
- `ScriptsStore` - Main store for scripts list, search, filters
- `ScriptEditorStore` - Store for script editor form state and validation

## Files Structure

```
lib/features/scripts/
├── domain/
│   ├── entities/          # Freezed entities
│   ├── repositories/      # Repository interfaces
│   └── usecases/         # Business logic use cases
├── data/
│   ├── services/         # API service
│   └── repositories/     # Repository implementations
├── application/
│   └── stores/           # MobX stores
└── presentation/
    ├── pages/            # Main pages
    └── widgets/          # Reusable widgets
```

## API Integration

Backend endpoints:
- `GET /_api/v1/scripts` - List all scripts
- `GET /_api/v1/scripts/:id` - Get script by ID
- `POST /_api/v1/scripts` - Create script
- `PUT /_api/v1/scripts/:id` - Update script
- `DELETE /_api/v1/scripts/:id` - Delete script
- `PATCH /_api/v1/scripts/:id/toggle` - Toggle enabled status
- `POST /_api/v1/scripts/test` - Test script execution
- `POST /_api/v1/scripts/:id/compile` - Compile source code to WASM

## Development

### Running Code Generation
```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

### Linting
```bash
flutter analyze
```

All Freezed-related analyzer warnings are properly handled in `analysis_options.yaml`.
