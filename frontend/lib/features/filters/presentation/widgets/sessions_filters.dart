import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';

import '../../application/stores/sessions_filters_store.dart';
import '../../../../core/di/di.dart';
import '../../../tags/application/stores/tags_store.dart';

class SessionsFilters extends StatelessWidget {
  const SessionsFilters({super.key, required this.onApply});

  final VoidCallback onApply;

  Color _getMethodColor(String method) {
    switch (method.toUpperCase()) {
      case 'GET':
        return Colors.purple;
      case 'POST':
        return Colors.blue;
      case 'PUT':
      case 'PATCH':
        return Colors.orange;
      case 'DELETE':
        return Colors.red;
      default:
        return Colors.grey;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Observer(
      builder: (_) {
        final store = context.read<SessionsFiltersStore>();
        return Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            IntrinsicWidth(
              child: DropdownButtonFormField<String>(
                value: store.httpMethod,
                isDense: true,
                iconSize: 18,
                style: TextStyle(
                  fontSize: 12,
                  color: Theme.of(context).colorScheme.onSurface,
                ),
                decoration: const InputDecoration(
                  isDense: true,
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 8,
                  ),
                ),
                selectedItemBuilder: (context) {
                  return [
                    'any',
                    'GET',
                    'POST',
                    'PUT',
                    'DELETE',
                    'PATCH',
                    'OPTIONS',
                  ].map((method) {
                    if (method == 'any') {
                      return const Text(
                        'Any method',
                        style: TextStyle(fontSize: 12),
                      );
                    }
                    return Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 10,
                          height: 10,
                          decoration: BoxDecoration(
                            color: _getMethodColor(method),
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(method, style: const TextStyle(fontSize: 12)),
                      ],
                    );
                  }).toList();
                },
                items: [
                  const DropdownMenuItem(
                    value: 'any',
                    child: Text('Any method', style: TextStyle(fontSize: 12)),
                  ),
                  DropdownMenuItem(
                    value: 'GET',
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 10,
                          height: 10,
                          decoration: BoxDecoration(
                            color: _getMethodColor('GET'),
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        const Text('GET', style: TextStyle(fontSize: 12)),
                      ],
                    ),
                  ),
                  DropdownMenuItem(
                    value: 'POST',
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 10,
                          height: 10,
                          decoration: BoxDecoration(
                            color: _getMethodColor('POST'),
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        const Text('POST', style: TextStyle(fontSize: 12)),
                      ],
                    ),
                  ),
                  DropdownMenuItem(
                    value: 'PUT',
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 10,
                          height: 10,
                          decoration: BoxDecoration(
                            color: _getMethodColor('PUT'),
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        const Text('PUT', style: TextStyle(fontSize: 12)),
                      ],
                    ),
                  ),
                  DropdownMenuItem(
                    value: 'DELETE',
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 10,
                          height: 10,
                          decoration: BoxDecoration(
                            color: _getMethodColor('DELETE'),
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        const Text('DELETE', style: TextStyle(fontSize: 12)),
                      ],
                    ),
                  ),
                  DropdownMenuItem(
                    value: 'PATCH',
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 10,
                          height: 10,
                          decoration: BoxDecoration(
                            color: _getMethodColor('PATCH'),
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        const Text('PATCH', style: TextStyle(fontSize: 12)),
                      ],
                    ),
                  ),
                  DropdownMenuItem(
                    value: 'OPTIONS',
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 10,
                          height: 10,
                          decoration: BoxDecoration(
                            color: _getMethodColor('OPTIONS'),
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        const Text('OPTIONS', style: TextStyle(fontSize: 12)),
                      ],
                    ),
                  ),
                ],
                onChanged: (v) {
                  store.setHttpMethod(v ?? 'any');
                  onApply();
                },
              ),
            ),
            IntrinsicWidth(
              child: DropdownButtonFormField<String>(
                value: store.httpStatus,
                isDense: true,
                iconSize: 18,
                style: TextStyle(
                  fontSize: 12,
                  color: Theme.of(context).colorScheme.onSurface,
                ),
                decoration: const InputDecoration(
                  isDense: true,
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 8,
                  ),
                ),
                items: const [
                  DropdownMenuItem(
                    value: 'any',
                    child: Text('Any status', style: TextStyle(fontSize: 12)),
                  ),
                  DropdownMenuItem(
                    value: '2xx',
                    child: Text('2xx', style: TextStyle(fontSize: 12)),
                  ),
                  DropdownMenuItem(
                    value: '3xx',
                    child: Text('3xx', style: TextStyle(fontSize: 12)),
                  ),
                  DropdownMenuItem(
                    value: '4xx',
                    child: Text('4xx', style: TextStyle(fontSize: 12)),
                  ),
                  DropdownMenuItem(
                    value: '5xx',
                    child: Text('5xx', style: TextStyle(fontSize: 12)),
                  ),
                ],
                onChanged: (v) {
                  store.setHttpStatus(v ?? 'any');
                  onApply();
                },
              ),
            ),
            IntrinsicWidth(
              child: DropdownButtonFormField<String>(
                value: store.groupBy,
                isDense: true,
                iconSize: 18,
                style: TextStyle(
                  fontSize: 12,
                  color: Theme.of(context).colorScheme.onSurface,
                ),
                decoration: const InputDecoration(
                  isDense: true,
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 8,
                  ),
                ),
                items: const [
                  DropdownMenuItem(
                    value: 'none',
                    child: Text('No grouping', style: TextStyle(fontSize: 12)),
                  ),
                  DropdownMenuItem(
                    value: 'domain',
                    child: Text(
                      'Group by domain',
                      style: TextStyle(fontSize: 12),
                    ),
                  ),
                  DropdownMenuItem(
                    value: 'route',
                    child: Text(
                      'Group by route',
                      style: TextStyle(fontSize: 12),
                    ),
                  ),
                ],
                onChanged: (v) {
                  store.setGroupBy(v ?? 'none');
                  onApply();
                },
              ),
            ),
            SizedBox(
              width: 200,
              child: TextField(
                style: const TextStyle(fontSize: 12),
                decoration: const InputDecoration(labelText: 'MIME contains'),
                onChanged: (v) {
                  store.setHttpMime(v);
                  onApply();
                },
              ),
            ),
            SizedBox(
              width: 120,
              child: TextField(
                style: const TextStyle(fontSize: 12),
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(labelText: 'Min ms'),
                onChanged: (v) {
                  store.setHttpMinDurationMs(int.tryParse(v) ?? 0);
                  onApply();
                },
              ),
            ),
            SizedBox(
              width: 160,
              child: TextField(
                style: const TextStyle(fontSize: 12),
                decoration: const InputDecoration(labelText: 'Header key'),
                onChanged: (v) {
                  store.setHeaderKey(v);
                  onApply();
                },
              ),
            ),
            SizedBox(
              width: 180,
              child: TextField(
                style: const TextStyle(fontSize: 12),
                decoration: const InputDecoration(labelText: 'Header value'),
                onChanged: (v) {
                  store.setHeaderVal(v);
                  onApply();
                },
              ),
            ),
            /*
            SizedBox(
              width: 200,
              child: TextField(
                style: const TextStyle(fontSize: 12),
                decoration: const InputDecoration(labelText: 'target'),
                onChanged: (v) {
                  store.setTarget(v);
                  onApply();
                },
                onSubmitted: (_) => onApply(),
              ),
            ),
            */
            // Tags filter section
            _TagsFilterSection(onApply: onApply),
            // Reset button appears only when there are deviations from defaults
            if (store.hasActive)
              IntrinsicWidth(
                child: TextButton.icon(
                  style: TextButton.styleFrom(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 6,
                    ),
                    textStyle: const TextStyle(fontSize: 12),
                  ),
                  onPressed: () {
                    // Reset all values to defaults and apply
                    store
                      ..setTarget('')
                      ..setHttpMethod('any')
                      ..setHttpStatus('any')
                      ..setHttpMime('')
                      ..setHttpMinDurationMs(0)
                      ..setGroupBy('none')
                      ..setHeaderKey('')
                      ..setHeaderVal('')
                      ..clearTags();
                    onApply();
                  },
                  icon: const Icon(Icons.restart_alt, size: 16),
                  label: const Text('Reset'),
                ),
              ),
            /*
            Observer(
              builder: (_) {
                final loading = sessions.loading;
                return loading
                    ? const SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                    : const SizedBox.shrink();
              },
            ),
            */
          ],
        );
      },
    );
  }
}

// Widget for tags filtering section
class _TagsFilterSection extends StatefulWidget {
  const _TagsFilterSection({required this.onApply});

  final VoidCallback onApply;

  @override
  State<_TagsFilterSection> createState() => _TagsFilterSectionState();
}

class _TagsFilterSectionState extends State<_TagsFilterSection> {
  bool _loaded = false;

  @override
  void initState() {
    super.initState();
    _loadTags();
  }

  Future<void> _loadTags() async {
    if (_loaded) return;
    try {
      await sl<TagsStore>().loadPredefinedTags();
      _loaded = true;
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    final tagsStore = sl<TagsStore>();
    final filtersStore = context.read<SessionsFiltersStore>();

    return Observer(
      builder: (_) {
        final tags = tagsStore.predefinedTags;
        if (tags.isEmpty) {
          return const SizedBox.shrink();
        }

        final selectedTags = filtersStore.selectedTags;

        return MenuAnchor(
          menuChildren: tags.map((tag) {
            final isSelected = selectedTags.contains(tag.name);
            return MenuItemButton(
              leadingIcon: isSelected
                  ? const Icon(Icons.check_box, size: 18)
                  : const Icon(Icons.check_box_outline_blank, size: 18),
              onPressed: () {
                filtersStore.toggleTag(tag.name);
                widget.onApply();
              },
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    width: 12,
                    height: 12,
                    decoration: BoxDecoration(
                      color: Color(
                        int.parse(tag.color.substring(1), radix: 16) +
                            0xFF000000,
                      ),
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(tag.name, style: const TextStyle(fontSize: 12)),
                ],
              ),
            );
          }).toList(),
          builder: (context, controller, child) {
            return IntrinsicWidth(
              child: OutlinedButton.icon(
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 6,
                  ),
                  textStyle: const TextStyle(fontSize: 12),
                ),
                onPressed: () {
                  controller.isOpen ? controller.close() : controller.open();
                },
                icon: Icon(
                  Icons.label,
                  size: 16,
                  color: selectedTags.isNotEmpty
                      ? Theme.of(context).colorScheme.primary
                      : null,
                ),
                label: Text(
                  selectedTags.isEmpty
                      ? 'Filter by tags'
                      : 'Tags (${selectedTags.length})',
                  style: TextStyle(
                    color: selectedTags.isNotEmpty
                        ? Theme.of(context).colorScheme.primary
                        : null,
                  ),
                ),
              ),
            );
          },
        );
      },
    );
  }
}
