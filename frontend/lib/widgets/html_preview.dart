import 'package:flutter/material.dart';
import 'package:webview_flutter/webview_flutter.dart';

/// Simple HTML previewer via WebView
/// Supports loading HTML string with optional baseUrl for relative links
class HtmlPreview extends StatefulWidget {
  const HtmlPreview({super.key, required this.html, this.baseUrl});

  final String html;
  final String? baseUrl;

  @override
  State<HtmlPreview> createState() => _HtmlPreviewState();
}

class _HtmlPreviewState extends State<HtmlPreview> {
  late final WebViewController _controller;

  @override
  void initState() {
    super.initState();
    _controller = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..loadHtmlString(widget.html, baseUrl: widget.baseUrl);
  }

  @override
  void didUpdateWidget(HtmlPreview oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Reload HTML if content or baseUrl changed
    if (oldWidget.html != widget.html || oldWidget.baseUrl != widget.baseUrl) {
      _controller.loadHtmlString(widget.html, baseUrl: widget.baseUrl);
    }
  }

  @override
  void dispose() {
    // WebViewController doesn't require explicit disposal in webview_flutter 4.x
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: DecoratedBox(
        decoration: BoxDecoration(color: Theme.of(context).colorScheme.surface),
        child: WebViewWidget(controller: _controller),
      ),
    );
  }
}
