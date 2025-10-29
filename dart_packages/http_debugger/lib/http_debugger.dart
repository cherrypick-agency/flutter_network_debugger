library http_debugger;

// Экспортируем платформенно-специфичную реализацию: на IO — рабочую,
// на Web и др. — заглушку без побочных эффектов.
export 'src/overrides_stub.dart' if (dart.library.io) 'src/overrides_io.dart';
export 'src/http_client_wrapper.dart';
