import 'package:flutter/material.dart';
import '../models/byte_selection.dart';
import '../models/hex_config.dart';
import 'hex_viewer_widget.dart';

/// Controls panel for hex viewer (grouping, copy, etc.)
class HexControls extends StatelessWidget {
  final HexConfig config;
  final ByteSelection? selection;
  final void Function(CopyFormat format) onCopy;
  final ValueChanged<HexConfig> onConfigChanged;

  const HexControls({
    super.key,
    required this.config,
    required this.selection,
    required this.onCopy,
    required this.onConfigChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(8),
      child: Row(
        children: [
          // Grouping selector
          Text('Grouping:', style: Theme.of(context).textTheme.labelSmall),
          const SizedBox(width: 8),
          DropdownButton<int>(
            value: config.groupSize,
            isDense: true,
            underline: Container(),
            items: const [
              DropdownMenuItem(value: 0, child: Text('None')),
              DropdownMenuItem(value: 4, child: Text('4-4-4-4')),
              DropdownMenuItem(value: 8, child: Text('8-8')),
            ],
            onChanged: (value) {
              if (value != null) {
                onConfigChanged(config.copyWith(groupSize: value));
              }
            },
          ),

          const SizedBox(width: 16),

          // Bytes per line selector
          Text('Width:', style: Theme.of(context).textTheme.labelSmall),
          const SizedBox(width: 8),
          DropdownButton<int>(
            value: config.bytesPerLine,
            isDense: true,
            underline: Container(),
            items: const [
              DropdownMenuItem(value: 8, child: Text('8')),
              DropdownMenuItem(value: 16, child: Text('16')),
              DropdownMenuItem(value: 32, child: Text('32')),
            ],
            onChanged: (value) {
              if (value != null) {
                onConfigChanged(config.copyWith(bytesPerLine: value));
              }
            },
          ),

          const Spacer(),

          // Copy button with menu
          if (selection != null && config.enableCopy)
            MenuAnchor(
              builder: (context, controller, child) {
                return TextButton.icon(
                  onPressed: () {
                    if (controller.isOpen) {
                      controller.close();
                    } else {
                      controller.open();
                    }
                  },
                  icon: const Icon(Icons.copy, size: 16),
                  label: Text('Copy (${selection!.length} bytes)'),
                );
              },
              menuChildren: [
                MenuItemButton(
                  leadingIcon: const Icon(Icons.code, size: 16),
                  child: const Text('Copy as Hex'),
                  onPressed: () => onCopy(CopyFormat.hex),
                ),
                MenuItemButton(
                  leadingIcon: const Icon(Icons.text_fields, size: 16),
                  child: const Text('Copy as Text'),
                  onPressed: () => onCopy(CopyFormat.text),
                ),
                MenuItemButton(
                  leadingIcon: const Icon(Icons.storage, size: 16),
                  child: const Text('Copy as Binary'),
                  onPressed: () => onCopy(CopyFormat.binary),
                ),
              ],
            ),
        ],
      ),
    );
  }
}
