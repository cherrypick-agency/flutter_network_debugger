import 'package:web/web.dart' as web;
import 'package:js/js_util.dart' as js_util;

String detectOS() {
  final plat = (web.window.navigator.platform).toLowerCase();
  final ua = (web.window.navigator.userAgent).toLowerCase();
  if (plat.contains('win') || ua.contains('windows')) return 'win';
  if (plat.contains('mac') ||
      ua.contains('mac os') ||
      ua.contains('macos') ||
      ua.contains('iphone') ||
      ua.contains('ipad'))
    return 'mac';
  return 'linux';
}

String detectArch() {
  final ua = (web.window.navigator.userAgent).toLowerCase();
  if (ua.contains('aarch64') ||
      ua.contains('arm64') ||
      ua.contains('armv8') ||
      ua.contains('apple m'))
    return 'arm64';
  if (ua.contains('wow64') || ua.contains('win64')) return 'amd64';
  if (ua.contains('win32') || ua.contains('ia32')) return '386';
  return 'amd64';
}

String macArchLabel(String arch) => arch == 'arm64' ? 'Apple Silicon' : 'Intel';

Future<String?> detectArchPrecise() async {
  try {
    final nav = web.window.navigator as dynamic;
    final uad = nav.userAgentData;
    if (uad == null || uad.getHighEntropyValues == null) return null;
    final promise =
        uad.getHighEntropyValues(
              js_util.jsify(["architecture", "bitness", "platform"]),
            )
            as Object;
    final result = await js_util.promiseToFuture<Object>(promise);
    final dyn = result as dynamic;
    final arch = (dyn.architecture ?? '').toString().toLowerCase();
    final bit = (dyn.bitness ?? '').toString().toLowerCase();
    if (arch.contains('arm')) return 'arm64';
    if (bit == '32') return '386';
    // Если userAgentData доступен и не ARM — считаем amd64
    return 'amd64';
  } catch (_) {
    // Падение/нет userAgentData — попробуем эвристику через WebGL на macOS
    try {
      if (detectOS() != 'mac') return null;
      final canvas = web.document.createElement('canvas');
      final gl =
          js_util.callMethod(canvas, 'getContext', ['webgl']) ??
          js_util.callMethod(canvas, 'getContext', ['experimental-webgl']);
      if (gl == null) return null;

      final ext = js_util.callMethod(gl, 'getExtension', [
        'WEBGL_debug_renderer_info',
      ]);
      if (ext != null) {
        final unmasked = js_util.getProperty(ext, 'UNMASKED_RENDERER_WEBGL');
        final renderer = js_util.callMethod(gl, 'getParameter', [unmasked]);
        final r = (renderer?.toString() ?? '').toLowerCase();
        // Если в Chromium выдаёт конкретный Apple M*, это ARM
        if (r.contains('apple') && !r.contains('apple gpu')) return 'arm64';
      }

      // В Safari рендерер маскируется как "Apple GPU" — используем отсутствующую S3TC_sRGB как слабый сигнал ARM
      final supported =
          js_util.callMethod(gl, 'getSupportedExtensions', []) as Object?;
      if (supported is List) {
        final hasS3tcSrgb = supported.any(
          (e) =>
              e.toString().toLowerCase() ==
              'webgl_compressed_texture_s3tc_srgb',
        );
        if (!hasS3tcSrgb) return 'arm64';
      }
      return null;
    } catch (_) {
      return null;
    }
  }
}
