import 'package:mobx/mobx.dart';

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

  @action
  Future<void> load() async {
    loading = true;
    try {
      config = await _repo.getConfig();
      rules = ObservableList.of(await _repo.listRules());
    } finally {
      loading = false;
    }
  }

  @action
  Future<void> saveConfig(InterceptConfig cfg) async {
    final sanitized = _sanitizeConfig(cfg);
    await _repo.setConfig(sanitized);
    config = sanitized;
  }

  @action
  Future<void> replaceRules(List<InterceptRule> newRules) async {
    await _repo.replaceRules(newRules);
    rules = ObservableList.of(newRules);
  }

  InterceptConfig _sanitizeConfig(InterceptConfig c) {
    final prev = config;
    final timeoutMs =
        c.timeoutMs > 0 ? c.timeoutMs : (prev?.timeoutMs ?? c.timeoutMs);
    final queueMax = c.queueMax >= 0 ? c.queueMax : (prev?.queueMax ?? 0);
    final bodyMaxBytes = c.bodyMaxBytes > 0
        ? c.bodyMaxBytes
        : (prev?.bodyMaxBytes ?? c.bodyMaxBytes);
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
