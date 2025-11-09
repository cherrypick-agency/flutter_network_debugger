import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import '../../application/stores/script_editor_store.dart';
import '../../domain/entities/script_test_result.dart';

/// Test tab for testing script execution
class ScriptTestTab extends StatefulWidget {
  final ScriptEditorStore store;

  const ScriptTestTab({super.key, required this.store});

  @override
  State<ScriptTestTab> createState() => _ScriptTestTabState();
}

class _ScriptTestTabState extends State<ScriptTestTab> {
  final _methodController = TextEditingController(text: 'GET');
  final _urlController = TextEditingController(
    text: 'https://api.example.com/users',
  );
  final _bodyController = TextEditingController();
  final List<HeaderPair> _headers = [];

  @override
  void dispose() {
    _methodController.dispose();
    _urlController.dispose();
    _bodyController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Observer(
      builder: (_) {
        return Row(
          children: [
            // Left: Request Builder
            Expanded(flex: 1, child: _buildRequestBuilder()),

            // Divider
            const VerticalDivider(width: 1),

            // Right: Results Viewer
            Expanded(flex: 1, child: _buildResultsViewer()),
          ],
        );
      },
    );
  }

  Widget _buildRequestBuilder() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Test Request', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 8),
          Text(
            'Build a sample HTTP request to test your script',
            style: TextStyle(color: Colors.grey[600]),
          ),
          const SizedBox(height: 24),

          // Method & URL
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Method dropdown
              SizedBox(
                width: 120,
                child: DropdownButtonFormField<String>(
                  value: _methodController.text,
                  decoration: const InputDecoration(
                    labelText: 'Method',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(value: 'GET', child: Text('GET')),
                    DropdownMenuItem(value: 'POST', child: Text('POST')),
                    DropdownMenuItem(value: 'PUT', child: Text('PUT')),
                    DropdownMenuItem(value: 'PATCH', child: Text('PATCH')),
                    DropdownMenuItem(value: 'DELETE', child: Text('DELETE')),
                    DropdownMenuItem(value: 'HEAD', child: Text('HEAD')),
                    DropdownMenuItem(value: 'OPTIONS', child: Text('OPTIONS')),
                  ],
                  onChanged: (value) {
                    if (value != null) {
                      _methodController.text = value;
                    }
                  },
                ),
              ),
              const SizedBox(width: 12),
              // URL
              Expanded(
                child: TextFormField(
                  controller: _urlController,
                  decoration: const InputDecoration(
                    labelText: 'URL',
                    border: OutlineInputBorder(),
                    hintText: 'https://api.example.com/users',
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),

          // Headers section
          Row(
            children: [
              Text('Headers', style: Theme.of(context).textTheme.titleMedium),
              const Spacer(),
              IconButton(
                icon: const Icon(Icons.add, size: 20),
                tooltip: 'Add Header',
                onPressed: () {
                  setState(() {
                    _headers.add(HeaderPair());
                  });
                },
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (_headers.isEmpty)
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.grey[100],
                borderRadius: BorderRadius.circular(8),
              ),
              child: Center(
                child: Text(
                  'No headers. Click + to add.',
                  style: TextStyle(color: Colors.grey[600]),
                ),
              ),
            )
          else
            ..._headers.asMap().entries.map((entry) {
              final index = entry.key;
              final header = entry.value;
              return Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: header.keyController,
                        decoration: const InputDecoration(
                          hintText: 'Header name',
                          border: OutlineInputBorder(),
                          isDense: true,
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: TextField(
                        controller: header.valueController,
                        decoration: const InputDecoration(
                          hintText: 'Header value',
                          border: OutlineInputBorder(),
                          isDense: true,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.delete, size: 20),
                      onPressed: () {
                        setState(() {
                          header.dispose();
                          _headers.removeAt(index);
                        });
                      },
                    ),
                  ],
                ),
              );
            }),
          const SizedBox(height: 24),

          // Body section
          Text('Body', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          TextField(
            controller: _bodyController,
            decoration: const InputDecoration(
              border: OutlineInputBorder(),
              hintText: 'Request body (JSON, text, etc.)',
            ),
            maxLines: 8,
            minLines: 4,
          ),
          const SizedBox(height: 24),

          // Execute button
          SizedBox(
            width: double.infinity,
            child: FilledButton.icon(
              onPressed: widget.store.isTesting ? null : _executeTest,
              icon: widget.store.isTesting
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
                      ),
                    )
                  : const Icon(Icons.play_arrow),
              label: Text(
                widget.store.isTesting ? 'Testing...' : 'Execute Test',
              ),
              style: FilledButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
          ),
          const SizedBox(height: 12),

          // Copy as cURL button
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: _copyCurlToClipboard,
              icon: const Icon(Icons.content_copy),
              label: const Text('Copy as cURL'),
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildResultsViewer() {
    final testResult = widget.store.testResult;
    final testError = widget.store.testError;

    if (testResult == null && testError == null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.science_outlined, size: 80, color: Colors.grey[400]),
            const SizedBox(height: 16),
            Text(
              'No test results yet',
              style: Theme.of(
                context,
              ).textTheme.titleLarge?.copyWith(color: Colors.grey[600]),
            ),
            const SizedBox(height: 8),
            Text(
              'Execute a test to see results here',
              style: TextStyle(color: Colors.grey[500]),
            ),
          ],
        ),
      );
    }

    if (testError != null) {
      return _buildErrorResult(testError);
    }

    return _buildSuccessResult(testResult!);
  }

  Widget _buildErrorResult(String error) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.error, color: Colors.red, size: 32),
              const SizedBox(width: 12),
              Text(
                'Test Failed',
                style: Theme.of(
                  context,
                ).textTheme.titleLarge?.copyWith(color: Colors.red),
              ),
            ],
          ),
          const SizedBox(height: 24),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.red[50],
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.red[200]!),
            ),
            child: SelectableText(
              error,
              style: TextStyle(fontFamily: 'monospace', color: Colors.red[900]),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSuccessResult(ScriptTestResult result) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Status header
          Row(
            children: [
              Icon(
                result.success ? Icons.check_circle : Icons.error,
                color: result.success ? Colors.green : Colors.red,
                size: 32,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      result.statusText,
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                        color: result.success ? Colors.green : Colors.red,
                      ),
                    ),
                    if (result.durationMs != null)
                      Text(
                        'Execution time: ${result.durationFormatted}',
                        style: TextStyle(color: Colors.grey[600], fontSize: 12),
                      ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),

          // Error message if failed
          if (!result.success && result.error != null) ...[
            _buildSection(
              'Error',
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.red[50],
                  borderRadius: BorderRadius.circular(8),
                ),
                child: SelectableText(
                  result.error!,
                  style: TextStyle(
                    fontFamily: 'monospace',
                    color: Colors.red[900],
                  ),
                ),
              ),
            ),
            const SizedBox(height: 16),
          ],

          // Modified Request
          if (result.modifiedRequest != null) ...[
            _buildSection(
              'Modified Request',
              _buildModifiedHTTP(result.modifiedRequest!, isRequest: true),
            ),
            const SizedBox(height: 16),
          ],

          // Modified Response
          if (result.modifiedResponse != null) ...[
            _buildSection(
              'Modified Response',
              _buildModifiedHTTP(result.modifiedResponse!, isRequest: false),
            ),
            const SizedBox(height: 16),
          ],

          // Logs
          if (result.logs.isNotEmpty) ...[
            _buildSection(
              'Logs',
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.grey[100],
                  borderRadius: BorderRadius.circular(8),
                ),
                child: SelectableText(
                  result.logsFormatted,
                  style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildSection(String title, Widget child) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: Theme.of(
            context,
          ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 8),
        child,
      ],
    );
  }

  Widget _buildModifiedHTTP(ModifiedHTTP modified, {required bool isRequest}) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.blue[50],
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.blue[200]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (modified.method != null) ...[
            _buildField('Method', modified.method!),
            const SizedBox(height: 8),
          ],
          if (modified.url != null) ...[
            _buildField('URL', modified.url!),
            const SizedBox(height: 8),
          ],
          if (!isRequest && modified.status != null) ...[
            _buildField('Status', modified.status.toString()),
            const SizedBox(height: 8),
          ],
          if (modified.headers != null && modified.headers!.isNotEmpty) ...[
            _buildField('Headers', _formatHeaders(modified.headers!)),
            const SizedBox(height: 8),
          ],
          if (modified.body != null) ...[_buildField('Body', modified.body!)],
        ],
      ),
    );
  }

  Widget _buildField(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontWeight: FontWeight.bold,
            fontSize: 12,
            color: Colors.black87,
          ),
        ),
        const SizedBox(height: 4),
        SelectableText(
          value,
          style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
        ),
      ],
    );
  }

  String _formatHeaders(Map<String, List<String>> headers) {
    return headers.entries
        .map((e) => '${e.key}: ${e.value.join(', ')}')
        .join('\n');
  }

  /// Generate cURL command from current request data
  String _generateCurlCommand() {
    final method = _methodController.text;
    final url = _urlController.text.trim();
    final body = _bodyController.text.trim();

    // Start with basic curl command
    final buffer = StringBuffer('curl');

    // Add method (skip for GET)
    if (method != 'GET') {
      buffer.write(' -X $method');
    }

    // Add URL (quoted to handle special characters)
    buffer.write(' \'$url\'');

    // Add headers
    for (final header in _headers) {
      final key = header.keyController.text.trim();
      final value = header.valueController.text.trim();
      if (key.isNotEmpty && value.isNotEmpty) {
        // Escape single quotes in header value
        final escapedValue = value.replaceAll('\'', '\'\\\'\'');
        buffer.write(' \\\n  -H \'$key: $escapedValue\'');
      }
    }

    // Add body (if present)
    if (body.isNotEmpty) {
      // Escape single quotes in body
      final escapedBody = body.replaceAll('\'', '\'\\\'\'');
      buffer.write(' \\\n  -d \'$escapedBody\'');
    }

    return buffer.toString();
  }

  /// Copy cURL command to clipboard
  Future<void> _copyCurlToClipboard() async {
    // Validate
    if (_urlController.text.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please enter a URL first'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    // Generate cURL command
    final curlCommand = _generateCurlCommand();

    // Copy to clipboard
    await Clipboard.setData(ClipboardData(text: curlCommand));

    // Show success message
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('cURL command copied to clipboard'),
        backgroundColor: Colors.green,
        duration: Duration(seconds: 2),
      ),
    );
  }

  Future<void> _executeTest() async {
    // Validate
    if (_urlController.text.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please enter a URL'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    // Build headers map
    final headers = <String, List<String>>{};
    for (final header in _headers) {
      final key = header.keyController.text.trim();
      final value = header.valueController.text.trim();
      if (key.isNotEmpty && value.isNotEmpty) {
        headers[key] = [value];
      }
    }

    // Build test request
    final testRequest = TestRequest(
      method: _methodController.text,
      url: _urlController.text.trim(),
      headers: headers,
      body: _bodyController.text.trim().isEmpty
          ? null
          : _bodyController.text.trim(),
    );

    // Execute test
    await widget.store.testScript(testRequest);
  }
}

/// Helper class for header key-value pairs
class HeaderPair {
  final TextEditingController keyController = TextEditingController();
  final TextEditingController valueController = TextEditingController();

  void dispose() {
    keyController.dispose();
    valueController.dispose();
  }
}
