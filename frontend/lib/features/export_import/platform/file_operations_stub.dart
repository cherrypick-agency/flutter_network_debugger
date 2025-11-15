// Stub implementation for non-web platforms
// This file provides type stubs to allow compilation on non-web platforms
// where dart:html is not available.

import 'dart:async';

/// Stub for html.File (non-web platforms)
class File {
  String get name => '';
  int get size => 0;
  String get type => '';
}

/// Stub for html.FileReader (non-web platforms)
class FileReader {
  dynamic result;

  void readAsArrayBuffer(File file) {
    throw UnsupportedError(
      'File operations are only supported on web platform',
    );
  }

  void readAsText(File file) {
    throw UnsupportedError(
      'File operations are only supported on web platform',
    );
  }

  Stream<dynamic> get onLoad => Stream.empty();
  Stream<dynamic> get onLoadEnd => Stream.empty();
  Stream<dynamic> get onError => Stream.empty();
}

/// Stub for html.FileUploadInputElement (non-web platforms)
class FileUploadInputElement {
  String accept = '';
  List<File>? files;

  void click() {
    throw UnsupportedError(
      'File operations are only supported on web platform',
    );
  }

  Stream<dynamic> get onChange => Stream.empty();
}

/// Stub for html.DataTransfer (non-web platforms)
class DataTransfer {
  List<File> get files => [];
}

/// Stub for html.Event (non-web platforms)
class Event {
  void preventDefault() {}
  void stopPropagation() {}
  DataTransfer get dataTransfer => DataTransfer();
}

/// Stub for html.Document (non-web platforms)
class Document {
  Stream<Event> get onDragOver => Stream.empty();
  Stream<Event> get onDragLeave => Stream.empty();
  Stream<Event> get onDrop => Stream.empty();
}

/// Stub for html.Window (non-web platforms)
class Window {
  void open(String url, String target) {
    throw UnsupportedError(
      'Window operations are only supported on web platform',
    );
  }
}

/// Global stubs
final document = Document();
final window = Window();
