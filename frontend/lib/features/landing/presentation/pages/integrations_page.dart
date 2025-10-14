import 'dart:convert';
import 'dart:io' show Platform;

import 'package:flutter/material.dart';

import '../../../../core/di/di.dart';
import '../../../../features/landing/utils/open_url.dart';
import 'integrations_platform.dart';
import 'package:app_http_client/application/app_http_client.dart'
    as http_client;

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

  @override
  void initState() {
    super.initState();
    _baseUrl = sl<http_client.AppHttpClient>().defaultHost;
    _loadStatus();
  }

  Future<void> _loadStatus() async {
    setState(() => _loading = true);
    try {
      final client = sl.get<Object>() as dynamic;
      final resp = await client.get(path: '/_api/v1/mitm/status');
      final data =
          (resp.data is Map)
              ? (resp.data as Map).cast<String, dynamic>()
              : jsonDecode(resp.data as String) as Map<String, dynamic>;
      _enabled = data['enabled'] == true;
      _hasCA = data['hasCA'] == true;
    } catch (_) {
      // ignore errors — show default values
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _generateCA() async {
    setState(() => _loading = true);
    try {
      final client = sl.get<Object>() as dynamic;
      await client.post(
        path: '/_api/v1/mitm/ca/generate',
        body: {"cn": "network-debugger dev CA"},
      );
      _hasCA = true;
      if (mounted) setState(() {});
    } catch (_) {
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _downloadCA() {
    final url = _baseUrl + '/_api/v1/mitm/ca';
    openUrl(url);
  }

  Future<void> _autoIntegrateMacOS() async {
    if (!Platform.isMacOS) return;
    setState(() => _loading = true);
    try {
      await autoIntegrateMacOS(_baseUrl);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Done: CA installed and system proxy enabled'),
          ),
        );
      }
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
    return Scaffold(
      appBar: AppBar(
        title: const Text('Integration: System Proxy and Certificate'),
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 880),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Wrap(
                    spacing: 12,
                    runSpacing: 12,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Chip(
                        label: Text(
                          _enabled ? 'MITM enabled' : 'MITM disabled',
                        ),
                        backgroundColor:
                            _enabled ? cs.primaryContainer : cs.surfaceVariant,
                      ),
                      Chip(
                        label: Text(
                          _hasCA ? 'CA installed (runtime)' : 'CA missing',
                        ),
                        backgroundColor:
                            _hasCA ? cs.secondaryContainer : cs.surfaceVariant,
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
                    ],
                  ),
                  const SizedBox(height: 16),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Step 1. Prepare root certificate (CA)',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                          const SizedBox(height: 8),
                          const Text(
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
                              IconButton(
                                onPressed: _loading ? null : _loadStatus,
                                tooltip: 'Refresh status',
                                icon: const Icon(Icons.refresh),
                              ),
                            ],
                          ),
                          const SizedBox(height: 8),
                          const Text(
                            'Important: Keep the CA private key secure. This dev CA is intended for local development only.',
                          ),
                          if (macNativeAvailable())
                            Padding(
                              padding: const EdgeInsets.only(top: 12),
                              child: ElevatedButton.icon(
                                onPressed:
                                    _loading ? null : _autoIntegrateMacOS,
                                icon: const Icon(Icons.shield_moon),
                                label: const Text(
                                  'Auto-setup (macOS): CA + system proxy',
                                ),
                              ),
                            ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Step 2. Install CA in trusted certificates (macOS)',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                          const SizedBox(height: 8),
                          const Text('Option A (GUI):'),
                          const SizedBox(height: 8),
                          const _Bullet(
                            text: 'Download CA (.crt) using the button above',
                          ),
                          const _Bullet(
                            text:
                                'Open Keychain Access → System → Certificates',
                          ),
                          const _Bullet(
                            text:
                                'Import the .crt file, then double-click the certificate → Trust → Always Trust',
                          ),
                          const SizedBox(height: 12),
                          const Text(
                            'Option B (CLI, requires administrator password):',
                          ),
                          const SizedBox(height: 8),
                          SelectableText(
                            'sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ~/Downloads/network-debugger-dev-ca.crt',
                            style: TextStyle(color: cs.onSurfaceVariant),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: const [
                          Text(
                            'Step 3. Enable system proxy (macOS)',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                          SizedBox(height: 8),
                          _Bullet(
                            text:
                                'System Settings → Network → Wi‑Fi → Details → Proxies',
                          ),
                          _Bullet(
                            text:
                                'HTTP Proxy and HTTPS Proxy: 127.0.0.1, port as in ADDR setting (default 9091)',
                          ),
                          _Bullet(
                            text: 'Save and restart applications/browsers',
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: const [
                          Text(
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
                  if (macNativeAvailable())
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              'Rollback settings (macOS)',
                              style: TextStyle(fontWeight: FontWeight.w600),
                            ),
                            const SizedBox(height: 8),
                            Wrap(
                              spacing: 8,
                              runSpacing: 8,
                              children: [
                                OutlinedButton.icon(
                                  onPressed:
                                      _loading
                                          ? null
                                          : () async {
                                            setState(() => _loading = true);
                                            try {
                                              await rollbackMacOS(_baseUrl);
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
                                            } catch (_) {
                                            } finally {
                                              if (mounted)
                                                setState(
                                                  () => _loading = false,
                                                );
                                            }
                                          },
                                  icon: const Icon(Icons.delete_forever),
                                  label: const Text('Remove dev CA'),
                                ),
                              ],
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
      children: [const Text('• '), Expanded(child: Text(text))],
    );
  }
}
