import 'package:dio/dio.dart';

/// Stub for [forwardProxyAttach] on platforms without dart:io support.
///
/// On web platforms forward proxy is not supported, so the function
/// simply returns the passed Dio instance unchanged.
Dio forwardProxyAttach(
  Dio dio, {
  required String proxyHostPort,
  bool allowBadCerts = false,
}) =>
    dio;
