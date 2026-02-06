import 'package:dio/dio.dart';
import 'package:app_http_client/application/app_http_client.dart';

/// API service for Scripts REST endpoints
/// Handles HTTP communication with backend
class ScriptsApiService {
  final AppHttpClient _httpClient;
  static const String _basePath = '/_api/v1/scripts';

  ScriptsApiService(this._httpClient);

  /// GET /_api/v1/scripts
  /// Retrieve all scripts
  Future<List<Map<String, dynamic>>> getAll() async {
    final response = await _httpClient.get(path: _basePath);
    return List<Map<String, dynamic>>.from(response.data as List);
  }

  /// GET /_api/v1/scripts/{id}
  /// Retrieve script by ID
  Future<Map<String, dynamic>> getById(String id) async {
    final response = await _httpClient.get(path: '$_basePath/$id');
    return response.data as Map<String, dynamic>;
  }

  /// POST /_api/v1/scripts
  /// Create new script
  Future<Map<String, dynamic>> create(Map<String, dynamic> scriptJson) async {
    final response = await _httpClient.post(path: _basePath, body: scriptJson);
    return response.data as Map<String, dynamic>;
  }

  /// PUT /_api/v1/scripts/{id}
  /// Update existing script
  Future<Map<String, dynamic>> update(
    String id,
    Map<String, dynamic> scriptJson,
  ) async {
    final response = await _httpClient.put(
      path: '$_basePath/$id',
      body: scriptJson,
    );
    return response.data as Map<String, dynamic>;
  }

  /// DELETE /_api/v1/scripts/{id}
  /// Delete script
  Future<void> delete(String id) async {
    await _httpClient.delete(path: '$_basePath/$id');
  }

  /// PATCH /_api/v1/scripts/{id}/toggle
  /// Toggle script enabled status
  Future<Map<String, dynamic>> toggle(String id, bool enabled) async {
    final response = await _httpClient.patch(
      path: '$_basePath/$id/toggle',
      body: {'enabled': enabled},
    );
    return response.data as Map<String, dynamic>;
  }

  /// POST /_api/v1/scripts/test
  /// Test script execution with sample request
  Future<Map<String, dynamic>> test(Map<String, dynamic> testPayload) async {
    final response = await _httpClient.post(
      path: '$_basePath/test',
      body: testPayload,
    );
    return response.data as Map<String, dynamic>;
  }

  /// GET /_api/v1/scripts/examples
  /// Retrieve example scripts library
  Future<List<Map<String, dynamic>>> getExamples() async {
    final response = await _httpClient.get(path: '$_basePath/examples');
    return List<Map<String, dynamic>>.from(response.data as List);
  }

  /// POST /_api/v1/scripts/{id}/compile
  /// Compile script source code to WASM
  Future<Map<String, dynamic>> compile(
    String id, {
    bool optimize = true,
  }) async {
    final response = await _httpClient.post(
      path: '$_basePath/$id/compile',
      body: {'optimize': optimize},
    );
    return response.data as Map<String, dynamic>;
  }

  /// GET /_api/v1/scripts/compilers
  /// Compiler availability (from system or cache) for unlocking compilation
  /// Returns map language -> canCompile
  Future<Map<String, bool>> getCompilersAvailability() async {
    final response = await _httpClient.get(path: '$_basePath/compilers');
    final data = response.data;
    if (data is Map<String, dynamic>) {
      final all = (data['all'] as Map?)?.cast<String, dynamic>() ?? const {};
      return all.map((k, v) => MapEntry(k.toLowerCase(), v == true));
    }
    return {};
  }

  /// GET /_api/v1/scripts/{id}/export-zip
  /// Export script as ZIP file (metadata + source + dependencies + WASM)
  /// Returns the download URL for the ZIP file
  String getExportZipUrl(String id) {
    return '$_basePath/$id/export-zip';
  }

  /// POST /_api/v1/scripts/import-zip
  /// Import script from ZIP file
  /// Note: This needs to be handled via FormData in the UI layer
  /// fileBytes: ZIP file contents as bytes
  /// Returns the created script
  String get importZipUrl => '$_basePath/import-zip';

  /// POST /_api/v1/scripts/import-zip (with FormData)
  /// Import script from ZIP file using multipart upload
  Future<Map<String, dynamic>> importFromZip(List<int> zipBytes) async {
    final response = await _httpClient.filesPost(
      path: '$_basePath/import-zip',
      files: [
        MultipartFile.fromBytes(zipBytes, filename: 'script.zip'),
      ],
    );
    return response.data as Map<String, dynamic>;
  }

  /// POST /_api/v1/scripts/{id}/upload-project
  /// Upload ZIP with project files
  Future<Map<String, dynamic>> uploadProject(String id, List<int> zipBytes) async {
    final response = await _httpClient.filesPost(
      path: '$_basePath/$id/upload-project',
      files: [
        MultipartFile.fromBytes(zipBytes, filename: 'project.zip'),
      ],
    );
    return response.data as Map<String, dynamic>;
  }

  /// GET /_api/v1/scripts/{id}/download-project
  /// Returns download URL for project ZIP
  String getDownloadProjectUrl(String id) {
    return '$_basePath/$id/download-project';
  }

  /// POST /_api/v1/scripts/validate
  /// Validate script syntax without compilation
  Future<Map<String, dynamic>> validateSyntax({
    required String sourceCode,
    required String language,
    Map<String, String>? dependencies,
  }) async {
    final response = await _httpClient.post(
      path: '$_basePath/validate',
      body: {
        'sourceCode': sourceCode,
        'language': language,
        if (dependencies != null) 'dependencies': dependencies,
      },
    );
    return response.data as Map<String, dynamic>;
  }

  /// GET /_api/v1/scripts/{id}/files
  /// List all files in script project
  Future<Map<String, dynamic>> listProjectFiles(String id) async {
    final response = await _httpClient.get(path: '$_basePath/$id/files');
    return response.data as Map<String, dynamic>;
  }
}
