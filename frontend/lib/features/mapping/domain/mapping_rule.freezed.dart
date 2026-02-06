// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'mapping_rule.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;
/// @nodoc
mixin _$MappingRule {

 String get id; bool get enabled; int get priority; String get kind; bool get stopProcessing; List<String> get methods; String get hostPattern; String get pathPattern; String get patternType; String? get filePath; String? get blobPath; int? get statusOverride; String? get contentTypeOverride; String? get targetURLTemplate; bool get preserveHost; DateTime? get createdAt; DateTime? get updatedAt;
/// Create a copy of MappingRule
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$MappingRuleCopyWith<MappingRule> get copyWith => _$MappingRuleCopyWithImpl<MappingRule>(this as MappingRule, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is MappingRule&&(identical(other.id, id) || other.id == id)&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.priority, priority) || other.priority == priority)&&(identical(other.kind, kind) || other.kind == kind)&&(identical(other.stopProcessing, stopProcessing) || other.stopProcessing == stopProcessing)&&const DeepCollectionEquality().equals(other.methods, methods)&&(identical(other.hostPattern, hostPattern) || other.hostPattern == hostPattern)&&(identical(other.pathPattern, pathPattern) || other.pathPattern == pathPattern)&&(identical(other.patternType, patternType) || other.patternType == patternType)&&(identical(other.filePath, filePath) || other.filePath == filePath)&&(identical(other.blobPath, blobPath) || other.blobPath == blobPath)&&(identical(other.statusOverride, statusOverride) || other.statusOverride == statusOverride)&&(identical(other.contentTypeOverride, contentTypeOverride) || other.contentTypeOverride == contentTypeOverride)&&(identical(other.targetURLTemplate, targetURLTemplate) || other.targetURLTemplate == targetURLTemplate)&&(identical(other.preserveHost, preserveHost) || other.preserveHost == preserveHost)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}


@override
int get hashCode => Object.hash(runtimeType,id,enabled,priority,kind,stopProcessing,const DeepCollectionEquality().hash(methods),hostPattern,pathPattern,patternType,filePath,blobPath,statusOverride,contentTypeOverride,targetURLTemplate,preserveHost,createdAt,updatedAt);

@override
String toString() {
  return 'MappingRule(id: $id, enabled: $enabled, priority: $priority, kind: $kind, stopProcessing: $stopProcessing, methods: $methods, hostPattern: $hostPattern, pathPattern: $pathPattern, patternType: $patternType, filePath: $filePath, blobPath: $blobPath, statusOverride: $statusOverride, contentTypeOverride: $contentTypeOverride, targetURLTemplate: $targetURLTemplate, preserveHost: $preserveHost, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class $MappingRuleCopyWith<$Res>  {
  factory $MappingRuleCopyWith(MappingRule value, $Res Function(MappingRule) _then) = _$MappingRuleCopyWithImpl;
@useResult
$Res call({
 String id, bool enabled, int priority, String kind, bool stopProcessing, List<String> methods, String hostPattern, String pathPattern, String patternType, String? filePath, String? blobPath, int? statusOverride, String? contentTypeOverride, String? targetURLTemplate, bool preserveHost, DateTime? createdAt, DateTime? updatedAt
});




}
/// @nodoc
class _$MappingRuleCopyWithImpl<$Res>
    implements $MappingRuleCopyWith<$Res> {
  _$MappingRuleCopyWithImpl(this._self, this._then);

  final MappingRule _self;
  final $Res Function(MappingRule) _then;

/// Create a copy of MappingRule
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? enabled = null,Object? priority = null,Object? kind = null,Object? stopProcessing = null,Object? methods = null,Object? hostPattern = null,Object? pathPattern = null,Object? patternType = null,Object? filePath = freezed,Object? blobPath = freezed,Object? statusOverride = freezed,Object? contentTypeOverride = freezed,Object? targetURLTemplate = freezed,Object? preserveHost = null,Object? createdAt = freezed,Object? updatedAt = freezed,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,priority: null == priority ? _self.priority : priority // ignore: cast_nullable_to_non_nullable
as int,kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as String,stopProcessing: null == stopProcessing ? _self.stopProcessing : stopProcessing // ignore: cast_nullable_to_non_nullable
as bool,methods: null == methods ? _self.methods : methods // ignore: cast_nullable_to_non_nullable
as List<String>,hostPattern: null == hostPattern ? _self.hostPattern : hostPattern // ignore: cast_nullable_to_non_nullable
as String,pathPattern: null == pathPattern ? _self.pathPattern : pathPattern // ignore: cast_nullable_to_non_nullable
as String,patternType: null == patternType ? _self.patternType : patternType // ignore: cast_nullable_to_non_nullable
as String,filePath: freezed == filePath ? _self.filePath : filePath // ignore: cast_nullable_to_non_nullable
as String?,blobPath: freezed == blobPath ? _self.blobPath : blobPath // ignore: cast_nullable_to_non_nullable
as String?,statusOverride: freezed == statusOverride ? _self.statusOverride : statusOverride // ignore: cast_nullable_to_non_nullable
as int?,contentTypeOverride: freezed == contentTypeOverride ? _self.contentTypeOverride : contentTypeOverride // ignore: cast_nullable_to_non_nullable
as String?,targetURLTemplate: freezed == targetURLTemplate ? _self.targetURLTemplate : targetURLTemplate // ignore: cast_nullable_to_non_nullable
as String?,preserveHost: null == preserveHost ? _self.preserveHost : preserveHost // ignore: cast_nullable_to_non_nullable
as bool,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

}


/// Adds pattern-matching-related methods to [MappingRule].
extension MappingRulePatterns on MappingRule {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _MappingRule value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _MappingRule() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _MappingRule value)  $default,){
final _that = this;
switch (_that) {
case _MappingRule():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _MappingRule value)?  $default,){
final _that = this;
switch (_that) {
case _MappingRule() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  bool enabled,  int priority,  String kind,  bool stopProcessing,  List<String> methods,  String hostPattern,  String pathPattern,  String patternType,  String? filePath,  String? blobPath,  int? statusOverride,  String? contentTypeOverride,  String? targetURLTemplate,  bool preserveHost,  DateTime? createdAt,  DateTime? updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _MappingRule() when $default != null:
return $default(_that.id,_that.enabled,_that.priority,_that.kind,_that.stopProcessing,_that.methods,_that.hostPattern,_that.pathPattern,_that.patternType,_that.filePath,_that.blobPath,_that.statusOverride,_that.contentTypeOverride,_that.targetURLTemplate,_that.preserveHost,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  bool enabled,  int priority,  String kind,  bool stopProcessing,  List<String> methods,  String hostPattern,  String pathPattern,  String patternType,  String? filePath,  String? blobPath,  int? statusOverride,  String? contentTypeOverride,  String? targetURLTemplate,  bool preserveHost,  DateTime? createdAt,  DateTime? updatedAt)  $default,) {final _that = this;
switch (_that) {
case _MappingRule():
return $default(_that.id,_that.enabled,_that.priority,_that.kind,_that.stopProcessing,_that.methods,_that.hostPattern,_that.pathPattern,_that.patternType,_that.filePath,_that.blobPath,_that.statusOverride,_that.contentTypeOverride,_that.targetURLTemplate,_that.preserveHost,_that.createdAt,_that.updatedAt);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  bool enabled,  int priority,  String kind,  bool stopProcessing,  List<String> methods,  String hostPattern,  String pathPattern,  String patternType,  String? filePath,  String? blobPath,  int? statusOverride,  String? contentTypeOverride,  String? targetURLTemplate,  bool preserveHost,  DateTime? createdAt,  DateTime? updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _MappingRule() when $default != null:
return $default(_that.id,_that.enabled,_that.priority,_that.kind,_that.stopProcessing,_that.methods,_that.hostPattern,_that.pathPattern,_that.patternType,_that.filePath,_that.blobPath,_that.statusOverride,_that.contentTypeOverride,_that.targetURLTemplate,_that.preserveHost,_that.createdAt,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc


class _MappingRule extends MappingRule {
  const _MappingRule({required this.id, this.enabled = true, this.priority = 100, this.kind = 'local', this.stopProcessing = true, final  List<String> methods = const [], this.hostPattern = '', this.pathPattern = '', this.patternType = 'glob', this.filePath, this.blobPath, this.statusOverride, this.contentTypeOverride, this.targetURLTemplate, this.preserveHost = false, this.createdAt, this.updatedAt}): _methods = methods,super._();
  

@override final  String id;
@override@JsonKey() final  bool enabled;
@override@JsonKey() final  int priority;
@override@JsonKey() final  String kind;
@override@JsonKey() final  bool stopProcessing;
 final  List<String> _methods;
@override@JsonKey() List<String> get methods {
  if (_methods is EqualUnmodifiableListView) return _methods;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_methods);
}

@override@JsonKey() final  String hostPattern;
@override@JsonKey() final  String pathPattern;
@override@JsonKey() final  String patternType;
@override final  String? filePath;
@override final  String? blobPath;
@override final  int? statusOverride;
@override final  String? contentTypeOverride;
@override final  String? targetURLTemplate;
@override@JsonKey() final  bool preserveHost;
@override final  DateTime? createdAt;
@override final  DateTime? updatedAt;

/// Create a copy of MappingRule
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$MappingRuleCopyWith<_MappingRule> get copyWith => __$MappingRuleCopyWithImpl<_MappingRule>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _MappingRule&&(identical(other.id, id) || other.id == id)&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.priority, priority) || other.priority == priority)&&(identical(other.kind, kind) || other.kind == kind)&&(identical(other.stopProcessing, stopProcessing) || other.stopProcessing == stopProcessing)&&const DeepCollectionEquality().equals(other._methods, _methods)&&(identical(other.hostPattern, hostPattern) || other.hostPattern == hostPattern)&&(identical(other.pathPattern, pathPattern) || other.pathPattern == pathPattern)&&(identical(other.patternType, patternType) || other.patternType == patternType)&&(identical(other.filePath, filePath) || other.filePath == filePath)&&(identical(other.blobPath, blobPath) || other.blobPath == blobPath)&&(identical(other.statusOverride, statusOverride) || other.statusOverride == statusOverride)&&(identical(other.contentTypeOverride, contentTypeOverride) || other.contentTypeOverride == contentTypeOverride)&&(identical(other.targetURLTemplate, targetURLTemplate) || other.targetURLTemplate == targetURLTemplate)&&(identical(other.preserveHost, preserveHost) || other.preserveHost == preserveHost)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}


@override
int get hashCode => Object.hash(runtimeType,id,enabled,priority,kind,stopProcessing,const DeepCollectionEquality().hash(_methods),hostPattern,pathPattern,patternType,filePath,blobPath,statusOverride,contentTypeOverride,targetURLTemplate,preserveHost,createdAt,updatedAt);

@override
String toString() {
  return 'MappingRule(id: $id, enabled: $enabled, priority: $priority, kind: $kind, stopProcessing: $stopProcessing, methods: $methods, hostPattern: $hostPattern, pathPattern: $pathPattern, patternType: $patternType, filePath: $filePath, blobPath: $blobPath, statusOverride: $statusOverride, contentTypeOverride: $contentTypeOverride, targetURLTemplate: $targetURLTemplate, preserveHost: $preserveHost, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$MappingRuleCopyWith<$Res> implements $MappingRuleCopyWith<$Res> {
  factory _$MappingRuleCopyWith(_MappingRule value, $Res Function(_MappingRule) _then) = __$MappingRuleCopyWithImpl;
@override @useResult
$Res call({
 String id, bool enabled, int priority, String kind, bool stopProcessing, List<String> methods, String hostPattern, String pathPattern, String patternType, String? filePath, String? blobPath, int? statusOverride, String? contentTypeOverride, String? targetURLTemplate, bool preserveHost, DateTime? createdAt, DateTime? updatedAt
});




}
/// @nodoc
class __$MappingRuleCopyWithImpl<$Res>
    implements _$MappingRuleCopyWith<$Res> {
  __$MappingRuleCopyWithImpl(this._self, this._then);

  final _MappingRule _self;
  final $Res Function(_MappingRule) _then;

/// Create a copy of MappingRule
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? enabled = null,Object? priority = null,Object? kind = null,Object? stopProcessing = null,Object? methods = null,Object? hostPattern = null,Object? pathPattern = null,Object? patternType = null,Object? filePath = freezed,Object? blobPath = freezed,Object? statusOverride = freezed,Object? contentTypeOverride = freezed,Object? targetURLTemplate = freezed,Object? preserveHost = null,Object? createdAt = freezed,Object? updatedAt = freezed,}) {
  return _then(_MappingRule(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,priority: null == priority ? _self.priority : priority // ignore: cast_nullable_to_non_nullable
as int,kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as String,stopProcessing: null == stopProcessing ? _self.stopProcessing : stopProcessing // ignore: cast_nullable_to_non_nullable
as bool,methods: null == methods ? _self._methods : methods // ignore: cast_nullable_to_non_nullable
as List<String>,hostPattern: null == hostPattern ? _self.hostPattern : hostPattern // ignore: cast_nullable_to_non_nullable
as String,pathPattern: null == pathPattern ? _self.pathPattern : pathPattern // ignore: cast_nullable_to_non_nullable
as String,patternType: null == patternType ? _self.patternType : patternType // ignore: cast_nullable_to_non_nullable
as String,filePath: freezed == filePath ? _self.filePath : filePath // ignore: cast_nullable_to_non_nullable
as String?,blobPath: freezed == blobPath ? _self.blobPath : blobPath // ignore: cast_nullable_to_non_nullable
as String?,statusOverride: freezed == statusOverride ? _self.statusOverride : statusOverride // ignore: cast_nullable_to_non_nullable
as int?,contentTypeOverride: freezed == contentTypeOverride ? _self.contentTypeOverride : contentTypeOverride // ignore: cast_nullable_to_non_nullable
as String?,targetURLTemplate: freezed == targetURLTemplate ? _self.targetURLTemplate : targetURLTemplate // ignore: cast_nullable_to_non_nullable
as String?,preserveHost: null == preserveHost ? _self.preserveHost : preserveHost // ignore: cast_nullable_to_non_nullable
as bool,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}


}

// dart format on
