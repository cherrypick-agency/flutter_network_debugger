import 'dart:convert';
import 'package:http/http.dart' as http;
import 'logger.dart';

/// Handles interactions with GitHub Releases API.
class GitHubRelease {
  final String owner;
  final String repo;
  final http.Client? _client;
  final Logger _logger = Logger('GitHubRelease');

  GitHubRelease({
    required this.owner,
    required this.repo,
    http.Client? client,
  }) : _client = client;

  http.Client get client => _client ?? http.Client();

  /// Fetches the latest release from GitHub.
  Future<ReleaseInfo> getLatestRelease({
    Duration timeout = const Duration(seconds: 30),
  }) async {
    _logger.debug('Fetching latest release from $owner/$repo');

    final url = Uri.parse(
      'https://api.github.com/repos/$owner/$repo/releases/latest',
    );

    try {
      final response = await client.get(
        url,
        headers: {
          'Accept': 'application/vnd.github.v3+json',
        },
      ).timeout(
        timeout,
        onTimeout: () {
          _logger.error('Timeout fetching latest release');
          throw GitHubReleaseException(
            'Timeout fetching latest release after ${timeout.inSeconds}s',
          );
        },
      );

      if (response.statusCode != 200) {
        _logger.error('Failed to fetch latest release: ${response.statusCode}');
        throw GitHubReleaseException(
          'Failed to fetch latest release: ${response.statusCode} ${response.reasonPhrase}',
        );
      }

      final json = jsonDecode(response.body) as Map<String, dynamic>;
      final release = ReleaseInfo.fromJson(json);
      _logger.info('Found latest release: ${release.tagName}');
      return release;
    } on GitHubReleaseException {
      rethrow;
    } catch (e) {
      throw GitHubReleaseException(
        'Error fetching latest release: $e',
      );
    }
  }

  /// Fetches a specific release by tag name.
  Future<ReleaseInfo> getRelease(
    String tagName, {
    Duration timeout = const Duration(seconds: 30),
  }) async {
    final url = Uri.parse(
      'https://api.github.com/repos/$owner/$repo/releases/tags/$tagName',
    );

    try {
      final response = await client.get(
        url,
        headers: {
          'Accept': 'application/vnd.github.v3+json',
        },
      ).timeout(
        timeout,
        onTimeout: () {
          throw GitHubReleaseException(
            'Timeout fetching release "$tagName" after ${timeout.inSeconds}s',
          );
        },
      );

      if (response.statusCode == 404) {
        throw GitHubReleaseException(
          'Release with tag "$tagName" not found',
        );
      }

      if (response.statusCode != 200) {
        throw GitHubReleaseException(
          'Failed to fetch release: ${response.statusCode} ${response.reasonPhrase}',
        );
      }

      final json = jsonDecode(response.body) as Map<String, dynamic>;
      return ReleaseInfo.fromJson(json);
    } on GitHubReleaseException {
      rethrow;
    } catch (e) {
      throw GitHubReleaseException(
        'Error fetching release "$tagName": $e',
      );
    }
  }

  /// Finds the download URL for a specific asset by name.
  String? findAssetUrl(ReleaseInfo release, String assetName) {
    final asset = release.assets.firstWhere(
      (a) => a.name == assetName,
      orElse: () => throw GitHubReleaseException(
        'Asset "$assetName" not found in release ${release.tagName}. '
        'Available assets: ${release.assets.map((a) => a.name).join(', ')}',
      ),
    );
    return asset.browserDownloadUrl;
  }

  /// Returns all asset download URLs for a release.
  ///
  /// Useful for finding checksum files or alternative downloads.
  List<String> getAllAssetUrls(ReleaseInfo release) {
    return release.assets.map((a) => a.browserDownloadUrl).toList();
  }
}

/// Information about a GitHub release.
class ReleaseInfo {
  final String tagName;
  final String name;
  final String body;
  final bool prerelease;
  final bool draft;
  final DateTime publishedAt;
  final List<ReleaseAsset> assets;

  ReleaseInfo({
    required this.tagName,
    required this.name,
    required this.body,
    required this.prerelease,
    required this.draft,
    required this.publishedAt,
    required this.assets,
  });

  factory ReleaseInfo.fromJson(Map<String, dynamic> json) {
    return ReleaseInfo(
      tagName: json['tag_name'] as String,
      name: json['name'] as String? ?? '',
      body: json['body'] as String? ?? '',
      prerelease: json['prerelease'] as bool? ?? false,
      draft: json['draft'] as bool? ?? false,
      publishedAt: DateTime.parse(json['published_at'] as String),
      assets: (json['assets'] as List<dynamic>)
          .map((a) => ReleaseAsset.fromJson(a as Map<String, dynamic>))
          .toList(),
    );
  }

  @override
  String toString() {
    return 'ReleaseInfo{tagName: $tagName, name: $name, assets: ${assets.length}}';
  }
}

/// Information about a release asset.
class ReleaseAsset {
  final String name;
  final int size;
  final String browserDownloadUrl;
  final String contentType;

  ReleaseAsset({
    required this.name,
    required this.size,
    required this.browserDownloadUrl,
    required this.contentType,
  });

  factory ReleaseAsset.fromJson(Map<String, dynamic> json) {
    return ReleaseAsset(
      name: json['name'] as String,
      size: json['size'] as int,
      browserDownloadUrl: json['browser_download_url'] as String,
      contentType:
          json['content_type'] as String? ?? 'application/octet-stream',
    );
  }

  @override
  String toString() {
    return 'ReleaseAsset{name: $name, size: $size}';
  }
}

/// Exception thrown when GitHub Release operations fail.
class GitHubReleaseException implements Exception {
  final String message;

  GitHubReleaseException(this.message);

  @override
  String toString() => 'GitHubReleaseException: $message';
}
