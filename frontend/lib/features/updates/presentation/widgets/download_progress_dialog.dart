import 'package:flutter/material.dart';
import 'dart:async';

/// Результат диалога загрузки
enum DownloadDialogResult { completed, cancelled, error }

/// Состояние загрузки
class DownloadProgress {
  final int downloaded; // байты
  final int total; // байты
  final double speed; // байт/сек
  final String? error;

  DownloadProgress({
    required this.downloaded,
    required this.total,
    required this.speed,
    this.error,
  });

  double get progress => total > 0 ? downloaded / total : 0.0;

  int get remainingBytes => total - downloaded;

  Duration get estimatedTimeRemaining {
    if (speed <= 0) return Duration.zero;
    final remainingSec = remainingBytes / speed;
    return Duration(seconds: remainingSec.toInt());
  }

  String get downloadedMB => (downloaded / 1024 / 1024).toStringAsFixed(1);
  String get totalMB => (total / 1024 / 1024).toStringAsFixed(1);
  String get speedMBps => (speed / 1024 / 1024).toStringAsFixed(2);

  String get progressPercent => '${(progress * 100).toStringAsFixed(0)}%';
}

/// Показывает диалог прогресса загрузки обновления
Future<DownloadDialogResult> showDownloadProgressDialog(
  BuildContext context, {
  required Stream<DownloadProgress> progressStream,
  required VoidCallback onCancel,
}) async {
  final result = await showDialog<DownloadDialogResult>(
    context: context,
    barrierDismissible: false,
    builder: (ctx) => _DownloadProgressDialog(
      progressStream: progressStream,
      onCancel: onCancel,
    ),
  );

  return result ?? DownloadDialogResult.cancelled;
}

class _DownloadProgressDialog extends StatefulWidget {
  final Stream<DownloadProgress> progressStream;
  final VoidCallback onCancel;

  const _DownloadProgressDialog({
    required this.progressStream,
    required this.onCancel,
  });

  @override
  State<_DownloadProgressDialog> createState() =>
      _DownloadProgressDialogState();
}

class _DownloadProgressDialogState extends State<_DownloadProgressDialog> {
  DownloadProgress? _currentProgress;
  StreamSubscription<DownloadProgress>? _subscription;
  bool _cancelled = false;
  bool _dialogClosed = false; // Флаг для предотвращения множественных pop()

  @override
  void initState() {
    super.initState();
    _subscription = widget.progressStream.listen(
      (progress) {
        if (!mounted || _dialogClosed) return;

        setState(() {
          _currentProgress = progress;
        });

        // Автоматически закрываем диалог при завершении
        if (progress.downloaded >= progress.total && progress.total > 0) {
          Future.delayed(const Duration(milliseconds: 500), () {
            if (mounted && !_cancelled && !_dialogClosed) {
              _closeDialog(DownloadDialogResult.completed);
            }
          });
        }

        // Показываем ошибку
        if (progress.error != null && mounted && !_dialogClosed) {
          Future.delayed(const Duration(milliseconds: 100), () {
            if (mounted && !_dialogClosed) {
              _closeDialog(DownloadDialogResult.error);
            }
          });
        }
      },
      onError: (error) {
        if (!mounted || _dialogClosed) return;
        _closeDialog(DownloadDialogResult.error);
      },
      onDone: () {
        if (!mounted || _dialogClosed) return;
        if (!_cancelled &&
            _currentProgress != null &&
            _currentProgress!.downloaded >= _currentProgress!.total) {
          _closeDialog(DownloadDialogResult.completed);
        }
      },
    );
  }

  /// Безопасное закрытие диалога (только один раз)
  void _closeDialog(DownloadDialogResult result) {
    if (_dialogClosed) return;
    _dialogClosed = true;
    if (mounted) {
      Navigator.of(context).pop(result);
    }
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }

  void _handleCancel() {
    setState(() {
      _cancelled = true;
    });
    widget.onCancel();
    _closeDialog(DownloadDialogResult.cancelled);
  }

  @override
  Widget build(BuildContext context) {
    final progress = _currentProgress;

    return PopScope(
      canPop: false,
      child: AlertDialog(
        title: const Row(
          children: [
            Icon(Icons.download, size: 24),
            SizedBox(width: 12),
            Text('Downloading Update'),
          ],
        ),
        content: SizedBox(
          width: 400,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (progress != null) ...[
                // Progress Bar
                LinearProgressIndicator(
                  value: progress.progress,
                  minHeight: 8,
                  backgroundColor: Colors.grey.withValues(alpha: 0.2),
                  valueColor: AlwaysStoppedAnimation<Color>(
                    Theme.of(context).colorScheme.primary,
                  ),
                ),
                const SizedBox(height: 16),

                // Progress Info
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      progress.progressPercent,
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    Text(
                      '${progress.downloadedMB} / ${progress.totalMB} MB',
                      style: Theme.of(
                        context,
                      ).textTheme.bodyMedium?.copyWith(color: Colors.grey),
                    ),
                  ],
                ),
                const SizedBox(height: 12),

                // Speed and Time Remaining
                Row(
                  children: [
                    Icon(Icons.speed, size: 16, color: Colors.grey[600]),
                    const SizedBox(width: 4),
                    Text(
                      '${progress.speedMBps} MB/s',
                      style: Theme.of(
                        context,
                      ).textTheme.bodySmall?.copyWith(color: Colors.grey[600]),
                    ),
                    const SizedBox(width: 16),
                    Icon(Icons.timer, size: 16, color: Colors.grey[600]),
                    const SizedBox(width: 4),
                    Text(
                      _formatDuration(progress.estimatedTimeRemaining),
                      style: Theme.of(
                        context,
                      ).textTheme.bodySmall?.copyWith(color: Colors.grey[600]),
                    ),
                  ],
                ),

                if (progress.error != null) ...[
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.red.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: Colors.red.withValues(alpha: 0.3),
                      ),
                    ),
                    child: Row(
                      children: [
                        const Icon(
                          Icons.error_outline,
                          color: Colors.red,
                          size: 20,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            progress.error!,
                            style: const TextStyle(
                              fontSize: 12,
                              color: Colors.red,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ] else ...[
                const Center(
                  child: Column(
                    children: [
                      CircularProgressIndicator(),
                      SizedBox(height: 16),
                      Text('Initializing download...'),
                    ],
                  ),
                ),
              ],
            ],
          ),
        ),
        actions: [
          TextButton.icon(
            onPressed: _handleCancel,
            icon: const Icon(Icons.close),
            label: const Text('Cancel'),
          ),
        ],
      ),
    );
  }

  String _formatDuration(Duration duration) {
    if (duration.inSeconds < 1) return '< 1s';
    if (duration.inSeconds < 60) return '${duration.inSeconds}s';
    if (duration.inMinutes < 60) {
      final sec = duration.inSeconds % 60;
      return '${duration.inMinutes}m ${sec}s';
    }
    return '${duration.inHours}h ${duration.inMinutes % 60}m';
  }
}
