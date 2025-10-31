import 'package:flutter/widgets.dart';
import 'dnd_types.dart';

class DndDropArea extends StatelessWidget {
  final void Function(List<int> bytes, String filename) onDrop;
  final ValueChanged<bool>? onHoverChanged;
  final void Function(List<DndFile> files)? onDropMany;
  const DndDropArea({
    super.key,
    required this.onDrop,
    this.onHoverChanged,
    this.onDropMany,
  });
  @override
  Widget build(BuildContext context) {
    // Пустышка: ничего не делает
    return const SizedBox.shrink();
  }
}
