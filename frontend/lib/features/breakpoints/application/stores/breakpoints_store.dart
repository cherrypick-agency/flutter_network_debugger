import 'package:mobx/mobx.dart';

import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../domain/entities/intercept_config.dart';
import '../../domain/entities/intercept_rule.dart';
import '../../domain/repositories/breakpoints_repository.dart';

part 'breakpoints_store.g.dart';

class BreakpointsStore = _BreakpointsStore with _$BreakpointsStore;

abstract class _BreakpointsStore with Store {
  final BreakpointsRepository _repo;

  _BreakpointsStore(this._repo);

  @observable
  InterceptConfig? config;

  @observable
  ObservableList<InterceptRule> rules = ObservableList<InterceptRule>();

  @observable
  bool loading = false;

  @observable
  String? lastError;

  @action
  Future<void> load() async {
    loading = true;
    lastError = null;
    try {
      config = await _repo.getConfig();
      rules = ObservableList.of(await _repo.listRules());
    } catch (e) {
      lastError = e.toString();
      sl<NotificationsService>().error('Load config', e.toString());
    } finally {
      loading = false;
    }
  }

  @action
  Future<void> saveConfig(InterceptConfig cfg) async {
    final sanitized = _sanitizeConfig(cfg);
    try {
      await _repo.setConfig(sanitized);
      config = sanitized;
    } catch (e) {
      sl<NotificationsService>().error('Save config', e.toString());
      rethrow;
    }
  }

  @action
  Future<void> replaceRules(List<InterceptRule> newRules) async {
    try {
      await _repo.replaceRules(newRules);
      rules = ObservableList.of(newRules);
    } catch (e) {
      sl<NotificationsService>().error('Save rules', e.toString());
      rethrow;
    }
  }

  InterceptConfig _sanitizeConfig(InterceptConfig c) {
    final prev = config;
    var timeoutMs =
        c.timeoutMs > 0 ? c.timeoutMs : (prev?.timeoutMs ?? c.timeoutMs);
    if (timeoutMs > 300000) timeoutMs = 300000;
    var queueMax = c.queueMax >= 0 ? c.queueMax : (prev?.queueMax ?? 0);
    if (queueMax > 10000) queueMax = 10000;
    var bodyMaxBytes = c.bodyMaxBytes > 0
        ? c.bodyMaxBytes
        : (prev?.bodyMaxBytes ?? c.bodyMaxBytes);
    if (bodyMaxBytes > 104857600) bodyMaxBytes = 104857600;
    final overflow =
        (c.overflow == 'auto-continue-oldest' || c.overflow == 'drop-new')
            ? c.overflow
            : (prev?.overflow ?? 'auto-continue-oldest');
    return c.copyWith(
      timeoutMs: timeoutMs,
      queueMax: queueMax,
      bodyMaxBytes: bodyMaxBytes,
      overflow: overflow,
    );
  }
}
