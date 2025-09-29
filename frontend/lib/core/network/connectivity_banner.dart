import 'dart:async';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:flutter/foundation.dart';

class ConnectivityBanner extends StatefulWidget {
  const ConnectivityBanner({super.key, required this.baseUrl});
  final String Function() baseUrl;

  @override
  State<ConnectivityBanner> createState() => _ConnectivityBannerState();
}

class ConnectivityState {
  static final ValueNotifier<bool> connected = ValueNotifier<bool>(true);
}

class _ConnectivityBannerState extends State<ConnectivityBanner> {
  Timer? _timer;
  bool _connected = true;
  // legacy flag (not used anymore)
  Duration _intervalConnected = const Duration(seconds: 10);
  Duration _intervalDisconnected = const Duration(seconds: 3);
  bool _checking = false;
  bool _dismissed = false;

  @override
  void initState() {
    super.initState();
    _start();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _start() {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (_) => _tick());
  }

  DateTime _lastRun = DateTime.fromMillisecondsSinceEpoch(0);
  Future<void> _tick() async {
    if (_checking) return;
    final interval = _connected ? _intervalConnected : _intervalDisconnected;
    if (DateTime.now().difference(_lastRun) < interval) return;
    _lastRun = DateTime.now();
    await _checkOnce();
  }

  Future<void> _checkOnce() async {
    _checking = true;
    try {
      final url = Uri.parse(_normalizeBase(widget.baseUrl()) + '/_api/v1/version');
      final resp = await http
          .get(url)
          .timeout(const Duration(seconds: 2));
      final ok = resp.statusCode >= 200 && resp.statusCode < 500;
      _setConnected(ok);
    } catch (_) {
      _setConnected(false);
    } finally {
      _checking = false;
    }
  }

  String _normalizeBase(String b) {
    if (b.endsWith('/')) return b.substring(0, b.length - 1);
    return b;
  }

  void _setConnected(bool v) {
    if (_connected == v) {
      if (v) {
        // при восстановлении связи снова разрешаем показывать баннер в будущем
        if (_dismissed) setState(() { _dismissed = false; });
      }
      return;
    }
    setState(() { _connected = v; if (v) _dismissed = false; });
    ConnectivityState.connected.value = v;
    // overlay больше не используем — баннер рисуется в build и сдвигает контент
  }

  @override
  Widget build(BuildContext context) {
    if (_connected || _dismissed) return const SizedBox.shrink();
    final theme = Theme.of(context);
    final bg = theme.colorScheme.errorContainer;
    final on = theme.colorScheme.onErrorContainer;
    final insets = MediaQuery.of(context).padding; // используем только top для статуса
    return Container(
      width: double.infinity,
      color: bg,
      padding: EdgeInsets.only(top: insets.top),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Icon(Icons.wifi_off, color: on, size: 16),
            const SizedBox(width: 6),
            Expanded(
              child: Text(
                'No connection to backend',
                style: theme.textTheme.bodySmall?.copyWith(color: on),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 8),
            TextButton(
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                minimumSize: const Size(40, 28),
                textStyle: theme.textTheme.labelSmall,
              ),
              onPressed: () { _checkOnce(); },
              child: Text('Retry', style: TextStyle(color: on)),
            ),
            const SizedBox(width: 6),
            TextButton(
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                minimumSize: const Size(40, 28),
                textStyle: theme.textTheme.labelSmall,
              ),
              onPressed: () { setState(() { _dismissed = true; }); },
              child: Text('Dismiss', style: TextStyle(color: on)),
            ),
          ],
        ),
      ),
    );
  }
}


