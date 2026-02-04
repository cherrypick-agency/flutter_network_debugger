import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import '../providers/font_provider.dart';
import '../../infrastructure/utils/ttf_parser.dart';

/// Widget for selecting and managing custom fonts
/// Following Single Responsibility Principle
class CustomFontSelector extends StatefulWidget {
  final FontProvider provider;

  const CustomFontSelector({super.key, required this.provider});

  @override
  State<CustomFontSelector> createState() => _CustomFontSelectorState();
}

class _CustomFontSelectorState extends State<CustomFontSelector> {
  static const int _maxFontSizeBytes = 5 * 1024 * 1024; // 5 MB

  FontProvider get provider => widget.provider;

  @override
  void initState() {
    super.initState();
    provider.addListener(_onProviderChanged);
  }

  @override
  void dispose() {
    provider.removeListener(_onProviderChanged);
    super.dispose();
  }

  void _onProviderChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  Future<void> _pickFont() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['ttf', 'otf'],
        withData: true,
      );

      if (result == null || result.files.isEmpty) {
        return;
      }

      final file = result.files.first;
      if (file.bytes == null) {
        _showError('Failed to read font file');
        return;
      }

      final fontData = file.bytes!;

      // Validate file size
      if (fontData.length > _maxFontSizeBytes) {
        _showError(
          'Font file too large: ${_formatFileSize(fontData.length)}. '
          'Maximum size: ${_formatFileSize(_maxFontSizeBytes)}',
        );
        return;
      }

      // Validate file format
      if (!_isValidFontFile(fontData)) {
        _showError(
          'Invalid font file format. Only TTF and OTF fonts are supported.',
        );
        return;
      }

      // Extract family name from font metadata
      final fileName = file.name;
      String familyName =
          TtfParser.extractFontFamily(fontData) ??
          fileName.replaceAll(RegExp(r'\.(ttf|otf)$'), '');

      // Clean up family name - remove style suffixes if present
      familyName = familyName
          .replaceAll(
            RegExp(
              r'\s+(Regular|Bold|Italic|Light|Medium|SemiBold|Black|Thin|Heavy)$',
              caseSensitive: false,
            ),
            '',
          )
          .trim();

      // If family name is empty after cleanup, use filename
      if (familyName.isEmpty) {
        familyName = fileName.replaceAll(RegExp(r'\.(ttf|otf)$'), '');
      }

      // Load for preview (don't save)
      await provider.loadFontForPreview(
        fontData: fontData,
        fileName: fileName,
        familyName: familyName,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Font "$familyName" loaded for preview'),
            backgroundColor: Colors.blue,
          ),
        );
      }
    } catch (e) {
      _showError('Failed to load font: $e');
    }
  }

  void _removeFont() {
    // Mark font for removal (will be removed on Save)
    provider.markFontForRemoval();
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Font marked for removal (save to apply)'),
          backgroundColor: Colors.orange,
        ),
      );
    }
  }

  bool _isValidFontFile(Uint8List data) {
    if (data.length < 4) return false;

    // Check TTF magic number: 0x00010000
    if (data[0] == 0x00 &&
        data[1] == 0x01 &&
        data[2] == 0x00 &&
        data[3] == 0x00) {
      return true;
    }

    // Check OTF magic number: 'OTTO'
    final magic = String.fromCharCodes(data.sublist(0, 4));
    if (magic == 'OTTO') {
      return true;
    }

    // Check TrueType Collection: 'ttcf'
    if (magic == 'ttcf') {
      return true;
    }

    return false;
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }

  void _showError(String message) {
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(message), backgroundColor: Colors.red),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final hasFont = provider.effectivelyHasCustomFont;
    final font = provider.effectiveFont;
    final fontFamily = provider.previewFontFamily;
    final isPending = provider.hasPendingChanges;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Text('Custom font'),
            const SizedBox(width: 12),
            if (hasFont) ...[
              Expanded(
                child: Row(
                  children: [
                    Flexible(
                      child: Text(
                        font?.name ?? '',
                        style: const TextStyle(
                          fontSize: 12,
                          color: Colors.grey,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (isPending) ...[
                      const SizedBox(width: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 6,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.orange.withValues(alpha: 0.2),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: const Text(
                          'unsaved',
                          style: TextStyle(fontSize: 10, color: Colors.orange),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(width: 8),
              TextButton(
                onPressed: provider.isLoading ? null : _removeFont,
                child: const Text('Remove'),
              ),
            ] else ...[
              Expanded(
                child: Row(
                  children: [
                    const Text(
                      'System font',
                      style: TextStyle(fontSize: 12, color: Colors.grey),
                    ),
                    if (isPending) ...[
                      const SizedBox(width: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 6,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.orange.withValues(alpha: 0.2),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: const Text(
                          'unsaved',
                          style: TextStyle(fontSize: 10, color: Colors.orange),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(width: 8),
              ElevatedButton.icon(
                onPressed: provider.isLoading ? null : _pickFont,
                icon: const Icon(Icons.upload_file, size: 16),
                label: const Text('Load font'),
              ),
            ],
          ],
        ),
        if (hasFont) ...[
          const SizedBox(height: 12),
          const Text(
            'Preview:',
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w500,
              color: Colors.grey,
            ),
          ),
          const SizedBox(height: 4),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Small (12px) - The quick brown fox',
                style: TextStyle(fontFamily: fontFamily, fontSize: 12),
              ),
              const SizedBox(height: 2),
              Text(
                'Medium (14px) - The quick brown fox',
                style: TextStyle(fontFamily: fontFamily, fontSize: 14),
              ),
              const SizedBox(height: 2),
              Text(
                'Large (18px) - The quick brown fox',
                style: TextStyle(fontFamily: fontFamily, fontSize: 18),
              ),
            ],
          ),
        ],
        if (provider.isLoading) ...[
          const SizedBox(height: 8),
          const LinearProgressIndicator(),
        ],
        if (provider.error != null) ...[
          const SizedBox(height: 8),
          Text(
            provider.error!,
            style: const TextStyle(color: Colors.red, fontSize: 12),
          ),
        ],
      ],
    );
  }
}
