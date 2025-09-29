import 'package:get_it/get_it.dart';

class ContainerDI {
  ContainerDI(this._getIt);
  final GetIt _getIt;

  T call<T extends Object>() => _getIt<T>();

  void registerSingleton<T extends Object>(T instance) => _getIt.registerSingleton<T>(instance);

  void unregister<T extends Object>() {
    if (_getIt.isRegistered<T>()) {
      _getIt.unregister<T>();
    }
  }
}
