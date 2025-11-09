// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'compiler_info.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$CompilerInfo {

 String get language; String get version; CompilerStatus get status; String get installedPath; int get size; int get downloadSize; String get error;
/// Create a copy of CompilerInfo
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$CompilerInfoCopyWith<CompilerInfo> get copyWith => _$CompilerInfoCopyWithImpl<CompilerInfo>(this as CompilerInfo, _$identity);

  /// Serializes this CompilerInfo to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is CompilerInfo&&(identical(other.language, language) || other.language == language)&&(identical(other.version, version) || other.version == version)&&(identical(other.status, status) || other.status == status)&&(identical(other.installedPath, installedPath) || other.installedPath == installedPath)&&(identical(other.size, size) || other.size == size)&&(identical(other.downloadSize, downloadSize) || other.downloadSize == downloadSize)&&(identical(other.error, error) || other.error == error));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,language,version,status,installedPath,size,downloadSize,error);

@override
String toString() {
  return 'CompilerInfo(language: $language, version: $version, status: $status, installedPath: $installedPath, size: $size, downloadSize: $downloadSize, error: $error)';
}


}

/// @nodoc
abstract mixin class $CompilerInfoCopyWith<$Res>  {
  factory $CompilerInfoCopyWith(CompilerInfo value, $Res Function(CompilerInfo) _then) = _$CompilerInfoCopyWithImpl;
@useResult
$Res call({
 String language, String version, CompilerStatus status, String installedPath, int size, int downloadSize, String error
});


$CompilerStatusCopyWith<$Res> get status;

}
/// @nodoc
class _$CompilerInfoCopyWithImpl<$Res>
    implements $CompilerInfoCopyWith<$Res> {
  _$CompilerInfoCopyWithImpl(this._self, this._then);

  final CompilerInfo _self;
  final $Res Function(CompilerInfo) _then;

/// Create a copy of CompilerInfo
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? language = null,Object? version = null,Object? status = null,Object? installedPath = null,Object? size = null,Object? downloadSize = null,Object? error = null,}) {
  return _then(_self.copyWith(
language: null == language ? _self.language : language // ignore: cast_nullable_to_non_nullable
as String,version: null == version ? _self.version : version // ignore: cast_nullable_to_non_nullable
as String,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as CompilerStatus,installedPath: null == installedPath ? _self.installedPath : installedPath // ignore: cast_nullable_to_non_nullable
as String,size: null == size ? _self.size : size // ignore: cast_nullable_to_non_nullable
as int,downloadSize: null == downloadSize ? _self.downloadSize : downloadSize // ignore: cast_nullable_to_non_nullable
as int,error: null == error ? _self.error : error // ignore: cast_nullable_to_non_nullable
as String,
  ));
}
/// Create a copy of CompilerInfo
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$CompilerStatusCopyWith<$Res> get status {
  
  return $CompilerStatusCopyWith<$Res>(_self.status, (value) {
    return _then(_self.copyWith(status: value));
  });
}
}


/// Adds pattern-matching-related methods to [CompilerInfo].
extension CompilerInfoPatterns on CompilerInfo {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _CompilerInfo value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _CompilerInfo() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _CompilerInfo value)  $default,){
final _that = this;
switch (_that) {
case _CompilerInfo():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _CompilerInfo value)?  $default,){
final _that = this;
switch (_that) {
case _CompilerInfo() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String language,  String version,  CompilerStatus status,  String installedPath,  int size,  int downloadSize,  String error)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _CompilerInfo() when $default != null:
return $default(_that.language,_that.version,_that.status,_that.installedPath,_that.size,_that.downloadSize,_that.error);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String language,  String version,  CompilerStatus status,  String installedPath,  int size,  int downloadSize,  String error)  $default,) {final _that = this;
switch (_that) {
case _CompilerInfo():
return $default(_that.language,_that.version,_that.status,_that.installedPath,_that.size,_that.downloadSize,_that.error);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String language,  String version,  CompilerStatus status,  String installedPath,  int size,  int downloadSize,  String error)?  $default,) {final _that = this;
switch (_that) {
case _CompilerInfo() when $default != null:
return $default(_that.language,_that.version,_that.status,_that.installedPath,_that.size,_that.downloadSize,_that.error);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _CompilerInfo extends CompilerInfo {
  const _CompilerInfo({required this.language, required this.version, required this.status, required this.installedPath, required this.size, required this.downloadSize, this.error = ''}): super._();
  factory _CompilerInfo.fromJson(Map<String, dynamic> json) => _$CompilerInfoFromJson(json);

@override final  String language;
@override final  String version;
@override final  CompilerStatus status;
@override final  String installedPath;
@override final  int size;
@override final  int downloadSize;
@override@JsonKey() final  String error;

/// Create a copy of CompilerInfo
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$CompilerInfoCopyWith<_CompilerInfo> get copyWith => __$CompilerInfoCopyWithImpl<_CompilerInfo>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$CompilerInfoToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _CompilerInfo&&(identical(other.language, language) || other.language == language)&&(identical(other.version, version) || other.version == version)&&(identical(other.status, status) || other.status == status)&&(identical(other.installedPath, installedPath) || other.installedPath == installedPath)&&(identical(other.size, size) || other.size == size)&&(identical(other.downloadSize, downloadSize) || other.downloadSize == downloadSize)&&(identical(other.error, error) || other.error == error));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,language,version,status,installedPath,size,downloadSize,error);

@override
String toString() {
  return 'CompilerInfo(language: $language, version: $version, status: $status, installedPath: $installedPath, size: $size, downloadSize: $downloadSize, error: $error)';
}


}

/// @nodoc
abstract mixin class _$CompilerInfoCopyWith<$Res> implements $CompilerInfoCopyWith<$Res> {
  factory _$CompilerInfoCopyWith(_CompilerInfo value, $Res Function(_CompilerInfo) _then) = __$CompilerInfoCopyWithImpl;
@override @useResult
$Res call({
 String language, String version, CompilerStatus status, String installedPath, int size, int downloadSize, String error
});


@override $CompilerStatusCopyWith<$Res> get status;

}
/// @nodoc
class __$CompilerInfoCopyWithImpl<$Res>
    implements _$CompilerInfoCopyWith<$Res> {
  __$CompilerInfoCopyWithImpl(this._self, this._then);

  final _CompilerInfo _self;
  final $Res Function(_CompilerInfo) _then;

/// Create a copy of CompilerInfo
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? language = null,Object? version = null,Object? status = null,Object? installedPath = null,Object? size = null,Object? downloadSize = null,Object? error = null,}) {
  return _then(_CompilerInfo(
language: null == language ? _self.language : language // ignore: cast_nullable_to_non_nullable
as String,version: null == version ? _self.version : version // ignore: cast_nullable_to_non_nullable
as String,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as CompilerStatus,installedPath: null == installedPath ? _self.installedPath : installedPath // ignore: cast_nullable_to_non_nullable
as String,size: null == size ? _self.size : size // ignore: cast_nullable_to_non_nullable
as int,downloadSize: null == downloadSize ? _self.downloadSize : downloadSize // ignore: cast_nullable_to_non_nullable
as int,error: null == error ? _self.error : error // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

/// Create a copy of CompilerInfo
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$CompilerStatusCopyWith<$Res> get status {
  
  return $CompilerStatusCopyWith<$Res>(_self.status, (value) {
    return _then(_self.copyWith(status: value));
  });
}
}

// dart format on
