import 'package:flutter/material.dart';

class ComposeHistoryList extends StatelessWidget {
  const ComposeHistoryList({
    super.key,
    required this.items,
    this.currentTemplateId,
    this.onTapSessionId,
    required this.onTapTemplate,
  });

  final List<Map<String, dynamic>>
  items; // {id, templateId, method, url, status, when}
  final String? currentTemplateId;
  final void Function(String? sessionId)? onTapSessionId;
  final void Function(String templateId) onTapTemplate;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const Center(child: Text('No history yet'));
    }
    return ListView.separated(
      itemCount: items.length,
      separatorBuilder: (_, __) => const Divider(height: 1),
      itemBuilder: (context, i) {
        final it = items[i];
        final method = (it['method'] ?? '').toString();
        final url = (it['url'] ?? '').toString();
        final status = (it['status'] ?? '').toString();
        final templateId = (it['templateId'] ?? '').toString();
        final sessionId = (it['sessionId'] ?? '').toString();
        return ListTile(
          dense: true,
          selected:
              currentTemplateId != null &&
              templateId.isNotEmpty &&
              templateId == currentTemplateId,
          leading: CircleAvatar(
            radius: 10,
            backgroundColor: _statusColor(context, int.tryParse(status) ?? 0),
            child: const SizedBox.shrink(),
          ),
          title: Text(
            '$method $url',
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          subtitle: Text('Status: $status'),
          onTap: () {
            onTapSessionId?.call(sessionId.isNotEmpty ? sessionId : null);
            if (templateId.isNotEmpty) onTapTemplate(templateId);
          },
        );
      },
    );
  }

  Color _statusColor(BuildContext context, int status) {
    final cs = Theme.of(context).colorScheme;
    if (status >= 200 && status < 300) return cs.primary;
    if (status >= 400 && status < 500) return cs.tertiary;
    if (status >= 500) return cs.error;
    return cs.outlineVariant;
  }
}
