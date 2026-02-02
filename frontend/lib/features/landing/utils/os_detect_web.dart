import 'dart:js_interop';

import 'package:web/web.dart' as web;

String detectOS() {
  final plat = (web.window.navigator.platform).toLowerCase();
  final ua = (web.window.navigator.userAgent).toLowerCase();
  if (plat.contains('win') || ua.contains('windows')) return 'win';
  if (plat.contains('mac') ||
      ua.contains('mac os') ||
      ua.contains('macos') ||
      ua.contains('iphone') ||
      ua.contains('ipad')) {
    return 'mac';
  }
  return 'linux';
}

String detectArch() {
  final ua = (web.window.navigator.userAgent).toLowerCase();
  if (ua.contains('aarch64') ||
      ua.contains('arm64') ||
      ua.contains('armv8') ||
      ua.contains('apple m')) {
    return 'arm64';
  }
  if (ua.contains('wow64') || ua.contains('win64')) return 'amd64';
  if (ua.contains('win32') || ua.contains('ia32')) return '386';
  return 'amd64';
}

String macArchLabel(String arch) => arch == 'arm64' ? 'Apple Silicon' : 'Intel';

/// Биндинги для NavigatorUAData (User-Agent Client Hints API)
@JS('navigator.userAgentData')
external NavigatorUAData? get _userAgentData;

extension type NavigatorUAData._(JSObject _) implements JSObject {
  external JSPromise<UADataValues> getHighEntropyValues(
    JSArray<JSString> hints,
  );
}

extension type UADataValues._(JSObject _) implements JSObject {
  external String? get architecture;
  external String? get bitness;
  external String? get platform;
}

/// Биндинги для WebGL debug extension
extension type WebGLDebugRendererInfo._(JSObject _) implements JSObject {
  @JS('UNMASKED_RENDERER_WEBGL')
  external int get unmaskedRendererWebGL;
}

Future<String?> detectArchPrecise() async {
  try {
    final uad = _userAgentData;
    if (uad == null) return null;

    final hints = <JSString>[
      'architecture'.toJS,
      'bitness'.toJS,
      'platform'.toJS,
    ].toJS;
    final result = await uad.getHighEntropyValues(hints).toDart;

    final arch = (result.architecture ?? '').toLowerCase();
    final bit = (result.bitness ?? '').toLowerCase();

    if (arch.contains('arm')) return 'arm64';
    if (bit == '32') return '386';
    // Если userAgentData доступен и не ARM — считаем amd64
    return 'amd64';
  } catch (_) {
    // Падение/нет userAgentData — попробуем эвристику через WebGL на macOS
    try {
      if (detectOS() != 'mac') return null;

      final canvas =
          web.document.createElement('canvas') as web.HTMLCanvasElement;
      final gl =
          canvas.getContext('webgl') ?? canvas.getContext('experimental-webgl');
      if (gl == null) return null;

      final glObj = gl as web.WebGLRenderingContext;
      final ext = glObj.getExtension('WEBGL_debug_renderer_info');

      if (ext != null) {
        final extObj = ext as WebGLDebugRendererInfo;
        final renderer = glObj.getParameter(extObj.unmaskedRendererWebGL);
        final r = (renderer?.toString() ?? '').toLowerCase();
        // Если в Chromium выдаёт конкретный Apple M*, это ARM
        if (r.contains('apple') && !r.contains('apple gpu')) return 'arm64';
      }

      // В Safari рендерер маскируется как "Apple GPU" — используем отсутствующую S3TC_sRGB как слабый сигнал ARM
      final supported = glObj.getSupportedExtensions();
      if (supported != null) {
        final extensions = supported.toDart;
        final hasS3tcSrgb = extensions.any(
          (e) => e.toDart.toLowerCase() == 'webgl_compressed_texture_s3tc_srgb',
        );
        if (!hasS3tcSrgb) return 'arm64';
      }
      return null;
    } catch (_) {
      return null;
    }
  }
}
