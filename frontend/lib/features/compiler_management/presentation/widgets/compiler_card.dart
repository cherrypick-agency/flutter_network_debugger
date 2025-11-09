import 'package:flutter/material.dart';
import '../../domain/entities/compiler_info.dart';
import '../../domain/entities/download_progress.dart';

class CompilerCard extends StatelessWidget {
  final CompilerInfo compiler;
  final DownloadProgress? progress;
  final VoidCallback? onInstall;
  final VoidCallback? onUninstall;

  const CompilerCard({
    super.key,
    required this.compiler,
    this.progress,
    this.onInstall,
    this.onUninstall,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Card(
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: compiler.isInstalled
              ? colorScheme.primary.withOpacity(0.3)
              : colorScheme.outline.withOpacity(0.1),
          width: 1,
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                _buildIcon(colorScheme),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _getDisplayName(compiler.language),
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        compiler.version,
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                ),
                _buildStatusBadge(theme),
              ],
            ),
            const SizedBox(height: 12),
            if (progress != null) ...[
              _buildProgressBar(theme, progress!),
              const SizedBox(height: 8),
              Text(
                progress!.progressText,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: colorScheme.onSurfaceVariant,
                ),
              ),
            ] else ...[
              Row(
                children: [
                  Icon(
                    Icons.storage_outlined,
                    size: 16,
                    color: colorScheme.onSurfaceVariant,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    compiler.displaySize,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ],
            const SizedBox(height: 12),
            _buildActions(theme),
          ],
        ),
      ),
    );
  }

  Widget _buildIcon(ColorScheme colorScheme) {
    final icon = _getLanguageIcon(compiler.language);
    final color = compiler.isInstalled
        ? colorScheme.primary
        : colorScheme.onSurfaceVariant;

    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Icon(icon, size: 28, color: color),
    );
  }

  Widget _buildStatusBadge(ThemeData theme) {
    final colorScheme = theme.colorScheme;

    Color backgroundColor;
    Color textColor;
    String text;

    if (compiler.isInstalling || progress != null) {
      backgroundColor = colorScheme.tertiary.withOpacity(0.1);
      textColor = colorScheme.tertiary;
      text = 'Installing';
    } else if (compiler.isInstalled) {
      backgroundColor = colorScheme.primary.withOpacity(0.1);
      textColor = colorScheme.primary;
      text = 'Installed';
    } else {
      backgroundColor = colorScheme.surfaceVariant;
      textColor = colorScheme.onSurfaceVariant;
      text = 'Not Installed';
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        text,
        style: theme.textTheme.labelSmall?.copyWith(
          color: textColor,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  Widget _buildProgressBar(ThemeData theme, DownloadProgress progress) {
    return LinearProgressIndicator(
      value: progress.percentage / 100,
      backgroundColor: theme.colorScheme.surfaceVariant,
      valueColor: AlwaysStoppedAnimation<Color>(theme.colorScheme.tertiary),
      borderRadius: BorderRadius.circular(4),
      minHeight: 6,
    );
  }

  Widget _buildActions(ThemeData theme) {
    if (compiler.isInstalling || progress != null) {
      return SizedBox(
        width: double.infinity,
        child: OutlinedButton.icon(
          onPressed: null,
          icon: const SizedBox(
            width: 16,
            height: 16,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
          label: const Text('Installing...'),
        ),
      );
    }

    if (compiler.isInstalled) {
      return SizedBox(
        width: double.infinity,
        child: OutlinedButton.icon(
          onPressed: onUninstall,
          icon: const Icon(Icons.delete_outline, size: 18),
          label: const Text('Uninstall'),
          style: OutlinedButton.styleFrom(
            foregroundColor: theme.colorScheme.error,
            side: BorderSide(color: theme.colorScheme.error),
          ),
        ),
      );
    }

    return SizedBox(
      width: double.infinity,
      child: FilledButton.icon(
        onPressed: onInstall,
        icon: const Icon(Icons.download, size: 18),
        label: const Text('Install'),
      ),
    );
  }

  IconData _getLanguageIcon(String language) {
    switch (language.toLowerCase()) {
      case 'zig':
        return Icons.flash_on;
      case 'kotlin':
        return Icons.code;
      case 'swift':
        return Icons.apple;
      case 'tinygo':
      case 'go':
        return Icons.pets;
      case 'rust':
        return Icons.security;
      case 'assemblyscript':
        return Icons.javascript;
      case 'wasi-sdk':
      case 'c':
      case 'cpp':
        return Icons.build;
      default:
        return Icons.code_outlined;
    }
  }

  String _getDisplayName(String language) {
    switch (language.toLowerCase()) {
      case 'tinygo':
        return 'TinyGo';
      case 'assemblyscript':
        return 'AssemblyScript';
      case 'wasi-sdk':
        return 'WASI-SDK (C/C++)';
      default:
        return language[0].toUpperCase() + language.substring(1);
    }
  }
}
