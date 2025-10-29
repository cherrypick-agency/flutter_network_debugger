import 'dart:convert';
import 'dart:io';

import '../../../../core/di/di.dart';
import 'package:app_http_client/application/app_http_client.dart' as http_client;

bool nativeAutomationAvailable() => Platform.isMacOS || Platform.isWindows || Platform.isLinux;

Future<bool> isSystemProxyEnabled() async {
  try {
    if (Platform.isMacOS) {
      final servicesRes = await Process.run('bash', [
        '-lc',
        'networksetup -listallnetworkservices | tail -n +2 | sed "s/^\\* \\?//"'
      ]);
      if (servicesRes.exitCode != 0) return false;
      final services = (servicesRes.stdout as String)
          .split('\n')
          .map((e) => e.trim())
          .where((e) => e.isNotEmpty)
          .toList();
      for (final svc in services) {
        final w = await Process.run('networksetup', ['-getwebproxy', svc]);
        final s = await Process.run('networksetup', ['-getsecurewebproxy', svc]);
        final ws = (w.stdout ?? '').toString().toLowerCase();
        final ss = (s.stdout ?? '').toString().toLowerCase();
        bool on(String out) => out.contains('enabled: yes');
        if (on(ws) || on(ss)) return true;
      }
      // Fallback: read effective system proxy
      final sc = await Process.run('scutil', ['--proxy']);
      if (sc.exitCode == 0) {
        final out = (sc.stdout ?? '').toString().toLowerCase();
        if (out.contains('httpenable : 1') || out.contains('httpsenable : 1')) {
          return true;
        }
      }
      return false;
    }
    if (Platform.isWindows) {
      final res = await Process.run(
        'reg',
        [
          'query',
          'HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings',
          '/v',
          'ProxyEnable'
        ],
      );
      if (res.exitCode == 0) {
        final out = (res.stdout ?? '').toString().toLowerCase();
        if (out.contains('0x1')) return true;
      }
      // Also check WinHTTP proxy
      final nh = await Process.run('netsh', ['winhttp', 'show', 'proxy']);
      if (nh.exitCode == 0) {
        final s = (nh.stdout ?? '').toString().toLowerCase();
        // When no proxy: "Direct access (no proxy server)"
        if (!s.contains('direct access')) return true;
      }
      return false;
    }
    if (Platform.isLinux) {
      // Try GNOME proxy mode
      final g = await Process.run('bash', [
        '-lc',
        'gsettings get org.gnome.system.proxy mode 2>/dev/null || echo none'
      ]);
      final mode = (g.stdout ?? '').toString().trim();
      if (mode.contains('manual')) return true;
      // Fallback: env var
      final e = await Process.run('bash', ['-lc', r'echo $http_proxy$https_proxy']);
      return ((e.stdout ?? '') as String).trim().isNotEmpty;
    }
  } catch (_) {
    return false;
  }
  return false;
}

Future<bool> autoIntegrate(String baseUrl) async {
  if (Platform.isMacOS) {
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
    // Verify
    return await isSystemProxyEnabled();
  }

  if (Platform.isWindows) {
    try {
      // Ensure CA exists
      final api = sl<http_client.AppHttpClient>();
      await api.post(
        path: '/_api/v1/mitm/ca/generate',
        body: {"cn": "network-debugger dev CA"},
      );
    } catch (_) {}
    // Download CA to temp
    final tmpPath = Directory.systemTemp.path +
        Platform.pathSeparator +
        'network-debugger-dev-ca.crt';
    try {
      final api = sl<http_client.AppHttpClient>();
      final resp = await api.get(path: '/_api/v1/mitm/ca');
      final pem =
          (resp.data is String)
              ? resp.data as String
              : utf8.decode((resp.data as List).cast<int>());
      await File(tmpPath).writeAsString(pem);
    } catch (_) {}
    // Import into CurrentUser Root store (no admin required)
    await Process.run('certutil', ['-user', '-addstore', 'Root', tmpPath]);
    // Enable user proxy settings
    final port = _tryParsePort(baseUrl) ?? 9091;
    await Process.run('reg', [
      'add',
      'HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings',
      '/v',
      'ProxyEnable',
      '/t',
      'REG_DWORD',
      '/d',
      '1',
      '/f'
    ]);
    await Process.run('reg', [
      'add',
      'HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings',
      '/v',
      'ProxyServer',
      '/t',
      'REG_SZ',
      '/d',
      '127.0.0.1:' + port.toString(),
      '/f'
    ]);
    await Process.run('reg', [
      'add',
      'HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings',
      '/v',
      'ProxyOverride',
      '/t',
      'REG_SZ',
      '/d',
      '<local>',
      '/f'
    ]);
    // Set WinHTTP proxy (requires elevation). Try direct, then elevate via PowerShell if needed.
    final setWinHttp = await Process.run(
      'netsh',
      ['winhttp', 'set', 'proxy', '127.0.0.1:' + port.toString(), 'bypass-list=localhost'],
    );
    if (setWinHttp.exitCode != 0) {
      final ps = 'Start-Process -Verb RunAs cmd -ArgumentList "/c netsh winhttp set proxy 127.0.0.1:' +
          port.toString() +
          ' bypass-list=localhost"';
      await Process.run('powershell', ['-NoProfile', '-Command', ps]);
    }
    return await isSystemProxyEnabled();
  }

  if (Platform.isLinux) {
    // 1) Ensure CA exists and save PEM to temp
    final tmpPath = Directory.systemTemp.path +
        Platform.pathSeparator +
        'network-debugger-dev-ca.crt';
    try {
      final api = sl<http_client.AppHttpClient>();
      await api.post(
        path: '/_api/v1/mitm/ca/generate',
        body: {"cn": "network-debugger dev CA"},
      );
      final resp = await api.get(path: '/_api/v1/mitm/ca');
      final pem =
          (resp.data is String)
              ? resp.data as String
              : utf8.decode((resp.data as List).cast<int>());
      await File(tmpPath).writeAsString(pem);
    } catch (_) {}

    // 2) Install to system trust store (Debian/Ubuntu or RHEL family)
    final debPath = '/usr/local/share/ca-certificates/network-debugger-dev-ca.crt';
    final rhelPath = '/etc/pki/ca-trust/source/anchors/network-debugger-dev-ca.crt';
    // Try Debian/Ubuntu path first; otherwise RHEL/Fedora
    final installDeb = 'cp "' + tmpPath + '" "' + debPath + '" && update-ca-certificates';
    final installRhel = 'cp "' + tmpPath + '" "' + rhelPath + '" && update-ca-trust extract';
    // Use pkexec to elevate; ignore errors if pkexec is missing
    await Process.run('bash', [
      '-lc',
      'command -v pkexec >/dev/null 2>&1 && pkexec bash -lc "' + installDeb.replaceAll('"', '\\"') + '" || sudo -n bash -lc "' + installDeb.replaceAll('"', '\\"') + '" || true'
    ]);
    await Process.run('bash', [
      '-lc',
      'test -f "' + debPath + '" || (command -v pkexec >/dev/null 2>&1 && pkexec bash -lc "' + installRhel.replaceAll('"', '\\"') + '" || sudo -n bash -lc "' + installRhel.replaceAll('"', '\\"') + '") || true'
    ]);

    // 3) Enable proxy: prefer GNOME gsettings; otherwise drop env file
    final port = _tryParsePort(baseUrl) ?? 9091;
    final setGnome =
        'gsettings set org.gnome.system.proxy mode manual && ' +
        'gsettings set org.gnome.system.proxy.http host 127.0.0.1 && ' +
        'gsettings set org.gnome.system.proxy.http port ' + port.toString() + ' && ' +
        'gsettings set org.gnome.system.proxy.https host 127.0.0.1 && ' +
        'gsettings set org.gnome.system.proxy.https port ' + port.toString();
    final gRes = await Process.run('bash', ['-lc', setGnome]);
    if (gRes.exitCode != 0) {
      final home = Platform.environment['HOME'] ?? '';
      if (home.isNotEmpty) {
        final envDir = Directory(home + '/.config/environment.d');
        try {
          if (!envDir.existsSync()) envDir.createSync(recursive: true);
          final file = File(envDir.path + '/90-network-debugger-proxy.conf');
          await file.writeAsString(
            'http_proxy=http://127.0.0.1:' + port.toString() + '\n' +
            'https_proxy=http://127.0.0.1:' + port.toString() + '\n',
          );
        } catch (_) {}
      }
    }
    return await isSystemProxyEnabled();
  }
  return false;
}

Future<void> rollback(String baseUrl) async {
  if (Platform.isMacOS) {
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
    return;
  }
  if (Platform.isWindows) {
    await Process.run('reg', [
      'add',
      'HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings',
      '/v',
      'ProxyEnable',
      '/t',
      'REG_DWORD',
      '/d',
      '0',
      '/f'
    ]);
    // Reset WinHTTP proxy (requires elevation)
    final reset = await Process.run('netsh', ['winhttp', 'reset', 'proxy']);
    if (reset.exitCode != 0) {
      final ps =
          'Start-Process -Verb RunAs cmd -ArgumentList "/c netsh winhttp reset proxy"';
      await Process.run('powershell', ['-NoProfile', '-Command', ps]);
    }
  }
  if (Platform.isLinux) {
    // Disable GNOME proxy if available
    await Process.run('bash', [
      '-lc',
      'gsettings set org.gnome.system.proxy mode none 2>/dev/null || true'
    ]);
    // Remove env file fallback
    final home = Platform.environment['HOME'] ?? '';
    if (home.isNotEmpty) {
      final f = File(home + '/.config/environment.d/90-network-debugger-proxy.conf');
      try {
        if (f.existsSync()) f.deleteSync();
      } catch (_) {}
    }
  }
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
      'command -v pkexec >/dev/null 2>&1 && pkexec bash -lc "rm -f /usr/local/share/ca-certificates/network-debugger-dev-ca.crt && update-ca-certificates" || sudo -n bash -lc "rm -f /usr/local/share/ca-certificates/network-debugger-dev-ca.crt && update-ca-certificates" || true'
    ]);
    await Process.run('bash', [
      '-lc',
      'command -v pkexec >/dev/null 2>&1 && pkexec bash -lc "rm -f /etc/pki/ca-trust/source/anchors/network-debugger-dev-ca.crt && update-ca-trust extract" || sudo -n bash -lc "rm -f /etc/pki/ca-trust/source/anchors/network-debugger-dev-ca.crt && update-ca-trust extract" || true'
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
