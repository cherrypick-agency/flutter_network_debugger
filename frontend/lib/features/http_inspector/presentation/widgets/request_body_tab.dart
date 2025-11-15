import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../inspector/presentation/utils/body_view_mode.dart';
import '../../../inspector/presentation/utils/body_content_analyzer.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_search_controller.dart';
import '../../../../widgets/json_search_buttons.dart';
import 'form_data_view.dart';

/// Callback typedef for body rendering
typedef BodyViewChipsBuilder =
    Widget Function({
      required String body,
      required BodyViewController controller,
      required ContentAnalysisResult analysis,
      String? title,
      String? contentType,
      String? frameId,
      int? bodySize,
    });

/// Callback typedef for body content rendering as slivers
typedef BodyContentSliverBuilder =
    List<Widget> Function({
      required String body,
      required BodyViewController controller,
      required ContentAnalysisResult analysis,
      required String? contentType,
      required bool isRequest,
      String? frameId,
      int? bodySize,
      JsonSearchController? searchController,
    });

/// Callback typedef for content analysis
typedef ContentAnalyzer =
    ContentAnalysisResult Function(String body, String? contentType);

/// Request body tab widget
///
/// Displays:
/// 1. Form Data section (if multipart/urlencoded)
/// 2. Body section with view mode chips (Pretty, Raw, Tree, Hex, etc.)
/// 3. Cookies section (if present)
/// 4. Fixed search/copy buttons overlay for JSON content
///
/// Follows SRP - only responsible for layout and delegation to specialized builders.
/// Doesn't contain business logic for body analysis or rendering.
class RequestBodyTab extends StatefulWidget {
  const RequestBodyTab({
    super.key,
    required this.body,
    required this.contentType,
    required this.controller,
    required this.analyzeContent,
    required this.buildViewChips,
    required this.buildBodyContent,
    this.form,
    this.cookies,
    this.frameId,
    this.bodySize,
    this.fullBody,
    this.isLoading = false,
  });

  /// Raw body string (may be truncated preview)
  final String body;

  /// Full body string (loaded from API if needed)
  final String? fullBody;

  /// Content-Type header value
  final String contentType;

  /// Body view controller for managing current view mode
  final BodyViewController controller;

  /// Form data (for multipart/urlencoded requests)
  /// Structure: {type: "multipart", fields: [...], files: [...]}
  final Map<String, dynamic>? form;

  /// Cookies from Cookie header
  /// Map of cookie name -> value
  final Map<String, String>? cookies;

  /// Frame ID for loading full body data
  final String? frameId;

  /// Total body size (for truncation detection)
  final int? bodySize;

  /// Whether full body is currently loading
  final bool isLoading;

  /// Callback to analyze body content
  final ContentAnalyzer analyzeContent;

  /// Callback to build view mode chips
  final BodyViewChipsBuilder buildViewChips;

  /// Callback to build body content as slivers
  final BodyContentSliverBuilder buildBodyContent;

  @override
  State<RequestBodyTab> createState() => _RequestBodyTabState();
}

class _RequestBodyTabState extends State<RequestBodyTab> {
  late final JsonSearchController _searchController;

  @override
  void initState() {
    super.initState();
    _searchController = JsonSearchController();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Determine if we should hide raw body (when form data is present)
    final ctLower = widget.contentType.toLowerCase();
    final hasFormObj = widget.form != null && widget.form!.isNotEmpty;
    final isFormCt =
        ctLower.contains('multipart/form-data') ||
        ctLower.contains('application/x-www-form-urlencoded');
    final hideRawBody = hasFormObj || isFormCt;

    // Check if we're showing JSON content (need search buttons)
    final bodyForAnalysis = widget.fullBody ?? widget.body;
    final analysis = widget.analyzeContent(bodyForAnalysis, widget.contentType);
    final isJsonContent =
        analysis.isJson &&
        (widget.controller.current == BodyViewMode.pretty ||
            widget.controller.current == BodyViewMode.jsonTree);

    return Stack(
      children: [
        CustomScrollView(
          slivers: [
            // Form Data section (if present)
            if (hasFormObj || isFormCt)
              SliverToBoxAdapter(
                child: FormDataView(
                  form: widget.form,
                  contentType: widget.contentType,
                  rawBody: widget.body,
                ),
              ),
            const SliverToBoxAdapter(child: SizedBox(height: 6)),

            // Body section (if not hidden by form data)
            if (widget.body.isNotEmpty && !hideRawBody)
              ..._buildBodySection(context),

            const SliverToBoxAdapter(child: SizedBox(height: 8)),

            // Cookies section (if present)
            if (widget.cookies != null && widget.cookies!.isNotEmpty)
              ..._buildCookiesSection(context),
          ],
        ),
        // Fixed search/copy buttons overlay (only for JSON content)
        if (isJsonContent)
          Positioned(
            top: 8,
            right: 8,
            child: JsonSearchButtons(
              controller: _searchController,
              jsonString: bodyForAnalysis,
            ),
          ),
      ],
    );
  }

  /// Build body section with view mode chips and content
  List<Widget> _buildBodySection(BuildContext context) {
    // Use full body for analysis if available, otherwise use preview
    final bodyForAnalysis = widget.fullBody ?? widget.body;
    final analysis = widget.analyzeContent(bodyForAnalysis, widget.contentType);

    return [
      // View mode chips (scrollable, not pinned)
      SliverToBoxAdapter(
        child: Padding(
          padding: const EdgeInsets.only(bottom: 6),
          child: widget.buildViewChips(
            body: widget.body,
            controller: widget.controller,
            analysis: analysis,
            title: 'Request Body',
            contentType: widget.contentType,
            frameId: widget.frameId,
            bodySize: widget.bodySize,
          ),
        ),
      ),
      const SliverToBoxAdapter(child: SizedBox(height: 6)),

      // Body content as slivers (eliminates nested scroll issues)
      ...widget.buildBodyContent(
        body: widget.body,
        controller: widget.controller,
        analysis: analysis,
        contentType: widget.contentType,
        isRequest: true,
        frameId: widget.frameId,
        bodySize: widget.bodySize,
        searchController: _searchController,
      ),
    ];
  }

  /// Build cookies section
  List<Widget> _buildCookiesSection(BuildContext context) {
    return [
      SliverToBoxAdapter(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Cookies', style: context.appText.subtitle),
            const SizedBox(height: 4),
            ...widget.cookies!.entries.map(
              (e) => Padding(
                padding: const EdgeInsets.only(bottom: 2),
                child: _CookieItem(name: e.key, value: e.value),
              ),
            ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    ];
  }
}

/// Cookie item widget with copy functionality
class _CookieItem extends StatefulWidget {
  const _CookieItem({required this.name, required this.value});

  final String name;
  final String value;

  @override
  State<_CookieItem> createState() => _CookieItemState();
}

class _CookieItemState extends State<_CookieItem> {
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

    return SelectableText.rich(
      TextSpan(
        children: [
          TextSpan(text: '${widget.name}: ', style: nameStyle),
          TextSpan(text: widget.value, style: valueStyle),
        ],
      ),
    );
  }
}
