import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';

Future<void> main() async {
  final dio = Dio(
    BaseOptions(baseUrl: 'https://api.example.com'),
  );

  // Attach reverse proxy interceptor for local debugging
  // Set resetCaptureOnHotRestart: true to automatically clear previous sessions
  // and start a new capture on each hot restart
  DioDebugger.attach(
    dio,
    resetCaptureOnHotRestart: true,
  );

  final response = await dio.get('/health');
  print('Status: \'${response.statusCode}\'');
}
