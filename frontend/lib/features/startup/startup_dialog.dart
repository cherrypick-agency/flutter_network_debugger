import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'port_checker.dart';

class StartupConfig {
  final int apiPort;
  final int forwardProxyPort;
  final bool serverAlreadyRunning;

  StartupConfig({
    required this.apiPort,
    required this.forwardProxyPort,
    this.serverAlreadyRunning = false,
  });

  static StartupConfig getDefaults() {
    return StartupConfig(apiPort: 9092, forwardProxyPort: 9091);
  }
}

/// Shows port configuration/check dialog at application startup
Future<StartupConfig?> showStartupDialog(
  BuildContext context, {
  StartupConfig? initialConfig,
}) async {
  final config = initialConfig ?? StartupConfig.getDefaults();

  return await showDialog<StartupConfig>(
    context: context,
    barrierDismissible: false,
    barrierColor: Colors.black.withValues(alpha: 0.8),
    builder: (ctx) => _StartupDialog(initialConfig: config),
  );
}

class _StartupDialog extends StatefulWidget {
  final StartupConfig initialConfig;

  const _StartupDialog({required this.initialConfig});

  @override
  State<_StartupDialog> createState() => _StartupDialogState();
}

class _StartupDialogState extends State<_StartupDialog> {
  late TextEditingController _apiPortCtrl;
  late TextEditingController _proxyPortCtrl;
  PortStatus? _portStatus;
  bool _checking = true;
  String _appVersion = '';
  Timer? _autoRefreshTimer;

  @override
  void initState() {
    super.initState();
    _apiPortCtrl = TextEditingController(
      text: widget.initialConfig.apiPort.toString(),
    );
    _proxyPortCtrl = TextEditingController(
      text: widget.initialConfig.forwardProxyPort.toString(),
    );
    _loadAppVersion();
    _checkPorts();
    _startAutoRefresh();
  }

  void _startAutoRefresh() {
    _autoRefreshTimer?.cancel();
    _autoRefreshTimer = Timer.periodic(const Duration(seconds: 5), (_) {
      if (mounted && !_checking) {
        _checkPorts(silent: true);
      }
    });
  }

  Future<void> _loadAppVersion() async {
    try {
      final packageInfo = await PackageInfo.fromPlatform();
      if (mounted) {
        setState(() {
          _appVersion = 'v${packageInfo.version}';
        });
      }
    } catch (e) {
      // Fallback version
      if (mounted) {
        setState(() {
          _appVersion = 'v?';
        });
      }
    }
  }

  @override
  void dispose() {
    _autoRefreshTimer?.cancel();
    _apiPortCtrl.dispose();
    _proxyPortCtrl.dispose();
    super.dispose();
  }

  Future<void> _checkPorts({bool silent = false}) async {
    if (!silent) {
      setState(() => _checking = true);
    }

    final apiPort = int.tryParse(_apiPortCtrl.text) ?? 9092;
    final proxyPort = int.tryParse(_proxyPortCtrl.text) ?? 9091;

    final status = await checkPorts(apiPort: apiPort, proxyPort: proxyPort);

    if (mounted) {
      // Show snackbar on status change (only for silent mode)
      if (silent && _portStatus != null) {
        _showStatusChangeSnackbar(_portStatus!, status);
      }

      setState(() {
        _portStatus = status;
        _checking = false;
      });

      // Automatically close dialog if both servers are already running
      if (status.bothRunning) {
        _onNext();
      }
    }
  }

  void _showStatusChangeSnackbar(PortStatus oldStatus, PortStatus newStatus) {
    if (!context.mounted) return;

    String? message;
    Color? backgroundColor;

    // Both started
    if (!oldStatus.bothRunning && newStatus.bothRunning) {
      message = 'All servers are ready!';
      backgroundColor = Colors.green;
    }
    // API server started
    else if (!oldStatus.apiRunning && newStatus.apiRunning) {
      message = 'API server is now running';
      backgroundColor = Colors.green;
    }
    // Proxy server started
    else if (!oldStatus.proxyRunning && newStatus.proxyRunning) {
      message = 'Proxy server is now running';
      backgroundColor = Colors.green;
    }
    // API server stopped
    else if (oldStatus.apiRunning && !newStatus.apiRunning) {
      message = 'API server stopped';
      backgroundColor = Colors.orange;
    }
    // Proxy server stopped
    else if (oldStatus.proxyRunning && !newStatus.proxyRunning) {
      message = 'Proxy server stopped';
      backgroundColor = Colors.orange;
    }

    if (message != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: backgroundColor,
          duration: const Duration(seconds: 2),
        ),
      );
    }
  }

  void _onStartServer() {
    final apiPort = int.tryParse(_apiPortCtrl.text) ?? 9092;
    final proxyPort = int.tryParse(_proxyPortCtrl.text) ?? 9091;

    // Port validation
    if (apiPort < 1024 || apiPort > 65535) {
      _showError('API port must be between 1024 and 65535');
      return;
    }

    if (proxyPort < 1024 || proxyPort > 65535) {
      _showError('Proxy port must be between 1024 and 65535');
      return;
    }

    if (apiPort == proxyPort) {
      _showError('Ports must be different');
      return;
    }

    Navigator.of(context).pop(
      StartupConfig(
        apiPort: apiPort,
        forwardProxyPort: proxyPort,
        serverAlreadyRunning: false,
      ),
    );
  }

  void _onNext() {
    final apiPort = int.tryParse(_apiPortCtrl.text) ?? 9092;
    final proxyPort = int.tryParse(_proxyPortCtrl.text) ?? 9091;

    Navigator.of(context).pop(
      StartupConfig(
        apiPort: apiPort,
        forwardProxyPort: proxyPort,
        serverAlreadyRunning: true,
      ),
    );
  }

  void _showError(String message) {
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(message), backgroundColor: Colors.red),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      child: AlertDialog(
        backgroundColor:
            (Theme.of(context).dialogTheme.backgroundColor ??
                    Theme.of(context).colorScheme.surface)
                .withValues(alpha: 0.95),
        title: Row(
          children: [
            Icon(
              _checking
                  ? Icons.hourglass_empty
                  : (_portStatus?.bothStopped ?? true
                        ? Icons.settings
                        : Icons.check_circle),
              size: 24,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                _checking
                    ? 'Checking Server Status'
                    : (_portStatus?.bothStopped ?? true
                          ? 'Server Configuration'
                          : 'Server Status'),
              ),
            ),
            if (_appVersion.isNotEmpty)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.grey.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  _appVersion,
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
          ],
        ),
        content: SizedBox(
          width: 400,
          child: _checking ? _buildCheckingUI() : _buildContentUI(),
        ),
        actions: _buildActions(),
      ),
    );
  }

  Widget _buildCheckingUI() {
    return const Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        CircularProgressIndicator(),
        SizedBox(height: 16),
        Text('Checking port availability...'),
      ],
    );
  }

  Widget _buildContentUI() {
    if (_portStatus == null) {
      return const Text('Failed to check ports');
    }

    if (_portStatus!.bothStopped) {
      return _buildConfigurationUI();
    } else {
      return _buildStatusUI();
    }
  }

  Widget _buildConfigurationUI() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Configure server ports before starting',
          style: TextStyle(fontSize: 14, color: Colors.grey),
        ),
        const SizedBox(height: 24),
        TextField(
          controller: _apiPortCtrl,
          keyboardType: TextInputType.number,
          inputFormatters: [FilteringTextInputFormatter.digitsOnly],
          decoration: const InputDecoration(
            labelText: 'API Port',
            hintText: '9092',
            helperText: 'Port for REST API and UI',
            prefixIcon: Icon(Icons.computer),
            border: OutlineInputBorder(),
          ),
        ),
        const SizedBox(height: 16),
        TextField(
          controller: _proxyPortCtrl,
          keyboardType: TextInputType.number,
          inputFormatters: [FilteringTextInputFormatter.digitsOnly],
          decoration: const InputDecoration(
            labelText: 'Forward Proxy Port',
            hintText: '9091',
            helperText: 'Port for HTTP/HTTPS proxy',
            prefixIcon: Icon(Icons.swap_horiz),
            border: OutlineInputBorder(),
          ),
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.blue.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.blue.withValues(alpha: 0.3)),
          ),
          child: const Row(
            children: [
              Icon(Icons.info_outline, color: Colors.blue, size: 20),
              SizedBox(width: 8),
              Expanded(
                child: Text(
                  'These settings sync with Settings page',
                  style: TextStyle(fontSize: 12, color: Colors.blue),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildStatusUI() {
    final status = _portStatus!;

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildPortStatusRow(
          label: 'API',
          port: status.apiPort,
          isRunning: status.apiRunning,
        ),
        const SizedBox(height: 12),
        _buildPortStatusRow(
          label: 'Proxy',
          port: status.proxyPort,
          isRunning: status.proxyRunning,
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: (status.bothRunning ? Colors.green : Colors.orange)
                .withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: (status.bothRunning ? Colors.green : Colors.orange)
                  .withValues(alpha: 0.3),
            ),
          ),
          child: Row(
            children: [
              Icon(
                status.bothRunning ? Icons.check_circle : Icons.warning,
                color: status.bothRunning ? Colors.green : Colors.orange,
                size: 20,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  status.bothRunning
                      ? 'Servers are ready to use'
                      : 'Start the missing server and click Refresh',
                  style: TextStyle(
                    fontSize: 12,
                    color: status.bothRunning ? Colors.green : Colors.orange,
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildPortStatusRow({
    required String label,
    required int port,
    required bool isRunning,
  }) {
    return Row(
      children: [
        Icon(
          isRunning ? Icons.check_circle : Icons.cancel,
          color: isRunning ? const Color(0xFF4CAF50) : const Color(0xFFF44336),
          size: 24,
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Text(
            isRunning
                ? '$label already running on $port'
                : 'Please start $label server',
            style: TextStyle(
              fontSize: 14,
              color: isRunning
                  ? const Color(0xFF4CAF50)
                  : const Color(0xFFF44336),
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
      ],
    );
  }

  List<Widget> _buildActions() {
    if (_checking) {
      return [];
    }

    if (_portStatus == null) {
      return [
        TextButton(
          onPressed: () => Navigator.of(context).pop(null),
          child: const Text('Cancel'),
        ),
      ];
    }

    if (_portStatus!.bothStopped) {
      // State 1: Both not running - show Refresh + Start Server
      return [
        TextButton(
          onPressed: () => Navigator.of(context).pop(null),
          child: const Text('Cancel'),
        ),
        TextButton.icon(
          onPressed: _checkPorts,
          icon: const Icon(Icons.refresh),
          label: const Text('Refresh'),
        ),
        ElevatedButton.icon(
          onPressed: _onStartServer,
          icon: const Icon(Icons.play_arrow),
          label: const Text('Start Server'),
        ),
      ];
    } else if (_portStatus!.bothRunning) {
      // State 3: Both running - show Refresh + Next
      return [
        TextButton(
          onPressed: () => Navigator.of(context).pop(null),
          child: const Text('Cancel'),
        ),
        TextButton.icon(
          onPressed: _checkPorts,
          icon: const Icon(Icons.refresh),
          label: const Text('Refresh'),
        ),
        ElevatedButton.icon(
          onPressed: _onNext,
          icon: const Icon(Icons.arrow_forward),
          label: const Text('Next'),
        ),
      ];
    } else {
      // State 2: Partially running - show Refresh
      return [
        TextButton(
          onPressed: () => Navigator.of(context).pop(null),
          child: const Text('Cancel'),
        ),
        ElevatedButton.icon(
          onPressed: _checkPorts,
          icon: const Icon(Icons.refresh),
          label: const Text('Refresh'),
        ),
      ];
    }
  }
}
