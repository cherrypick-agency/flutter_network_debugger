import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../types/body_callbacks.dart';

/// Export mode for cURL command generation
enum CurlExportMode {
  /// Compact single-line format
  compact,

  /// Multi-line format with backslashes
  multiline,

  /// Multi-line format with additional options (--location, --insecure)
  withOptions,
}

/// Action buttons for HTTP request operations
///
/// Provides two main actions:
/// 1. Copy cURL command (with dropdown menu for different formats)
/// 2. Repeat request
///
/// Follows SRP (Single Responsibility Principle) - only handles action buttons,
/// doesn't contain business logic for curl generation or request execution.
class RequestActionButtons extends StatelessWidget {
  const RequestActionButtons({
    super.key,
    required this.url,
    required this.method,
    required this.headers,
    required this.body,
    this.form,
    required this.onCopyCurl,
    this.onRepeat,
    this.isLoading = false,
  });

  /// The HTTP request URL
  final String url;

  /// HTTP method (GET, POST, etc.)
  final String method;

  /// Request headers
  final Map<String, String> headers;

  /// Request body (raw string)
  final String body;

  /// Form data (for multipart/urlencoded requests)
  final Map<String, dynamic>? form;

  /// Callback when Copy cURL button is pressed (compact mode by default)
  /// Receives the generated cURL command
  final CurlCallback onCopyCurl;

  /// Optional callback when Repeat button is pressed
  /// If null, button will be disabled
  final VoidCallback? onRepeat;

  /// Whether the Repeat operation is currently loading
  final bool isLoading;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 8,
      runSpacing: 4,
      children: [
        _buildCopyCurlButton(context),
        if (url.isNotEmpty) _buildRepeatButton(),
      ],
    );
  }

  /// Build Copy cURL button with dropdown menu
  Widget _buildCopyCurlButton(BuildContext context) {
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
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 12),
            textStyle: const TextStyle(fontSize: 12),
            minimumSize: Size.zero,
            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.content_paste, size: 14),
              const SizedBox(width: 4),
              const Text('Copy as cURL', style: TextStyle(fontSize: 12)),
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
          onPressed: () {
            final curl = _buildCurl(CurlExportMode.compact);
            onCopyCurl(curl);
            Clipboard.setData(ClipboardData(text: curl));
          },
          child: const Text('Compact'),
        ),
        MenuItemButton(
          leadingIcon: const Icon(Icons.wrap_text, size: 16),
          onPressed: () {
            final curl = _buildCurl(CurlExportMode.multiline);
            onCopyCurl(curl);
            Clipboard.setData(ClipboardData(text: curl));
          },
          child: const Text('Multiline'),
        ),
        MenuItemButton(
          leadingIcon: const Icon(Icons.settings, size: 16),
          onPressed: () {
            final curl = _buildCurl(CurlExportMode.withOptions);
            onCopyCurl(curl);
            Clipboard.setData(ClipboardData(text: curl));
          },
          child: const Text('With options'),
        ),
      ],
    );
  }

  /// Build Repeat button
  Widget _buildRepeatButton() {
    if (isLoading) {
      return const SizedBox(
        width: 16,
        height: 16,
        child: CircularProgressIndicator(strokeWidth: 1.5),
      );
    }

    return TextButton.icon(
      onPressed: onRepeat,
      style: TextButton.styleFrom(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 12),
        textStyle: const TextStyle(fontSize: 12),
        minimumSize: Size.zero,
        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
      ),
      icon: const Icon(Icons.refresh, size: 14),
      label: const Text('Repeat', style: TextStyle(fontSize: 12)),
    );
  }

  /// Generate cURL command from request data
  ///
  /// This is a pure function that doesn't depend on widget state,
  /// making it easy to test and maintain.
  String _buildCurl(CurlExportMode mode) {
    // Validate URL
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

    // Determine if we should use form-data format
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

    // Start command
    buffer.write("curl -X $method$newline'");
    buffer.write(_escapeShellArg(url));
    buffer.write("'");

    // Add headers
    headersToInclude.forEach((k, v) {
      // Skip Content-Type and Content-Length for form-data (curl adds them automatically)
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
      _appendFormData(buffer, newline);
    } else if (body.isNotEmpty) {
      // Use --data for regular body
      buffer.write("$newline--data '${_escapeShellArg(body)}'");
    }

    // Add additional options for withOptions mode
    if (mode == CurlExportMode.withOptions) {
      buffer.write("$newline--location");
      buffer.write("$newline--insecure");
    }

    buffer.write("$newline--compressed");

    return buffer.toString();
  }

  /// Append form data to cURL command
  void _appendFormData(StringBuffer buffer, String newline) {
    // Form structure: {type: "multipart", fields: [...], files: [...]}
    final fields =
        (form?['fields'] as List?)?.cast<Map<String, dynamic>>() ?? [];
    final files = (form?['files'] as List?)?.cast<Map<String, dynamic>>() ?? [];

    // Add regular fields
    for (final field in fields) {
      final name = field['name']?.toString() ?? '';
      final value =
          field['valuePreview']?.toString() ?? field['value']?.toString() ?? '';
      if (name.isNotEmpty) {
        buffer.write("$newline-F '$name=${_escapeShellArg(value)}'");
      }
    }

    // Add file uploads (Note: file paths need to be manually adjusted)
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

  /// Escape shell argument for safe inclusion in curl command
  ///
  /// Handles single quotes by replacing them with '\''
  String _escapeShellArg(String arg) {
    return arg.replaceAll("'", "'\\''");
  }
}
