import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:app_http_client/application/app_http_client.dart';
import 'package:highlight_selectable/highlight_selectable.dart';
import 'package:highlight_selectable/theme_map.dart';
import '../../../inspector/presentation/utils/graphql_detect.dart';
import '../../../inspector/presentation/utils/graphql_formatter.dart';
import '../../../inspector/presentation/utils/body_view_mode.dart';
import '../../../inspector/presentation/utils/body_content_analyzer.dart';
import '../../../inspector/presentation/widgets/jwt_viewer.dart';
import '../../../inspector/application/stores/session_details_store.dart';
import '../../../../core/network/error_utils.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_viewer.dart';
import '../../../../core/di/di.dart';
import '../../../../widgets/html_preview.dart';
import 'form_data_view.dart';
import 'highlighted_url_text.dart';
import '../../../tags/presentation/widgets/tags_editor.dart';

enum CurlExportMode { compact, multiline, withOptions }

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
  // View controllers for request and response bodies
  final _reqViewController = BodyViewController();
  final _respViewController = BodyViewController();
  final _contentAnalyzer = BodyContentAnalyzer();

  bool _loadingFetch = false;
  int? _respTtfbMs;
  int? _respTotalMs;
  String? _bodyOverride;
  String? _highlightThemeKey;

  Future<void> _openTagsDialog() async {
    final sid = widget.sessionId;
    if (sid == null || sid.isEmpty) return;
    await showDialog<void>(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          title: const Text('Tags'),
          content: SizedBox(
            width: 520,
            child: SingleChildScrollView(child: TagsEditor(sessionId: sid)),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(),
              child: const Text('Close'),
            ),
          ],
        );
      },
    );
  }

  void _ensureHighlightThemeInitialized(BuildContext context) {
    if (_highlightThemeKey != null) return;
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final def = isDark ? 'a11y-dark' : 'a11y-light';
    _fetchHighlightThemeFromBackend(isDark: isDark).then((saved) {
      if (!mounted) return;
      setState(() {
        _highlightThemeKey = (saved != null && themeMap.containsKey(saved))
            ? saved
            : def;
      });
    });
    _highlightThemeKey = def;
  }

  Future<String?> _fetchHighlightThemeFromBackend({
    required bool isDark,
  }) async {
    try {
      final api = sl<AppHttpClient>();
      final res = await api.get(path: '/_api/v1/settings');
      final data = (res.data as Map).cast<String, dynamic>();
      final light = (data['highlightThemeLight'] ?? '').toString();
      final dark = (data['highlightThemeDark'] ?? '').toString();
      final legacy = (data['highlightTheme'] ?? '').toString();
      if (isDark) {
        final k = dark.isNotEmpty ? dark : legacy;
        return k.isEmpty ? null : k;
      } else {
        final k = light.isNotEmpty ? light : legacy;
        return k.isEmpty ? null : k;
      }
    } catch (_) {
      return null;
    }
  }

  Map<String, TextStyle> _currentHlTheme() {
    final key = _highlightThemeKey;
    return themeMap[key] ?? themeMap['a11y-dark']!;
  }

  String _detectLanguage(String text, String? contentType) {
    final ct = (contentType ?? '').toLowerCase();
    if (ct.contains('html') || _looksLikeHtml(text)) {
      return 'html';
    }
    if (ct.contains('xml') || text.trimLeft().startsWith('<?xml')) {
      return 'xml';
    }
    if (ct.contains('javascript') || ct.contains('ecmascript')) {
      return 'javascript';
    }
    if (ct.contains('typescript')) {
      return 'typescript';
    }
    if (ct.contains('css')) {
      return 'css';
    }
    if (GraphqlLanguageDetector.isLikelyGraphql(text)) {
      return 'graphql';
    }
    return 'plaintext';
  }

  Widget _highlightBlock(String body, {String? contentType}) {
    if (body.isEmpty) return const SizedBox.shrink();
    final lang = _detectLanguage(body, contentType);
    final theme = _currentHlTheme();
    final bgColor = theme['root']?.backgroundColor;
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Theme.of(context).dividerColor),
        color: bgColor,
      ),
      clipBehavior: Clip.antiAlias,
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: HighlightSelectable(
          body,
          language: lang,
          theme: theme,
          selectable: true,
          showCopyButton: true,
          showEditButton: false,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    _ensureHighlightThemeInitialized(context);
    final req = _findByType(widget.frames, 'http_request');
    final resp = _findByType(widget.frames, 'http_response');
    final reqTs = _tsOf(widget.frames, 'http_request');
    final respTs = _tsOf(widget.frames, 'http_response');
    final durationMs = (reqTs != null && respTs != null)
        ? respTs.difference(reqTs).inMilliseconds
        : null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: Wrap(
                spacing: 8,
                crossAxisAlignment: WrapCrossAlignment.center,
                children: [
                  SizedBox(width: 4),
                  Text('HTTP Details', style: context.appText.title),
                  if (durationMs != null) _chip(context, '$durationMs ms'),
                ],
              ),
            ),
            TextButton.icon(
              onPressed:
                  (widget.sessionId == null || (widget.sessionId ?? '').isEmpty)
                  ? null
                  : _openTagsDialog,
              icon: const Icon(Icons.label_outline, size: 16),
              label: const Text('Tags'),
            ),
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
    if (req == null) {
      final method = (widget.httpMeta?['method'] ?? '')
          .toString()
          .toUpperCase();
      if (method == 'CONNECT') return _connectPlaceholder(context);

      // Check loading/error states from store
      final details = context.watch<SessionDetailsStore>();

      // Show loading indicator if currently loading and no frames yet
      if (details.loading && widget.frames.isEmpty) {
        return Center(
          child: Padding(
            padding: const EdgeInsets.all(32.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const CircularProgressIndicator(),
                const SizedBox(height: 16),
                Text('Loading request data...', style: context.appText.body),
              ],
            ),
          ),
        );
      }

      // Show error state if there was a load error
      if (details.loadError != null) {
        return Center(
          child: Padding(
            padding: const EdgeInsets.all(32.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Icons.error_outline,
                  size: 48,
                  color: context.appColors.danger,
                ),
                const SizedBox(height: 16),
                Text('Failed to load data', style: context.appText.subtitle),
                const SizedBox(height: 8),
                Text(
                  details.loadError!,
                  style: context.appText.body.copyWith(
                    color: context.appColors.textSecondary,
                  ),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
        );
      }

      // No data (successful load but no HTTP request found)
      return Center(
        child: Text('No request data', style: context.appText.body),
      );
    }
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
    final ctHeader = headers.entries
        .firstWhere(
          (e) => e.key.toLowerCase() == 'content-type',
          orElse: () => const MapEntry('', ''),
        )
        .value;
    final ctLower = ctHeader.toLowerCase();
    final hasFormObj = req['form'] is Map && (req['form'] as Map).isNotEmpty;
    final isFormCt =
        ctLower.contains('multipart/form-data') ||
        ctLower.contains('application/x-www-form-urlencoded');
    final hideRawBody = hasFormObj || isFormCt;
    final url = (req['url'] ?? '').toString();
    final uri = _tryParseUri(url);
    // Нормализуем возможные артефакты вида "?%3F..." и ключи параметров, начинающиеся с '?'
    final rawQp = uri?.queryParametersAll ?? <String, List<String>>{};
    final qp = _normalizeQueryKeys(rawQp);
    final normalizedUrl = (uri == null) ? null : _rebuildUrlWithQuery(uri, qp);
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
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              SelectableText(
                '${(req['method'] ?? '').toString().toUpperCase()}  ',
                style: context.appText.subtitle,
              ),
              HighlightedUrlText(
                url: normalizedUrl ?? url,
                baseStyle: context.appText.subtitle,
              ),
            ],
          ),
        ),
        const SizedBox(height: 6),
        // Кнопки на новой строке, чтобы не мешали URL
        Wrap(
          spacing: 8,
          children: [
            TextButton.icon(
              onPressed: () {
                Clipboard.setData(ClipboardData(text: normalizedUrl ?? url));
              },
              icon: const Icon(Icons.link, size: 16),
              label: const Text('Copy URL'),
            ),
            MenuAnchor(
              builder: (context, controller, child) {
                return TextButton.icon(
                  onPressed: () {
                    // Default: copy compact version
                    final curl = _buildCurl(req, CurlExportMode.compact);
                    Clipboard.setData(ClipboardData(text: curl));
                  },
                  onLongPress: () {
                    if (controller.isOpen) {
                      controller.close();
                    } else {
                      controller.open();
                    }
                  },
                  icon: const Icon(Icons.content_paste, size: 16),
                  label: const Text('Copy as cURL'),
                );
              },
              menuChildren: [
                MenuItemButton(
                  leadingIcon: const Icon(Icons.remove, size: 16),
                  child: const Text('Compact'),
                  onPressed: () {
                    final curl = _buildCurl(req, CurlExportMode.compact);
                    Clipboard.setData(ClipboardData(text: curl));
                  },
                ),
                MenuItemButton(
                  leadingIcon: const Icon(Icons.wrap_text, size: 16),
                  child: const Text('Multiline'),
                  onPressed: () {
                    final curl = _buildCurl(req, CurlExportMode.multiline);
                    Clipboard.setData(ClipboardData(text: curl));
                  },
                ),
                MenuItemButton(
                  leadingIcon: const Icon(Icons.settings, size: 16),
                  child: const Text('With options'),
                  onPressed: () {
                    final curl = _buildCurl(req, CurlExportMode.withOptions);
                    Clipboard.setData(ClipboardData(text: curl));
                  },
                ),
              ],
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
              child: _KeyValueItem(name: e.key, value: e.value.join(', ')),
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
              child: _KeyValueItem(name: e.key, value: e.value),
            ),
          ),
          const SizedBox(height: 8),
        ],
        // Form Data (multipart/urlencoded) – если есть, показываем перед сырым Body
        FormDataView(
          form: (req['form'] is Map)
              ? (req['form'] as Map).cast<String, dynamic>()
              : null,
          contentType: ctHeader,
          rawBody: body,
        ),
        const SizedBox(height: 6),
        if (body.isNotEmpty && !hideRawBody) ...[
          // Analyze content and initialize view controller
          Builder(
            builder: (ctx) {
              final analysis = _contentAnalyzer.analyze(
                body,
                contentType: ctHeader,
              );
              _reqViewController.setAvailableModes(analysis.availableModes);

              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    children: [
                      Text('Body', style: context.appText.subtitle),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _buildBodyViewChips(
                          body: body,
                          controller: _reqViewController,
                          analysis: analysis,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  _renderBodyContent(
                    body: body,
                    controller: _reqViewController,
                    analysis: analysis,
                    contentType: ctHeader,
                  ),
                ],
              );
            },
          ),
        ],
        const SizedBox(height: 8),
        Text('Headers', style: context.appText.subtitle),
        const SizedBox(height: 4),
        ...headers.entries.map(
          (e) => Padding(
            padding: const EdgeInsets.only(bottom: 2),
            child: _HeaderItem(
              name: e.key,
              value: e.value,
              raw: headersRaw[e.key],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildResponse(BuildContext context, Map<String, dynamic>? resp) {
    if (resp == null) {
      final method = (widget.httpMeta?['method'] ?? '')
          .toString()
          .toUpperCase();
      if (method == 'CONNECT') return _connectPlaceholder(context);

      // Check loading/error states from store
      final details = context.watch<SessionDetailsStore>();

      // Show loading indicator if currently loading and no frames yet
      if (details.loading && widget.frames.isEmpty) {
        return Center(
          child: Padding(
            padding: const EdgeInsets.all(32.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const CircularProgressIndicator(),
                const SizedBox(height: 16),
                Text('Loading response data...', style: context.appText.body),
              ],
            ),
          ),
        );
      }

      // Show error state if there was a load error
      if (details.loadError != null) {
        return Center(
          child: Padding(
            padding: const EdgeInsets.all(32.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Icons.error_outline,
                  size: 48,
                  color: context.appColors.danger,
                ),
                const SizedBox(height: 16),
                Text('Failed to load data', style: context.appText.subtitle),
                const SizedBox(height: 8),
                Text(
                  details.loadError!,
                  style: context.appText.body.copyWith(
                    color: context.appColors.textSecondary,
                  ),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
        );
      }

      // No data (successful load but no HTTP response found)
      return Center(
        child: Text('No response data', style: context.appText.body),
      );
    }
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
    final ctHeader = headers.entries
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
    final durationMs = (reqTs != null && respTs != null)
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
                  _chip(context, 'Total: $_respTotalMs ms')
                else if (durationMs != null)
                  _chip(context, 'Total: $durationMs ms'),
                if (_respTtfbMs != null)
                  _chip(context, 'TTFB: $_respTtfbMs ms'),
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
        if (body.isNotEmpty) ...[
          // Analyze content and initialize view controller
          Builder(
            builder: (ctx) {
              final analysis = _contentAnalyzer.analyze(
                body,
                contentType: ctHeader,
              );
              _respViewController.setAvailableModes(analysis.availableModes);

              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    children: [
                      Text('Body', style: context.appText.subtitle),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Wrap(
                          spacing: 8,
                          runSpacing: 4,
                          crossAxisAlignment: WrapCrossAlignment.center,
                          children: [
                            ...analysis.availableModes.map((mode) {
                              return FilterChip(
                                label: Text(BodyViewController.getLabel(mode)),
                                selected: _respViewController.isActive(mode),
                                onSelected: (_) {
                                  setState(() {
                                    _respViewController.switchTo(mode);
                                  });
                                },
                              );
                            }),
                            const SizedBox(width: 4),
                            TextButton.icon(
                              onPressed: () {
                                Clipboard.setData(ClipboardData(text: body));
                              },
                              icon: const Icon(Icons.copy, size: 16),
                              label: const Text('Copy'),
                            ),
                            if (url.isNotEmpty)
                              _loadingFetch
                                  ? const SizedBox(
                                      width: 16,
                                      height: 16,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                      ),
                                    )
                                  : IconButton(
                                      onPressed: () => _refetch(context, url),
                                      icon: const Icon(Icons.refresh, size: 16),
                                      tooltip: 'Refetch full body',
                                    ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  _renderBodyContent(
                    body: body,
                    controller: _respViewController,
                    analysis: analysis,
                    contentType: ctHeader,
                    baseUrl: url.isNotEmpty ? url : null,
                  ),
                ],
              );
            },
          ),
        ],
        const SizedBox(height: 8),
        Text('Headers', style: context.appText.subtitle),
        const SizedBox(height: 4),
        ...headers.entries.map(
          (e) => Padding(
            padding: const EdgeInsets.only(bottom: 2),
            child: _HeaderItem(
              name: e.key,
              value: e.value,
              raw: headersRaw[e.key],
            ),
          ),
        ),
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
              child: _KeyValueItem(name: e.key, value: e.value),
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
            child: _KeyValueItem(name: e.key, value: e.value),
          ),
        ),
      ],
    );
  }

  Widget _connectPlaceholder(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('CONNECT tunnel', style: context.appText.subtitle),
          const SizedBox(height: 6),
          SelectableText(
            'This is an HTTPS tunnel. To see Request/Response you need MITM enabled and the dev CA trusted by the OS.',
            style: context.appText.body,
          ),
          const SizedBox(height: 6),
          SelectableText(
            'Check: 1) Proxy: ON, 2) OS trust: ON, 3) MITM enabled, 4) QUIC disabled in browser, 5) domain not in deny list.',
            style: context.appText.monospace,
          ),
        ],
      ),
    );
  }

  bool _looksLikeHtml(String s) {
    final t = s.trimLeft();
    if (t.isEmpty) return false;
    final lower = t.toLowerCase();
    return lower.startsWith('<!doctype html') ||
        lower.startsWith('<html') ||
        lower.contains('<head>') ||
        lower.contains('<body');
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
    final cc = respHdr.entries
        .firstWhere(
          (e) => e.key.toLowerCase() == 'cache-control',
          orElse: () => const MapEntry('', ''),
        )
        .value;
    final etag = respHdr.entries
        .firstWhere(
          (e) => e.key.toLowerCase() == 'etag',
          orElse: () => const MapEntry('', ''),
        )
        .value;
    final ageStr = respHdr.entries
        .firstWhere(
          (e) => e.key.toLowerCase() == 'age',
          orElse: () => const MapEntry('', ''),
        )
        .value;
    final age = int.tryParse(ageStr) ?? 0;
    String st = 'MISS';
    if (status == 304) {
      st = 'REVALIDATED';
    } else if (age > 0) {
      st = 'HIT';
    }
    out['status'] = st;
    if (cc.isNotEmpty) {
      out['cache-control'] = cc;
    }
    if (etag.isNotEmpty) {
      out['etag'] = etag;
    }
    if (age > 0) {
      out['age'] = age;
    }
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
      final reqMethod = _getFold(
        reqHdr,
        'Access-Control-Request-Method',
      ).toUpperCase();
      final reqHeaders = _splitCsv(
        _getFold(reqHdr, 'Access-Control-Request-Headers'),
      );
      final originOk = (allowOrigin == '*' || allowOrigin == origin);
      final methodOk = allowMethods.isEmpty
          ? true
          : allowMethods.map((e) => e.toUpperCase()).contains(reqMethod);
      final headersOk = reqHeaders.every(
        (h) => allowHeaders.any((a) => a.toLowerCase() == h.toLowerCase()),
      );
      ok = originOk && methodOk && headersOk;
      if (!originOk) {
        reason = 'origin';
      } else if (!methodOk) {
        reason = 'method';
      } else if (!headersOk) {
        reason = 'headers';
      }
    } else {
      final originOk = (allowOrigin == '*' || allowOrigin == origin);
      ok = originOk;
      if (!originOk) {
        reason = 'origin';
      }
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

  Map<String, List<String>> _normalizeQueryKeys(Map<String, List<String>> src) {
    if (src.isEmpty) return const {};
    final out = <String, List<String>>{};
    src.forEach((k, values) {
      final key = k.startsWith('?') ? k.substring(1) : k;
      final dst = out.putIfAbsent(key, () => <String>[]);
      dst.addAll(values);
    });
    return out;
  }

  String _rebuildUrlWithQuery(Uri uri, Map<String, List<String>> qp) {
    // Собираем URL из базовой части + нормализованный query
    final base = uri.replace(query: null, queryParameters: null);
    if (qp.isEmpty) return base.toString();
    final parts = <String>[];
    qp.forEach((k, values) {
      for (final v in values) {
        parts.add(
          '${Uri.encodeQueryComponent(k)}=${Uri.encodeQueryComponent(v)}',
        );
      }
    });
    final q = parts.join('&');
    return base.replace(query: q).toString();
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
    if (out.isEmpty) {
      out.add(SelectableText('—', style: context.appText.monospace));
    }
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
      final client = sl<AppHttpClient>();
      final res = await client.get(path: '/httpproxy', query: {'_target': url});
      final data = res.data?.toString() ?? '';
      // show modal with full body
      if (!mounted) {
        return;
      }
      await showModalBottomSheet(
        // ignore: use_build_context_synchronously
        context: context,
        isScrollControlled: true,
        builder: (_) => Padding(
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
                  child: SelectableText(data, style: context.appText.monospace),
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
      if (!mounted) {
        return;
      }
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
      final client = sl<AppHttpClient>();
      final res = await client.get(path: '/httpproxy', query: {'_target': url});
      final data = res.data?.toString() ?? '';
      setState(() {
        _bodyOverride = data;
      });
    } catch (_) {
      // тихо игнорируем
    } finally {
      if (mounted) {
        setState(() {
          _loadingFetch = false;
        });
      }
    }
  }

  Widget _chip(BuildContext context, String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(text, style: Theme.of(context).textTheme.labelSmall),
    );
  }

  String _buildCurl(
    Map<String, dynamic> req, [
    CurlExportMode mode = CurlExportMode.compact,
  ]) {
    final method = (req['method'] ?? 'GET').toString().toUpperCase();
    final url = (req['url'] ?? '').toString();

    // Validate URL
    if (url.isEmpty) {
      return '# Error: URL is empty';
    }

    final headers =
        (req['headers'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};
    final body = (req['body'] ?? '').toString();
    final form = req['form'] is Map
        ? Map<String, dynamic>.from(req['form'] as Map)
        : null;

    // Extract cookies from headers
    String? cookieValue;
    final headersToInclude = <String, String>{};
    headers.forEach((k, v) {
      if (k.toLowerCase() == 'cookie') {
        cookieValue = v;
      } else {
        headersToInclude[k] = v;
      }
    });

    // Determine if we should use form-data format
    final contentType =
        headers['Content-Type'] ?? headers['content-type'] ?? '';
    final useFormData =
        form != null &&
        (contentType.contains('multipart/form-data') ||
            contentType.contains('application/x-www-form-urlencoded'));

    final b = StringBuffer();
    final isMultiline =
        mode == CurlExportMode.multiline || mode == CurlExportMode.withOptions;
    final newline = isMultiline ? ' \\\n  ' : ' ';

    // Start command
    b.write("curl -X $method$newline'");
    b.write(_escapeShellArg(url));
    b.write("'");

    // Add headers
    headersToInclude.forEach((k, v) {
      // Skip Content-Type and Content-Length for form-data (curl adds them automatically)
      if (useFormData &&
          (k.toLowerCase() == 'content-type' ||
              k.toLowerCase() == 'content-length')) {
        return;
      }
      b.write("$newline-H '$k: ${_escapeShellArg(v)}'");
    });

    // Add cookie
    if ((cookieValue ?? '').isNotEmpty) {
      b.write("$newline--cookie '${_escapeShellArg(cookieValue!)}'");
    }

    // Add body or form data
    if (useFormData) {
      // Use -F for form data
      // Form structure: {type: "multipart", fields: [...], files: [...]}
      final fields =
          (form['fields'] as List?)?.cast<Map<String, dynamic>>() ?? [];
      final files =
          (form['files'] as List?)?.cast<Map<String, dynamic>>() ?? [];

      // Add regular fields
      for (final field in fields) {
        final name = field['name']?.toString() ?? '';
        final value =
            field['valuePreview']?.toString() ??
            field['value']?.toString() ??
            '';
        if (name.isNotEmpty) {
          b.write("$newline-F '$name=${_escapeShellArg(value)}'");
        }
      }

      // Add file uploads (Note: file paths need to be manually adjusted)
      if (files.isNotEmpty) {
        b.write("$newline# Note: Replace file paths with actual local paths");
      }
      for (final file in files) {
        final name = file['name']?.toString() ?? '';
        final filename = file['filename']?.toString() ?? 'file';
        if (name.isNotEmpty) {
          b.write("$newline-F '$name=@$filename'");
        }
      }
    } else if (body.isNotEmpty) {
      // Use --data for regular body
      b.write("$newline--data '${_escapeShellArg(body)}'");
    }

    // Add additional options for withOptions mode
    if (mode == CurlExportMode.withOptions) {
      b.write("$newline--location");
      b.write("$newline--insecure");
    }

    b.write("$newline--compressed");

    return b.toString();
  }

  String _escapeShellArg(String arg) {
    return arg.replaceAll("'", "'\\''");
  }

  /// Build body view mode chips (DRY - reusable for request and response)
  Widget _buildBodyViewChips({
    required String body,
    required BodyViewController controller,
    required ContentAnalysisResult analysis,
  }) {
    return Wrap(
      spacing: 8,
      runSpacing: 4,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        ...analysis.availableModes.map((mode) {
          return FilterChip(
            label: Text(BodyViewController.getLabel(mode)),
            selected: controller.isActive(mode),
            onSelected: (_) {
              setState(() {
                controller.switchTo(mode);
              });
            },
          );
        }),
        const SizedBox(width: 4),
        TextButton.icon(
          onPressed: () {
            Clipboard.setData(ClipboardData(text: body));
          },
          icon: const Icon(Icons.copy, size: 16),
          label: const Text('Copy'),
        ),
      ],
    );
  }

  /// Render body content based on current view mode (DRY - reusable)
  Widget _renderBodyContent({
    required String body,
    required BodyViewController controller,
    required ContentAnalysisResult analysis,
    required String? contentType,
    String? baseUrl,
  }) {
    final mode = controller.current;

    switch (mode) {
      case BodyViewMode.jwtDecoded:
        return JwtViewer(token: body, jwtData: analysis.jwtData);

      case BodyViewMode.base64Decoded:
        final decoded = analysis.base64Data?.decoded ?? body;
        final isJson = analysis.base64Data?.isJson ?? false;
        if (isJson) {
          return JsonViewer(jsonString: decoded, forceTree: false);
        }
        return SelectableText(decoded, style: context.appText.monospace);

      case BodyViewMode.jsonTree:
        return JsonViewer(jsonString: body, forceTree: true);

      case BodyViewMode.pretty:
        if (analysis.isJson) {
          return JsonViewer(jsonString: body, forceTree: false);
        }
        // Auto-format GraphQL queries in Pretty mode
        if (GraphqlLanguageDetector.isLikelyGraphql(body)) {
          final formatted = GraphQLFormatter.format(body);
          return _highlightBlock(formatted, contentType: contentType);
        }
        return _highlightBlock(body, contentType: contentType);

      case BodyViewMode.htmlPreview:
        return SizedBox(
          height: 480,
          child: HtmlPreview(html: body, baseUrl: baseUrl),
        );

      case BodyViewMode.raw:
        return SelectableText(body, style: context.appText.monospace);
    }
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

class _KeyValueItem extends StatelessWidget {
  const _KeyValueItem({required this.name, required this.value});
  final String name;
  final String value;

  @override
  Widget build(BuildContext context) {
    return SelectableText.rich(
      TextSpan(
        children: [
          TextSpan(
            text: '$name: ',
            style: context.appText.monospace.copyWith(
              fontWeight: FontWeight.w600,
            ),
          ),
          TextSpan(
            text: value,
            style: context.appText.monospace.copyWith(
              color: context.appColors.textSecondary,
            ),
          ),
        ],
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
    final valueStyle = Theme.of(context).textTheme.bodySmall?.copyWith(
      fontFamily: 'monospace',
      color: context.appColors.textSecondary,
    );
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
                                    color: _iconHover
                                        ? Theme.of(context).colorScheme.primary
                                              .withValues(alpha: 0.12)
                                        : Colors.transparent,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: Icon(
                                    _reveal
                                        ? Icons.visibility_off
                                        : Icons.visibility,
                                    size: iconSize,
                                    color: _iconHover
                                        ? Theme.of(context).colorScheme.primary
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
                                  color: _iconHover
                                      ? Theme.of(context).colorScheme.primary
                                            .withValues(alpha: 0.12)
                                      : Colors.transparent,
                                  borderRadius: BorderRadius.circular(4),
                                ),
                                child: Icon(
                                  Icons.copy,
                                  size: iconSize,
                                  color: _iconHover
                                      ? Theme.of(context).colorScheme.primary
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
