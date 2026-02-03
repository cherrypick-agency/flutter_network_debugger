import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// cURL command export mode
enum CurlExportMode {
  /// Compact single-line format
  compact,

  /// Multiline format with backslash
  multiline,

  /// Multiline with additional options (--location, --insecure)
  withOptions,
}

/// cURL copy button with dropdown menu for formats.
/// Reused in Inspector and Compose.
class CopyCurlButton extends StatelessWidget {
  const CopyCurlButton({
    super.key,
    required this.url,
    required this.method,
    required this.headers,
    this.body = '',
    this.form,
    this.onCopied,
    this.showSnackBar = true,
  });

  /// Request URL
  final String url;

  /// HTTP method (GET, POST, etc.)
  final String method;

  /// Request headers
  final Map<String, String> headers;

  /// Request body (raw string)
  final String body;

  /// Form data (for multipart/urlencoded)
  final Map<String, dynamic>? form;

  /// Callback after copying (optional)
  final void Function(String curl)? onCopied;

  /// Show SnackBar after copying
  final bool showSnackBar;

  @override
  Widget build(BuildContext context) {
    return MenuAnchor(
      builder: (context, controller, child) {
        return OutlinedButton(
          onPressed: () {
            if (controller.isOpen) {
              controller.close();
            } else {
              controller.open();
            }
          },
          style: OutlinedButton.styleFrom(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
            textStyle: const TextStyle(fontSize: 12),
            minimumSize: Size.zero,
            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.content_paste, size: 14),
              const SizedBox(width: 4),
              const Text('cURL', style: TextStyle(fontSize: 12)),
              const SizedBox(width: 2),
              Icon(
                controller.isOpen ? Icons.arrow_drop_up : Icons.arrow_drop_down,
                size: 16,
              ),
            ],
          ),
        );
      },
      menuChildren: [
        MenuItemButton(
          leadingIcon: const Icon(Icons.remove, size: 16),
          onPressed: () => _copyWithMode(context, CurlExportMode.compact),
          child: const Text('Compact'),
        ),
        MenuItemButton(
          leadingIcon: const Icon(Icons.wrap_text, size: 16),
          onPressed: () => _copyWithMode(context, CurlExportMode.multiline),
          child: const Text('Multiline'),
        ),
        MenuItemButton(
          leadingIcon: const Icon(Icons.settings, size: 16),
          onPressed: () => _copyWithMode(context, CurlExportMode.withOptions),
          child: const Text('With options'),
        ),
      ],
    );
  }

  void _copyWithMode(BuildContext context, CurlExportMode mode) {
    final curl = buildCurl(
      url: url,
      method: method,
      headers: headers,
      body: body,
      form: form,
      mode: mode,
    );
    Clipboard.setData(ClipboardData(text: curl));
    onCopied?.call(curl);

    if (showSnackBar) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('cURL copied')));
    }
  }

  /// Generate cURL command from request data.
  /// Static method for use without widget.
  static String buildCurl({
    required String url,
    required String method,
    required Map<String, String> headers,
    String body = '',
    Map<String, dynamic>? form,
    CurlExportMode mode = CurlExportMode.compact,
  }) {
    if (url.isEmpty) {
      return '# Error: URL is empty';
    }

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

    // Check if form-data format is needed
    final contentType =
        headers['Content-Type'] ?? headers['content-type'] ?? '';
    final useFormData =
        form != null &&
        (contentType.contains('multipart/form-data') ||
            contentType.contains('application/x-www-form-urlencoded'));

    final buffer = StringBuffer();
    final isMultiline =
        mode == CurlExportMode.multiline || mode == CurlExportMode.withOptions;
    final newline = isMultiline ? ' \\\n  ' : ' ';

    buffer.write("curl -X $method$newline'");
    buffer.write(_escapeShellArg(url));
    buffer.write("'");

    // Add headers
    headersToInclude.forEach((k, v) {
      // Skip Content-Type and Content-Length for form-data
      if (useFormData &&
          (k.toLowerCase() == 'content-type' ||
              k.toLowerCase() == 'content-length')) {
        return;
      }
      buffer.write("$newline-H '$k: ${_escapeShellArg(v)}'");
    });

    // Add cookie
    if ((cookieValue ?? '').isNotEmpty) {
      buffer.write("$newline--cookie '${_escapeShellArg(cookieValue!)}'");
    }

    // Add body or form data
    if (useFormData) {
      _appendFormData(buffer, newline, form);
    } else if (body.isNotEmpty) {
      buffer.write("$newline--data '${_escapeShellArg(body)}'");
    }

    // Additional options for withOptions mode
    if (mode == CurlExportMode.withOptions) {
      buffer.write("$newline--location");
      buffer.write("$newline--insecure");
    }

    buffer.write("$newline--compressed");

    return buffer.toString();
  }

  static void _appendFormData(
    StringBuffer buffer,
    String newline,
    Map<String, dynamic>? form,
  ) {
    final fields =
        (form?['fields'] as List?)?.cast<Map<String, dynamic>>() ?? [];
    final files = (form?['files'] as List?)?.cast<Map<String, dynamic>>() ?? [];

    for (final field in fields) {
      final name = field['name']?.toString() ?? '';
      final value =
          field['valuePreview']?.toString() ?? field['value']?.toString() ?? '';
      if (name.isNotEmpty) {
        buffer.write("$newline-F '$name=${_escapeShellArg(value)}'");
      }
    }

    if (files.isNotEmpty) {
      buffer.write(
        "$newline# Note: Replace file paths with actual local paths",
      );
    }
    for (final file in files) {
      final name = file['name']?.toString() ?? '';
      final filename = file['filename']?.toString() ?? 'file';
      if (name.isNotEmpty) {
        buffer.write("$newline-F '$name=@$filename'");
      }
    }
  }

  static String _escapeShellArg(String arg) {
    return arg.replaceAll("'", "'\\''");
  }
}
