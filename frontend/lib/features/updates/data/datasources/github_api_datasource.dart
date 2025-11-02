import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:logging/logging.dart';

/// DataSource для работы с GitHub Releases API
class GitHubApiDataSource {
  final String githubOwner;
  final String githubRepo;
  final _log = Logger('GitHubApiDataSource');

  GitHubApiDataSource({required this.githubOwner, required this.githubRepo});

  /// Получает последний релиз из GitHub
  Future<Map<String, dynamic>?> fetchLatestRelease() async {
    try {
      final url = Uri.parse(
        'https://api.github.com/repos/$githubOwner/$githubRepo/releases/latest',
      );

      final response = await http
          .get(url, headers: {'Accept': 'application/vnd.github.v3+json'})
          .timeout(const Duration(seconds: 10));

      _checkRateLimit(response);

      if (response.statusCode == 403) {
        _logRateLimitExceeded(response);
        return null;
      }

      if (response.statusCode != 200) {
        _log.warning('Failed to fetch latest release: ${response.statusCode}');
        return null;
      }

      return jsonDecode(response.body) as Map<String, dynamic>;
    } catch (e, stack) {
      _log.warning('Error fetching latest release', e, stack);
      return null;
    }
  }

  /// Получает список релизов с пагинацией
  Future<List<Map<String, dynamic>>> fetchAllReleases({
    int page = 1,
    int perPage = 10,
  }) async {
    try {
      _log.info(
        'Fetching releases from GitHub (page: $page, perPage: $perPage)',
      );

      final url = Uri.parse(
        'https://api.github.com/repos/$githubOwner/$githubRepo/releases?page=$page&per_page=$perPage',
      );

      final response = await http
          .get(url, headers: {'Accept': 'application/vnd.github.v3+json'})
          .timeout(const Duration(seconds: 10));

      _checkRateLimit(response);

      if (response.statusCode == 403) {
        _logRateLimitExceeded(response);
        throw Exception('GitHub API rate limit exceeded');
      }

      if (response.statusCode != 200) {
        _log.warning('Failed to fetch releases: ${response.statusCode}');
        throw Exception('Failed to fetch releases: ${response.statusCode}');
      }

      final List<dynamic> releasesJson = jsonDecode(response.body) as List;
      return releasesJson.map((e) => e as Map<String, dynamic>).toList();
    } catch (e, stack) {
      _log.warning('Error fetching releases', e, stack);
      rethrow;
    }
  }

  /// Получает конкретный релиз по тегу
  Future<Map<String, dynamic>?> fetchReleaseByTag(String tag) async {
    try {
      _log.info('Fetching release by tag: $tag');

      final url = Uri.parse(
        'https://api.github.com/repos/$githubOwner/$githubRepo/releases/tags/$tag',
      );

      final response = await http
          .get(url, headers: {'Accept': 'application/vnd.github.v3+json'})
          .timeout(const Duration(seconds: 10));

      _checkRateLimit(response);

      if (response.statusCode == 404) {
        _log.info('Release not found: $tag');
        return null;
      }

      if (response.statusCode == 403) {
        _log.warning('GitHub API rate limit exceeded');
        return null;
      }

      if (response.statusCode != 200) {
        _log.warning('Failed to fetch release: ${response.statusCode}');
        return null;
      }

      return jsonDecode(response.body) as Map<String, dynamic>;
    } catch (e, stack) {
      _log.warning('Error fetching release by tag', e, stack);
      return null;
    }
  }

  /// Выбирает подходящий asset для текущей платформы
  Map<String, dynamic>? selectPlatformAsset(List assets) {
    if (assets.isEmpty) return null;

    // Эта логика будет использоваться из repository impl
    // Здесь только возвращаем первый asset для демонстрации
    // В реальном коде нужно проверять Platform.isMacOS/Windows/Linux
    return null; // Будет реализовано в repository
  }

  /// Проверяет rate limit и логирует предупреждение
  void _checkRateLimit(http.Response response) {
    final remaining = response.headers['x-ratelimit-remaining'];
    if (remaining != null) {
      final limit = int.tryParse(remaining);
      if (limit != null && limit < 10) {
        _log.warning('GitHub API rate limit low: $limit requests remaining');
      }
    }
  }

  /// Логирует информацию о превышении rate limit
  void _logRateLimitExceeded(http.Response response) {
    final resetTime = response.headers['x-ratelimit-reset'];
    if (resetTime != null) {
      final reset = DateTime.fromMillisecondsSinceEpoch(
        int.parse(resetTime) * 1000,
      );
      _log.warning('GitHub API rate limit exceeded. Resets at: $reset');
    }
  }
}
