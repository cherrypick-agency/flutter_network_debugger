import '../../domain/entities/frame.dart';
import '../../domain/repositories/inspector_repository.dart';

class ListFramesUseCase {
  ListFramesUseCase(this._repo);
  final InspectorRepository _repo;
  Future<List<Frame>> call(String sessionId, {String? from, int limit = 100}) =>
      _repo.listFrames(sessionId, from: from, limit: limit);
}


