import 'dart:convert';
import 'dart:io' show Platform;

import 'package:flutter/material.dart';

import '../../../../core/di/di.dart';
import '../../../../features/landing/utils/open_url.dart';
import 'integrations_platform.dart';
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

  // Адрес прокси для подсказок/копирования
  String get _proxyAddr {
    try {
      final u = Uri.parse(_baseUrl);
      final p = u.hasPort ? u.port : 9091;
      return '127.0.0.1:' + p.toString();
    } catch (_) {
      return '127.0.0.1:9091';
    }
  }

  @override
  void initState() {
    super.initState();
    _baseUrl = sl<http_client.AppHttpClient>().defaultHost;
    _loadStatus();
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
      _sysProxy = await isSystemProxyEnabled();
    } catch (_) {
      // ignore errors — show default values
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _generateCA() async {
    setState(() => _loading = true);
    try {
      final api = sl<http_client.AppHttpClient>();
      await api.post(
        path: '/_api/v1/mitm/ca/generate',
        body: {"cn": "network-debugger dev CA"},
      );
      _hasCA = true;
      if (mounted) {
        setState(() {});
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(const SnackBar(content: Text('Dev CA generated.')));
      }
      // Обновим статус, чтобы подтянуть enabled/hasCA из бэка
      await _loadStatus();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to generate CA: ${e.toString()}')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _downloadCA() {
    final url = _baseUrl + '/_api/v1/mitm/ca';
    openUrl(url);
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
    final osLabel =
        Platform.isMacOS
            ? 'macOS'
            : Platform.isWindows
            ? 'Windows'
            : Platform.isLinux
            ? 'Linux'
            : 'OS';
    return Scaffold(
      appBar: AppBar(
        title: const Text('Integration: System Proxy and Certificate'),
        scrolledUnderElevation: 0,
      ),
      body: SingleChildScrollView(
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
                      Chip(label: Text('Detected OS: ' + osLabel)),
                      Chip(
                        label: Text(
                          _enabled ? 'MITM enabled' : 'MITM disabled',
                        ),
                        backgroundColor:
                            _enabled
                                ? context.appColors.success.withOpacity(0.25)
                                : cs.surfaceVariant,
                      ),
                      Chip(
                        label: Text(
                          _hasCA ? 'CA installed (runtime)' : 'CA missing',
                        ),
                        backgroundColor:
                            _hasCA
                                ? context.appColors.success.withOpacity(0.25)
                                : cs.surfaceVariant,
                      ),
                      Chip(
                        label: Text(_sysProxy ? 'Proxy: ON' : 'Proxy: OFF'),
                        backgroundColor:
                            _sysProxy
                                ? context.appColors.success.withOpacity(0.25)
                                : cs.surfaceVariant,
                      ),
                      Tooltip(
                        message: 'Копировать адрес прокси',
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
                            child: CircularProgressIndicator(strokeWidth: 2),
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
                                            child: SingleChildScrollView(
                                              child: SelectableText(txt),
                                            ),
                                          ),
                                          actions: [
                                            TextButton(
                                              onPressed: () async {
                                                await Clipboard.setData(
                                                  ClipboardData(text: txt),
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
                                                  () => Navigator.pop(context),
                                              child: const Text('Close'),
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
                  if (!(_sysProxy && _hasCA))
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Row(
                          children: [
                            Icon(
                              Icons.info_outline,
                              color: cs.onSurfaceVariant,
                            ),
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
                  // Большая заметная кнопка авто-настройки — если что-то не готово
                  if (nativeAutomationAvailable() && (!(_sysProxy && _hasCA)))
                    Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: FilledButton.icon(
                        onPressed: _loading ? null : _autoIntegrate,
                        icon: const Icon(Icons.shield_moon),
                        label: Text('Auto-setup ($osLabel): CA + system proxy'),
                        style: FilledButton.styleFrom(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 14,
                          ),
                        ),
                      ),
                    ),
                  const SizedBox(height: 12),
                  // Rollback блок сразу под статусами
                  if (nativeAutomationAvailable() && (_sysProxy || _hasCA))
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const SelectableText(
                              'Rollback settings',
                              style: TextStyle(fontWeight: FontWeight.w600),
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
                                              setState(() => _loading = true);
                                              try {
                                                await rollback(_baseUrl);
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
                                                    () => _loading = false,
                                                  );
                                              }
                                            },
                                    icon: const Icon(
                                      Icons.settings_backup_restore,
                                    ),
                                    label: const Text('Disable system proxy'),
                                  ),
                                if (_hasCA)
                                  OutlinedButton.icon(
                                    onPressed:
                                        _loading
                                            ? null
                                            : () async {
                                              setState(() => _loading = true);
                                              try {
                                                await deleteDevCA();
                                                if (mounted) {
                                                  ScaffoldMessenger.of(
                                                    context,
                                                  ).showSnackBar(
                                                    const SnackBar(
                                                      content: Text(
                                                        'Dev CA removed from System Keychain',
                                                      ),
                                                    ),
                                                  );
                                                }
                                                await _loadStatus();
                                              } catch (_) {
                                              } finally {
                                                if (mounted)
                                                  setState(
                                                    () => _loading = false,
                                                  );
                                              }
                                            },
                                    style: OutlinedButton.styleFrom(
                                      foregroundColor: cs.error,
                                      side: BorderSide(color: cs.error),
                                    ),
                                    icon: const Icon(Icons.delete_forever),
                                    label: const Text('Remove dev CA'),
                                  ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),
                  if (!(nativeAutomationAvailable() && _sysProxy && _hasCA))
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const SelectableText(
                              'Step 1. Prepare root certificate (CA)',
                              style: TextStyle(fontWeight: FontWeight.w600),
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
                                ElevatedButton.icon(
                                  onPressed: _loading ? null : _generateCA,
                                  icon: const Icon(Icons.auto_fix_high),
                                  label: const Text('Generate dev CA'),
                                ),
                                OutlinedButton.icon(
                                  onPressed: _hasCA ? _downloadCA : null,
                                  icon: const Icon(Icons.download),
                                  label: const Text('Download CA (.crt)'),
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
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const SelectableText(
                              'Step 2. Install CA in trusted certificates',
                              style: TextStyle(fontWeight: FontWeight.w600),
                            ),
                            const SizedBox(height: 8),
                            const SelectableText('Option A (GUI):'),
                            const SizedBox(height: 8),
                            const _Bullet(
                              text: 'Download CA (.crt) using the button above',
                            ),
                            if (Platform.isMacOS) ...const [
                              _Bullet(
                                text:
                                    'Open Keychain Access → System → Certificates',
                              ),
                              _Bullet(
                                text:
                                    'Import the .crt file, then Trust → Always Trust',
                              ),
                            ] else if (Platform.isWindows) ...const [
                              _Bullet(
                                text:
                                    'Open certmgr.msc → Trusted Root Certification Authorities → Certificates',
                              ),
                              _Bullet(
                                text:
                                    'Actions → All Tasks → Import… → select the .crt and finish the wizard',
                              ),
                            ] else if (Platform.isLinux) ...const [
                              _Bullet(
                                text:
                                    'Debian/Ubuntu: copy to /usr/local/share/ca-certificates and run update-ca-certificates (root)',
                              ),
                              _Bullet(
                                text:
                                    'Fedora/RHEL: place in /etc/pki/ca-trust/source/anchors/ and run update-ca-trust extract (root)',
                              ),
                            ],
                            const SizedBox(height: 12),
                            const SelectableText('Option B (CLI):'),
                            const SizedBox(height: 8),
                            if (Platform.isMacOS)
                              TerminalCommand(
                                command:
                                    'sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ~/Downloads/network-debugger-dev-ca.crt',
                              )
                            else if (Platform.isWindows)
                              TerminalCommand(
                                command:
                                    'certutil -user -addstore Root %USERPROFILE%\\Downloads\\network-debugger-dev-ca.crt',
                              )
                            else if (Platform.isLinux) ...[
                              const TerminalCommand(
                                command:
                                    'sudo cp ~/Downloads/network-debugger-dev-ca.crt /usr/local/share/ca-certificates/',
                              ),
                              const SizedBox(height: 4),
                              const TerminalCommand(
                                command: 'sudo update-ca-certificates',
                              ),
                            ],
                          ],
                        ),
                      ),
                    ),
                  const SizedBox(height: 16),
                  if (!(_sysProxy && _hasCA))
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const SelectableText(
                              'Step 3. Enable system proxy',
                              style: TextStyle(fontWeight: FontWeight.w600),
                            ),
                            const SizedBox(height: 8),
                            if (Platform.isMacOS) ...const [
                              _Bullet(
                                text:
                                    'System Settings → Network → Wi‑Fi → Details → Proxies',
                              ),
                            ] else if (Platform.isWindows) ...const [
                              _Bullet(
                                text:
                                    'Settings → Network & Internet → Proxy → Use a proxy server',
                              ),
                            ] else if (Platform.isLinux) ...const [
                              _Bullet(
                                text:
                                    'Desktop settings → Network → Network Proxy (varies by distro/DE)',
                              ),
                            ],
                            _Bullet(text: 'HTTP/HTTPS proxy: ' + _proxyAddr),
                            const _Bullet(
                              text: 'Save and restart applications/browsers',
                            ),
                            const SizedBox(height: 8),
                            Wrap(
                              spacing: 8,
                              children: [
                                OutlinedButton.icon(
                                  onPressed:
                                      _loading ? null : openSystemProxySettings,
                                  icon: const Icon(Icons.open_in_new),
                                  label: const Text(
                                    'Open system proxy settings',
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),
                  const SizedBox(height: 16),
                  if (!(_sysProxy && _hasCA))
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: const [
                            SelectableText(
                              'Step 4. Verification',
                              style: TextStyle(fontWeight: FontWeight.w600),
                            ),
                            SizedBox(height: 8),
                            _Bullet(
                              text:
                                  'Open any HTTPS website/client — requests will appear in the inspector',
                            ),
                            _Bullet(
                              text:
                                  'Apps with certificate pinning will not allow MITM — use dev builds without pinning',
                            ),
                          ],
                        ),
                      ),
                    ),
                  const SizedBox(height: 16),

                  if (nativeAutomationAvailable() && _sysProxy && _hasCA)
                    Padding(
                      padding: const EdgeInsets.only(top: 8),
                      child: Card(
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Row(
                            children: [
                              const Icon(
                                Icons.check_circle,
                                color: Colors.green,
                              ),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  'Forward proxy is ready: system proxy is enabled and dev CA is active.',
                                  style: Theme.of(context).textTheme.bodyMedium,
                                ),
                              ),
                              IconButton(
                                onPressed: _loading ? null : _loadStatus,
                                icon: const Icon(Icons.refresh),
                                tooltip: 'Refresh status',
                              ),
                            ],
                          ),
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
