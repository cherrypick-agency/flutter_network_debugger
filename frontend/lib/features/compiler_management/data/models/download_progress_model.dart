import '../../domain/entities/download_progress.dart';

class DownloadProgressModel {
  final String stage;
  final double percentage;
  final int bytesDownloaded;
  final int totalBytes;
  final String message;

  DownloadProgressModel({
    required this.stage,
    required this.percentage,
    required this.bytesDownloaded,
    required this.totalBytes,
    required this.message,
  });

  factory DownloadProgressModel.fromJson(Map<String, dynamic> json) {
    return DownloadProgressModel(
      stage: json['stage'] as String? ?? '',
      percentage: (json['percentage'] as num?)?.toDouble() ?? 0.0,
      bytesDownloaded: json['bytes_downloaded'] as int? ?? 0,
      totalBytes: json['total_bytes'] as int? ?? 0,
      message: json['message'] as String? ?? '',
    );
  }

  DownloadProgress toEntity() {
    return DownloadProgress(
      stage: stage,
      percentage: percentage,
      bytesDownloaded: bytesDownloaded,
      totalBytes: totalBytes,
      message: message,
    );
  }
}
