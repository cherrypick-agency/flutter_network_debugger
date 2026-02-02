import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';

class MappingRuleErrorBanner extends StatelessWidget {
  const MappingRuleErrorBanner({
    super.key,
    required this.title,
    required this.message,
    this.onClose,
  });

  final String title;
  final String message;
  final VoidCallback? onClose;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Theme.of(context).colorScheme.errorContainer,
      borderRadius: BorderRadius.circular(12),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        child: Row(
          children: [
            Icon(
              Icons.error_outline,
              color: Theme.of(context).colorScheme.onErrorContainer,
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: context.appText.subtitle.copyWith(
                      color: Theme.of(context).colorScheme.onErrorContainer,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    message,
                    style: context.appText.body.copyWith(
                      color: Theme.of(context).colorScheme.onErrorContainer,
                    ),
                  ),
                ],
              ),
            ),
            if (onClose != null) ...[
              const SizedBox(width: 8),
              IconButton(
                onPressed: onClose,
                icon: Icon(
                  Icons.close,
                  color: Theme.of(context).colorScheme.onErrorContainer,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
