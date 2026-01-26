import 'dart:io';

Directory findRepoRoot() {
  var dir = Directory.current;
  for (var i = 0; i < 10; i++) {
    final goMod = File('${dir.path}${Platform.pathSeparator}go.mod');
    final cmdMain = File(
      '${dir.path}${Platform.pathSeparator}cmd${Platform.pathSeparator}network-debugger${Platform.pathSeparator}main.go',
    );
    if (goMod.existsSync() && cmdMain.existsSync()) return dir;
    dir = dir.parent;
  }
  throw StateError(
      'repo root not found (expected go.mod + cmd/network-debugger/main.go)');
}
