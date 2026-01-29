import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';
import 'package:test/test.dart';

void main() {
  test('attach adds ReverseProxyInterceptor in reverse mode', () {
    final dio = Dio(BaseOptions(baseUrl: 'https://api.example.test'));

    DioDebugger.attach(
      dio,
      enabled: true,
      upstreamBaseUrl: 'https://api.example.test',
      proxyBaseUrl: 'http://localhost:9091',
      proxyHttpPath: '/httpproxy',
    );

    final has = dio.interceptors.any((i) => i is ReverseProxyInterceptor);
    expect(has, isTrue);
  });
}
