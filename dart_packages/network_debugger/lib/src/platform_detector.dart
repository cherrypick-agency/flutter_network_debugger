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
  static String getBinaryName({String binaryBaseName = 'network-debugger-web'}) {
    final platform = getPlatformIdentifier();
    final ext = getFileExtension();
    return '${binaryBaseName}_$platform$ext';
  }

  /// Returns the archive name for download.
  /// Example: 'network-debugger-web_darwin_arm64.tar.gz' or 'network-debugger-web_windows_amd64.zip'
  static String getArchiveName({String binaryBaseName = 'network-debugger-web'}) {
    final platform = getPlatformIdentifier();
    final ext = getArchiveExtension();
    return '${binaryBaseName}_$platform$ext';
  }

  static String _getOS() {
    if (Platform.isMacOS) return 'darwin';
    if (Platform.isWindows) return 'windows';
    if (Platform.isLinux) return 'linux';
    throw UnsupportedError('Unsupported operating system: ${Platform.operatingSystem}');
  }

  static String _getArchitecture() {
    // Platform.version contains architecture information
    final version = Platform.version;

    // Try to detect from Dart VM version string
    // Example: "2.19.0 (stable) on 'macos_arm64'"
    if (version.contains('arm64') || version.contains('aarch64')) {
      return 'arm64';
    }

    if (version.contains('x64') || version.contains('amd64')) {
      return 'amd64';
    }

    if (version.contains('ia32') || version.contains('x86')) {
      // Only Windows has 386 builds
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
