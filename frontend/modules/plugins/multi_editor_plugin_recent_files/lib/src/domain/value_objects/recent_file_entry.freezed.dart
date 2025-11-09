// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'recent_file_entry.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RecentFileEntry {

 String get fileId; String get fileName; String get filePath; DateTime get lastOpened;
/// Create a copy of RecentFileEntry
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RecentFileEntryCopyWith<RecentFileEntry> get copyWith => _$RecentFileEntryCopyWithImpl<RecentFileEntry>(this as RecentFileEntry, _$identity);

  /// Serializes this RecentFileEntry to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RecentFileEntry&&(identical(other.fileId, fileId) || other.fileId == fileId)&&(identical(other.fileName, fileName) || other.fileName == fileName)&&(identical(other.filePath, filePath) || other.filePath == filePath)&&(identical(other.lastOpened, lastOpened) || other.lastOpened == lastOpened));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,fileId,fileName,filePath,lastOpened);

@override
String toString() {
  return 'RecentFileEntry(fileId: $fileId, fileName: $fileName, filePath: $filePath, lastOpened: $lastOpened)';
}


}

/// @nodoc
abstract mixin class $RecentFileEntryCopyWith<$Res>  {
  factory $RecentFileEntryCopyWith(RecentFileEntry value, $Res Function(RecentFileEntry) _then) = _$RecentFileEntryCopyWithImpl;
@useResult
$Res call({
 String fileId, String fileName, String filePath, DateTime lastOpened
});




}
/// @nodoc
class _$RecentFileEntryCopyWithImpl<$Res>
    implements $RecentFileEntryCopyWith<$Res> {
  _$RecentFileEntryCopyWithImpl(this._self, this._then);

  final RecentFileEntry _self;
  final $Res Function(RecentFileEntry) _then;

/// Create a copy of RecentFileEntry
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fileId = null,Object? fileName = null,Object? filePath = null,Object? lastOpened = null,}) {
  return _then(_self.copyWith(
fileId: null == fileId ? _self.fileId : fileId // ignore: cast_nullable_to_non_nullable
as String,fileName: null == fileName ? _self.fileName : fileName // ignore: cast_nullable_to_non_nullable
as String,filePath: null == filePath ? _self.filePath : filePath // ignore: cast_nullable_to_non_nullable
as String,lastOpened: null == lastOpened ? _self.lastOpened : lastOpened // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

}


/// Adds pattern-matching-related methods to [RecentFileEntry].
extension RecentFileEntryPatterns on RecentFileEntry {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RecentFileEntry value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RecentFileEntry() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RecentFileEntry value)  $default,){
final _that = this;
switch (_that) {
case _RecentFileEntry():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RecentFileEntry value)?  $default,){
final _that = this;
switch (_that) {
case _RecentFileEntry() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String fileId,  String fileName,  String filePath,  DateTime lastOpened)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RecentFileEntry() when $default != null:
return $default(_that.fileId,_that.fileName,_that.filePath,_that.lastOpened);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String fileId,  String fileName,  String filePath,  DateTime lastOpened)  $default,) {final _that = this;
switch (_that) {
case _RecentFileEntry():
return $default(_that.fileId,_that.fileName,_that.filePath,_that.lastOpened);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String fileId,  String fileName,  String filePath,  DateTime lastOpened)?  $default,) {final _that = this;
switch (_that) {
case _RecentFileEntry() when $default != null:
return $default(_that.fileId,_that.fileName,_that.filePath,_that.lastOpened);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RecentFileEntry extends RecentFileEntry {
  const _RecentFileEntry({required this.fileId, required this.fileName, required this.filePath, required this.lastOpened}): super._();
  factory _RecentFileEntry.fromJson(Map<String, dynamic> json) => _$RecentFileEntryFromJson(json);

@override final  String fileId;
@override final  String fileName;
@override final  String filePath;
@override final  DateTime lastOpened;

/// Create a copy of RecentFileEntry
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RecentFileEntryCopyWith<_RecentFileEntry> get copyWith => __$RecentFileEntryCopyWithImpl<_RecentFileEntry>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RecentFileEntryToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _RecentFileEntry&&(identical(other.fileId, fileId) || other.fileId == fileId)&&(identical(other.fileName, fileName) || other.fileName == fileName)&&(identical(other.filePath, filePath) || other.filePath == filePath)&&(identical(other.lastOpened, lastOpened) || other.lastOpened == lastOpened));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,fileId,fileName,filePath,lastOpened);

@override
String toString() {
  return 'RecentFileEntry(fileId: $fileId, fileName: $fileName, filePath: $filePath, lastOpened: $lastOpened)';
}


}

/// @nodoc
abstract mixin class _$RecentFileEntryCopyWith<$Res> implements $RecentFileEntryCopyWith<$Res> {
  factory _$RecentFileEntryCopyWith(_RecentFileEntry value, $Res Function(_RecentFileEntry) _then) = __$RecentFileEntryCopyWithImpl;
@override @useResult
$Res call({
 String fileId, String fileName, String filePath, DateTime lastOpened
});




}
/// @nodoc
class __$RecentFileEntryCopyWithImpl<$Res>
    implements _$RecentFileEntryCopyWith<$Res> {
  __$RecentFileEntryCopyWithImpl(this._self, this._then);

  final _RecentFileEntry _self;
  final $Res Function(_RecentFileEntry) _then;

/// Create a copy of RecentFileEntry
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fileId = null,Object? fileName = null,Object? filePath = null,Object? lastOpened = null,}) {
  return _then(_RecentFileEntry(
fileId: null == fileId ? _self.fileId : fileId // ignore: cast_nullable_to_non_nullable
as String,fileName: null == fileName ? _self.fileName : fileName // ignore: cast_nullable_to_non_nullable
as String,filePath: null == filePath ? _self.filePath : filePath // ignore: cast_nullable_to_non_nullable
as String,lastOpened: null == lastOpened ? _self.lastOpened : lastOpened // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}


}

// dart format on
