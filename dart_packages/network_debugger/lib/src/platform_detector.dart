import 'dart:ffi';
import 'dart:io';

/// Detects the current platform and returns the identifier used in binary names.
class PlatformDetector {
  /// Returns platform identifier (e.g., 'darwin_arm64', 'windows_amd64', 'linux_amd64').
  static String getPlatformIdentifier() {
    final os = _getOS();
    final arch = _getArchitecture();
    return '${os}_$arch';
  }

  /// Returns the file extension for the current platform.
  static String getFileExtension() {
    return Platform.isWindows ? '.exe' : '';
  }

  /// Returns the archive extension for the current platform.
  static String getArchiveExtension() {
    return Platform.isWindows ? '.zip' : '.tar.gz';
  }

  /// Returns the binary name with platform identifier.
  /// Example: 'network-debugger-web_darwin_arm64' or 'network-debugger-web_windows_amd64.exe'
  static String getBinaryName(
      {String binaryBaseName = 'network-debugger-web',}) {
    final platform = getPlatformIdentifier();
    final ext = getFileExtension();
    return '${binaryBaseName}_$platform$ext';
  }

  /// Returns the archive name for download.
  /// Example: 'network-debugger-web_darwin_arm64.tar.gz' or 'network-debugger-web_windows_amd64.zip'
  static String getArchiveName(
      {String binaryBaseName = 'network-debugger-web',}) {
    final platform = getPlatformIdentifier();
    final ext = getArchiveExtension();
    return '${binaryBaseName}_$platform$ext';
  }

  static String _getOS() {
    if (Platform.isMacOS) return 'darwin';
    if (Platform.isWindows) return 'windows';
    if (Platform.isLinux) return 'linux';
    throw UnsupportedError(
        'Unsupported operating system: ${Platform.operatingSystem}',
    );
  }

  static String _getArchitecture() {
    // Use dart:ffi for precise architecture detection
    final abi = Abi.current();

    // Check for ARM64 architectures
    if (abi == Abi.androidArm64 ||
        abi == Abi.linuxArm64 ||
        abi == Abi.macosArm64 ||
        abi == Abi.windowsArm64) {
      return 'arm64';
    }

    // Check for x64/AMD64 architectures
    if (abi == Abi.androidX64 ||
        abi == Abi.linuxX64 ||
        abi == Abi.macosX64 ||
        abi == Abi.windowsX64) {
      return 'amd64';
    }

    // Check for Windows 32-bit (IA32)
    if (abi == Abi.windowsIA32) {
      return '386';
    }

    // Fallback: try to detect from Dart VM version string
    // This handles any edge cases not covered by Abi enum
    final version = Platform.version;
    if (version.contains('arm64') || version.contains('aarch64')) {
      return 'arm64';
    }
    if (version.contains('ia32') || version.contains('x86')) {
      return Platform.isWindows ? '386' : 'amd64';
    }

    // Default to amd64 for most platforms
    return 'amd64';
  }

  /// Checks if the current platform is supported.
  static bool isSupported() {
    try {
      getPlatformIdentifier();
      return true;
    } catch (_) {
      return false;
    }
  }

  /// Returns a human-readable platform description.
  static String getPlatformDescription() {
    final os = Platform.operatingSystem;
    final version = Platform.version;
    final identifier = getPlatformIdentifier();
    return '$os ($identifier) - Dart $version';
  }
}
