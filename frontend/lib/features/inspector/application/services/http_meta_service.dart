import 'dart:convert';
import 'dart:collection';

import 'package:app_http_client/application/app_http_client.dart';

import '../../../../core/di/di.dart';
import '../stores/home_ui_store.dart';
import '../stores/sessions_store.dart';

class HttpMetaService {
  Future<void> warmup({int limit = 50, int concurrency = 6}) async {
    final store = sl<SessionsStore>();
    final ui = sl<HomeUiStore>();
    final ids = <String>[];
    for (final s in store.items.take(limit)) {
      final id = s.id;
      if (!ui.httpMeta.containsKey(id)) ids.add(id);
    }
    await _warmupIds(ids, concurrency: concurrency);
  }

  Future<void> _warmupIds(List<String> ids, {int concurrency = 6}) async {
    if (ids.isEmpty) return;
    final client = sl<AppHttpClient>();
    final ui = sl<HomeUiStore>();
    final q = Queue<String>()..addAll(ids);
    final workers = <Future<void>>[];
    final int conc = concurrency.clamp(1, 16);
    for (int i = 0; i < conc; i++) {
      workers.add(
        Future<void>(() async {
          while (true) {
            String? id;
            if (q.isNotEmpty) {
              id = q.removeFirst();
            }
            if (id == null) break;
            if (ui.httpMeta.containsKey(id)) continue;
            try {
              final r = await client.get(
                path: '/_api/v1/sessions/$id/frames',
                query: {'limit': '2'},
              );
              final meta = _extractMeta(r.data);
              if (meta.isNotEmpty) ui.httpMeta[id] = meta;
            } catch (_) {}
            await Future.delayed(const Duration(milliseconds: 8));
          }
        }),
      );
    }
    await Future.wait(workers);
  }

  Map<String, dynamic> _extractMeta(dynamic responseData) {
    final m = <String, dynamic>{};
    try {
      final data = responseData is Map<String, dynamic> ? responseData : {};
      final items =
          (data['items'] as List<dynamic>).cast<Map<String, dynamic>>();
      Map<String, dynamic>? req;
      Map<String, dynamic>? resp;
      DateTime? tReq;
      DateTime? tResp;
      for (final f in items) {
        final p = f['preview']?.toString() ?? '';
        try {
          final mp = jsonDecode(p);
          if (mp['type'] == 'http_request') {
            req = mp;
            tReq = DateTime.tryParse(f['ts']?.toString() ?? '');
          }
          if (mp['type'] == 'http_response') {
            resp = mp;
            tResp = DateTime.tryParse(f['ts']?.toString() ?? '');
          }
        } catch (_) {}
      }
      if (req != null) m['method'] = (req['method'] ?? '').toString();
      if (resp != null) {
        m['status'] = (resp['status'] ?? '').toString();
        final headers =
            (resp['headers'] as Map?)?.map(
              (k, v) => MapEntry(k.toString(), v.toString()),
            ) ??
            {};
        final ctEntry = headers.entries.firstWhere(
          (e) => e.key.toLowerCase() == 'content-type',
          orElse: () => const MapEntry('', ''),
        );
        m['mime'] = ctEntry.value;
        final upg =
            headers.entries
                .firstWhere(
                  (e) => e.key.toLowerCase() == 'upgrade',
                  orElse: () => const MapEntry('', ''),
                )
                .value;
        m['streaming'] =
            (m['mime']?.toString().contains('text/event-stream') ?? false) ||
            (upg.toString().toLowerCase() == 'websocket');
        m['headers'] = headers;
      }
      if (tReq != null && tResp != null) {
        m['durationMs'] = tResp.difference(tReq).inMilliseconds;
      }
    } catch (_) {}
    return m;
  }
}
