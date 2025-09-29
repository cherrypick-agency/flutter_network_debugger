import 'package:flutter/material.dart';
import 'app_theme.dart';

extension ThemeCtx on BuildContext {
  AppColors get appColors => Theme.of(this).extension<AppColors>()!;
  AppTextStyles get appText => Theme.of(this).extension<AppTextStyles>()!;
}



