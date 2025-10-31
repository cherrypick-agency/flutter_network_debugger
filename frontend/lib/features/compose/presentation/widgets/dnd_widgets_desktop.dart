// ignore_for_file: uri_does_not_exist, undefined_identifier, undefined_class
import 'package:flutter/widgets.dart';
import 'package:desktop_drop/desktop_drop.dart';
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
    return DropTarget(
      onDragEntered: (_) {
        try {
          onHoverChanged?.call(true);
        } catch (_) {}
      },
      onDragExited: (_) {
        try {
          onHoverChanged?.call(false);
        } catch (_) {}
      },
      onDragDone: (details) async {
        try {
          if (details.files.isNotEmpty) {
            if (onDropMany != null) {
              final list = <DndFile>[];
              for (final f in details.files) {
                final bytes = await f.readAsBytes();
                list.add(DndFile(bytes: bytes, filename: f.name));
              }
              onDropMany!(list);
            } else {
              final f = details.files.first;
              final bytes = await f.readAsBytes();
              onDrop(bytes, f.name);
            }
          }
        } catch (_) {}
        try {
          onHoverChanged?.call(false);
        } catch (_) {}
      },
      child: const SizedBox.expand(),
    );
  }
}
