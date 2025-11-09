// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'recent_files_list.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RecentFilesList {

 List<RecentFileEntry> get entries; int get maxEntries;
/// Create a copy of RecentFilesList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RecentFilesListCopyWith<RecentFilesList> get copyWith => _$RecentFilesListCopyWithImpl<RecentFilesList>(this as RecentFilesList, _$identity);

  /// Serializes this RecentFilesList to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RecentFilesList&&const DeepCollectionEquality().equals(other.entries, entries)&&(identical(other.maxEntries, maxEntries) || other.maxEntries == maxEntries));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(entries),maxEntries);

@override
String toString() {
  return 'RecentFilesList(entries: $entries, maxEntries: $maxEntries)';
}


}

/// @nodoc
abstract mixin class $RecentFilesListCopyWith<$Res>  {
  factory $RecentFilesListCopyWith(RecentFilesList value, $Res Function(RecentFilesList) _then) = _$RecentFilesListCopyWithImpl;
@useResult
$Res call({
 List<RecentFileEntry> entries, int maxEntries
});




}
/// @nodoc
class _$RecentFilesListCopyWithImpl<$Res>
    implements $RecentFilesListCopyWith<$Res> {
  _$RecentFilesListCopyWithImpl(this._self, this._then);

  final RecentFilesList _self;
  final $Res Function(RecentFilesList) _then;

/// Create a copy of RecentFilesList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? entries = null,Object? maxEntries = null,}) {
  return _then(_self.copyWith(
entries: null == entries ? _self.entries : entries // ignore: cast_nullable_to_non_nullable
as List<RecentFileEntry>,maxEntries: null == maxEntries ? _self.maxEntries : maxEntries // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

}


/// Adds pattern-matching-related methods to [RecentFilesList].
extension RecentFilesListPatterns on RecentFilesList {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RecentFilesList value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RecentFilesList() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RecentFilesList value)  $default,){
final _that = this;
switch (_that) {
case _RecentFilesList():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RecentFilesList value)?  $default,){
final _that = this;
switch (_that) {
case _RecentFilesList() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<RecentFileEntry> entries,  int maxEntries)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RecentFilesList() when $default != null:
return $default(_that.entries,_that.maxEntries);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<RecentFileEntry> entries,  int maxEntries)  $default,) {final _that = this;
switch (_that) {
case _RecentFilesList():
return $default(_that.entries,_that.maxEntries);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<RecentFileEntry> entries,  int maxEntries)?  $default,) {final _that = this;
switch (_that) {
case _RecentFilesList() when $default != null:
return $default(_that.entries,_that.maxEntries);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RecentFilesList extends RecentFilesList {
  const _RecentFilesList({final  List<RecentFileEntry> entries = const [], this.maxEntries = 10}): _entries = entries,super._();
  factory _RecentFilesList.fromJson(Map<String, dynamic> json) => _$RecentFilesListFromJson(json);

 final  List<RecentFileEntry> _entries;
@override@JsonKey() List<RecentFileEntry> get entries {
  if (_entries is EqualUnmodifiableListView) return _entries;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_entries);
}

@override@JsonKey() final  int maxEntries;

/// Create a copy of RecentFilesList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RecentFilesListCopyWith<_RecentFilesList> get copyWith => __$RecentFilesListCopyWithImpl<_RecentFilesList>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RecentFilesListToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _RecentFilesList&&const DeepCollectionEquality().equals(other._entries, _entries)&&(identical(other.maxEntries, maxEntries) || other.maxEntries == maxEntries));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(_entries),maxEntries);

@override
String toString() {
  return 'RecentFilesList(entries: $entries, maxEntries: $maxEntries)';
}


}

/// @nodoc
abstract mixin class _$RecentFilesListCopyWith<$Res> implements $RecentFilesListCopyWith<$Res> {
  factory _$RecentFilesListCopyWith(_RecentFilesList value, $Res Function(_RecentFilesList) _then) = __$RecentFilesListCopyWithImpl;
@override @useResult
$Res call({
 List<RecentFileEntry> entries, int maxEntries
});




}
/// @nodoc
class __$RecentFilesListCopyWithImpl<$Res>
    implements _$RecentFilesListCopyWith<$Res> {
  __$RecentFilesListCopyWithImpl(this._self, this._then);

  final _RecentFilesList _self;
  final $Res Function(_RecentFilesList) _then;

/// Create a copy of RecentFilesList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? entries = null,Object? maxEntries = null,}) {
  return _then(_RecentFilesList(
entries: null == entries ? _self._entries : entries // ignore: cast_nullable_to_non_nullable
as List<RecentFileEntry>,maxEntries: null == maxEntries ? _self.maxEntries : maxEntries // ignore: cast_nullable_to_non_nullable
as int,
  ));
}


}

// dart format on
