import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'assessment_scale_client.dart';

const String defaultIepPep3RecordsPagePath = String.fromEnvironment(
  'IEP_PEP3_RECORDS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/page',
);
const String defaultIepErxinRecordsPagePath = String.fromEnvironment(
  'IEP_ERXIN_RECORDS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/page',
);

class IepAssessmentRecordApiException implements Exception {
  const IepAssessmentRecordApiException(
    this.message, {
    this.unauthorized = false,
  });

  final String message;
  final bool unauthorized;

  @override
  String toString() => message;
}

class IepAssessmentRecordPage {
  const IepAssessmentRecordPage({
    required this.items,
    required this.total,
    required this.current,
    required this.size,
  });

  static const IepAssessmentRecordPage empty = IepAssessmentRecordPage(
    items: <IepAssessmentRecordSummary>[],
    total: 0,
    current: 1,
    size: 0,
  );

  final List<IepAssessmentRecordSummary> items;
  final int total;
  final int current;
  final int size;
}

class IepAssessmentRecordSummary {
  const IepAssessmentRecordSummary({
    required this.id,
    required this.source,
    required this.studentId,
    required this.studentName,
    required this.assessmentCode,
    required this.assessmentName,
    required this.birthDate,
    required this.assessmentDate,
    required this.examinerName,
    required this.updatedTime,
    this.studentGender = '',
    this.scaleCategory = '',
    this.ageYears = 0,
    this.ageMonths = 0,
    this.ageDays = 0,
    this.iepPlanStatus = '',
  });

  factory IepAssessmentRecordSummary.fromJson(
    Map<String, dynamic> json, {
    required String source,
  }) {
    return IepAssessmentRecordSummary(
      id: _intFrom(json['id']),
      source: source,
      studentId: _intFrom(json['studentId']),
      studentName: _stringFrom(json['studentName']),
      studentGender: _stringFrom(json['studentGender']),
      assessmentCode: _stringFrom(json['assessmentCode']),
      assessmentName: _stringFrom(json['assessmentName']),
      scaleCategory: _stringFrom(json['scaleCategory']),
      birthDate: _dateStringFrom(json['birthDate']),
      assessmentDate: _dateStringFrom(json['assessmentDate']),
      ageYears: _intFrom(json['ageYears']),
      ageMonths: _intFrom(json['ageMonths']),
      ageDays: _intFrom(json['ageDays']),
      examinerName: _stringFrom(json['examinerName']),
      iepPlanStatus: _stringFrom(json['iepPlanStatus']),
      updatedTime: _stringFrom(json['updatedTime']),
    );
  }

  final int id;
  final String source;
  final int studentId;
  final String studentName;
  final String studentGender;
  final String assessmentCode;
  final String assessmentName;
  final String scaleCategory;
  final String birthDate;
  final String assessmentDate;
  final int ageYears;
  final int ageMonths;
  final int ageDays;
  final String examinerName;
  final String iepPlanStatus;
  final String updatedTime;
}

abstract interface class IepAssessmentRecordClient {
  Future<IepAssessmentRecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 20,
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  });
}

class ApiIepAssessmentRecordClient implements IepAssessmentRecordClient {
  const ApiIepAssessmentRecordClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.pep3RecordsPagePath = defaultIepPep3RecordsPagePath,
    this.erxinRecordsPagePath = defaultIepErxinRecordsPagePath,
    this.httpClient,
  });

  final String educationBaseUrl;
  final String pep3RecordsPagePath;
  final String erxinRecordsPagePath;
  final http.Client? httpClient;

  @override
  Future<IepAssessmentRecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 20,
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  }) async {
    final int normalizedPage = pageIndex < 1 ? 1 : pageIndex;
    final int normalizedSize = pageSize < 1 ? 20 : pageSize;
    final int sourcePageSize = normalizedPage * normalizedSize;
    final Map<String, dynamic> payload = <String, dynamic>{
      'pageRequestModel': <String, int>{
        'pageIndex': 1,
        'pageSize': sourcePageSize,
      },
      'queryModel': <String, dynamic>{
        if (searchKey.trim().isNotEmpty) 'searchKey': searchKey.trim(),
        if (assessmentDateBegin.trim().isNotEmpty)
          'assessmentDateBegin': assessmentDateBegin.trim(),
        if (assessmentDateEnd.trim().isNotEmpty)
          'assessmentDateEnd': assessmentDateEnd.trim(),
      },
    };

    final List<_SourceRecordPage> pages =
        await Future.wait(<Future<_SourceRecordPage>>[
      _fetchSourcePage(
        token,
        path: pep3RecordsPagePath,
        source: 'PEP3',
        payload: payload,
      ),
      _fetchSourcePage(
        token,
        path: erxinRecordsPagePath,
        source: 'ERXIN',
        payload: payload,
      ),
    ]);
    final List<IepAssessmentRecordSummary> merged =
        pages.expand((page) => page.items).toList()..sort(_compareRecordDesc);
    final int start = (normalizedPage - 1) * normalizedSize;
    final List<IepAssessmentRecordSummary> visible = start >= merged.length
        ? <IepAssessmentRecordSummary>[]
        : merged.skip(start).take(normalizedSize).toList(growable: false);
    return IepAssessmentRecordPage(
      items: visible,
      total: pages.fold<int>(0, (int sum, _SourceRecordPage page) {
        return sum + page.total;
      }),
      current: normalizedPage,
      size: normalizedSize,
    );
  }

  Future<_SourceRecordPage> _fetchSourcePage(
    String token, {
    required String path,
    required String source,
    required Map<String, dynamic> payload,
  }) async {
    final http.Client client = httpClient ?? http.Client();
    final bool shouldCloseClient = httpClient == null;
    final http.Response response;
    try {
      response = await client
          .post(
            _uri(path),
            headers: _headers(token),
            body: jsonEncode(payload),
          )
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const IepAssessmentRecordApiException('评估记录接口响应超时，请检查网络');
    } on Object catch (error) {
      throw IepAssessmentRecordApiException('无法连接评估记录接口：$error');
    } finally {
      if (shouldCloseClient) {
        client.close();
      }
    }
    final Object? data = _handleResponse(response);
    if (data is! Map) {
      return const _SourceRecordPage(
        items: <IepAssessmentRecordSummary>[],
        total: 0,
      );
    }
    return _SourceRecordPage.fromJson(
      Map<String, dynamic>.from(data),
      source: source,
    );
  }

  Uri _uri(String path) {
    final String trimmedBase =
        educationBaseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
    final String normalizedPath = path.startsWith('/') ? path : '/$path';
    return Uri.parse('$trimmedBase$normalizedPath');
  }

  Map<String, String> _headers(String token) {
    return <String, String>{
      'Accept': 'application/json',
      'Content-Type': 'application/json; charset=utf-8',
      if (token.trim().isNotEmpty) 'Authorization': 'Bearer ${token.trim()}',
      if (token.trim().isNotEmpty) 'X-Access-Token': token.trim(),
    };
  }

  Object? _handleResponse(http.Response response) {
    final Object? decoded = response.body.trim().isEmpty
        ? null
        : jsonDecode(utf8.decode(response.bodyBytes));
    if (response.statusCode == 401 || response.statusCode == 403) {
      throw IepAssessmentRecordApiException(
        _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw IepAssessmentRecordApiException(
        _messageFromPayload(decoded) ?? '评估记录加载失败',
      );
    }
    if (decoded is Map) {
      final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
      if (envelope['success'] == false) {
        throw IepAssessmentRecordApiException(
          _messageFromPayload(envelope) ?? '评估记录加载失败',
        );
      }
      if (envelope.containsKey('data')) {
        return envelope['data'];
      }
    }
    return decoded;
  }
}

class _SourceRecordPage {
  const _SourceRecordPage({
    required this.items,
    required this.total,
  });

  factory _SourceRecordPage.fromJson(
    Map<String, dynamic> json, {
    required String source,
  }) {
    return _SourceRecordPage(
      items: _listFrom(json['items'])
          .map(
            (Map<String, dynamic> item) =>
                IepAssessmentRecordSummary.fromJson(item, source: source),
          )
          .toList(),
      total: _intFrom(json['total']),
    );
  }

  final List<IepAssessmentRecordSummary> items;
  final int total;
}

int _compareRecordDesc(
  IepAssessmentRecordSummary left,
  IepAssessmentRecordSummary right,
) {
  final int dateCompare = _dateSortValue(right.assessmentDate)
      .compareTo(_dateSortValue(left.assessmentDate));
  if (dateCompare != 0) {
    return dateCompare;
  }
  return right.id.compareTo(left.id);
}

int _dateSortValue(String value) {
  final DateTime? parsed = DateTime.tryParse(value.trim());
  if (parsed == null) {
    return 0;
  }
  return parsed.millisecondsSinceEpoch;
}

List<Map<String, dynamic>> _listFrom(Object? value) {
  if (value is! List) {
    return <Map<String, dynamic>>[];
  }
  return value
      .whereType<Map>()
      .map((Map<dynamic, dynamic> item) => Map<String, dynamic>.from(item))
      .toList();
}

int _intFrom(Object? value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  return int.tryParse('${value ?? ''}') ?? 0;
}

String _stringFrom(Object? value) => '${value ?? ''}'.trim();

String _dateStringFrom(Object? value) {
  final String raw = _stringFrom(value);
  if (raw.isEmpty) {
    return '';
  }
  final DateTime? parsed = DateTime.tryParse(raw);
  if (parsed == null) {
    return raw.contains('T') ? raw.split('T').first : raw;
  }
  return '${parsed.year.toString().padLeft(4, '0')}-'
      '${parsed.month.toString().padLeft(2, '0')}-'
      '${parsed.day.toString().padLeft(2, '0')}';
}

String? _messageFromPayload(Object? payload) {
  if (payload is Map) {
    for (final String key in <String>['message', 'msg', 'error']) {
      final Object? value = payload[key];
      if (value != null && '$value'.trim().isNotEmpty) {
        return '$value';
      }
    }
  }
  return null;
}
