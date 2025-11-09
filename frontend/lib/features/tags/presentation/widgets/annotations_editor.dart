import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';

import '../../../../core/di/di.dart';
import '../../application/stores/tags_store.dart';
import '../../domain/entities/session_annotation.dart';

class AnnotationsEditor extends StatefulWidget {
  const AnnotationsEditor({required this.sessionId, super.key});

  final String sessionId;

  @override
  State<AnnotationsEditor> createState() => _AnnotationsEditorState();
}

class _AnnotationsEditorState extends State<AnnotationsEditor> {
  final _keyController = TextEditingController();
  final _valueController = TextEditingController();
  String? _editingKey;

  @override
  void initState() {
    super.initState();
    final store = sl<TagsStore>();
    store.loadSessionAnnotations(widget.sessionId);
  }

  @override
  void dispose() {
    _keyController.dispose();
    _valueController.dispose();
    super.dispose();
  }

  Future<void> _saveAnnotation() async {
    final key = _keyController.text.trim();
    final value = _valueController.text.trim();

    if (key.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Key is required'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    final store = sl<TagsStore>();
    try {
      await store.upsertSessionAnnotation(widget.sessionId, key, value);
      _keyController.clear();
      _valueController.clear();
      setState(() {
        _editingKey = null;
      });
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Annotation "$key" saved')));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to save annotation: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _deleteAnnotation(String key) async {
    final store = sl<TagsStore>();
    try {
      await store.deleteSessionAnnotation(widget.sessionId, key);
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Annotation "$key" deleted')));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to delete annotation: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  void _startEditing(SessionAnnotation annotation) {
    setState(() {
      _editingKey = annotation.key;
      _keyController.text = annotation.key;
      _valueController.text = annotation.value;
    });
  }

  void _cancelEditing() {
    setState(() {
      _editingKey = null;
      _keyController.clear();
      _valueController.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    final store = sl<TagsStore>();

    return Observer(
      builder: (_) {
        final annotations = store.getSessionAnnotationsFromCache(
          widget.sessionId,
        );

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Add/Edit form
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    TextField(
                      controller: _keyController,
                      decoration: InputDecoration(
                        labelText: 'Key',
                        hintText: 'e.g., description, jira-ticket',
                        border: const OutlineInputBorder(),
                        enabled: _editingKey == null,
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _valueController,
                      decoration: const InputDecoration(
                        labelText: 'Value',
                        hintText: 'Enter value',
                        border: OutlineInputBorder(),
                      ),
                      maxLines: 3,
                    ),
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        if (_editingKey != null)
                          TextButton(
                            onPressed: _cancelEditing,
                            child: const Text('Cancel'),
                          ),
                        const SizedBox(width: 8),
                        ElevatedButton.icon(
                          onPressed: _saveAnnotation,
                          icon: const Icon(Icons.save, size: 18),
                          label: Text(_editingKey != null ? 'Update' : 'Add'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),

            const SizedBox(height: 16),

            // Annotations list
            if (annotations.isEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: Text(
                  'No annotations yet',
                  style: Theme.of(
                    context,
                  ).textTheme.bodySmall?.copyWith(color: Colors.grey),
                ),
              )
            else
              ListView.separated(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                itemCount: annotations.length,
                separatorBuilder: (_, __) => const Divider(),
                itemBuilder: (context, index) {
                  final annotation = annotations[index];
                  final isEditing = _editingKey == annotation.key;

                  return ListTile(
                    title: Text(
                      annotation.key,
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                    subtitle: Text(
                      annotation.value,
                      maxLines: 3,
                      overflow: TextOverflow.ellipsis,
                    ),
                    trailing: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        IconButton(
                          icon: const Icon(Icons.edit, size: 20),
                          onPressed: isEditing
                              ? null
                              : () => _startEditing(annotation),
                          tooltip: 'Edit',
                        ),
                        IconButton(
                          icon: const Icon(Icons.delete, size: 20),
                          onPressed: () => _deleteAnnotation(annotation.key),
                          tooltip: 'Delete',
                          color: Colors.red,
                        ),
                      ],
                    ),
                  );
                },
              ),
          ],
        );
      },
    );
  }
}
