import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../core/network/error_utils.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_viewer.dart';
import '../../../../core/di/di.dart';

class HttpDetailsPanel extends StatefulWidget {
  const HttpDetailsPanel({
    super.key,
    required this.sessionId,
    required this.frames,
    this.httpMeta,
  });
  final String? sessionId;
  final List<dynamic> frames;
  final Map<String, dynamic>? httpMeta;

  @override
  State<HttpDetailsPanel> createState() => _HttpDetailsPanelState();
}

class _HttpDetailsPanelState extends State<HttpDetailsPanel> {
  bool _prettyReq = true;
  bool _prettyResp = true;
  bool _treeReq = false;
  bool _treeResp = false;
  bool _loadingFetch = false;
  int? _respTtfbMs;
  int? _respTotalMs;
  String? _bodyOverride;

  @override
  Widget build(BuildContext context) {
    final req = _findByType(widget.frames, 'http_request');
    final resp = _findByType(widget.frames, 'http_response');
    final reqTs = _tsOf(widget.frames, 'http_request');
    final respTs = _tsOf(widget.frames, 'http_response');
    final durationMs =
        (reqTs != null && respTs != null)
            ? respTs.difference(reqTs).inMilliseconds
            : null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 8,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            SizedBox(width: 4),
            Text('HTTP Details', style: context.appText.title),
            if (durationMs != null) _chip(context, '$durationMs ms'),
          ],
        ),
        const SizedBox(height: 8),
        Expanded(
          child: Row(
            children: [
              Expanded(
                child: _Card(
                  title: 'Request',
                  child: _buildRequest(context, req),
                ),
              ),
              const VerticalDivider(width: 1),
              Expanded(
                child: _Card(
                  title: 'Response',
                  child: _buildResponse(context, resp),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildRequest(BuildContext context, Map<String, dynamic>? req) {
    if (req == null) return Text('Нет данных', style: context.appText.body);
    final headers =
        (req['headers'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};
    final headersRaw =
        (req['headersRaw'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};
    final body = _normalizeMaybeQuotedJson((req['body'] ?? '').toString());
    final url = (req['url'] ?? '').toString();
    final uri = _tryParseUri(url);
    final qp = uri?.queryParametersAll ?? <String, List<String>>{};
    // cookies: prefer raw (unmasked) if available
    final cookieHeader =
        headersRaw.entries
                .firstWhere(
                  (e) => e.key.toLowerCase() == 'cookie',
                  orElse: () => const MapEntry('', ''),
                )
                .value
                .isNotEmpty
            ? headersRaw.entries
                .firstWhere(
                  (e) => e.key.toLowerCase() == 'cookie',
                  orElse: () => const MapEntry('', ''),
                )
                .value
            : headers.entries
                .firstWhere(
                  (e) => e.key.toLowerCase() == 'cookie',
                  orElse: () => const MapEntry('', ''),
                )
                .value;
    final reqCookies = _parseRequestCookies(cookieHeader);
    return ListView(
      children: [
        // Полная строка URL в отдельной строке со скроллом
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: SelectableText(
            '${(req['method'] ?? '').toString().toUpperCase()}  $url',
            style: context.appText.subtitle,
          ),
        ),
        const SizedBox(height: 6),
        // Кнопки на новой строке, чтобы не мешали URL
        Wrap(
          spacing: 8,
          children: [
            TextButton.icon(
              onPressed: () {
                Clipboard.setData(ClipboardData(text: url));
              },
              icon: const Icon(Icons.link, size: 16),
              label: const Text('Copy URL'),
            ),
            TextButton.icon(
              onPressed: () {
                final curl = _buildCurl(req);
                Clipboard.setData(ClipboardData(text: curl));
              },
              icon: const Icon(Icons.content_paste, size: 16),
              label: const Text('Copy as cURL'),
            ),
          ],
        ),
        const SizedBox(height: 8),
        if (qp.isNotEmpty) ...[
          Text('Query Params', style: context.appText.subtitle),
          const SizedBox(height: 4),
          ...qp.entries.map(
            (e) => Padding(
              padding: const EdgeInsets.only(bottom: 2),
              child: SelectableText(
                '${e.key}: ${e.value.join(', ')}',
                style: context.appText.monospace,
              ),
            ),
          ),
          const SizedBox(height: 8),
        ],
        if (reqCookies.isNotEmpty) ...[
          Text('Cookies', style: context.appText.subtitle),
          const SizedBox(height: 4),
          ...reqCookies.entries.map(
            (e) => Padding(
              padding: const EdgeInsets.only(bottom: 2),
              child: SelectableText(
                '${e.key}: ${e.value}',
                style: context.appText.monospace,
              ),
            ),
          ),
          const SizedBox(height: 8),
        ],
        if (body.isNotEmpty)
          Row(
            children: [
              Text('Body', style: context.appText.subtitle),
              const SizedBox(width: 12),
              FilterChip(
                label: const Text('Pretty'),
                selected: _prettyReq && !_treeReq,
                onSelected: (v) {
                  setState(() {
                    _treeReq = false;
                    _prettyReq = true;
                  });
                },
              ),
              const SizedBox(width: 8),
              if (_isJson(body))
                FilterChip(
                  label: const Text('Tree'),
                  selected: _treeReq,
                  onSelected: (v) {
                    setState(() {
                      _treeReq = v;
                      _prettyReq = !v;
                    });
                  },
                ),
              const SizedBox(width: 8),
              TextButton.icon(
                onPressed: () {
                  Clipboard.setData(ClipboardData(text: body));
                },
                icon: const Icon(Icons.copy, size: 16),
                label: const Text('Copy'),
              ),
            ],
          ),
        if (body.isNotEmpty) const SizedBox(height: 6),
        if (body.isNotEmpty)
          (_isJson(body) && _treeReq)
              ? JsonViewer(jsonString: body, forceTree: true)
              : (_isJson(body) && _prettyReq)
              ? JsonViewer(jsonString: body, forceTree: false)
              : SelectableText(body, style: context.appText.monospace),
        const SizedBox(height: 8),
        Text('Headers', style: context.appText.subtitle),
        const SizedBox(height: 4),
        ...headers.entries
            .map(
              (e) => Padding(
                padding: const EdgeInsets.only(bottom: 2),
                child: _HeaderItem(
                  name: e.key,
                  value: e.value,
                  raw: headersRaw[e.key],
                ),
              ),
            )
            .toList(),
      ],
    );
  }

  Widget _buildResponse(BuildContext context, Map<String, dynamic>? resp) {
    if (resp == null) return Text('Нет данных', style: context.appText.body);
    final headers =
        (resp['headers'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};
    final headersRaw =
        (resp['headersRaw'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};
    final rawBody = (resp['body'] ?? '').toString();
    final ctHeader =
        headers.entries
            .firstWhere(
              (e) => e.key.toLowerCase() == 'content-type',
              orElse: () => const MapEntry('', ''),
            )
            .value
            .toLowerCase();
    final isJsonCt = ctHeader.contains('json') || ctHeader.contains('+json');
    final body = _normalizeMaybeQuotedJson(
      (_bodyOverride ?? rawBody).toString(),
    );
    final req = _findByType(widget.frames, 'http_request');
    final url = req?['url']?.toString() ?? '';
    final reqHeaders =
        (req?['headers'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};
    final reqTs = _tsOf(widget.frames, 'http_request');
    final respTs = _tsOf(widget.frames, 'http_response');
    final durationMs =
        (reqTs != null && respTs != null)
            ? respTs.difference(reqTs).inMilliseconds
            : null;
    final status = (resp['status'] ?? 0) as int;
    final color = _statusColor(context, status);
    // Cache & CORS quick analysis
    final cache = _computeCacheMeta(status, headers);
    final cors = _computeCorsMeta(
      (req?['method'] ?? '').toString(),
      reqHeaders,
      headers,
    );
    final errMessage = widget.httpMeta?['errorMessage']?.toString();
    // Автоподтяжка полного тела, если по заголовку это JSON, но превью не парсится
    if (_bodyOverride == null &&
        isJsonCt &&
        body.isNotEmpty &&
        !_isJson(body) &&
        !_loadingFetch &&
        url.isNotEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _refetchInline(url));
    }

    return ListView(
      children: [
        Text(
          'Status: $status',
          style: context.appText.subtitle.copyWith(color: color),
        ),
        if (status == 0 && (errMessage != null && errMessage.isNotEmpty))
          Padding(
            padding: const EdgeInsets.only(top: 6),
            child: Text(
              'Transport Error: $errMessage',
              style: Theme.of(context).textTheme.labelSmall?.copyWith(
                color: Theme.of(context).colorScheme.error,
              ),
            ),
          ),
        if (durationMs != null || _respTtfbMs != null || _respTotalMs != null)
          Padding(
            padding: const EdgeInsets.only(top: 6, bottom: 6),
            child: Wrap(
              spacing: 8,
              children: [
                if (_respTotalMs != null)
                  _chip(context, 'Total: ${_respTotalMs} ms')
                else if (durationMs != null)
                  _chip(context, 'Total: ${durationMs} ms'),
                if (_respTtfbMs != null)
                  _chip(context, 'TTFB: ${_respTtfbMs} ms'),
                if (_respTotalMs != null &&
                    _respTtfbMs != null &&
                    (_respTotalMs! - _respTtfbMs!) >= 0)
                  _chip(
                    context,
                    'Download: ${_respTotalMs! - _respTtfbMs!} ms',
                  ),
              ],
            ),
          ),
        const SizedBox(height: 8),
        if (body.isNotEmpty)
          Row(
            children: [
              Text('Body', style: context.appText.subtitle),
              const SizedBox(width: 12),
              FilterChip(
                label: const Text('Pretty'),
                selected: _prettyResp && !_treeResp,
                onSelected: (v) {
                  setState(() {
                    _treeResp = false;
                    _prettyResp = true;
                  });
                },
              ),
              const SizedBox(width: 8),
              if (_isJson(body))
                FilterChip(
                  label: const Text('JSON Viewer'),
                  selected: _treeResp,
                  onSelected: (v) {
                    setState(() {
                      _treeResp = v;
                      _prettyResp = !v;
                    });
                  },
                ),
              const SizedBox(width: 8),
              TextButton.icon(
                onPressed: () {
                  Clipboard.setData(ClipboardData(text: body));
                },
                icon: const Icon(Icons.copy, size: 16),
                label: const Text('Copy'),
              ),
              const SizedBox(width: 8),
              if (url.isNotEmpty)
                _loadingFetch
                    ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                    : IconButton(
                      onPressed: () => _refetch(context, url),
                      icon: const Icon(Icons.refresh, size: 16),
                    ),
            ],
          ),
        if (body.isNotEmpty) const SizedBox(height: 6),
        if (body.isNotEmpty)
          (_isJson(body) && _treeResp)
              ? JsonViewer(jsonString: body, forceTree: true)
              : (_isJson(body) && _prettyResp)
              ? JsonViewer(jsonString: body, forceTree: false)
              : SelectableText(body, style: context.appText.monospace),
        const SizedBox(height: 8),
        Text('Headers', style: context.appText.subtitle),
        const SizedBox(height: 4),
        ...headers.entries
            .map(
              (e) => Padding(
                padding: const EdgeInsets.only(bottom: 2),
                child: _HeaderItem(
                  name: e.key,
                  value: e.value,
                  raw: headersRaw[e.key],
                ),
              ),
            )
            .toList(),
        const SizedBox(height: 12),
        // Security section (TLS & Cookies)
        Text('Security', style: context.appText.subtitle),
        const SizedBox(height: 6),
        ..._securityRows(resp, headers),
        const SizedBox(height: 12),
        // Cache & CORS section
        Text('Cache & CORS', style: context.appText.subtitle),
        const SizedBox(height: 6),
        Wrap(
          spacing: 8,
          runSpacing: 4,
          children: [
            if (cache['status'] != null)
              _chip(context, 'cache: ${cache['status']}'),
            _chip(context, cors['ok'] == true ? 'CORS OK' : 'CORS Fail'),
            if ((headers['Vary'] ?? headers['vary']) != null)
              _chip(context, 'Vary: ${(headers['Vary'] ?? headers['vary'])}'),
          ],
        ),
        const SizedBox(height: 6),
        // Cache table
        if (cache.isNotEmpty) ...[
          Text('Cache', style: Theme.of(context).textTheme.labelLarge),
          const SizedBox(height: 4),
          ..._kvList(cache).map(
            (e) => Padding(
              padding: const EdgeInsets.only(bottom: 2),
              child: SelectableText(
                '${e.key}: ${e.value}',
                style: context.appText.monospace,
              ),
            ),
          ),
          const SizedBox(height: 8),
        ],
        // CORS table
        Text('CORS', style: Theme.of(context).textTheme.labelLarge),
        const SizedBox(height: 4),
        ..._kvList({
          'origin': reqHeaders['Origin'] ?? reqHeaders['origin'] ?? '',
          'allowedOrigin': cors['allowedOrigin'] ?? '',
          'allowedMethods': (cors['allowedMethods'] ?? []).toString(),
          'allowedHeaders': (cors['allowedHeaders'] ?? []).toString(),
          'vary': headers['Vary'] ?? headers['vary'] ?? '',
          'preflight': (cors['preflight'] == true).toString(),
        }).map(
          (e) => Padding(
            padding: const EdgeInsets.only(bottom: 2),
            child: SelectableText(
              '${e.key}: ${e.value}',
              style: context.appText.monospace,
            ),
          ),
        ),
      ],
    );
  }

  Map<String, dynamic>? _findByType(List<dynamic> frames, String type) {
    for (final it in frames) {
      final mp = (it as Map<String, dynamic>);
      final preview = mp['preview']?.toString() ?? '';
      try {
        final obj = jsonDecode(preview) as Map<String, dynamic>;
        if (obj['type'] == type) return obj;
      } catch (_) {}
    }
    return null;
  }

  DateTime? _tsOf(List<dynamic> frames, String type) {
    for (final it in frames) {
      final mp = (it as Map<String, dynamic>);
      final preview = mp['preview']?.toString() ?? '';
      try {
        final obj = jsonDecode(preview) as Map<String, dynamic>;
        if (obj['type'] == type) {
          return DateTime.tryParse(mp['ts']?.toString() ?? '');
        }
      } catch (_) {}
    }
    return null;
  }

  bool _isJson(String s) {
    try {
      jsonDecode(s);
      return true;
    } catch (_) {
      return false;
    }
  }

  String _normalizeMaybeQuotedJson(String src) {
    if (src.isEmpty) return src;
    final t = src.trim();
    if ((t.startsWith('"') && t.endsWith('"')) ||
        (t.startsWith("'") && t.endsWith("'"))) {
      try {
        final unq = jsonDecode(t);
        if (unq is String) return unq;
      } catch (_) {}
    }
    return src;
  }

  Map<String, dynamic> _computeCacheMeta(
    int status,
    Map<String, String> respHdr,
  ) {
    final out = <String, dynamic>{};
    final cc =
        respHdr.entries
            .firstWhere(
              (e) => e.key.toLowerCase() == 'cache-control',
              orElse: () => const MapEntry('', ''),
            )
            .value;
    final etag =
        respHdr.entries
            .firstWhere(
              (e) => e.key.toLowerCase() == 'etag',
              orElse: () => const MapEntry('', ''),
            )
            .value;
    final ageStr =
        respHdr.entries
            .firstWhere(
              (e) => e.key.toLowerCase() == 'age',
              orElse: () => const MapEntry('', ''),
            )
            .value;
    final age = int.tryParse(ageStr) ?? 0;
    String st = 'MISS';
    if (status == 304)
      st = 'REVALIDATED';
    else if (age > 0)
      st = 'HIT';
    out['status'] = st;
    if (cc.isNotEmpty) out['cache-control'] = cc;
    if (etag.isNotEmpty) out['etag'] = etag;
    if (age > 0) out['age'] = age;
    return out;
  }

  Map<String, dynamic> _computeCorsMeta(
    String method,
    Map<String, String> reqHdr,
    Map<String, String> respHdr,
  ) {
    final out = <String, dynamic>{};
    final origin = _getFold(reqHdr, 'Origin');
    final allowOrigin = _getFold(respHdr, 'Access-Control-Allow-Origin');
    final allowMethods = _splitCsv(
      _getFold(respHdr, 'Access-Control-Allow-Methods'),
    );
    final allowHeaders = _splitCsv(
      _getFold(respHdr, 'Access-Control-Allow-Headers'),
    );
    final isPreflight =
        method.toUpperCase() == 'OPTIONS' &&
        _getFold(reqHdr, 'Access-Control-Request-Method').isNotEmpty;
    bool ok;
    String reason = '';
    if (origin.isEmpty) {
      ok = true;
    } else if (isPreflight) {
      final reqMethod =
          _getFold(reqHdr, 'Access-Control-Request-Method').toUpperCase();
      final reqHeaders = _splitCsv(
        _getFold(reqHdr, 'Access-Control-Request-Headers'),
      );
      final originOk = (allowOrigin == '*' || allowOrigin == origin);
      final methodOk =
          allowMethods.isEmpty
              ? true
              : allowMethods.map((e) => e.toUpperCase()).contains(reqMethod);
      final headersOk = reqHeaders.every(
        (h) => allowHeaders.any((a) => a.toLowerCase() == h.toLowerCase()),
      );
      ok = originOk && methodOk && headersOk;
      if (!originOk)
        reason = 'origin';
      else if (!methodOk)
        reason = 'method';
      else if (!headersOk)
        reason = 'headers';
    } else {
      final originOk = (allowOrigin == '*' || allowOrigin == origin);
      ok = originOk;
      if (!originOk) reason = 'origin';
    }
    out['ok'] = ok;
    if (reason.isNotEmpty) out['reason'] = reason;
    out['allowedOrigin'] = allowOrigin;
    out['allowedMethods'] = allowMethods;
    out['allowedHeaders'] = allowHeaders;
    out['preflight'] = isPreflight;
    return out;
  }

  String _getFold(Map<String, String> h, String key) {
    final lk = key.toLowerCase();
    for (final e in h.entries) {
      if (e.key.toLowerCase() == lk) return e.value;
    }
    return '';
  }

  List<MapEntry<String, String>> _kvList(Map<String, dynamic> m) {
    return m.entries
        .map(
          (e) => MapEntry(
            e.key,
            e.value is List ? (e.value as List).join(', ') : e.value.toString(),
          ),
        )
        .toList();
  }

  Uri? _tryParseUri(String s) {
    try {
      return Uri.parse(s);
    } catch (_) {
      return null;
    }
  }

  Map<String, String> _parseRequestCookies(String cookieHeader) {
    if (cookieHeader.isEmpty) return const {};
    final out = <String, String>{};
    final parts = cookieHeader.split(';');
    for (final p in parts) {
      final i = p.indexOf('=');
      if (i <= 0) continue;
      final k = p.substring(0, i).trim();
      final v = p.substring(i + 1).trim();
      if (k.isNotEmpty) out[k] = v;
    }
    return out;
  }

  Color _statusColor(BuildContext context, int status) {
    final cs = Theme.of(context).colorScheme;
    if (status >= 500) return cs.error;
    if (status >= 400) return cs.tertiary;
    if (status >= 300) return cs.primary;
    return Colors.green;
  }

  List<Widget> _securityRows(
    Map<String, dynamic> resp,
    Map<String, String> headers,
  ) {
    final List<Widget> out = [];
    // TLS summary from response preview if present
    final tls = resp['tls'];
    if (tls is Map) {
      out.addAll([
        SelectableText(
          'TLS: ${tls['version'] ?? ''} ${tls['cipherSuite'] ?? ''}  ALPN: ${tls['alpn'] ?? ''}',
          style: context.appText.monospace,
        ),
        if ((tls['serverName'] ?? '').toString().isNotEmpty)
          SelectableText(
            'SNI: ${tls['serverName']}',
            style: context.appText.monospace,
          ),
      ]);
      final certs =
          (tls['peerCertificates'] as List?)?.cast<Map<String, dynamic>>() ??
          const [];
      if (certs.isNotEmpty) {
        out.add(const SizedBox(height: 4));
        out.add(
          Text('Certificates', style: Theme.of(context).textTheme.labelLarge),
        );
        for (final c in certs) {
          out.add(
            Padding(
              padding: const EdgeInsets.only(bottom: 2),
              child: SelectableText(
                '- ${c['subject']} | Issuer: ${c['issuer']} | NotAfter: ${c['notAfter']}',
                style: context.appText.monospace,
              ),
            ),
          );
        }
      }
      out.add(const SizedBox(height: 6));
    }
    // Cookies summary
    final cs = resp['cookieSummary'];
    if (cs is Map) {
      out.add(Text('Cookies', style: Theme.of(context).textTheme.labelLarge));
      out.add(
        SelectableText(
          'Set-Cookie: ${cs['setCookieCount']} | Secure: ${cs['secure']} | HttpOnly: ${cs['httpOnly']} | SameSite Lax/Strict/None: ${cs['sameSiteLax']}/${cs['sameSiteStrict']}/${cs['sameSiteNone']}',
          style: context.appText.monospace,
        ),
      );
    }
    if (out.isEmpty)
      out.add(SelectableText('—', style: context.appText.monospace));
    return out;
  }

  List<String> _splitCsv(String s) {
    if (s.isEmpty) return const [];
    return s
        .split(',')
        .map((e) => e.trim())
        .where((e) => e.isNotEmpty)
        .toList(growable: false);
  }

  Future<void> _refetch(BuildContext context, String url) async {
    setState(() {
      _loadingFetch = true;
    });
    try {
      // ignore: invalid_use_of_protected_member
      final client = sl.get<Object>();
      final res = await (client as dynamic).get(
        path: '/httpproxy',
        query: {'_target': url},
      );
      final data = res.data?.toString() ?? '';
      // show modal with full body
      // ignore: use_build_context_synchronously
      await showModalBottomSheet(
        context: context,
        isScrollControlled: true,
        builder:
            (_) => Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Fetched Body',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Expanded(
                    child: SingleChildScrollView(
                      child: SelectableText(
                        data,
                        style: context.appText.monospace,
                      ),
                    ),
                  ),
                  Align(
                    alignment: Alignment.centerRight,
                    child: TextButton.icon(
                      onPressed: () {
                        Clipboard.setData(ClipboardData(text: data));
                      },
                      icon: const Icon(Icons.copy, size: 16),
                      label: const Text('Copy'),
                    ),
                  ),
                ],
              ),
            ),
      );
    } catch (e) {
      final msg = resolveErrorMessage(e);
      // ignore: use_build_context_synchronously
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('${msg.title}: ${msg.description}')),
      );
    } finally {
      setState(() {
        _loadingFetch = false;
      });
    }
  }

  Future<void> _refetchInline(String url) async {
    setState(() {
      _loadingFetch = true;
    });
    try {
      // ignore: invalid_use_of_protected_member
      final client = sl.get<Object>();
      final res = await (client as dynamic).get(
        path: '/httpproxy',
        query: {'_target': url},
      );
      final data = res.data?.toString() ?? '';
      setState(() {
        _bodyOverride = data;
      });
    } catch (_) {
      // тихо игнорируем
    } finally {
      if (mounted)
        setState(() {
          _loadingFetch = false;
        });
    }
  }

  Widget _chip(BuildContext context, String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceVariant,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(text, style: Theme.of(context).textTheme.labelSmall),
    );
  }

  String _buildCurl(Map<String, dynamic> req) {
    final method = (req['method'] ?? 'GET').toString().toUpperCase();
    final url = (req['url'] ?? '').toString();
    final headers =
        (req['headers'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};
    final body = (req['body'] ?? '').toString();
    final b = StringBuffer();
    b.write("curl -X $method '");
    b.write(url.replaceAll("'", "'\\''"));
    b.write("'");
    headers.forEach((k, v) {
      final vv = v.replaceAll("'", "'\\''");
      b.write(" -H '$k: $vv'");
    });
    if (body.isNotEmpty) {
      final bb = body.replaceAll("'", "'\\''");
      b.write(" --data '$bb'");
    }
    b.write(' --compressed');
    return b.toString();
  }
}

class _Card extends StatelessWidget {
  const _Card({required this.title, required this.child});
  final String title;
  final Widget child;
  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 1,
      margin: const EdgeInsets.all(8),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: Theme.of(
                context,
              ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}

class _HeaderItem extends StatefulWidget {
  const _HeaderItem({required this.name, required this.value, this.raw});
  final String name;
  final String value;
  final String? raw;
  @override
  State<_HeaderItem> createState() => _HeaderItemState();
}

class _HeaderItemState extends State<_HeaderItem> {
  bool _hover = false;
  bool _iconHover = false;
  bool _reveal = false;
  @override
  Widget build(BuildContext context) {
    final nameStyle = Theme.of(context).textTheme.bodySmall?.copyWith(
      fontFamily: 'monospace',
      fontWeight: FontWeight.w600,
    );
    final valueStyle = Theme.of(
      context,
    ).textTheme.bodySmall?.copyWith(fontFamily: 'monospace');
    final iconSize = valueStyle?.fontSize ?? 12;
    final iconColor = valueStyle?.color;
    final lname = widget.name.toLowerCase();
    final isSensitive =
        lname == 'authorization' ||
        lname == 'cookie' ||
        lname == 'set-cookie' ||
        lname.contains('token') ||
        lname.contains('secret') ||
        lname.contains('api-key') ||
        lname.contains('apikey');
    final raw = widget.raw ?? widget.value;
    final shownValue = (isSensitive && !_reveal) ? '***' : raw;
    return MouseRegion(
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('${widget.name}: ', style: nameStyle),
          Expanded(
            child: SelectableText.rich(
              TextSpan(
                style: valueStyle,
                children: [
                  TextSpan(text: shownValue),
                  if (_hover)
                    WidgetSpan(
                      alignment: PlaceholderAlignment.baseline,
                      baseline: TextBaseline.alphabetic,
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          if (isSensitive)
                            MouseRegion(
                              cursor: SystemMouseCursors.click,
                              onEnter: (_) => setState(() => _iconHover = true),
                              onExit: (_) => setState(() => _iconHover = false),
                              child: GestureDetector(
                                onTap: () {
                                  setState(() {
                                    _reveal = !_reveal;
                                  });
                                },
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 2,
                                    vertical: 1,
                                  ),
                                  margin: const EdgeInsets.only(left: 6),
                                  decoration: BoxDecoration(
                                    color:
                                        _iconHover
                                            ? Theme.of(context)
                                                .colorScheme
                                                .primary
                                                .withOpacity(0.12)
                                            : Colors.transparent,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: Icon(
                                    _reveal
                                        ? Icons.visibility_off
                                        : Icons.visibility,
                                    size: iconSize,
                                    color:
                                        _iconHover
                                            ? Theme.of(
                                              context,
                                            ).colorScheme.primary
                                            : iconColor,
                                  ),
                                ),
                              ),
                            ),
                          MouseRegion(
                            cursor: SystemMouseCursors.click,
                            onEnter: (_) => setState(() => _iconHover = true),
                            onExit: (_) => setState(() => _iconHover = false),
                            child: GestureDetector(
                              onTap: () {
                                Clipboard.setData(
                                  ClipboardData(text: '${widget.name}: $raw'),
                                );
                              },
                              child: Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 2,
                                  vertical: 1,
                                ),
                                margin: const EdgeInsets.only(left: 6),
                                decoration: BoxDecoration(
                                  color:
                                      _iconHover
                                          ? Theme.of(context)
                                              .colorScheme
                                              .primary
                                              .withOpacity(0.12)
                                          : Colors.transparent,
                                  borderRadius: BorderRadius.circular(4),
                                ),
                                child: Icon(
                                  Icons.copy,
                                  size: iconSize,
                                  color:
                                      _iconHover
                                          ? Theme.of(
                                            context,
                                          ).colorScheme.primary
                                          : iconColor,
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
