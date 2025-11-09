import 'package:equatable/equatable.dart';

class ProcessInfo extends Equatable {
  const ProcessInfo({
    required this.pid,
    required this.name,
    this.executablePath,
    this.bundleId,
    this.icon,
    this.detectedAt,
  });

  final int pid;
  final String name;
  final String? executablePath;
  final String? bundleId;
  final ProcessIcon? icon;
  final DateTime? detectedAt;

  @override
  List<Object?> get props => [
    pid,
    name,
    executablePath,
    bundleId,
    icon,
    detectedAt,
  ];
}

class ProcessIcon extends Equatable {
  const ProcessIcon({required this.format, required this.data});

  final String format; // png, svg, etc.
  final String data; // base64 encoded

  @override
  List<Object?> get props => [format, data];
}

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
    this.isSocketIo = false,
    this.processInfo,
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
  // Признак, что в сессии присутствуют Socket.IO события (по данным бэка)
  final bool isSocketIo;
  // Информация о процессе, который создал эту сессию
  final ProcessInfo? processInfo;

  @override
  List<Object?> get props => [
    id,
    target,
    clientAddr,
    startedAt,
    closedAt,
    error,
    kind,
    httpMeta,
    sizes,
    isSocketIo,
    processInfo,
  ];
}
