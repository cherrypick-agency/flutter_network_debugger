import 'dart:js_interop' as js;
import 'package:web/web.dart' as web;
import '../../utils/os_detect.dart';

Future<String> saveFileToDownloads(String fileName, String content) async {
  // Web: trigger browser download using Blob
  final blob = web.Blob([content.toJS].toJS);
  final url = web.URL.createObjectURL(blob);
  final anchor =
      web.document.createElement('a') as web.HTMLAnchorElement
        ..href = url
        ..download = fileName
        ..style.display = 'none';
  web.document.body!.appendChild(anchor);
  anchor.click();
  web.document.body!.removeChild(anchor);
  web.URL.revokeObjectURL(url);
  // Return path with environment variable for display in CLI commands
  final os = detectOS();
  if (os == 'win') {
    return '%USERPROFILE%\\Downloads\\' + fileName;
  }
  return '\$HOME/Downloads/' + fileName;
}

String getPathSeparator() {
  final os = detectOS();
  return os == 'win' ? '\\' : '/';
}

String? getEnvironmentVariable(String name) {
  return null; // Not available in web
}

String quoteForShell(String path) {
  final os = detectOS();
  if (os == 'win') {
    return '"' + path.replaceAll('/', '\\') + '"';
  }
  // macOS/Linux: use double quotes for paths with env variables
  if (path.contains('\$')) {
    return '"' + path.replaceAll('"', '\\"') + '"';
  }
  // Otherwise use single quotes with escaping
  return "'" + path.replaceAll("'", "'\\''") + "'";
}

String resolveDownloadsPath(String fileName) {
  final os = detectOS();
  if (os == 'win') {
    return '%USERPROFILE%\\Downloads\\' + fileName;
  }
  return '\$HOME/Downloads/' + fileName;
}

String getOperatingSystemLabel() {
  final os = detectOS();
  if (os == 'mac') return 'macOS';
  if (os == 'win') return 'Windows';
  return 'Linux';
}
