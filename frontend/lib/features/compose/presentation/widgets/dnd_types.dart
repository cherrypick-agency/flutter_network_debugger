import 'package:flutter/foundation.dart';

@immutable
class DndFile {
  final List<int> bytes;
  final String filename;
  const DndFile({required this.bytes, required this.filename});
}
