import '../entities/intercept_config.dart';
import '../entities/intercept_item.dart';
import '../entities/intercept_rule.dart';
import '../entities/decisions.dart';

abstract class BreakpointsRepository {
  Future<InterceptConfig> getConfig();
  Future<void> setConfig(InterceptConfig cfg);

  Future<List<InterceptRule>> listRules();
  Future<void> replaceRules(List<InterceptRule> rules);

  Future<List<InterceptItem>> listPending({int? limit});
  Future<InterceptItem> getItem(String id);
  Future<void> continueRequest(String id, RequestDecision decision);
  Future<void> continueResponse(String id, ResponseDecision decision);
  Future<void> cancel(String id);
}
