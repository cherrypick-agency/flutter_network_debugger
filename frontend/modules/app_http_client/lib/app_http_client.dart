import 'dart:async';

import 'package:injectable/injectable.dart';
import 'package:modules_basis/module.dart' as module;
import 'package:modules_basis/modules_basis.dart' show ContainerDI;

import 'application/app_http_client.dart';
import 'application/dio/dio_http_client_impl.dart';
import 'domain/tokens_data.dart';
import 'application/tokens_storage.dart';

typedef BaseUrlType = String Function();
typedef OnTokenRefreshedType = void Function(TokensData);

@lazySingleton
class AppHttpClientModule extends module.Module<AppHttpClientModuleResult> {
  final BaseUrlType baseURL;
  final OnTokenRefreshedType onTokenRefreshed;
  final ContainerDI containerDI;

  AppHttpClientModule(
    this.baseURL,
    this.onTokenRefreshed,
    this.containerDI,
  );

  @override
  Future<AppHttpClientModuleResult> execute() async {
    TokensStorage tokensStorage;
    try {
      tokensStorage = containerDI<TokensStorage>();
    } catch (_) {
      tokensStorage = TokensStorageImpl();
      containerDI.registerSingleton<TokensStorage>(tokensStorage);
    }

    final http = DioHttpClientImpl(
      defaultHost: baseURL,
      onTokenRefreshed: onTokenRefreshed,
      tokensStorage: tokensStorage,
    );

    await http.init();

    containerDI.registerSingleton<AppHttpClient>(http);

    return AppHttpClientModuleResult(http);
  }

  @disposeMethod
  void dispose() {
    containerDI.unregister<AppHttpClient>();
  }
}

class AppHttpClientModuleResult extends module.ModuleResult {
  final AppHttpClient httpClient;

  AppHttpClientModuleResult(this.httpClient);
}

/// 2. Функция-инициализатор микропакета
// @InjectableInit.microPackage()
// void initAppHttpClientPackage(ContainerDI getIt) => $initGetIt(getIt);

