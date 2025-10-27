import 'package:test/test.dart';
import 'package:network_debugger/src/github_release.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'dart:convert';

void main() {
  group('GitHubRelease', () {
    test('getLatestRelease parses response correctly', () async {
      final mockClient = MockClient((request) async {
        expect(request.url.toString(), contains('/releases/latest'));
        return http.Response(
          jsonEncode({
            'tag_name': 'v1.0.0',
            'name': 'Release 1.0.0',
            'body': 'Release notes',
            'prerelease': false,
            'draft': false,
            'published_at': '2025-01-01T00:00:00Z',
            'assets': [
              {
                'name': 'binary.tar.gz',
                'size': 1024,
                'browser_download_url': 'https://example.com/binary.tar.gz',
                'content_type': 'application/gzip',
              }
            ],
          }),
          200,
        );
      });

      final client = GitHubRelease(
        owner: 'test',
        repo: 'repo',
        client: mockClient,
      );

      final release = await client.getLatestRelease();

      expect(release.tagName, equals('v1.0.0'));
      expect(release.name, equals('Release 1.0.0'));
      expect(release.assets.length, equals(1));
      expect(release.assets[0].name, equals('binary.tar.gz'));
    });

    test('getLatestRelease throws on error', () async {
      final mockClient = MockClient((request) async {
        return http.Response('Not Found', 404);
      });

      final client = GitHubRelease(
        owner: 'test',
        repo: 'repo',
        client: mockClient,
      );

      expect(
        () => client.getLatestRelease(),
        throwsA(isA<GitHubReleaseException>()),
      );
    });

    test('getRelease fetches specific version', () async {
      final mockClient = MockClient((request) async {
        expect(request.url.toString(), contains('/tags/v1.2.3'));
        return http.Response(
          jsonEncode({
            'tag_name': 'v1.2.3',
            'name': 'Release 1.2.3',
            'body': '',
            'prerelease': false,
            'draft': false,
            'published_at': '2025-01-01T00:00:00Z',
            'assets': [],
          }),
          200,
        );
      });

      final client = GitHubRelease(
        owner: 'test',
        repo: 'repo',
        client: mockClient,
      );

      final release = await client.getRelease('v1.2.3');
      expect(release.tagName, equals('v1.2.3'));
    });

    test('findAssetUrl returns correct URL', () {
      final release = ReleaseInfo(
        tagName: 'v1.0.0',
        name: 'Test',
        body: '',
        prerelease: false,
        draft: false,
        publishedAt: DateTime.now(),
        assets: [
          ReleaseAsset(
            name: 'binary_darwin_arm64.tar.gz',
            size: 1024,
            browserDownloadUrl: 'https://example.com/binary.tar.gz',
            contentType: 'application/gzip',
          ),
        ],
      );

      final client = GitHubRelease(owner: 'test', repo: 'repo');
      final url = client.findAssetUrl(release, 'binary_darwin_arm64.tar.gz');
      expect(url, equals('https://example.com/binary.tar.gz'));
    });

    test('findAssetUrl throws when asset not found', () {
      final release = ReleaseInfo(
        tagName: 'v1.0.0',
        name: 'Test',
        body: '',
        prerelease: false,
        draft: false,
        publishedAt: DateTime.now(),
        assets: [],
      );

      final client = GitHubRelease(owner: 'test', repo: 'repo');
      expect(
        () => client.findAssetUrl(release, 'nonexistent.tar.gz'),
        throwsA(isA<GitHubReleaseException>()),
      );
    });
  });
}
