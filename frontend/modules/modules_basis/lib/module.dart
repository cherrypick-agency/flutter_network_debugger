abstract class Module<TResult extends ModuleResult> {
  TResult? executeResult;

  /// Optional preparation hook used by ModuleInvoker
  void prepare({dynamic moduleInvoker}) {}

  Future<TResult> execute();
}

class ModuleResult {}
