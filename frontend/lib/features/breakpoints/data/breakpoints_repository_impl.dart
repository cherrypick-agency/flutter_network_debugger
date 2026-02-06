import '../domain/entities/decisions.dart';
import '../domain/entities/intercept_config.dart';
import '../domain/entities/intercept_item.dart';
import '../domain/entities/intercept_rule.dart';
import '../domain/repositories/breakpoints_repository.dart';
import 'breakpoints_api.dart';

class BreakpointsRepositoryImpl implements BreakpointsRepository {
  BreakpointsRepositoryImpl(this._api);
  final BreakpointsApi _api;

  @override
  Future<InterceptConfig> getConfig() async {
    final m = await _api.getConfig();
    return InterceptConfig.fromJson(m);
  }

  @override
  Future<void> setConfig(InterceptConfig cfg) async {
    await _api.setConfig(cfg.toJson());
  }

  @override
  Future<List<InterceptRule>> listRules() async {
    final list = await _api.listRules();
    return list
        .whereType<Map>()
        .map((e) => InterceptRule.fromJson(Map<String, dynamic>.from(e)))
        .toList();
  }

  @override
  Future<void> replaceRules(List<InterceptRule> rules) async {
    await _api.replaceRules(rules.map((r) => r.toJson()).toList());
  }

  @override
  Future<List<InterceptItem>> listPending({int? limit}) async {
    final list = await _api.listPending(limit: limit);
    return list
        .whereType<Map>()
        .map((e) => InterceptItem.fromJson(Map<String, dynamic>.from(e)))
        .toList();
  }

  @override
  Future<InterceptItem> getItem(String id) async {
    final m = await _api.getItem(id);
    return InterceptItem.fromJson(m);
  }

  @override
  Future<void> continueRequest(String id, RequestDecision decision) async {
    await _api.continueItem(id, decision.toJson());
  }

  @override
  Future<void> continueResponse(String id, ResponseDecision decision) async {
    await _api.continueItem(id, decision.toJson());
  }

  @override
  Future<void> cancel(String id) async {
    await _api.cancel(id);
  }
}
