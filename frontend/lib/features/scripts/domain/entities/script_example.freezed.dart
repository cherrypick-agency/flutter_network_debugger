// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'script_example.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ScriptExample {

 String get id; String get name; String get description; String get language; ExampleDifficulty get difficulty; ExampleCategory get category; String get triggerType; String get sourceCode;
/// Create a copy of ScriptExample
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ScriptExampleCopyWith<ScriptExample> get copyWith => _$ScriptExampleCopyWithImpl<ScriptExample>(this as ScriptExample, _$identity);

  /// Serializes this ScriptExample to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ScriptExample&&(identical(other.id, id) || other.id == id)&&(identical(other.name, name) || other.name == name)&&(identical(other.description, description) || other.description == description)&&(identical(other.language, language) || other.language == language)&&(identical(other.difficulty, difficulty) || other.difficulty == difficulty)&&(identical(other.category, category) || other.category == category)&&(identical(other.triggerType, triggerType) || other.triggerType == triggerType)&&(identical(other.sourceCode, sourceCode) || other.sourceCode == sourceCode));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,name,description,language,difficulty,category,triggerType,sourceCode);

@override
String toString() {
  return 'ScriptExample(id: $id, name: $name, description: $description, language: $language, difficulty: $difficulty, category: $category, triggerType: $triggerType, sourceCode: $sourceCode)';
}


}

/// @nodoc
abstract mixin class $ScriptExampleCopyWith<$Res>  {
  factory $ScriptExampleCopyWith(ScriptExample value, $Res Function(ScriptExample) _then) = _$ScriptExampleCopyWithImpl;
@useResult
$Res call({
 String id, String name, String description, String language, ExampleDifficulty difficulty, ExampleCategory category, String triggerType, String sourceCode
});




}
/// @nodoc
class _$ScriptExampleCopyWithImpl<$Res>
    implements $ScriptExampleCopyWith<$Res> {
  _$ScriptExampleCopyWithImpl(this._self, this._then);

  final ScriptExample _self;
  final $Res Function(ScriptExample) _then;

/// Create a copy of ScriptExample
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? name = null,Object? description = null,Object? language = null,Object? difficulty = null,Object? category = null,Object? triggerType = null,Object? sourceCode = null,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,description: null == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String,language: null == language ? _self.language : language // ignore: cast_nullable_to_non_nullable
as String,difficulty: null == difficulty ? _self.difficulty : difficulty // ignore: cast_nullable_to_non_nullable
as ExampleDifficulty,category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as ExampleCategory,triggerType: null == triggerType ? _self.triggerType : triggerType // ignore: cast_nullable_to_non_nullable
as String,sourceCode: null == sourceCode ? _self.sourceCode : sourceCode // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [ScriptExample].
extension ScriptExamplePatterns on ScriptExample {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ScriptExample value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ScriptExample() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ScriptExample value)  $default,){
final _that = this;
switch (_that) {
case _ScriptExample():
return $default(_that);}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ScriptExample value)?  $default,){
final _that = this;
switch (_that) {
case _ScriptExample() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  String name,  String description,  String language,  ExampleDifficulty difficulty,  ExampleCategory category,  String triggerType,  String sourceCode)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ScriptExample() when $default != null:
return $default(_that.id,_that.name,_that.description,_that.language,_that.difficulty,_that.category,_that.triggerType,_that.sourceCode);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  String name,  String description,  String language,  ExampleDifficulty difficulty,  ExampleCategory category,  String triggerType,  String sourceCode)  $default,) {final _that = this;
switch (_that) {
case _ScriptExample():
return $default(_that.id,_that.name,_that.description,_that.language,_that.difficulty,_that.category,_that.triggerType,_that.sourceCode);}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  String name,  String description,  String language,  ExampleDifficulty difficulty,  ExampleCategory category,  String triggerType,  String sourceCode)?  $default,) {final _that = this;
switch (_that) {
case _ScriptExample() when $default != null:
return $default(_that.id,_that.name,_that.description,_that.language,_that.difficulty,_that.category,_that.triggerType,_that.sourceCode);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ScriptExample extends ScriptExample {
  const _ScriptExample({required this.id, required this.name, required this.description, required this.language, required this.difficulty, required this.category, required this.triggerType, required this.sourceCode}): super._();
  factory _ScriptExample.fromJson(Map<String, dynamic> json) => _$ScriptExampleFromJson(json);

@override final  String id;
@override final  String name;
@override final  String description;
@override final  String language;
@override final  ExampleDifficulty difficulty;
@override final  ExampleCategory category;
@override final  String triggerType;
@override final  String sourceCode;

/// Create a copy of ScriptExample
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ScriptExampleCopyWith<_ScriptExample> get copyWith => __$ScriptExampleCopyWithImpl<_ScriptExample>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ScriptExampleToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _ScriptExample&&(identical(other.id, id) || other.id == id)&&(identical(other.name, name) || other.name == name)&&(identical(other.description, description) || other.description == description)&&(identical(other.language, language) || other.language == language)&&(identical(other.difficulty, difficulty) || other.difficulty == difficulty)&&(identical(other.category, category) || other.category == category)&&(identical(other.triggerType, triggerType) || other.triggerType == triggerType)&&(identical(other.sourceCode, sourceCode) || other.sourceCode == sourceCode));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,name,description,language,difficulty,category,triggerType,sourceCode);

@override
String toString() {
  return 'ScriptExample(id: $id, name: $name, description: $description, language: $language, difficulty: $difficulty, category: $category, triggerType: $triggerType, sourceCode: $sourceCode)';
}


}

/// @nodoc
abstract mixin class _$ScriptExampleCopyWith<$Res> implements $ScriptExampleCopyWith<$Res> {
  factory _$ScriptExampleCopyWith(_ScriptExample value, $Res Function(_ScriptExample) _then) = __$ScriptExampleCopyWithImpl;
@override @useResult
$Res call({
 String id, String name, String description, String language, ExampleDifficulty difficulty, ExampleCategory category, String triggerType, String sourceCode
});




}
/// @nodoc
class __$ScriptExampleCopyWithImpl<$Res>
    implements _$ScriptExampleCopyWith<$Res> {
  __$ScriptExampleCopyWithImpl(this._self, this._then);

  final _ScriptExample _self;
  final $Res Function(_ScriptExample) _then;

/// Create a copy of ScriptExample
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? name = null,Object? description = null,Object? language = null,Object? difficulty = null,Object? category = null,Object? triggerType = null,Object? sourceCode = null,}) {
  return _then(_ScriptExample(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,description: null == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String,language: null == language ? _self.language : language // ignore: cast_nullable_to_non_nullable
as String,difficulty: null == difficulty ? _self.difficulty : difficulty // ignore: cast_nullable_to_non_nullable
as ExampleDifficulty,category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as ExampleCategory,triggerType: null == triggerType ? _self.triggerType : triggerType // ignore: cast_nullable_to_non_nullable
as String,sourceCode: null == sourceCode ? _self.sourceCode : sourceCode // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
