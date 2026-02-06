import 'package:flutter/foundation.dart';
import '../../domain/mapping_rule.dart';
import '../../domain/repositories/mapping_repository.dart';

class MappingStore extends ChangeNotifier {
  MappingStore(this._repo);
  final MappingRepository _repo;

  List<MappingRule> _rules = const [];
  bool _loading = false;
  String? _lastError;

  List<MappingRule> get rules => _rules;
  bool get loading => _loading;
  String? get lastError => _lastError;

  Future<void> load() async {
    _loading = true;
    _lastError = null;
    notifyListeners();
    try {
      _rules = await _repo.listRules();
    } catch (e) {
      _lastError = e.toString();
      rethrow;
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> upsert(MappingRule r) async {
    _lastError = null;
    try {
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
    } catch (e) {
      _lastError = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  Future<void> delete(String id) async {
    _lastError = null;
    try {
      await _repo.delete(id);
      _rules = _rules.where((e) => e.id != id).toList();
      notifyListeners();
    } catch (e) {
      _lastError = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  Future<void> reorder(List<String> ids) async {
    _lastError = null;
    try {
      await _repo.reorder(ids);
      // After reorder, priorities change on the backend. It's simpler and more reliable to reload.
      await load();
    } catch (e) {
      _lastError = e.toString();
      notifyListeners();
      rethrow;
    }
  }
}
