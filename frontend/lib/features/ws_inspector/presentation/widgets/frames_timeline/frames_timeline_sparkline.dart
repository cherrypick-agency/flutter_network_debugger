import 'package:flutter/material.dart';
import 'dart:math' as math;
import '../../../../../theme/context_ext.dart';

class FramesTimelineSparkline extends StatelessWidget {
  const FramesTimelineSparkline({
    super.key,
    required this.frames,
    required this.minTs,
    required this.maxTs,
    this.height = 28,
    this.metric = SparkMetric.messagesPerSecond,
  });

  final List<Map<String, dynamic>> frames;
  final DateTime minTs;
  final DateTime maxTs;
  final double height;
  final SparkMetric metric;

  @override
  Widget build(BuildContext context) {
    final bins = _buildBins();
    final maxVal = bins.isEmpty ? 0.0 : bins.map((e) => e.value).reduce(math.max);
    final color = switch (metric) {
      SparkMetric.messagesPerSecond => context.appColors.primary,
      SparkMetric.bytesPerSecond => context.appColors.success,
    };
    return SizedBox(
      height: height,
      width: double.infinity,
      child: CustomPaint(
        painter: _SparklinePainter(bins: bins, maxVal: maxVal, color: color),
      ),
    );
  }

  List<_Bin> _buildBins() {
    final totalMs = maxTs.difference(minTs).inMilliseconds;
    final bucketMs = math.max(1000, (totalMs / 50).round()); // ~50 точек
    final bins = <int, double>{};
    for (final f in frames) {
      DateTime? ts;
      try { ts = DateTime.parse((f['ts'] ?? '').toString()); } catch (_) {}
      if (ts == null) continue;
      final idx = ((ts.millisecondsSinceEpoch - minTs.millisecondsSinceEpoch) / bucketMs).floor();
      if (!bins.containsKey(idx)) bins[idx] = 0;
      switch (metric) {
        case SparkMetric.messagesPerSecond:
          bins[idx] = (bins[idx] ?? 0) + 1;
          break;
        case SparkMetric.bytesPerSecond:
          final szRaw = f['size'];
          final sz = szRaw is int ? szRaw.toDouble() : double.tryParse(szRaw?.toString() ?? '0') ?? 0.0;
          bins[idx] = (bins[idx] ?? 0) + sz;
          break;
      }
    }
    return bins.entries.map((e) => _Bin(index: e.key, value: e.value)).toList()..sort((a,b)=>a.index.compareTo(b.index));
  }
}

enum SparkMetric { messagesPerSecond, bytesPerSecond }

class _Bin {
  final int index;
  final double value;
  _Bin({required this.index, required this.value});
}

class _SparklinePainter extends CustomPainter {
  _SparklinePainter({required this.bins, required this.maxVal, required this.color});
  final List<_Bin> bins;
  final double maxVal;
  final Color color;

  @override
  void paint(Canvas canvas, Size size) {
    if (bins.isEmpty || maxVal <= 0) return;
    final paint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.5
      ..color = color;
    final path = Path();
    final dx = size.width / math.max(1, bins.length - 1);
    for (var i = 0; i < bins.length; i++) {
      final x = i * dx;
      final v = bins[i].value / maxVal;
      final y = size.height - v * (size.height - 2) - 1;
      if (i == 0) path.moveTo(x, y); else path.lineTo(x, y);
    }
    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(covariant _SparklinePainter oldDelegate) {
    return oldDelegate.bins != bins || oldDelegate.maxVal != maxVal || oldDelegate.color != color;
  }
}


