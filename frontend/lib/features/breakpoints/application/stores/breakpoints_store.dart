import 'package:flutter/foundation.dart';
import '../../domain/entities/intercept_config.dart';
import '../../domain/entities/intercept_rule.dart';
import '../../domain/repositories/breakpoints_repository.dart';

class BreakpointsStore extends ChangeNotifier {
  BreakpointsStore(this._repo);
  final BreakpointsRepository _repo;

  InterceptConfig? _config;
  List<InterceptRule> _rules = const [];
  bool _loading = false;

  InterceptConfig? get config => _config;
  List<InterceptRule> get rules => _rules;
  bool get loading => _loading;

  InterceptConfig _sanitizeConfig(InterceptConfig c) {
    final prev = _config;
    final timeoutMs = c.timeoutMs > 0
        ? c.timeoutMs
        : (prev?.timeoutMs ?? c.timeoutMs);
    final queueMax = c.queueMax >= 0 ? c.queueMax : (prev?.queueMax ?? 0);
    final bodyMaxBytes = c.bodyMaxBytes > 0
        ? c.bodyMaxBytes
        : (prev?.bodyMaxBytes ?? c.bodyMaxBytes);
    final overflow =
        (c.overflow == 'auto-continue-oldest' || c.overflow == 'drop-new')
        ? c.overflow
        : (prev?.overflow ?? 'auto-continue-oldest');
    return InterceptConfig(
      enabled: c.enabled,
      requests: c.requests,
      responses: c.responses,
      timeoutMs: timeoutMs,
      queueMax: queueMax,
      bodyMaxBytes: bodyMaxBytes,
      reencode: c.reencode,
      overflow: overflow,
    );
  }

  Future<void> load() async {
    _loading = true;
    notifyListeners();
    try {
      _config = await _repo.getConfig();
      _rules = await _repo.listRules();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> saveConfig(InterceptConfig c) async {
    final sanitized = _sanitizeConfig(c);
    await _repo.setConfig(sanitized);
    _config = sanitized;
    notifyListeners();
  }

  Future<void> replaceRules(List<InterceptRule> list) async {
    await _repo.replaceRules(list);
    _rules = list;
    notifyListeners();
  }
}
