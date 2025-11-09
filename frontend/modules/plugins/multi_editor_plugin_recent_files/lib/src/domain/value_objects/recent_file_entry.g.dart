// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'recent_file_entry.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RecentFileEntry _$RecentFileEntryFromJson(Map<String, dynamic> json) =>
    _RecentFileEntry(
      fileId: json['fileId'] as String,
      fileName: json['fileName'] as String,
      filePath: json['filePath'] as String,
      lastOpened: DateTime.parse(json['lastOpened'] as String),
    );

Map<String, dynamic> _$RecentFileEntryToJson(_RecentFileEntry instance) =>
    <String, dynamic>{
      'fileId': instance.fileId,
      'fileName': instance.fileName,
      'filePath': instance.filePath,
      'lastOpened': instance.lastOpened.toIso8601String(),
    };
