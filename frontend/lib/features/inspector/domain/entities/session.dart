import 'package:equatable/equatable.dart';

class Session extends Equatable {
  const Session({
    required this.id,
    required this.target,
    this.clientAddr,
    this.startedAt,
    this.closedAt,
    this.error,
    this.kind,
    this.httpMeta,
    this.sizes,
  });
  final String id;
  final String target;
  final String? clientAddr;
  final DateTime? startedAt;
  final DateTime? closedAt;
  final String? error;
  final String? kind; // ws | http
  final Map<String, dynamic>? httpMeta; // server-provided
  final Map<String, dynamic>? sizes;

  @override
  List<Object?> get props => [id, target, clientAddr, startedAt, closedAt, error, kind, httpMeta, sizes];
}


