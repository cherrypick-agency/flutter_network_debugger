// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'file_statistics.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_FileStatistics _$FileStatisticsFromJson(Map<String, dynamic> json) =>
    _FileStatistics(
      fileId: json['fileId'] as String,
      lines: (json['lines'] as num).toInt(),
      characters: (json['characters'] as num).toInt(),
      words: (json['words'] as num).toInt(),
      bytes: (json['bytes'] as num).toInt(),
      calculatedAt: DateTime.parse(json['calculatedAt'] as String),
    );

Map<String, dynamic> _$FileStatisticsToJson(_FileStatistics instance) =>
    <String, dynamic>{
      'fileId': instance.fileId,
      'lines': instance.lines,
      'characters': instance.characters,
      'words': instance.words,
      'bytes': instance.bytes,
      'calculatedAt': instance.calculatedAt.toIso8601String(),
    };
