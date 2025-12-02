import 'package:flutter/material.dart';
import '../types/body_callbacks.dart';
import '../../../../widgets/copy_curl_button.dart';

// Re-export для обратной совместимости
export '../../../../widgets/copy_curl_button.dart' show CurlExportMode;

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
        CopyCurlButton(
          url: url,
          method: method,
          headers: headers,
          body: body,
          form: form,
          onCopied: onCopyCurl,
          showSnackBar: false,
        ),
        if (url.isNotEmpty) _buildRepeatButton(),
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
}
