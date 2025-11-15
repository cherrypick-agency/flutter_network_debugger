import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:hex_viewer/hex_viewer.dart';

void main() => runApp(const HexViewerExampleApp());

class HexViewerExampleApp extends StatelessWidget {
  const HexViewerExampleApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Hex Viewer Example',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        useMaterial3: true,
      ),
      darkTheme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.blue,
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
      ),
      home: const HexViewerExamplePage(),
    );
  }
}

class HexViewerExamplePage extends StatefulWidget {
  const HexViewerExamplePage({super.key});

  @override
  State<HexViewerExamplePage> createState() => _HexViewerExamplePageState();
}

class _HexViewerExamplePageState extends State<HexViewerExamplePage> {
  ByteSelection? _selection;
  int _bytesPerLine = 16;
  int _groupSize = 8;

  // Sample data: JSON string encoded as UTF-8
  late final List<int> _sampleData;

  @override
  void initState() {
    super.initState();
    // Create sample data: a JSON object
    const sampleJson = {
      'message': 'Hello, Hex Viewer!',
      'version': '0.1.0',
      'features': ['selection', 'copy', 'grouping'],
      'unicode': '你好世界 🌍',
    };
    _sampleData = utf8.encode(json.encode(sampleJson));
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Hex Viewer Example'),
        backgroundColor: colorScheme.inversePrimary,
      ),
      body: Column(
        children: [
          // Controls
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Configuration',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Bytes per line: $_bytesPerLine'),
                          Slider(
                            value: _bytesPerLine.toDouble(),
                            min: 8,
                            max: 32,
                            divisions: 3,
                            label: '$_bytesPerLine',
                            onChanged: (value) {
                              setState(() {
                                _bytesPerLine = value.toInt();
                              });
                            },
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Group size: $_groupSize'),
                          Slider(
                            value: _groupSize.toDouble(),
                            min: 1,
                            max: 16,
                            divisions: 3,
                            label: '$_groupSize',
                            onChanged: (value) {
                              setState(() {
                                _groupSize = value.toInt();
                              });
                            },
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                if (_selection != null) ...[
                  const Divider(),
                  Text(
                    'Selected: ${_selection!.length} bytes '
                    '(0x${_selection!.start.toRadixString(16).toUpperCase()} - '
                    '0x${_selection!.end.toRadixString(16).toUpperCase()})',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: colorScheme.primary,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ],
            ),
          ),
          const Divider(height: 1),
          // Hex Viewer
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: HexViewer(
                data: _sampleData,
                config: HexConfig(
                  bytesPerLine: _bytesPerLine,
                  groupSize: _groupSize,
                  colorScheme: ByteColorScheme.fromTheme(Theme.of(context)),
                ),
                onSelectionChanged: (selection) {
                  setState(() {
                    _selection = selection;
                  });
                },
              ),
            ),
          ),
        ],
      ),
    );
  }
}
