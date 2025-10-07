// Fallback for non-web platforms (tests, mobile). Provide sane defaults.
String detectOS() => 'linux';
String detectArch() => 'amd64';
String macArchLabel(String arch) => arch == 'arm64' ? 'Apple Silicon' : 'Intel';
Future<String?> detectArchPrecise() async => null;
