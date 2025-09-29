import 'package:equatable/equatable.dart';

class Frame extends Equatable {
  const Frame({required this.id, required this.ts, required this.direction, required this.opcode, required this.size, required this.preview});
  final String id;
  final DateTime ts;
  final String direction;
  final String opcode;
  final int size;
  final String preview;
  @override
  List<Object?> get props => [id, ts, direction, opcode, size, preview];
}


