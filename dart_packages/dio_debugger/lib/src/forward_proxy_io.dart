import 'dart:io';

import 'package:dio/dio.dart';
import 'package:dio/io.dart';

/// Attaches a forward proxy to a Dio instance.
///
/// Configures the HTTP client to route all requests through the specified
/// proxy server. Works only on platforms with dart:io support.
///
/// [proxyHostPort] - proxy server address in "host:port" format
///   (e.g. "localhost:8080" or "proxy.example.com:3128").
/// [allowBadCerts] - if `true`, allows self-signed SSL certificates.
///
/// Returns the same [dio] instance for method chaining.
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
