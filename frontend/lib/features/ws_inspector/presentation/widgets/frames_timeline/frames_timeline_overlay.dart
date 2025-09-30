import 'package:flutter/material.dart';
import '../../../../../theme/context_ext.dart';

class FramesTimelineOverlay extends StatelessWidget {
  const FramesTimelineOverlay({
    super.key,
    required this.position,
    required this.frame,
  });

    // Позиция курсора/тапа в локальных координатах таймлайна
  final Offset position;
  // Данные фрейма (минимум: id, ts, direction, opcode, size, preview?)
  final Map<String, dynamic> frame;

  @override
  Widget build(BuildContext context) {
    final colors = context.appColors;
    final txt = context.appText;
    final opcode = (frame['opcode'] ?? '').toString();
    final dir = (frame['direction'] ?? '').toString();
    final size = (frame['size'] ?? '').toString();
    final ts = (frame['ts'] ?? '').toString();
    final preview = (frame['preview'] ?? '').toString();

    final theme = Theme.of(context);
    final bg = theme.colorScheme.surface.withOpacity(0.95);
    final border = colors.border;

    return Positioned(
      left: position.dx + 8,
      top: (position.dy - 44).clamp(0, double.infinity),
      child: Material(
        color: Colors.transparent,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
          decoration: BoxDecoration(
            color: bg,
            border: Border.all(color: border),
            borderRadius: BorderRadius.circular(6),
            boxShadow: [
              BoxShadow(color: Colors.black.withOpacity(0.08), blurRadius: 8, offset: const Offset(0, 2)),
            ],
          ),
          constraints: const BoxConstraints(maxWidth: 320),
          child: DefaultTextStyle(
            style: txt.body.copyWith(color: colors.textPrimary),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text('$dir  ·  $opcode  ·  $size B', style: txt.subtitle),
                const SizedBox(height: 4),
                Text(ts, style: txt.body.copyWith(color: colors.textSecondary)),
                if (preview.isNotEmpty) ...[
                  const SizedBox(height: 6),
                  Text(preview, maxLines: 2, overflow: TextOverflow.ellipsis, style: txt.monospace.copyWith(color: colors.textSecondary)),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}


