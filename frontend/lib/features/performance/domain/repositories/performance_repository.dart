import '../entities/performance_overview.dart';

abstract class PerformanceRepository {
  Future<PerformanceOverview> getOverview(DateTime from, DateTime to);
}
