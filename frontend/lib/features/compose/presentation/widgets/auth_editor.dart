import 'package:flutter/material.dart';

class AuthEditor extends StatelessWidget {
  const AuthEditor({
    super.key,
    required this.authType,
    required this.onAuthTypeChanged,
    required this.basicUser,
    required this.basicPass,
    required this.bearer,
    required this.apiKeyHeader,
    required this.apiKey,
  });

  final String authType; // none|basic|bearer|apiKey
  final ValueChanged<String> onAuthTypeChanged;
  final TextEditingController basicUser;
  final TextEditingController basicPass;
  final TextEditingController bearer;
  final TextEditingController apiKeyHeader;
  final TextEditingController apiKey;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          DropdownButtonFormField<String>(
            value: authType,
            onChanged: (v) => onAuthTypeChanged(v ?? 'none'),
            items:
                const ['none', 'basic', 'bearer', 'apiKey']
                    .map((t) => DropdownMenuItem(value: t, child: Text(t)))
                    .toList(),
          ),
          const SizedBox(height: 8),
          if (authType == 'basic') ...[
            TextFormField(
              controller: basicUser,
              decoration: const InputDecoration(labelText: 'Username'),
            ),
            TextFormField(
              controller: basicPass,
              decoration: const InputDecoration(labelText: 'Password'),
              obscureText: true,
            ),
          ] else if (authType == 'bearer') ...[
            TextFormField(
              controller: bearer,
              decoration: const InputDecoration(labelText: 'Bearer token'),
            ),
          ] else if (authType == 'apiKey') ...[
            TextFormField(
              controller: apiKeyHeader,
              decoration: const InputDecoration(labelText: 'Header name'),
            ),
            TextFormField(
              controller: apiKey,
              decoration: const InputDecoration(labelText: 'API Key'),
            ),
          ],
        ],
      ),
    );
  }
}
