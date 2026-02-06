import 'package:flutter/foundation.dart';
import '../../domain/mapping_config.dart';
import '../../domain/repositories/mapping_repository.dart';

class MappingConfigStore extends ChangeNotifier {
  MappingConfigStore(this._repo);
  final MappingRepository _repo;

  MappingConfig? _config;
  bool _loading = false;
  String? _lastError;

  MappingConfig? get config => _config;
  bool get loading => _loading;
  String? get lastError => _lastError;

  Future<void> load() async {
    _loading = true;
    _lastError = null;
    notifyListeners();
    try {
      _config = await _repo.getConfig();
    } catch (e) {
      _lastError = e.toString();
      rethrow;
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> save(MappingConfig c) async {
    _lastError = null;
    try {
      final response = await _repo.setConfig(c);
      _config = response;
      notifyListeners();
    } catch (e) {
      _lastError = e.toString();
      notifyListeners();
      rethrow;
    }
  }
}
