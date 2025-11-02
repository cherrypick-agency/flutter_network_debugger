import 'dart:io';

Future<String> saveFileToDownloads(String fileName, String content) async {
  final home =
      Platform.environment['HOME'] ??
      Platform.environment['USERPROFILE'] ??
      '.';
  String path;
  if (Platform.isWindows) {
    final base = Platform.environment['USERPROFILE'] ?? home;
    path = base + '\\Downloads\\' + fileName;
  } else {
    path = home + '/Downloads/' + fileName;
  }

  try {
    final file = File(path);
    await file.parent.create(recursive: true);
    await file.writeAsString(content);
    return path;
  } catch (_) {
    // Fallback to temp
    final tmpBase = Directory.systemTemp.path;
    path = tmpBase + (Platform.isWindows ? '\\' : '/') + fileName;
    await File(path).writeAsString(content);
    return path;
  }
}

String getPathSeparator() {
  return Platform.isWindows ? '\\' : '/';
}

String? getEnvironmentVariable(String name) {
  return Platform.environment[name];
}

String quoteForShell(String path) {
  if (Platform.isWindows) {
    return '"' + path.replaceAll('/', '\\') + '"';
  }
  // macOS/Linux: single quotes with escaping
  return "'" + path.replaceAll("'", "'\\''") + "'";
}

String resolveDownloadsPath(String fileName) {
  final home =
      Platform.environment['HOME'] ??
      Platform.environment['USERPROFILE'] ??
      '.';
  if (Platform.isWindows) {
    final base = Platform.environment['USERPROFILE'] ?? home;
    return base + '\\Downloads\\' + fileName;
  }
  return home + '/Downloads/' + fileName;
}

String getOperatingSystemLabel() {
  if (Platform.isMacOS) return 'macOS';
  if (Platform.isWindows) return 'Windows';
  if (Platform.isLinux) return 'Linux';
  return 'OS';
}
