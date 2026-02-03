import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';

class EditorPanelToolbar extends StatelessWidget {
  const EditorPanelToolbar({
    super.key,
    required this.isRequest,
    required this.methodController,
    required this.urlController,
    required this.statusController,
    required this.responseContentType,
    required this.submitting,
    required this.onAddAuth,
    required this.onAddJson,
    required this.onDrop,
    required this.onCancel,
    required this.onContinue,
  });

  final bool isRequest;
  final TextEditingController methodController;
  final TextEditingController urlController;
  final TextEditingController statusController;
  final String? responseContentType;
  final bool submitting;

  final VoidCallback onAddAuth;
  final VoidCallback onAddJson;
  final VoidCallback onDrop;
  final VoidCallback onCancel;
  final VoidCallback onContinue;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        if (isRequest)
          SizedBox(
            width: 100,
            child: TextField(
              controller: methodController,
              decoration: const InputDecoration(labelText: 'Method'),
            ),
          ),
        if (isRequest) const SizedBox(width: 8),
        if (isRequest)
          Expanded(
            child: TextField(
              controller: urlController,
              decoration: const InputDecoration(labelText: 'URL'),
            ),
          ),
        if (!isRequest)
          SizedBox(
            width: 120,
            child: TextField(
              controller: statusController,
              decoration: const InputDecoration(labelText: 'Status'),
              keyboardType: TextInputType.number,
            ),
          ),
        if (!isRequest) const SizedBox(width: 8),
        if (!isRequest)
          Expanded(
            child: Text(
              responseContentType ?? '',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: context.appText.body,
            ),
          ),
        const SizedBox(width: 12),
        TextButton(onPressed: onAddAuth, child: const Text('+Auth')),
        const SizedBox(width: 4),
        TextButton(onPressed: onAddJson, child: const Text('+JSON')),
        if (isRequest)
          FilledButton(
            onPressed: submitting ? null : onDrop,
            child: const Text('Drop'),
          ),
        const SizedBox(width: 8),
        OutlinedButton(
          onPressed: submitting ? null : onCancel,
          child: const Text('Cancel'),
        ),
        const SizedBox(width: 8),
        ElevatedButton(
          onPressed: submitting ? null : onContinue,
          child: const Text('Continue'),
        ),
      ],
    );
  }
}
