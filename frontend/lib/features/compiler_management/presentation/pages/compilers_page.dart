import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import '../stores/compiler_list_store.dart';
import '../stores/installation_progress_store.dart';
import '../widgets/compiler_card.dart';

class CompilersPage extends StatefulWidget {
  final CompilerListStore store;
  final InstallationProgressStore progressStore;

  const CompilersPage({
    super.key,
    required this.store,
    required this.progressStore,
  });

  @override
  State<CompilersPage> createState() => _CompilersPageState();
}

class _CompilersPageState extends State<CompilersPage> {
  @override
  void initState() {
    super.initState();
    widget.store.loadCompilers();
  }

  @override
  void dispose() {
    widget.progressStore.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Compiler Management'),
        actions: [
          Observer(
            builder: (_) {
              final size = widget.store.totalCacheSize;
              if (size > 0) {
                return Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Chip(
                    avatar: Icon(
                      Icons.storage,
                      size: 16,
                      color: colorScheme.onSurfaceVariant,
                    ),
                    label: Text(_formatBytes(size)),
                    backgroundColor: colorScheme.surfaceVariant,
                  ),
                );
              }
              return const SizedBox.shrink();
            },
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => widget.store.loadCompilers(),
            tooltip: 'Refresh',
          ),
        ],
      ),
      body: Observer(
        builder: (_) {
          if (widget.store.isLoading && widget.store.compilers.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (widget.store.error != null) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.error_outline, size: 48, color: colorScheme.error),
                  const SizedBox(height: 16),
                  Text(
                    'Error loading compilers',
                    style: theme.textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    widget.store.error!,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 16),
                  FilledButton.icon(
                    onPressed: () {
                      widget.store.clearError();
                      widget.store.loadCompilers();
                    },
                    icon: const Icon(Icons.refresh),
                    label: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (widget.store.compilers.isEmpty) {
            return const Center(child: Text('No compilers available'));
          }

          return RefreshIndicator(
            onRefresh: widget.store.loadCompilers,
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (widget.store.installedCompilers.isNotEmpty) ...[
                  Text(
                    'Installed',
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  ...widget.store.installedCompilers.map(
                    (compiler) => Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Observer(
                        builder: (_) => CompilerCard(
                          compiler: compiler,
                          progress: widget.progressStore.getProgress(
                            compiler.language,
                          ),
                          onUninstall: () {
                            _showUninstallDialog(compiler.language);
                          },
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 24),
                ],
                if (widget.store.availableCompilers.isNotEmpty) ...[
                  Text(
                    'Available for Download',
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  ...widget.store.availableCompilers.map(
                    (compiler) => Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Observer(
                        builder: (_) => CompilerCard(
                          compiler: compiler,
                          progress: widget.progressStore.getProgress(
                            compiler.language,
                          ),
                          onInstall: () {
                            _install(compiler.language);
                          },
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          );
        },
      ),
    );
  }

  void _install(String language) {
    widget.progressStore.startWatching(language);
    widget.store.install(language);
  }

  void _showUninstallDialog(String language) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Uninstall Compiler'),
        content: Text('Are you sure you want to uninstall $language?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () {
              Navigator.of(context).pop();
              widget.store.uninstall(language);
            },
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
            ),
            child: const Text('Uninstall'),
          ),
        ],
      ),
    );
  }

  String _formatBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(1)} GB';
  }
}
