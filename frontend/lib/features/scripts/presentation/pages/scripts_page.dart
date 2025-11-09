import 'package:flutter/material.dart';

/// Scripts Management Page
/// TODO: Add full implementation with Monaco editor, filters, etc.
class ScriptsPage extends StatelessWidget {
  const ScriptsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Scripts'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            tooltip: 'Create Script',
            onPressed: () {
              // TODO: Open script editor dialog
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Script editor coming soon!')),
              );
            },
          ),
        ],
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.code, size: 100, color: Colors.grey),
            const SizedBox(height: 24),
            Text(
              'Scripts Feature',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'Create and manage scripts to modify HTTP traffic',
              style: TextStyle(color: Colors.grey),
            ),
            const SizedBox(height: 32),
            const Text(
              '✅ Domain Layer (Entities, Use Cases)',
              style: TextStyle(color: Colors.green),
            ),
            const Text(
              '✅ Data Layer (API, Repository)',
              style: TextStyle(color: Colors.green),
            ),
            const Text(
              '✅ Application Layer (MobX Stores)',
              style: TextStyle(color: Colors.green),
            ),
            const SizedBox(height: 16),
            const Text(
              '🚧 Presentation Layer - In Development',
              style: TextStyle(color: Colors.orange),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () {
                // TODO: Show implementation plan
              },
              icon: const Icon(Icons.info_outline),
              label: const Text('View Implementation Plan'),
            ),
          ],
        ),
      ),
    );
  }
}
