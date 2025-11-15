import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:file_picker/file_picker.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';
import '../../platform/file_operations.dart' as html;
import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../theme/context_ext.dart';
import '../../../inspector/domain/repositories/inspector_repository.dart';
import '../../application/stores/export_import_store.dart';

/// Shows the Export/Import dialog
Future<void> showExportImportDialog(
  BuildContext context, {
  List<String>? visibleSessionIds,
  int? visibleSessionsCount,
  int? totalSessionsCount,
}) async {
  return await showDialog<void>(
    context: context,
    barrierDismissible: false,
    builder: (ctx) => _ExportImportDialog(
      visibleSessionIds: visibleSessionIds,
      visibleSessionsCount: visibleSessionsCount,
      totalSessionsCount: totalSessionsCount,
    ),
  );
}

class _ExportImportDialog extends StatefulWidget {
  final List<String>? visibleSessionIds;
  final int? visibleSessionsCount;
  final int? totalSessionsCount;

  const _ExportImportDialog({
    this.visibleSessionIds,
    this.visibleSessionsCount,
    this.totalSessionsCount,
  });

  @override
  State<_ExportImportDialog> createState() => _ExportImportDialogState();
}

class _ExportImportDialogState extends State<_ExportImportDialog> {
  late final ExportImportStore store;

  @override
  void initState() {
    super.initState();
    final repo = sl<InspectorRepository>();
    store = ExportImportStore(repo);

    // Set preview counts
    if (widget.visibleSessionsCount != null) {
      store.setVisibleSessionsCount(widget.visibleSessionsCount!);
    }
    if (widget.totalSessionsCount != null) {
      store.setTotalSessionsCount(widget.totalSessionsCount!);
    }
  }

  @override
  void dispose() {
    // Reset store state to clear file data and prevent memory leaks
    store.reset();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Provider<ExportImportStore>.value(
      value: store,
      child: Focus(
        autofocus: true,
        onKeyEvent: (node, event) {
          // Only handle key down events
          if (event is! KeyDownEvent) {
            return KeyEventResult.ignored;
          }

          // ESC - Close dialog
          if (event.logicalKey == LogicalKeyboardKey.escape) {
            Navigator.of(context).pop();
            return KeyEventResult.handled;
          }

          // Enter - Perform export if ready
          if (event.logicalKey == LogicalKeyboardKey.enter) {
            if (store.state == ExportImportState.exportSettings &&
                store.canExport) {
              final sessionIds = store.exportScope == ExportScope.visible
                  ? widget.visibleSessionIds
                  : null;
              store.performExport(sessionIds);
              return KeyEventResult.handled;
            }
          }

          return KeyEventResult.ignored;
        },
        child: Center(
          child: Material(
            color: Theme.of(context).colorScheme.surface,
            elevation: 12,
            borderRadius: BorderRadius.circular(16),
            child: SizedBox(
              width: 600,
              child: Observer(
                builder: (_) {
                  return Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      _buildHeader(context),
                      const Divider(height: 1),
                      _buildContent(context),
                    ],
                  );
                },
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(BuildContext context) {
    return Observer(
      builder: (_) {
        String title = 'Export / Import';
        if (store.state == ExportImportState.exportSettings) {
          title = 'Export Settings';
        } else if (store.state == ExportImportState.importSettings) {
          title = 'Import Settings';
        } else if (store.state == ExportImportState.exporting) {
          title = 'Exporting...';
        } else if (store.state == ExportImportState.importing) {
          title = 'Importing...';
        } else if (store.state == ExportImportState.success) {
          title = 'Success';
        } else if (store.state == ExportImportState.error) {
          title = 'Error';
        }

        return Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          child: Row(
            children: [
              Expanded(child: Text(title, style: context.appText.title)),
              if (store.state == ExportImportState.exportSettings ||
                  store.state == ExportImportState.importSettings)
                IconButton(
                  icon: const Icon(Icons.arrow_back),
                  tooltip: 'Back',
                  onPressed: () => store.backToModeSelection(),
                ),
              IconButton(
                icon: const Icon(Icons.close),
                tooltip: 'Close',
                onPressed: () => Navigator.of(context).pop(),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildContent(BuildContext context) {
    return Observer(
      builder: (_) {
        switch (store.state) {
          case ExportImportState.selectMode:
            return ModeSelectionScreen(
              visibleSessionIds: widget.visibleSessionIds,
            );
          case ExportImportState.exportSettings:
            return ExportSettingsScreen(
              visibleSessionIds: widget.visibleSessionIds,
            );
          case ExportImportState.importSettings:
            return const ImportSettingsScreen();
          case ExportImportState.importing:
          case ExportImportState.exporting:
            return const ProgressScreen();
          case ExportImportState.success:
            return const SuccessScreen();
          case ExportImportState.error:
            return const ErrorScreen();
        }
      },
    );
  }
}

/// Mode selection: Export or Import
class ModeSelectionScreen extends StatefulWidget {
  final List<String>? visibleSessionIds;

  const ModeSelectionScreen({super.key, this.visibleSessionIds});

  @override
  State<ModeSelectionScreen> createState() => _ModeSelectionScreenState();
}

class _ModeSelectionScreenState extends State<ModeSelectionScreen> {
  bool _isDragging = false;
  bool _isLoadingFile = false;
  final _dropZoneKey = GlobalKey();

  // HTML event listeners subscriptions (for cleanup)
  StreamSubscription? _dragOverSub;
  StreamSubscription? _dragLeaveSub;
  StreamSubscription? _dropSub;
  StreamSubscription? _fileInputSub;

  @override
  void initState() {
    super.initState();
    // Add HTML drag & drop listeners after first frame
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _setupDragAndDrop();
    });
  }

  @override
  void dispose() {
    // Cancel HTML event listeners to prevent memory leaks
    _dragOverSub?.cancel();
    _dragLeaveSub?.cancel();
    _dropSub?.cancel();
    _fileInputSub?.cancel();
    super.dispose();
  }

  void _setupDragAndDrop() {
    final element = _dropZoneKey.currentContext?.findRenderObject();
    if (element == null) return;

    // Note: HTML drag & drop events are handled at the document level
    // Store subscriptions for cleanup in dispose()
    _dragOverSub = html.document.onDragOver.listen((event) {
      event.preventDefault();
      if (!_isDragging && mounted) {
        setState(() => _isDragging = true);
      }
    });

    _dragLeaveSub = html.document.onDragLeave.listen((event) {
      if (_isDragging && mounted) {
        setState(() => _isDragging = false);
      }
    });

    _dropSub = html.document.onDrop.listen((event) {
      event.preventDefault();
      if (!mounted) return;

      setState(() {
        _isDragging = false;
        _isLoadingFile = true;
      });

      final files = event.dataTransfer.files;
      if (files == null || files.isEmpty) {
        if (mounted) {
          setState(() => _isLoadingFile = false);
        }
        return;
      }

      final file = files[0];
      _handleFileSelection(file);
    });
  }

  void _handleFileSelection(html.File file) {
    if (!mounted) return;

    setState(() => _isLoadingFile = true);

    // Capture store BEFORE async operation to avoid context across async gap
    final store = context.read<ExportImportStore>();
    final reader = html.FileReader();

    reader.onLoadEnd.listen((e) {
      if (!mounted) return;
      setState(() => _isLoadingFile = false);

      // Safe type check instead of unsafe cast
      final bytes = reader.result;
      if (bytes is! Uint8List) {
        if (!mounted) return;
        store.errorMessage = 'Failed to read file data';
        store.state = ExportImportState.error;
        return;
      }

      if (!mounted) return;
      store.selectMode(ExportImportMode.import);
      store.setImportFile(bytes, file.name, file.size);

      // Navigate to import settings screen
      if (!mounted) return;
      if (store.state != ExportImportState.error) {
        store.state = ExportImportState.importSettings;
      }
    });

    reader.onError.listen((e) {
      if (!mounted) return;
      setState(() => _isLoadingFile = false);
      store.errorMessage = 'Error reading file';
      store.state = ExportImportState.error;
    });

    reader.readAsArrayBuffer(file);
  }

  @override
  Widget build(BuildContext context) {
    final store = context.read<ExportImportStore>();

    return AnimatedContainer(
      key: _dropZoneKey,
      duration: const Duration(milliseconds: 200),
      decoration: BoxDecoration(
        color: _isDragging
            ? Colors.green.withValues(alpha: 0.1)
            : Colors.transparent,
        border: _isDragging
            ? Border.all(
                color: Colors.green,
                width: 2,
                strokeAlign: BorderSide.strokeAlignInside,
              )
            : null,
        borderRadius: BorderRadius.circular(12),
      ),
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (_isDragging)
            Container(
              padding: const EdgeInsets.all(16),
              margin: const EdgeInsets.only(bottom: 16),
              decoration: BoxDecoration(
                color: Colors.green.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.file_upload, color: Colors.green),
                  const SizedBox(width: 8),
                  Text(
                    'Drop HAR file here to import',
                    style: TextStyle(
                      color: Colors.green.shade700,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
          if (_isLoadingFile)
            Container(
              padding: const EdgeInsets.all(16),
              margin: const EdgeInsets.only(bottom: 16),
              decoration: BoxDecoration(
                color: Colors.blue.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                  const SizedBox(width: 12),
                  Text(
                    'Reading file...',
                    style: TextStyle(
                      color: Colors.blue.shade700,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
          Text(
            _isLoadingFile
                ? 'Please wait...'
                : _isDragging
                ? 'Release to import'
                : 'Choose an action',
            style: context.appText.subtitle.copyWith(
              color: context.appColors.textSecondary,
            ),
          ),
          const SizedBox(height: 24),
          Row(
            children: [
              Expanded(
                child: _ModeButton(
                  icon: Icons.upload_rounded,
                  label: 'Export',
                  description: 'Export sessions to HAR file',
                  color: Colors.blue,
                  onPressed: () => store.selectMode(ExportImportMode.export),
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: _ModeButton(
                  icon: Icons.download_rounded,
                  label: 'Import',
                  description: 'Import HAR file or drag & drop',
                  color: Colors.green,
                  onPressed: () => _pickFile(context, store),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  Future<void> _pickFile(BuildContext context, ExportImportStore store) async {
    if (kIsWeb) {
      // Web: use html file picker
      final input = html.FileUploadInputElement()
        ..accept = '.har,application/json';
      input.click();

      // Store subscription for cleanup
      _fileInputSub?.cancel();
      _fileInputSub = input.onChange.listen((e) {
        if (!mounted) return;

        final files = input.files;
        if (files == null || files.isEmpty) return;

        final file = files[0];
        _handleFileSelection(file);
      });
    } else {
      // Desktop/Mobile: use file_picker package
      try {
        final result = await FilePicker.platform.pickFiles(
          type: FileType.custom,
          allowedExtensions: ['har', 'json'],
          withData: true,
          allowMultiple: false,
        );

        if (result == null || result.files.isEmpty) {
          // User cancelled
          return;
        }

        final file = result.files.first;
        if (file.bytes == null) {
          if (!mounted) return;
          store.errorMessage = 'Failed to read file data';
          store.state = ExportImportState.error;
          sl<NotificationsService>().error(
            'Error Reading File',
            'Could not read file data',
          );
          return;
        }

        // Extract bytes before setting to avoid force unwrap
        final bytes = file.bytes!;

        if (!mounted) return;

        // Select import mode and set file data
        store.selectMode(ExportImportMode.import);
        store.setImportFile(bytes, file.name, file.size);

        // Navigate to import settings screen
        if (!mounted) return;
        if (store.state != ExportImportState.error) {
          store.state = ExportImportState.importSettings;
        }
      } catch (e) {
        if (!mounted) return;
        store.errorMessage = 'Error selecting file: ${e.toString()}';
        store.state = ExportImportState.error;
        sl<NotificationsService>().error('Error Selecting File', e.toString());
      }
    }
  }
}

class _ModeButton extends StatefulWidget {
  final IconData icon;
  final String label;
  final String description;
  final Color color;
  final VoidCallback onPressed;

  const _ModeButton({
    required this.icon,
    required this.label,
    required this.description,
    required this.color,
    required this.onPressed,
  });

  @override
  State<_ModeButton> createState() => _ModeButtonState();
}

class _ModeButtonState extends State<_ModeButton> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeInOut,
        transform: Matrix4.diagonal3Values(
          _isHovered ? 1.02 : 1.0,
          _isHovered ? 1.02 : 1.0,
          1.0,
        ),
        child: Material(
          color: widget.color.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(12),
          child: InkWell(
            onTap: widget.onPressed,
            borderRadius: BorderRadius.circular(12),
            child: Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: widget.color.withValues(alpha: _isHovered ? 0.5 : 0.3),
                  width: 2,
                ),
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(widget.icon, size: 48, color: widget.color),
                  const SizedBox(height: 12),
                  Text(
                    widget.label,
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    widget.description,
                    style: TextStyle(
                      fontSize: 12,
                      color: context.appColors.textSecondary,
                    ),
                    textAlign: TextAlign.center,
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// Export settings screen
class ExportSettingsScreen extends StatelessWidget {
  final List<String>? visibleSessionIds;

  const ExportSettingsScreen({super.key, this.visibleSessionIds});

  @override
  Widget build(BuildContext context) {
    final store = context.read<ExportImportStore>();

    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Observer(
        builder: (_) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            mainAxisSize: MainAxisSize.min,
            children: [
              // Scope selection
              Text(
                'Export Scope',
                style: context.appText.subtitle.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 12),
              _ScopeOption(
                title: 'Visible Sessions',
                description: store.visibleScopeDescription,
                isSelected: store.exportScope == ExportScope.visible,
                onTap: () => store.setExportScope(ExportScope.visible),
                enabled: (store.visibleSessionsCount ?? 0) > 0,
              ),
              const SizedBox(height: 8),
              _ScopeOption(
                title: 'All Sessions',
                description: store.allScopeDescription,
                isSelected: store.exportScope == ExportScope.all,
                onTap: () => store.setExportScope(ExportScope.all),
                enabled: (store.totalSessionsCount ?? 0) > 0,
              ),
              const SizedBox(height: 24),

              // Export options
              Text(
                'Export Options',
                style: context.appText.subtitle.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 12),
              CheckboxListTile(
                value: store.includeBodies,
                onChanged: (_) => store.toggleIncludeBodies(),
                title: const Text('Include Request/Response Bodies'),
                subtitle: const Text(
                  'Bodies can significantly increase file size',
                ),
                dense: true,
                contentPadding: EdgeInsets.zero,
              ),
              CheckboxListTile(
                value: store.includeSensitive,
                onChanged: (_) => store.toggleIncludeSensitive(),
                title: const Text('Include Sensitive Headers'),
                subtitle: const Text('Authorization, cookies, etc.'),
                dense: true,
                contentPadding: EdgeInsets.zero,
              ),
              CheckboxListTile(
                value: store.minify,
                onChanged: (_) => store.toggleMinify(),
                title: const Text('Minify JSON'),
                subtitle: const Text('Reduces file size, harder to read'),
                dense: true,
                contentPadding: EdgeInsets.zero,
              ),
              const SizedBox(height: 24),

              // Preview info
              if (store.estimatedFileSize != null)
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.blue.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(
                      color: Colors.blue.withValues(alpha: 0.3),
                    ),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.info_outline, color: Colors.blue),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              'Estimated file size',
                              style: TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              store.estimatedFileSize!,
                              style: const TextStyle(
                                fontSize: 16,
                                color: Colors.blue,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              const SizedBox(height: 24),

              // Export button
              SizedBox(
                height: 48,
                child: ElevatedButton.icon(
                  onPressed: store.canExport
                      ? () {
                          final sessionIds =
                              store.exportScope == ExportScope.visible
                              ? visibleSessionIds
                              : null;
                          store.performExport(sessionIds);
                        }
                      : null,
                  icon: const Icon(Icons.upload_rounded),
                  label: const Text('Export HAR'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blue,
                    foregroundColor: Colors.white,
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _ScopeOption extends StatelessWidget {
  final String title;
  final String description;
  final bool isSelected;
  final VoidCallback onTap;
  final bool enabled;

  const _ScopeOption({
    required this.title,
    required this.description,
    required this.isSelected,
    required this.onTap,
    this.enabled = true,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: isSelected
          ? Colors.blue.withValues(alpha: 0.15)
          : Colors.grey.withValues(alpha: 0.05),
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: enabled ? onTap : null,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: isSelected
                  ? Colors.blue
                  : Colors.grey.withValues(alpha: 0.2),
              width: 2,
            ),
          ),
          child: Row(
            children: [
              Icon(
                isSelected
                    ? Icons.radio_button_checked
                    : Icons.radio_button_unchecked,
                color: isSelected ? Colors.blue : Colors.grey,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: enabled ? null : context.appColors.textSecondary,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      description,
                      style: TextStyle(
                        fontSize: 12,
                        color: context.appColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Import settings screen
class ImportSettingsScreen extends StatelessWidget {
  const ImportSettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final store = context.read<ExportImportStore>();

    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Observer(
        builder: (_) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            mainAxisSize: MainAxisSize.min,
            children: [
              // File preview
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.blue.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.blue.withValues(alpha: 0.3)),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.file_present,
                      color: Colors.blue,
                      size: 32,
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            store.importFileName ?? 'Unknown file',
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                            ),
                            overflow: TextOverflow.ellipsis,
                          ),
                          const SizedBox(height: 4),
                          Row(
                            children: [
                              if (store.importFileSize != null)
                                Text(
                                  'Size: ${(store.importFileSize! / 1024).toStringAsFixed(1)} KB',
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: context.appColors.textSecondary,
                                  ),
                                ),
                              if (store.importFileSize != null &&
                                  store.harEntriesCount != null)
                                Text(
                                  ' • ',
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: context.appColors.textSecondary,
                                  ),
                                ),
                              if (store.harEntriesCount != null)
                                Text(
                                  'Entries: ${store.harEntriesCount}',
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: context.appColors.textSecondary,
                                  ),
                                ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // Import mode selection
              Text(
                'Import Mode',
                style: context.appText.subtitle.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 12),

              _ImportModeOption(
                title: 'Add to existing',
                description:
                    'Keep all current sessions and add imported ones (safe)',
                mode: ImportMode.merge,
                isSelected: store.importMode == ImportMode.merge,
                onTap: () => store.setImportMode(ImportMode.merge),
              ),
              const SizedBox(height: 8),

              _ImportModeOption(
                title: 'Replace all sessions',
                description:
                    'Delete ALL sessions before import (⚠️ destructive)',
                mode: ImportMode.replaceAll,
                isSelected: store.importMode == ImportMode.replaceAll,
                onTap: () => store.setImportMode(ImportMode.replaceAll),
                isDestructive: true,
              ),
              const SizedBox(height: 8),

              _ImportModeOption(
                title: 'Replace imported only',
                description:
                    'Delete previously imported sessions, keep proxy sessions',
                mode: ImportMode.replaceImported,
                isSelected: store.importMode == ImportMode.replaceImported,
                onTap: () => store.setImportMode(ImportMode.replaceImported),
              ),

              const SizedBox(height: 24),

              // Import button
              SizedBox(
                height: 48,
                child: ElevatedButton.icon(
                  onPressed: () => store.performImport(),
                  icon: const Icon(Icons.download_rounded),
                  label: const Text('Import HAR'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.green,
                    foregroundColor: Colors.white,
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _ImportModeOption extends StatelessWidget {
  final String title;
  final String description;
  final ImportMode mode;
  final bool isSelected;
  final VoidCallback onTap;
  final bool isDestructive;

  const _ImportModeOption({
    required this.title,
    required this.description,
    required this.mode,
    required this.isSelected,
    required this.onTap,
    this.isDestructive = false,
  });

  @override
  Widget build(BuildContext context) {
    final borderColor = isSelected
        ? (isDestructive ? Colors.red : Colors.green)
        : Colors.grey.withValues(alpha: 0.2);

    final backgroundColor = isSelected
        ? (isDestructive
              ? Colors.red.withValues(alpha: 0.15)
              : Colors.green.withValues(alpha: 0.15))
        : Colors.grey.withValues(alpha: 0.05);

    final iconColor = isSelected
        ? (isDestructive ? Colors.red : Colors.green)
        : Colors.grey;

    return Material(
      color: backgroundColor,
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: borderColor, width: 2),
          ),
          child: Row(
            children: [
              Icon(
                isSelected
                    ? Icons.radio_button_checked
                    : Icons.radio_button_unchecked,
                color: iconColor,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      description,
                      style: TextStyle(
                        fontSize: 12,
                        color: context.appColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Progress screen for import/export operations
class ProgressScreen extends StatelessWidget {
  const ProgressScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final store = context.read<ExportImportStore>();

    return Padding(
      padding: const EdgeInsets.all(48),
      child: Observer(
        builder: (_) {
          return Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                width: 80,
                height: 80,
                child: CircularProgressIndicator(
                  value: store.progress > 0 ? store.progress : null,
                  strokeWidth: 6,
                ),
              ),
              const SizedBox(height: 24),
              Text(
                store.progressMessage ?? 'Processing...',
                style: context.appText.subtitle,
                textAlign: TextAlign.center,
              ),
              if (store.progress > 0) ...[
                const SizedBox(height: 12),
                Text(
                  '${(store.progress * 100).toStringAsFixed(0)}%',
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.blue,
                  ),
                ),
              ],
            ],
          );
        },
      ),
    );
  }
}

/// Success screen
class SuccessScreen extends StatelessWidget {
  const SuccessScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final store = context.read<ExportImportStore>();

    return Padding(
      padding: const EdgeInsets.all(48),
      child: Observer(
        builder: (_) {
          return Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  color: Colors.green.withValues(alpha: 0.15),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.check_circle,
                  size: 48,
                  color: Colors.green,
                ),
              ),
              const SizedBox(height: 24),
              Text(
                store.successMessage ?? 'Operation completed successfully',
                style: context.appText.subtitle,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 32),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Close'),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

/// Error screen
class ErrorScreen extends StatelessWidget {
  const ErrorScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final store = context.read<ExportImportStore>();

    return Padding(
      padding: const EdgeInsets.all(48),
      child: Observer(
        builder: (_) {
          return Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  color: Colors.red.withValues(alpha: 0.15),
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.error, size: 48, color: Colors.red),
              ),
              const SizedBox(height: 24),
              Text(
                store.errorMessage ?? 'Operation failed',
                style: context.appText.subtitle.copyWith(color: Colors.red),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 32),
              Row(
                children: [
                  Expanded(
                    child: SizedBox(
                      height: 48,
                      child: OutlinedButton(
                        onPressed: () => store.backToModeSelection(),
                        child: const Text('Try Again'),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: SizedBox(
                      height: 48,
                      child: ElevatedButton(
                        onPressed: () => Navigator.of(context).pop(),
                        child: const Text('Close'),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          );
        },
      ),
    );
  }
}
