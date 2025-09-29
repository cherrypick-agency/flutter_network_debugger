import 'dart:async';

import 'module.dart';

class ModuleInvoker {
  final initializedModules = <Module>{};
  final streamInitializing = StreamController<Module>.broadcast();
  // final providers = <Provider>{};

  Future<void> use(Module module) async {
    // module.addAll(module.providers);

    module.executeResult = await module.execute();

    initializedModules.add(module);
    streamInitializing.sink.add(module);
  }

  Future<void> useAll(Iterable<Module> modules) {
    for (var m in modules) {
      m.prepare(
        moduleInvoker: this,
      );
    }

    return Future.wait(modules.map(use));
  }
}
