class MappingRule {
  MappingRule({
    required this.id,
    required this.enabled,
    required this.priority,
    required this.kind,
    required this.stopProcessing,
    required this.methods,
    required this.hostPattern,
    required this.pathPattern,
    required this.patternType,
    this.filePath,
    this.blobPath,
    this.statusOverride,
    this.contentTypeOverride,
    this.targetURLTemplate,
    this.preserveHost = false,
  });

  final String id;
  final bool enabled;
  final int priority;
  final String kind; // local|remote
  final bool stopProcessing;
  final List<String> methods;
  final String hostPattern;
  final String pathPattern;
  final String patternType; // glob|regex
  final String? filePath;
  final String? blobPath;
  final int? statusOverride;
  final String? contentTypeOverride;
  final String? targetURLTemplate;
  final bool preserveHost;

  Map<String, dynamic> toJson() => {
    'id': id,
    'enabled': enabled,
    'priority': priority,
    'kind': kind,
    'stopProcessing': stopProcessing,
    'methods': methods,
    'hostPattern': hostPattern,
    'pathPattern': pathPattern,
    'patternType': patternType,
    'filePath': filePath,
    'blobPath': blobPath,
    // Бэк сейчас не принимает null для этих полей, поэтому отправляем безопасные
    // значения по умолчанию.
    'statusOverride': statusOverride ?? 200,
    'contentTypeOverride': contentTypeOverride ?? '',
    'targetURLTemplate': targetURLTemplate ?? '',
    'preserveHost': preserveHost,
  };

  static MappingRule fromJson(Map<String, dynamic> j) => MappingRule(
    id: (j['id'] ?? '') as String,
    enabled: (j['enabled'] ?? true) as bool,
    priority: (j['priority'] ?? 0) as int,
    kind: (j['kind'] ?? 'remote') as String,
    stopProcessing: (j['stopProcessing'] ?? true) as bool,
    methods: ((j['methods'] as List<dynamic>?) ?? const <dynamic>[])
        .map((e) => (e ?? '').toString())
        .where((e) => e.isNotEmpty)
        .toList(),
    hostPattern: (j['hostPattern'] ?? '') as String,
    pathPattern: (j['pathPattern'] ?? '') as String,
    patternType: (j['patternType'] ?? 'glob') as String,
    filePath: (j['filePath'] as String?),
    blobPath: (j['blobPath'] as String?),
    statusOverride: (j['statusOverride'] as int?),
    contentTypeOverride: (j['contentTypeOverride'] as String?),
    targetURLTemplate: (j['targetURLTemplate'] as String?),
    preserveHost: (j['preserveHost'] ?? false) as bool,
  );
}
