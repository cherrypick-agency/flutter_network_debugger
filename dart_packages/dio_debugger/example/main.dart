import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';

Future<void> main() async {
  final dio = Dio(
    BaseOptions(baseUrl: 'https://api.example.com'),
  );

  // Attach reverse proxy interceptor for local debugging
  DioDebugger.attach(dio);

  final response = await dio.get('/health');
  print('Status: \'${response.statusCode}\'');
}
