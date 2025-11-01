import 'package:flutter/foundation.dart';
import '../../domain/mapping_config.dart';
import '../../domain/repositories/mapping_repository.dart';

class MappingConfigStore extends ChangeNotifier {
  MappingConfigStore(this._repo);
  final MappingRepository _repo;

  MappingConfig? _config;
  bool _loading = false;

  MappingConfig? get config => _config;
  bool get loading => _loading;

  Future<void> load() async {
    _loading = true;
    notifyListeners();
    try {
      _config = await _repo.getConfig();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> save(MappingConfig c) async {
    await _repo.setConfig(c);
    _config = c;
    notifyListeners();
  }
}
