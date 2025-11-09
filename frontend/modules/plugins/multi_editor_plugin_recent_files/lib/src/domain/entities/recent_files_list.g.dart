// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'recent_files_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RecentFilesList _$RecentFilesListFromJson(Map<String, dynamic> json) =>
    _RecentFilesList(
      entries:
          (json['entries'] as List<dynamic>?)
              ?.map((e) => RecentFileEntry.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      maxEntries: (json['maxEntries'] as num?)?.toInt() ?? 10,
    );

Map<String, dynamic> _$RecentFilesListToJson(_RecentFilesList instance) =>
    <String, dynamic>{
      'entries': instance.entries,
      'maxEntries': instance.maxEntries,
    };
