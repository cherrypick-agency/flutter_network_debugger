import '../domain/entities/performance_overview.dart';
import '../domain/repositories/performance_repository.dart';
import 'performance_api.dart';

class PerformanceRepositoryImpl implements PerformanceRepository {
  PerformanceRepositoryImpl(this._api);
  final PerformanceApi _api;

  @override
  Future<PerformanceOverview> getOverview(DateTime from, DateTime to) async {
    final json = await _api.getOverview(from, to);
    return PerformanceOverview.fromJson(json);
  }
}
