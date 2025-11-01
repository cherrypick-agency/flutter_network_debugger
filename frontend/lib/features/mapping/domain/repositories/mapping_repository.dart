import '../mapping_config.dart';
import '../mapping_rule.dart';

abstract class MappingRepository {
  Future<MappingConfig> getConfig();
  Future<void> setConfig(MappingConfig cfg);

  Future<List<MappingRule>> listRules();
  Future<MappingRule> upsert(MappingRule rule);
  Future<void> reorder(List<String> ids);
  Future<void> delete(String id);

  Future<Map<String, dynamic>> uploadFile(String fileName, List<int> bytes);
}
