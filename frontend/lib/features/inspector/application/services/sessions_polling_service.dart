import 'dart:async';

class SessionsPollingService {
  Timer? _poll;

  void start({required Future<void> Function() onTick}) {
    _poll?.cancel();
    _poll = Timer.periodic(const Duration(seconds: 2), (_) async {
      await onTick();
    });
  }

  void stop() {
    _poll?.cancel();
    _poll = null;
  }
}
