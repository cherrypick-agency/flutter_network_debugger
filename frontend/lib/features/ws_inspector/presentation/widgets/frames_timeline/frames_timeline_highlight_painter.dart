import 'package:flutter/material.dart';
import 'dart:math' as math;

class FramesTimelineHighlightPainter extends CustomPainter {
  FramesTimelineHighlightPainter({
    required this.hoverFrame,
    required this.minTs,
    required this.maxTs,
    required this.padding,
    required this.color,
  });

  final Map<String, dynamic>? hoverFrame;
  final DateTime minTs;
  final DateTime maxTs;
  final EdgeInsets padding;
  final Color color;

  double _tsToX(DateTime ts, double width) {
    final total = maxTs.difference(minTs).inMilliseconds;
    final dx = ts.difference(minTs).inMilliseconds;
    if (total <= 0) return padding.left;
    final innerW = width - padding.horizontal;
    final t = dx.clamp(0, total) / total;
    return padding.left + innerW * t;
  }

  double _radiusForSize(dynamic sizeRaw) {
    final intSize = switch (sizeRaw) {
      int v => v,
      double v => v.toInt(),
      String s => int.tryParse(s) ?? 0,
      _ => 0,
    };
    final r = 1.5 + (intSize <= 0 ? 0.0 : math.sqrt(intSize.toDouble()) / 6);
    if (r < 1.5) return 1.5;
    if (r > 6.0) return 6.0;
    return r;
  }

  @override
  void paint(Canvas canvas, Size size) {
    final f = hoverFrame;
    if (f == null) return;
    DateTime? ts;
    try { ts = DateTime.parse((f['ts'] ?? '').toString()); } catch (_) {}
    if (ts == null) return;

    final x = _tsToX(ts, size.width);
    final centerY = size.height / 2;
    final laneOffset = math.min(12.0, (size.height - padding.vertical) / 4);
    final dir = (f['direction'] ?? '').toString();
    final y = dir == 'upstream->client' ? (centerY - laneOffset) : (centerY + laneOffset);
    final r = _radiusForSize(f['size']);

    final paint = Paint()
      ..color = color.withOpacity(1.0)
      ..style = PaintingStyle.fill;
    canvas.drawCircle(Offset(x, y), r, paint);
  }

  @override
  bool shouldRepaint(covariant FramesTimelineHighlightPainter oldDelegate) {
    return oldDelegate.hoverFrame != hoverFrame ||
        oldDelegate.minTs != minTs ||
        oldDelegate.maxTs != maxTs ||
        oldDelegate.padding != padding ||
        oldDelegate.color != color;
  }
}


