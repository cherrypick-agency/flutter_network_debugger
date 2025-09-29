import '../entities/session.dart';
import '../entities/frame.dart';
import '../entities/event.dart';

abstract class InspectorRepository {
  Future<List<Session>> listSessions({String? q, String? target});
  Future<List<Frame>> listFrames(String sessionId, {String? from, int limit = 100});
  Future<List<EventEntity>> listEvents(String sessionId, {String? from, int limit = 100});
  Future<List<Map<String, dynamic>>> aggregateSessions({String groupBy = 'domain'});
}


