import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:file_picker/file_picker.dart';
import 'package:dio/dio.dart';
import '../../application/stores/scripts_store.dart';
import '../../domain/entities/script.dart';
import '../../domain/entities/match_rules.dart';
import '../widgets/script_editor_dialog.dart';
import '../widgets/scripts_filters_dialog.dart';
import '../widgets/examples_library_dialog.dart';

/// Full Scripts Management Page with list and CRUD
class ScriptsPageFull extends StatefulWidget {
  final ScriptsStore store;

  const ScriptsPageFull({super.key, required this.store});

  @override
  State<ScriptsPageFull> createState() => _ScriptsPageFullState();
}

class _ScriptsPageFullState extends State<ScriptsPageFull> {
  @override
  void initState() {
    super.initState();
    // Load scripts on init
    widget.store.loadScripts();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Scripts'),
        actions: [
          // Import
          Observer(
            builder: (_) {
              final isImporting = widget.store.isImporting;
              return IconButton(
                icon: isImporting
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.upload_file),
                tooltip: isImporting
                    ? 'Importing...'
                    : 'Import Script from ZIP',
                onPressed: isImporting ? null : _importScript,
              );
            },
          ),
          // Examples Library
          IconButton(
            icon: const Icon(Icons.library_books),
            tooltip: 'Example Scripts Library',
            onPressed: _showExamples,
          ),
          // Search
          IconButton(icon: const Icon(Icons.search), onPressed: _showSearch),
          // Filter
          Observer(
            builder: (_) {
              final hasActiveFilters =
                  widget.store.runtimeFilter != null ||
                  widget.store.triggerTypeFilter != null ||
                  widget.store.enabledFilter != null;

              return Badge(
                isLabelVisible: hasActiveFilters,
                child: IconButton(
                  icon: const Icon(Icons.filter_list),
                  onPressed: _showFilters,
                ),
              );
            },
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: Column(
        children: [
          // Warning: section is not yet complete, screens are for demonstration
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(
                      Icons.warning_amber,
                      color: Theme.of(context).colorScheme.tertiary,
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: const [
                          Text(
                            'Scripts is a work-in-progress — these screens are primarily for demonstration. Functionality and behavior may change.',
                          ),
                          SizedBox(height: 6),
                          Text(
                            'Coming soon: write scripts and install plugins in multiple languages (Rust, Go, TypeScript, Python, Dart).',
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          Expanded(
            child: Observer(
              builder: (_) {
                if (widget.store.isLoading && widget.store.scripts.isEmpty) {
                  return const Center(child: CircularProgressIndicator());
                }

                if (widget.store.errorMessage != null) {
                  return _buildError(widget.store.errorMessage!);
                }

                if (widget.store.filteredScripts.isEmpty) {
                  return _buildEmpty();
                }

                return _buildList();
              },
            ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _createScript,
        icon: const Icon(Icons.add),
        label: const Text('Create Script'),
      ),
    );
  }

  Widget _buildList() {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: widget.store.filteredScripts.length,
      itemBuilder: (context, index) {
        final script = widget.store.filteredScripts[index];
        return _buildScriptCard(script);
      },
    );
  }

  Widget _buildScriptCard(Script script) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: () => _editScript(script),
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  // Name
                  Expanded(
                    child: Text(
                      script.name,
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  // Enabled toggle
                  Observer(
                    builder: (_) => Switch(
                      value: script.enabled,
                      onChanged: (value) {
                        widget.store.toggleScript(script.id, value);
                      },
                    ),
                  ),
                ],
              ),
              if (script.description != null &&
                  script.description!.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(
                  script.description!,
                  style: TextStyle(color: Colors.grey[600]),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
              const SizedBox(height: 12),
              // Badges
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  _badge(script.runtimeDisplay, Icons.memory, Colors.blue),
                  _badge(script.languageDisplay, Icons.code, Colors.purple),
                  _badge(
                    script.triggerType.name.toUpperCase(),
                    Icons.flash_on,
                    Colors.orange,
                  ),
                  _badge(
                    'Priority: ${script.priority}',
                    Icons.priority_high,
                    Colors.green,
                  ),
                  if (script.matchRules != null && !script.matchRules!.isEmpty)
                    _badge('Filtered', Icons.filter_alt, Colors.teal),
                ],
              ),
              const SizedBox(height: 12),
              // Actions
              Row(
                children: [
                  TextButton.icon(
                    onPressed: () => _editScript(script),
                    icon: const Icon(Icons.edit, size: 18),
                    label: const Text('Edit'),
                  ),
                  Observer(
                    builder: (_) {
                      final isExporting =
                          widget.store.exportingScriptId == script.id;
                      return TextButton.icon(
                        onPressed: isExporting
                            ? null
                            : () => _exportScript(script),
                        icon: isExporting
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.download, size: 18),
                        label: Text(isExporting ? 'Exporting...' : 'Export'),
                      );
                    },
                  ),
                  TextButton.icon(
                    onPressed: () => _deleteScript(script),
                    icon: const Icon(Icons.delete, size: 18),
                    label: const Text('Delete'),
                  ),
                  const Spacer(),
                  if (script.createdAt != null)
                    Text(
                      _formatDate(script.createdAt!),
                      style: TextStyle(color: Colors.grey[500], fontSize: 12),
                    ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _badge(String label, IconData icon, Color color) {
    return Chip(
      avatar: Icon(icon, size: 16, color: color),
      label: Text(label),
      visualDensity: VisualDensity.compact,
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
    );
  }

  Widget _buildEmpty() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.code_off, size: 100, color: Colors.grey),
          const SizedBox(height: 24),
          Text(
            'No Scripts Yet',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 8),
          const Text(
            'Create your first script to modify HTTP traffic',
            style: TextStyle(color: Colors.grey),
          ),
          const SizedBox(height: 32),
          ElevatedButton.icon(
            onPressed: _createScript,
            icon: const Icon(Icons.add),
            label: const Text('Create First Script'),
          ),
        ],
      ),
    );
  }

  Widget _buildError(String error) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.error_outline, size: 64, color: Colors.red),
          const SizedBox(height: 16),
          Text('Error', style: Theme.of(context).textTheme.headlineSmall),
          const SizedBox(height: 8),
          Text(
            error,
            style: const TextStyle(color: Colors.red),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          ElevatedButton(
            onPressed: () {
              widget.store.clearError();
              widget.store.loadScripts();
            },
            child: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  void _createScript() {
    ScriptEditorDialog.show(context, widget.store);
  }

  void _editScript(Script script) {
    ScriptEditorDialog.show(context, widget.store, editingScript: script);
  }

  Future<void> _exportScript(Script script) async {
    // Set exporting flag
    widget.store.exportingScriptId = script.id;
    widget.store.isExporting = true;

    try {
      // Get export URL from store
      final exportUrl = widget.store.getExportZipUrl(script.id);

      // Download the file
      final dio = Dio();
      final response = await dio.get<List<int>>(
        exportUrl,
        options: Options(responseType: ResponseType.bytes),
      );

      if (response.data == null) {
        throw Exception('Failed to download file');
      }

      // Save file using file_picker (works for both web and desktop)
      final fileName =
          '${script.name.replaceAll(RegExp(r'[^\w\s-]'), '_')}.zip';

      final result = await FilePicker.platform.saveFile(
        dialogTitle: 'Save Script Export',
        fileName: fileName,
        bytes: Uint8List.fromList(response.data!),
      );

      if (result != null && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Script exported successfully'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } on DioException catch (e) {
      // Translate DioException to user-friendly error messages
      String errorMessage;
      if (e.type == DioExceptionType.connectionTimeout ||
          e.type == DioExceptionType.sendTimeout ||
          e.type == DioExceptionType.receiveTimeout) {
        errorMessage =
            'Connection timeout. Please check your network connection.';
      } else if (e.type == DioExceptionType.connectionError) {
        errorMessage = 'Cannot connect to server. Is the backend running?';
      } else if (e.response != null) {
        final statusCode = e.response!.statusCode;
        if (statusCode == 404) {
          errorMessage = 'Script not found. It may have been deleted.';
        } else if (statusCode == 500) {
          errorMessage = 'Server error while exporting script.';
        } else {
          errorMessage = 'Export failed with status code $statusCode';
        }
      } else {
        errorMessage = 'Network error: ${e.message ?? "Unknown error"}';
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(errorMessage), backgroundColor: Colors.red),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Export failed: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      // Clear exporting flag
      widget.store.exportingScriptId = null;
      widget.store.isExporting = false;
    }
  }

  Future<void> _deleteScript(Script script) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Script?'),
        content: Text('Are you sure you want to delete "${script.name}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirm == true) {
      try {
        await widget.store.deleteScript(script.id);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Script deleted successfully'),
              backgroundColor: Colors.green,
            ),
          );
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Error: ${e.toString()}'),
              backgroundColor: Colors.red,
            ),
          );
        }
      }
    }
  }

  void _showExamples() {
    ExamplesLibraryDialog.show(context, widget.store);
  }

  void _showSearch() {
    showSearch(context: context, delegate: ScriptSearchDelegate(widget.store));
  }

  void _showFilters() {
    ScriptsFiltersDialog.show(context, widget.store);
  }

  Future<void> _importScript() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['zip'],
        withData: true,
      );

      if (result != null && result.files.isNotEmpty) {
        final file = result.files.first;
        if (file.bytes == null) {
          throw Exception('Failed to read file data');
        }

        final bytes = file.bytes!;

        // Validate file size (max 10MB)
        const maxSizeBytes = 10 * 1024 * 1024; // 10MB
        if (bytes.length > maxSizeBytes) {
          throw Exception('File is too large. Maximum size is 10MB.');
        }

        // Validate ZIP magic bytes (0x50 0x4B = "PK")
        if (bytes.length < 4 || bytes[0] != 0x50 || bytes[1] != 0x4B) {
          throw Exception(
            'Invalid ZIP file. Please select a valid ZIP archive.',
          );
        }

        await widget.store.importFromZip(bytes);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Script imported successfully'),
              backgroundColor: Colors.green,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Import failed: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inDays > 7) {
      return '${date.day}/${date.month}/${date.year}';
    } else if (diff.inDays > 0) {
      return '${diff.inDays}d ago';
    } else if (diff.inHours > 0) {
      return '${diff.inHours}h ago';
    } else if (diff.inMinutes > 0) {
      return '${diff.inMinutes}m ago';
    } else {
      return 'Just now';
    }
  }
}

/// Search delegate for scripts
class ScriptSearchDelegate extends SearchDelegate<Script?> {
  final ScriptsStore store;

  ScriptSearchDelegate(this.store);

  @override
  List<Widget> buildActions(BuildContext context) {
    return [
      if (query.isNotEmpty)
        IconButton(
          icon: const Icon(Icons.clear),
          onPressed: () {
            query = '';
            store.setSearchQuery('');
          },
        ),
    ];
  }

  @override
  Widget buildLeading(BuildContext context) {
    return IconButton(
      icon: const Icon(Icons.arrow_back),
      onPressed: () {
        store.setSearchQuery('');
        close(context, null);
      },
    );
  }

  @override
  Widget buildResults(BuildContext context) {
    store.setSearchQuery(query);
    return _buildSearchResults(context);
  }

  @override
  Widget buildSuggestions(BuildContext context) {
    store.setSearchQuery(query);
    return _buildSearchResults(context);
  }

  Widget _buildSearchResults(BuildContext context) {
    return Observer(
      builder: (_) {
        final results = store.filteredScripts;

        if (query.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.search, size: 64, color: Colors.grey[400]),
                const SizedBox(height: 16),
                Text(
                  'Search scripts',
                  style: Theme.of(
                    context,
                  ).textTheme.titleLarge?.copyWith(color: Colors.grey[600]),
                ),
                const SizedBox(height: 8),
                Text(
                  'Enter script name or description',
                  style: TextStyle(color: Colors.grey[500]),
                ),
              ],
            ),
          );
        }

        if (results.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.search_off, size: 64, color: Colors.grey[400]),
                const SizedBox(height: 16),
                Text(
                  'No scripts found',
                  style: Theme.of(
                    context,
                  ).textTheme.titleLarge?.copyWith(color: Colors.grey[600]),
                ),
                const SizedBox(height: 8),
                Text(
                  'Try different keywords',
                  style: TextStyle(color: Colors.grey[500]),
                ),
              ],
            ),
          );
        }

        return ListView.builder(
          padding: const EdgeInsets.all(16),
          itemCount: results.length,
          itemBuilder: (context, index) {
            final script = results[index];
            return Card(
              margin: const EdgeInsets.only(bottom: 8),
              child: ListTile(
                title: Text(script.name),
                subtitle: script.description != null
                    ? Text(
                        script.description!,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      )
                    : null,
                trailing: Chip(
                  label: Text(script.runtimeDisplay),
                  visualDensity: VisualDensity.compact,
                ),
                onTap: () {
                  store.setSearchQuery('');
                  close(context, script);
                },
              ),
            );
          },
        );
      },
    );
  }
}
