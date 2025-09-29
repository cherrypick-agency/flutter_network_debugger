import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:injectable/injectable.dart';

/// Ключи, которые приходят с бэкенда в ответах.
abstract class TokensStorageKeys {
  static const String token = 'token';
  static const String refreshToken = 'refresh_token';
}

abstract class TokensStorage {
  String? get token;

  String? get refreshToken;

  Future<void> init();

  Future<void> save(String? token, String? refreshToken);

  Future<void> clear();
}

/// Реализация, использующая защищённое хранилище KeyChain/KeyStore.
@LazySingleton(as: TokensStorage)
class TokensStorageImpl implements TokensStorage {
  late final FlutterSecureStorage _storage;

  String? _token;
  String? _refreshToken;

  TokensStorageImpl() {
    _storage = const FlutterSecureStorage();
  }

  @override
  String? get token => _token;

  @override
  String? get refreshToken => _refreshToken;

  @override
  Future<void> init() async {
    _token = await _storage.read(key: TokensStorageKeys.token);
    _refreshToken = await _storage.read(key: TokensStorageKeys.refreshToken);
  }

  @override
  Future<void> save(String? token, String? refreshToken) async {
    final futures = <Future<void>>[];

    if (token != null) {
      _token = token;
      futures.add(_storage.write(key: TokensStorageKeys.token, value: token));
    }

    if (refreshToken != null) {
      _refreshToken = refreshToken;
      futures.add(
          _storage.write(key: TokensStorageKeys.refreshToken, value: refreshToken));
    }

    await Future.wait(futures);
  }

  @override
  Future<void> clear() async {
    _token = null;
    _refreshToken = null;
    await Future.wait([
      _storage.delete(key: TokensStorageKeys.token),
      _storage.delete(key: TokensStorageKeys.refreshToken),
    ]);
  }
}

/// Тестовая реализация для unit-тестов.
class TokensTestStorage implements TokensStorage {
  @override
  String? token;

  @override
  String? refreshToken;

  @override
  Future<void> init() async {}

  @override
  Future<void> clear() async {
    token = null;
    refreshToken = null;
  }

  @override
  Future<void> save(String? token, String? refreshToken) async {
    this.token = token ?? this.token;
    this.refreshToken = refreshToken ?? this.refreshToken;
  }
}
