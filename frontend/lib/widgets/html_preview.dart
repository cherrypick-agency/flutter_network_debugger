import 'package:flutter/material.dart';
import 'package:webview_flutter/webview_flutter.dart';

/// Простой превьювер HTML через WebView
/// Поддерживает загрузку HTML-строки с опциональным baseUrl для относительных ссылок
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
    _controller =
        WebViewController()
          ..setJavaScriptMode(JavaScriptMode.unrestricted)
          ..loadHtmlString(widget.html, baseUrl: widget.baseUrl);
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
