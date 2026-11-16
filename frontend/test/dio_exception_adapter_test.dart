import 'package:app_http_client/application/app_http_exception.dart';
import 'package:app_http_client/application/dio/dio_exception_adapter.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  const adapter = DioExceptionAdapter();

  test('maps Dio response transform timeout to receive timeout', () {
    expect(
      adapter.adapt(DioExceptionType.transformTimeout),
      AppHttpErrorType.receiveTimeout,
    );
  });

  test('preserves existing Dio timeout mappings', () {
    expect(
      adapter.adapt(DioExceptionType.connectionTimeout),
      AppHttpErrorType.connectionTimeout,
    );
    expect(
      adapter.adapt(DioExceptionType.sendTimeout),
      AppHttpErrorType.sendTimeout,
    );
    expect(
      adapter.adapt(DioExceptionType.receiveTimeout),
      AppHttpErrorType.receiveTimeout,
    );
  });
}
