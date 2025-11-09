import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/compiler_info_model.dart';
import '../models/download_progress_model.dart';

class CompilerApi {
  final String baseUrl;
  final http.Client client;

  CompilerApi({required this.baseUrl, http.Client? client})
    : client = client ?? http.Client();

  Future<List<CompilerInfoModel>> getCompilers() async {
    final response = await client.get(
      Uri.parse('$baseUrl/_api/v1/compilers/status'),
    );

    if (response.statusCode == 200) {
      final Map<String, dynamic> json = jsonDecode(response.body);
      final List<dynamic> compilers = json['compilers'] ?? [];
      return compilers.map((e) => CompilerInfoModel.fromJson(e)).toList();
    } else {
      throw Exception('Failed to load compilers: ${response.statusCode}');
    }
  }

  Future<CompilerInfoModel> getCompiler(String language) async {
    final response = await client.get(
      Uri.parse('$baseUrl/_api/v1/compilers/$language/status'),
    );

    if (response.statusCode == 200) {
      return CompilerInfoModel.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to load compiler: ${response.statusCode}');
    }
  }

  Future<void> installCompiler(String language) async {
    final response = await client.post(
      Uri.parse('$baseUrl/_api/v1/compilers/$language/install'),
    );

    if (response.statusCode != 200 && response.statusCode != 202) {
      throw Exception('Failed to install compiler: ${response.statusCode}');
    }
  }

  Future<void> uninstallCompiler(String language) async {
    final response = await client.delete(
      Uri.parse('$baseUrl/_api/v1/compilers/$language'),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to uninstall compiler: ${response.statusCode}');
    }
  }

  Stream<DownloadProgressModel> watchProgress(String language) async* {
    final uri = Uri.parse('$baseUrl/_api/v1/compilers/$language/progress');

    final request = http.Request('GET', uri);
    request.headers['Accept'] = 'text/event-stream';
    request.headers['Cache-Control'] = 'no-cache';

    final response = await client.send(request);

    if (response.statusCode != 200) {
      throw Exception('Failed to watch progress: ${response.statusCode}');
    }

    await for (final chunk in response.stream.transform(utf8.decoder)) {
      final lines = chunk.split('\n');
      for (final line in lines) {
        if (line.startsWith('data: ')) {
          final data = line.substring(6).trim();
          if (data.isNotEmpty) {
            try {
              final json = jsonDecode(data);
              yield DownloadProgressModel.fromJson(json);
            } catch (e) {
              // Skip malformed JSON
            }
          }
        }
      }
    }
  }

  Future<int> getCacheSize() async {
    final response = await client.get(
      Uri.parse('$baseUrl/_api/v1/compilers/cache/size'),
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body);
      return json['size'] as int? ?? 0;
    } else {
      throw Exception('Failed to get cache size: ${response.statusCode}');
    }
  }

  Future<void> clearCache() async {
    final response = await client.delete(
      Uri.parse('$baseUrl/_api/v1/compilers/cache'),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to clear cache: ${response.statusCode}');
    }
  }
}
