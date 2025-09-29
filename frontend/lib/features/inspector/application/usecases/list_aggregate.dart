import '../../domain/repositories/inspector_repository.dart';

class ListAggregateUseCase {
  ListAggregateUseCase(this._repo);
  final InspectorRepository _repo;
  Future<List<Map<String, dynamic>>> call({String groupBy = 'domain'}) => _repo.aggregateSessions(groupBy: groupBy);
}


