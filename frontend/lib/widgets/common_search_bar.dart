import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

// Универсальный SearchBar с навигацией и переключателями
class CommonSearchBar extends StatelessWidget {
  const CommonSearchBar({
    super.key,
    required this.controller,
    this.focusNode,
    required this.countText,
    required this.matchCase,
    required this.wholeWord,
    required this.useRegex,
    required this.canNavigate,
    required this.onChanged,
    required this.onNext,
    required this.onPrev,
    required this.onClose,
    required this.onToggleMatchCase,
    required this.onToggleWholeWord,
    required this.onToggleRegex,
  });

  final TextEditingController controller;
  final FocusNode? focusNode;
  final String countText;
  final bool matchCase;
  final bool wholeWord;
  final bool useRegex;
  final bool canNavigate;
  final VoidCallback onChanged;
  final VoidCallback onNext;
  final VoidCallback onPrev;
  final VoidCallback onClose;
  final VoidCallback onToggleMatchCase;
  final VoidCallback onToggleWholeWord;
  final VoidCallback onToggleRegex;

  @override
  Widget build(BuildContext context) {
    final surface = Theme.of(context).colorScheme.surfaceVariant;
    final onSurface = Theme.of(context).colorScheme.onSurfaceVariant;
    final primary = Theme.of(context).colorScheme.primary;
    final textStyle =
        Theme.of(context).textTheme.labelSmall ?? const TextStyle(fontSize: 12);

    return Shortcuts(
      shortcuts: <LogicalKeySet, Intent>{
        LogicalKeySet(LogicalKeyboardKey.enter): const _NextIntent(),
        LogicalKeySet(LogicalKeyboardKey.shift, LogicalKeyboardKey.enter):
            const _PrevIntent(),
        LogicalKeySet(LogicalKeyboardKey.escape): const _CloseIntent(),
      },
      child: Actions(
        actions: <Type, Action<Intent>>{
          _NextIntent: CallbackAction<_NextIntent>(
            onInvoke: (_) {
              if (canNavigate) onNext();
              return null;
            },
          ),
          _PrevIntent: CallbackAction<_PrevIntent>(
            onInvoke: (_) {
              if (canNavigate) onPrev();
              return null;
            },
          ),
          _CloseIntent: CallbackAction<_CloseIntent>(
            onInvoke: (_) {
              onClose();
              return null;
            },
          ),
        },
        child: Focus(
          autofocus: true,
          child: Material(
            elevation: 1,
            borderRadius: BorderRadius.circular(8),
            color: surface.withValues(alpha: 0.75),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
              // Сужаемся по доступной ширине родителя, но не шире 480
              constraints: const BoxConstraints(maxWidth: 480),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.search, size: 16),
                  const SizedBox(width: 4),
                  Expanded(
                    child: SizedBox(
                      height: 20,
                      child: TextField(
                        controller: controller,
                        autofocus: true,
                        focusNode: focusNode,
                        style: textStyle,
                        decoration: InputDecoration(
                          isDense: true,
                          hintText: 'Search',
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(6),
                          ),
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 6,
                          ),
                        ),
                        onChanged: (_) => onChanged(),
                        onSubmitted: (_) => onNext(),
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  Text(countText, style: textStyle.copyWith(color: onSurface)),
                  const SizedBox(width: 4),
                  IconButton(
                    visualDensity: VisualDensity.compact,
                    constraints: const BoxConstraints(
                      minWidth: 28,
                      minHeight: 28,
                    ),
                    padding: const EdgeInsets.all(2),
                    tooltip: 'Previous match',
                    onPressed: canNavigate ? onPrev : null,
                    icon: const Icon(Icons.keyboard_arrow_up, size: 18),
                  ),
                  IconButton(
                    visualDensity: VisualDensity.compact,
                    constraints: const BoxConstraints(
                      minWidth: 28,
                      minHeight: 28,
                    ),
                    padding: const EdgeInsets.all(2),
                    tooltip: 'Next match',
                    onPressed: canNavigate ? onNext : null,
                    icon: const Icon(Icons.keyboard_arrow_down, size: 18),
                  ),
                  const SizedBox(width: 4),
                  IconButton(
                    visualDensity: VisualDensity.compact,
                    constraints: const BoxConstraints(
                      minWidth: 28,
                      minHeight: 28,
                    ),
                    padding: const EdgeInsets.all(2),
                    tooltip: 'Match case',
                    onPressed: onToggleMatchCase,
                    icon: Icon(
                      Icons.abc,
                      size: 18,
                      color: matchCase ? primary : onSurface,
                    ),
                  ),
                  IconButton(
                    visualDensity: VisualDensity.compact,
                    constraints: const BoxConstraints(
                      minWidth: 28,
                      minHeight: 28,
                    ),
                    padding: const EdgeInsets.all(2),
                    tooltip: 'Match whole word',
                    onPressed: onToggleWholeWord,
                    icon: Icon(
                      Icons.format_shapes,
                      size: 18,
                      color: wholeWord ? primary : onSurface,
                    ),
                  ),
                  IconButton(
                    visualDensity: VisualDensity.compact,
                    constraints: const BoxConstraints(
                      minWidth: 28,
                      minHeight: 28,
                    ),
                    padding: const EdgeInsets.all(2),
                    tooltip: 'Use regular expression',
                    onPressed: onToggleRegex,
                    icon: Icon(
                      Icons.pattern,
                      size: 18,
                      color: useRegex ? primary : onSurface,
                    ),
                  ),
                  const SizedBox(width: 2),
                  IconButton(
                    visualDensity: VisualDensity.compact,
                    constraints: const BoxConstraints(
                      minWidth: 28,
                      minHeight: 28,
                    ),
                    padding: const EdgeInsets.all(2),
                    tooltip: 'Close',
                    onPressed: onClose,
                    icon: const Icon(Icons.close, size: 18),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _NextIntent extends Intent {
  const _NextIntent();
}

class _PrevIntent extends Intent {
  const _PrevIntent();
}

class _CloseIntent extends Intent {
  const _CloseIntent();
}
