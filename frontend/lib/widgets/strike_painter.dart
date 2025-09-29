import 'package:flutter/material.dart';

class ChipStrikePainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final p = Paint()
      ..color = Colors.red.withOpacity(0.6)
      ..strokeWidth = 1.5;
    final y = size.height / 2;
    canvas.drawLine(Offset(4, y), Offset(size.width - 4, y), p);
  }

  @override
  bool shouldRepaint(covariant ChipStrikePainter oldDelegate) => false;
}
