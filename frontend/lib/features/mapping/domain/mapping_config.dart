class MappingConfig {
  const MappingConfig({required this.enabled, required this.uploadMaxMB});
  final bool enabled;
  final int uploadMaxMB;

  Map<String, dynamic> toJson() => {
    'enabled': enabled,
    'uploadMaxMB': uploadMaxMB,
  };

  static MappingConfig fromJson(Map<String, dynamic> j) => MappingConfig(
    enabled: (j['enabled'] ?? true) as bool,
    uploadMaxMB: (j['uploadMaxMB'] ?? 20) as int,
  );
}
