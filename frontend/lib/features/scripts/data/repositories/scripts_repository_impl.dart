import '../../domain/entities/script.dart';
import '../../domain/entities/script_test_result.dart';
import '../../domain/repositories/scripts_repository.dart';
import '../services/scripts_api_service.dart';

/// Implementation of ScriptsRepository
/// Handles data mapping and API calls
class ScriptsRepositoryImpl implements ScriptsRepository {
  final ScriptsApiService _apiService;

  ScriptsRepositoryImpl(this._apiService);

  @override
  Future<List<Script>> getAll() async {
    final jsonList = await _apiService.getAll();
    return jsonList.map((json) => Script.fromJson(json)).toList();
  }

  @override
  Future<Script> getById(String id) async {
    final json = await _apiService.getById(id);
    return Script.fromJson(json);
  }

  @override
  Future<Script> create(Script script) async {
    final json = script.toJson();
    final responseJson = await _apiService.create(json);
    // Бэкенд в режиме source/dependencies может вернуть только { "id": "..." }.
    // В таком случае дотягиваем полноценный объект отдельным запросом.
    final maybeName = responseJson['name'];
    final maybeRuntime = responseJson['runtime'];
    final maybeLanguage = responseJson['language'];
    final maybeTrigger = responseJson['triggerType'];
    final hasFullEntity =
        maybeName != null &&
        maybeRuntime != null &&
        maybeLanguage != null &&
        maybeTrigger != null;

    if (hasFullEntity) {
      return Script.fromJson(responseJson);
    }

    final id = responseJson['id'];
    if (id is String && id.isNotEmpty) {
      final fetched = await _apiService.getById(id);
      return Script.fromJson(fetched);
    }

    // В крайнем случае пробуем распарсить как есть (даст понятную ошибку, если формат неожиданный)
    return Script.fromJson(responseJson);
  }

  @override
  Future<Script> update(String id, Script script) async {
    final json = script.toJson();
    final responseJson = await _apiService.update(id, json);
    return Script.fromJson(responseJson);
  }

  @override
  Future<void> delete(String id) async {
    await _apiService.delete(id);
  }

  @override
  Future<void> toggle(String id, bool enabled) async {
    await _apiService.toggle(id, enabled);
  }

  @override
  Future<ScriptTestResult> test(Script script, TestRequest testRequest) async {
    final payload = {
      'script': script.toJson(),
      'testRequest': testRequest.toJson(),
    };

    final responseJson = await _apiService.test(payload);
    return ScriptTestResult.fromJson(responseJson);
  }

  @override
  Future<Script> compile(String id, {bool optimize = true}) async {
    // 1) Запускаем компиляцию (API возвращает статус/логи)
    await _apiService.compile(id, optimize: optimize);
    // 2) Дотягиваем актуальный объект скрипта после компиляции
    final json = await _apiService.getById(id);
    return Script.fromJson(json);
  }

  @override
  String getExportZipUrl(String id) {
    return _apiService.getExportZipUrl(id);
  }

  @override
  Future<Script> importFromZip(List<int> zipBytes) async {
    final responseJson = await _apiService.importFromZip(zipBytes);
    return Script.fromJson(responseJson);
  }
}
