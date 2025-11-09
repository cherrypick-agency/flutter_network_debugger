// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'compiler_status.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;
CompilerStatus _$CompilerStatusFromJson(
  Map<String, dynamic> json
) {
        switch (json['runtimeType']) {
                  case 'installed':
          return _Installed.fromJson(
            json
          );
                case 'installing':
          return _Installing.fromJson(
            json
          );
                case 'notInstalled':
          return _NotInstalled.fromJson(
            json
          );
        
          default:
            throw CheckedFromJsonException(
  json,
  'runtimeType',
  'CompilerStatus',
  'Invalid union type "${json['runtimeType']}"!'
);
        }
      
}

/// @nodoc
mixin _$CompilerStatus {



  /// Serializes this CompilerStatus to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is CompilerStatus);
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => runtimeType.hashCode;

@override
String toString() {
  return 'CompilerStatus()';
}


}

/// @nodoc
class $CompilerStatusCopyWith<$Res>  {
$CompilerStatusCopyWith(CompilerStatus _, $Res Function(CompilerStatus) __);
}


/// Adds pattern-matching-related methods to [CompilerStatus].
extension CompilerStatusPatterns on CompilerStatus {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>({TResult Function( _Installed value)?  installed,TResult Function( _Installing value)?  installing,TResult Function( _NotInstalled value)?  notInstalled,required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Installed() when installed != null:
return installed(_that);case _Installing() when installing != null:
return installing(_that);case _NotInstalled() when notInstalled != null:
return notInstalled(_that);case _:
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

@optionalTypeArgs TResult map<TResult extends Object?>({required TResult Function( _Installed value)  installed,required TResult Function( _Installing value)  installing,required TResult Function( _NotInstalled value)  notInstalled,}){
final _that = this;
switch (_that) {
case _Installed():
return installed(_that);case _Installing():
return installing(_that);case _NotInstalled():
return notInstalled(_that);}
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>({TResult? Function( _Installed value)?  installed,TResult? Function( _Installing value)?  installing,TResult? Function( _NotInstalled value)?  notInstalled,}){
final _that = this;
switch (_that) {
case _Installed() when installed != null:
return installed(_that);case _Installing() when installing != null:
return installing(_that);case _NotInstalled() when notInstalled != null:
return notInstalled(_that);case _:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>({TResult Function()?  installed,TResult Function()?  installing,TResult Function()?  notInstalled,required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Installed() when installed != null:
return installed();case _Installing() when installing != null:
return installing();case _NotInstalled() when notInstalled != null:
return notInstalled();case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>({required TResult Function()  installed,required TResult Function()  installing,required TResult Function()  notInstalled,}) {final _that = this;
switch (_that) {
case _Installed():
return installed();case _Installing():
return installing();case _NotInstalled():
return notInstalled();}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>({TResult? Function()?  installed,TResult? Function()?  installing,TResult? Function()?  notInstalled,}) {final _that = this;
switch (_that) {
case _Installed() when installed != null:
return installed();case _Installing() when installing != null:
return installing();case _NotInstalled() when notInstalled != null:
return notInstalled();case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Installed implements CompilerStatus {
  const _Installed({final  String? $type}): $type = $type ?? 'installed';
  factory _Installed.fromJson(Map<String, dynamic> json) => _$InstalledFromJson(json);



@JsonKey(name: 'runtimeType')
final String $type;



@override
Map<String, dynamic> toJson() {
  return _$InstalledToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _Installed);
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => runtimeType.hashCode;

@override
String toString() {
  return 'CompilerStatus.installed()';
}


}




/// @nodoc
@JsonSerializable()

class _Installing implements CompilerStatus {
  const _Installing({final  String? $type}): $type = $type ?? 'installing';
  factory _Installing.fromJson(Map<String, dynamic> json) => _$InstallingFromJson(json);



@JsonKey(name: 'runtimeType')
final String $type;



@override
Map<String, dynamic> toJson() {
  return _$InstallingToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _Installing);
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => runtimeType.hashCode;

@override
String toString() {
  return 'CompilerStatus.installing()';
}


}




/// @nodoc
@JsonSerializable()

class _NotInstalled implements CompilerStatus {
  const _NotInstalled({final  String? $type}): $type = $type ?? 'notInstalled';
  factory _NotInstalled.fromJson(Map<String, dynamic> json) => _$NotInstalledFromJson(json);



@JsonKey(name: 'runtimeType')
final String $type;



@override
Map<String, dynamic> toJson() {
  return _$NotInstalledToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _NotInstalled);
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => runtimeType.hashCode;

@override
String toString() {
  return 'CompilerStatus.notInstalled()';
}


}




// dart format on
