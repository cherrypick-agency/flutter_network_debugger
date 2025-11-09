// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'script.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Script {

 String get id; String get name; String? get description; ScriptRuntime get runtime; String get code; String get language; TriggerType get triggerType; int get priority; bool get enabled; MatchRules? get matchRules; ScriptConfig? get config; DateTime? get createdAt; DateTime? get updatedAt;// Compilation fields
 String? get sourceCode; Map<String, String>? get dependencies; String? get compilationStatus; String? get compilationError; DateTime? get lastCompiledAt;
/// Create a copy of Script
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ScriptCopyWith<Script> get copyWith => _$ScriptCopyWithImpl<Script>(this as Script, _$identity);

  /// Serializes this Script to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Script&&(identical(other.id, id) || other.id == id)&&(identical(other.name, name) || other.name == name)&&(identical(other.description, description) || other.description == description)&&(identical(other.runtime, runtime) || other.runtime == runtime)&&(identical(other.code, code) || other.code == code)&&(identical(other.language, language) || other.language == language)&&(identical(other.triggerType, triggerType) || other.triggerType == triggerType)&&(identical(other.priority, priority) || other.priority == priority)&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.matchRules, matchRules) || other.matchRules == matchRules)&&(identical(other.config, config) || other.config == config)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt)&&(identical(other.sourceCode, sourceCode) || other.sourceCode == sourceCode)&&const DeepCollectionEquality().equals(other.dependencies, dependencies)&&(identical(other.compilationStatus, compilationStatus) || other.compilationStatus == compilationStatus)&&(identical(other.compilationError, compilationError) || other.compilationError == compilationError)&&(identical(other.lastCompiledAt, lastCompiledAt) || other.lastCompiledAt == lastCompiledAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,name,description,runtime,code,language,triggerType,priority,enabled,matchRules,config,createdAt,updatedAt,sourceCode,const DeepCollectionEquality().hash(dependencies),compilationStatus,compilationError,lastCompiledAt);

@override
String toString() {
  return 'Script(id: $id, name: $name, description: $description, runtime: $runtime, code: $code, language: $language, triggerType: $triggerType, priority: $priority, enabled: $enabled, matchRules: $matchRules, config: $config, createdAt: $createdAt, updatedAt: $updatedAt, sourceCode: $sourceCode, dependencies: $dependencies, compilationStatus: $compilationStatus, compilationError: $compilationError, lastCompiledAt: $lastCompiledAt)';
}


}

/// @nodoc
abstract mixin class $ScriptCopyWith<$Res>  {
  factory $ScriptCopyWith(Script value, $Res Function(Script) _then) = _$ScriptCopyWithImpl;
@useResult
$Res call({
 String id, String name, String? description, ScriptRuntime runtime, String code, String language, TriggerType triggerType, int priority, bool enabled, MatchRules? matchRules, ScriptConfig? config, DateTime? createdAt, DateTime? updatedAt, String? sourceCode, Map<String, String>? dependencies, String? compilationStatus, String? compilationError, DateTime? lastCompiledAt
});


$MatchRulesCopyWith<$Res>? get matchRules;$ScriptConfigCopyWith<$Res>? get config;

}
/// @nodoc
class _$ScriptCopyWithImpl<$Res>
    implements $ScriptCopyWith<$Res> {
  _$ScriptCopyWithImpl(this._self, this._then);

  final Script _self;
  final $Res Function(Script) _then;

/// Create a copy of Script
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? name = null,Object? description = freezed,Object? runtime = null,Object? code = null,Object? language = null,Object? triggerType = null,Object? priority = null,Object? enabled = null,Object? matchRules = freezed,Object? config = freezed,Object? createdAt = freezed,Object? updatedAt = freezed,Object? sourceCode = freezed,Object? dependencies = freezed,Object? compilationStatus = freezed,Object? compilationError = freezed,Object? lastCompiledAt = freezed,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String?,runtime: null == runtime ? _self.runtime : runtime // ignore: cast_nullable_to_non_nullable
as ScriptRuntime,code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as String,language: null == language ? _self.language : language // ignore: cast_nullable_to_non_nullable
as String,triggerType: null == triggerType ? _self.triggerType : triggerType // ignore: cast_nullable_to_non_nullable
as TriggerType,priority: null == priority ? _self.priority : priority // ignore: cast_nullable_to_non_nullable
as int,enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,matchRules: freezed == matchRules ? _self.matchRules : matchRules // ignore: cast_nullable_to_non_nullable
as MatchRules?,config: freezed == config ? _self.config : config // ignore: cast_nullable_to_non_nullable
as ScriptConfig?,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,sourceCode: freezed == sourceCode ? _self.sourceCode : sourceCode // ignore: cast_nullable_to_non_nullable
as String?,dependencies: freezed == dependencies ? _self.dependencies : dependencies // ignore: cast_nullable_to_non_nullable
as Map<String, String>?,compilationStatus: freezed == compilationStatus ? _self.compilationStatus : compilationStatus // ignore: cast_nullable_to_non_nullable
as String?,compilationError: freezed == compilationError ? _self.compilationError : compilationError // ignore: cast_nullable_to_non_nullable
as String?,lastCompiledAt: freezed == lastCompiledAt ? _self.lastCompiledAt : lastCompiledAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}
/// Create a copy of Script
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$MatchRulesCopyWith<$Res>? get matchRules {
    if (_self.matchRules == null) {
    return null;
  }

  return $MatchRulesCopyWith<$Res>(_self.matchRules!, (value) {
    return _then(_self.copyWith(matchRules: value));
  });
}/// Create a copy of Script
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ScriptConfigCopyWith<$Res>? get config {
    if (_self.config == null) {
    return null;
  }

  return $ScriptConfigCopyWith<$Res>(_self.config!, (value) {
    return _then(_self.copyWith(config: value));
  });
}
}


/// Adds pattern-matching-related methods to [Script].
extension ScriptPatterns on Script {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Script value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Script() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Script value)  $default,){
final _that = this;
switch (_that) {
case _Script():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Script value)?  $default,){
final _that = this;
switch (_that) {
case _Script() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  String name,  String? description,  ScriptRuntime runtime,  String code,  String language,  TriggerType triggerType,  int priority,  bool enabled,  MatchRules? matchRules,  ScriptConfig? config,  DateTime? createdAt,  DateTime? updatedAt,  String? sourceCode,  Map<String, String>? dependencies,  String? compilationStatus,  String? compilationError,  DateTime? lastCompiledAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Script() when $default != null:
return $default(_that.id,_that.name,_that.description,_that.runtime,_that.code,_that.language,_that.triggerType,_that.priority,_that.enabled,_that.matchRules,_that.config,_that.createdAt,_that.updatedAt,_that.sourceCode,_that.dependencies,_that.compilationStatus,_that.compilationError,_that.lastCompiledAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  String name,  String? description,  ScriptRuntime runtime,  String code,  String language,  TriggerType triggerType,  int priority,  bool enabled,  MatchRules? matchRules,  ScriptConfig? config,  DateTime? createdAt,  DateTime? updatedAt,  String? sourceCode,  Map<String, String>? dependencies,  String? compilationStatus,  String? compilationError,  DateTime? lastCompiledAt)  $default,) {final _that = this;
switch (_that) {
case _Script():
return $default(_that.id,_that.name,_that.description,_that.runtime,_that.code,_that.language,_that.triggerType,_that.priority,_that.enabled,_that.matchRules,_that.config,_that.createdAt,_that.updatedAt,_that.sourceCode,_that.dependencies,_that.compilationStatus,_that.compilationError,_that.lastCompiledAt);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  String name,  String? description,  ScriptRuntime runtime,  String code,  String language,  TriggerType triggerType,  int priority,  bool enabled,  MatchRules? matchRules,  ScriptConfig? config,  DateTime? createdAt,  DateTime? updatedAt,  String? sourceCode,  Map<String, String>? dependencies,  String? compilationStatus,  String? compilationError,  DateTime? lastCompiledAt)?  $default,) {final _that = this;
switch (_that) {
case _Script() when $default != null:
return $default(_that.id,_that.name,_that.description,_that.runtime,_that.code,_that.language,_that.triggerType,_that.priority,_that.enabled,_that.matchRules,_that.config,_that.createdAt,_that.updatedAt,_that.sourceCode,_that.dependencies,_that.compilationStatus,_that.compilationError,_that.lastCompiledAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Script extends Script {
  const _Script({required this.id, required this.name, this.description, required this.runtime, required this.code, required this.language, required this.triggerType, this.priority = 10, this.enabled = true, this.matchRules, this.config, this.createdAt, this.updatedAt, this.sourceCode, final  Map<String, String>? dependencies, this.compilationStatus, this.compilationError, this.lastCompiledAt}): _dependencies = dependencies,super._();
  factory _Script.fromJson(Map<String, dynamic> json) => _$ScriptFromJson(json);

@override final  String id;
@override final  String name;
@override final  String? description;
@override final  ScriptRuntime runtime;
@override final  String code;
@override final  String language;
@override final  TriggerType triggerType;
@override@JsonKey() final  int priority;
@override@JsonKey() final  bool enabled;
@override final  MatchRules? matchRules;
@override final  ScriptConfig? config;
@override final  DateTime? createdAt;
@override final  DateTime? updatedAt;
// Compilation fields
@override final  String? sourceCode;
 final  Map<String, String>? _dependencies;
@override Map<String, String>? get dependencies {
  final value = _dependencies;
  if (value == null) return null;
  if (_dependencies is EqualUnmodifiableMapView) return _dependencies;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(value);
}

@override final  String? compilationStatus;
@override final  String? compilationError;
@override final  DateTime? lastCompiledAt;

/// Create a copy of Script
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ScriptCopyWith<_Script> get copyWith => __$ScriptCopyWithImpl<_Script>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ScriptToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _Script&&(identical(other.id, id) || other.id == id)&&(identical(other.name, name) || other.name == name)&&(identical(other.description, description) || other.description == description)&&(identical(other.runtime, runtime) || other.runtime == runtime)&&(identical(other.code, code) || other.code == code)&&(identical(other.language, language) || other.language == language)&&(identical(other.triggerType, triggerType) || other.triggerType == triggerType)&&(identical(other.priority, priority) || other.priority == priority)&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.matchRules, matchRules) || other.matchRules == matchRules)&&(identical(other.config, config) || other.config == config)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt)&&(identical(other.sourceCode, sourceCode) || other.sourceCode == sourceCode)&&const DeepCollectionEquality().equals(other._dependencies, _dependencies)&&(identical(other.compilationStatus, compilationStatus) || other.compilationStatus == compilationStatus)&&(identical(other.compilationError, compilationError) || other.compilationError == compilationError)&&(identical(other.lastCompiledAt, lastCompiledAt) || other.lastCompiledAt == lastCompiledAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,name,description,runtime,code,language,triggerType,priority,enabled,matchRules,config,createdAt,updatedAt,sourceCode,const DeepCollectionEquality().hash(_dependencies),compilationStatus,compilationError,lastCompiledAt);

@override
String toString() {
  return 'Script(id: $id, name: $name, description: $description, runtime: $runtime, code: $code, language: $language, triggerType: $triggerType, priority: $priority, enabled: $enabled, matchRules: $matchRules, config: $config, createdAt: $createdAt, updatedAt: $updatedAt, sourceCode: $sourceCode, dependencies: $dependencies, compilationStatus: $compilationStatus, compilationError: $compilationError, lastCompiledAt: $lastCompiledAt)';
}


}

/// @nodoc
abstract mixin class _$ScriptCopyWith<$Res> implements $ScriptCopyWith<$Res> {
  factory _$ScriptCopyWith(_Script value, $Res Function(_Script) _then) = __$ScriptCopyWithImpl;
@override @useResult
$Res call({
 String id, String name, String? description, ScriptRuntime runtime, String code, String language, TriggerType triggerType, int priority, bool enabled, MatchRules? matchRules, ScriptConfig? config, DateTime? createdAt, DateTime? updatedAt, String? sourceCode, Map<String, String>? dependencies, String? compilationStatus, String? compilationError, DateTime? lastCompiledAt
});


@override $MatchRulesCopyWith<$Res>? get matchRules;@override $ScriptConfigCopyWith<$Res>? get config;

}
/// @nodoc
class __$ScriptCopyWithImpl<$Res>
    implements _$ScriptCopyWith<$Res> {
  __$ScriptCopyWithImpl(this._self, this._then);

  final _Script _self;
  final $Res Function(_Script) _then;

/// Create a copy of Script
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? name = null,Object? description = freezed,Object? runtime = null,Object? code = null,Object? language = null,Object? triggerType = null,Object? priority = null,Object? enabled = null,Object? matchRules = freezed,Object? config = freezed,Object? createdAt = freezed,Object? updatedAt = freezed,Object? sourceCode = freezed,Object? dependencies = freezed,Object? compilationStatus = freezed,Object? compilationError = freezed,Object? lastCompiledAt = freezed,}) {
  return _then(_Script(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String?,runtime: null == runtime ? _self.runtime : runtime // ignore: cast_nullable_to_non_nullable
as ScriptRuntime,code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as String,language: null == language ? _self.language : language // ignore: cast_nullable_to_non_nullable
as String,triggerType: null == triggerType ? _self.triggerType : triggerType // ignore: cast_nullable_to_non_nullable
as TriggerType,priority: null == priority ? _self.priority : priority // ignore: cast_nullable_to_non_nullable
as int,enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,matchRules: freezed == matchRules ? _self.matchRules : matchRules // ignore: cast_nullable_to_non_nullable
as MatchRules?,config: freezed == config ? _self.config : config // ignore: cast_nullable_to_non_nullable
as ScriptConfig?,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,sourceCode: freezed == sourceCode ? _self.sourceCode : sourceCode // ignore: cast_nullable_to_non_nullable
as String?,dependencies: freezed == dependencies ? _self._dependencies : dependencies // ignore: cast_nullable_to_non_nullable
as Map<String, String>?,compilationStatus: freezed == compilationStatus ? _self.compilationStatus : compilationStatus // ignore: cast_nullable_to_non_nullable
as String?,compilationError: freezed == compilationError ? _self.compilationError : compilationError // ignore: cast_nullable_to_non_nullable
as String?,lastCompiledAt: freezed == lastCompiledAt ? _self.lastCompiledAt : lastCompiledAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

/// Create a copy of Script
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$MatchRulesCopyWith<$Res>? get matchRules {
    if (_self.matchRules == null) {
    return null;
  }

  return $MatchRulesCopyWith<$Res>(_self.matchRules!, (value) {
    return _then(_self.copyWith(matchRules: value));
  });
}/// Create a copy of Script
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ScriptConfigCopyWith<$Res>? get config {
    if (_self.config == null) {
    return null;
  }

  return $ScriptConfigCopyWith<$Res>(_self.config!, (value) {
    return _then(_self.copyWith(config: value));
  });
}
}

// dart format on
