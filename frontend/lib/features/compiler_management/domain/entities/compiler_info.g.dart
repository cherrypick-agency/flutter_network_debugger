// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'compiler_info.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_CompilerInfo _$CompilerInfoFromJson(Map<String, dynamic> json) =>
    _CompilerInfo(
      language: json['language'] as String,
      version: json['version'] as String,
      status: CompilerStatus.fromJson(json['status'] as Map<String, dynamic>),
      installedPath: json['installedPath'] as String,
      size: (json['size'] as num).toInt(),
      downloadSize: (json['downloadSize'] as num).toInt(),
      error: json['error'] as String? ?? '',
    );

Map<String, dynamic> _$CompilerInfoToJson(_CompilerInfo instance) =>
    <String, dynamic>{
      'language': instance.language,
      'version': instance.version,
      'status': instance.status,
      'installedPath': instance.installedPath,
      'size': instance.size,
      'downloadSize': instance.downloadSize,
      'error': instance.error,
    };
