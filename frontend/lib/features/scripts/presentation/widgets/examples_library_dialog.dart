import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../domain/entities/script_example.dart';
import '../../domain/entities/script.dart';
import '../../application/stores/scripts_store.dart';
import '../../data/services/scripts_api_service.dart';
import '../../../../core/di/di.dart';

class ExamplesLibraryDialog extends StatefulWidget {
  final ScriptsStore scriptsStore;

  const ExamplesLibraryDialog({super.key, required this.scriptsStore});

  static Future<void> show(BuildContext context, ScriptsStore store) async {
    return showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => ExamplesLibraryDialog(scriptsStore: store),
    );
  }

  @override
  State<ExamplesLibraryDialog> createState() => _ExamplesLibraryDialogState();
}

class _ExamplesLibraryDialogState extends State<ExamplesLibraryDialog> {
  final _apiService = sl<ScriptsApiService>();
  List<ScriptExample> _examples = [];
  bool _isLoading = true;
  String? _errorMessage;
  ExampleDifficulty? _difficultyFilter;
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    _loadExamples();
  }

  Future<void> _loadExamples() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final data = await _apiService.getExamples();
      final examples = data
          .map((json) => ScriptExample.fromJson(json))
          .toList();

      setState(() {
        _examples = examples;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = 'Failed to load examples: ${e.toString()}';
        _isLoading = false;
      });
    }
  }

  List<ScriptExample> get _filteredExamples {
    var filtered = _examples;

    if (_difficultyFilter != null) {
      filtered = filtered
          .where((e) => e.difficulty == _difficultyFilter)
          .toList();
    }

    if (_searchQuery.isNotEmpty) {
      final query = _searchQuery.toLowerCase();
      filtered = filtered
          .where(
            (e) =>
                e.name.toLowerCase().contains(query) ||
                e.description.toLowerCase().contains(query),
          )
          .toList();
    }

    return filtered;
  }

  Future<void> _importExample(ScriptExample example) async {
    try {
      TriggerType triggerType;
      switch (example.triggerType) {
        case 'request':
          triggerType = TriggerType.request;
          break;
        case 'response':
          triggerType = TriggerType.response;
          break;
        case 'both':
          triggerType = TriggerType.both;
          break;
        default:
          triggerType = TriggerType.request;
      }

      final newScript = Script(
        id: '',
        name: example.name,
        description: example.description,
        runtime: ScriptRuntime.extism,
        code: '',
        language: example.language,
        triggerType: triggerType,
        priority: 10,
        enabled: false,
        sourceCode: example.sourceCode,
      );

      await widget.scriptsStore.createScript(newScript);

      if (!mounted) return;

      Navigator.of(context).pop();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Example "${example.name}" imported successfully!'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      if (!mounted) return;

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to import example: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  void _copyToClipboard(String code) {
    Clipboard.setData(ClipboardData(text: code));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Code copied to clipboard'),
        duration: Duration(seconds: 2),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Dialog.fullscreen(
      child: Scaffold(
        appBar: AppBar(
          title: const Row(
            children: [
              Icon(Icons.library_books),
              SizedBox(width: 12),
              Text('Example Scripts Library'),
            ],
          ),
          leading: IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => Navigator.of(context).pop(),
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: _loadExamples,
              tooltip: 'Reload Examples',
            ),
          ],
        ),
        body: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                children: [
                  TextField(
                    decoration: const InputDecoration(
                      hintText: 'Search examples...',
                      prefixIcon: Icon(Icons.search),
                      border: OutlineInputBorder(),
                    ),
                    onChanged: (value) {
                      setState(() => _searchQuery = value);
                    },
                  ),
                  const SizedBox(height: 12),
                  SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(
                      children: [
                        ChoiceChip(
                          label: const Text('All Levels'),
                          selected: _difficultyFilter == null,
                          onSelected: (selected) {
                            setState(() => _difficultyFilter = null);
                          },
                        ),
                        const SizedBox(width: 8),
                        ChoiceChip(
                          label: const Text('Beginner'),
                          selected:
                              _difficultyFilter == ExampleDifficulty.beginner,
                          onSelected: (selected) {
                            setState(
                              () => _difficultyFilter = selected
                                  ? ExampleDifficulty.beginner
                                  : null,
                            );
                          },
                        ),
                        const SizedBox(width: 8),
                        ChoiceChip(
                          label: const Text('Intermediate'),
                          selected:
                              _difficultyFilter ==
                              ExampleDifficulty.intermediate,
                          onSelected: (selected) {
                            setState(
                              () => _difficultyFilter = selected
                                  ? ExampleDifficulty.intermediate
                                  : null,
                            );
                          },
                        ),
                        const SizedBox(width: 8),
                        ChoiceChip(
                          label: const Text('Advanced'),
                          selected:
                              _difficultyFilter == ExampleDifficulty.advanced,
                          onSelected: (selected) {
                            setState(
                              () => _difficultyFilter = selected
                                  ? ExampleDifficulty.advanced
                                  : null,
                            );
                          },
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            Expanded(child: _buildBody()),
          ],
        ),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 64, color: Colors.red),
            const SizedBox(height: 16),
            Text(_errorMessage!, style: const TextStyle(color: Colors.red)),
            const SizedBox(height: 16),
            ElevatedButton.icon(
              onPressed: _loadExamples,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    final filtered = _filteredExamples;

    if (filtered.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.search_off, size: 64, color: Colors.grey),
            SizedBox(height: 16),
            Text('No examples found'),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: filtered.length,
      itemBuilder: (context, index) {
        final example = filtered[index];
        return _ExampleCard(
          example: example,
          onImport: () => _importExample(example),
          onCopyCode: () => _copyToClipboard(example.sourceCode),
        );
      },
    );
  }
}

class _ExampleCard extends StatelessWidget {
  final ScriptExample example;
  final VoidCallback onImport;
  final VoidCallback onCopyCode;

  const _ExampleCard({
    required this.example,
    required this.onImport,
    required this.onCopyCode,
  });

  Color _getDifficultyColor() {
    switch (example.difficulty) {
      case ExampleDifficulty.beginner:
        return Colors.green;
      case ExampleDifficulty.intermediate:
        return Colors.orange;
      case ExampleDifficulty.advanced:
        return Colors.red;
    }
  }

  String _getDifficultyLabel() {
    switch (example.difficulty) {
      case ExampleDifficulty.beginner:
        return 'Beginner';
      case ExampleDifficulty.intermediate:
        return 'Intermediate';
      case ExampleDifficulty.advanced:
        return 'Advanced';
    }
  }

  String _getCategoryLabel() {
    switch (example.category) {
      case ExampleCategory.requestModification:
        return 'Request Modification';
      case ExampleCategory.responseModification:
        return 'Response Modification';
      case ExampleCategory.both:
        return 'Both (Request + Response)';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    example.name,
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                ),
                Chip(
                  label: Text(
                    _getDifficultyLabel(),
                    style: const TextStyle(color: Colors.white, fontSize: 12),
                  ),
                  backgroundColor: _getDifficultyColor(),
                  padding: EdgeInsets.zero,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              example.description,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                Chip(
                  avatar: const Icon(Icons.code, size: 16),
                  label: Text(example.language.toUpperCase()),
                  labelStyle: const TextStyle(fontSize: 12),
                ),
                Chip(
                  avatar: const Icon(Icons.category, size: 16),
                  label: Text(_getCategoryLabel()),
                  labelStyle: const TextStyle(fontSize: 12),
                ),
                Chip(
                  avatar: const Icon(Icons.flash_on, size: 16),
                  label: Text('Trigger: ${example.triggerType}'),
                  labelStyle: const TextStyle(fontSize: 12),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.grey[900],
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                example.sourceCode.length > 200
                    ? '${example.sourceCode.substring(0, 200)}...'
                    : example.sourceCode,
                style: const TextStyle(
                  fontFamily: 'monospace',
                  fontSize: 12,
                  color: Colors.green,
                ),
                maxLines: 8,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                ElevatedButton.icon(
                  onPressed: onImport,
                  icon: const Icon(Icons.add),
                  label: const Text('Use This Template'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blue,
                    foregroundColor: Colors.white,
                  ),
                ),
                const SizedBox(width: 8),
                OutlinedButton.icon(
                  onPressed: onCopyCode,
                  icon: const Icon(Icons.copy),
                  label: const Text('Copy Code'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
