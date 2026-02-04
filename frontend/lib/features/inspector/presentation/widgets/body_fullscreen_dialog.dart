import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:highlight_selectable/highlight_selectable.dart';
import 'package:app_http_client/application/app_http_client.dart';
import '../utils/body_view_mode.dart';
import '../utils/body_content_analyzer.dart';
import '../utils/graphql_detect.dart';
import '../utils/graphql_formatter.dart';
import '../../../../widgets/json_viewer.dart';
import '../../../../widgets/html_preview.dart';
import '../../../../widgets/json_search_controller.dart';
import '../../../../widgets/json_search_buttons.dart';
import '../../../../theme/context_ext.dart';
import '../../../../core/di/di.dart';
import 'jwt_viewer.dart';
import 'hex_body_renderer.dart';

class BodyFullScreenDialog extends StatefulWidget {
  const BodyFullScreenDialog({
    super.key,
    required this.title,
    required this.body,
    required this.analysis,
    required this.initialMode,
    required this.getHighlightTheme,
    required this.detectLanguage,
    this.contentType,
    this.baseUrl,
    this.sessionId,
    this.frameId,
    this.bodySize,
  });

  final String title;
  final String body;
  final ContentAnalysisResult analysis;
  final BodyViewMode initialMode;
  final Map<String, TextStyle> Function() getHighlightTheme;
  final String Function(String body, String? contentType) detectLanguage;
  final String? contentType;
  final String? baseUrl;
  final String? sessionId;
  final String? frameId;
  final int? bodySize;

  @override
  State<BodyFullScreenDialog> createState() => _BodyFullScreenDialogState();
}

class _BodyFullScreenDialogState extends State<BodyFullScreenDialog> {
  late final BodyViewController _viewController;
  late final JsonSearchController _searchController;

  @override
  void initState() {
    super.initState();
    _viewController = BodyViewController();
    _viewController.setAvailableModes(widget.analysis.availableModes);
    _viewController.switchTo(widget.initialMode);
    _searchController = JsonSearchController();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog.fullscreen(
      child: Scaffold(
        appBar: AppBar(
          title: Text(widget.title),
          leading: IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => Navigator.of(context).pop(),
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.copy),
              tooltip: 'Copy',
              onPressed: () {
                Clipboard.setData(ClipboardData(text: widget.body));
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Copied to clipboard'),
                    duration: Duration(seconds: 1),
                  ),
                );
              },
            ),
          ],
        ),
        body: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // FilterChips (modes: Pretty/Tree/Raw/etc)
            Padding(
              padding: const EdgeInsets.all(16.0),
              child: Wrap(
                spacing: 8,
                runSpacing: 4,
                crossAxisAlignment: WrapCrossAlignment.center,
                children: [
                  ...widget.analysis.availableModes.map((mode) {
                    return FilterChip(
                      label: Text(
                        BodyViewController.getLabel(mode),
                        style: Theme.of(context).textTheme.labelSmall,
                      ),
                      selected: _viewController.isActive(mode),
                      visualDensity: const VisualDensity(
                        horizontal: -2,
                        vertical: -2,
                      ),
                      labelPadding: const EdgeInsets.symmetric(horizontal: 4),
                      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                      onSelected: (_) {
                        setState(() {
                          _viewController.switchTo(mode);
                        });
                      },
                    );
                  }),
                ],
              ),
            ),
            // Stack with content and fixed buttons on top
            Expanded(
              child: Stack(
                children: [
                  // Content with increased left padding
                  // Hex mode doesn't use ScrollView since HexViewer has its own scrolling
                  if (_viewController.current != BodyViewMode.hex)
                    SingleChildScrollView(
                      padding: const EdgeInsets.only(
                        left: 48.0,
                        right: 16.0,
                        bottom: 16.0,
                        top: 16.0,
                      ),
                      child: _renderBodyContent(),
                    )
                  else
                    Padding(
                      padding: const EdgeInsets.only(
                        left: 48.0,
                        right: 16.0,
                        bottom: 16.0,
                        top: 16.0,
                      ),
                      child: _renderBodyContent(),
                    ),
                  // Fixed search/copy buttons (JSON only)
                  if (_isJsonMode())
                    Positioned(
                      top: 8,
                      right: 8,
                      child: ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 400),
                        child: JsonSearchButtons(
                          controller: _searchController,
                          jsonString: widget.body,
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Check if current mode displays JSON (needs search buttons)
  bool _isJsonMode() {
    final mode = _viewController.current;
    return mode == BodyViewMode.jsonTree ||
        mode == BodyViewMode.pretty && widget.analysis.isJson ||
        mode == BodyViewMode.base64Decoded &&
            (widget.analysis.base64Data?.isJson ?? false);
  }

  Widget _renderBodyContent() {
    final mode = _viewController.current;

    switch (mode) {
      case BodyViewMode.jwtDecoded:
        return JwtViewer(token: widget.body, jwtData: widget.analysis.jwtData);

      case BodyViewMode.base64Decoded:
        final decoded = widget.analysis.base64Data?.decoded ?? widget.body;
        final isJson = widget.analysis.base64Data?.isJson ?? false;
        if (isJson) {
          return JsonViewer(
            jsonString: decoded,
            forceTree: false,
            unlimitedHeight: true,
            externalSearchController: _searchController,
          );
        }
        return SelectableText(decoded, style: context.appText.monospace);

      case BodyViewMode.jsonTree:
        return JsonViewer(
          jsonString: widget.body,
          forceTree: true,
          unlimitedHeight: true,
          externalSearchController: _searchController,
        );

      case BodyViewMode.pretty:
        if (widget.analysis.isJson) {
          return JsonViewer(
            jsonString: widget.body,
            forceTree: false,
            unlimitedHeight: true,
            externalSearchController: _searchController,
          );
        }
        // Auto-format GraphQL queries in Pretty mode
        if (GraphqlLanguageDetector.isLikelyGraphql(widget.body)) {
          try {
            final formatted = GraphQLFormatter.format(widget.body);
            return _highlightBlock(formatted, contentType: widget.contentType);
          } catch (e) {
            // Fallback to raw if formatting fails
            return _highlightBlock(
              widget.body,
              contentType: widget.contentType,
            );
          }
        }
        return _highlightBlock(widget.body, contentType: widget.contentType);

      case BodyViewMode.htmlPreview:
        return HtmlPreview(html: widget.body, baseUrl: widget.baseUrl);

      case BodyViewMode.raw:
        return SelectableText(widget.body, style: context.appText.monospace);

      case BodyViewMode.hex:
        return HexBodyRenderer(
          body: widget.body,
          sessionId: widget.sessionId,
          frameId: widget.frameId,
          bodySize: widget.bodySize,
          apiBaseUrl: sl<AppHttpClient>().defaultHost,
        );
    }
  }

  Widget _highlightBlock(String text, {String? contentType}) {
    if (text.isEmpty) return const SizedBox.shrink();

    final language = widget.detectLanguage(text, contentType);
    final theme = widget.getHighlightTheme();
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
          text,
          language: language,
          theme: theme,
          selectable: true,
          showCopyButton: true,
          showEditButton: false,
        ),
      ),
    );
  }
}
