import 'dart:io';

import 'package:dio/dio.dart';
import 'package:dio/io.dart';

Dio forwardProxyAttach(
  Dio dio, {
  required String proxyHostPort, // host:port
  bool allowBadCerts = false,
}) {
  dio.httpClientAdapter = IOHttpClientAdapter(
    createHttpClient: () {
      final client = HttpClient();
      client.findProxy = (uri) => 'PROXY $proxyHostPort';
      if (allowBadCerts) {
        client.badCertificateCallback = (cert, host, port) => true;
      }
      return client;
    },
  );
  return dio;
}
