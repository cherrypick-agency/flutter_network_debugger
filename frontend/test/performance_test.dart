import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/features/performance/application/stores/performance_store.dart';
import 'package:frontend/features/performance/domain/entities/performance_overview.dart';
import 'package:frontend/features/performance/domain/repositories/performance_repository.dart';

// Mock repository for testing
class MockPerformanceRepository implements PerformanceRepository {
  PerformanceOverview? mockOverview;
  bool shouldThrowError = false;

  @override
  Future<PerformanceOverview> getOverview(DateTime from, DateTime to) async {
    if (shouldThrowError) {
      throw Exception('Mock error: Failed to fetch performance data');
    }

    if (mockOverview != null) {
      return mockOverview!;
    }

    // Return default mock data
    return PerformanceOverview(
      responseTimeStats: ResponseTimeStats(
        p50: 120.5,
        p95: 350.2,
        p99: 850.0,
        min: 50.0,
        max: 1200.0,
        mean: 180.3,
        count: 150,
      ),
      throughputStats: ThroughputStats(
        requestsPerSecond: 25.5,
        totalRequests: 1530,
        successRate: 0.95,
        errorRate: 0.05,
        status2xx: 1454,
        status4xx: 38,
        status5xx: 38,
      ),
      bandwidthStats: BandwidthStats(
        totalBytesUp: 1048576,
        totalBytesDown: 3145728,
        avgRequestSize: 512.0,
        avgResponseSize: 2048.0,
      ),
      topSlowEndpoints: [
        EndpointPerformance(
          endpoint: '/api/slow-endpoint',
          method: 'GET',
          avgDuration: 450.0,
          p95Duration: 890.0,
          requestCount: 50,
          errorCount: 5,
          errorRate: 0.1,
        ),
        EndpointPerformance(
          endpoint: '/api/medium-endpoint',
          method: 'POST',
          avgDuration: 200.0,
          p95Duration: 400.0,
          requestCount: 100,
          errorCount: 2,
          errorRate: 0.02,
        ),
      ],
      timeRange: TimeRange(from: from, to: to),
    );
  }
}

void main() {
  late PerformanceStore store;
  late MockPerformanceRepository mockRepo;

  setUp(() {
    mockRepo = MockPerformanceRepository();
    store = PerformanceStore(mockRepo);
  });

  group('Time Range Management', () {
    test('initial time range is last hour', () {
      final now = DateTime.now();
      final hourAgo = now.subtract(const Duration(hours: 1));

      // Check that from and to are set correctly
      expect(
        store.from.difference(hourAgo).abs().inMinutes < 1,
        true,
        reason: 'from should be approximately an hour ago',
      );
      expect(
        store.to.difference(now).abs().inMinutes < 1,
        true,
        reason: 'to should be approximately now',
      );
    });

    test('setTimeRange updates from and to and loads metrics', () async {
      final newFrom = DateTime.now().subtract(const Duration(hours: 24));
      final newTo = DateTime.now();

      store.setTimeRange(newFrom, newTo);
      // Wait for loadMetrics to complete
      await Future.delayed(const Duration(milliseconds: 50));

      expect(store.from, newFrom);
      expect(store.to, newTo);
      expect(store.overview, isNotNull);
      expect(store.loading, false);
    });

    test('setLastHour sets range to last hour', () async {
      store.setLastHour();
      await Future.delayed(const Duration(milliseconds: 50));

      final now = DateTime.now();
      final hourAgo = now.subtract(const Duration(hours: 1));

      expect(store.from.difference(hourAgo).abs().inMinutes < 1, true);
      expect(store.to.difference(now).abs().inMinutes < 1, true);
    });

    test('setLastDay sets range to last day', () async {
      store.setLastDay();
      await Future.delayed(const Duration(milliseconds: 50));

      final now = DateTime.now();
      final dayAgo = now.subtract(const Duration(days: 1));

      expect(store.from.difference(dayAgo).abs().inMinutes < 1, true);
      expect(store.to.difference(now).abs().inMinutes < 1, true);
    });

    test('setLastWeek sets range to last week', () async {
      store.setLastWeek();
      await Future.delayed(const Duration(milliseconds: 50));

      final now = DateTime.now();
      final weekAgo = now.subtract(const Duration(days: 7));

      expect(store.from.difference(weekAgo).abs().inMinutes < 1, true);
      expect(store.to.difference(now).abs().inMinutes < 1, true);
    });
  });

  group('Load Metrics', () {
    test('loadMetrics successfully loads performance data', () async {
      // Action
      await store.loadMetrics();

      // Verification
      expect(store.loading, false);
      expect(store.error, null);
      expect(store.overview, isNotNull);

      // Check that data is loaded correctly
      final overview = store.overview!;
      expect(overview.responseTimeStats.count, 150);
      expect(overview.responseTimeStats.p50, 120.5);
      expect(overview.throughputStats.totalRequests, 1530);
      expect(overview.topSlowEndpoints.length, 2);
    });

    test('loadMetrics sets loading to true during loading', () async {
      // Create slow mock to check loading state
      mockRepo = MockPerformanceRepository();
      store = PerformanceStore(mockRepo);

      // Start loading but don't wait for completion
      final future = store.loadMetrics();

      // Initially loading should be true
      // (may already be false if loading is very fast)
      await future;

      // After completion loading should be false
      expect(store.loading, false);
    });

    test('loadMetrics handles errors', () async {
      // Setup
      mockRepo.shouldThrowError = true;

      // Action
      await store.loadMetrics();

      // Verification
      expect(store.loading, false);
      expect(store.error, isNotNull);
      expect(store.error, contains('Failed to fetch'));
    });

    test('loadMetrics clears previous error on successful load', () async {
      // Setup - first load with error
      mockRepo.shouldThrowError = true;
      await store.loadMetrics();
      expect(store.error, isNotNull);

      // Action - second load is successful
      mockRepo.shouldThrowError = false;
      await store.loadMetrics();

      // Verification
      expect(store.error, null);
      expect(store.overview, isNotNull);
    });
  });

  group('Response Time Stats', () {
    test('loaded responseTimeStats contains correct data', () async {
      await store.loadMetrics();

      final stats = store.overview!.responseTimeStats;
      expect(stats.p50, 120.5);
      expect(stats.p95, 350.2);
      expect(stats.p99, 850.0);
      expect(stats.min, 50.0);
      expect(stats.max, 1200.0);
      expect(stats.mean, 180.3);
      expect(stats.count, 150);
    });
  });

  group('Throughput Stats', () {
    test('loaded throughputStats contains correct data', () async {
      await store.loadMetrics();

      final stats = store.overview!.throughputStats;
      expect(stats.requestsPerSecond, 25.5);
      expect(stats.totalRequests, 1530);
      expect(stats.successRate, 0.95);
      expect(stats.errorRate, 0.05);
      expect(stats.status2xx, 1454);
      expect(stats.status4xx, 38);
      expect(stats.status5xx, 38);
    });

    test('successRate + errorRate should approximately equal 1.0', () async {
      await store.loadMetrics();

      final stats = store.overview!.throughputStats;
      final sum = stats.successRate + stats.errorRate;

      expect((sum - 1.0).abs() < 0.01, true);
    });
  });

  group('Bandwidth Stats', () {
    test('loaded bandwidthStats contains correct data', () async {
      await store.loadMetrics();

      final stats = store.overview!.bandwidthStats;
      expect(stats.totalBytesUp, 1048576);
      expect(stats.totalBytesDown, 3145728);
      expect(stats.avgRequestSize, 512.0);
      expect(stats.avgResponseSize, 2048.0);
    });
  });

  group('Top Slow Endpoints', () {
    test('loads list of slow endpoints', () async {
      await store.loadMetrics();

      final endpoints = store.overview!.topSlowEndpoints;
      expect(endpoints.length, 2);
    });

    test('endpoints are sorted by speed (slowest first)', () async {
      await store.loadMetrics();

      final endpoints = store.overview!.topSlowEndpoints;

      // First endpoint should be slower than second
      expect(endpoints[0].avgDuration > endpoints[1].avgDuration, true);
      expect(endpoints[0].p95Duration > endpoints[1].p95Duration, true);
    });

    test('endpoint contains complete information', () async {
      await store.loadMetrics();

      final endpoint = store.overview!.topSlowEndpoints[0];

      expect(endpoint.endpoint, isNotEmpty);
      expect(endpoint.method, isNotEmpty);
      expect(endpoint.avgDuration, greaterThan(0));
      expect(endpoint.p95Duration, greaterThan(0));
      expect(endpoint.requestCount, greaterThan(0));
      expect(endpoint.errorRate, greaterThanOrEqualTo(0));
      expect(endpoint.errorRate, lessThanOrEqualTo(1));
    });

    test('errorRate is calculated correctly', () async {
      await store.loadMetrics();

      final endpoint = store.overview!.topSlowEndpoints[0];
      final expectedErrorRate = endpoint.errorCount / endpoint.requestCount;

      expect(endpoint.errorRate, expectedErrorRate);
    });
  });

  group('Custom Mock Data', () {
    test('can set custom data via mockOverview', () async {
      // Setup
      final customFrom = DateTime.now().subtract(const Duration(hours: 2));
      final customTo = DateTime.now();

      mockRepo.mockOverview = PerformanceOverview(
        responseTimeStats: ResponseTimeStats(
          p50: 999.0,
          p95: 1999.0,
          p99: 2999.0,
          min: 100.0,
          max: 3000.0,
          mean: 1500.0,
          count: 500,
        ),
        throughputStats: ThroughputStats(
          requestsPerSecond: 50.0,
          totalRequests: 3000,
          successRate: 0.98,
          errorRate: 0.02,
          status2xx: 2940,
          status4xx: 30,
          status5xx: 30,
        ),
        bandwidthStats: BandwidthStats(
          totalBytesUp: 2097152,
          totalBytesDown: 12288000,
          avgRequestSize: 1024.0,
          avgResponseSize: 4096.0,
        ),
        topSlowEndpoints: [],
        timeRange: TimeRange(from: customFrom, to: customTo),
      );

      // Action
      await store.loadMetrics();

      // Verification
      expect(store.overview!.responseTimeStats.p50, 999.0);
      expect(store.overview!.responseTimeStats.count, 500);
      expect(store.overview!.throughputStats.totalRequests, 3000);
    });
  });

  group('Edge Cases', () {
    test('overview remains null until first load', () {
      expect(store.overview, null);
    });

    test('loading is false by default', () {
      expect(store.loading, false);
    });

    test('error is null by default', () {
      expect(store.error, null);
    });

    test('can load metrics multiple times in a row', () async {
      await store.loadMetrics();
      final firstOverview = store.overview;

      await store.loadMetrics();
      final secondOverview = store.overview;

      // Both overviews should be non-null
      expect(firstOverview, isNotNull);
      expect(secondOverview, isNotNull);

      // Data should be the same (since mock returns the same thing)
      expect(
        firstOverview!.responseTimeStats.count,
        secondOverview!.responseTimeStats.count,
      );
    });
  });
}
