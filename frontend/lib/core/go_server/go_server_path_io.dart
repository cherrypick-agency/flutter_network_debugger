import 'dart:io';
import 'package:path/path.dart' as path;

/// Desktop implementation for finding Go server binary
Future<String?> getGoServerPathImpl() async {
  // Определяем имя бинарника в зависимости от платформы
  final String binaryName;
  final String arch;

  if (Platform.isMacOS) {
    // Определяем архитектуру для macOS
    final result = await Process.run('uname', ['-m']);
    final machine = (result.stdout as String).trim();
    arch = machine == 'arm64' ? 'arm64' : 'amd64';
    binaryName = 'server_darwin_$arch';
  } else if (Platform.isWindows) {
    // Для Windows пока только amd64
    arch = 'amd64';
    binaryName = 'server_windows_$arch.exe';
  } else if (Platform.isLinux) {
    // Определяем архитектуру для Linux
    final result = await Process.run('uname', ['-m']);
    final machine = (result.stdout as String).trim();
    arch = machine.contains('aarch64') || machine.contains('arm64')
        ? 'arm64'
        : 'amd64';
    binaryName = 'server_linux_$arch';
  } else {
    return null;
  }

  // Пути где может находиться бинарник
  final possiblePaths = <String>[
    // В resources рядом с executable (production)
    path.join(
      path.dirname(Platform.resolvedExecutable),
      'resources',
      binaryName,
    ),
    // В Contents/Resources для macOS app bundle
    path.join(
      path.dirname(Platform.resolvedExecutable),
      '..',
      'Resources',
      binaryName,
    ),
    // Для development - в корне проекта
    path.join(
      path.dirname(Platform.resolvedExecutable),
      '..',
      '..',
      '..',
      'resources',
      binaryName,
    ),
    // В корне проекта Go (для разработки)
    path.join(
      path.dirname(Platform.resolvedExecutable),
      '..',
      '..',
      '..',
      '..',
      'cmd',
      'network-debugger',
      'network-debugger${Platform.isWindows ? ".exe" : ""}',
    ),
  ];

  // Ищем первый существующий файл
  for (final serverPath in possiblePaths) {
    final file = File(serverPath);
    if (await file.exists()) {
      // Проверяем что файл исполняемый на Unix системах
      if (!Platform.isWindows) {
        final stat = await file.stat();
        if (stat.mode & 0x49 == 0) {
          // 0x49 = 0111 (executable bits)
          // Делаем файл исполняемым
          await Process.run('chmod', ['+x', serverPath]);
        }
      }
      return serverPath;
    }
  }

  return null;
}
