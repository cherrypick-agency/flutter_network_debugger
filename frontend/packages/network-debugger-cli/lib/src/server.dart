import 'dart:async';
import 'dart:io';

Future<void> serveArtifacts({required String dir, int port = 8099}) async {
  final server = await HttpServer.bind(InternetAddress.loopbackIPv4, port);
  stdout.writeln('[network-debugger] Serving $dir at http://127.0.0.1:$port');
  await for (final req in server) {
    try {
      final path = req.uri.path;
      final localPath = path.startsWith('/') ? path.substring(1) : path;
      final file = File('${dir}${Platform.pathSeparator}$localPath');
      if (!await file.exists()) {
        req.response.statusCode = HttpStatus.notFound;
        await req.response.close();
        continue;
      }
      final stream = file.openRead();
      await req.response.addStream(stream);
      await req.response.close();
    } catch (_) {
      req.response.statusCode = HttpStatus.internalServerError;
      await req.response.close();
    }
  }
}
