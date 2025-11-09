import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';

class MockEventBus implements EventBus {
  final StreamController<EditorEvent> _controller =
      StreamController<EditorEvent>.broadcast();

  @override
  void publish(EditorEvent event) {
    _controller.add(event);
  }

  @override
  Stream<EditorEvent> get stream => _controller.stream;

  @override
  Stream<T> on<T extends EditorEvent>() {
    return _controller.stream.where((event) => event is T).cast<T>();
  }

  @override
  void dispose() {
    _controller.close();
  }
}
