import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:desktop_drop/desktop_drop.dart';
import 'package:file_picker/file_picker.dart';

/// WASM file upload widget with drag-n-drop support
class WasmUploadZone extends StatefulWidget {
  final ValueChanged<String> onWasmUploaded; // base64 encoded WASM
  final VoidCallback? onClear;
  final String? currentFileName;

  const WasmUploadZone({
    super.key,
    required this.onWasmUploaded,
    this.onClear,
    this.currentFileName,
  });

  @override
  State<WasmUploadZone> createState() => _WasmUploadZoneState();
}

class _WasmUploadZoneState extends State<WasmUploadZone> {
  bool _isDragging = false;
  bool _isProcessing = false;
  String? _uploadedFileName;
  int? _fileSize;
  String? _error;

  static const int _maxFileSizeMB = 10;
  static const int _maxFileSizeBytes = _maxFileSizeMB * 1024 * 1024;

  @override
  void initState() {
    super.initState();
    _uploadedFileName = widget.currentFileName;
  }

  @override
  Widget build(BuildContext context) {
    return DropTarget(
      onDragEntered: (_) => setState(() => _isDragging = true),
      onDragExited: (_) => setState(() => _isDragging = false),
      onDragDone: (details) async {
        setState(() => _isDragging = false);
        if (details.files.isNotEmpty) {
          await _processFile(
            details.files.first.path,
            details.files.first.readAsBytes,
          );
        }
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        decoration: BoxDecoration(
          border: Border.all(
            color: _isDragging
                ? Theme.of(context).colorScheme.primary
                : (_error != null ? Colors.red : Colors.grey.shade400),
            width: _isDragging ? 3 : 2,
            strokeAlign: BorderSide.strokeAlignInside,
          ),
          borderRadius: BorderRadius.circular(12),
          color: _isDragging
              ? Theme.of(context).colorScheme.primaryContainer.withOpacity(0.1)
              : Colors.transparent,
        ),
        child: InkWell(
          onTap: _isProcessing ? null : _pickFile,
          borderRadius: BorderRadius.circular(12),
          child: Padding(
            padding: const EdgeInsets.all(32),
            child: _isProcessing
                ? _buildProcessing()
                : (_uploadedFileName != null
                      ? _buildUploaded()
                      : _buildEmpty()),
          ),
        ),
      ),
    );
  }

  Widget _buildEmpty() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(
          _isDragging ? Icons.file_download : Icons.cloud_upload_outlined,
          size: 64,
          color: _error != null
              ? Colors.red
              : Theme.of(context).colorScheme.primary,
        ),
        const SizedBox(height: 16),
        Text(
          _isDragging ? 'Drop WASM file here' : 'Upload WASM File',
          style: Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 8),
        Text(
          'Drag & drop a .wasm file or click to browse',
          style: TextStyle(color: Colors.grey[600]),
        ),
        const SizedBox(height: 4),
        Text(
          'Maximum file size: $_maxFileSizeMB MB',
          style: TextStyle(color: Colors.grey[500], fontSize: 12),
        ),
        if (_error != null) ...[
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.red.shade50,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.error_outline, color: Colors.red, size: 20),
                const SizedBox(width: 8),
                Flexible(
                  child: Text(
                    _error!,
                    style: const TextStyle(color: Colors.red),
                  ),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildUploaded() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Icon(Icons.check_circle, size: 64, color: Colors.green),
        const SizedBox(height: 16),
        Text(
          'WASM File Uploaded',
          style: Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 8),
        Text(
          _uploadedFileName!,
          style: const TextStyle(fontFamily: 'monospace'),
        ),
        if (_fileSize != null) ...[
          const SizedBox(height: 4),
          Text(
            _formatFileSize(_fileSize!),
            style: TextStyle(color: Colors.grey[600], fontSize: 12),
          ),
        ],
        const SizedBox(height: 16),
        Wrap(
          spacing: 8,
          children: [
            OutlinedButton.icon(
              onPressed: _pickFile,
              icon: const Icon(Icons.swap_horiz, size: 18),
              label: const Text('Replace'),
            ),
            OutlinedButton.icon(
              onPressed: () {
                setState(() {
                  _uploadedFileName = null;
                  _fileSize = null;
                  _error = null;
                });
                widget.onClear?.call();
              },
              icon: const Icon(Icons.clear, size: 18),
              label: const Text('Clear'),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildProcessing() {
    return const Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        CircularProgressIndicator(),
        SizedBox(height: 16),
        Text('Processing WASM file...'),
      ],
    );
  }

  Future<void> _pickFile() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['wasm', 'wat'],
        withData: true,
      );

      if (result != null && result.files.isNotEmpty) {
        final file = result.files.first;
        if (file.bytes != null) {
          await _processFileBytes(file.name, file.bytes!);
        }
      }
    } catch (e) {
      setState(() {
        _error = 'Failed to pick file: ${e.toString()}';
      });
    }
  }

  Future<void> _processFile(
    String path,
    Future<Uint8List> Function() readBytes,
  ) async {
    setState(() {
      _isProcessing = true;
      _error = null;
    });

    try {
      final bytes = await readBytes();
      final fileName = path.split('/').last;
      await _processFileBytes(fileName, bytes);
    } catch (e) {
      setState(() {
        _error = 'Failed to read file: ${e.toString()}';
        _isProcessing = false;
      });
    }
  }

  Future<void> _processFileBytes(String fileName, Uint8List bytes) async {
    try {
      // Validate file extension
      if (!fileName.endsWith('.wasm') && !fileName.endsWith('.wat')) {
        throw Exception('Only .wasm and .wat files are supported');
      }

      // Validate file size
      if (bytes.length > _maxFileSizeBytes) {
        throw Exception('File size exceeds $_maxFileSizeMB MB limit');
      }

      // Convert to base64
      final base64String = base64Encode(bytes);

      setState(() {
        _uploadedFileName = fileName;
        _fileSize = bytes.length;
        _isProcessing = false;
        _error = null;
      });

      // Notify parent
      widget.onWasmUploaded(base64String);
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isProcessing = false;
      });
    }
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(2)} MB';
  }
}
