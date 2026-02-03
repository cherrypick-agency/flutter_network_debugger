import 'package:flutter/material.dart';

import '../../theme/context_ext.dart';

/// Compact banner for displaying loading/error state inside panels.
///
/// The idea is simple: avoid spamming global notifications on background requests,
/// and show the state right where the user is currently looking.
class InlineStatusBanner extends StatelessWidget {
  const InlineStatusBanner({
    super.key,
    required this.loading,
    required this.errorText,
    this.onRetry,
    this.loadingText,
  });

  final bool loading;
  final String? errorText;
  final VoidCallback? onRetry;
  final String? loadingText;

  @override
  Widget build(BuildContext context) {
    final err = errorText?.trim();
    if ((err == null || err.isEmpty) && !loading) {
      return const SizedBox.shrink();
    }

    if (err != null && err.isNotEmpty) {
      final scheme = Theme.of(context).colorScheme;
      return Material(
        color: scheme.errorContainer,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          child: Row(
            children: [
              Icon(
                Icons.error_outline,
                color: scheme.onErrorContainer,
                size: 18,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  err,
                  style: context.appText.body.copyWith(
                    color: scheme.onErrorContainer,
                  ),
                  maxLines: 3,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (onRetry != null) ...[
                const SizedBox(width: 8),
                TextButton(
                  onPressed: onRetry,
                  child: Text(
                    'Retry',
                    style: TextStyle(color: scheme.onErrorContainer),
                  ),
                ),
              ],
            ],
          ),
        ),
      );
    }

    // loading
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const LinearProgressIndicator(minHeight: 2),
        if ((loadingText ?? '').trim().isNotEmpty)
          Padding(
            padding: const EdgeInsets.fromLTRB(10, 6, 10, 0),
            child: Text(
              loadingText!.trim(),
              style: context.appText.body.copyWith(
                color: context.appColors.textSecondary,
              ),
            ),
          ),
      ],
    );
  }
}
