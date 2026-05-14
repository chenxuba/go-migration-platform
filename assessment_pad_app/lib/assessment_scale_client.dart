import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

const String defaultAssessmentEducationApiBaseUrl = String.fromEnvironment(
  'EDUCATION_API_BASE_URL',
  defaultValue: 'http://127.0.0.1:8083',
);
const String defaultAssessmentScaleCategoriesPath = String.fromEnvironment(
  'ASSESSMENT_SCALE_CATEGORIES_PATH',
  defaultValue: '/api/v1/assessments/scales/categories',
);
const String defaultAssessmentScaleLibraryPath = String.fromEnvironment(
  'ASSESSMENT_SCALE_LIBRARY_PATH',
  defaultValue: '/api/v1/assessments/scales/library',
);
const String defaultAssessmentDraftsPagePath = String.fromEnvironment(
  'ASSESSMENT_DRAFTS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/pep3/drafts/page',
);
const String defaultAssessmentStudentCandidatesPath = String.fromEnvironment(
  'ASSESSMENT_STUDENT_CANDIDATES_PATH',
  defaultValue: '/api/v1/assessments/scales/student-candidates',
);

class AssessmentStudentStatuses {
  const AssessmentStudentStatuses._();

  static const int intention = 0;
  static const int enrolled = 1;
  static const int history = 2;
}

class AssessmentScaleApiException implements Exception {
  const AssessmentScaleApiException(this.message, {this.unauthorized = false});

  final String message;
  final bool unauthorized;

  @override
  String toString() => message;
}

class AssessmentScaleLibrary {
  const AssessmentScaleLibrary({
    required this.items,
    required this.summary,
    required this.filterOptions,
  });

  factory AssessmentScaleLibrary.fromJson(Map<String, dynamic> json) {
    return AssessmentScaleLibrary(
      items: _listFrom(json['items'])
          .map(
              (Map<String, dynamic> item) => AssessmentScaleItem.fromJson(item))
          .toList(),
      summary: AssessmentScaleLibrarySummary.fromJson(
        _mapFrom(json['summary']),
      ),
      filterOptions: AssessmentScaleFilterOptions.fromJson(
        _mapFrom(json['filterOptions']),
      ),
    );
  }

  static const AssessmentScaleLibrary empty = AssessmentScaleLibrary(
    items: <AssessmentScaleItem>[],
    summary: AssessmentScaleLibrarySummary(),
    filterOptions: AssessmentScaleFilterOptions(),
  );

  final List<AssessmentScaleItem> items;
  final AssessmentScaleLibrarySummary summary;
  final AssessmentScaleFilterOptions filterOptions;
}

class AssessmentScaleLibrarySummary {
  const AssessmentScaleLibrarySummary({
    this.total = 0,
    this.available = 0,
    this.unavailable = 0,
    this.monthUsage = 0,
    this.usageCount = 0,
    this.reservedAuths = 0,
  });

  factory AssessmentScaleLibrarySummary.fromJson(Map<String, dynamic> json) {
    return AssessmentScaleLibrarySummary(
      total: _intFrom(json['total']),
      available: _intFrom(json['available']),
      unavailable: _intFrom(json['unavailable']),
      monthUsage: _intFrom(json['monthUsage']),
      usageCount: _intFrom(json['usageCount']),
      reservedAuths: _intFrom(json['reservedAuths']),
    );
  }

  final int total;
  final int available;
  final int unavailable;
  final int monthUsage;
  final int usageCount;
  final int reservedAuths;
}

class AssessmentScaleFilterOptions {
  const AssessmentScaleFilterOptions({
    this.categories = const <String>[],
    this.categoryCounts = const <String, int>{},
    this.scenarios = const <String>[],
    this.statuses = const <String>[],
  });

  factory AssessmentScaleFilterOptions.fromJson(Map<String, dynamic> json) {
    return AssessmentScaleFilterOptions(
      categories: _stringListFrom(json['categories']),
      categoryCounts: _stringIntMapFrom(json['categoryCounts']),
      scenarios: _stringListFrom(json['scenarios']),
      statuses: _stringListFrom(json['statuses']),
    );
  }

  final List<String> categories;
  final Map<String, int> categoryCounts;
  final List<String> scenarios;
  final List<String> statuses;
}

class AssessmentScaleItem {
  const AssessmentScaleItem({
    required this.id,
    required this.name,
    required this.code,
    required this.category,
    required this.scenario,
    required this.ageRange,
    required this.ageMinMonths,
    required this.ageMaxMonths,
    required this.duration,
    required this.durationMinMinutes,
    required this.durationMaxMinutes,
    required this.currentVersion,
    required this.itemCount,
    required this.domainCount,
    required this.monthUsage,
    required this.usageCount,
    required this.latestUse,
    required this.dataStatus,
    required this.status,
    required this.statusText,
    required this.updatedAt,
    required this.summary,
    required this.posterUrl,
    required this.executionEntry,
    required this.apiPackage,
  });

  factory AssessmentScaleItem.fromJson(Map<String, dynamic> json) {
    return AssessmentScaleItem(
      id: _intFrom(json['id']),
      name: '${json['name'] ?? ''}',
      code: '${json['code'] ?? ''}',
      category: '${json['category'] ?? ''}',
      scenario: '${json['scenario'] ?? ''}',
      ageRange: '${json['ageRange'] ?? ''}',
      ageMinMonths: _intFrom(json['ageMinMonths']),
      ageMaxMonths: _intFrom(json['ageMaxMonths']),
      duration: '${json['duration'] ?? ''}',
      durationMinMinutes: _intFrom(json['durationMinMinutes']),
      durationMaxMinutes: _intFrom(json['durationMaxMinutes']),
      currentVersion: '${json['currentVersion'] ?? ''}',
      itemCount: _intFrom(json['itemCount']),
      domainCount: _intFrom(json['domainCount']),
      monthUsage: _intFrom(json['monthUsage']),
      usageCount: _intFrom(json['usageCount']),
      latestUse: '${json['latestUse'] ?? ''}',
      dataStatus: '${json['dataStatus'] ?? ''}',
      status: '${json['status'] ?? ''}',
      statusText: '${json['statusText'] ?? ''}',
      updatedAt: '${json['updatedAt'] ?? ''}',
      summary: '${json['summary'] ?? ''}',
      posterUrl: '${json['posterUrl'] ?? ''}',
      executionEntry: '${json['executionEntry'] ?? ''}',
      apiPackage: '${json['apiPackage'] ?? ''}',
    );
  }

  final int id;
  final String name;
  final String code;
  final String category;
  final String scenario;
  final String ageRange;
  final int ageMinMonths;
  final int ageMaxMonths;
  final String duration;
  final int durationMinMinutes;
  final int durationMaxMinutes;
  final String currentVersion;
  final int itemCount;
  final int domainCount;
  final int monthUsage;
  final int usageCount;
  final String latestUse;
  final String dataStatus;
  final String status;
  final String statusText;
  final String updatedAt;
  final String summary;
  final String posterUrl;
  final String executionEntry;
  final String apiPackage;

  bool get available => status == 'available';

  List<String> get tags {
    final List<String> values = <String>[];
    if (itemCount > 0) {
      values.add('$itemCount题');
    }
    final String durationLabel = displayDuration;
    if (durationLabel.isNotEmpty) {
      values.add(durationLabel);
    }
    if (ageRange.trim().isNotEmpty) {
      values.add(ageRange.trim());
    }
    if (values.isEmpty && scenario.trim().isNotEmpty) {
      values.add(scenario.trim());
    }
    if (values.isEmpty && currentVersion.trim().isNotEmpty) {
      values.add(currentVersion.trim());
    }
    return values.take(3).toList();
  }

  String get displayDuration {
    if (duration.trim().isNotEmpty) {
      return duration.trim();
    }
    if (durationMinMinutes <= 0 && durationMaxMinutes <= 0) {
      return '';
    }
    if (durationMinMinutes > 0 &&
        durationMaxMinutes > 0 &&
        durationMinMinutes != durationMaxMinutes) {
      return '$durationMinMinutes-$durationMaxMinutes分钟';
    }
    final int minutes =
        durationMaxMinutes > 0 ? durationMaxMinutes : durationMinMinutes;
    return '$minutes分钟';
  }
}

class AssessmentDraftPage {
  const AssessmentDraftPage({
    required this.items,
    required this.total,
    required this.current,
    required this.size,
  });

  factory AssessmentDraftPage.fromJson(Map<String, dynamic> json) {
    return AssessmentDraftPage(
      items: _listFrom(json['items'])
          .map((Map<String, dynamic> item) =>
              AssessmentDraftSummary.fromJson(item))
          .toList(),
      total: _intFrom(json['total']),
      current: _intFrom(json['current']),
      size: _intFrom(json['size']),
    );
  }

  static const AssessmentDraftPage empty = AssessmentDraftPage(
    items: <AssessmentDraftSummary>[],
    total: 0,
    current: 1,
    size: 0,
  );

  final List<AssessmentDraftSummary> items;
  final int total;
  final int current;
  final int size;
}

class AssessmentDraftSummary {
  const AssessmentDraftSummary({
    required this.id,
    required this.studentName,
    required this.assessmentCode,
    required this.assessmentName,
    required this.scaleVersion,
    required this.examinerName,
    required this.status,
    required this.answeredItemCount,
    required this.rawScoreCount,
    required this.completionPercent,
    required this.progressItemCount,
    required this.progressQuestionDisplayPreference,
    required this.createdTime,
    required this.updatedTime,
  });

  factory AssessmentDraftSummary.fromJson(Map<String, dynamic> json) {
    final Map<String, dynamic> progress = _mapFrom(json['progress']);
    return AssessmentDraftSummary(
      id: _intFrom(json['id']),
      studentName: '${json['studentName'] ?? ''}',
      assessmentCode: '${json['assessmentCode'] ?? ''}',
      assessmentName: '${json['assessmentName'] ?? ''}',
      scaleVersion: '${json['scaleVersion'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      status: '${json['status'] ?? ''}',
      answeredItemCount: _intFrom(json['answeredItemCount']),
      rawScoreCount: _intFrom(json['rawScoreCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      progressItemCount: _intFrom(progress['itemCount']),
      progressQuestionDisplayPreference:
          '${progress['questionDisplayPreference'] ?? ''}',
      createdTime: '${json['createdTime'] ?? ''}',
      updatedTime: '${json['updatedTime'] ?? ''}',
    );
  }

  final int id;
  final String studentName;
  final String assessmentCode;
  final String assessmentName;
  final String scaleVersion;
  final String examinerName;
  final String status;
  final int answeredItemCount;
  final int rawScoreCount;
  final double completionPercent;
  final int progressItemCount;
  final String progressQuestionDisplayPreference;
  final String createdTime;
  final String updatedTime;

  int get completionPercentInt {
    final double normalized =
        completionPercent <= 1 ? completionPercent * 100 : completionPercent;
    return normalized.round().clamp(0, 100);
  }

  String get displayUpdatedDate {
    final String raw =
        updatedTime.trim().isNotEmpty ? updatedTime.trim() : createdTime.trim();
    if (raw.length >= 10) {
      return raw.substring(0, 10);
    }
    return raw;
  }
}

class AssessmentStudentCandidatePage {
  const AssessmentStudentCandidatePage({
    required this.items,
    required this.total,
    required this.current,
    required this.size,
  });

  factory AssessmentStudentCandidatePage.fromJson(Map<String, dynamic> json) {
    return AssessmentStudentCandidatePage(
      items: _listFrom(json['items'])
          .map((Map<String, dynamic> item) =>
              AssessmentStudentCandidate.fromJson(item))
          .toList(),
      total: _intFrom(json['total']),
      current: _intFrom(json['current']),
      size: _intFrom(json['size']),
    );
  }

  static const AssessmentStudentCandidatePage empty =
      AssessmentStudentCandidatePage(
    items: <AssessmentStudentCandidate>[],
    total: 0,
    current: 1,
    size: 0,
  );

  final List<AssessmentStudentCandidate> items;
  final int total;
  final int current;
  final int size;
}

class AssessmentStudentCandidate {
  const AssessmentStudentCandidate({
    required this.id,
    required this.shortName,
    required this.name,
    required this.avatarUrl,
    required this.gender,
    required this.age,
    required this.birthDate,
    required this.contactPhone,
    required this.latestAssessment,
    this.studentStatus = AssessmentStudentStatuses.enrolled,
    this.studentStatusText = '',
  });

  factory AssessmentStudentCandidate.fromJson(Map<String, dynamic> json) {
    return AssessmentStudentCandidate(
      id: _intFrom(json['id']),
      shortName: '${json['shortName'] ?? ''}',
      name: '${json['name'] ?? ''}',
      avatarUrl: '${json['avatarUrl'] ?? ''}',
      gender: '${json['gender'] ?? ''}',
      age: '${json['age'] ?? ''}',
      birthDate: '${json['birthDate'] ?? ''}',
      contactPhone: '${json['contactPhone'] ?? ''}',
      latestAssessment: '${json['latestAssessment'] ?? ''}',
      studentStatus: json.containsKey('studentStatus')
          ? _intFrom(json['studentStatus'])
          : AssessmentStudentStatuses.enrolled,
      studentStatusText: '${json['studentStatusText'] ?? ''}',
    );
  }

  final int id;
  final String shortName;
  final String name;
  final String avatarUrl;
  final String gender;
  final String age;
  final String birthDate;
  final String contactPhone;
  final String latestAssessment;
  final int studentStatus;
  final String studentStatusText;

  String get displayName => name.trim().isNotEmpty ? name.trim() : '未命名学员';

  String get displayShortName {
    if (shortName.trim().isNotEmpty) {
      return shortName.trim();
    }
    final List<int> runes = displayName.runes.toList();
    if (runes.isEmpty) {
      return '学';
    }
    return String.fromCharCode(runes.first);
  }
}

abstract interface class AssessmentScaleClient {
  Future<List<String>> fetchCategories(String token);

  Future<AssessmentScaleLibrary> fetchScaleLibrary(
    String token, {
    String keyword = '',
    String category = '',
  });

  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    bool latestOnly = false,
  });

  Future<AssessmentStudentCandidatePage> fetchStudentCandidates(
    String token, {
    String scaleCode = '',
    String keyword = '',
    int studentStatus = AssessmentStudentStatuses.enrolled,
    int pageIndex = 1,
    int pageSize = 20,
  });
}

class ApiAssessmentScaleClient implements AssessmentScaleClient {
  const ApiAssessmentScaleClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.categoriesPath = defaultAssessmentScaleCategoriesPath,
    this.libraryPath = defaultAssessmentScaleLibraryPath,
    this.draftsPagePath = defaultAssessmentDraftsPagePath,
    this.studentCandidatesPath = defaultAssessmentStudentCandidatesPath,
  });

  final String educationBaseUrl;
  final String categoriesPath;
  final String libraryPath;
  final String draftsPagePath;
  final String studentCandidatesPath;

  @override
  Future<List<String>> fetchCategories(String token) async {
    final Object? data = await _getJson(
      _uri(educationBaseUrl, categoriesPath),
      token,
    );
    return _stringListFrom(data);
  }

  @override
  Future<AssessmentScaleLibrary> fetchScaleLibrary(
    String token, {
    String keyword = '',
    String category = '',
  }) async {
    final Map<String, String> query = <String, String>{};
    if (keyword.trim().isNotEmpty) {
      query['keyword'] = keyword.trim();
    }
    if (category.trim().isNotEmpty) {
      query['category'] = category.trim();
    }
    final Uri uri = _uri(educationBaseUrl, libraryPath).replace(
      queryParameters: query.isEmpty ? null : query,
    );
    final Object? data = await _getJson(uri, token);
    if (data is! Map) {
      return AssessmentScaleLibrary.empty;
    }
    return AssessmentScaleLibrary.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    bool latestOnly = false,
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, draftsPagePath),
      token,
      <String, dynamic>{
        'pageRequestModel': <String, int>{
          'pageIndex': pageIndex,
          'pageSize': pageSize,
        },
        'queryModel': <String, dynamic>{
          if (latestOnly) 'latestOnly': true,
        },
        if (latestOnly) 'latestOnly': true,
      },
    );
    if (data is! Map) {
      return AssessmentDraftPage.empty;
    }
    return AssessmentDraftPage.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AssessmentStudentCandidatePage> fetchStudentCandidates(
    String token, {
    String scaleCode = '',
    String keyword = '',
    int studentStatus = AssessmentStudentStatuses.enrolled,
    int pageIndex = 1,
    int pageSize = 20,
  }) async {
    final Map<String, String> query = <String, String>{
      'pageIndex': '$pageIndex',
      'pageSize': '$pageSize',
    };
    if (scaleCode.trim().isNotEmpty) {
      query['scaleCode'] = scaleCode.trim();
    }
    if (keyword.trim().isNotEmpty) {
      query['keyword'] = keyword.trim();
    }
    query['studentStatus'] = '$studentStatus';
    final Object? data = await _getJson(
      _uri(educationBaseUrl, studentCandidatesPath).replace(
        queryParameters: query,
      ),
      token,
    );
    if (data is! Map) {
      return AssessmentStudentCandidatePage.empty;
    }
    return AssessmentStudentCandidatePage.fromJson(
      Map<String, dynamic>.from(data),
    );
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
      throw const AssessmentScaleApiException('量表接口响应超时，请检查网络');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接量表接口：$error');
    }
    return _handleResponse(response, fallbackMessage: '量表数据加载失败');
  }

  Future<Object?> _postJson(
    Uri uri,
    String token,
    Map<String, dynamic> body,
  ) async {
    final http.Response response;
    try {
      response = await http
          .post(
            uri,
            headers: <String, String>{
              'Accept': 'application/json',
              'Content-Type': 'application/json',
              'Authorization': 'Bearer $token',
            },
            body: jsonEncode(body),
          )
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const AssessmentScaleApiException('草稿接口响应超时，请检查网络');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接草稿接口：$error');
    }
    return _handleResponse(response, fallbackMessage: '草稿数据加载失败');
  }

  Future<Object?> _handleResponse(
    http.Response response, {
    required String fallbackMessage,
  }) async {
    final Object? decoded = await _decodeResponse(response.body);
    if (response.statusCode == 401) {
      throw AssessmentScaleApiException(
        _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw AssessmentScaleApiException(
        _messageFromPayload(decoded) ?? fallbackMessage,
      );
    }
    if (decoded is Map) {
      final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
      if (envelope['success'] == false) {
        throw AssessmentScaleApiException(
          _messageFromPayload(envelope) ?? fallbackMessage,
        );
      }
      if (envelope.containsKey('data')) {
        return envelope['data'];
      }
    }
    return decoded;
  }
}

const int _assessmentScaleBackgroundDecodeThreshold = 24 * 1024;

Uri _uri(String baseUrl, String path) {
  final String trimmedBase = baseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
  final String normalizedPath = path.startsWith('/') ? path : '/$path';
  return Uri.parse('$trimmedBase$normalizedPath');
}

Future<Object?> _decodeResponse(String body) async {
  if (body.trim().isEmpty) {
    return null;
  }
  try {
    if (body.length >= _assessmentScaleBackgroundDecodeThreshold) {
      return await compute(_decodeJsonPayload, body);
    }
    return _decodeJsonPayload(body);
  } on FormatException {
    return body;
  }
}

Object? _decodeJsonPayload(String body) => jsonDecode(body);

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

Map<String, dynamic> _mapFrom(Object? raw) {
  if (raw is Map) {
    return Map<String, dynamic>.from(raw);
  }
  return <String, dynamic>{};
}

List<Map<String, dynamic>> _listFrom(Object? raw) {
  if (raw is! List) {
    return <Map<String, dynamic>>[];
  }
  return raw
      .whereType<Map>()
      .map((Map item) => Map<String, dynamic>.from(item))
      .toList();
}

List<String> _stringListFrom(Object? raw) {
  if (raw is List) {
    return raw
        .map((Object? item) => '${item ?? ''}'.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
  }
  if (raw is String) {
    return raw
        .split(',')
        .map((String item) => item.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
  }
  return const <String>[];
}

Map<String, int> _stringIntMapFrom(Object? raw) {
  if (raw is! Map) {
    return const <String, int>{};
  }
  final Map<String, int> out = <String, int>{};
  raw.forEach((Object? key, Object? value) {
    final String textKey = '${key ?? ''}'.trim();
    if (textKey.isNotEmpty) {
      out[textKey] = _intFrom(value);
    }
  });
  return out;
}

int _intFrom(Object? raw) {
  if (raw is int) {
    return raw;
  }
  if (raw is num) {
    return raw.toInt();
  }
  return int.tryParse('${raw ?? ''}') ?? 0;
}

double _doubleFrom(Object? raw) {
  if (raw is num) {
    return raw.toDouble();
  }
  return double.tryParse('${raw ?? ''}') ?? 0;
}
