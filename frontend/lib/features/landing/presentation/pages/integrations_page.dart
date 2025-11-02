import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart' show kIsWeb;

import '../../../../core/di/di.dart';
import 'integrations_platform.dart';
import '../../utils/os_detect.dart';
import 'file_helper.dart';
import 'package:app_http_client/application/app_http_client.dart'
    as http_client;
import '../../../../theme/context_ext.dart';
import 'package:flutter/services.dart';

// Integration screen: help with CA installation and system proxy setup (macOS)
class IntegrationsPage extends StatefulWidget {
  const IntegrationsPage({super.key});

  @override
  State<IntegrationsPage> createState() => _IntegrationsPageState();
}

class _IntegrationsPageState extends State<IntegrationsPage> {
  bool _loading = false;
  bool _hasCA = false;
  bool _enabled = false;
  String _baseUrl = '';
  bool _sysProxy = false;
  bool _osCAInstalled = false;
  String? _lastCAPath;
  String? _caFpRuntime;
  String? _caFpSystem;
  String? _caPath;
  bool _initLoaded = false;

  // Адрес прокси для подсказок/копирования
  int? _proxyPort; // forward-proxy port from backend settings
  String get _proxyAddr {
    final p = (_proxyPort ?? 9091);
    return '127.0.0.1:' + p.toString();
  }

  @override
  void initState() {
    super.initState();
    _baseUrl = sl<http_client.AppHttpClient>().defaultHost;
    _loadStatus();
  }

  Future<void> _refreshStatusSilent() async {
    try {
      final api = sl<http_client.AppHttpClient>();
      final resp = await api.get(path: '/_api/v1/mitm/status');
      final data =
          (resp.data is Map)
              ? (resp.data as Map).cast<String, dynamic>()
              : jsonDecode(resp.data as String) as Map<String, dynamic>;
      _enabled = data['enabled'] == true;
      _hasCA = data['hasCA'] == true;
      _caFpRuntime = (data['caFingerprint'] ?? '').toString();
      _caPath = (data['caCertPath'] ?? '').toString();
      _sysProxy = await isSystemProxyEnabled();
      _osCAInstalled = await isDevCAInstalledSystem();
      _caFpSystem = await systemDevCAFingerprint();
      // load forward proxy port (best-effort)
      try {
        final pc = await api.get(path: '/_api/v1/proxy/config');
        final map =
            (pc.data is Map)
                ? (pc.data as Map).cast<String, dynamic>()
                : jsonDecode(pc.data as String) as Map<String, dynamic>;
        final p = map['forward']?['port'];
        if (p is int && p > 0) {
          _proxyPort = p;
        } else {
          final addr = (map['forward']?['addr'] ?? '').toString();
          final i = addr.lastIndexOf(':');
          if (i >= 0) {
            final pp = int.tryParse(addr.substring(i + 1));
            if (pp != null && pp > 0) _proxyPort = pp;
          }
        }
      } catch (_) {}
      if (mounted) setState(() {});
    } catch (_) {}
  }

  Future<void> _pollStatus({int attempts = 8}) async {
    for (int i = 0; i < attempts; i++) {
      await Future.delayed(const Duration(seconds: 1));
      await _refreshStatusSilent();
    }
  }

  Future<void> _loadStatus() async {
    setState(() => _loading = true);
    try {
      final api = sl<http_client.AppHttpClient>();
      final resp = await api.get(path: '/_api/v1/mitm/status');
      final data =
          (resp.data is Map)
              ? (resp.data as Map).cast<String, dynamic>()
              : jsonDecode(resp.data as String) as Map<String, dynamic>;
      _enabled = data['enabled'] == true;
      _hasCA = data['hasCA'] == true;
      _caFpRuntime = (data['caFingerprint'] ?? '').toString();
      _caPath = (data['caCertPath'] ?? '').toString();
      _sysProxy = await isSystemProxyEnabled();
      _osCAInstalled = await isDevCAInstalledSystem();
      _caFpSystem = await systemDevCAFingerprint();
      // fetch proxy port
      try {
        final api2 = sl<http_client.AppHttpClient>();
        final pc = await api2.get(path: '/_api/v1/proxy/config');
        final map =
            (pc.data is Map)
                ? (pc.data as Map).cast<String, dynamic>()
                : jsonDecode(pc.data as String) as Map<String, dynamic>;
        final p = map['forward']?['port'];
        if (p is int && p > 0) {
          _proxyPort = p;
        } else {
          final addr = (map['forward']?['addr'] ?? '').toString();
          final i = addr.lastIndexOf(':');
          if (i >= 0) {
            final pp = int.tryParse(addr.substring(i + 1));
            if (pp != null && pp > 0) _proxyPort = pp;
          }
        }
      } catch (_) {}
    } catch (_) {
      // ignore errors — show default values
    } finally {
      if (mounted)
        setState(() {
          _loading = false;
          _initLoaded = true; // первый рендер после загрузки статусов
        });
    }
  }

  // deprecated: replaced by _generateAndDownloadCA

  Future<void> _generateAndDownloadCA() async {
    setState(() => _loading = true);
    try {
      final api = sl<http_client.AppHttpClient>();
      // 1) Ensure CA exists
      await api.post(
        path: '/_api/v1/mitm/ca/generate',
        body: {"cn": "network-debugger dev CA"},
      );
      // 2) Download
      final resp = await api.get(path: '/_api/v1/mitm/ca');
      final pem =
          (resp.data is String)
              ? resp.data as String
              : utf8.decode((resp.data as List).cast<int>());
      // 3) Save to Downloads using web-compatible helper
      final path = await saveFileToDownloads(
        'network-debugger-dev-ca.crt',
        pem,
      );

      _hasCA = true;
      _lastCAPath = path;
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Dev CA saved to: ' + path),
            action:
                !kIsWeb
                    ? SnackBarAction(
                      label: 'Show',
                      onPressed: () => _revealFile(path),
                    )
                    : null,
          ),
        );
      }
      await _refreshStatusSilent();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to download CA: ' + e.toString())),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _regeneratePersistCA() async {
    setState(() => _loading = true);
    try {
      final api = sl<http_client.AppHttpClient>();
      final resp = await api.post(path: '/_api/v1/mitm/ca/regenerate');
      final data =
          (resp.data is Map)
              ? (resp.data as Map).cast<String, dynamic>()
              : jsonDecode(resp.data as String) as Map<String, dynamic>;
      final certPath = (data['certPath'] ?? '').toString();
      final fp = (data['fingerprint'] ?? '').toString();
      if (certPath.isNotEmpty) _caPath = certPath;
      if (fp.isNotEmpty) _caFpRuntime = fp;
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Dev CA regenerated' +
                  (certPath.isNotEmpty ? ': ' + certPath : ''),
            ),
          ),
        );
      }
      await _refreshStatusSilent();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Regenerate failed: ' + e.toString())),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  // Функция открытия папки убрана — иконка скрыта

  String _suggestedCAPath() {
    return _lastCAPath ?? resolveDownloadsPath('network-debugger-dev-ca.crt');
  }

  Future<void> _revealFile(String path) async {
    if (kIsWeb) return; // Not available in web
    await revealInFileManager(path);
  }

  Future<void> _autoIntegrate() async {
    setState(() => _loading = true);
    try {
      final ok = await autoIntegrate(_baseUrl);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              ok
                  ? 'Done: CA installed and system proxy enabled'
                  : 'Auto-setup finished but proxy seems OFF. Please check permissions/network services and try again.',
            ),
          ),
        );
      }
      await _loadStatus();
      // Автопулинг статуса после действия
      // Fire-and-forget обновление статуса
      _pollStatus();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Auto-setup failed. Please check administrator privileges.',
            ),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  // no-op helper removed; logic lives in the platform module

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final osLabel = getOperatingSystemLabel();
    final bool caMismatch =
        _osCAInstalled &&
        _hasCA &&
        (_caFpRuntime != null && _caFpRuntime!.isNotEmpty) &&
        (_caFpSystem != null && _caFpSystem!.isNotEmpty) &&
        _caFpRuntime!.toLowerCase() != _caFpSystem!.toLowerCase();
    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            const Text('Integration: System Proxy and Certificate'),
            const SizedBox(width: 8),
            Tooltip(
              message:
                  'System proxy allows you to debug all network requests from your system and applications by routing traffic through the network debugger. This enables you to inspect HTTPS traffic, modify requests/responses, and analyze all network activity.',
              child: Icon(
                Icons.info_outline,
                size: 18,
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
              ),
            ),
          ],
        ),
        scrolledUnderElevation: 0,
      ),
      body:
          !_initLoaded
              ? const Center(
                child: Padding(
                  padding: EdgeInsets.all(24),
                  child: CircularProgressIndicator(),
                ),
              )
              : SingleChildScrollView(
                child: Center(
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 880),
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Wrap(
                            spacing: 12,
                            runSpacing: 12,
                            crossAxisAlignment: WrapCrossAlignment.center,
                            children: [
                              /*
                      // Временно скрываем чип с определённой ОС
                      Tooltip(
                        message: 'Detected operating system',
                        child: Chip(label: Text('Detected OS: ' + osLabel)),
                      ),
                      */
                              Tooltip(
                                message:
                                    'MITM endpoint enabled on the proxy server',
                                child: Chip(
                                  label: Text(
                                    _enabled ? 'MITM enabled' : 'MITM disabled',
                                  ),
                                  backgroundColor:
                                      _enabled
                                          ? context.appColors.success
                                              .withOpacity(0.25)
                                          : cs.surfaceVariant,
                                ),
                              ),
                              Tooltip(
                                message:
                                    _hasCA
                                        ? (((_caPath ?? '').isNotEmpty
                                                ? ('Path: ' + _caPath! + '\n')
                                                : '') +
                                            ((_caFpRuntime ?? '').isNotEmpty
                                                ? ('FP: ' + _caFpRuntime!)
                                                : ''))
                                        : 'Dev CA exists in proxy runtime (not about OS trust)',
                                child: Chip(
                                  label: Text(
                                    _hasCA
                                        ? 'Dev CA (proxy)'
                                        : 'CA missing (proxy)',
                                  ),
                                  backgroundColor:
                                      _hasCA
                                          ? context.appColors.success
                                              .withOpacity(0.25)
                                          : cs.surfaceVariant,
                                ),
                              ),
                              if (!kIsWeb)
                                Tooltip(
                                  message:
                                      _osCAInstalled
                                          ? ((_caFpSystem ?? '').isNotEmpty
                                              ? ('OS FP: ' + _caFpSystem!)
                                              : 'OS trust: ON')
                                          : 'System trust store (macOS System keychain / Windows Root) contains dev CA',
                                  child: Chip(
                                    label: Text(
                                      _osCAInstalled
                                          ? 'OS trust: ON'
                                          : 'OS trust: OFF',
                                    ),
                                    backgroundColor:
                                        _osCAInstalled
                                            ? context.appColors.success
                                                .withOpacity(0.25)
                                            : cs.surfaceVariant,
                                  ),
                                ),
                              if (!kIsWeb)
                                Tooltip(
                                  message: 'System HTTP/HTTPS proxy status',
                                  child: Chip(
                                    label: Text(
                                      _sysProxy ? 'Proxy: ON' : 'Proxy: OFF',
                                    ),
                                    backgroundColor:
                                        _sysProxy
                                            ? context.appColors.success
                                                .withOpacity(0.25)
                                            : cs.surfaceVariant,
                                  ),
                                ),
                              if ((_caPath ?? '').isNotEmpty)
                                Tooltip(
                                  message:
                                      'Path to persisted dev CA used by proxy',
                                  child: ActionChip(
                                    label: const Text('CA persisted'),
                                    onPressed: () async {
                                      await Clipboard.setData(
                                        ClipboardData(text: _caPath!),
                                      );
                                      if (!mounted) return;
                                      ScaffoldMessenger.of(
                                        context,
                                      ).showSnackBar(
                                        SnackBar(
                                          content: Text(
                                            'Path copied: ' + _caPath!,
                                          ),
                                        ),
                                      );
                                    },
                                  ),
                                ),
                              /*
                      // Временно убираем вспомогательные иконки (папка/копирование отпечатков CA)
                      if ((_caPath ?? '').isNotEmpty)
                        IconButton(
                          tooltip: 'Open folder with CA',
                          onPressed: () async {
                            final dir = File(_caPath!).parent.path;
                            await _openDir(dir);
                          },
                          icon: const Icon(Icons.folder_open),
                        ),
                      if ((_caFpRuntime ?? '').isNotEmpty)
                        IconButton(
                          tooltip: 'Copy proxy CA fingerprint',
                          onPressed: () async {
                            await Clipboard.setData(
                              ClipboardData(text: _caFpRuntime!),
                            );
                            if (!mounted) return;
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(
                                content: Text('Fingerprint copied'),
                              ),
                            );
                          },
                          icon: const Icon(Icons.copy),
                        ),
                      if (_osCAInstalled && (_caFpSystem ?? '').isNotEmpty)
                        IconButton(
                          tooltip: 'Copy OS CA fingerprint',
                          onPressed: () async {
                            await Clipboard.setData(
                              ClipboardData(text: _caFpSystem!),
                            );
                            if (!mounted) return;
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(
                                content: Text('Fingerprint copied'),
                              ),
                            );
                          },
                          icon: const Icon(Icons.content_copy),
                        ),
                      */
                              if (caMismatch && !kIsWeb)
                                Tooltip(
                                  message:
                                      'OS-trusted CA differs from proxy runtime CA',
                                  child: Chip(
                                    label: const Text('CA mismatch'),
                                    backgroundColor:
                                        Theme.of(
                                          context,
                                        ).colorScheme.errorContainer,
                                  ),
                                ),
                              Tooltip(
                                message: 'Copy proxy address',
                                child: ActionChip(
                                  label: Text('Addr: ' + _proxyAddr),
                                  onPressed: () async {
                                    await Clipboard.setData(
                                      ClipboardData(text: _proxyAddr),
                                    );
                                    if (!mounted) return;
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      const SnackBar(
                                        content: Text('Proxy address copied'),
                                      ),
                                    );
                                  },
                                ),
                              ),
                              if (_loading)
                                const Padding(
                                  padding: EdgeInsets.only(left: 8),
                                  child: SizedBox(
                                    width: 16,
                                    height: 16,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                    ),
                                  ),
                                ),
                              IconButton(
                                onPressed: _loading ? null : _loadStatus,
                                tooltip: 'Refresh status',
                                icon: const Icon(Icons.refresh),
                              ),
                              IconButton(
                                onPressed:
                                    _loading
                                        ? null
                                        : () async {
                                          final txt = await proxyDiagnostics();
                                          final summary =
                                              StringBuffer()
                                                ..writeln('Config summary:')
                                                ..writeln(
                                                  '  Proxy addr: ' + _proxyAddr,
                                                )
                                                ..writeln(
                                                  '  MITM enabled: ' +
                                                      _enabled.toString(),
                                                )
                                                ..writeln(
                                                  '  Dev CA (proxy): ' +
                                                      _hasCA.toString(),
                                                )
                                                ..writeln(
                                                  '  OS trust: ' +
                                                      _osCAInstalled.toString(),
                                                )
                                                ..writeln(
                                                  '  CA path: ' +
                                                      ((_caPath ?? '').isEmpty
                                                          ? 'N/A'
                                                          : _caPath!),
                                                )
                                                ..writeln(
                                                  '  Proxy CA FP: ' +
                                                      ((_caFpRuntime ?? '')
                                                              .isEmpty
                                                          ? 'N/A'
                                                          : _caFpRuntime!),
                                                )
                                                ..writeln(
                                                  '  OS CA FP: ' +
                                                      ((_caFpSystem ?? '')
                                                              .isEmpty
                                                          ? 'N/A'
                                                          : _caFpSystem!),
                                                )
                                                ..writeln(
                                                  '\n--- system output ---\n',
                                                );
                                          final full = summary.toString() + txt;
                                          if (!mounted) return;
                                          showDialog(
                                            context: context,
                                            builder:
                                                (_) => AlertDialog(
                                                  title: const Text(
                                                    'Proxy diagnostics',
                                                  ),
                                                  content: SizedBox(
                                                    width: 700,
                                                    child:
                                                        SingleChildScrollView(
                                                          child: SelectableText(
                                                            full,
                                                          ),
                                                        ),
                                                  ),
                                                  actions: [
                                                    TextButton(
                                                      onPressed: () async {
                                                        await Clipboard.setData(
                                                          ClipboardData(
                                                            text: full,
                                                          ),
                                                        );
                                                        if (!mounted) return;
                                                        ScaffoldMessenger.of(
                                                          context,
                                                        ).showSnackBar(
                                                          const SnackBar(
                                                            content: Text(
                                                              'Diagnostics copied',
                                                            ),
                                                          ),
                                                        );
                                                      },
                                                      child: const Text('Copy'),
                                                    ),
                                                    TextButton(
                                                      onPressed:
                                                          () => Navigator.pop(
                                                            context,
                                                          ),
                                                      child: const Text(
                                                        'Close',
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                          );
                                        },
                                tooltip: 'Diagnostics',
                                icon: const Icon(Icons.terminal),
                              ),
                            ],
                          ),
                          const SizedBox(height: 16),
                          if (kIsWeb)
                            Card(
                              child: Container(
                                decoration: BoxDecoration(
                                  color: cs.primary.withOpacity(0.06),
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(
                                    color: cs.primary.withOpacity(0.25),
                                  ),
                                ),
                                padding: const EdgeInsets.all(16),
                                child: Row(
                                  children: [
                                    Icon(Icons.info_outline, color: cs.primary),
                                    const SizedBox(width: 12),
                                    const Expanded(
                                      child: SelectableText(
                                        'Web version: view instructions and download CA. System diagnostics (OS trust, proxy status) and auto-setup are only available in the desktop application.',
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          if (!kIsWeb && !(_sysProxy && _hasCA))
                            Card(
                              child: Container(
                                decoration: BoxDecoration(
                                  color: cs.primary.withOpacity(0.06),
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(
                                    color: cs.primary.withOpacity(0.25),
                                  ),
                                ),
                                padding: const EdgeInsets.all(16),
                                child: Row(
                                  children: [
                                    Icon(Icons.info_outline, color: cs.primary),
                                    const SizedBox(width: 12),
                                    Expanded(
                                      child: SelectableText(
                                        'Not fully configured. Use Auto-setup or ' +
                                            'follow the steps below. Proxy address: ' +
                                            _proxyAddr,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          // Auto-setup button — only if something is missing
                          if (nativeAutomationAvailable() &&
                              (!(_sysProxy && _hasCA)))
                            Center(
                              child: Padding(
                                padding: const EdgeInsets.only(top: 12),
                                child: FilledButton.icon(
                                  onPressed: _loading ? null : _autoIntegrate,
                                  icon: const Icon(Icons.shield_moon),
                                  label: Text(
                                    'Auto-setup ($osLabel): CA + system proxy',
                                  ),
                                  style: FilledButton.styleFrom(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 16,
                                      vertical: 14,
                                    ),
                                  ),
                                ),
                              ),
                            ),
                          const SizedBox(height: 12),
                          // Rollback block right under statuses
                          if (nativeAutomationAvailable() &&
                              (_sysProxy || _hasCA))
                            Card(
                              child: Padding(
                                padding: const EdgeInsets.all(16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    const SelectableText(
                                      'Rollback settings',
                                      style: TextStyle(
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    Wrap(
                                      spacing: 8,
                                      runSpacing: 8,
                                      children: [
                                        if (_sysProxy)
                                          OutlinedButton.icon(
                                            onPressed:
                                                _loading
                                                    ? null
                                                    : () async {
                                                      setState(
                                                        () => _loading = true,
                                                      );
                                                      try {
                                                        await rollback(
                                                          _baseUrl,
                                                        );
                                                        if (mounted) {
                                                          ScaffoldMessenger.of(
                                                            context,
                                                          ).showSnackBar(
                                                            const SnackBar(
                                                              content: Text(
                                                                'Proxy disabled for system services',
                                                              ),
                                                            ),
                                                          );
                                                        }
                                                        await _loadStatus();
                                                      } catch (_) {
                                                      } finally {
                                                        if (mounted)
                                                          setState(
                                                            () =>
                                                                _loading =
                                                                    false,
                                                          );
                                                      }
                                                    },
                                            icon: const Icon(
                                              Icons.settings_backup_restore,
                                            ),
                                            label: const Text(
                                              'Disable system proxy',
                                            ),
                                          ),
                                        if (_hasCA)
                                          OutlinedButton.icon(
                                            onPressed:
                                                _loading
                                                    ? null
                                                    : () async {
                                                      setState(
                                                        () => _loading = true,
                                                      );
                                                      try {
                                                        await deleteDevCA();
                                                        if (mounted) {
                                                          ScaffoldMessenger.of(
                                                            context,
                                                          ).showSnackBar(
                                                            const SnackBar(
                                                              content: Text(
                                                                'Dev CA removed',
                                                              ),
                                                            ),
                                                          );
                                                        }
                                                        await _loadStatus();
                                                      } catch (_) {
                                                      } finally {
                                                        if (mounted)
                                                          setState(
                                                            () =>
                                                                _loading =
                                                                    false,
                                                          );
                                                      }
                                                    },
                                            style: OutlinedButton.styleFrom(
                                              foregroundColor: cs.error,
                                              side: BorderSide(color: cs.error),
                                            ),
                                            icon: const Icon(
                                              Icons.delete_forever,
                                            ),
                                            label: const Text('Remove dev CA'),
                                          ),
                                        // Fallback CLI instructions if automatic removal fails
                                        if (_osCAInstalled)
                                          Padding(
                                            padding: const EdgeInsets.only(
                                              top: 8,
                                            ),
                                            child: Column(
                                              crossAxisAlignment:
                                                  CrossAxisAlignment.start,
                                              children: [
                                                const SelectableText(
                                                  'OR remove via CLI:',
                                                ),
                                                const SizedBox(height: 6),
                                                if (detectOS() == 'mac')
                                                  const TerminalCommand(
                                                    command:
                                                        'sudo security delete-certificate -c "network-debugger dev CA" /Library/Keychains/System.keychain',
                                                  )
                                                else if (detectOS() == 'win')
                                                  const TerminalCommand(
                                                    command:
                                                        'certutil -user -delstore Root "network-debugger dev CA"',
                                                  )
                                                else ...[
                                                  const TerminalCommand(
                                                    command:
                                                        'sudo find /usr/local/share/ca-certificates -name "network-debugger-dev-ca.crt" -delete && sudo update-ca-certificates',
                                                  ),
                                                  const SizedBox(height: 6),
                                                  const TerminalCommand(
                                                    command:
                                                        'sudo find /etc/pki/ca-trust/source/anchors -name "network-debugger-dev-ca.crt" -delete && sudo update-ca-trust extract',
                                                  ),
                                                ],
                                              ],
                                            ),
                                          ),
                                      ],
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          // Notice when OS trust remains but runtime CA in proxy is missing (e.g., after server restart)
                          if (_osCAInstalled && !_hasCA && !kIsWeb)
                            Card(
                              child: Padding(
                                padding: const EdgeInsets.all(16),
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Icon(
                                      Icons.warning_amber,
                                      color: cs.tertiary,
                                    ),
                                    const SizedBox(width: 12),
                                    const Expanded(
                                      child: SelectableText(
                                        'OS trust is ON, but the proxy runtime CA is missing (likely after server restart). Use "Download CA (.crt)" to generate it again. System trust will stay intact.',
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          // Mismatch between OS-trusted CA and proxy runtime CA
                          if (_osCAInstalled &&
                              _hasCA &&
                              (_caFpRuntime != null &&
                                  _caFpRuntime!.isNotEmpty) &&
                              (_caFpSystem != null &&
                                  _caFpSystem!.isNotEmpty) &&
                              _caFpRuntime!.toLowerCase() !=
                                  _caFpSystem!.toLowerCase() &&
                              !kIsWeb)
                            Card(
                              child: Padding(
                                padding: const EdgeInsets.all(16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Icon(
                                          Icons.report_problem,
                                          color: cs.tertiary,
                                        ),
                                        const SizedBox(width: 12),
                                        Expanded(
                                          child: SelectableText(
                                            'OS trust contains a different CA than the proxy runtime CA. Reinstall the dev CA so they match, otherwise HTTPS interception will fail.',
                                          ),
                                        ),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    SelectableText(
                                      'Proxy CA (sha1): ' + _caFpRuntime!,
                                    ),
                                    if (_caFpSystem != null)
                                      SelectableText(
                                        'OS CA (sha1):    ' + _caFpSystem!,
                                      ),
                                  ],
                                ),
                              ),
                            ),
                          if (!(nativeAutomationAvailable() &&
                              _sysProxy &&
                              _hasCA))
                            Card(
                              child: Container(
                                decoration:
                                    _hasCA
                                        ? BoxDecoration(
                                          color: context.appColors.success
                                              .withOpacity(0.06),
                                          borderRadius: BorderRadius.circular(
                                            12,
                                          ),
                                          border: Border.all(
                                            color: context.appColors.success
                                                .withOpacity(0.25),
                                          ),
                                        )
                                        : null,
                                padding: const EdgeInsets.all(16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    const SelectableText(
                                      'Step 1. Prepare root certificate (CA)',
                                      style: TextStyle(
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    const SelectableText(
                                      'You can generate a temporary dev CA (for local debugging only), or use an already prepared one.',
                                    ),
                                    const SizedBox(height: 12),
                                    Wrap(
                                      spacing: 8,
                                      runSpacing: 8,
                                      children: [
                                        FilledButton.icon(
                                          onPressed:
                                              _loading
                                                  ? null
                                                  : _generateAndDownloadCA,
                                          icon: const Icon(Icons.download),
                                          label: const Text(
                                            'Download CA (.crt)',
                                          ),
                                        ),
                                        OutlinedButton.icon(
                                          onPressed:
                                              _loading
                                                  ? null
                                                  : _regeneratePersistCA,
                                          icon: const Icon(Icons.autorenew),
                                          label: const Text(
                                            'Regenerate dev CA',
                                          ),
                                        ),
                                        if (_lastCAPath != null && !kIsWeb)
                                          TextButton.icon(
                                            onPressed:
                                                _loading
                                                    ? null
                                                    : () => _revealFile(
                                                      _lastCAPath!,
                                                    ),
                                            icon: const Icon(Icons.folder_open),
                                            label: const Text('Show'),
                                          ),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    const SelectableText(
                                      'Important: Keep the CA private key secure. This dev CA is intended for local development only.',
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          const SizedBox(height: 16),
                          if (!(_sysProxy && _hasCA))
                            Card(
                              child: Container(
                                decoration:
                                    _osCAInstalled
                                        ? BoxDecoration(
                                          color: context.appColors.success
                                              .withOpacity(0.06),
                                          borderRadius: BorderRadius.circular(
                                            12,
                                          ),
                                          border: Border.all(
                                            color: context.appColors.success
                                                .withOpacity(0.25),
                                          ),
                                        )
                                        : null,
                                padding: const EdgeInsets.all(16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    const SelectableText(
                                      'Step 2. Install CA in trusted certificates',
                                      style: TextStyle(
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    const SelectableText('Option A (GUI):'),
                                    const SizedBox(height: 8),
                                    const _Bullet(
                                      text:
                                          'Download CA (.crt) using the button above',
                                    ),
                                    if (detectOS() == 'mac') ...const [
                                      _Bullet(
                                        text:
                                            'Open Keychain Access → System → Certificates',
                                      ),
                                      _Bullet(
                                        text:
                                            'Import the .crt file, then Trust → Always Trust',
                                      ),
                                    ] else if (detectOS() == 'win') ...const [
                                      _Bullet(
                                        text:
                                            'Open certmgr.msc → Trusted Root Certification Authorities → Certificates',
                                      ),
                                      _Bullet(
                                        text:
                                            'Actions → All Tasks → Import… → select the .crt and finish the wizard',
                                      ),
                                    ] else ...const [
                                      _Bullet(
                                        text:
                                            'Debian/Ubuntu: copy to /usr/local/share/ca-certificates and run update-ca-certificates (root)',
                                      ),
                                      _Bullet(
                                        text:
                                            'Fedora/RHEL: place in /etc/pki/ca-trust/source/anchors/ and run update-ca-trust extract (root)',
                                      ),
                                    ],
                                    if (_lastCAPath != null) ...[
                                      const SizedBox(height: 12),
                                      const SelectableText('Option B (CLI):'),
                                      const SizedBox(height: 8),
                                      if (detectOS() == 'mac')
                                        TerminalCommand(
                                          command:
                                              'sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ' +
                                              quoteForShell(_suggestedCAPath()),
                                        )
                                      else if (detectOS() == 'win')
                                        TerminalCommand(
                                          command:
                                              'certutil -user -addstore Root ' +
                                              quoteForShell(_suggestedCAPath()),
                                        )
                                      else ...[
                                        TerminalCommand(
                                          command:
                                              'sudo cp ' +
                                              quoteForShell(
                                                _suggestedCAPath(),
                                              ) +
                                              ' /usr/local/share/ca-certificates/',
                                        ),
                                        const SizedBox(height: 4),
                                        const TerminalCommand(
                                          command:
                                              'sudo update-ca-certificates',
                                        ),
                                      ],
                                    ],
                                  ],
                                ),
                              ),
                            ),
                          const SizedBox(height: 16),
                          if (!(_sysProxy && _hasCA))
                            Card(
                              child: Container(
                                decoration:
                                    _sysProxy
                                        ? BoxDecoration(
                                          color: context.appColors.success
                                              .withOpacity(0.06),
                                          borderRadius: BorderRadius.circular(
                                            12,
                                          ),
                                          border: Border.all(
                                            color: context.appColors.success
                                                .withOpacity(0.25),
                                          ),
                                        )
                                        : null,
                                padding: const EdgeInsets.all(16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    const SelectableText(
                                      'Step 3. Enable system proxy',
                                      style: TextStyle(
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    if (detectOS() == 'mac') ...const [
                                      _Bullet(
                                        text:
                                            'System Settings → Network → Wi‑Fi → Details → Proxies',
                                      ),
                                    ] else if (detectOS() == 'win') ...const [
                                      _Bullet(
                                        text:
                                            'Settings → Network & Internet → Proxy → Use a proxy server',
                                      ),
                                    ] else ...const [
                                      _Bullet(
                                        text:
                                            'Desktop settings → Network → Network Proxy (varies by distro/DE)',
                                      ),
                                    ],
                                    _Bullet(
                                      text: 'HTTP/HTTPS proxy: ' + _proxyAddr,
                                    ),
                                    const _Bullet(
                                      text:
                                          'Save and restart applications/browsers',
                                    ),
                                    const SizedBox(height: 8),
                                    Wrap(
                                      spacing: 8,
                                      children: [
                                        OutlinedButton.icon(
                                          onPressed:
                                              _loading
                                                  ? null
                                                  : openSystemProxySettings,
                                          icon: const Icon(Icons.open_in_new),
                                          label: const Text(
                                            'Open system proxy settings',
                                          ),
                                        ),
                                        TextButton.icon(
                                          onPressed:
                                              _loading
                                                  ? null
                                                  : () {
                                                    // Начать наблюдение за статусом, чтобы не нажимать Refresh
                                                    _pollStatus();
                                                    ScaffoldMessenger.of(
                                                      context,
                                                    ).showSnackBar(
                                                      const SnackBar(
                                                        content: Text(
                                                          'Watching for status changes…',
                                                        ),
                                                      ),
                                                    );
                                                  },
                                          icon: const Icon(Icons.visibility),
                                          label: const Text(
                                            'Auto‑refresh status',
                                          ),
                                        ),
                                      ],
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          const SizedBox(height: 16),

                          if (nativeAutomationAvailable() &&
                              _sysProxy &&
                              _hasCA &&
                              _osCAInstalled)
                            Padding(
                              padding: const EdgeInsets.only(top: 8),
                              child: Card(
                                child: Padding(
                                  padding: const EdgeInsets.all(16),
                                  child: Row(
                                    children: [
                                      Icon(
                                        Icons.check_circle,
                                        color: context.appColors.success,
                                      ),
                                      const SizedBox(width: 8),
                                      Expanded(
                                        child: Text(
                                          'Forward proxy is ready: system proxy is enabled and dev CA is active.',
                                          style:
                                              Theme.of(
                                                context,
                                              ).textTheme.bodyMedium,
                                        ),
                                      ),
                                      IconButton(
                                        onPressed:
                                            _loading ? null : _loadStatus,
                                        icon: const Icon(Icons.refresh),
                                        tooltip: 'Refresh status',
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          const SizedBox(height: 12),
                          Card(
                            child: Container(
                              decoration: BoxDecoration(
                                color: cs.primary.withOpacity(0.06),
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(
                                  color: cs.primary.withOpacity(0.25),
                                ),
                              ),
                              padding: const EdgeInsets.all(16),
                              child: Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Icon(Icons.info_outline, color: cs.primary),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        const SelectableText(
                                          'Verification',
                                          style: TextStyle(
                                            fontWeight: FontWeight.w600,
                                          ),
                                        ),
                                        const SizedBox(height: 8),
                                        const _Bullet(
                                          text:
                                              'Open any HTTPS website/client — requests will appear in the inspector',
                                        ),
                                        const _Bullet(
                                          text:
                                              'Apps with certificate pinning will not allow MITM — use dev builds without pinning',
                                        ),
                                        if (kIsWeb) ...[
                                          const SizedBox(height: 8),
                                          const _Bullet(
                                            text:
                                                'Web version note: For full diagnostics (OS trust status, system proxy detection), use the desktop application',
                                          ),
                                        ],
                                      ],
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
    );
  }
}

class _Bullet extends StatelessWidget {
  const _Bullet({required this.text});
  final String text;
  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [const Text('• '), Expanded(child: SelectableText(text))],
    );
  }
}

class TerminalCommand extends StatelessWidget {
  const TerminalCommand({super.key, required this.command});
  final String command;
  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Container(
      decoration: BoxDecoration(
        color: cs.surfaceVariant.withOpacity(0.4),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: cs.outline.withOpacity(0.6)),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      child: Row(
        children: [
          Expanded(
            child: SelectableText(command, style: context.appText.monospace),
          ),
          IconButton(
            tooltip: 'Copy',
            icon: const Icon(Icons.copy),
            onPressed: () async {
              await Clipboard.setData(ClipboardData(text: command));
              ScaffoldMessenger.of(
                context,
              ).showSnackBar(const SnackBar(content: Text('Command copied')));
            },
          ),
        ],
      ),
    );
  }
}
