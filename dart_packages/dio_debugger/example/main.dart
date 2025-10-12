import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';

Future<void> main() async {
  final dio = Dio(
    BaseOptions(baseUrl: 'https://api.example.com'),
  );

  // Attach reverse/forward proxy interceptor for local debugging
  DioDebugger.attach(
    dio,
    upstreamBaseUrl: 'https://api.example.com',
    proxyBaseUrl: 'http://localhost:9091',
    proxyHttpPath: '/httpproxy',
  );

  final response = await dio.get('/health');
  print('Status: \'${response.statusCode}\'');
}
