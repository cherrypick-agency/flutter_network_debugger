import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import '../../application/stores/script_editor_store.dart';
import '../../domain/entities/script.dart';
import '../../../compiler_management/presentation/stores/compiler_list_store.dart';
import '../../../../core/di/di.dart';

/// Form for script settings (name, description, priority, etc.)
class ScriptSettingsForm extends StatelessWidget {
  final ScriptEditorStore store;

  const ScriptSettingsForm({super.key, required this.store});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Name field
          Observer(
            builder: (_) => TextField(
              decoration: InputDecoration(
                labelText: 'Script Name',
                hintText:
                    'e.g., Add Authentication Header (auto-generated if empty)',
                errorText: store.nameError,
                border: const OutlineInputBorder(),
              ),
              controller: TextEditingController(text: store.name)
                ..selection = TextSelection.fromPosition(
                  TextPosition(offset: store.name.length),
                ),
              onChanged: store.setName,
            ),
          ),
          const SizedBox(height: 16),

          // Description field
          Observer(
            builder: (_) => TextField(
              decoration: const InputDecoration(
                labelText: 'Description',
                hintText: 'Optional description of what this script does',
                border: OutlineInputBorder(),
              ),
              controller: TextEditingController(text: store.description)
                ..selection = TextSelection.fromPosition(
                  TextPosition(offset: store.description.length),
                ),
              onChanged: store.setDescription,
              maxLines: 3,
            ),
          ),
          const SizedBox(height: 24),

          // Runtime selector
          Text('Runtime', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          Observer(
            builder: (_) => SegmentedButton<ScriptRuntime>(
              segments: const [
                ButtonSegment(
                  value: ScriptRuntime.extism,
                  label: Text('Extism (WASM)'),
                  icon: Icon(Icons.memory),
                ),
                ButtonSegment(
                  value: ScriptRuntime.dart,
                  label: Text('Dart Subprocess'),
                  icon: Icon(Icons.code),
                ),
              ],
              selected: {store.runtime},
              onSelectionChanged: (Set<ScriptRuntime> selected) {
                store.setRuntime(selected.first);
              },
            ),
          ),
          const SizedBox(height: 24),

          // Language selector (for Extism)
          if (store.runtime == ScriptRuntime.extism) ...[
            Text('Language', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Observer(
              builder: (_) => Wrap(
                spacing: 8,
                children: [
                  ChoiceChip(
                    label: const Text('🦀 Rust'),
                    selected: store.language == 'rust',
                    onSelected: (_) => store.setLanguage('rust'),
                  ),
                  ChoiceChip(
                    label: const Text('🟨 JavaScript'),
                    selected: store.language == 'javascript',
                    onSelected: (_) => store.setLanguage('javascript'),
                  ),
                  ChoiceChip(
                    label: const Text('🔷 TypeScript'),
                    selected: store.language == 'typescript',
                    onSelected: (_) => store.setLanguage('typescript'),
                  ),
                  ChoiceChip(
                    label: const Text('🐹 Go'),
                    selected: store.language == 'go',
                    onSelected: (_) => store.setLanguage('go'),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            // Compiler status banner
            _buildCompilerStatusBanner(context, store.language),
            const SizedBox(height: 24),
          ],

          // Trigger Type
          Text(
            'When to Execute',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          Observer(
            builder: (_) => Column(
              children: [
                RadioListTile<TriggerType>(
                  title: const Text('Before Request'),
                  subtitle: const Text(
                    'Modify request before sending to upstream',
                  ),
                  value: TriggerType.request,
                  groupValue: store.triggerType,
                  onChanged: (value) => store.setTriggerType(value!),
                ),
                RadioListTile<TriggerType>(
                  title: const Text('After Response'),
                  subtitle: const Text(
                    'Modify response before sending to client',
                  ),
                  value: TriggerType.response,
                  groupValue: store.triggerType,
                  onChanged: (value) => store.setTriggerType(value!),
                ),
                RadioListTile<TriggerType>(
                  title: const Text('Both'),
                  subtitle: const Text('Execute on both request and response'),
                  value: TriggerType.both,
                  groupValue: store.triggerType,
                  onChanged: (value) => store.setTriggerType(value!),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          // Priority slider
          Text(
            'Priority (${store.priority})',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const Text(
            'Higher priority scripts execute first',
            style: TextStyle(color: Colors.grey, fontSize: 12),
          ),
          const SizedBox(height: 8),
          Observer(
            builder: (_) => Slider(
              value: store.priority.toDouble(),
              min: 0,
              max: 100,
              divisions: 20,
              label: store.priority.toString(),
              onChanged: (value) => store.setPriority(value.toInt()),
            ),
          ),
          const SizedBox(height: 24),

          // Enabled toggle
          Observer(
            builder: (_) => SwitchListTile(
              title: const Text('Enabled'),
              subtitle: const Text(
                'Script will execute automatically when enabled',
              ),
              value: store.enabled,
              onChanged: store.setEnabled,
            ),
          ),
          const SizedBox(height: 24),

          // Advanced Configuration
          ExpansionTile(
            title: const Text('Advanced Configuration'),
            initiallyExpanded: false,
            children: [
              const SizedBox(height: 8),
              // Timeout
              Observer(
                builder: (_) => TextField(
                  decoration: const InputDecoration(
                    labelText: 'Timeout (ms)',
                    hintText: 'Default: 5000',
                    border: OutlineInputBorder(),
                    suffixText: 'ms',
                  ),
                  controller:
                      TextEditingController(
                          text: store.timeoutMs?.toString() ?? '',
                        )
                        ..selection = TextSelection.fromPosition(
                          TextPosition(
                            offset: (store.timeoutMs?.toString() ?? '').length,
                          ),
                        ),
                  keyboardType: TextInputType.number,
                  onChanged: (value) {
                    final intValue = int.tryParse(value);
                    store.setTimeoutMs(intValue);
                  },
                ),
              ),
              const SizedBox(height: 16),
              // Memory Limit
              Observer(
                builder: (_) => TextField(
                  decoration: const InputDecoration(
                    labelText: 'Memory Limit (MB)',
                    hintText: 'Default: 128',
                    border: OutlineInputBorder(),
                    suffixText: 'MB',
                  ),
                  controller:
                      TextEditingController(
                          text: store.memoryLimitMB?.toString() ?? '',
                        )
                        ..selection = TextSelection.fromPosition(
                          TextPosition(
                            offset:
                                (store.memoryLimitMB?.toString() ?? '').length,
                          ),
                        ),
                  keyboardType: TextInputType.number,
                  onChanged: (value) {
                    final intValue = int.tryParse(value);
                    store.setMemoryLimitMB(intValue);
                  },
                ),
              ),
              const SizedBox(height: 16),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildCompilerStatusBanner(BuildContext context, String language) {
    // Get compiler store from DI
    final compilerStore = sl<CompilerListStore>();

    // Map language to compiler name
    final compilerName = _getCompilerForLanguage(language);

    if (compilerName == null) {
      return const SizedBox.shrink(); // No compiler needed for Dart
    }

    return Observer(
      builder: (_) {
        // Find compiler info
        final compiler = compilerStore.compilers.cast().firstWhere(
          (c) => c?.language.toLowerCase() == compilerName.toLowerCase(),
          orElse: () => null,
        );

        if (compiler == null) {
          return const SizedBox.shrink();
        }

        final theme = Theme.of(context);
        final colorScheme = theme.colorScheme;

        // Determine banner color and message
        final isInstalled = compiler.isInstalled;
        final backgroundColor = isInstalled
            ? colorScheme.primaryContainer
            : colorScheme.errorContainer;
        final textColor = isInstalled
            ? colorScheme.onPrimaryContainer
            : colorScheme.onErrorContainer;
        final icon = isInstalled ? Icons.check_circle : Icons.warning_amber;
        final message = isInstalled
            ? '$compilerName compiler is installed'
            : '$compilerName compiler required - Install to compile your code';

        return Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: backgroundColor,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            children: [
              Icon(icon, color: textColor, size: 20),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  message,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: textColor,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
              if (!isInstalled) ...[
                const SizedBox(width: 8),
                TextButton(
                  onPressed: () {
                    Navigator.of(context).pushNamed('/compilers');
                  },
                  style: TextButton.styleFrom(
                    foregroundColor: textColor,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 4,
                    ),
                  ),
                  child: const Text('Install'),
                ),
              ],
            ],
          ),
        );
      },
    );
  }

  String? _getCompilerForLanguage(String language) {
    switch (language.toLowerCase()) {
      case 'rust':
        return 'rust';
      case 'go':
        return 'tinygo';
      case 'javascript':
      case 'typescript':
        return 'assemblyscript';
      case 'zig':
        return 'zig';
      case 'kotlin':
        return 'kotlin';
      case 'swift':
        return 'swift';
      case 'c':
      case 'cpp':
      case 'c++':
        return 'wasi-sdk';
      case 'dart':
        return null; // Dart doesn't need compilation
      default:
        return null;
    }
  }
}
