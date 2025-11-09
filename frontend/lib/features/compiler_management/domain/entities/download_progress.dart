import 'package:freezed_annotation/freezed_annotation.dart';

part 'download_progress.freezed.dart';
part 'download_progress.g.dart';

@freezed
sealed class DownloadProgress with _$DownloadProgress {
  const DownloadProgress._();

  const factory DownloadProgress({
    required String stage,
    required double percentage,
    required int bytesDownloaded,
    required int totalBytes,
    required String message,
  }) = _DownloadProgress;

  factory DownloadProgress.fromJson(Map<String, dynamic> json) =>
      _$DownloadProgressFromJson(json);

  bool get isDownloading => stage == 'downloading';
  bool get isExtracting => stage == 'extracting';
  bool get isVerifying => stage == 'verifying';
  bool get isComplete => stage == 'complete';

  String get progressText {
    if (bytesDownloaded > 0 && totalBytes > 0) {
      final downloaded = _formatBytes(bytesDownloaded);
      final total = _formatBytes(totalBytes);
      return '$downloaded / $total';
    }
    return message;
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
