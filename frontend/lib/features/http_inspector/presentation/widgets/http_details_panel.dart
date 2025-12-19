import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:app_http_client/application/app_http_client.dart';
import 'package:highlight_selectable/highlight_selectable.dart';
import 'package:highlight_selectable/theme_map.dart';
import 'package:http/http.dart' as http;
import '../../../inspector/presentation/utils/graphql_detect.dart';
import '../../../inspector/presentation/utils/graphql_formatter.dart';
import '../../../inspector/presentation/utils/body_view_mode.dart';
import '../../../inspector/presentation/utils/body_content_analyzer.dart';
import '../../../inspector/presentation/widgets/jwt_viewer.dart';
import '../../../inspector/presentation/widgets/body_fullscreen_dialog.dart';
import '../../../inspector/application/stores/session_details_store.dart';
import '../../../../core/network/error_utils.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_viewer.dart';
import '../../../../widgets/json_search_controller.dart';
import '../../../../core/di/di.dart';
import '../../../../widgets/html_preview.dart';
import '../../../../widgets/highlighted_url_widget.dart';
import '../../../../widgets/http_method_chip.dart';
import '../../../tags/presentation/widgets/tags_editor.dart';
import '../../../inspector/presentation/widgets/hex_body_renderer.dart';
import '../../../compose/domain/usecases/convert_request_to_template.dart';
import '../../../compose/presentation/pages/compose_page.dart';
import 'request_action_buttons.dart';
import 'request_body_tab.dart';
import 'request_info_tab.dart';
import 'response_body_tab.dart';
import 'response_info_tab.dart';

// Callback type aliases for clean interfaces
typedef BodyAnalyzer =
    ContentAnalysisResult Function(String body, String? contentType);

typedef BodyViewChipsBuilder =
    Widget Function({
      required String body,
      required BodyViewController controller,
      required ContentAnalysisResult analysis,
      String? title,
      String? contentType,
      String? baseUrl,
      String? frameId,
      int? bodySize,
      int? bodyRawSize,
    });

typedef BodyContentSliverRenderer =
    List<Widget> Function({
      required String body,
      required BodyViewController controller,
      required ContentAnalysisResult analysis,
      required String? contentType,
      required bool isRequest,
      String? baseUrl,
      String? frameId,
      int? bodySize,
      int? bodyRawSize,
    });

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

class _HttpDetailsPanelState extends State<HttpDetailsPanel>
    with TickerProviderStateMixin {
  // View controllers for request and response bodies
  final _reqViewController = BodyViewController();
  final _respViewController = BodyViewController();
  final _contentAnalyzer = BodyContentAnalyzer();

  // Tab controllers for Request and Response sections
  late TabController _reqTabController;
  late TabController _respTabController;

  bool _loadingFetch = false;
  int? _respTtfbMs;
  int? _respTotalMs;
  String? _bodyOverride;
  String? _highlightThemeKey;

  // Full body cache (panel-level) - loaded once and shared across all view modes
  String? _fullReqBody;
  String? _fullRespBody;
  bool _reqBodyLoading = false;
  bool _respBodyLoading = false;
  String? _reqBodyError;
  String? _respBodyError;

  // Cache for expensive content analysis (to avoid re-analyzing on every rebuild)
  String? _lastReqBody;
  String? _lastReqContentType;
  ContentAnalysisResult? _reqAnalysis;
  String? _lastRespBody;
  String? _lastRespContentType;
  ContentAnalysisResult? _respAnalysis;

  @override
  void initState() {
    super.initState();
    // Initialize with 2 tabs (will be recreated if needed)
    _reqTabController = TabController(length: 2, vsync: this);
    _respTabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _reqTabController.dispose();
    _respTabController.dispose();
    super.dispose();
  }

  /// Recreate tab controller if number of tabs changed
  void _ensureTabController({
    required TabController controller,
    required int newLength,
    required bool isRequest,
  }) {
    if (controller.length != newLength) {
      controller.dispose();
      final newController = TabController(length: newLength, vsync: this);
      if (isRequest) {
        _reqTabController = newController;
      } else {
        _respTabController = newController;
      }
    }
  }

  /// Compare frames by content (id) instead of reference to avoid unnecessary reloads
  bool _framesEqual(List<dynamic> a, List<dynamic> b) {
    if (a.length != b.length) return false;
    for (int i = 0; i < a.length; i++) {
      final aMap = a[i] as Map<String, dynamic>?;
      final bMap = b[i] as Map<String, dynamic>?;
      if (aMap?['id'] != bMap?['id']) return false;
    }
    return true;
  }

  @override
  void didUpdateWidget(HttpDetailsPanel oldWidget) {
    super.didUpdateWidget(oldWidget);

    // Reset cache if frames or sessionId changed (different HTTP request selected)
    // Compare frames by content (id) to avoid spurious reloads from Observer rebuilds
    if (!_framesEqual(oldWidget.frames, widget.frames) ||
        oldWidget.sessionId != widget.sessionId) {
      setState(() {
        // Clear analysis cache
        _lastReqBody = null;
        _lastReqContentType = null;
        _reqAnalysis = null;
        _lastRespBody = null;
        _lastRespContentType = null;
        _respAnalysis = null;

        // Clear full body cache and reset loading states
        _fullReqBody = null;
        _fullRespBody = null;
        _reqBodyLoading = false;
        _respBodyLoading = false;
        _reqBodyError = null;
        _respBodyError = null;
      });

      // Trigger loading for new frames (ONCE, not in build())
      _triggerFullBodyLoadIfNeeded();
    }
  }

  /// Trigger full body loading for request and response (called once on frame change)
  void _triggerFullBodyLoadIfNeeded() {
    // Extract request data
    final req = _findByType(widget.frames, 'http_request');
    final reqFrame = _findFrameByType(widget.frames, 'http_request');

    if (req != null && reqFrame != null) {
      final reqBody = (req['body'] ?? '').toString();
      final reqFrameId = reqFrame['frame']?['id']?.toString();
      // Use Frame.Size (ContentLength) as total size
      final reqBodySize = reqFrame['frame']?['size'] as int?;
      // Use bodyRawSize from preview (actual bytes read) to determine if truncated
      final reqBodyRawSize = req['bodyRawSize'] as int?;
      _loadFullBodyIfNeeded(
        frameId: reqFrameId,
        preview: reqBody,
        bodySize: reqBodySize,
        bodyRawSize: reqBodyRawSize,
        isRequest: true,
      );
    }

    // Extract response data
    final resp = _findByType(widget.frames, 'http_response');
    final respFrame = _findFrameByType(widget.frames, 'http_response');

    if (resp != null && respFrame != null) {
      final respBody = (resp['body'] ?? '').toString();
      final respFrameId = respFrame['frame']?['id']?.toString();
      // Use Frame.Size (ContentLength) as total size
      final respBodySize = respFrame['frame']?['size'] as int?;
      // Use bodyRawSize from preview (actual bytes read) to determine if truncated
      final respBodyRawSize = resp['bodyRawSize'] as int?;
      _loadFullBodyIfNeeded(
        frameId: respFrameId,
        preview: respBody,
        bodySize: respBodySize,
        bodyRawSize: respBodyRawSize,
        isRequest: false,
      );
    }
  }

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

  /// Открывает Compose с данными текущего запроса для редактирования
  void _openComposeWithRequest(Map<String, dynamic> req) {
    final headersRaw =
        (req['headersRaw'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};

    final useCase = sl<ConvertRequestToTemplateUseCase>();
    final template = useCase(
      requestData: req,
      headersRaw: headersRaw.isNotEmpty ? headersRaw : null,
    );

    Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => ComposePage(initialTemplate: template),
      ),
    );
  }

  /// Load full body data from API if needed (truncated preview)
  /// This is panel-level caching - loaded once and shared across all view modes for efficiency
  Future<void> _loadFullBodyIfNeeded({
    required String? frameId,
    required String preview,
    required int? bodySize,
    int? bodyRawSize,
    required bool isRequest,
  }) async {
    final sessionId = widget.sessionId;
    if (sessionId == null || frameId == null || bodySize == null) return;

    // Check if full data is needed (preview is truncated)
    // Use bodyRawSize (actual bytes read into preview) if available
    // Body is truncated only if ContentLength > bodyRawSize (bytes actually read)
    if (bodyRawSize != null) {
      // Precise check: if all bytes were read into preview, no need to load full body
      if (bodySize <= bodyRawSize) return;
    } else {
      // Fallback for old data without bodyRawSize
      final previewLength = utf8.encode(preview).length;
      if (bodySize <= previewLength) return;
    }

    // Check if already loaded, loading, or has error (prevent auto-retry on error)
    if (isRequest) {
      if (_fullReqBody != null || _reqBodyLoading || _reqBodyError != null) {
        return;
      }
    } else {
      if (_fullRespBody != null || _respBodyLoading || _respBodyError != null) {
        return;
      }
    }

    if (!mounted) return;

    // Start loading
    setState(() {
      if (isRequest) {
        _reqBodyLoading = true;
      } else {
        _respBodyLoading = true;
      }
    });

    try {
      // Use http package for direct API call (similar to HexBodyRenderer)
      final baseUrl = sl<AppHttpClient>().defaultHost;
      final url = '$baseUrl/api/v1/sessions/$sessionId/frames/$frameId/body';
      final response = await http
          .get(Uri.parse(url))
          .timeout(
            const Duration(seconds: 30),
            onTimeout: () {
              throw TimeoutException('Request timed out after 30 seconds');
            },
          );

      if (!mounted) return;

      if (response.statusCode == 200) {
        // Check if backend returned preview fallback (body file not available)
        final bodySource = response.headers['x-body-source'];
        if (bodySource == 'preview') {
          // Backend returned preview - full body not available (body capture disabled)
          if (mounted) {
            setState(() {
              if (isRequest) {
                _reqBodyLoading = false;
                _reqBodyError =
                    'Full body not available (body capture disabled)';
              } else {
                _respBodyLoading = false;
                _respBodyError =
                    'Full body not available (body capture disabled)';
              }
            });
          }
          return;
        }

        final fullBody = utf8.decode(response.bodyBytes, allowMalformed: true);
        setState(() {
          if (isRequest) {
            _fullReqBody = fullBody;
            _reqBodyLoading = false;
            _reqBodyError = null; // Clear error on success
            // Invalidate analysis cache to force re-analysis on full data
            _lastReqBody = null;
            _reqAnalysis = null;
          } else {
            _fullRespBody = fullBody;
            _respBodyLoading = false;
            _respBodyError = null; // Clear error on success
            // Invalidate analysis cache to force re-analysis on full data
            _lastRespBody = null;
            _respAnalysis = null;
          }
        });
      } else if (response.statusCode == 404) {
        // Frame body not found
        if (mounted) {
          setState(() {
            if (isRequest) {
              _reqBodyLoading = false;
              _reqBodyError = 'Frame body not found';
            } else {
              _respBodyLoading = false;
              _respBodyError = 'Frame body not found';
            }
          });
        }
      } else if (response.statusCode == 413) {
        // Body too large
        if (mounted) {
          setState(() {
            if (isRequest) {
              _reqBodyLoading = false;
              _reqBodyError = 'Body too large (exceeds 100MB limit)';
            } else {
              _respBodyLoading = false;
              _respBodyError = 'Body too large (exceeds 100MB limit)';
            }
          });
        }
      } else {
        // Other HTTP errors
        if (mounted) {
          setState(() {
            if (isRequest) {
              _reqBodyLoading = false;
              _reqBodyError = 'Failed to load: HTTP ${response.statusCode}';
            } else {
              _respBodyLoading = false;
              _respBodyError = 'Failed to load: HTTP ${response.statusCode}';
            }
          });
        }
      }
    } catch (e) {
      // Log error for debugging
      debugPrint(
        'Failed to load full ${isRequest ? "request" : "response"} body: $e',
      );

      // Save error and stop loading
      if (mounted) {
        setState(() {
          if (isRequest) {
            _reqBodyLoading = false;
            _reqBodyError = 'Network error: ${e.toString()}';
          } else {
            _respBodyLoading = false;
            _respBodyError = 'Network error: ${e.toString()}';
          }
        });
      }
    }
  }

  void _openBodyFullScreen({
    required String title,
    required String body,
    required ContentAnalysisResult analysis,
    required BodyViewController controller,
    String? contentType,
    String? baseUrl,
    String? frameId,
    int? bodySize,
  }) {
    showDialog<void>(
      context: context,
      builder: (ctx) {
        return BodyFullScreenDialog(
          title: title,
          body: body,
          analysis: analysis,
          initialMode: controller.current,
          getHighlightTheme: _currentHlTheme,
          detectLanguage: _detectLanguage,
          contentType: contentType,
          baseUrl: baseUrl,
          sessionId: widget.sessionId,
          frameId: frameId,
          bodySize: bodySize,
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

  /// Analyze request body with caching to avoid expensive re-analysis on rebuild
  ContentAnalysisResult _analyzeRequestBody(String body, String? contentType) {
    // Cache key based on body length and content type
    // This automatically invalidates when full data loads (different length)
    final cacheKey = '${body.length}:$contentType';
    final lastCacheKey = _lastReqBody != null
        ? '${_lastReqBody!.length}:$_lastReqContentType'
        : null;

    // Analyze if cache is empty OR if cache key changed
    if (_reqAnalysis == null || cacheKey != lastCacheKey) {
      _lastReqBody = body;
      _lastReqContentType = contentType;
      _reqAnalysis = _contentAnalyzer.analyze(body, contentType: contentType);
      _reqViewController.setAvailableModes(_reqAnalysis!.availableModes);
    }
    return _reqAnalysis!;
  }

  /// Analyze response body with caching to avoid expensive re-analysis on rebuild
  ContentAnalysisResult _analyzeResponseBody(String body, String? contentType) {
    // Cache key based on body length and content type
    // This automatically invalidates when full data loads (different length)
    final cacheKey = '${body.length}:$contentType';
    final lastCacheKey = _lastRespBody != null
        ? '${_lastRespBody!.length}:$_lastRespContentType'
        : null;

    // Analyze if cache is empty OR if cache key changed
    if (_respAnalysis == null || cacheKey != lastCacheKey) {
      _lastRespBody = body;
      _lastRespContentType = contentType;
      _respAnalysis = _contentAnalyzer.analyze(body, contentType: contentType);
      _respViewController.setAvailableModes(_respAnalysis!.availableModes);
    }
    return _respAnalysis!;
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
    final reqFrame = _findFrameByType(widget.frames, 'http_request');
    final respFrame = _findFrameByType(widget.frames, 'http_response');
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
            Padding(
              padding: const EdgeInsets.only(left: 10),
              child: Text('Details', style: context.appText.title),
            ),
            if (req != null && req['method'] != null) ...[
              const SizedBox(width: 12),
              HttpMethodChip(
                method: req['method'].toString(),
                style: context.appText.subtitle.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
            if (req != null && req['url'] != null) ...[
              const SizedBox(width: 8),
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: HighlightedUrlWidget(
                    url: req['url'].toString(),
                    baseStyle: context.appText.subtitle,
                    selectable: true,
                  ),
                ),
              ),
              IconButton(
                onPressed: () {
                  final url = req['url']?.toString() ?? '';
                  final uri = _tryParseUri(url);
                  final rawQp =
                      uri?.queryParametersAll ?? <String, List<String>>{};
                  final qp = _normalizeQueryKeys(rawQp);
                  final normalizedUrl = (uri == null)
                      ? null
                      : _rebuildUrlWithQuery(uri, qp);
                  Clipboard.setData(ClipboardData(text: normalizedUrl ?? url));
                },
                icon: const Icon(Icons.copy, size: 18),
                tooltip: 'Copy URL',
                iconSize: 18,
                padding: const EdgeInsets.all(8),
                constraints: const BoxConstraints(),
              ),
              IconButton(
                onPressed: () => _openComposeWithRequest(req),
                icon: const Icon(Icons.edit_outlined, size: 18),
                tooltip: 'Edit in Compose',
                iconSize: 18,
                padding: const EdgeInsets.all(8),
                constraints: const BoxConstraints(),
              ),
            ],
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
        Expanded(
          child: Row(
            children: [
              Expanded(
                child: Builder(
                  builder: (context) {
                    // Determine if request has body
                    final reqBody = (req?['body'] ?? '').toString();
                    final hasReqBody = reqBody.isNotEmpty;

                    // Build tabs list
                    final reqTabs = <Tab>[
                      if (hasReqBody) const Tab(text: 'Body'),
                      const Tab(text: 'Info'),
                    ];

                    // Ensure tab controller has correct length
                    _ensureTabController(
                      controller: _reqTabController,
                      newLength: reqTabs.length,
                      isRequest: true,
                    );

                    return _Card(
                      title: 'Request',
                      actions: req != null
                          ? [_buildRequestActionButtons(context, req)]
                          : null,
                      tabController: _reqTabController,
                      tabs: reqTabs,
                      child: _buildRequest(
                        context,
                        req,
                        reqFrame,
                        hasBodyTab: hasReqBody,
                      ),
                    );
                  },
                ),
              ),
              Expanded(
                child: Builder(
                  builder: (context) {
                    // Determine if response has body
                    final respBody = (resp?['body'] ?? '').toString();
                    final hasRespBody = respBody.isNotEmpty;

                    // Build tabs list
                    final respTabs = <Tab>[
                      if (hasRespBody) const Tab(text: 'Body'),
                      const Tab(text: 'Info'),
                    ];

                    // Ensure tab controller has correct length
                    _ensureTabController(
                      controller: _respTabController,
                      newLength: respTabs.length,
                      isRequest: false,
                    );

                    return _ResponseCard(
                      resp: resp,
                      durationMs: durationMs,
                      respTtfbMs: _respTtfbMs,
                      respTotalMs: _respTotalMs,
                      tabController: _respTabController,
                      tabs: respTabs,
                      child: _buildResponse(
                        context,
                        resp,
                        respFrame,
                        hasBodyTab: hasRespBody,
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  /// Build action buttons for Request card header
  Widget _buildRequestActionButtons(
    BuildContext context,
    Map<String, dynamic> req,
  ) {
    final headers =
        (req['headers'] as Map?)?.map(
          (k, v) => MapEntry(k.toString(), v.toString()),
        ) ??
        <String, String>{};
    final body = _normalizeMaybeQuotedJson((req['body'] ?? '').toString());
    final url = (req['url'] ?? '').toString();
    final uri = _tryParseUri(url);
    final rawQp = uri?.queryParametersAll ?? <String, List<String>>{};
    final qp = _normalizeQueryKeys(rawQp);
    final normalizedUrl = (uri == null) ? null : _rebuildUrlWithQuery(uri, qp);

    return RequestActionButtons(
      url: normalizedUrl ?? url,
      method: req['method']?.toString() ?? 'GET',
      headers: headers,
      body: body,
      form: req['form'] is Map
          ? Map<String, dynamic>.from(req['form'] as Map)
          : null,
      onCopyCurl: (curl) {
        // Curl already copied by the widget, no additional action needed
      },
      onRepeat: url.isNotEmpty
          ? () => _refetchOriginalRequest(context, req)
          : null,
      isLoading: _loadingFetch,
    );
  }

  Widget _buildRequest(
    BuildContext context,
    Map<String, dynamic>? req,
    Map<String, dynamic>? reqFrame, {
    required bool hasBodyTab,
  }) {
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
    final uri = _tryParseUri((req['url'] ?? '').toString());
    final rawQp = uri?.queryParametersAll ?? <String, List<String>>{};
    final qp = _normalizeQueryKeys(rawQp);
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

    // Return TabBarView with conditional Body tab
    return TabBarView(
      controller: _reqTabController,
      children: [
        // Body Tab (only if hasBodyTab is true)
        if (hasBodyTab)
          RequestBodyTab(
            body: body,
            fullBody: _fullReqBody,
            contentType: ctHeader,
            controller: _reqViewController,
            analyzeContent: _analyzeRequestBody,
            buildViewChips: _buildBodyViewChips,
            buildBodyContent: _renderBodyContentAsSliver,
            form: (req['form'] is Map)
                ? (req['form'] as Map).cast<String, dynamic>()
                : null,
            cookies: reqCookies,
            frameId: reqFrame?['frame']?['id']?.toString(),
            bodySize: reqFrame?['frame']?['size'] as int?,
            bodyRawSize: req['bodyRawSize'] as int?,
            isLoading: _reqBodyLoading,
          ),

        // Info Tab (always present)
        RequestInfoTab(
          queryParams: qp,
          headers: headers,
          headersRaw: headersRaw,
        ),
      ],
    );
  }

  Widget _buildResponse(
    BuildContext context,
    Map<String, dynamic>? resp,
    Map<String, dynamic>? respFrame, {
    required bool hasBodyTab,
  }) {
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
    final status = (resp['status'] ?? 0) as int;
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

    // Return TabBarView with conditional Body tab
    return TabBarView(
      controller: _respTabController,
      children: [
        // Body Tab (only if hasBodyTab is true)
        if (hasBodyTab)
          ResponseBodyTab(
            body: body,
            fullBody: _fullRespBody,
            contentType: ctHeader,
            controller: _respViewController,
            analyzeContent: _analyzeResponseBody,
            buildViewChips: _buildBodyViewChips,
            buildBodyContent: _renderBodyContentAsSliver,
            statusCode: status,
            errorMessage: errMessage,
            baseUrl: url,
            frameId: respFrame?['frame']?['id']?.toString(),
            bodySize: respFrame?['frame']?['size'] as int?,
            bodyRawSize: resp['bodyRawSize'] as int?,
            isLoading: _respBodyLoading,
          ),

        // Info Tab (always present)
        ResponseInfoTab(
          headers: headers,
          headersRaw: headersRaw,
          tlsInfo: resp['preview']?['tlsInfo'] as Map<String, dynamic>?,
          cookieSummary:
              resp['preview']?['cookieSummary'] as Map<String, dynamic>?,
          cacheInfo: cache,
          corsInfo: cors,
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

  /// Find full frame (not just preview) by type
  /// Returns map with 'frame' (raw frame) and 'data' (parsed preview)
  Map<String, dynamic>? _findFrameByType(List<dynamic> frames, String type) {
    for (final it in frames) {
      final mp = (it as Map<String, dynamic>);
      final preview = mp['preview']?.toString() ?? '';
      try {
        final obj = jsonDecode(preview) as Map<String, dynamic>;
        if (obj['type'] == type) {
          return {
            'frame': mp, // Full frame with id, size, etc.
            'data': obj, // Parsed preview data
          };
        }
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

  List<String> _splitCsv(String s) {
    if (s.isEmpty) return const [];
    return s
        .split(',')
        .map((e) => e.trim())
        .where((e) => e.isNotEmpty)
        .toList(growable: false);
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

  Future<void> _refetchOriginalRequest(
    BuildContext context,
    Map<String, dynamic>? req,
  ) async {
    if (req == null) return;

    setState(() {
      _loadingFetch = true;
    });

    try {
      final method = (req['method'] ?? 'GET').toString().toUpperCase();
      final url = (req['url'] ?? '').toString();
      final headers =
          (req['headers'] as Map<String, dynamic>?)?.map(
            (k, v) => MapEntry(k.toString(), v.toString()),
          ) ??
          <String, String>{};
      final bodyStr = (req['body'] ?? '').toString();

      if (url.isEmpty) {
        throw Exception('URL is empty');
      }

      final client = sl<AppHttpClient>();

      // Prepare request based on method
      final query = {'_target': url};

      // Try to parse body as JSON Map
      Map<String, dynamic>? bodyMap;
      if (bodyStr.isNotEmpty) {
        try {
          final decoded = jsonDecode(bodyStr);
          if (decoded is Map) {
            bodyMap = decoded.cast<String, dynamic>();
          }
        } catch (_) {
          // If not valid JSON, leave bodyMap as null
        }
      }

      dynamic res;
      switch (method) {
        case 'POST':
          res = await client.post(
            path: '/httpproxy',
            query: query,
            body: bodyMap,
            headers: headers,
          );
          break;
        case 'PUT':
          res = await client.put(
            path: '/httpproxy',
            query: query,
            body: bodyMap,
            headers: headers,
          );
          break;
        case 'DELETE':
          res = await client.delete(
            path: '/httpproxy',
            query: query,
            body: bodyMap,
            headers: headers,
          );
          break;
        case 'PATCH':
          res = await client.patch(
            path: '/httpproxy',
            query: query,
            body: bodyMap,
            headers: headers,
          );
          break;
        default: // GET, HEAD, OPTIONS, etc.
          res = await client.get(
            path: '/httpproxy',
            query: query,
            headers: headers,
          );
      }

      final responseData = res.data?.toString() ?? '';

      // Show result in modal
      if (!mounted) return;

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
                'Repeated Request Response',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 4),
              Text(
                '$method $url',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context).colorScheme.secondary,
                ),
              ),
              const SizedBox(height: 8),
              Expanded(
                child: SingleChildScrollView(
                  child: SelectableText(
                    responseData,
                    style: context.appText.monospace,
                  ),
                ),
              ),
              Align(
                alignment: Alignment.centerRight,
                child: TextButton.icon(
                  onPressed: () {
                    Clipboard.setData(ClipboardData(text: responseData));
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
      if (!mounted) return;
      // ignore: use_build_context_synchronously
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('${msg.title}: ${msg.description}')),
      );
    } finally {
      if (mounted) {
        setState(() {
          _loadingFetch = false;
        });
      }
    }
  }

  /// Build body view mode chips (DRY - reusable for request and response)
  Widget _buildBodyViewChips({
    required String body,
    required BodyViewController controller,
    required ContentAnalysisResult analysis,
    String? title,
    String? contentType,
    String? baseUrl,
    String? frameId,
    int? bodySize,
    int? bodyRawSize,
  }) {
    return Padding(
      padding: const EdgeInsets.only(top: 7),
      child: Wrap(
        spacing: 8,
        runSpacing: 4,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: [
          IconButton(
            icon: const Icon(Icons.fullscreen, size: 24),
            tooltip: 'Full Screen',
            visualDensity: VisualDensity.compact,
            padding: const EdgeInsets.all(8),
            constraints: const BoxConstraints(),
            onPressed: () {
              // Determine if this is request or response based on controller
              final isRequest = controller == _reqViewController;
              // Use full body from cache if available
              final actualBody = isRequest
                  ? (_fullReqBody ?? body)
                  : (_fullRespBody ?? body);

              _openBodyFullScreen(
                title: title ?? 'Body',
                body: actualBody,
                analysis: analysis,
                controller: controller,
                contentType: contentType,
                baseUrl: baseUrl,
                frameId: frameId,
                bodySize: bodySize,
              );
            },
          ),
          ...analysis.availableModes.map((mode) {
            // Determine if data is loading for this controller
            final isLoading =
                (controller == _reqViewController && _reqBodyLoading) ||
                (controller == _respViewController && _respBodyLoading);

            return FilterChip(
              label: Text(
                BodyViewController.getLabel(mode),
                style: Theme.of(context).textTheme.labelSmall,
              ),
              selected: controller.isActive(mode),
              visualDensity: const VisualDensity(horizontal: -2, vertical: -2),
              labelPadding: const EdgeInsets.symmetric(horizontal: 4),
              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
              // Disable chip if data is loading
              onSelected: isLoading
                  ? null
                  : (_) {
                      setState(() {
                        controller.switchTo(mode);
                      });
                    },
            );
          }),
        ],
      ),
    );
  }

  /// Render body content based on current view mode (DRY - reusable)
  /// Render body content as slivers for CustomScrollView integration
  ///
  /// This eliminates nested scroll conflicts by properly integrating body view modes
  /// (especially HexBodyRenderer) as slivers rather than wrapped widgets.
  List<Widget> _renderBodyContentAsSliver({
    required String body,
    required BodyViewController controller,
    required ContentAnalysisResult analysis,
    required String? contentType,
    required bool isRequest,
    String? baseUrl,
    String? frameId,
    int? bodySize,
    int? bodyRawSize,
    JsonSearchController? searchController,
  }) {
    // Use full body from panel-level cache if available, otherwise use preview
    final actualBody = isRequest
        ? (_fullReqBody ?? body)
        : (_fullRespBody ?? body);
    final isLoading = isRequest ? _reqBodyLoading : _respBodyLoading;
    final error = isRequest ? _reqBodyError : _respBodyError;
    final mode = controller.current;

    // PRIORITY 1: Show error if loading failed
    if (error != null) {
      return [
        SliverToBoxAdapter(
          child: Center(
            child: Padding(
              padding: const EdgeInsets.all(32.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.error_outline,
                    size: 48,
                    color: Theme.of(context).colorScheme.error,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Error loading full body',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    error,
                    style: Theme.of(context).textTheme.bodySmall,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton.icon(
                    onPressed: () {
                      // Retry: clear error and trigger reload
                      setState(() {
                        if (isRequest) {
                          _reqBodyError = null;
                        } else {
                          _respBodyError = null;
                        }
                      });
                      // Trigger reload
                      _loadFullBodyIfNeeded(
                        frameId: frameId,
                        preview: body,
                        bodySize: bodySize,
                        isRequest: isRequest,
                      );
                    },
                    icon: const Icon(Icons.refresh),
                    label: const Text('Retry'),
                  ),
                  const SizedBox(height: 8),
                  TextButton(
                    onPressed: () {
                      // Fallback: clear error to show preview
                      setState(() {
                        if (isRequest) {
                          _reqBodyError = null;
                        } else {
                          _respBodyError = null;
                        }
                      });
                    },
                    child: const Text('Show preview instead'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ];
    }

    // PRIORITY 2: Show loading indicator for modes that display full text (not for hex - it has its own loading)
    if (isLoading && mode != BodyViewMode.hex) {
      return [
        const SliverToBoxAdapter(
          child: Center(
            child: Padding(
              padding: EdgeInsets.all(32.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  CircularProgressIndicator(),
                  SizedBox(height: 16),
                  Text('Loading full body data...'),
                ],
              ),
            ),
          ),
        ),
      ];
    }

    // Check if we're displaying truncated preview (not loading, but body is incomplete)
    // BUG FIX: Only show warning if full body hasn't been loaded yet
    final hasFullBody = isRequest
        ? _fullReqBody != null
        : _fullRespBody != null;
    // Use bodyRawSize (actual bytes read) for precise truncation check
    // Body is truncated only if ContentLength > bodyRawSize
    // Calculate displayed preview size for truncation check and message
    final displayedPreviewSize = bodyRawSize ?? utf8.encode(body).length;
    final bool isTruncated =
        !isLoading &&
        !hasFullBody &&
        bodySize != null &&
        bodySize > displayedPreviewSize;

    // Get mode content slivers
    final modeSlivers = _renderModeContentAsSliver(
      mode: mode,
      body: actualBody,
      analysis: analysis,
      contentType: contentType,
      baseUrl: baseUrl,
      frameId: frameId,
      bodySize: bodySize,
      searchController: searchController,
    );

    // Prepend warning banner if showing truncated preview
    if (isTruncated) {
      return [
        SliverToBoxAdapter(
          child: Container(
            width: double.infinity,
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.orange.withValues(alpha: 0.15),
              border: Border(left: BorderSide(color: Colors.orange, width: 4)),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.warning_amber, size: 18, color: Colors.orange),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    '⚠ Displaying truncated preview ($displayedPreviewSize / $bodySize bytes). Loading full data...',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ),
              ],
            ),
          ),
        ),
        const SliverToBoxAdapter(child: SizedBox(height: 8)),
        ...modeSlivers,
      ];
    }

    return modeSlivers;
  }

  /// Render mode-specific content as slivers
  List<Widget> _renderModeContentAsSliver({
    required BodyViewMode mode,
    required String body,
    required ContentAnalysisResult analysis,
    required String? contentType,
    String? baseUrl,
    String? frameId,
    int? bodySize,
    JsonSearchController? searchController,
  }) {
    switch (mode) {
      case BodyViewMode.hex:
        // BUG 2 fix: Use HexBodyRenderer widget for interactivity (clickable bytes)
        // instead of static slivers which are read-only
        return [
          SliverToBoxAdapter(
            child: HexBodyRenderer(
              body: body,
              sessionId: widget.sessionId,
              frameId: frameId,
              bodySize: bodySize,
              apiBaseUrl: baseUrl,
            ),
          ),
        ];

      // Other modes: wrap in SliverToBoxAdapter (they don't have nested scroll issues)
      case BodyViewMode.jwtDecoded:
        return [
          SliverToBoxAdapter(
            child: JwtViewer(token: body, jwtData: analysis.jwtData),
          ),
        ];

      case BodyViewMode.base64Decoded:
        final decoded = analysis.base64Data?.decoded ?? body;
        final isJson = analysis.base64Data?.isJson ?? false;
        if (isJson) {
          return [
            SliverToBoxAdapter(
              child: JsonViewer(
                jsonString: decoded,
                forceTree: false,
                unlimitedHeight: true,
                externalSearchController: searchController,
              ),
            ),
          ];
        }
        return [
          SliverToBoxAdapter(
            child: SelectableText(decoded, style: context.appText.monospace),
          ),
        ];

      case BodyViewMode.jsonTree:
        return [
          SliverToBoxAdapter(
            child: JsonViewer(
              jsonString: body,
              forceTree: true,
              unlimitedHeight: true,
              externalSearchController: searchController,
            ),
          ),
        ];

      case BodyViewMode.pretty:
        if (analysis.isJson) {
          return [
            SliverToBoxAdapter(
              child: JsonViewer(
                jsonString: body,
                forceTree: false,
                unlimitedHeight: true,
                externalSearchController: searchController,
              ),
            ),
          ];
        }
        // Auto-format GraphQL queries in Pretty mode
        if (GraphqlLanguageDetector.isLikelyGraphql(body)) {
          try {
            final formatted = GraphQLFormatter.format(body);
            return [
              SliverToBoxAdapter(
                child: _highlightBlock(formatted, contentType: contentType),
              ),
            ];
          } catch (e) {
            // Fallback to raw if formatting fails
            return [
              SliverToBoxAdapter(
                child: _highlightBlock(body, contentType: contentType),
              ),
            ];
          }
        }
        return [
          SliverToBoxAdapter(
            child: _highlightBlock(body, contentType: contentType),
          ),
        ];

      case BodyViewMode.htmlPreview:
        return [
          SliverToBoxAdapter(
            child: SizedBox(
              height: 480,
              child: HtmlPreview(html: body, baseUrl: baseUrl),
            ),
          ),
        ];

      case BodyViewMode.raw:
        return [
          SliverToBoxAdapter(
            child: SelectableText(body, style: context.appText.monospace),
          ),
        ];
    }
  }
}

class _Card extends StatelessWidget {
  const _Card({
    required this.title,
    required this.child,
    this.actions,
    this.tabController,
    this.tabs,
  });
  final String title;
  final Widget child;
  final List<Widget>? actions;
  final TabController? tabController;
  final List<Tab>? tabs;

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
            Row(
              children: [
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (actions != null) ...[const Spacer(), ...actions!],
              ],
            ),
            if (tabs != null && tabController != null) ...[
              const SizedBox(height: 2),
              SizedBox(
                height: 32,
                child: TabBar(
                  controller: tabController,
                  tabs: tabs!,
                  labelColor: Theme.of(context).colorScheme.primary,
                  indicatorSize: TabBarIndicatorSize.label,
                  labelPadding: const EdgeInsets.symmetric(
                    horizontal: 14,
                    vertical: 4,
                  ),
                  labelStyle: Theme.of(
                    context,
                  ).textTheme.labelSmall?.copyWith(fontSize: 13, height: 1.0),
                  unselectedLabelStyle: Theme.of(
                    context,
                  ).textTheme.labelSmall?.copyWith(fontSize: 13, height: 1.0),
                  isScrollable: true,
                  tabAlignment: TabAlignment.start,
                ),
              ),
            ],
            const SizedBox(height: 2),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}

class _ResponseCard extends StatelessWidget {
  const _ResponseCard({
    required this.resp,
    required this.durationMs,
    required this.respTtfbMs,
    required this.respTotalMs,
    required this.child,
    this.tabController,
    this.tabs,
  });
  final Map<String, dynamic>? resp;
  final int? durationMs;
  final int? respTtfbMs;
  final int? respTotalMs;
  final Widget child;
  final TabController? tabController;
  final List<Tab>? tabs;

  @override
  Widget build(BuildContext context) {
    final status = resp?['status'] as int? ?? 0;
    final color = _statusColor(context, status);
    final totalTime = respTotalMs ?? durationMs;

    return Card(
      elevation: 1,
      margin: const EdgeInsets.all(8),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Text(
                  'Response',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const Spacer(),
                if (totalTime != null) ...[
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: _durationBg(context, totalTime),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      'Total: $totalTime ms',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: _durationFg(context, totalTime),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                ],
                Text(
                  'Status: $status',
                  style: Theme.of(
                    context,
                  ).textTheme.bodyMedium?.copyWith(color: color),
                ),
              ],
            ),
            if (tabs != null && tabController != null) ...[
              const SizedBox(height: 2),
              SizedBox(
                height: 32,
                child: TabBar(
                  controller: tabController,
                  tabs: tabs!,
                  labelColor: Theme.of(context).colorScheme.primary,
                  indicatorSize: TabBarIndicatorSize.label,
                  labelPadding: const EdgeInsets.symmetric(
                    horizontal: 14,
                    vertical: 4,
                  ),
                  labelStyle: Theme.of(
                    context,
                  ).textTheme.labelSmall?.copyWith(fontSize: 13, height: 1.0),
                  unselectedLabelStyle: Theme.of(
                    context,
                  ).textTheme.labelSmall?.copyWith(fontSize: 13, height: 1.0),
                  isScrollable: true,
                  tabAlignment: TabAlignment.start,
                ),
              ),
            ],
            const SizedBox(height: 2),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }

  Color _statusColor(BuildContext context, int status) {
    final cs = Theme.of(context).colorScheme;
    if (status >= 500) return cs.error;
    if (status >= 400) return cs.tertiary;
    if (status >= 300) return cs.primary;
    return Colors.green;
  }

  Color _durationBg(BuildContext context, int ms) {
    final cs = Theme.of(context).colorScheme;
    if (ms < 300) return Colors.green.withValues(alpha: 0.12);
    if (ms < 1000) return cs.tertiary.withValues(alpha: 0.12);
    return cs.error.withValues(alpha: 0.12);
  }

  Color _durationFg(BuildContext context, int ms) {
    final cs = Theme.of(context).colorScheme;
    if (ms < 300) return Colors.green;
    if (ms < 1000) return cs.tertiary;
    return cs.error;
  }
}
