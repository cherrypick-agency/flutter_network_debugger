/// Domain entity representing a custom font
class CustomFont {
  final String id;
  final String name;
  final String familyName;
  final DateTime uploadedAt;
  final int sizeInBytes;

  const CustomFont({
    required this.id,
    required this.name,
    required this.familyName,
    required this.uploadedAt,
    required this.sizeInBytes,
  });

  CustomFont copyWith({
    String? id,
    String? name,
    String? familyName,
    DateTime? uploadedAt,
    int? sizeInBytes,
  }) {
    return CustomFont(
      id: id ?? this.id,
      name: name ?? this.name,
      familyName: familyName ?? this.familyName,
      uploadedAt: uploadedAt ?? this.uploadedAt,
      sizeInBytes: sizeInBytes ?? this.sizeInBytes,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is CustomFont &&
          runtimeType == other.runtimeType &&
          id == other.id &&
          name == other.name &&
          familyName == other.familyName;

  @override
  int get hashCode => id.hashCode ^ name.hashCode ^ familyName.hashCode;

  @override
  String toString() {
    return 'CustomFont(id: $id, name: $name, familyName: $familyName, size: $sizeInBytes bytes)';
  }
}
