// ignore_for_file: uri_does_not_exist, undefined_identifier, undefined_class
import 'package:flutter/widgets.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'dnd_types.dart';

class DndDropArea extends StatefulWidget {
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
  State<DndDropArea> createState() => _DndDropAreaState();
}

class _DndDropAreaState extends State<DndDropArea> {
  DropzoneViewController? _ctrl;
  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        const SizedBox.expand(),
        DropzoneView(
          onCreated: (c) {
            _ctrl = c;
          },
          onHover: () {
            try {
              widget.onHoverChanged?.call(true);
            } catch (_) {}
          },
          onLeave: () {
            try {
              widget.onHoverChanged?.call(false);
            } catch (_) {}
          },
          onDrop: (ev) async {
            try {
              final c = _ctrl;
              if (c == null) return;
              final name = await c.getFilename(ev);
              final data = await c.getFileData(ev);
              widget.onDrop(data, name);
              try {
                widget.onHoverChanged?.call(false);
              } catch (_) {}
            } catch (_) {}
          },
          onDropMultiple: (evs) async {
            try {
              final c = _ctrl;
              if (c == null) return;
              final list = <DndFile>[];
              if (evs == null) return;
              for (final ev in evs) {
                final name = await c.getFilename(ev);
                final data = await c.getFileData(ev);
                list.add(DndFile(bytes: data, filename: name));
              }
              try {
                widget.onDropMany?.call(list);
              } catch (_) {}
              try {
                widget.onHoverChanged?.call(false);
              } catch (_) {}
            } catch (_) {}
          },
        ),
      ],
    );
  }
}
