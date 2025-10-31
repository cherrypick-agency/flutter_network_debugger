import 'package:app_http_client/application/app_http_client.dart';
import '../../../core/di/di.dart';

class ThrottleProfileDTO {
  final String? id;
  final String name;
  final int downKbps;
  final int upKbps;
  final int packetLossPct;
  final int latencyMs;
  final int latencyJitter;
  final bool offline;

  const ThrottleProfileDTO({
    this.id,
    required this.name,
    required this.downKbps,
    required this.upKbps,
    required this.packetLossPct,
    this.latencyMs = 0,
    this.latencyJitter = 0,
    required this.offline,
  });

  Map<String, dynamic> toJson() => {
    if (id != null) 'id': id,
    'name': name,
    'downKbps': downKbps,
    'upKbps': upKbps,
    'packetLossPct': packetLossPct,
    'latencyMs': latencyMs,
    'latencyJitter': latencyJitter,
    'offline': offline,
  };

  static ThrottleProfileDTO fromJson(Map<String, dynamic> j) =>
      ThrottleProfileDTO(
        id: j['id'] as String?,
        name: (j['name'] ?? '') as String,
        downKbps: (j['downKbps'] ?? 0) as int,
        upKbps: (j['upKbps'] ?? 0) as int,
        packetLossPct: (j['packetLossPct'] ?? 0) as int,
        latencyMs: (j['latencyMs'] ?? 0) as int,
        latencyJitter: (j['latencyJitter'] ?? 0) as int,
        offline: (j['offline'] ?? false) as bool,
      );
}

class ThrottleService {
  AppHttpClient get _api => sl<AppHttpClient>();

  Future<Map<String, dynamic>> current() async {
    final r = await _api.get<Map<String, dynamic>>(path: '/_api/v1/throttle');
    return r.data ?? <String, dynamic>{};
  }

  Future<void> apply({
    required bool enabled,
    required int downKbps,
    required int upKbps,
    required int packetLossPct,
    int latencyMs = 0,
    int latencyJitter = 0,
    required bool offline,
  }) async {
    await _api.post(
      path: '/_api/v1/throttle',
      body: {
        'enabled': enabled,
        'downKbps': downKbps,
        'upKbps': upKbps,
        'packetLossPct': packetLossPct,
        'latencyMs': latencyMs,
        'latencyJitter': latencyJitter,
        'offline': offline,
      },
    );
  }

  Future<List<ThrottleProfileDTO>> listProfiles() async {
    final r = await _api.get<List<dynamic>>(path: '/_api/v1/throttle/profiles');
    final data = r.data ?? <dynamic>[];
    return data
        .whereType<Map<String, dynamic>>()
        .map(ThrottleProfileDTO.fromJson)
        .toList();
  }

  Future<ThrottleProfileDTO> upsert(ThrottleProfileDTO p) async {
    final r = await _api.post<Map<String, dynamic>>(
      path: '/_api/v1/throttle/profiles',
      body: p.toJson(),
    );
    return ThrottleProfileDTO.fromJson(r.data ?? <String, dynamic>{});
  }

  Future<void> deleteProfile(String id) async {
    await _api.delete(path: '/_api/v1/throttle/profiles/$id');
  }
}
