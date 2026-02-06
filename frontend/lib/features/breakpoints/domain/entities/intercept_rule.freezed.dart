// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'intercept_rule.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RuleStringMatch {

 String get equals; String get prefix; String get suffix; String get contains; List<String> get anyOf; String get regex;
/// Create a copy of RuleStringMatch
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<RuleStringMatch> get copyWith => _$RuleStringMatchCopyWithImpl<RuleStringMatch>(this as RuleStringMatch, _$identity);

  /// Serializes this RuleStringMatch to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RuleStringMatch&&(identical(other.equals, equals) || other.equals == equals)&&(identical(other.prefix, prefix) || other.prefix == prefix)&&(identical(other.suffix, suffix) || other.suffix == suffix)&&(identical(other.contains, contains) || other.contains == contains)&&const DeepCollectionEquality().equals(other.anyOf, anyOf)&&(identical(other.regex, regex) || other.regex == regex));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,equals,prefix,suffix,contains,const DeepCollectionEquality().hash(anyOf),regex);

@override
String toString() {
  return 'RuleStringMatch(equals: $equals, prefix: $prefix, suffix: $suffix, contains: $contains, anyOf: $anyOf, regex: $regex)';
}


}

/// @nodoc
abstract mixin class $RuleStringMatchCopyWith<$Res>  {
  factory $RuleStringMatchCopyWith(RuleStringMatch value, $Res Function(RuleStringMatch) _then) = _$RuleStringMatchCopyWithImpl;
@useResult
$Res call({
 String equals, String prefix, String suffix, String contains, List<String> anyOf, String regex
});




}
/// @nodoc
class _$RuleStringMatchCopyWithImpl<$Res>
    implements $RuleStringMatchCopyWith<$Res> {
  _$RuleStringMatchCopyWithImpl(this._self, this._then);

  final RuleStringMatch _self;
  final $Res Function(RuleStringMatch) _then;

/// Create a copy of RuleStringMatch
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? equals = null,Object? prefix = null,Object? suffix = null,Object? contains = null,Object? anyOf = null,Object? regex = null,}) {
  return _then(_self.copyWith(
equals: null == equals ? _self.equals : equals // ignore: cast_nullable_to_non_nullable
as String,prefix: null == prefix ? _self.prefix : prefix // ignore: cast_nullable_to_non_nullable
as String,suffix: null == suffix ? _self.suffix : suffix // ignore: cast_nullable_to_non_nullable
as String,contains: null == contains ? _self.contains : contains // ignore: cast_nullable_to_non_nullable
as String,anyOf: null == anyOf ? _self.anyOf : anyOf // ignore: cast_nullable_to_non_nullable
as List<String>,regex: null == regex ? _self.regex : regex // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [RuleStringMatch].
extension RuleStringMatchPatterns on RuleStringMatch {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RuleStringMatch value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RuleStringMatch() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RuleStringMatch value)  $default,){
final _that = this;
switch (_that) {
case _RuleStringMatch():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RuleStringMatch value)?  $default,){
final _that = this;
switch (_that) {
case _RuleStringMatch() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String equals,  String prefix,  String suffix,  String contains,  List<String> anyOf,  String regex)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RuleStringMatch() when $default != null:
return $default(_that.equals,_that.prefix,_that.suffix,_that.contains,_that.anyOf,_that.regex);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String equals,  String prefix,  String suffix,  String contains,  List<String> anyOf,  String regex)  $default,) {final _that = this;
switch (_that) {
case _RuleStringMatch():
return $default(_that.equals,_that.prefix,_that.suffix,_that.contains,_that.anyOf,_that.regex);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String equals,  String prefix,  String suffix,  String contains,  List<String> anyOf,  String regex)?  $default,) {final _that = this;
switch (_that) {
case _RuleStringMatch() when $default != null:
return $default(_that.equals,_that.prefix,_that.suffix,_that.contains,_that.anyOf,_that.regex);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RuleStringMatch extends RuleStringMatch {
  const _RuleStringMatch({this.equals = '', this.prefix = '', this.suffix = '', this.contains = '', final  List<String> anyOf = const [], this.regex = ''}): _anyOf = anyOf,super._();
  factory _RuleStringMatch.fromJson(Map<String, dynamic> json) => _$RuleStringMatchFromJson(json);

@override@JsonKey() final  String equals;
@override@JsonKey() final  String prefix;
@override@JsonKey() final  String suffix;
@override@JsonKey() final  String contains;
 final  List<String> _anyOf;
@override@JsonKey() List<String> get anyOf {
  if (_anyOf is EqualUnmodifiableListView) return _anyOf;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_anyOf);
}

@override@JsonKey() final  String regex;

/// Create a copy of RuleStringMatch
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RuleStringMatchCopyWith<_RuleStringMatch> get copyWith => __$RuleStringMatchCopyWithImpl<_RuleStringMatch>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RuleStringMatchToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _RuleStringMatch&&(identical(other.equals, equals) || other.equals == equals)&&(identical(other.prefix, prefix) || other.prefix == prefix)&&(identical(other.suffix, suffix) || other.suffix == suffix)&&(identical(other.contains, contains) || other.contains == contains)&&const DeepCollectionEquality().equals(other._anyOf, _anyOf)&&(identical(other.regex, regex) || other.regex == regex));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,equals,prefix,suffix,contains,const DeepCollectionEquality().hash(_anyOf),regex);

@override
String toString() {
  return 'RuleStringMatch(equals: $equals, prefix: $prefix, suffix: $suffix, contains: $contains, anyOf: $anyOf, regex: $regex)';
}


}

/// @nodoc
abstract mixin class _$RuleStringMatchCopyWith<$Res> implements $RuleStringMatchCopyWith<$Res> {
  factory _$RuleStringMatchCopyWith(_RuleStringMatch value, $Res Function(_RuleStringMatch) _then) = __$RuleStringMatchCopyWithImpl;
@override @useResult
$Res call({
 String equals, String prefix, String suffix, String contains, List<String> anyOf, String regex
});




}
/// @nodoc
class __$RuleStringMatchCopyWithImpl<$Res>
    implements _$RuleStringMatchCopyWith<$Res> {
  __$RuleStringMatchCopyWithImpl(this._self, this._then);

  final _RuleStringMatch _self;
  final $Res Function(_RuleStringMatch) _then;

/// Create a copy of RuleStringMatch
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? equals = null,Object? prefix = null,Object? suffix = null,Object? contains = null,Object? anyOf = null,Object? regex = null,}) {
  return _then(_RuleStringMatch(
equals: null == equals ? _self.equals : equals // ignore: cast_nullable_to_non_nullable
as String,prefix: null == prefix ? _self.prefix : prefix // ignore: cast_nullable_to_non_nullable
as String,suffix: null == suffix ? _self.suffix : suffix // ignore: cast_nullable_to_non_nullable
as String,contains: null == contains ? _self.contains : contains // ignore: cast_nullable_to_non_nullable
as String,anyOf: null == anyOf ? _self._anyOf : anyOf // ignore: cast_nullable_to_non_nullable
as List<String>,regex: null == regex ? _self.regex : regex // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$RuleHeaderMatch {

 RuleStringMatch get name; RuleStringMatch get value;
/// Create a copy of RuleHeaderMatch
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RuleHeaderMatchCopyWith<RuleHeaderMatch> get copyWith => _$RuleHeaderMatchCopyWithImpl<RuleHeaderMatch>(this as RuleHeaderMatch, _$identity);

  /// Serializes this RuleHeaderMatch to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RuleHeaderMatch&&(identical(other.name, name) || other.name == name)&&(identical(other.value, value) || other.value == value));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,name,value);

@override
String toString() {
  return 'RuleHeaderMatch(name: $name, value: $value)';
}


}

/// @nodoc
abstract mixin class $RuleHeaderMatchCopyWith<$Res>  {
  factory $RuleHeaderMatchCopyWith(RuleHeaderMatch value, $Res Function(RuleHeaderMatch) _then) = _$RuleHeaderMatchCopyWithImpl;
@useResult
$Res call({
 RuleStringMatch name, RuleStringMatch value
});


$RuleStringMatchCopyWith<$Res> get name;$RuleStringMatchCopyWith<$Res> get value;

}
/// @nodoc
class _$RuleHeaderMatchCopyWithImpl<$Res>
    implements $RuleHeaderMatchCopyWith<$Res> {
  _$RuleHeaderMatchCopyWithImpl(this._self, this._then);

  final RuleHeaderMatch _self;
  final $Res Function(RuleHeaderMatch) _then;

/// Create a copy of RuleHeaderMatch
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? name = null,Object? value = null,}) {
  return _then(_self.copyWith(
name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as RuleStringMatch,value: null == value ? _self.value : value // ignore: cast_nullable_to_non_nullable
as RuleStringMatch,
  ));
}
/// Create a copy of RuleHeaderMatch
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res> get name {
  
  return $RuleStringMatchCopyWith<$Res>(_self.name, (value) {
    return _then(_self.copyWith(name: value));
  });
}/// Create a copy of RuleHeaderMatch
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res> get value {
  
  return $RuleStringMatchCopyWith<$Res>(_self.value, (value) {
    return _then(_self.copyWith(value: value));
  });
}
}


/// Adds pattern-matching-related methods to [RuleHeaderMatch].
extension RuleHeaderMatchPatterns on RuleHeaderMatch {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RuleHeaderMatch value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RuleHeaderMatch() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RuleHeaderMatch value)  $default,){
final _that = this;
switch (_that) {
case _RuleHeaderMatch():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RuleHeaderMatch value)?  $default,){
final _that = this;
switch (_that) {
case _RuleHeaderMatch() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( RuleStringMatch name,  RuleStringMatch value)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RuleHeaderMatch() when $default != null:
return $default(_that.name,_that.value);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( RuleStringMatch name,  RuleStringMatch value)  $default,) {final _that = this;
switch (_that) {
case _RuleHeaderMatch():
return $default(_that.name,_that.value);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( RuleStringMatch name,  RuleStringMatch value)?  $default,) {final _that = this;
switch (_that) {
case _RuleHeaderMatch() when $default != null:
return $default(_that.name,_that.value);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RuleHeaderMatch extends RuleHeaderMatch {
  const _RuleHeaderMatch({this.name = const RuleStringMatch(), this.value = const RuleStringMatch()}): super._();
  factory _RuleHeaderMatch.fromJson(Map<String, dynamic> json) => _$RuleHeaderMatchFromJson(json);

@override@JsonKey() final  RuleStringMatch name;
@override@JsonKey() final  RuleStringMatch value;

/// Create a copy of RuleHeaderMatch
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RuleHeaderMatchCopyWith<_RuleHeaderMatch> get copyWith => __$RuleHeaderMatchCopyWithImpl<_RuleHeaderMatch>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RuleHeaderMatchToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _RuleHeaderMatch&&(identical(other.name, name) || other.name == name)&&(identical(other.value, value) || other.value == value));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,name,value);

@override
String toString() {
  return 'RuleHeaderMatch(name: $name, value: $value)';
}


}

/// @nodoc
abstract mixin class _$RuleHeaderMatchCopyWith<$Res> implements $RuleHeaderMatchCopyWith<$Res> {
  factory _$RuleHeaderMatchCopyWith(_RuleHeaderMatch value, $Res Function(_RuleHeaderMatch) _then) = __$RuleHeaderMatchCopyWithImpl;
@override @useResult
$Res call({
 RuleStringMatch name, RuleStringMatch value
});


@override $RuleStringMatchCopyWith<$Res> get name;@override $RuleStringMatchCopyWith<$Res> get value;

}
/// @nodoc
class __$RuleHeaderMatchCopyWithImpl<$Res>
    implements _$RuleHeaderMatchCopyWith<$Res> {
  __$RuleHeaderMatchCopyWithImpl(this._self, this._then);

  final _RuleHeaderMatch _self;
  final $Res Function(_RuleHeaderMatch) _then;

/// Create a copy of RuleHeaderMatch
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? name = null,Object? value = null,}) {
  return _then(_RuleHeaderMatch(
name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as RuleStringMatch,value: null == value ? _self.value : value // ignore: cast_nullable_to_non_nullable
as RuleStringMatch,
  ));
}

/// Create a copy of RuleHeaderMatch
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res> get name {
  
  return $RuleStringMatchCopyWith<$Res>(_self.name, (value) {
    return _then(_self.copyWith(name: value));
  });
}/// Create a copy of RuleHeaderMatch
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res> get value {
  
  return $RuleStringMatchCopyWith<$Res>(_self.value, (value) {
    return _then(_self.copyWith(value: value));
  });
}
}


/// @nodoc
mixin _$RuleStatusMatch {

@JsonKey(name: 'equals') List<int> get statusEquals; int get from; int get to; bool get is4xx; bool get is5xx;
/// Create a copy of RuleStatusMatch
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RuleStatusMatchCopyWith<RuleStatusMatch> get copyWith => _$RuleStatusMatchCopyWithImpl<RuleStatusMatch>(this as RuleStatusMatch, _$identity);

  /// Serializes this RuleStatusMatch to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RuleStatusMatch&&const DeepCollectionEquality().equals(other.statusEquals, statusEquals)&&(identical(other.from, from) || other.from == from)&&(identical(other.to, to) || other.to == to)&&(identical(other.is4xx, is4xx) || other.is4xx == is4xx)&&(identical(other.is5xx, is5xx) || other.is5xx == is5xx));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(statusEquals),from,to,is4xx,is5xx);

@override
String toString() {
  return 'RuleStatusMatch(statusEquals: $statusEquals, from: $from, to: $to, is4xx: $is4xx, is5xx: $is5xx)';
}


}

/// @nodoc
abstract mixin class $RuleStatusMatchCopyWith<$Res>  {
  factory $RuleStatusMatchCopyWith(RuleStatusMatch value, $Res Function(RuleStatusMatch) _then) = _$RuleStatusMatchCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'equals') List<int> statusEquals, int from, int to, bool is4xx, bool is5xx
});




}
/// @nodoc
class _$RuleStatusMatchCopyWithImpl<$Res>
    implements $RuleStatusMatchCopyWith<$Res> {
  _$RuleStatusMatchCopyWithImpl(this._self, this._then);

  final RuleStatusMatch _self;
  final $Res Function(RuleStatusMatch) _then;

/// Create a copy of RuleStatusMatch
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? statusEquals = null,Object? from = null,Object? to = null,Object? is4xx = null,Object? is5xx = null,}) {
  return _then(_self.copyWith(
statusEquals: null == statusEquals ? _self.statusEquals : statusEquals // ignore: cast_nullable_to_non_nullable
as List<int>,from: null == from ? _self.from : from // ignore: cast_nullable_to_non_nullable
as int,to: null == to ? _self.to : to // ignore: cast_nullable_to_non_nullable
as int,is4xx: null == is4xx ? _self.is4xx : is4xx // ignore: cast_nullable_to_non_nullable
as bool,is5xx: null == is5xx ? _self.is5xx : is5xx // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [RuleStatusMatch].
extension RuleStatusMatchPatterns on RuleStatusMatch {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RuleStatusMatch value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RuleStatusMatch() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RuleStatusMatch value)  $default,){
final _that = this;
switch (_that) {
case _RuleStatusMatch():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RuleStatusMatch value)?  $default,){
final _that = this;
switch (_that) {
case _RuleStatusMatch() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'equals')  List<int> statusEquals,  int from,  int to,  bool is4xx,  bool is5xx)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RuleStatusMatch() when $default != null:
return $default(_that.statusEquals,_that.from,_that.to,_that.is4xx,_that.is5xx);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'equals')  List<int> statusEquals,  int from,  int to,  bool is4xx,  bool is5xx)  $default,) {final _that = this;
switch (_that) {
case _RuleStatusMatch():
return $default(_that.statusEquals,_that.from,_that.to,_that.is4xx,_that.is5xx);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'equals')  List<int> statusEquals,  int from,  int to,  bool is4xx,  bool is5xx)?  $default,) {final _that = this;
switch (_that) {
case _RuleStatusMatch() when $default != null:
return $default(_that.statusEquals,_that.from,_that.to,_that.is4xx,_that.is5xx);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RuleStatusMatch extends RuleStatusMatch {
  const _RuleStatusMatch({@JsonKey(name: 'equals') final  List<int> statusEquals = const [], this.from = 0, this.to = 0, this.is4xx = false, this.is5xx = false}): _statusEquals = statusEquals,super._();
  factory _RuleStatusMatch.fromJson(Map<String, dynamic> json) => _$RuleStatusMatchFromJson(json);

 final  List<int> _statusEquals;
@override@JsonKey(name: 'equals') List<int> get statusEquals {
  if (_statusEquals is EqualUnmodifiableListView) return _statusEquals;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_statusEquals);
}

@override@JsonKey() final  int from;
@override@JsonKey() final  int to;
@override@JsonKey() final  bool is4xx;
@override@JsonKey() final  bool is5xx;

/// Create a copy of RuleStatusMatch
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RuleStatusMatchCopyWith<_RuleStatusMatch> get copyWith => __$RuleStatusMatchCopyWithImpl<_RuleStatusMatch>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RuleStatusMatchToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _RuleStatusMatch&&const DeepCollectionEquality().equals(other._statusEquals, _statusEquals)&&(identical(other.from, from) || other.from == from)&&(identical(other.to, to) || other.to == to)&&(identical(other.is4xx, is4xx) || other.is4xx == is4xx)&&(identical(other.is5xx, is5xx) || other.is5xx == is5xx));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(_statusEquals),from,to,is4xx,is5xx);

@override
String toString() {
  return 'RuleStatusMatch(statusEquals: $statusEquals, from: $from, to: $to, is4xx: $is4xx, is5xx: $is5xx)';
}


}

/// @nodoc
abstract mixin class _$RuleStatusMatchCopyWith<$Res> implements $RuleStatusMatchCopyWith<$Res> {
  factory _$RuleStatusMatchCopyWith(_RuleStatusMatch value, $Res Function(_RuleStatusMatch) _then) = __$RuleStatusMatchCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'equals') List<int> statusEquals, int from, int to, bool is4xx, bool is5xx
});




}
/// @nodoc
class __$RuleStatusMatchCopyWithImpl<$Res>
    implements _$RuleStatusMatchCopyWith<$Res> {
  __$RuleStatusMatchCopyWithImpl(this._self, this._then);

  final _RuleStatusMatch _self;
  final $Res Function(_RuleStatusMatch) _then;

/// Create a copy of RuleStatusMatch
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? statusEquals = null,Object? from = null,Object? to = null,Object? is4xx = null,Object? is5xx = null,}) {
  return _then(_RuleStatusMatch(
statusEquals: null == statusEquals ? _self._statusEquals : statusEquals // ignore: cast_nullable_to_non_nullable
as List<int>,from: null == from ? _self.from : from // ignore: cast_nullable_to_non_nullable
as int,to: null == to ? _self.to : to // ignore: cast_nullable_to_non_nullable
as int,is4xx: null == is4xx ? _self.is4xx : is4xx // ignore: cast_nullable_to_non_nullable
as bool,is5xx: null == is5xx ? _self.is5xx : is5xx // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}


/// @nodoc
mixin _$InterceptWhen {

 List<String> get method; List<String> get scheme;@JsonKey(includeIfNull: false) RuleStringMatch? get host;@JsonKey(includeIfNull: false) RuleStringMatch? get port;@JsonKey(includeIfNull: false) RuleStringMatch? get path;@JsonKey(includeIfNull: false) RuleStringMatch? get contentType;@JsonKey(name: 'responseStatus', includeIfNull: false) RuleStatusMatch? get responseStatus;@JsonKey(includeIfNull: false) RuleHeaderMatch? get header;@JsonKey(includeIfNull: false) String? get bodyContains;
/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$InterceptWhenCopyWith<InterceptWhen> get copyWith => _$InterceptWhenCopyWithImpl<InterceptWhen>(this as InterceptWhen, _$identity);

  /// Serializes this InterceptWhen to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is InterceptWhen&&const DeepCollectionEquality().equals(other.method, method)&&const DeepCollectionEquality().equals(other.scheme, scheme)&&(identical(other.host, host) || other.host == host)&&(identical(other.port, port) || other.port == port)&&(identical(other.path, path) || other.path == path)&&(identical(other.contentType, contentType) || other.contentType == contentType)&&(identical(other.responseStatus, responseStatus) || other.responseStatus == responseStatus)&&(identical(other.header, header) || other.header == header)&&(identical(other.bodyContains, bodyContains) || other.bodyContains == bodyContains));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(method),const DeepCollectionEquality().hash(scheme),host,port,path,contentType,responseStatus,header,bodyContains);

@override
String toString() {
  return 'InterceptWhen(method: $method, scheme: $scheme, host: $host, port: $port, path: $path, contentType: $contentType, responseStatus: $responseStatus, header: $header, bodyContains: $bodyContains)';
}


}

/// @nodoc
abstract mixin class $InterceptWhenCopyWith<$Res>  {
  factory $InterceptWhenCopyWith(InterceptWhen value, $Res Function(InterceptWhen) _then) = _$InterceptWhenCopyWithImpl;
@useResult
$Res call({
 List<String> method, List<String> scheme,@JsonKey(includeIfNull: false) RuleStringMatch? host,@JsonKey(includeIfNull: false) RuleStringMatch? port,@JsonKey(includeIfNull: false) RuleStringMatch? path,@JsonKey(includeIfNull: false) RuleStringMatch? contentType,@JsonKey(name: 'responseStatus', includeIfNull: false) RuleStatusMatch? responseStatus,@JsonKey(includeIfNull: false) RuleHeaderMatch? header,@JsonKey(includeIfNull: false) String? bodyContains
});


$RuleStringMatchCopyWith<$Res>? get host;$RuleStringMatchCopyWith<$Res>? get port;$RuleStringMatchCopyWith<$Res>? get path;$RuleStringMatchCopyWith<$Res>? get contentType;$RuleStatusMatchCopyWith<$Res>? get responseStatus;$RuleHeaderMatchCopyWith<$Res>? get header;

}
/// @nodoc
class _$InterceptWhenCopyWithImpl<$Res>
    implements $InterceptWhenCopyWith<$Res> {
  _$InterceptWhenCopyWithImpl(this._self, this._then);

  final InterceptWhen _self;
  final $Res Function(InterceptWhen) _then;

/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? method = null,Object? scheme = null,Object? host = freezed,Object? port = freezed,Object? path = freezed,Object? contentType = freezed,Object? responseStatus = freezed,Object? header = freezed,Object? bodyContains = freezed,}) {
  return _then(_self.copyWith(
method: null == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as List<String>,scheme: null == scheme ? _self.scheme : scheme // ignore: cast_nullable_to_non_nullable
as List<String>,host: freezed == host ? _self.host : host // ignore: cast_nullable_to_non_nullable
as RuleStringMatch?,port: freezed == port ? _self.port : port // ignore: cast_nullable_to_non_nullable
as RuleStringMatch?,path: freezed == path ? _self.path : path // ignore: cast_nullable_to_non_nullable
as RuleStringMatch?,contentType: freezed == contentType ? _self.contentType : contentType // ignore: cast_nullable_to_non_nullable
as RuleStringMatch?,responseStatus: freezed == responseStatus ? _self.responseStatus : responseStatus // ignore: cast_nullable_to_non_nullable
as RuleStatusMatch?,header: freezed == header ? _self.header : header // ignore: cast_nullable_to_non_nullable
as RuleHeaderMatch?,bodyContains: freezed == bodyContains ? _self.bodyContains : bodyContains // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}
/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res>? get host {
    if (_self.host == null) {
    return null;
  }

  return $RuleStringMatchCopyWith<$Res>(_self.host!, (value) {
    return _then(_self.copyWith(host: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res>? get port {
    if (_self.port == null) {
    return null;
  }

  return $RuleStringMatchCopyWith<$Res>(_self.port!, (value) {
    return _then(_self.copyWith(port: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res>? get path {
    if (_self.path == null) {
    return null;
  }

  return $RuleStringMatchCopyWith<$Res>(_self.path!, (value) {
    return _then(_self.copyWith(path: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res>? get contentType {
    if (_self.contentType == null) {
    return null;
  }

  return $RuleStringMatchCopyWith<$Res>(_self.contentType!, (value) {
    return _then(_self.copyWith(contentType: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStatusMatchCopyWith<$Res>? get responseStatus {
    if (_self.responseStatus == null) {
    return null;
  }

  return $RuleStatusMatchCopyWith<$Res>(_self.responseStatus!, (value) {
    return _then(_self.copyWith(responseStatus: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleHeaderMatchCopyWith<$Res>? get header {
    if (_self.header == null) {
    return null;
  }

  return $RuleHeaderMatchCopyWith<$Res>(_self.header!, (value) {
    return _then(_self.copyWith(header: value));
  });
}
}


/// Adds pattern-matching-related methods to [InterceptWhen].
extension InterceptWhenPatterns on InterceptWhen {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _InterceptWhen value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _InterceptWhen() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _InterceptWhen value)  $default,){
final _that = this;
switch (_that) {
case _InterceptWhen():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _InterceptWhen value)?  $default,){
final _that = this;
switch (_that) {
case _InterceptWhen() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<String> method,  List<String> scheme, @JsonKey(includeIfNull: false)  RuleStringMatch? host, @JsonKey(includeIfNull: false)  RuleStringMatch? port, @JsonKey(includeIfNull: false)  RuleStringMatch? path, @JsonKey(includeIfNull: false)  RuleStringMatch? contentType, @JsonKey(name: 'responseStatus', includeIfNull: false)  RuleStatusMatch? responseStatus, @JsonKey(includeIfNull: false)  RuleHeaderMatch? header, @JsonKey(includeIfNull: false)  String? bodyContains)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _InterceptWhen() when $default != null:
return $default(_that.method,_that.scheme,_that.host,_that.port,_that.path,_that.contentType,_that.responseStatus,_that.header,_that.bodyContains);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<String> method,  List<String> scheme, @JsonKey(includeIfNull: false)  RuleStringMatch? host, @JsonKey(includeIfNull: false)  RuleStringMatch? port, @JsonKey(includeIfNull: false)  RuleStringMatch? path, @JsonKey(includeIfNull: false)  RuleStringMatch? contentType, @JsonKey(name: 'responseStatus', includeIfNull: false)  RuleStatusMatch? responseStatus, @JsonKey(includeIfNull: false)  RuleHeaderMatch? header, @JsonKey(includeIfNull: false)  String? bodyContains)  $default,) {final _that = this;
switch (_that) {
case _InterceptWhen():
return $default(_that.method,_that.scheme,_that.host,_that.port,_that.path,_that.contentType,_that.responseStatus,_that.header,_that.bodyContains);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<String> method,  List<String> scheme, @JsonKey(includeIfNull: false)  RuleStringMatch? host, @JsonKey(includeIfNull: false)  RuleStringMatch? port, @JsonKey(includeIfNull: false)  RuleStringMatch? path, @JsonKey(includeIfNull: false)  RuleStringMatch? contentType, @JsonKey(name: 'responseStatus', includeIfNull: false)  RuleStatusMatch? responseStatus, @JsonKey(includeIfNull: false)  RuleHeaderMatch? header, @JsonKey(includeIfNull: false)  String? bodyContains)?  $default,) {final _that = this;
switch (_that) {
case _InterceptWhen() when $default != null:
return $default(_that.method,_that.scheme,_that.host,_that.port,_that.path,_that.contentType,_that.responseStatus,_that.header,_that.bodyContains);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _InterceptWhen extends InterceptWhen {
  const _InterceptWhen({final  List<String> method = const [], final  List<String> scheme = const [], @JsonKey(includeIfNull: false) this.host, @JsonKey(includeIfNull: false) this.port, @JsonKey(includeIfNull: false) this.path, @JsonKey(includeIfNull: false) this.contentType, @JsonKey(name: 'responseStatus', includeIfNull: false) this.responseStatus, @JsonKey(includeIfNull: false) this.header, @JsonKey(includeIfNull: false) this.bodyContains}): _method = method,_scheme = scheme,super._();
  factory _InterceptWhen.fromJson(Map<String, dynamic> json) => _$InterceptWhenFromJson(json);

 final  List<String> _method;
@override@JsonKey() List<String> get method {
  if (_method is EqualUnmodifiableListView) return _method;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_method);
}

 final  List<String> _scheme;
@override@JsonKey() List<String> get scheme {
  if (_scheme is EqualUnmodifiableListView) return _scheme;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_scheme);
}

@override@JsonKey(includeIfNull: false) final  RuleStringMatch? host;
@override@JsonKey(includeIfNull: false) final  RuleStringMatch? port;
@override@JsonKey(includeIfNull: false) final  RuleStringMatch? path;
@override@JsonKey(includeIfNull: false) final  RuleStringMatch? contentType;
@override@JsonKey(name: 'responseStatus', includeIfNull: false) final  RuleStatusMatch? responseStatus;
@override@JsonKey(includeIfNull: false) final  RuleHeaderMatch? header;
@override@JsonKey(includeIfNull: false) final  String? bodyContains;

/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$InterceptWhenCopyWith<_InterceptWhen> get copyWith => __$InterceptWhenCopyWithImpl<_InterceptWhen>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$InterceptWhenToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _InterceptWhen&&const DeepCollectionEquality().equals(other._method, _method)&&const DeepCollectionEquality().equals(other._scheme, _scheme)&&(identical(other.host, host) || other.host == host)&&(identical(other.port, port) || other.port == port)&&(identical(other.path, path) || other.path == path)&&(identical(other.contentType, contentType) || other.contentType == contentType)&&(identical(other.responseStatus, responseStatus) || other.responseStatus == responseStatus)&&(identical(other.header, header) || other.header == header)&&(identical(other.bodyContains, bodyContains) || other.bodyContains == bodyContains));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(_method),const DeepCollectionEquality().hash(_scheme),host,port,path,contentType,responseStatus,header,bodyContains);

@override
String toString() {
  return 'InterceptWhen(method: $method, scheme: $scheme, host: $host, port: $port, path: $path, contentType: $contentType, responseStatus: $responseStatus, header: $header, bodyContains: $bodyContains)';
}


}

/// @nodoc
abstract mixin class _$InterceptWhenCopyWith<$Res> implements $InterceptWhenCopyWith<$Res> {
  factory _$InterceptWhenCopyWith(_InterceptWhen value, $Res Function(_InterceptWhen) _then) = __$InterceptWhenCopyWithImpl;
@override @useResult
$Res call({
 List<String> method, List<String> scheme,@JsonKey(includeIfNull: false) RuleStringMatch? host,@JsonKey(includeIfNull: false) RuleStringMatch? port,@JsonKey(includeIfNull: false) RuleStringMatch? path,@JsonKey(includeIfNull: false) RuleStringMatch? contentType,@JsonKey(name: 'responseStatus', includeIfNull: false) RuleStatusMatch? responseStatus,@JsonKey(includeIfNull: false) RuleHeaderMatch? header,@JsonKey(includeIfNull: false) String? bodyContains
});


@override $RuleStringMatchCopyWith<$Res>? get host;@override $RuleStringMatchCopyWith<$Res>? get port;@override $RuleStringMatchCopyWith<$Res>? get path;@override $RuleStringMatchCopyWith<$Res>? get contentType;@override $RuleStatusMatchCopyWith<$Res>? get responseStatus;@override $RuleHeaderMatchCopyWith<$Res>? get header;

}
/// @nodoc
class __$InterceptWhenCopyWithImpl<$Res>
    implements _$InterceptWhenCopyWith<$Res> {
  __$InterceptWhenCopyWithImpl(this._self, this._then);

  final _InterceptWhen _self;
  final $Res Function(_InterceptWhen) _then;

/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? method = null,Object? scheme = null,Object? host = freezed,Object? port = freezed,Object? path = freezed,Object? contentType = freezed,Object? responseStatus = freezed,Object? header = freezed,Object? bodyContains = freezed,}) {
  return _then(_InterceptWhen(
method: null == method ? _self._method : method // ignore: cast_nullable_to_non_nullable
as List<String>,scheme: null == scheme ? _self._scheme : scheme // ignore: cast_nullable_to_non_nullable
as List<String>,host: freezed == host ? _self.host : host // ignore: cast_nullable_to_non_nullable
as RuleStringMatch?,port: freezed == port ? _self.port : port // ignore: cast_nullable_to_non_nullable
as RuleStringMatch?,path: freezed == path ? _self.path : path // ignore: cast_nullable_to_non_nullable
as RuleStringMatch?,contentType: freezed == contentType ? _self.contentType : contentType // ignore: cast_nullable_to_non_nullable
as RuleStringMatch?,responseStatus: freezed == responseStatus ? _self.responseStatus : responseStatus // ignore: cast_nullable_to_non_nullable
as RuleStatusMatch?,header: freezed == header ? _self.header : header // ignore: cast_nullable_to_non_nullable
as RuleHeaderMatch?,bodyContains: freezed == bodyContains ? _self.bodyContains : bodyContains // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res>? get host {
    if (_self.host == null) {
    return null;
  }

  return $RuleStringMatchCopyWith<$Res>(_self.host!, (value) {
    return _then(_self.copyWith(host: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res>? get port {
    if (_self.port == null) {
    return null;
  }

  return $RuleStringMatchCopyWith<$Res>(_self.port!, (value) {
    return _then(_self.copyWith(port: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res>? get path {
    if (_self.path == null) {
    return null;
  }

  return $RuleStringMatchCopyWith<$Res>(_self.path!, (value) {
    return _then(_self.copyWith(path: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStringMatchCopyWith<$Res>? get contentType {
    if (_self.contentType == null) {
    return null;
  }

  return $RuleStringMatchCopyWith<$Res>(_self.contentType!, (value) {
    return _then(_self.copyWith(contentType: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleStatusMatchCopyWith<$Res>? get responseStatus {
    if (_self.responseStatus == null) {
    return null;
  }

  return $RuleStatusMatchCopyWith<$Res>(_self.responseStatus!, (value) {
    return _then(_self.copyWith(responseStatus: value));
  });
}/// Create a copy of InterceptWhen
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RuleHeaderMatchCopyWith<$Res>? get header {
    if (_self.header == null) {
    return null;
  }

  return $RuleHeaderMatchCopyWith<$Res>(_self.header!, (value) {
    return _then(_self.copyWith(header: value));
  });
}
}


/// @nodoc
mixin _$InterceptRule {

 String get id; bool get enabled; int get priority; String get action; bool get once; bool get stopProcessing; InterceptWhen get when; DateTime? get createdAt; DateTime? get updatedAt;
/// Create a copy of InterceptRule
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$InterceptRuleCopyWith<InterceptRule> get copyWith => _$InterceptRuleCopyWithImpl<InterceptRule>(this as InterceptRule, _$identity);

  /// Serializes this InterceptRule to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is InterceptRule&&(identical(other.id, id) || other.id == id)&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.priority, priority) || other.priority == priority)&&(identical(other.action, action) || other.action == action)&&(identical(other.once, once) || other.once == once)&&(identical(other.stopProcessing, stopProcessing) || other.stopProcessing == stopProcessing)&&(identical(other.when, when) || other.when == when)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,enabled,priority,action,once,stopProcessing,when,createdAt,updatedAt);

@override
String toString() {
  return 'InterceptRule(id: $id, enabled: $enabled, priority: $priority, action: $action, once: $once, stopProcessing: $stopProcessing, when: $when, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class $InterceptRuleCopyWith<$Res>  {
  factory $InterceptRuleCopyWith(InterceptRule value, $Res Function(InterceptRule) _then) = _$InterceptRuleCopyWithImpl;
@useResult
$Res call({
 String id, bool enabled, int priority, String action, bool once, bool stopProcessing, InterceptWhen when, DateTime? createdAt, DateTime? updatedAt
});


$InterceptWhenCopyWith<$Res> get when;

}
/// @nodoc
class _$InterceptRuleCopyWithImpl<$Res>
    implements $InterceptRuleCopyWith<$Res> {
  _$InterceptRuleCopyWithImpl(this._self, this._then);

  final InterceptRule _self;
  final $Res Function(InterceptRule) _then;

/// Create a copy of InterceptRule
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? enabled = null,Object? priority = null,Object? action = null,Object? once = null,Object? stopProcessing = null,Object? when = null,Object? createdAt = freezed,Object? updatedAt = freezed,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,priority: null == priority ? _self.priority : priority // ignore: cast_nullable_to_non_nullable
as int,action: null == action ? _self.action : action // ignore: cast_nullable_to_non_nullable
as String,once: null == once ? _self.once : once // ignore: cast_nullable_to_non_nullable
as bool,stopProcessing: null == stopProcessing ? _self.stopProcessing : stopProcessing // ignore: cast_nullable_to_non_nullable
as bool,when: null == when ? _self.when : when // ignore: cast_nullable_to_non_nullable
as InterceptWhen,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}
/// Create a copy of InterceptRule
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$InterceptWhenCopyWith<$Res> get when {
  
  return $InterceptWhenCopyWith<$Res>(_self.when, (value) {
    return _then(_self.copyWith(when: value));
  });
}
}


/// Adds pattern-matching-related methods to [InterceptRule].
extension InterceptRulePatterns on InterceptRule {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _InterceptRule value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _InterceptRule() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _InterceptRule value)  $default,){
final _that = this;
switch (_that) {
case _InterceptRule():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _InterceptRule value)?  $default,){
final _that = this;
switch (_that) {
case _InterceptRule() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  bool enabled,  int priority,  String action,  bool once,  bool stopProcessing,  InterceptWhen when,  DateTime? createdAt,  DateTime? updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _InterceptRule() when $default != null:
return $default(_that.id,_that.enabled,_that.priority,_that.action,_that.once,_that.stopProcessing,_that.when,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  bool enabled,  int priority,  String action,  bool once,  bool stopProcessing,  InterceptWhen when,  DateTime? createdAt,  DateTime? updatedAt)  $default,) {final _that = this;
switch (_that) {
case _InterceptRule():
return $default(_that.id,_that.enabled,_that.priority,_that.action,_that.once,_that.stopProcessing,_that.when,_that.createdAt,_that.updatedAt);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  bool enabled,  int priority,  String action,  bool once,  bool stopProcessing,  InterceptWhen when,  DateTime? createdAt,  DateTime? updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _InterceptRule() when $default != null:
return $default(_that.id,_that.enabled,_that.priority,_that.action,_that.once,_that.stopProcessing,_that.when,_that.createdAt,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _InterceptRule extends InterceptRule {
  const _InterceptRule({this.id = '', this.enabled = true, this.priority = 10, this.action = 'both', this.once = false, this.stopProcessing = false, this.when = const InterceptWhen(), this.createdAt, this.updatedAt}): super._();
  factory _InterceptRule.fromJson(Map<String, dynamic> json) => _$InterceptRuleFromJson(json);

@override@JsonKey() final  String id;
@override@JsonKey() final  bool enabled;
@override@JsonKey() final  int priority;
@override@JsonKey() final  String action;
@override@JsonKey() final  bool once;
@override@JsonKey() final  bool stopProcessing;
@override@JsonKey() final  InterceptWhen when;
@override final  DateTime? createdAt;
@override final  DateTime? updatedAt;

/// Create a copy of InterceptRule
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$InterceptRuleCopyWith<_InterceptRule> get copyWith => __$InterceptRuleCopyWithImpl<_InterceptRule>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$InterceptRuleToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _InterceptRule&&(identical(other.id, id) || other.id == id)&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.priority, priority) || other.priority == priority)&&(identical(other.action, action) || other.action == action)&&(identical(other.once, once) || other.once == once)&&(identical(other.stopProcessing, stopProcessing) || other.stopProcessing == stopProcessing)&&(identical(other.when, when) || other.when == when)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,enabled,priority,action,once,stopProcessing,when,createdAt,updatedAt);

@override
String toString() {
  return 'InterceptRule(id: $id, enabled: $enabled, priority: $priority, action: $action, once: $once, stopProcessing: $stopProcessing, when: $when, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$InterceptRuleCopyWith<$Res> implements $InterceptRuleCopyWith<$Res> {
  factory _$InterceptRuleCopyWith(_InterceptRule value, $Res Function(_InterceptRule) _then) = __$InterceptRuleCopyWithImpl;
@override @useResult
$Res call({
 String id, bool enabled, int priority, String action, bool once, bool stopProcessing, InterceptWhen when, DateTime? createdAt, DateTime? updatedAt
});


@override $InterceptWhenCopyWith<$Res> get when;

}
/// @nodoc
class __$InterceptRuleCopyWithImpl<$Res>
    implements _$InterceptRuleCopyWith<$Res> {
  __$InterceptRuleCopyWithImpl(this._self, this._then);

  final _InterceptRule _self;
  final $Res Function(_InterceptRule) _then;

/// Create a copy of InterceptRule
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? enabled = null,Object? priority = null,Object? action = null,Object? once = null,Object? stopProcessing = null,Object? when = null,Object? createdAt = freezed,Object? updatedAt = freezed,}) {
  return _then(_InterceptRule(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,priority: null == priority ? _self.priority : priority // ignore: cast_nullable_to_non_nullable
as int,action: null == action ? _self.action : action // ignore: cast_nullable_to_non_nullable
as String,once: null == once ? _self.once : once // ignore: cast_nullable_to_non_nullable
as bool,stopProcessing: null == stopProcessing ? _self.stopProcessing : stopProcessing // ignore: cast_nullable_to_non_nullable
as bool,when: null == when ? _self.when : when // ignore: cast_nullable_to_non_nullable
as InterceptWhen,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

/// Create a copy of InterceptRule
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$InterceptWhenCopyWith<$Res> get when {
  
  return $InterceptWhenCopyWith<$Res>(_self.when, (value) {
    return _then(_self.copyWith(when: value));
  });
}
}

// dart format on
