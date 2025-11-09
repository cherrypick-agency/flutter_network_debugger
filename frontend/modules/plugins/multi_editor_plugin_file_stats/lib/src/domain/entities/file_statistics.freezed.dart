// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'file_statistics.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$FileStatistics {

 String get fileId; int get lines; int get characters; int get words; int get bytes; DateTime get calculatedAt;
/// Create a copy of FileStatistics
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$FileStatisticsCopyWith<FileStatistics> get copyWith => _$FileStatisticsCopyWithImpl<FileStatistics>(this as FileStatistics, _$identity);

  /// Serializes this FileStatistics to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is FileStatistics&&(identical(other.fileId, fileId) || other.fileId == fileId)&&(identical(other.lines, lines) || other.lines == lines)&&(identical(other.characters, characters) || other.characters == characters)&&(identical(other.words, words) || other.words == words)&&(identical(other.bytes, bytes) || other.bytes == bytes)&&(identical(other.calculatedAt, calculatedAt) || other.calculatedAt == calculatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,fileId,lines,characters,words,bytes,calculatedAt);

@override
String toString() {
  return 'FileStatistics(fileId: $fileId, lines: $lines, characters: $characters, words: $words, bytes: $bytes, calculatedAt: $calculatedAt)';
}


}

/// @nodoc
abstract mixin class $FileStatisticsCopyWith<$Res>  {
  factory $FileStatisticsCopyWith(FileStatistics value, $Res Function(FileStatistics) _then) = _$FileStatisticsCopyWithImpl;
@useResult
$Res call({
 String fileId, int lines, int characters, int words, int bytes, DateTime calculatedAt
});




}
/// @nodoc
class _$FileStatisticsCopyWithImpl<$Res>
    implements $FileStatisticsCopyWith<$Res> {
  _$FileStatisticsCopyWithImpl(this._self, this._then);

  final FileStatistics _self;
  final $Res Function(FileStatistics) _then;

/// Create a copy of FileStatistics
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fileId = null,Object? lines = null,Object? characters = null,Object? words = null,Object? bytes = null,Object? calculatedAt = null,}) {
  return _then(_self.copyWith(
fileId: null == fileId ? _self.fileId : fileId // ignore: cast_nullable_to_non_nullable
as String,lines: null == lines ? _self.lines : lines // ignore: cast_nullable_to_non_nullable
as int,characters: null == characters ? _self.characters : characters // ignore: cast_nullable_to_non_nullable
as int,words: null == words ? _self.words : words // ignore: cast_nullable_to_non_nullable
as int,bytes: null == bytes ? _self.bytes : bytes // ignore: cast_nullable_to_non_nullable
as int,calculatedAt: null == calculatedAt ? _self.calculatedAt : calculatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

}


/// Adds pattern-matching-related methods to [FileStatistics].
extension FileStatisticsPatterns on FileStatistics {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _FileStatistics value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _FileStatistics() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _FileStatistics value)  $default,){
final _that = this;
switch (_that) {
case _FileStatistics():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _FileStatistics value)?  $default,){
final _that = this;
switch (_that) {
case _FileStatistics() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String fileId,  int lines,  int characters,  int words,  int bytes,  DateTime calculatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _FileStatistics() when $default != null:
return $default(_that.fileId,_that.lines,_that.characters,_that.words,_that.bytes,_that.calculatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String fileId,  int lines,  int characters,  int words,  int bytes,  DateTime calculatedAt)  $default,) {final _that = this;
switch (_that) {
case _FileStatistics():
return $default(_that.fileId,_that.lines,_that.characters,_that.words,_that.bytes,_that.calculatedAt);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String fileId,  int lines,  int characters,  int words,  int bytes,  DateTime calculatedAt)?  $default,) {final _that = this;
switch (_that) {
case _FileStatistics() when $default != null:
return $default(_that.fileId,_that.lines,_that.characters,_that.words,_that.bytes,_that.calculatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _FileStatistics extends FileStatistics {
  const _FileStatistics({required this.fileId, required this.lines, required this.characters, required this.words, required this.bytes, required this.calculatedAt}): super._();
  factory _FileStatistics.fromJson(Map<String, dynamic> json) => _$FileStatisticsFromJson(json);

@override final  String fileId;
@override final  int lines;
@override final  int characters;
@override final  int words;
@override final  int bytes;
@override final  DateTime calculatedAt;

/// Create a copy of FileStatistics
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$FileStatisticsCopyWith<_FileStatistics> get copyWith => __$FileStatisticsCopyWithImpl<_FileStatistics>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$FileStatisticsToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _FileStatistics&&(identical(other.fileId, fileId) || other.fileId == fileId)&&(identical(other.lines, lines) || other.lines == lines)&&(identical(other.characters, characters) || other.characters == characters)&&(identical(other.words, words) || other.words == words)&&(identical(other.bytes, bytes) || other.bytes == bytes)&&(identical(other.calculatedAt, calculatedAt) || other.calculatedAt == calculatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,fileId,lines,characters,words,bytes,calculatedAt);

@override
String toString() {
  return 'FileStatistics(fileId: $fileId, lines: $lines, characters: $characters, words: $words, bytes: $bytes, calculatedAt: $calculatedAt)';
}


}

/// @nodoc
abstract mixin class _$FileStatisticsCopyWith<$Res> implements $FileStatisticsCopyWith<$Res> {
  factory _$FileStatisticsCopyWith(_FileStatistics value, $Res Function(_FileStatistics) _then) = __$FileStatisticsCopyWithImpl;
@override @useResult
$Res call({
 String fileId, int lines, int characters, int words, int bytes, DateTime calculatedAt
});




}
/// @nodoc
class __$FileStatisticsCopyWithImpl<$Res>
    implements _$FileStatisticsCopyWith<$Res> {
  __$FileStatisticsCopyWithImpl(this._self, this._then);

  final _FileStatistics _self;
  final $Res Function(_FileStatistics) _then;

/// Create a copy of FileStatistics
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fileId = null,Object? lines = null,Object? characters = null,Object? words = null,Object? bytes = null,Object? calculatedAt = null,}) {
  return _then(_FileStatistics(
fileId: null == fileId ? _self.fileId : fileId // ignore: cast_nullable_to_non_nullable
as String,lines: null == lines ? _self.lines : lines // ignore: cast_nullable_to_non_nullable
as int,characters: null == characters ? _self.characters : characters // ignore: cast_nullable_to_non_nullable
as int,words: null == words ? _self.words : words // ignore: cast_nullable_to_non_nullable
as int,bytes: null == bytes ? _self.bytes : bytes // ignore: cast_nullable_to_non_nullable
as int,calculatedAt: null == calculatedAt ? _self.calculatedAt : calculatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}


}

// dart format on
