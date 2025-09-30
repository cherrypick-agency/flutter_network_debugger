import 'package:flutter/material.dart';
import 'dart:math' as math;

class FramesTimelineStyle {
  final Color axisColor;
  final Color textColor;
  final Color binaryColor;
  final Color pingColor;
  final Color pongColor;
  final Color closeColor;
  final Color msgsColor;
  final Color bytesColor;

  const FramesTimelineStyle({
    required this.axisColor,
    required this.textColor,
    required this.binaryColor,
    required this.pingColor,
    required this.pongColor,
    required this.closeColor,
    required this.msgsColor,
    required this.bytesColor,
  });
}

class FramesTimelinePainter extends CustomPainter {
  FramesTimelinePainter({
    required this.frames,
    required this.minTs,
    required this.maxTs,
    required this.style,
    required this.padding,
  });

  final List<Map<String, dynamic>> frames;
  final DateTime minTs;
  final DateTime maxTs;
  final FramesTimelineStyle style;
  final EdgeInsets padding;

  double _tsToX(DateTime ts, double width) {
    final total = maxTs.difference(minTs).inMilliseconds;
    final dx = ts.difference(minTs).inMilliseconds;
    if (total <= 0) return padding.left;
    final innerW = width - padding.horizontal;
    final t = dx.clamp(0, total) / total;
    return padding.left + innerW * t;
  }

  Color _colorForFrame(Map<String, dynamic> f) {
    final opcode = (f['opcode'] ?? '').toString();
    final preview = (f['preview'] ?? '').toString();
    final sizeStr = (f['size'] ?? '').toString();
    final isEnginePingPong = opcode == 'text' && (preview == '2' || preview == '3') && sizeStr == '1';
    if (opcode == 'pong' || isEnginePingPong) return style.pongColor;
    if (opcode == 'ping') return style.pingColor;
    if (opcode == 'binary') return style.binaryColor;
    if (opcode == 'close') return style.closeColor;
    if (opcode == 'text') return style.textColor;
    return style.textColor;
  }

  double _radiusForSize(dynamic sizeRaw) {
    final intSize = switch (sizeRaw) {
      int v => v,
      double v => v.toInt(),
      String s => int.tryParse(s) ?? 0,
      _ => 0,
    };
    // Чем больше размер, тем крупнее точка, но теперь всё вдвое меньше
    final base = 1.5 + (intSize <= 0 ? 0.0 : math.sqrt(intSize.toDouble()) / 6);
    final scaled = base * 0.5; // x0.5
    final double r = scaled.clamp(0.75, 3.0);
    return r;
  }

  @override
  void paint(Canvas canvas, Size size) {
    final axisPaint = Paint()
      ..color = style.axisColor
      ..strokeWidth = 1;

    final centerY = size.height / 2;
    final laneOffset = math.min(12.0, (size.height - padding.vertical) / 4);
    final topLaneY = centerY - laneOffset;
    final bottomLaneY = centerY + laneOffset;

    // Ось времени
    canvas.drawLine(
      Offset(padding.left, centerY),
      Offset(size.width - padding.right, centerY),
      axisPaint,
    );

    // Точки
    final pointPaint = Paint();
    for (final f in frames) {
      final tsStr = (f['ts'] ?? '').toString();
      DateTime? ts;
      try { ts = DateTime.parse(tsStr); } catch (_) {}
      if (ts == null) continue;

      final x = _tsToX(ts, size.width);
      final dir = (f['direction'] ?? '').toString();
      final y = dir == 'upstream->client' ? topLaneY : bottomLaneY;
      final color = _colorForFrame(f);
      final r = _radiusForSize(f['size']);

      pointPaint
        ..color = color
        ..style = PaintingStyle.fill;
      canvas.drawCircle(Offset(x, y), r, pointPaint);
    }

    // Спарклайны на той же оси: msgs/s (синяя) и bytes/s (зелёная)
    if (frames.isNotEmpty) {
      final innerW = size.width - padding.horizontal;
      if (innerW > 4) {
        const binsCount = 50;
        final msgs = List<double>.filled(binsCount, 0);
        final bytes = List<double>.filled(binsCount, 0);
        final totalMs = maxTs.difference(minTs).inMilliseconds;
        final bucketMs = math.max(1, (totalMs / binsCount).floor());
        for (final f in frames) {
          DateTime? ts;
          try { ts = DateTime.parse((f['ts'] ?? '').toString()); } catch (_) {}
          if (ts == null) continue;
          final idx = ((ts.millisecondsSinceEpoch - minTs.millisecondsSinceEpoch) / bucketMs).floor();
          if (idx < 0 || idx >= binsCount) continue;
          msgs[idx] += 1;
          final szRaw = f['size'];
          final sz = szRaw is int ? (szRaw.toDouble()) : (double.tryParse(szRaw?.toString() ?? '0') ?? 0.0);
          bytes[idx] += sz;
        }
        final maxMsgs = msgs.fold<double>(0, (p, n) => n > p ? n : p);
        final maxBytes = bytes.fold<double>(0, (p, n) => n > p ? n : p);
        final amp = math.min(10.0, (size.height - padding.vertical) / 4);
        if (maxMsgs > 0) {
          final path = Path();
          for (var i = 0; i < binsCount; i++) {
            final t = binsCount == 1 ? 0.0 : (i / (binsCount - 1));
            final x = padding.left + innerW * t;
            final y = centerY - (msgs[i] / maxMsgs) * amp - 1;
            if (i == 0) path.moveTo(x, y); else path.lineTo(x, y);
          }
          final p = Paint()
            ..style = PaintingStyle.stroke
            ..strokeWidth = 1.2
            ..color = style.msgsColor;
          canvas.drawPath(path, p);
        }
        if (maxBytes > 0) {
          final path = Path();
          for (var i = 0; i < binsCount; i++) {
            final t = binsCount == 1 ? 0.0 : (i / (binsCount - 1));
            final x = padding.left + innerW * t;
            final y = centerY - (bytes[i] / maxBytes) * amp - 3; // немного сместим для различимости
            if (i == 0) path.moveTo(x, y); else path.lineTo(x, y);
          }
          final p = Paint()
            ..style = PaintingStyle.stroke
            ..strokeWidth = 1.2
            ..color = style.bytesColor;
          canvas.drawPath(path, p);
        }
      }
    }
  }

  @override
  bool shouldRepaint(covariant FramesTimelinePainter oldDelegate) {
    return oldDelegate.frames != frames ||
        oldDelegate.minTs != minTs ||
        oldDelegate.maxTs != maxTs ||
        oldDelegate.style != style ||
        oldDelegate.padding != padding;
  }
}


