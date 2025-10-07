import 'dart:async';
import 'dart:io';

Future<int> runBinary({required String execPath}) async {
  // Запускаем в обычном режиме, чтобы сигналы корректно доходили
  final proc = await Process.start(
    execPath,
    const <String>[],
    runInShell: Platform.isWindows,
  );

  // pipe stdout/stderr
  final subOut = proc.stdout.listen((data) {
    stdout.add(data);
  });
  final subErr = proc.stderr.listen((data) {
    stderr.add(data);
  });

  // Forward SIGINT/SIGTERM
  StreamSubscription? sigInt;
  StreamSubscription? sigTerm;
  if (!Platform.isWindows) {
    sigInt = ProcessSignal.sigint.watch().listen(
      (_) => proc.kill(ProcessSignal.sigint),
    );
    sigTerm = ProcessSignal.sigterm.watch().listen(
      (_) => proc.kill(ProcessSignal.sigterm),
    );
  }

  final code = await proc.exitCode;
  await subOut.cancel();
  await subErr.cancel();
  await sigInt?.cancel();
  await sigTerm?.cancel();
  return code;
}
