import 'package:flutter/material.dart';
import '../../../inspector/presentation/utils/body_view_mode.dart';
import '../../../inspector/presentation/utils/body_content_analyzer.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_search_controller.dart';
import '../../../../widgets/json_search_buttons.dart';

/// Callback typedef for body rendering (reused from request_body_tab)
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
    });

/// Callback typedef for body content rendering as slivers
typedef BodyContentSliverBuilder =
    List<Widget> Function({
      required String body,
      required BodyViewController controller,
      required ContentAnalysisResult analysis,
      required String? contentType,
      required bool isRequest,
      String? baseUrl,
      String? frameId,
      int? bodySize,
      JsonSearchController? searchController,
    });

/// Callback typedef for content analysis
typedef ContentAnalyzer =
    ContentAnalysisResult Function(String body, String? contentType);

/// Response body tab widget
///
/// Displays:
/// 1. Error banner (if 4xx/5xx status)
/// 2. Body section with view mode chips (Pretty, Raw, Tree, Hex, etc.)
/// 3. Fixed search/copy buttons overlay for JSON content
///
/// Follows SRP - only responsible for layout and delegation to specialized builders.
class ResponseBodyTab extends StatefulWidget {
  const ResponseBodyTab({
    super.key,
    required this.body,
    required this.contentType,
    required this.controller,
    required this.analyzeContent,
    required this.buildViewChips,
    required this.buildBodyContent,
    this.statusCode,
    this.errorMessage,
    this.baseUrl,
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

  /// HTTP status code (for error detection)
  final int? statusCode;

  /// Error message (if status is 0 - transport error)
  final String? errorMessage;

  /// Base URL for HTML preview
  final String? baseUrl;

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
  State<ResponseBodyTab> createState() => _ResponseBodyTabState();
}

class _ResponseBodyTabState extends State<ResponseBodyTab> {
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

  /// Check if this is an error response (4xx or 5xx)
  bool get _isErrorStatus {
    final status = widget.statusCode ?? 0;
    return status >= 400 && status < 600;
  }

  @override
  Widget build(BuildContext context) {
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
            // Error banner for 4xx/5xx responses
            if (_isErrorStatus) _buildErrorBanner(context),

            // Transport error banner (status == 0)
            if (widget.statusCode == 0 &&
                widget.errorMessage != null &&
                widget.errorMessage!.isNotEmpty)
              _buildTransportErrorBanner(context),

            // Body section
            if (widget.body.isNotEmpty) ..._buildBodySection(context),
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

  /// Build error banner for HTTP errors (4xx/5xx)
  Widget _buildErrorBanner(BuildContext context) {
    return SliverToBoxAdapter(
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(12),
        margin: const EdgeInsets.only(bottom: 8),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.error.withValues(alpha: 0.15),
          border: Border(
            left: BorderSide(
              color: Theme.of(context).colorScheme.error,
              width: 4,
            ),
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.error_outline,
              size: 18,
              color: Theme.of(context).colorScheme.error,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                'Error ${widget.statusCode} - ${_getErrorDescription(widget.statusCode ?? 0)}',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context).colorScheme.error,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Build transport error banner (network error, timeout, etc.)
  Widget _buildTransportErrorBanner(BuildContext context) {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Text(
          'Transport Error: ${widget.errorMessage}',
          style: Theme.of(context).textTheme.labelSmall?.copyWith(
            color: Theme.of(context).colorScheme.error,
          ),
        ),
      ),
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
            title: 'Response Body',
            contentType: widget.contentType,
            baseUrl: widget.baseUrl,
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
        isRequest: false,
        baseUrl: widget.baseUrl,
        frameId: widget.frameId,
        bodySize: widget.bodySize,
        searchController: _searchController,
      ),
    ];
  }

  /// Get human-readable error description
  String _getErrorDescription(int status) {
    if (status >= 500) return 'Server Error';
    if (status == 404) return 'Not Found';
    if (status == 403) return 'Forbidden';
    if (status == 401) return 'Unauthorized';
    if (status == 400) return 'Bad Request';
    if (status >= 400) return 'Client Error';
    return 'Error';
  }
}
