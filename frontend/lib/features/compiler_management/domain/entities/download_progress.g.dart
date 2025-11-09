// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'download_progress.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_DownloadProgress _$DownloadProgressFromJson(Map<String, dynamic> json) =>
    _DownloadProgress(
      stage: json['stage'] as String,
      percentage: (json['percentage'] as num).toDouble(),
      bytesDownloaded: (json['bytesDownloaded'] as num).toInt(),
      totalBytes: (json['totalBytes'] as num).toInt(),
      message: json['message'] as String,
    );

Map<String, dynamic> _$DownloadProgressToJson(_DownloadProgress instance) =>
    <String, dynamic>{
      'stage': instance.stage,
      'percentage': instance.percentage,
      'bytesDownloaded': instance.bytesDownloaded,
      'totalBytes': instance.totalBytes,
      'message': instance.message,
    };
