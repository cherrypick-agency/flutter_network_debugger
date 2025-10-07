import 'dart:convert';
import 'dart:io';

// Избавляемся от зависимости archive: используем встроенные утилиты ОС
import 'package:http/http.dart' as http;
import 'package:path/path.dart' as p;

import 'platform.dart';

Future<Directory> getCacheDir() async {
  final base = _appDataDir();
  final dir = Directory(p.join(base.path, 'bin-cache'));
  if (!await dir.exists()) await dir.create(recursive: true);
  return dir;
}

Directory _appDataDir() {
  if (Platform.isWindows) {
    final appData =
        Platform.environment['LOCALAPPDATA'] ?? Platform.environment['APPDATA'];
    return Directory(
      p.join(appData ?? Directory.current.path, 'network-debugger'),
    );
  }
  final home = Platform.environment['HOME'] ?? Directory.current.path;
  return Directory(p.join(home, '.cache', 'network-debugger'));
}

Future<String> ensureBinary({
  required PlatformSpec platform,
  required Directory cacheDir,
  required String repo,
  String? version,
  bool force = false,
  String? baseUrl,
  String? localDir,
  String? artifactName,
  bool noRemote = false,
  bool allowRemote = false,
  bool noCache = false,
}) async {
  final binName = binaryFileName(platform);
  final ext = archiveExtensionFor(platform);
  final tag = version ?? await _fetchLatestTag(repo);
  final pkgFile =
      artifactName?.isNotEmpty == true
          ? artifactName!
          : _archiveFileName(tag, platform, ext);
  final cacheSub = Directory(
    p.join(cacheDir.path, tag, '${platform.os}_${platform.arch}'),
  );
  final execPath = p.join(cacheSub.path, binName);

  if (!force && !noCache && await File(execPath).exists()) {
    return execPath;
  }

  // 1) local-dir
  if (localDir != null && localDir.trim().isNotEmpty) {
    final dir = Directory(localDir.trim());
    final fromBin = File(p.join(dir.path, binName));
    if (await fromBin.exists()) {
      // копируем в cacheSub
      await cacheSub.create(recursive: true);
      final dest = File(execPath);
      await fromBin.copy(dest.path);
      if (!Platform.isWindows) {
        await Process.run('chmod', ['+x', dest.path]);
      }
      return dest.path;
    }
    final archiveCandidate = File(p.join(dir.path, pkgFile));
    if (await archiveCandidate.exists()) {
      await cacheSub.create(recursive: true);
      await _extract(archiveCandidate, cacheSub, ext);
      final resolved = await _findExecutable(cacheSub, binName);
      if (resolved != null) {
        if (!Platform.isWindows) {
          await Process.run('chmod', ['+x', resolved]);
        }
        return resolved;
      }
      throw Exception('Binary not found inside local archive');
    }
    if (!allowRemote) {
      throw Exception('Local source not found and remote fallback is disabled');
    }
  }

  await cacheSub.create(recursive: true);
  final url =
      (baseUrl != null && baseUrl.isNotEmpty)
          ? _buildFromBase(baseUrl, pkgFile)
          : _buildDownloadUrl(repo, pkgFile);
  final tmp = File(p.join(cacheSub.path, pkgFile));
  stdout.writeln('[network-debugger] Загружаем $url');
  if (url.startsWith('file://')) {
    final path = url.substring('file://'.length);
    await File(path).copy(tmp.path);
  } else {
    await _downloadToFile(url, tmp);
  }
  stdout.writeln('[network-debugger] Распаковываем...');
  await _extract(tmp, cacheSub, ext);
  try {
    await tmp.delete();
  } catch (_) {}

  final resolved = await _findExecutable(cacheSub, binName);
  if (resolved == null) {
    throw Exception('Executable not found after extraction');
  }
  if (!Platform.isWindows) {
    await Process.run('chmod', ['+x', resolved]);
  }
  return resolved;
}

String _archiveFileName(String tag, PlatformSpec p, String ext) {
  // Имя файла соответствует артефактам из frontend/lib/features/landing/presentation/pages/download_page.dart
  // network-debugger-web_<os>_<arch>.(zip|tar.gz)
  final osPart = p.os; // windows | darwin | linux
  return 'network-debugger-web_${osPart}_${p.arch}.$ext';
}

String _buildDownloadUrl(String repo, String filename) {
  // По умолчанию используем GitHub Pages из /cmd/network-debugger-web/_web/assets/downloads/
  final base =
      Platform.environment['NETWORK_DEBUGGER_DOWNLOAD_BASE'] ??
      'https://belief.github.io/network-debugger/assets/downloads';
  return '$base/$filename';
}

String _buildFromBase(String base, String filename) {
  if (base.startsWith('file://')) {
    final root = base.substring('file://'.length);
    return 'file://' + p.join(root, filename);
  }
  final b = base.endsWith('/') ? base.substring(0, base.length - 1) : base;
  return '$b/$filename';
}

Future<String> _fetchLatestTag(String repo) async {
  // Пытаемся сначала через GitHub API, затем fallback на 'latest'
  try {
    final uri = Uri.parse('https://api.github.com/repos/$repo/releases/latest');
    final resp = await http
        .get(
          uri,
          headers: {
            'Accept': 'application/vnd.github+json',
            'User-Agent': 'network-debugger-cli',
          },
        )
        .timeout(const Duration(seconds: 8));
    if (resp.statusCode == 200) {
      final map = jsonDecode(resp.body) as Map<String, dynamic>;
      final tag = (map['tag_name'] ?? '').toString();
      if (tag.isNotEmpty) return tag;
    }
  } catch (_) {}
  return 'latest';
}

Future<void> _downloadToFile(String url, File dest) async {
  final req = await HttpClient().getUrl(Uri.parse(url));
  final resp = await req.close();
  if (resp.statusCode >= 400) {
    throw Exception('Failed to download ($url): HTTP ${resp.statusCode}');
  }
  final sink = dest.openWrite();
  await resp.listen(sink.add).asFuture();
  await sink.close();
}

Future<void> _extract(File archiveFile, Directory targetDir, String ext) async {
  if (ext == 'zip') {
    if (Platform.isWindows) {
      // PowerShell Expand-Archive доступен из коробки
      final r = await Process.run('powershell', [
        '-NoProfile',
        '-Command',
        'Expand-Archive -Path "${archiveFile.path}" -DestinationPath "${targetDir.path}" -Force',
      ]);
      if (r.exitCode != 0) {
        throw Exception('Expand-Archive failed: ${r.stderr}');
      }
      return;
    }
    // unix: используем unzip, если доступен
    final r = await Process.run('unzip', [
      '-o',
      archiveFile.path,
      '-d',
      targetDir.path,
    ]);
    if (r.exitCode != 0) {
      throw Exception('unzip failed: ${r.stderr}');
    }
    return;
  }
  // tar.gz — используем системный tar
  final r = await Process.run('tar', [
    '-xzf',
    archiveFile.path,
    '-C',
    targetDir.path,
  ]);
  if (r.exitCode != 0) {
    throw Exception('tar extract failed: ${r.stderr}');
  }
}

Future<String?> _findExecutable(Directory root, String binName) async {
  final direct = File(p.join(root.path, binName));
  if (await direct.exists()) return direct.path;
  try {
    await for (final e in root.list(recursive: true, followLinks: false)) {
      if (e is File && p.basename(e.path) == binName) return e.path;
    }
  } catch (_) {}
  return null;
}
