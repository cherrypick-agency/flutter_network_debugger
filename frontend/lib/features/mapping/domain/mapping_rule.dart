import 'package:freezed_annotation/freezed_annotation.dart';

import '../../../core/utils/json_cast.dart';

part 'mapping_rule.freezed.dart';

@Freezed(toJson: false, fromJson: false)
sealed class MappingRule with _$MappingRule {
  const MappingRule._();

  const factory MappingRule({
    required String id,
    @Default(true) bool enabled,
    @Default(100) int priority,
    @Default('local') String kind,
    @Default(true) bool stopProcessing,
    @Default([]) List<String> methods,
    @Default('') String hostPattern,
    @Default('') String pathPattern,
    @Default('glob') String patternType,
    String? filePath,
    String? blobPath,
    int? statusOverride,
    String? contentTypeOverride,
    String? targetURLTemplate,
    @Default(false) bool preserveHost,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _MappingRule;

  Map<String, dynamic> toJson() {
    final m = <String, dynamic>{
      'id': id,
      'enabled': enabled,
      'priority': priority,
      'kind': kind,
      'stopProcessing': stopProcessing,
      'methods': methods,
      'hostPattern': hostPattern,
      'pathPattern': pathPattern,
      'patternType': patternType,
    };
    if (kind == 'local') {
      m['filePath'] = filePath;
      m['blobPath'] = blobPath;
      m['statusOverride'] = statusOverride ?? 200;
      m['contentTypeOverride'] = contentTypeOverride ?? '';
    }
    if (kind == 'remote') {
      m['targetURLTemplate'] = targetURLTemplate ?? '';
      m['preserveHost'] = preserveHost;
    }
    if (updatedAt != null) {
      m['updatedAt'] = updatedAt!.toIso8601String();
    }
    return m;
  }

  static MappingRule fromJson(Map<String, dynamic> j) => MappingRule(
        id: (j['id'] ?? '').toString(),
        enabled: JsonCast.asBool(j['enabled'], fallback: true),
        priority: JsonCast.asInt(j['priority'], fallback: 100),
        kind: (j['kind'] ?? 'local').toString(),
        stopProcessing: JsonCast.asBool(j['stopProcessing'], fallback: true),
        methods: ((j['methods'] as List<dynamic>?) ?? const <dynamic>[])
            .map((e) => (e ?? '').toString())
            .where((e) => e.isNotEmpty)
            .toList(),
        hostPattern: (j['hostPattern'] ?? '').toString(),
        pathPattern: (j['pathPattern'] ?? '').toString(),
        patternType: (j['patternType'] ?? 'glob').toString(),
        filePath: (j['filePath'] as String?),
        blobPath: (j['blobPath'] as String?),
        statusOverride: j['statusOverride'] == null
            ? null
            : JsonCast.asInt(j['statusOverride'], fallback: 0),
        contentTypeOverride: (j['contentTypeOverride'] as String?),
        targetURLTemplate: (j['targetURLTemplate'] as String?),
        preserveHost: JsonCast.asBool(j['preserveHost'], fallback: false),
        createdAt: j['createdAt'] != null
            ? DateTime.tryParse(j['createdAt'].toString())
            : null,
        updatedAt: j['updatedAt'] != null
            ? DateTime.tryParse(j['updatedAt'].toString())
            : null,
      );
}
