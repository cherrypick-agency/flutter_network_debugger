bool nativeAutomationAvailable() => false;
Future<bool> isSystemProxyEnabled() async => false;
Future<bool> autoIntegrate(String baseUrl) async => false;
Future<void> rollback(String baseUrl) async {}
Future<void> deleteDevCA() async {}
Future<String> proxyDiagnostics() async => 'N/A';
Future<void> openSystemProxySettings() async {}
Future<bool> isDevCAInstalledSystem() async => false;
