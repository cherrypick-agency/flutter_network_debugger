import '../mapping_config.dart';
import '../mapping_rule.dart';

abstract class MappingRepository {
  Future<MappingConfig> getConfig();
  Future<MappingConfig> setConfig(MappingConfig cfg);

  Future<List<MappingRule>> listRules();
  Future<MappingRule> upsert(MappingRule rule);
  Future<void> reorder(List<String> ids);
  Future<void> delete(String id);

  Future<Map<String, dynamic>> uploadFile(String fileName, List<int> bytes);

  Future<Map<String, dynamic>> uploadStream({
    required String fileName,
    required Stream<List<int>> stream,
    required int length,
  });
}
