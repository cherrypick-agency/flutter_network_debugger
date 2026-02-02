import 'package:flutter/foundation.dart';
import '../../domain/mapping_rule.dart';
import '../../domain/repositories/mapping_repository.dart';

class MappingStore extends ChangeNotifier {
  MappingStore(this._repo);
  final MappingRepository _repo;

  List<MappingRule> _rules = const [];
  bool _loading = false;

  List<MappingRule> get rules => _rules;
  bool get loading => _loading;

  Future<void> load() async {
    _loading = true;
    notifyListeners();
    try {
      _rules = await _repo.listRules();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> upsert(MappingRule r) async {
    final saved = await _repo.upsert(r);
    final i = _rules.indexWhere((e) => e.id == saved.id);
    if (i >= 0) {
      _rules = List.of(_rules)..[i] = saved;
    } else {
      _rules = [..._rules, saved];
    }
    final sorted = _rules.toList()
      ..sort((a, b) => a.priority.compareTo(b.priority));
    _rules = sorted;
    notifyListeners();
  }

  Future<void> delete(String id) async {
    await _repo.delete(id);
    _rules = _rules.where((e) => e.id != id).toList();
    notifyListeners();
  }

  Future<void> reorder(List<String> ids) async {
    await _repo.reorder(ids);
    // локально переставим под тот же порядок
    final map = {for (final r in _rules) r.id: r};
    _rules = ids.map((e) => map[e]).whereType<MappingRule>().toList();
    notifyListeners();
  }
}
