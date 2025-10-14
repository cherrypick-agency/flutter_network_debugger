import 'dart:convert';
import 'dart:io';

import '../../../../core/di/di.dart';

bool macNativeAvailable() => Platform.isMacOS;

Future<void> autoIntegrateMacOS(String baseUrl) async {
  if (!Platform.isMacOS) return;
  // 1) Ensure CA exists
  try {
    final client = sl.get<Object>() as dynamic;
    await client.post(
      path: '/_api/v1/mitm/ca/generate',
      body: {"cn": "network-debugger dev CA"},
    );
  } catch (_) {}

  // 2) Download CA
  final tmpPath = '/tmp/network-debugger-dev-ca.crt';
  try {
    final client = sl.get<Object>() as dynamic;
    final resp = await client.get(path: '/_api/v1/mitm/ca');
    final pem =
        (resp.data is String)
            ? resp.data as String
            : utf8.decode((resp.data as List).cast<int>());
    final f = File(tmpPath);
    await f.writeAsString(pem);
  } catch (_) {}

  // 3) Enable system proxy and trust CA via AppleScript (admin prompt)
  final port = _tryParsePort(baseUrl) ?? 9091;
  final dollar = String.fromCharCode(36);
  final shell =
      "security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '" +
      tmpPath +
      "'" +
      "; networksetup -listallnetworkservices | tail -n +2 | sed \"s/^\\* \\?//\" | while IFS= read -r svc; do " +
      "networksetup -setwebproxy \"" +
      dollar +
      "svc\" 127.0.0.1 " +
      port.toString() +
      "; " +
      "networksetup -setsecurewebproxy \"" +
      dollar +
      "svc\" 127.0.0.1 " +
      port.toString() +
      "; " +
      "networksetup -setwebproxystate \"" +
      dollar +
      "svc\" on; " +
      "networksetup -setsecurewebproxystate \"" +
      dollar +
      "svc\" on; " +
      "done";
  final script =
      'do shell script "' +
      shell.replaceAll('"', '\\"') +
      '" with administrator privileges';
  // The sed expression keeps service names as-is; we don't inject shell variables from Dart.
  final res = await Process.run('osascript', ['-e', script]);
  if (res.exitCode != 0) {
    final fbShell =
        "if networksetup -listallnetworkservices | grep -q \"Wi-Fi\"; then " +
        "networksetup -setwebproxy \"Wi-Fi\" 127.0.0.1 " +
        port.toString() +
        "; " +
        "networksetup -setsecurewebproxy \"Wi-Fi\" 127.0.0.1 " +
        port.toString() +
        "; " +
        "networksetup -setwebproxystate \"Wi-Fi\" on; " +
        "networksetup -setsecurewebproxystate \"Wi-Fi\" on; " +
        "fi; " +
        "security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '" +
        tmpPath +
        "'";
    final fbScript =
        'do shell script "' +
        fbShell.replaceAll('"', '\\"') +
        '" with administrator privileges';
    await Process.run('osascript', ['-e', fbScript]);
  }
}

Future<void> rollbackMacOS(String baseUrl) async {
  if (!Platform.isMacOS) return;
  final dollar = String.fromCharCode(36);
  final shell =
      "networksetup -listallnetworkservices | tail -n +2 | sed \"s/^\\* \\?//\" | while IFS= read -r svc; do " +
      "networksetup -setwebproxystate \"" +
      dollar +
      "svc\" off; " +
      "networksetup -setsecurewebproxystate \"" +
      dollar +
      "svc\" off; " +
      "done";
  final script =
      'do shell script "' +
      shell.replaceAll('"', '\\"') +
      '" with administrator privileges';
  await Process.run('osascript', ['-e', script]);
}

Future<void> deleteDevCA() async {
  if (!Platform.isMacOS) return;
  final cn = 'network-debugger dev CA';
  try {
    final res = await Process.run('security', [
      'find-certificate',
      '-a',
      '-c',
      cn,
      '-Z',
      '/Library/Keychains/System.keychain',
    ]);
    if (res.exitCode != 0) return;
    final out = (res.stdout ?? '').toString().split('\n');
    final shas = <String>[];
    for (final raw in out) {
      final line = raw.trim();
      if (line.toLowerCase().startsWith('sha-1')) {
        final parts = line.split(':');
        if (parts.length > 1) {
          final sha = parts[1].trim().replaceAll(' ', '');
          if (sha.isNotEmpty) shas.add(sha);
        }
      }
    }
    for (final sha in shas) {
      final shell =
          'security delete-certificate -Z ' +
          sha +
          ' /Library/Keychains/System.keychain';
      final script =
          'do shell script "' +
          shell.replaceAll('"', '\\"') +
          '" with administrator privileges';
      await Process.run('osascript', ['-e', script]);
    }
  } catch (_) {}
}

int? _tryParsePort(String baseUrl) {
  try {
    final u = Uri.parse(baseUrl);
    return u.hasPort ? u.port : 80;
  } catch (_) {
    return null;
  }
}
