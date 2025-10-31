import 'dart:convert';

import 'package:flutter/material.dart';

import '../../domain/models.dart';
import '../../../http_inspector/presentation/widgets/http_details_panel.dart';
import '../../../../../core/di/di.dart';
import 'package:app_http_client/application/app_http_client.dart';

class ComposeHttpDetails extends StatefulWidget {
  const ComposeHttpDetails({
    super.key,
    required this.data,
    this.template,
    this.onOpenSession,
  });

  final Map<String, dynamic>? data;
  final ComposeTemplateDTO?
  template; // не используется после переключения на frames API
  final void Function(String sessionId)? onOpenSession;

  @override
  State<ComposeHttpDetails> createState() => _ComposeHttpDetailsState();
}

class _ComposeHttpDetailsState extends State<ComposeHttpDetails> {
  bool _loading = false;
  List<Map<String, dynamic>> _frames = const [];
  Map<String, dynamic>? _httpMeta;

  @override
  void initState() {
    super.initState();
    _reload();
  }

  @override
  void didUpdateWidget(covariant ComposeHttpDetails oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.data?['sessionId'] != widget.data?['sessionId']) {
      _reload();
    }
  }

  Future<void> _reload() async {
    final sid = widget.data?['sessionId']?.toString();
    final isError = widget.data?['error'] == true;
    if (!mounted) return;
    setState(() {
      _loading = true;
      _frames = const [];
      _httpMeta = null;
    });
    try {
      if (sid == null || sid.isEmpty) {
        // Ошибка транспорта: просто покажем сообщение через httpMeta
        if (isError) {
          _httpMeta = {
            'method': widget.template?.method ?? '',
            'errorMessage': (widget.data?['message'] ?? '').toString(),
          };
        }
        return;
      }
      final client = sl<AppHttpClient>();
      final res = await client.get<Map<String, dynamic>>(
        path: '/_api/v1/sessions/$sid/frames',
        query: {'limit': '200'},
      );
      final data = res.data is Map<String, dynamic> ? res.data! : {};
      final items = ((data['items'] as List?) ?? const [])
          .whereType<Map>()
          .map((e) => e.cast<String, dynamic>())
          .toList(growable: false);
      // Минимальная httpMeta: метод из http_request, статус из http_response
      String method = '';
      for (final m in items) {
        final p = m['preview']?.toString() ?? '';
        try {
          final mp = jsonDecode(p) as Map<String, dynamic>;
          final t = (mp['type'] ?? '').toString();
          if (t == 'http_request' && method.isEmpty) {
            method = (mp['method'] ?? '').toString();
          }
        } catch (_) {}
      }
      _frames = items;
      _httpMeta = {'method': method};
    } finally {
      if (mounted) {
        setState(() {
          _loading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (widget.data == null) {
      return const Center(child: Text('No response yet'));
    }
    final sessionId = widget.data!['sessionId']?.toString();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (sessionId != null && sessionId.isNotEmpty)
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 12, 12, 0),
            child: Align(
              alignment: Alignment.centerRight,
              child: TextButton.icon(
                onPressed: () => widget.onOpenSession?.call(sessionId),
                icon: const Icon(Icons.open_in_new),
                label: const Text('Open session'),
              ),
            ),
          ),
        const Divider(height: 1),
        Expanded(
          child:
              _loading
                  ? const Center(
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                  : HttpDetailsPanel(
                    sessionId: sessionId,
                    frames: _frames,
                    httpMeta: _httpMeta,
                  ),
        ),
      ],
    );
  }
}
