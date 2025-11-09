import 'package:equatable/equatable.dart';

class PredefinedTag extends Equatable {
  const PredefinedTag({
    required this.id,
    required this.name,
    required this.color,
    required this.category,
    required this.displayOrder,
    required this.createdAt,
  });

  final String id;
  final String name;
  final String color;
  final String category;
  final int displayOrder;
  final DateTime createdAt;

  factory PredefinedTag.fromJson(Map<String, dynamic> json) {
    return PredefinedTag(
      id: json['id'] as String,
      name: json['name'] as String,
      color: json['color'] as String,
      category: json['category'] as String,
      displayOrder: json['displayOrder'] as int,
      createdAt: DateTime.parse(json['createdAt'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'color': color,
      'category': category,
      'displayOrder': displayOrder,
      'createdAt': createdAt.toIso8601String(),
    };
  }

  @override
  List<Object?> get props => [
    id,
    name,
    color,
    category,
    displayOrder,
    createdAt,
  ];
}
