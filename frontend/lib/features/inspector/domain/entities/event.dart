import 'package:equatable/equatable.dart';

class EventEntity extends Equatable {
  const EventEntity({required this.id, required this.ts, required this.namespace, required this.event, this.ackId, required this.argsPreview});
  final String id;
  final DateTime ts;
  final String namespace;
  final String event;
  final int? ackId;
  final String argsPreview;
  @override
  List<Object?> get props => [id, ts, namespace, event, ackId, argsPreview];
}


