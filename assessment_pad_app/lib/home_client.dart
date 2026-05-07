import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';
import 'package:http/http.dart' as http;

import 'auth_client.dart';

const String defaultEducationApiBaseUrl = String.fromEnvironment(
  'EDUCATION_API_BASE_URL',
  defaultValue: 'http://127.0.0.1:8083',
);
const String defaultHomeSummaryPath = String.fromEnvironment(
  'HOME_SUMMARY_PATH',
  defaultValue: '/api/v1/pad/home/summary',
);
const String defaultCurrentSessionPath = String.fromEnvironment(
  'CURRENT_SESSION_PATH',
  defaultValue: '/api/v1/auth/me',
);

class HomeApiException implements Exception {
  const HomeApiException(this.message, {this.unauthorized = false});

  final String message;
  final bool unauthorized;

  @override
  String toString() => message;
}

class HomeSession {
  const HomeSession({
    this.nickName = '',
    this.orgName = '',
    this.avatar = '',
    this.username = '',
    this.mobile = '',
  });

  factory HomeSession.fromJson(Map<String, dynamic> json) {
    return HomeSession(
      nickName: '${json['nickName'] ?? ''}',
      orgName: '${json['orgName'] ?? ''}',
      avatar: '${json['avatar'] ?? ''}',
      username: '${json['username'] ?? ''}',
      mobile: '${json['mobile'] ?? ''}',
    );
  }

  static const HomeSession fallback = HomeSession();

  final String nickName;
  final String orgName;
  final String avatar;
  final String username;
  final String mobile;
}

class HomeSummary {
  const HomeSummary({
    required this.date,
    required this.weekday,
    required this.assessmentStats,
    required this.schedule,
    required this.weather,
  });

  factory HomeSummary.fromJson(Map<String, dynamic> json) {
    return HomeSummary(
      date: '${json['date'] ?? ''}',
      weekday: '${json['weekday'] ?? ''}',
      assessmentStats: HomeAssessmentStats.fromJson(
        _mapFrom(json['assessmentStats']),
      ),
      schedule: _listFrom(json['schedule'])
          .map((Map<String, dynamic> item) => HomeScheduleItem.fromJson(item))
          .toList(),
      weather: HomeWeather.fromJson(_mapFrom(json['weather'])),
    );
  }

  factory HomeSummary.fallback() {
    final DateTime now = DateTime.now();
    return HomeSummary(
      date:
          '${now.year.toString().padLeft(4, '0')}-${now.month.toString().padLeft(2, '0')}-${now.day.toString().padLeft(2, '0')}',
      weekday: _weekdayName(now.weekday),
      assessmentStats: const HomeAssessmentStats(
        enrolledStudents: 0,
        assessedStudents: 0,
        inProgressDrafts: 0,
        unassessedStudents: 0,
        completedRecords: 0,
        pendingIep: 0,
        draftIep: 0,
        generatedIep: 0,
        total: 0,
        coverageRate: 0,
      ),
      schedule: const <HomeScheduleItem>[],
      weather: const HomeWeather(
        city: '',
        condition: 'sunny',
        displayName: '',
      ),
    );
  }

  final String date;
  final String weekday;
  final HomeAssessmentStats assessmentStats;
  final List<HomeScheduleItem> schedule;
  final HomeWeather weather;
}

class HomeAssessmentStats {
  const HomeAssessmentStats({
    required this.enrolledStudents,
    required this.assessedStudents,
    required this.inProgressDrafts,
    required this.unassessedStudents,
    required this.completedRecords,
    required this.pendingIep,
    required this.draftIep,
    required this.generatedIep,
    required this.total,
    required this.coverageRate,
  });

  factory HomeAssessmentStats.fromJson(Map<String, dynamic> json) {
    return HomeAssessmentStats(
      enrolledStudents: _intFrom(json['enrolledStudents']),
      assessedStudents: _intFrom(json['assessedStudents']),
      inProgressDrafts: _intFrom(json['inProgressDrafts']),
      unassessedStudents: _intFrom(json['unassessedStudents']),
      completedRecords: _intFrom(json['completedRecords']),
      pendingIep: _intFrom(json['pendingIep']),
      draftIep: _intFrom(json['draftIep']),
      generatedIep: _intFrom(json['generatedIep']),
      total: _intFrom(json['total']),
      coverageRate: _doubleFrom(json['coverageRate'] ?? json['completionRate']),
    );
  }

  final int enrolledStudents;
  final int assessedStudents;
  final int inProgressDrafts;
  final int unassessedStudents;
  final int completedRecords;
  final int pendingIep;
  final int draftIep;
  final int generatedIep;
  final int total;
  final double coverageRate;
}

class HomeScheduleItem {
  const HomeScheduleItem({
    required this.time,
    required this.title,
    required this.place,
    required this.state,
  });

  factory HomeScheduleItem.fromJson(Map<String, dynamic> json) {
    return HomeScheduleItem(
      time: '${json['time'] ?? ''}',
      title: '${json['title'] ?? ''}',
      place: '${json['place'] ?? ''}',
      state: '${json['state'] ?? ''}',
    );
  }

  final String time;
  final String title;
  final String place;
  final String state;
}

class HomeWeather {
  const HomeWeather({
    required this.city,
    required this.condition,
    required this.displayName,
    this.temperature = 0,
    this.updatedAt = '',
    this.source = '',
  });

  factory HomeWeather.fromJson(Map<String, dynamic> json) {
    return HomeWeather(
      city: '${json['city'] ?? ''}',
      condition: '${json['condition'] ?? ''}',
      displayName: '${json['displayName'] ?? ''}',
      temperature: _doubleFrom(json['temperature']),
      updatedAt: '${json['updatedAt'] ?? ''}',
      source: '${json['source'] ?? ''}',
    );
  }

  final String city;
  final String condition;
  final String displayName;
  final double temperature;
  final String updatedAt;
  final String source;
}

abstract interface class HomeClient {
  Future<HomeSession> fetchCurrentSession(String token);

  Future<HomeSummary> fetchSummary(String token);
}

class ApiHomeClient implements HomeClient {
  const ApiHomeClient({
    this.educationBaseUrl = defaultEducationApiBaseUrl,
    this.loginBaseUrl = defaultLoginApiBaseUrl,
    this.homeSummaryPath = defaultHomeSummaryPath,
    this.currentSessionPath = defaultCurrentSessionPath,
  });

  final String educationBaseUrl;
  final String loginBaseUrl;
  final String homeSummaryPath;
  final String currentSessionPath;

  @override
  Future<HomeSession> fetchCurrentSession(String token) async {
    final Object? data = await _getJson(
      _uri(loginBaseUrl, currentSessionPath),
      token,
    );
    if (data is! Map) {
      return HomeSession.fallback;
    }
    return HomeSession.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<HomeSummary> fetchSummary(String token) async {
    final _HomeLocation? location = await _resolveCurrentHomeLocation();
    final Object? data = await _getJson(
      _withHomeLocation(_uri(educationBaseUrl, homeSummaryPath), location),
      token,
    );
    if (data is! Map) {
      throw const HomeApiException('首页接口返回格式不正确');
    }
    return HomeSummary.fromJson(Map<String, dynamic>.from(data));
  }

  Future<Object?> _getJson(Uri uri, String token) async {
    final http.Response response;
    try {
      response = await http.get(
        uri,
        headers: <String, String>{
          'Accept': 'application/json',
          'Authorization': 'Bearer $token',
        },
      ).timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const HomeApiException('首页接口响应超时，请检查网络');
    } on Object catch (error) {
      throw HomeApiException('无法连接首页接口：$error');
    }

    final Object? decoded = await _decodeResponse(response.body);
    if (response.statusCode == 401) {
      throw HomeApiException(
        _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw HomeApiException(_messageFromPayload(decoded) ?? '首页数据加载失败');
    }
    if (decoded is Map) {
      final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
      if (envelope['success'] == false) {
        throw HomeApiException(
          _messageFromPayload(envelope) ?? '首页数据加载失败',
        );
      }
      if (envelope.containsKey('data')) {
        return envelope['data'];
      }
    }
    return decoded;
  }
}

const int _homeBackgroundDecodeThreshold = 24 * 1024;

class _HomeLocation {
  const _HomeLocation({
    required this.latitude,
    required this.longitude,
  });

  final double latitude;
  final double longitude;
}

Future<_HomeLocation?> _resolveCurrentHomeLocation() async {
  try {
    final bool serviceEnabled = await Geolocator.isLocationServiceEnabled()
        .timeout(const Duration(milliseconds: 250), onTimeout: () => false);
    if (!serviceEnabled) {
      return null;
    }

    LocationPermission permission = await Geolocator.checkPermission().timeout(
        const Duration(milliseconds: 250),
        onTimeout: () => LocationPermission.denied);
    if (permission == LocationPermission.denied ||
        permission == LocationPermission.deniedForever ||
        permission == LocationPermission.unableToDetermine) {
      return null;
    }

    final Position? position = await Geolocator.getLastKnownPosition();
    if (position == null ||
        (position.latitude == 0 && position.longitude == 0)) {
      return null;
    }
    return _HomeLocation(
      latitude: position.latitude,
      longitude: position.longitude,
    );
  } on Object {
    return null;
  }
}

Uri _uri(String baseUrl, String path) {
  final String trimmedBase = baseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
  final String normalizedPath = path.startsWith('/') ? path : '/$path';
  return Uri.parse('$trimmedBase$normalizedPath');
}

Uri _withHomeLocation(Uri uri, _HomeLocation? location) {
  if (location == null) {
    return uri;
  }
  return uri.replace(
    queryParameters: <String, String>{
      ...uri.queryParameters,
      'latitude': location.latitude.toStringAsFixed(6),
      'longitude': location.longitude.toStringAsFixed(6),
    },
  );
}

Future<Object?> _decodeResponse(String body) async {
  if (body.trim().isEmpty) {
    return null;
  }
  try {
    if (body.length >= _homeBackgroundDecodeThreshold) {
      return await compute(_decodeHomeJsonPayload, body);
    }
    return _decodeHomeJsonPayload(body);
  } on FormatException {
    return body;
  }
}

Object? _decodeHomeJsonPayload(String body) => jsonDecode(body);

String? _messageFromPayload(Object? payload) {
  if (payload is Map) {
    final Object? message = payload['message'] ?? payload['msg'];
    if (message != null && '$message'.trim().isNotEmpty) {
      return '$message';
    }
  }
  if (payload is String && payload.trim().isNotEmpty) {
    return payload;
  }
  return null;
}

Map<String, dynamic> _mapFrom(Object? value) {
  if (value is Map) {
    return Map<String, dynamic>.from(value);
  }
  return <String, dynamic>{};
}

List<Map<String, dynamic>> _listFrom(Object? value) {
  if (value is! List) {
    return <Map<String, dynamic>>[];
  }
  return value
      .whereType<Map>()
      .map((Map item) => Map<String, dynamic>.from(item))
      .toList();
}

int _intFrom(Object? value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  return int.tryParse('$value') ?? 0;
}

double _doubleFrom(Object? value) {
  if (value is num) {
    return value.toDouble();
  }
  return double.tryParse('$value') ?? 0;
}

String _weekdayName(int weekday) {
  switch (weekday) {
    case DateTime.monday:
      return '星期一';
    case DateTime.tuesday:
      return '星期二';
    case DateTime.wednesday:
      return '星期三';
    case DateTime.thursday:
      return '星期四';
    case DateTime.friday:
      return '星期五';
    case DateTime.saturday:
      return '星期六';
    default:
      return '星期日';
  }
}
