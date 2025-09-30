import 'package:flutter/material.dart';
import '../../../../../theme/context_ext.dart';

class FramesTimelineLegend extends StatelessWidget {
  const FramesTimelineLegend({super.key});

  @override
  Widget build(BuildContext context) {
    final c = context.appColors;
    final t = context.appText;
    final baseSize = (t.body.fontSize ?? 12);
    final legendTextStyle = t.body.copyWith(fontSize: baseSize * 0.5); // в 2 раза меньше
    Widget dot(Color color) => Container(width: 5, height: 5, decoration: BoxDecoration(color: color, shape: BoxShape.circle));
    return Row(
      children: [
        dot(t.body.color ?? c.textPrimary), const SizedBox(width: 4), Text('Text', style: legendTextStyle), const SizedBox(width: 8),
        dot(c.warning), const SizedBox(width: 4), Text('Binary/Ping', style: legendTextStyle), const SizedBox(width: 8),
        dot(c.success), const SizedBox(width: 4), Text('Pong', style: legendTextStyle), const SizedBox(width: 8),
        dot(c.danger), const SizedBox(width: 4), Text('Close', style: legendTextStyle),
      ],
    );
  }
}


