import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';

import '../../../../core/di/di.dart';
import '../../../../core/network/error_utils.dart';
import '../../application/stores/tags_store.dart';
import '../../domain/entities/predefined_tag.dart';
import 'tag_chip.dart';

class TagsEditor extends StatefulWidget {
  const TagsEditor({required this.sessionId, super.key});

  final String sessionId;

  @override
  State<TagsEditor> createState() => _TagsEditorState();
}

class _TagsEditorState extends State<TagsEditor> {
  final _controller = TextEditingController();
  final _focusNode = FocusNode();
  List<PredefinedTag> _filteredTags = [];
  String? _errorText;

  @override
  void initState() {
    super.initState();
    final store = sl<TagsStore>();

    // Load predefined tags if not loaded
    if (store.predefinedTags.isEmpty) {
      store.loadPredefinedTags();
    }

    // Load session tags
    store.loadSessionTags(widget.sessionId);

    _controller.addListener(_onSearchChanged);
    _focusNode.addListener(() {
      // Перерисуем подсказки при смене фокуса
      _onSearchChanged();
      if (mounted) setState(() {});
    });
  }

  @override
  void dispose() {
    _controller.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _onSearchChanged() {
    final query = _controller.text.toLowerCase();
    final store = sl<TagsStore>();

    if (query.isEmpty) {
      setState(() {
        // Если поле в фокусе — показываем весь список предустановленных тегов,
        // иначе скрываем подсказки.
        _filteredTags = _focusNode.hasFocus
            ? store.predefinedTags.toList()
            : [];
        _errorText = null;
      });
      return;
    }

    setState(() {
      _filteredTags = store.predefinedTags
          .where((tag) => tag.name.toLowerCase().contains(query))
          .toList();
      _errorText = null;
    });
  }

  Future<void> _addTag(String tagName) async {
    if (tagName.trim().isEmpty) {
      setState(() {
        _errorText = 'Enter tag name';
      });
      _focusNode.requestFocus();
      return;
    }

    final store = sl<TagsStore>();
    try {
      // Если такого тега нет в базе предопределённых — добавим его для переиспользования
      final exists = store.predefinedTags.any(
        (t) => t.name.toLowerCase() == tagName.trim().toLowerCase(),
      );
      if (!exists) {
        await store.createPredefinedTag(
          name: tagName.trim(),
          color: '#808080',
          category: 'custom',
          displayOrder: 999,
        );
      }
      await store.addSessionTag(widget.sessionId, tagName.trim());
      _controller.clear();
      setState(() {
        _filteredTags = [];
        _errorText = null;
      });
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Tag "$tagName" added')));
      }
    } catch (e) {
      if (mounted) {
        final msg = resolveErrorMessage(e);
        final status = msg.details?['statusCode'];
        final method = (msg.details?['method'] ?? '').toString();
        final url = (msg.details?['url'] ?? '').toString();
        String shortUrl = url;
        try {
          final u = Uri.parse(url);
          shortUrl = u.path.isNotEmpty ? u.path : url;
        } catch (_) {}
        final text = (status != null || method.isNotEmpty || url.isNotEmpty)
            ? 'HTTP ${status ?? ''} $method $shortUrl'
            : '${msg.title}: ${msg.description}';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(text),
            backgroundColor: Theme.of(context).colorScheme.error,
          ),
        );
      }
    }
  }

  Future<void> _removeTag(String tagName) async {
    final store = sl<TagsStore>();
    try {
      await store.removeSessionTag(widget.sessionId, tagName);
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Tag "$tagName" removed')));
      }
    } catch (e) {
      if (mounted) {
        final msg = resolveErrorMessage(e);
        final status = msg.details?['statusCode'];
        final method = (msg.details?['method'] ?? '').toString();
        final url = (msg.details?['url'] ?? '').toString();
        String shortUrl = url;
        try {
          final u = Uri.parse(url);
          shortUrl = u.path.isNotEmpty ? u.path : url;
        } catch (_) {}
        final text = (status != null || method.isNotEmpty || url.isNotEmpty)
            ? 'HTTP ${status ?? ''} $method $shortUrl'
            : '${msg.title}: ${msg.description}';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(text),
            backgroundColor: Theme.of(context).colorScheme.error,
          ),
        );
      }
    }
  }

  Color _parseColor(String colorHex) {
    try {
      return Color(int.parse(colorHex.replaceFirst('#', '0xFF')));
    } catch (e) {
      return Colors.grey;
    }
  }

  @override
  Widget build(BuildContext context) {
    final store = sl<TagsStore>();

    return Observer(
      builder: (_) {
        final tags = store.getSessionTagsFromCache(widget.sessionId);
        final predefinedMap = {
          for (var tag in store.predefinedTags) tag.name: tag,
        };

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Input field with autocomplete
            TextField(
              controller: _controller,
              focusNode: _focusNode,
              decoration: InputDecoration(
                labelText: 'Add tag',
                hintText: 'Type to search or add custom tag',
                errorText: _errorText,
                suffixIcon: IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: () => _addTag(_controller.text),
                ),
                border: const OutlineInputBorder(),
              ),
              onSubmitted: _addTag,
            ),

            // Autocomplete suggestions (только когда в фокусе)
            if (_focusNode.hasFocus && _filteredTags.isNotEmpty) ...[
              const SizedBox(height: 8),
              Container(
                constraints: const BoxConstraints(maxHeight: 150),
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.grey.shade300),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: ListView.builder(
                  shrinkWrap: true,
                  itemCount: _filteredTags.length,
                  itemBuilder: (context, index) {
                    final tag = _filteredTags[index];
                    return ListTile(
                      dense: true,
                      leading: Container(
                        width: 16,
                        height: 16,
                        decoration: BoxDecoration(
                          color: _parseColor(tag.color),
                          shape: BoxShape.circle,
                        ),
                      ),
                      title: Text(tag.name),
                      subtitle: Text(
                        tag.category,
                        style: const TextStyle(fontSize: 11),
                      ),
                      onTap: () => _addTag(tag.name),
                    );
                  },
                ),
              ),
            ],

            const SizedBox(height: 16),

            // Current tags
            if (tags.isEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: Text(
                  'No tags yet',
                  style: Theme.of(
                    context,
                  ).textTheme.bodySmall?.copyWith(color: Colors.grey),
                ),
              )
            else
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: tags.map((sessionTag) {
                  final predefinedTag = predefinedMap[sessionTag.tagName];
                  final color = predefinedTag != null
                      ? _parseColor(predefinedTag.color)
                      : null;

                  return TagChip(
                    label: sessionTag.tagName,
                    color: color,
                    onDeleted: () => _removeTag(sessionTag.tagName),
                  );
                }).toList(),
              ),
          ],
        );
      },
    );
  }
}
