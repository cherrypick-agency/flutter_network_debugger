import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'dart:async';
import 'package:dio/dio.dart';
import '../../../../core/di/di.dart' show getIt;
import '../../application/services/updates_service.dart';
import '../../domain/entities/update_info.dart';
import '../../application/stores/updates_store.dart';
import '../widgets/download_progress_dialog.dart';
import '../../infrastructure/platform/platform_installer.dart';

class UpdatesPage extends StatefulWidget {
  const UpdatesPage({super.key});

  @override
  State<UpdatesPage> createState() => _UpdatesPageState();
}

class _UpdatesPageState extends State<UpdatesPage> {
  late final UpdatesStore _store;
  String _currentVersion = '';

  @override
  void initState() {
    super.initState();
    _store = getIt<UpdatesStore>();
    _loadCurrentVersion();
    _store.loadReleases();
  }

  Future<void> _loadCurrentVersion() async {
    try {
      final packageInfo = await PackageInfo.fromPlatform();
      setState(() {
        _currentVersion = packageInfo.version;
      });
    } catch (e) {
      // Log error for debugging
      debugPrint('Failed to get PackageInfo: $e');
      setState(() {
        // Fallback to version from pubspec.yaml instead of hardcoded 1.0.0
        _currentVersion = '0.1.4';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text('Updates'), centerTitle: false),
      body: Column(
        children: [
          // Header section с текущей версией и кнопкой проверки обновлений
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: colorScheme.surfaceContainerHighest.withValues(alpha: 0.3),
              border: Border(
                bottom: BorderSide(
                  color: colorScheme.outline.withValues(alpha: 0.2),
                ),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Текущая версия
                Row(
                  children: [
                    Icon(
                      Icons.info_outline,
                      size: 20,
                      color: colorScheme.primary,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      'Current Version: ',
                      style: theme.textTheme.bodyLarge?.copyWith(
                        color: colorScheme.onSurface.withValues(alpha: 0.7),
                      ),
                    ),
                    Text(
                      _currentVersion,
                      style: theme.textTheme.bodyLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: colorScheme.primary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),

                // Кнопка проверки обновлений
                Observer(
                  builder: (_) => ElevatedButton.icon(
                    onPressed: _store.isCheckingForUpdates
                        ? null
                        : () async {
                            await _store.checkForUpdates();
                            if (_store.availableUpdate != null && mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                SnackBar(
                                  content: Text(
                                    'Update available: ${_store.availableUpdate!.version}',
                                  ),
                                  backgroundColor: colorScheme.primary,
                                ),
                              );
                            } else if (mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text(
                                    'You are on the latest version',
                                  ),
                                ),
                              );
                            }
                          },
                    icon: _store.isCheckingForUpdates
                        ? SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: colorScheme.onPrimary,
                            ),
                          )
                        : const Icon(Icons.refresh),
                    label: const Text('Check for Updates'),
                  ),
                ),

                // Показ доступного обновления
                Observer(
                  builder: (_) {
                    if (_store.availableUpdate == null) {
                      return const SizedBox.shrink();
                    }

                    return Container(
                      margin: const EdgeInsets.only(top: 16),
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: colorScheme.primaryContainer,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                          color: colorScheme.primary.withValues(alpha: 0.3),
                        ),
                      ),
                      child: Row(
                        children: [
                          Icon(
                            Icons.system_update_alt,
                            color: colorScheme.onPrimaryContainer,
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Text(
                              'Update available: ${_store.availableUpdate!.version}',
                              style: theme.textTheme.bodyMedium?.copyWith(
                                color: colorScheme.onPrimaryContainer,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
              ],
            ),
          ),

          // Фильтры релизов
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
            child: Row(
              children: [
                Text(
                  'Filter:',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: colorScheme.onSurface.withValues(alpha: 0.7),
                  ),
                ),
                const SizedBox(width: 12),
                Observer(
                  builder: (_) => SegmentedButton<ReleaseFilter>(
                    segments: const [
                      ButtonSegment(
                        value: ReleaseFilter.all,
                        label: Text('All'),
                        icon: Icon(Icons.list, size: 16),
                      ),
                      ButtonSegment(
                        value: ReleaseFilter.stable,
                        label: Text('Stable'),
                        icon: Icon(Icons.verified, size: 16),
                      ),
                      ButtonSegment(
                        value: ReleaseFilter.prerelease,
                        label: Text('Pre-release'),
                        icon: Icon(Icons.science, size: 16),
                      ),
                    ],
                    selected: {_store.filter},
                    onSelectionChanged: (Set<ReleaseFilter> selected) {
                      _store.setFilter(selected.first);
                    },
                  ),
                ),
              ],
            ),
          ),

          // Список релизов
          Expanded(
            child: Observer(
              builder: (_) {
                // Состояние загрузки (первая загрузка)
                if (_store.isLoading && _store.releases.isEmpty) {
                  return const Center(child: CircularProgressIndicator());
                }

                // Ошибка
                if (_store.errorMessage != null && _store.releases.isEmpty) {
                  return Center(
                    child: Padding(
                      padding: const EdgeInsets.all(24),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.error_outline,
                            size: 48,
                            color: colorScheme.error,
                          ),
                          const SizedBox(height: 16),
                          Text(
                            _store.errorMessage!,
                            textAlign: TextAlign.center,
                            style: theme.textTheme.bodyLarge,
                          ),
                          const SizedBox(height: 16),
                          ElevatedButton.icon(
                            onPressed: () => _store.loadReleases(),
                            icon: const Icon(Icons.refresh),
                            label: const Text('Retry'),
                          ),
                        ],
                      ),
                    ),
                  );
                }

                // Пустой список
                if (_store.filteredReleases.isEmpty) {
                  return Center(
                    child: Padding(
                      padding: const EdgeInsets.all(24),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.inbox_outlined,
                            size: 48,
                            color: colorScheme.onSurface.withValues(alpha: 0.4),
                          ),
                          const SizedBox(height: 16),
                          Text(
                            'No releases found',
                            style: theme.textTheme.bodyLarge,
                          ),
                        ],
                      ),
                    ),
                  );
                }

                // Список релизов
                return ListView.builder(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 8,
                  ),
                  itemCount:
                      _store.filteredReleases.length +
                      (_store.hasMoreReleases ? 1 : 0),
                  itemBuilder: (context, index) {
                    // Кнопка Load More
                    if (index == _store.filteredReleases.length) {
                      return Padding(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        child: Center(
                          child: _store.isLoading
                              ? const CircularProgressIndicator()
                              : OutlinedButton.icon(
                                  onPressed: () =>
                                      _store.loadReleases(loadMore: true),
                                  icon: const Icon(Icons.expand_more),
                                  label: const Text('Load More'),
                                ),
                        ),
                      );
                    }

                    final release = _store.filteredReleases[index];
                    final isLatest = index == 0 && !release.isPrerelease;

                    return _ReleaseCard(
                      release: release,
                      isLatest: isLatest,
                      currentVersion: _currentVersion,
                    );
                  },
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _ReleaseCard extends StatefulWidget {
  final UpdateInfo release;
  final bool isLatest;
  final String currentVersion;

  const _ReleaseCard({
    required this.release,
    required this.isLatest,
    required this.currentVersion,
  });

  @override
  State<_ReleaseCard> createState() => _ReleaseCardState();
}

class _ReleaseCardState extends State<_ReleaseCard> {
  bool _isExpanded = false;

  /// Downloads and installs the release
  Future<void> _downloadRelease(
    BuildContext context,
    UpdateInfo release,
  ) async {
    final progressController = StreamController<DownloadProgress>();
    final cancelToken = CancelToken();

    try {
      final updatesService = getIt<UpdatesService>();

      // Запускаем загрузку в фоне
      final downloadFuture = updatesService.downloadUpdate(
        release,
        progressController: progressController,
        cancelToken: cancelToken,
      );

      // Показываем диалог прогресса
      // ignore: use_build_context_synchronously
      final downloadDialogFuture = showDownloadProgressDialog(
        context,
        progressStream: progressController.stream,
        onCancel: () => cancelToken.cancel(),
      );

      // Ждем оба future
      final results = await Future.wait([downloadFuture, downloadDialogFuture]);
      final filePath = results[0] as String?;
      final dialogResult = results[1] as DownloadDialogResult;

      if (!mounted) return;

      // Обрабатываем результат
      if (dialogResult == DownloadDialogResult.cancelled) {
        return; // Пользователь отменил
      }

      if (dialogResult == DownloadDialogResult.error || filePath == null) {
        // ignore: use_build_context_synchronously
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Download failed. Please try again.'),
            backgroundColor: Colors.red,
          ),
        );
        return;
      }

      // Успешная загрузка - открываем установщик
      // ignore: use_build_context_synchronously
      _openInstaller(context, filePath);
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
      );
    } finally {
      // ВАЖНО: Закрываем stream в finally блоке для предотвращения memory leak
      await progressController.close();
    }
  }

  /// Opens the downloaded installer
  Future<void> _openInstaller(BuildContext context, String filePath) async {
    try {
      final result = await openInstaller(filePath);

      if (!mounted) return;

      if (result.success) {
        // Показываем инструкции если есть
        if (result.instructions != null) {
          // ignore: use_build_context_synchronously
          await showDialog(
            context: context,
            builder: (ctx) => AlertDialog(
              title: const Row(
                children: [
                  Icon(Icons.info_outline, color: Colors.blue),
                  SizedBox(width: 12),
                  Text('Installation Instructions'),
                ],
              ),
              content: SingleChildScrollView(child: Text(result.instructions!)),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop(),
                  child: const Text('OK'),
                ),
              ],
            ),
          );
        } else {
          // ignore: use_build_context_synchronously
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Installer opened successfully'),
              backgroundColor: Colors.green,
            ),
          );
        }
      } else {
        // Ошибка открытия установщика
        // ignore: use_build_context_synchronously
        await showDialog(
          context: context,
          builder: (ctx) => AlertDialog(
            title: const Row(
              children: [
                Icon(Icons.error_outline, color: Colors.red),
                SizedBox(width: 12),
                Text('Installation Error'),
              ],
            ),
            content: Text(result.errorMessage ?? 'Failed to open installer'),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('OK'),
              ),
            ],
          ),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Error opening installer: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          InkWell(
            onTap: () => setState(() => _isExpanded = !_isExpanded),
            borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Версия и метаданные
                  Row(
                    children: [
                      // Версия
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: widget.release.isPrerelease
                              ? colorScheme.secondaryContainer
                              : colorScheme.primaryContainer,
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: Text(
                          widget.release.version,
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                            fontFamily: 'monospace',
                            color: widget.release.isPrerelease
                                ? colorScheme.onSecondaryContainer
                                : colorScheme.onPrimaryContainer,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),

                      // Badges
                      if (widget.isLatest)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: colorScheme.primary,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(
                                Icons.verified,
                                size: 14,
                                color: colorScheme.onPrimary,
                              ),
                              const SizedBox(width: 4),
                              Text(
                                'Latest',
                                style: theme.textTheme.labelSmall?.copyWith(
                                  color: colorScheme.onPrimary,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ],
                          ),
                        ),

                      if (widget.release.isPrerelease)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: colorScheme.secondary,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            'Pre-release',
                            style: theme.textTheme.labelSmall?.copyWith(
                              color: colorScheme.onSecondary,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),

                      const Spacer(),

                      // Expand icon
                      Icon(
                        _isExpanded ? Icons.expand_less : Icons.expand_more,
                        color: colorScheme.onSurface.withValues(alpha: 0.6),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),

                  // Дата и размер
                  Row(
                    children: [
                      Icon(
                        Icons.calendar_today,
                        size: 14,
                        color: colorScheme.onSurface.withValues(alpha: 0.6),
                      ),
                      const SizedBox(width: 6),
                      Text(
                        widget.release.formattedDate,
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: colorScheme.onSurface.withValues(alpha: 0.6),
                        ),
                      ),
                      const SizedBox(width: 16),
                      Icon(
                        Icons.storage,
                        size: 14,
                        color: colorScheme.onSurface.withValues(alpha: 0.6),
                      ),
                      const SizedBox(width: 6),
                      Text(
                        widget.release.formattedSize,
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: colorScheme.onSurface.withValues(alpha: 0.6),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),

          // Expanded changelog
          if (_isExpanded) ...[
            const Divider(height: 1),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Changelog header
                  Row(
                    children: [
                      Icon(
                        Icons.article_outlined,
                        size: 16,
                        color: colorScheme.primary,
                      ),
                      const SizedBox(width: 6),
                      Text(
                        'What\'s New',
                        style: theme.textTheme.titleSmall?.copyWith(
                          fontWeight: FontWeight.bold,
                          color: colorScheme.primary,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),

                  // Changelog with Markdown rendering
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: colorScheme.surfaceContainerHighest.withValues(
                        alpha: 0.3,
                      ),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: widget.release.changelog.isNotEmpty
                        ? MarkdownBody(
                            data: widget.release.changelog,
                            styleSheet: MarkdownStyleSheet(
                              p: theme.textTheme.bodyMedium,
                              h1: theme.textTheme.titleLarge,
                              h2: theme.textTheme.titleMedium,
                              h3: theme.textTheme.titleSmall,
                              code: theme.textTheme.bodySmall?.copyWith(
                                fontFamily: 'monospace',
                                backgroundColor:
                                    colorScheme.surfaceContainerHighest,
                              ),
                              codeblockDecoration: BoxDecoration(
                                color: colorScheme.surfaceContainerHighest,
                                borderRadius: BorderRadius.circular(4),
                              ),
                              listBullet: theme.textTheme.bodyMedium,
                              a: theme.textTheme.bodyMedium?.copyWith(
                                color: colorScheme.primary,
                                decoration: TextDecoration.underline,
                              ),
                            ),
                            onTapLink: (text, href, title) async {
                              if (href != null) {
                                final uri = Uri.parse(href);
                                if (await canLaunchUrl(uri)) {
                                  await launchUrl(uri);
                                }
                              }
                            },
                          )
                        : Text(
                            'No changelog available',
                            style: theme.textTheme.bodyMedium?.copyWith(
                              color: colorScheme.onSurface.withValues(
                                alpha: 0.5,
                              ),
                              fontStyle: FontStyle.italic,
                            ),
                          ),
                  ),
                  const SizedBox(height: 16),

                  // Action buttons
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      // View on GitHub
                      TextButton.icon(
                        onPressed: () async {
                          final uri = Uri.parse(widget.release.htmlUrl);
                          if (await canLaunchUrl(uri)) {
                            await launchUrl(uri);
                          }
                        },
                        icon: const Icon(Icons.open_in_new, size: 16),
                        label: const Text('View on GitHub'),
                      ),
                      const SizedBox(width: 8),

                      // Download button
                      ElevatedButton.icon(
                        onPressed: () =>
                            _downloadRelease(context, widget.release),
                        icon: const Icon(Icons.download, size: 16),
                        label: const Text('Download'),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }
}
