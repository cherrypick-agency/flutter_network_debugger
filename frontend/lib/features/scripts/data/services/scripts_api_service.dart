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
  Future<void> toggle(String id, bool enabled) async {
    await _httpClient.patch(
      path: '$_basePath/$id/toggle',
      body: {'enabled': enabled},
    );
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
    try {
      // Create FormData with the ZIP file
      final formData = FormData.fromMap({
        'file': MultipartFile.fromBytes(zipBytes, filename: 'script.zip'),
      });

      // Use dio directly for multipart upload (AppHttpClient may not support it)
      // Get the dio instance from _httpClient if available, otherwise create new
      final dio = Dio();

      // Get base URL from somewhere (we need to construct full URL)
      // AppHttpClient probably has baseUrl, but we can't access it
      // For now, use relative path and dio will use current host
      final response = await dio.post(
        importZipUrl,
        data: formData,
        options: Options(headers: {'Content-Type': 'multipart/form-data'}),
      );

      return response.data as Map<String, dynamic>;
    } on DioException catch (e) {
      // Translate DioException to user-friendly error messages
      if (e.type == DioExceptionType.connectionTimeout ||
          e.type == DioExceptionType.sendTimeout ||
          e.type == DioExceptionType.receiveTimeout) {
        throw Exception(
          'Connection timeout. Please check your network connection.',
        );
      } else if (e.type == DioExceptionType.connectionError) {
        throw Exception('Cannot connect to server. Is the backend running?');
      } else if (e.response != null) {
        final statusCode = e.response!.statusCode;
        final responseData = e.response!.data;

        if (statusCode == 413) {
          throw Exception('ZIP file is too large. Maximum size is 10MB.');
        } else if (statusCode == 400) {
          // Try to extract error message from response
          String errorMsg = 'Invalid ZIP file or request.';
          if (responseData is Map && responseData.containsKey('error')) {
            errorMsg = responseData['error'].toString();
          }
          throw Exception(errorMsg);
        } else if (statusCode == 404) {
          throw Exception(
            'Import endpoint not found. Please update your backend.',
          );
        } else if (statusCode == 500) {
          // Try to extract error message from server
          String errorMsg = 'Server error while processing ZIP file.';
          if (responseData is Map && responseData.containsKey('error')) {
            errorMsg = responseData['error'].toString();
          }
          throw Exception(errorMsg);
        } else {
          throw Exception('Upload failed with status code $statusCode');
        }
      } else {
        throw Exception('Network error: ${e.message ?? "Unknown error"}');
      }
    } catch (e) {
      // Re-throw if already an Exception, otherwise wrap it
      if (e is Exception) {
        rethrow;
      }
      throw Exception('Failed to import script: ${e.toString()}');
    }
  }
}
