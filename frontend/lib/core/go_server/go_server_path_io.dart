import 'dart:io';
import 'package:path/path.dart' as path;

/// Desktop implementation for finding Go server binary
Future<String?> getGoServerPathImpl() async {
  // Determine binary name based on platform
  final String binaryName;
  final String arch;

  if (Platform.isMacOS) {
    // Determine architecture for macOS
    final result = await Process.run('uname', ['-m']);
    final machine = (result.stdout as String).trim();
    arch = machine == 'arm64' ? 'arm64' : 'amd64';
    binaryName = 'server_darwin_$arch';
  } else if (Platform.isWindows) {
    // For Windows only amd64 for now
    arch = 'amd64';
    binaryName = 'server_windows_$arch.exe';
  } else if (Platform.isLinux) {
    // Determine architecture for Linux
    final result = await Process.run('uname', ['-m']);
    final machine = (result.stdout as String).trim();
    arch = machine.contains('aarch64') || machine.contains('arm64')
        ? 'arm64'
        : 'amd64';
    binaryName = 'server_linux_$arch';
  } else {
    return null;
  }

  // Paths where binary might be located
  final possiblePaths = <String>[
    // In resources next to executable (production)
    path.join(
      path.dirname(Platform.resolvedExecutable),
      'resources',
      binaryName,
    ),
    // In Contents/Resources for macOS app bundle
    path.join(
      path.dirname(Platform.resolvedExecutable),
      '..',
      'Resources',
      binaryName,
    ),
    // For development - in project root
    path.join(
      path.dirname(Platform.resolvedExecutable),
      '..',
      '..',
      '..',
      'resources',
      binaryName,
    ),
    // In Go project root (for development)
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

  // Find first existing file
  for (final serverPath in possiblePaths) {
    final file = File(serverPath);
    if (await file.exists()) {
      // Check if file is executable on Unix systems
      if (!Platform.isWindows) {
        final stat = await file.stat();
        if (stat.mode & 0x49 == 0) {
          // 0x49 = 0111 (executable bits)
          // Make file executable
          await Process.run('chmod', ['+x', serverPath]);
        }
      }
      return serverPath;
    }
  }

  return null;
}
