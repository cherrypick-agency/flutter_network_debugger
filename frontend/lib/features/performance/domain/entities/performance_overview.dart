import 'package:equatable/equatable.dart';

class PerformanceOverview extends Equatable {
  const PerformanceOverview({
    required this.responseTimeStats,
    required this.throughputStats,
    required this.bandwidthStats,
    required this.topSlowEndpoints,
    required this.timeRange,
  });

  final ResponseTimeStats responseTimeStats;
  final ThroughputStats throughputStats;
  final BandwidthStats bandwidthStats;
  final List<EndpointPerformance> topSlowEndpoints;
  final TimeRange timeRange;

  factory PerformanceOverview.fromJson(Map<String, dynamic> json) {
    return PerformanceOverview(
      responseTimeStats: ResponseTimeStats.fromJson(
        json['responseTimeStats'] as Map<String, dynamic>,
      ),
      throughputStats: ThroughputStats.fromJson(
        json['throughputStats'] as Map<String, dynamic>,
      ),
      bandwidthStats: BandwidthStats.fromJson(
        json['bandwidthStats'] as Map<String, dynamic>,
      ),
      topSlowEndpoints: (json['topSlowEndpoints'] as List)
          .map((e) => EndpointPerformance.fromJson(e as Map<String, dynamic>))
          .toList(),
      timeRange: TimeRange.fromJson(json['timeRange'] as Map<String, dynamic>),
    );
  }

  @override
  List<Object?> get props => [
    responseTimeStats,
    throughputStats,
    bandwidthStats,
    topSlowEndpoints,
    timeRange,
  ];
}

class ResponseTimeStats extends Equatable {
  const ResponseTimeStats({
    required this.p50,
    required this.p95,
    required this.p99,
    required this.min,
    required this.max,
    required this.mean,
    required this.count,
  });

  final double p50;
  final double p95;
  final double p99;
  final double min;
  final double max;
  final double mean;
  final int count;

  factory ResponseTimeStats.fromJson(Map<String, dynamic> json) {
    return ResponseTimeStats(
      p50: (json['p50'] as num).toDouble(),
      p95: (json['p95'] as num).toDouble(),
      p99: (json['p99'] as num).toDouble(),
      min: (json['min'] as num).toDouble(),
      max: (json['max'] as num).toDouble(),
      mean: (json['mean'] as num).toDouble(),
      count: json['count'] as int,
    );
  }

  @override
  List<Object?> get props => [p50, p95, p99, min, max, mean, count];
}

class ThroughputStats extends Equatable {
  const ThroughputStats({
    required this.requestsPerSecond,
    required this.totalRequests,
    required this.successRate,
    required this.errorRate,
    required this.status2xx,
    required this.status4xx,
    required this.status5xx,
  });

  final double requestsPerSecond;
  final int totalRequests;
  final double successRate;
  final double errorRate;
  final int status2xx;
  final int status4xx;
  final int status5xx;

  factory ThroughputStats.fromJson(Map<String, dynamic> json) {
    return ThroughputStats(
      requestsPerSecond: (json['requestsPerSecond'] as num).toDouble(),
      totalRequests: json['totalRequests'] as int,
      successRate: (json['successRate'] as num).toDouble(),
      errorRate: (json['errorRate'] as num).toDouble(),
      status2xx: json['status2xx'] as int,
      status4xx: json['status4xx'] as int,
      status5xx: json['status5xx'] as int,
    );
  }

  @override
  List<Object?> get props => [
    requestsPerSecond,
    totalRequests,
    successRate,
    errorRate,
    status2xx,
    status4xx,
    status5xx,
  ];
}

class BandwidthStats extends Equatable {
  const BandwidthStats({
    required this.totalBytesUp,
    required this.totalBytesDown,
    required this.avgRequestSize,
    required this.avgResponseSize,
  });

  final int totalBytesUp;
  final int totalBytesDown;
  final double avgRequestSize;
  final double avgResponseSize;

  factory BandwidthStats.fromJson(Map<String, dynamic> json) {
    return BandwidthStats(
      totalBytesUp: json['totalBytesUp'] as int,
      totalBytesDown: json['totalBytesDown'] as int,
      avgRequestSize: (json['avgRequestSize'] as num).toDouble(),
      avgResponseSize: (json['avgResponseSize'] as num).toDouble(),
    );
  }

  @override
  List<Object?> get props => [
    totalBytesUp,
    totalBytesDown,
    avgRequestSize,
    avgResponseSize,
  ];
}

class EndpointPerformance extends Equatable {
  const EndpointPerformance({
    required this.endpoint,
    required this.method,
    required this.avgDuration,
    required this.p95Duration,
    required this.requestCount,
    required this.errorCount,
    required this.errorRate,
  });

  final String endpoint;
  final String method;
  final double avgDuration;
  final double p95Duration;
  final int requestCount;
  final int errorCount;
  final double errorRate;

  factory EndpointPerformance.fromJson(Map<String, dynamic> json) {
    return EndpointPerformance(
      endpoint: json['endpoint'] as String,
      method: json['method'] as String,
      avgDuration: (json['avgDuration'] as num).toDouble(),
      p95Duration: (json['p95Duration'] as num).toDouble(),
      requestCount: json['requestCount'] as int,
      errorCount: json['errorCount'] as int,
      errorRate: (json['errorRate'] as num).toDouble(),
    );
  }

  @override
  List<Object?> get props => [
    endpoint,
    method,
    avgDuration,
    p95Duration,
    requestCount,
    errorCount,
    errorRate,
  ];
}

class TimeRange extends Equatable {
  const TimeRange({required this.from, required this.to});

  final DateTime from;
  final DateTime to;

  factory TimeRange.fromJson(Map<String, dynamic> json) {
    return TimeRange(
      from: DateTime.parse(json['from'] as String),
      to: DateTime.parse(json['to'] as String),
    );
  }

  @override
  List<Object?> get props => [from, to];
}
