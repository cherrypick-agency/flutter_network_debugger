import 'package:flutter/material.dart';
import '../../domain/entities/script.dart';
import '../../application/stores/scripts_store.dart';

/// Filters dialog for scripts list
class ScriptsFiltersDialog extends StatefulWidget {
  final ScriptsStore store;

  const ScriptsFiltersDialog({super.key, required this.store});

  @override
  State<ScriptsFiltersDialog> createState() => _ScriptsFiltersDialogState();

  static Future<void> show(BuildContext context, ScriptsStore store) async {
    return showDialog(
      context: context,
      builder: (context) => ScriptsFiltersDialog(store: store),
    );
  }
}

class _ScriptsFiltersDialogState extends State<ScriptsFiltersDialog> {
  late ScriptRuntime? _selectedRuntime;
  late TriggerType? _selectedTriggerType;
  late bool? _selectedEnabledFilter;

  @override
  void initState() {
    super.initState();
    _selectedRuntime = widget.store.runtimeFilter;
    _selectedTriggerType = widget.store.triggerTypeFilter;
    _selectedEnabledFilter = widget.store.enabledFilter;
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          const Icon(Icons.filter_list),
          const SizedBox(width: 12),
          const Text('Filter Scripts'),
          const Spacer(),
          IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
      content: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            // Runtime filter
            Text('Runtime', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              children: [
                ChoiceChip(
                  label: const Text('All'),
                  selected: _selectedRuntime == null,
                  onSelected: (selected) {
                    setState(() => _selectedRuntime = null);
                  },
                ),
                ChoiceChip(
                  label: const Text('Extism (WASM)'),
                  avatar: const Icon(Icons.memory, size: 16),
                  selected: _selectedRuntime == ScriptRuntime.extism,
                  onSelected: (selected) {
                    setState(
                      () => _selectedRuntime = selected
                          ? ScriptRuntime.extism
                          : null,
                    );
                  },
                ),
                ChoiceChip(
                  label: const Text('Dart'),
                  avatar: const Icon(Icons.code, size: 16),
                  selected: _selectedRuntime == ScriptRuntime.dart,
                  onSelected: (selected) {
                    setState(
                      () => _selectedRuntime = selected
                          ? ScriptRuntime.dart
                          : null,
                    );
                  },
                ),
              ],
            ),
            const SizedBox(height: 24),

            // Trigger type filter
            Text(
              'Trigger Type',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                ChoiceChip(
                  label: const Text('All'),
                  selected: _selectedTriggerType == null,
                  onSelected: (selected) {
                    setState(() => _selectedTriggerType = null);
                  },
                ),
                ChoiceChip(
                  label: const Text('Request'),
                  avatar: const Icon(Icons.arrow_upward, size: 16),
                  selected: _selectedTriggerType == TriggerType.request,
                  onSelected: (selected) {
                    setState(
                      () => _selectedTriggerType = selected
                          ? TriggerType.request
                          : null,
                    );
                  },
                ),
                ChoiceChip(
                  label: const Text('Response'),
                  avatar: const Icon(Icons.arrow_downward, size: 16),
                  selected: _selectedTriggerType == TriggerType.response,
                  onSelected: (selected) {
                    setState(
                      () => _selectedTriggerType = selected
                          ? TriggerType.response
                          : null,
                    );
                  },
                ),
                ChoiceChip(
                  label: const Text('Both'),
                  avatar: const Icon(Icons.swap_vert, size: 16),
                  selected: _selectedTriggerType == TriggerType.both,
                  onSelected: (selected) {
                    setState(
                      () => _selectedTriggerType = selected
                          ? TriggerType.both
                          : null,
                    );
                  },
                ),
              ],
            ),
            const SizedBox(height: 24),

            // Status filter
            Text('Status', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            SegmentedButton<String>(
              segments: const [
                ButtonSegment(value: 'all', label: Text('All')),
                ButtonSegment(
                  value: 'enabled',
                  label: Text('Enabled'),
                  icon: Icon(Icons.check_circle, size: 16),
                ),
                ButtonSegment(
                  value: 'disabled',
                  label: Text('Disabled'),
                  icon: Icon(Icons.cancel, size: 16),
                ),
              ],
              selected: {
                _selectedEnabledFilter == null
                    ? 'all'
                    : _selectedEnabledFilter!
                    ? 'enabled'
                    : 'disabled',
              },
              onSelectionChanged: (Set<String> selected) {
                setState(() {
                  switch (selected.first) {
                    case 'all':
                      _selectedEnabledFilter = null;
                      break;
                    case 'enabled':
                      _selectedEnabledFilter = true;
                      break;
                    case 'disabled':
                      _selectedEnabledFilter = false;
                      break;
                  }
                });
              },
            ),
          ],
        ),
      ),
      actions: [
        // Clear filters
        TextButton(
          onPressed: () {
            setState(() {
              _selectedRuntime = null;
              _selectedTriggerType = null;
              _selectedEnabledFilter = null;
            });
          },
          child: const Text('Clear All'),
        ),
        const Spacer(),
        // Cancel
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Cancel'),
        ),
        // Apply
        FilledButton(
          onPressed: () {
            widget.store.setRuntimeFilter(_selectedRuntime);
            widget.store.setTriggerTypeFilter(_selectedTriggerType);
            widget.store.setEnabledFilter(_selectedEnabledFilter);
            Navigator.of(context).pop();
          },
          child: const Text('Apply'),
        ),
      ],
    );
  }
}
