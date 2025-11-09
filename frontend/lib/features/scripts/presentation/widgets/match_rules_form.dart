import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import '../../application/stores/script_editor_store.dart';
import '../../domain/entities/match_rules.dart';

/// Visual builder for match rules (HTTP methods, path patterns, etc.)
class MatchRulesForm extends StatelessWidget {
  final ScriptEditorStore store;

  const MatchRulesForm({super.key, required this.store});

  static const List<String> _httpMethods = [
    'GET',
    'POST',
    'PUT',
    'DELETE',
    'PATCH',
    'HEAD',
    'OPTIONS',
  ];

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Toggle for custom rules
          Observer(
            builder: (_) => SwitchListTile(
              title: const Text('Use Custom Match Rules'),
              subtitle: const Text('Apply script only to specific requests'),
              value: store.useCustomMatchRules,
              onChanged: store.setUseCustomMatchRules,
            ),
          ),
          const SizedBox(height: 24),

          if (store.useCustomMatchRules) ...[
            // HTTP Methods selector
            Text(
              'HTTP Methods',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            const Text(
              'Select which HTTP methods should trigger this script. Leave empty to match all methods.',
              style: TextStyle(color: Colors.grey, fontSize: 12),
            ),
            const SizedBox(height: 12),
            Observer(
              builder: (_) => Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  ..._httpMethods.map((method) {
                    final isSelected = store.selectedMethods.contains(method);
                    return FilterChip(
                      label: Text(method),
                      selected: isSelected,
                      onSelected: (_) => store.toggleMethod(method),
                      selectedColor: Theme.of(
                        context,
                      ).colorScheme.primaryContainer,
                    );
                  }),
                  // Quick select buttons
                  if (store.selectedMethods.isEmpty)
                    ActionChip(
                      label: const Text('Select All'),
                      avatar: const Icon(Icons.select_all, size: 16),
                      onPressed: () {
                        for (final method in _httpMethods) {
                          if (!store.selectedMethods.contains(method)) {
                            store.toggleMethod(method);
                          }
                        }
                      },
                    )
                  else
                    ActionChip(
                      label: const Text('Clear All'),
                      avatar: const Icon(Icons.clear, size: 16),
                      onPressed: () {
                        for (final method in List.from(store.selectedMethods)) {
                          store.toggleMethod(method);
                        }
                      },
                    ),
                ],
              ),
            ),
            const SizedBox(height: 24),

            // Pattern Type selector
            Text(
              'Pattern Type',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Observer(
              builder: (_) => SegmentedButton<PatternType>(
                segments: const [
                  ButtonSegment(
                    value: PatternType.wildcard,
                    label: Text('Wildcard'),
                    tooltip: 'Use * for matching (e.g., /api/*)',
                  ),
                  ButtonSegment(
                    value: PatternType.prefix,
                    label: Text('Prefix'),
                    tooltip: 'Match paths starting with pattern',
                  ),
                  ButtonSegment(
                    value: PatternType.exact,
                    label: Text('Exact'),
                    tooltip: 'Exact path match',
                  ),
                ],
                selected: {store.patternType},
                onSelectionChanged: (Set<PatternType> selected) {
                  store.setPatternType(selected.first);
                },
              ),
            ),
            const SizedBox(height: 24),

            // Path Pattern
            Text(
              'Path Pattern',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Observer(
              builder: (_) => TextField(
                decoration: InputDecoration(
                  labelText: 'URL Path',
                  hintText: _getPathPatternHint(store.patternType),
                  border: const OutlineInputBorder(),
                  helperText: _getPathPatternHelp(store.patternType),
                  prefixIcon: const Icon(Icons.route),
                ),
                controller: TextEditingController(text: store.pathPattern)
                  ..selection = TextSelection.fromPosition(
                    TextPosition(offset: store.pathPattern.length),
                  ),
                onChanged: store.setPathPattern,
              ),
            ),
            const SizedBox(height: 24),

            // Host Pattern
            Text(
              'Host Pattern (Optional)',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Observer(
              builder: (_) => TextField(
                decoration: const InputDecoration(
                  labelText: 'Host',
                  hintText: 'example.com or api.*',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.dns),
                ),
                controller: TextEditingController(text: store.hostPattern)
                  ..selection = TextSelection.fromPosition(
                    TextPosition(offset: store.hostPattern.length),
                  ),
                onChanged: store.setHostPattern,
              ),
            ),
            const SizedBox(height: 32),

            // Preview panel
            _buildPreviewPanel(context),
          ] else ...[
            // No custom rules info
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    Icon(
                      Icons.info_outline,
                      color: Theme.of(context).colorScheme.primary,
                    ),
                    const SizedBox(width: 16),
                    const Expanded(
                      child: Text(
                        'This script will apply to all HTTP requests passing through the proxy.',
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildPreviewPanel(BuildContext context) {
    return Observer(
      builder: (_) {
        final matchRules = store.matchRules;
        if (matchRules == null || matchRules.isEmpty) {
          return const SizedBox.shrink();
        }

        return Card(
          color: Theme.of(context).colorScheme.surfaceContainerHighest,
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(
                      Icons.preview,
                      size: 20,
                      color: Theme.of(context).colorScheme.primary,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      'Match Preview',
                      style: Theme.of(context).textTheme.titleSmall,
                    ),
                  ],
                ),
                const SizedBox(height: 12),

                // Description
                Text(
                  matchRules.description,
                  style: const TextStyle(fontWeight: FontWeight.w500),
                ),
                const SizedBox(height: 16),

                // Examples that match
                if (matchRules.exampleMatches.isNotEmpty) ...[
                  const Text(
                    '✅ Will match:',
                    style: TextStyle(
                      color: Colors.green,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  ...matchRules.exampleMatches.map(
                    (example) => Padding(
                      padding: const EdgeInsets.only(left: 16, top: 4),
                      child: Text(
                        example,
                        style: const TextStyle(
                          fontFamily: 'monospace',
                          fontSize: 12,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                ],

                // Examples that don't match
                if (matchRules.exampleNonMatches.isNotEmpty) ...[
                  const Text(
                    '❌ Will NOT match:',
                    style: TextStyle(
                      color: Colors.red,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  ...matchRules.exampleNonMatches.map(
                    (example) => Padding(
                      padding: const EdgeInsets.only(left: 16, top: 4),
                      child: Text(
                        example,
                        style: TextStyle(
                          fontFamily: 'monospace',
                          fontSize: 12,
                          color: Colors.grey[600],
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        );
      },
    );
  }

  String _getPathPatternHint(PatternType type) {
    switch (type) {
      case PatternType.wildcard:
        return '/api/* or /users/*/profile';
      case PatternType.prefix:
        return '/api';
      case PatternType.exact:
        return '/api/users/123';
    }
  }

  String _getPathPatternHelp(PatternType type) {
    switch (type) {
      case PatternType.wildcard:
        return 'Use * to match any characters (e.g., /api/* matches /api/test, /api/users)';
      case PatternType.prefix:
        return 'Matches any path starting with this prefix';
      case PatternType.exact:
        return 'Matches only the exact path';
    }
  }
}
