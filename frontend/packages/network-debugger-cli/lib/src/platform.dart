import 'dart:io';

class PlatformSpec {
  final String os; // linux | macos | windows
  final String arch; // amd64 | arm64 | 386
  PlatformSpec(this.os, this.arch);
}

PlatformSpec detectPlatform() {
  final os = Platform.operatingSystem; // linux, macos, windows
  String arch = _detectArchBasic();
  // Unix: попытка уточнить через uname -m
  if (!Platform.isWindows) {
    try {
      final res = Process.runSync('uname', ['-m']);
      final m = (res.stdout.toString().trim().toLowerCase());
      if (m.contains('arm64') || m.contains('aarch64')) arch = 'arm64';
      if (m.contains('x86_64')) arch = 'amd64';
      if (m == 'i386' || m == 'i686' || m == 'x86') arch = '386';
    } catch (_) {}
  }
  String normOs;
  switch (os) {
    case 'macos':
      normOs = 'darwin';
      break;
    case 'linux':
      normOs = 'linux';
      break;
    case 'windows':
      normOs = 'windows';
      break;
    default:
      normOs = os;
  }
  return PlatformSpec(normOs, arch);
}

String _detectArchBasic() {
  // Windows: учитываем WOW64
  if (Platform.isWindows) {
    final a1 =
        (Platform.environment['PROCESSOR_ARCHITECTURE'] ?? '').toLowerCase();
    final a2 =
        (Platform.environment['PROCESSOR_ARCHITEW6432'] ?? '').toLowerCase();
    final raw = '$a1 $a2';
    if (raw.contains('arm64')) return 'arm64';
    if (raw.contains('amd64') || raw.contains('x86_64')) return 'amd64';
    return '386';
  }
  final raw =
      (Platform.environment['HOSTTYPE'] ?? Platform.environment['ARCH'] ?? '')
          .toLowerCase();
  if (raw.contains('aarch64') || raw.contains('arm64')) return 'arm64';
  if (raw.contains('x86') && raw.contains('64')) return 'amd64';
  if (raw.contains('386') || raw.contains('x86')) return '386';
  final wordSize = (sizeOfPointer() == 8) ? 64 : 32;
  return wordSize == 64 ? 'amd64' : '386';
}

int sizeOfPointer() {
  // Быстрая эвристика: на Dart VM 64-bit обычно intPtrSize == 8
  // Platform.version включает архитектуру, но ненадёжно. Это простой fallback.
  return (PidCurrent.is64BitProcess ? 8 : 4);
}

class PidCurrent {
  static bool get is64BitProcess {
    // На Dart нет прямого API, используем эвристику по адресному пространству.
    try {
      return Platform.version.contains('x64') ||
          Platform.version.contains('arm64');
    } catch (_) {
      return true;
    }
  }
}

String archiveExtensionFor(PlatformSpec p) {
  // На Windows — zip, на unix — tar.gz
  return p.os == 'windows' ? 'zip' : 'tar.gz';
}

String binaryFileName(PlatformSpec p) {
  final base = 'network-debugger-web';
  return p.os == 'windows' ? '$base.exe' : base;
}
