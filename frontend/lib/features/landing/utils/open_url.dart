import 'open_url_stub.dart' if (dart.library.html) 'open_url_web.dart' as impl;

void openUrl(String url) => impl.openUrl(url);
