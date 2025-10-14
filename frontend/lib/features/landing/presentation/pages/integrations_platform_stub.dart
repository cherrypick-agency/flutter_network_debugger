bool macNativeAvailable() => false;
Future<void> autoIntegrateMacOS(String baseUrl) async {
  // no-op for non-io targets
}

Future<void> rollbackMacOS(String baseUrl) async {}
Future<void> deleteDevCA() async {}
