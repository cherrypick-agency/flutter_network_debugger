import '../../domain/entities/event.dart';
import '../../domain/repositories/inspector_repository.dart';

class ListEventsUseCase {
  ListEventsUseCase(this._repo);
  final InspectorRepository _repo;
  Future<List<EventEntity>> call(String sessionId, {String? from, int limit = 100}) =>
      _repo.listEvents(sessionId, from: from, limit: limit);
}


