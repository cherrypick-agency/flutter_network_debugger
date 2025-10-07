import 'dart:convert';

import 'package:app_http_client/application/app_http_client.dart';

import '../../../../core/di/di.dart';
import '../stores/home_ui_store.dart';
import '../stores/sessions_store.dart';

class HttpMetaService {
  Future<void> warmup({int limit = 50}) async {
    final list = sl<SessionsStore>().items.take(limit).toList();
    final client = sl<AppHttpClient>();
    final ui = sl<HomeUiStore>();
    for (final s in list) {
      final id = s.id;
      if (ui.httpMeta.containsKey(id)) continue;
      try {
        final r = await client.get(
          path: '/api/sessions/$id/frames',
          query: {'limit': '2'},
        );
        final m = <String, dynamic>{};
        final data = r.data is Map<String, dynamic> ? r.data : {};
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
        if (m.isNotEmpty) {
          ui.httpMeta[id] = m;
        }
      } catch (_) {}
    }
  }
}
