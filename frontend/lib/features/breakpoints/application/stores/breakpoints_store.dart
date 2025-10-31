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
    await _repo.setConfig(c);
    _config = c;
    notifyListeners();
  }

  Future<void> replaceRules(List<InterceptRule> list) async {
    await _repo.replaceRules(list);
    _rules = list;
    notifyListeners();
  }
}
