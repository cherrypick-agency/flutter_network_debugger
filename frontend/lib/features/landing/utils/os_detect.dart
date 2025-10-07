import 'os_detect_stub.dart'
    if (dart.library.html) 'os_detect_web.dart'
    as impl;

String detectOS() => impl.detectOS();
String detectArch() => impl.detectArch();
String macArchLabel(String arch) => impl.macArchLabel(arch);
Future<String?> detectArchPrecise() => impl.detectArchPrecise();
