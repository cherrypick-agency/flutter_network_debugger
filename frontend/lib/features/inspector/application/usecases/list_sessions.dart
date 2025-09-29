import '../../domain/entities/session.dart';
import '../../domain/repositories/inspector_repository.dart';

class ListSessionsUseCase {
  ListSessionsUseCase(this._repo);
  final InspectorRepository _repo;
  Future<List<Session>> call({String? q, String? target}) => _repo.listSessions(q: q, target: target);
}


