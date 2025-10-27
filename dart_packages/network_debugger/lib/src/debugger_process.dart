import 'dart:async';
import 'dart:io';

/// Manages the network debugger process.
class DebuggerProcess {
  final String binaryPath;
  final int port;
  final bool autoOpenBrowser;
  final Map<String, String> environment;

  Process? _process;
  final _outputController = StreamController<String>.broadcast();
  final _errorController = StreamController<String>.broadcast();

  DebuggerProcess({
    required this.binaryPath,
    required this.port,
    this.autoOpenBrowser = true,
    Map<String, String>? environment,
  }) : environment = environment ?? {};

  /// Stream of stdout output from the process.
  Stream<String> get stdout => _outputController.stream;

  /// Stream of stderr output from the process.
  Stream<String> get stderr => _errorController.stream;

  /// Whether the process is currently running.
  bool get isRunning => _process != null;

  /// The URL where the debugger web UI is accessible.
  String get url => 'http://localhost:$port';

  /// Starts the debugger process.
  Future<void> start() async {
    if (_process != null) {
      throw ProcessException(
        'Process is already running',
      );
    }

    final binary = File(binaryPath);
    if (!await binary.exists()) {
      throw ProcessException(
        'Binary not found at: $binaryPath',
      );
    }

    // Prepare environment
    final env = <String, String>{
      ...Platform.environment,
      'ADDR': ':$port',
      if (!autoOpenBrowser) 'DEV_MODE': '1',
      ...environment,
    };

    try {
      _process = await Process.start(
        binaryPath,
        [],
        environment: env,
        mode: ProcessStartMode.normal,
      );

      // Pipe output streams
      _process!.stdout.transform(const SystemEncoding().decoder).listen(
        (line) {
          _outputController.add(line);
        },
        onError: (error) {
          _errorController.addError(error);
        },
      );

      _process!.stderr.transform(const SystemEncoding().decoder).listen(
        (line) {
          _errorController.add(line);
        },
        onError: (error) {
          _errorController.addError(error);
        },
      );

      // Wait a bit for the process to start
      await Future.delayed(const Duration(milliseconds: 500));

      // Check if process is still running
      if (!isRunning) {
        throw ProcessException(
          'Process exited immediately after start',
        );
      }
    } catch (e) {
      _process = null;
      throw ProcessException(
        'Failed to start process: $e',
      );
    }
  }

  /// Stops the debugger process gracefully.
  Future<void> stop({Duration timeout = const Duration(seconds: 5)}) async {
    if (_process == null) {
      return;
    }

    // Try graceful shutdown first
    _process!.kill(ProcessSignal.sigterm);

    // Wait for process to exit
    final completer = Completer<void>();
    Timer? timeoutTimer;

    _process!.exitCode.then((code) {
      timeoutTimer?.cancel();
      completer.complete();
    });

    timeoutTimer = Timer(timeout, () {
      if (!completer.isCompleted) {
        // Force kill if timeout
        _process?.kill(ProcessSignal.sigkill);
        completer.complete();
      }
    });

    await completer.future;
    _process = null;

    // Close streams
    await _outputController.close();
    await _errorController.close();
  }

  /// Waits for the process to become ready by checking if the port is listening.
  Future<bool> waitUntilReady({
    Duration timeout = const Duration(seconds: 10),
    Duration checkInterval = const Duration(milliseconds: 100),
  }) async {
    final endTime = DateTime.now().add(timeout);

    while (DateTime.now().isBefore(endTime)) {
      if (!isRunning) {
        return false;
      }

      // Try to connect to the port
      try {
        final socket = await Socket.connect(
          'localhost',
          port,
          timeout: const Duration(milliseconds: 500),
        );
        await socket.close();
        return true;
      } catch (_) {
        // Port not ready yet
      }

      await Future.delayed(checkInterval);
    }

    return false;
  }

  /// Gets the process ID if running.
  int? get pid => _process?.pid;
}

/// Instance of a running debugger.
class DebuggerInstance {
  final DebuggerProcess _process;

  DebuggerInstance(this._process);

  /// Stops the debugger.
  Future<void> stop() => _process.stop();

  /// Whether the debugger is running.
  bool get isRunning => _process.isRunning;

  /// The URL where the debugger is accessible.
  String get url => _process.url;

  /// The port the debugger is running on.
  int get port => _process.port;

  /// Stream of stdout output.
  Stream<String> get stdout => _process.stdout;

  /// Stream of stderr output.
  Stream<String> get stderr => _process.stderr;

  /// Process ID.
  int? get pid => _process.pid;

  /// Waits for the debugger to be ready.
  Future<bool> waitUntilReady({
    Duration timeout = const Duration(seconds: 10),
  }) => _process.waitUntilReady(timeout: timeout);
}

/// Exception thrown when process operations fail.
class ProcessException implements Exception {
  final String message;

  ProcessException(this.message);

  @override
  String toString() => 'ProcessException: $message';
}
