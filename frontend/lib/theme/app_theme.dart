import 'package:flutter/material.dart';

class AppColors extends ThemeExtension<AppColors> {
  final Color background;
  final Color surface;
  final Color primary;
  final Color textPrimary;
  final Color textSecondary;
  final Color success;
  final Color warning;
  final Color danger;
  final Color border;

  const AppColors({
    required this.background,
    required this.surface,
    required this.primary,
    required this.textPrimary,
    required this.textSecondary,
    required this.success,
    required this.warning,
    required this.danger,
    required this.border,
  });

  @override
  AppColors copyWith({
    Color? background,
    Color? surface,
    Color? primary,
    Color? textPrimary,
    Color? textSecondary,
    Color? success,
    Color? warning,
    Color? danger,
    Color? border,
  }) => AppColors(
    background: background ?? this.background,
    surface: surface ?? this.surface,
    primary: primary ?? this.primary,
    textPrimary: textPrimary ?? this.textPrimary,
    textSecondary: textSecondary ?? this.textSecondary,
    success: success ?? this.success,
    warning: warning ?? this.warning,
    danger: danger ?? this.danger,
    border: border ?? this.border,
  );

  @override
  AppColors lerp(ThemeExtension<AppColors>? other, double t) {
    if (other is! AppColors) return this;
    return AppColors(
      background: Color.lerp(background, other.background, t)!,
      surface: Color.lerp(surface, other.surface, t)!,
      primary: Color.lerp(primary, other.primary, t)!,
      textPrimary: Color.lerp(textPrimary, other.textPrimary, t)!,
      textSecondary: Color.lerp(textSecondary, other.textSecondary, t)!,
      success: Color.lerp(success, other.success, t)!,
      warning: Color.lerp(warning, other.warning, t)!,
      danger: Color.lerp(danger, other.danger, t)!,
      border: Color.lerp(border, other.border, t)!,
    );
  }
}

class AppTextStyles extends ThemeExtension<AppTextStyles> {
  final TextStyle title;
  final TextStyle subtitle;
  final TextStyle body;
  final TextStyle monospace;

  const AppTextStyles({
    required this.title,
    required this.subtitle,
    required this.body,
    required this.monospace,
  });

  @override
  AppTextStyles copyWith({
    TextStyle? title,
    TextStyle? subtitle,
    TextStyle? body,
    TextStyle? monospace,
  }) => AppTextStyles(
    title: title ?? this.title,
    subtitle: subtitle ?? this.subtitle,
    body: body ?? this.body,
    monospace: monospace ?? this.monospace,
  );

  @override
  AppTextStyles lerp(ThemeExtension<AppTextStyles>? other, double t) {
    if (other is! AppTextStyles) return this;
    return AppTextStyles(
      title: TextStyle.lerp(title, other.title, t)!,
      subtitle: TextStyle.lerp(subtitle, other.subtitle, t)!,
      body: TextStyle.lerp(body, other.body, t)!,
      monospace: TextStyle.lerp(monospace, other.monospace, t)!,
    );
  }
}

ThemeData buildLightTheme() {
  final scheme = ColorScheme.fromSeed(seedColor: const Color(0xFF3F51B5));
  return ThemeData(
    colorScheme: scheme,
    useMaterial3: true,
    extensions: [
      AppColors(
        background: scheme.background,
        surface: scheme.surface,
        primary: scheme.primary,
        textPrimary: scheme.onSurface,
        textSecondary: scheme.onSurface.withOpacity(0.7),
        success: const Color(0xFF2E7D32),
        warning: const Color(0xFFF9A825),
        danger: const Color(0xFFC62828),
        border: scheme.outline,
      ),
      const AppTextStyles(
        title: TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
        subtitle: TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
        body: TextStyle(fontSize: 13),
        monospace: TextStyle(fontSize: 12, fontFamily: 'monospace'),
      ),
    ],
  );
}

ThemeData buildDarkTheme() {
  final scheme = ColorScheme.fromSeed(seedColor: const Color(0xFF90CAF9), brightness: Brightness.dark);
  return ThemeData(
    colorScheme: scheme,
    useMaterial3: true,
    extensions: [
      AppColors(
        background: scheme.background,
        surface: scheme.surface,
        primary: scheme.primary,
        textPrimary: scheme.onSurface,
        textSecondary: scheme.onSurface.withOpacity(0.7),
        success: const Color(0xFF66BB6A),
        warning: const Color(0xFFFFCA28),
        danger: const Color(0xFFEF5350),
        border: scheme.outline,
      ),
      const AppTextStyles(
        title: TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
        subtitle: TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
        body: TextStyle(fontSize: 13),
        monospace: TextStyle(fontSize: 12, fontFamily: 'monospace'),
      ),
    ],
  );
}



