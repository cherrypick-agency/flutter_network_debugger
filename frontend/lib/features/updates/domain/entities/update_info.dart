/// Информация о доступном обновлении
class UpdateInfo {
  final String version;
  final String downloadUrl;
  final String changelog;
  final int sizeBytes;
  final String? sha256;
  final DateTime publishedAt;
  final bool isPrerelease;
  final String htmlUrl;

  UpdateInfo({
    required this.version,
    required this.downloadUrl,
    required this.changelog,
    this.sizeBytes = 0,
    this.sha256,
    required this.publishedAt,
    this.isPrerelease = false,
    required this.htmlUrl,
  });

  factory UpdateInfo.fromGitHubRelease(
    Map<String, dynamic> json,
    Map<String, dynamic> asset,
  ) {
    // Парсим дату публикации
    DateTime publishedAt;
    try {
      publishedAt = DateTime.parse(json['published_at'] as String);
    } catch (e) {
      publishedAt = DateTime.now();
    }

    return UpdateInfo(
      version: json['tag_name'] as String? ?? 'unknown',
      downloadUrl: asset['browser_download_url'] as String? ?? '',
      changelog: json['body'] as String? ?? '',
      sizeBytes: asset['size'] as int? ?? 0,
      // SHA256 will be computed after download
      sha256: null,
      publishedAt: publishedAt,
      isPrerelease: json['prerelease'] as bool? ?? false,
      htmlUrl: json['html_url'] as String? ?? '',
    );
  }

  String get formattedSize {
    if (sizeBytes == 0) return 'Unknown size';
    if (sizeBytes < 1024) return '$sizeBytes B';
    if (sizeBytes < 1024 * 1024) {
      return '${(sizeBytes / 1024).toStringAsFixed(1)} KB';
    }
    return '${(sizeBytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }

  String get formattedDate {
    final now = DateTime.now();
    final difference = now.difference(publishedAt);

    if (difference.inDays == 0) {
      return 'Today';
    } else if (difference.inDays == 1) {
      return 'Yesterday';
    } else if (difference.inDays < 7) {
      return '${difference.inDays} days ago';
    } else {
      // Format as "Jan 15, 2025"
      const months = [
        'Jan',
        'Feb',
        'Mar',
        'Apr',
        'May',
        'Jun',
        'Jul',
        'Aug',
        'Sep',
        'Oct',
        'Nov',
        'Dec',
      ];
      return '${months[publishedAt.month - 1]} ${publishedAt.day}, ${publishedAt.year}';
    }
  }

  bool isNewerThan(String currentVersion) {
    // Убираем 'v' префикс если есть
    final current = currentVersion.replaceFirst('v', '');
    final latest = version.replaceFirst('v', '');

    try {
      // Убираем build number (+4) если есть, оставляем только версию
      final currentVersionOnly = current.split('+').first;
      final latestVersionOnly = latest.split('+').first;

      final currentParts = currentVersionOnly
          .split('.')
          .map(int.parse)
          .toList();
      final latestParts = latestVersionOnly.split('.').map(int.parse).toList();

      // Сравниваем major.minor.patch
      for (int i = 0; i < 3; i++) {
        final currentPart = i < currentParts.length ? currentParts[i] : 0;
        final latestPart = i < latestParts.length ? latestParts[i] : 0;

        if (latestPart > currentPart) return true;
        if (latestPart < currentPart) return false;
      }

      return false; // Версии равны
    } catch (e) {
      // Если не удалось распарсить версии, считаем что обновление доступно
      return version != currentVersion;
    }
  }
}
