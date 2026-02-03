import '../domain/mapping_config.dart';
import '../domain/mapping_rule.dart';
import '../domain/repositories/mapping_repository.dart';
import 'mapping_api.dart';

class MappingRepositoryImpl implements MappingRepository {
  MappingRepositoryImpl(this._api);
  final MappingApi _api;

  @override
  Future<MappingConfig> getConfig() async {
    final m = await _api.getConfig();
    return MappingConfig.fromJson(m);
  }

  @override
  Future<void> setConfig(MappingConfig cfg) async {
    await _api.setConfig(cfg.toJson());
  }

  @override
  Future<List<MappingRule>> listRules() async {
    final list = await _api.listRules();
    return list
        .whereType<Map<String, dynamic>>()
        .map(MappingRule.fromJson)
        .toList();
  }

  @override
  Future<MappingRule> upsert(MappingRule rule) async {
    final m = await _api.upsertRule(rule.toJson());
    return MappingRule.fromJson(m);
  }

  @override
  Future<void> reorder(List<String> ids) async {
    await _api.reorder(ids);
  }

  @override
  Future<void> delete(String id) async {
    await _api.deleteRule(id);
  }

  @override
  Future<Map<String, dynamic>> uploadFile(
    String fileName,
    List<int> bytes,
  ) async {
    return _api.upload(fileName, bytes);
  }

  @override
  Future<Map<String, dynamic>> uploadStream({
    required String fileName,
    required Stream<List<int>> stream,
    required int length,
  }) async {
    return _api.uploadStream(
      fileName: fileName,
      stream: stream,
      length: length,
    );
  }
}
