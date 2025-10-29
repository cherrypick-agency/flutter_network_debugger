import 'dart:convert';
import 'dart:io';

import '../../../../core/di/di.dart';
import 'package:app_http_client/application/app_http_client.dart'
    as http_client;

bool nativeAutomationAvailable() =>
    Platform.isMacOS || Platform.isWindows || Platform.isLinux;

Future<bool> isSystemProxyEnabled() async {
  try {
    if (Platform.isMacOS) {
      // Try per-service first
      final services = await Process.run('bash', [
        '-lc',
        'networksetup -listallnetworkservices | tail -n +2 | sed "s/^\\* \\?//"',
      ]);
      final list = (services.stdout ?? '')
          .toString()
          .split('\n')
          .map((e) => e.trim())
          .where((e) => e.isNotEmpty);
      for (final s in list) {
        final w = await Process.run('networksetup', ['-getwebproxy', s]);
        final h = await Process.run('networksetup', ['-getsecurewebproxy', s]);
        bool on(String t) =>
            t.toString().toLowerCase().contains('enabled: yes');
        if (on(w.stdout ?? '') || on(h.stdout ?? '')) return true;
      }
      // Fallback to scutil effective proxy
      final sc = await Process.run('scutil', ['--proxy']);
      if (sc.exitCode == 0) {
        final out = (sc.stdout ?? '').toString().toLowerCase();
        if (out.contains('httpenable : 1') || out.contains('httpsenable : 1')) {
          return true;
        }
      }
      return false;
    }
    // For non-mac platforms here we return false; separate IO files can add support.
    return false;
  } catch (_) {
    return false;
  }
}

Future<bool> autoIntegrate(String baseUrl) async {
  if (Platform.isMacOS) {
    await autoIntegrateMacOS(baseUrl);
    return await isSystemProxyEnabled();
  }
  return false;
}

Future<void> rollback(String baseUrl) async {
  if (Platform.isMacOS) {
    await rollbackMacOS(baseUrl);
  }
}

Future<void> autoIntegrateMacOS(String baseUrl) async {
  if (!Platform.isMacOS) return;
  // 1) Ensure CA exists
  try {
    final api = sl<http_client.AppHttpClient>();
    await api.post(
      path: '/_api/v1/mitm/ca/generate',
      body: {"cn": "network-debugger dev CA"},
    );
  } catch (_) {}

  // 2) Download CA
  final tmpPath = '/tmp/network-debugger-dev-ca.crt';
  try {
    final api = sl<http_client.AppHttpClient>();
    final resp = await api.get(path: '/_api/v1/mitm/ca');
    final pem =
        (resp.data is String)
            ? resp.data as String
            : utf8.decode((resp.data as List).cast<int>());
    final f = File(tmpPath);
    await f.writeAsString(pem);
  } catch (_) {}

  // 3) Trust CA (admin prompt)
  final trustCmd =
      'security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "' +
      tmpPath +
      '"';
  await Process.run('osascript', [
    '-e',
    'do shell script "' +
        trustCmd.replaceAll('"', '\\"') +
        '" with administrator privileges',
  ]);

  // 4) Enable HTTP/HTTPS proxy for every service (call networksetup per service)
  final port = _tryParsePort(baseUrl) ?? 9091;
  final servicesRes = await Process.run('bash', [
    '-lc',
    'networksetup -listallnetworkservices | tail -n +2 | sed "s/^\\* \\?//"',
  ]);
  final services =
      (servicesRes.stdout ?? '')
          .toString()
          .split('\n')
          .map((e) => e.trim())
          .where((e) => e.isNotEmpty)
          .toList();
  for (final svc in services) {
    final qSvc = svc.replaceAll('"', '\\"');
    final cmds = [
      'networksetup -setwebproxy "' + qSvc + '" 127.0.0.1 ' + port.toString(),
      'networksetup -setsecurewebproxy "' +
          qSvc +
          '" 127.0.0.1 ' +
          port.toString(),
      'networksetup -setwebproxystate "' + qSvc + '" on',
      'networksetup -setsecurewebproxystate "' + qSvc + '" on',
    ];
    for (final c in cmds) {
      await Process.run('osascript', [
        '-e',
        'do shell script "' +
            c.replaceAll('"', '\\"') +
            '" with administrator privileges',
      ]);
    }
  }

  // 5) Fallback: напрямую проставим флаги в динамическом сторе через scutil
  try {
    final sc =
        "printf 'open\\nget State:/Network/Global/Proxies\\n'"
            "+ 'd.add HTTPEnable 1\\n'"
            "+ 'd.add HTTPProxy 127.0.0.1\\n'"
            "+ 'd.add HTTPPort " +
        port.toString() +
        "\\n'"
            "+ 'd.add HTTPSEnable 1\\n'"
            "+ 'd.add HTTPSProxy 127.0.0.1\\n'"
            "+ 'd.add HTTPSPort " +
        port.toString() +
        "\\n'"
            "+ 'set State:/Network/Global/Proxies\\nquit\\n' | scutil";
    await Process.run('osascript', [
      '-e',
      'do shell script "' +
          sc.replaceAll('"', '\\"') +
          '" with administrator privileges',
    ]);
  } catch (_) {}
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
  final cn = 'network-debugger dev CA';
  if (Platform.isMacOS) {
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
    return;
  }
  if (Platform.isWindows) {
    // Remove by CN from CurrentUser Root store
    await Process.run('certutil', ['-user', '-delstore', 'Root', cn]);
  }
  if (Platform.isLinux) {
    // Remove installed CA file and refresh trust store (Debian/Ubuntu or RHEL)
    await Process.run('bash', [
      '-lc',
      'command -v pkexec >/dev/null 2>&1 && pkexec bash -lc "rm -f /usr/local/share/ca-certificates/network-debugger-dev-ca.crt && update-ca-certificates" || sudo -n bash -lc "rm -f /usr/local/share/ca-certificates/network-debugger-dev-ca.crt && update-ca-certificates" || true',
    ]);
    await Process.run('bash', [
      '-lc',
      'command -v pkexec >/dev/null 2>&1 && pkexec bash -lc "rm -f /etc/pki/ca-trust/source/anchors/network-debugger-dev-ca.crt && update-ca-trust extract" || sudo -n bash -lc "rm -f /etc/pki/ca-trust/source/anchors/network-debugger-dev-ca.crt && update-ca-trust extract" || true',
    ]);
  }
}

int? _tryParsePort(String baseUrl) {
  try {
    final u = Uri.parse(baseUrl);
    return u.hasPort ? u.port : 80;
  } catch (_) {
    return null;
  }
}

Future<String> proxyDiagnostics() async {
  try {
    if (Platform.isMacOS) {
      final buf = StringBuffer();
      final services = await Process.run('bash', [
        '-lc',
        'networksetup -listallnetworkservices | tail -n +2 | sed "s/^\\* \\?//"',
      ]);
      buf.writeln('services:\n' + (services.stdout ?? '').toString());
      final list = (services.stdout ?? '')
          .toString()
          .split('\n')
          .map((e) => e.trim())
          .where((e) => e.isNotEmpty);
      for (final s in list) {
        final w = await Process.run('networksetup', ['-getwebproxy', s]);
        final h = await Process.run('networksetup', ['-getsecurewebproxy', s]);
        buf.writeln('\n[' + s + ']');
        buf.writeln((w.stdout ?? '').toString());
        buf.writeln((h.stdout ?? '').toString());
      }
      final sc = await Process.run('scutil', ['--proxy']);
      buf.writeln('\nscutil --proxy:\n' + (sc.stdout ?? '').toString());
      return buf.toString();
    }
  } catch (e) {
    return 'diag error: ' + e.toString();
  }
  return 'N/A';
}

Future<void> openSystemProxySettings() async {
  try {
    if (Platform.isMacOS) {
      // Try to open exact Network pane (pre-Ventura style anchor)
      await Process.run('open', [
        'x-apple.systempreferences:com.apple.preference.network?Wi-Fi',
      ]);
      // Also try the new System Settings extension (Ventura/Sonoma)
      await Process.run('open', [
        'x-apple.systempreferences:com.apple.Network-Settings.extension',
      ]);
      // Open generic Network pane as fallback
      await Process.run('open', [
        'x-apple.systempreferences:com.apple.preference.network',
      ]);
      return;
    }
    if (Platform.isWindows) {
      await Process.run('cmd', ['/c', 'start', 'ms-settings:network-proxy']);
      return;
    }
    if (Platform.isLinux) {
      // Best-effort: GNOME or KDE
      await Process.run('bash', [
        '-lc',
        'gnome-control-center network 2>/dev/null || systemsettings5 2>/dev/null || true',
      ]);
    }
  } catch (_) {}
}

Future<bool> isDevCAInstalledSystem() async {
  try {
    if (Platform.isMacOS) {
      final res = await Process.run('security', [
        'find-certificate',
        '-a',
        '-c',
        'network-debugger dev CA',
        '/Library/Keychains/System.keychain',
      ]);
      final out = (res.stdout ?? '').toString();
      return res.exitCode == 0 && out.trim().isNotEmpty;
    }
    if (Platform.isWindows) {
      final res = await Process.run('certutil', [
        '-user',
        '-store',
        'Root',
        'network-debugger dev CA',
      ]);
      final out = (res.stdout ?? '').toString().toLowerCase();
      return res.exitCode == 0 &&
          out.contains('certificate') &&
          !out.contains('not found');
    }
  } catch (_) {}
  return false;
}
