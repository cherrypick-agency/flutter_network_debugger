import 'package:flutter/material.dart';

class VisualDensityNotifier {
  static final ValueNotifier<VisualDensity> value =
      ValueNotifier<VisualDensity>(VisualDensity.standard);

  static VisualDensity fromString(String s) {
    switch (s) {
      case 'compact':
        return VisualDensity.compact;
      case 'comfortable':
        return VisualDensity.comfortable;
      default:
        return VisualDensity.standard;
    }
  }

  static String toStringValue(VisualDensity density) {
    if (density == VisualDensity.compact) return 'compact';
    if (density == VisualDensity.comfortable) return 'comfortable';
    return 'standard';
  }
}
