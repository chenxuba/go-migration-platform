import 'dart:convert';

import 'package:http/http.dart' as http;

import 'assessment_scale_client.dart';

const String defaultErxinTemplateSummaryPath = String.fromEnvironment(
  'ERXIN_TEMPLATE_SUMMARY_PATH',
  defaultValue: '/api/v1/assessments/erxin/form-template/summary',
);
const String defaultErxinTemplateItemPath = String.fromEnvironment(
  'ERXIN_TEMPLATE_ITEM_PATH',
  defaultValue: '/api/v1/assessments/erxin/form-template/item',
);

class ErxinAssessmentLaunchArgs {
  const ErxinAssessmentLaunchArgs({
    this.studentId = 0,
    this.studentName = '',
    this.studentAge = '',
    this.birthDate = '',
    this.assessmentDate = '',
    this.examinerName = '',
    this.scaleName = '儿心量表-II',
  });

  final int studentId;
  final String studentName;
  final String studentAge;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final String scaleName;
}

abstract class ErxinAssessmentClient {
  const ErxinAssessmentClient();

  Future<ErxinTemplateSummary> fetchTemplateSummary(String token);

  Future<ErxinAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  });
}

class ApiErxinAssessmentClient extends ErxinAssessmentClient {
  const ApiErxinAssessmentClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.templateSummaryPath = defaultErxinTemplateSummaryPath,
    this.templateItemPath = defaultErxinTemplateItemPath,
    this.httpClient,
  });

  final String educationBaseUrl;
  final String templateSummaryPath;
  final String templateItemPath;
  final http.Client? httpClient;

  @override
  Future<ErxinTemplateSummary> fetchTemplateSummary(String token) async {
    final http.Client client = httpClient ?? http.Client();
    final http.Response response = await client.get(
      _uri(educationBaseUrl, templateSummaryPath),
      headers: _authHeaders(token),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '儿心量表模板加载失败',
    );
    if (data == null) {
      return ErxinTemplateSummary.empty;
    }
    return ErxinTemplateSummary.fromJson(Map<String, dynamic>.from(data as Map));
  }

  @override
  Future<ErxinAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  }) async {
    final http.Client client = httpClient ?? http.Client();
    final Uri uri = _uri(
      educationBaseUrl,
      templateItemPath,
      <String, String>{'itemNo': '$itemNo'},
    );
    final http.Response response = await client.get(
      uri,
      headers: _authHeaders(token),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '题目说明加载失败',
    );
    if (data == null) {
      return ErxinAssessmentItem.empty;
    }
    return ErxinAssessmentItem.fromJson(Map<String, dynamic>.from(data as Map));
  }
}

class ErxinTemplateSummary {
  const ErxinTemplateSummary({
    required this.templateCode,
    required this.title,
    required this.scaleCode,
    required this.scaleVersion,
    required this.itemCount,
    required this.domains,
    required this.ageGroups,
  });

  factory ErxinTemplateSummary.fromJson(Map<String, dynamic> json) {
    return ErxinTemplateSummary(
      templateCode: '${json['templateCode'] ?? ''}',
      title: '${json['title'] ?? ''}',
      scaleCode: '${json['scaleCode'] ?? ''}',
      scaleVersion: '${json['scaleVersion'] ?? ''}',
      itemCount: _intFrom(json['itemCount']),
      domains: _listFrom(json['domains'])
          .map((Map<String, dynamic> item) => ErxinDomain.fromJson(item))
          .toList(),
      ageGroups: _listFrom(json['ageGroups'])
          .map((Map<String, dynamic> item) => ErxinAgeGroup.fromJson(item))
          .toList(),
    );
  }

  static const ErxinTemplateSummary empty = ErxinTemplateSummary(
    templateCode: '',
    title: '',
    scaleCode: 'ERXIN2',
    scaleVersion: '',
    itemCount: 0,
    domains: <ErxinDomain>[],
    ageGroups: <ErxinAgeGroup>[],
  );

  final String templateCode;
  final String title;
  final String scaleCode;
  final String scaleVersion;
  final int itemCount;
  final List<ErxinDomain> domains;
  final List<ErxinAgeGroup> ageGroups;
}

class ErxinDomain {
  const ErxinDomain({
    required this.domainCode,
    required this.domainName,
    required this.sortNo,
  });

  factory ErxinDomain.fromJson(Map<String, dynamic> json) {
    return ErxinDomain(
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      sortNo: _intFrom(json['sortNo']),
    );
  }

  final String domainCode;
  final String domainName;
  final int sortNo;
}

class ErxinAgeGroup {
  const ErxinAgeGroup({
    required this.ageMonth,
    required this.title,
    required this.items,
  });

  factory ErxinAgeGroup.fromJson(Map<String, dynamic> json) {
    return ErxinAgeGroup(
      ageMonth: _intFrom(json['ageMonth']),
      title: '${json['title'] ?? ''}',
      items: _listFrom(json['items'])
          .map((Map<String, dynamic> item) => ErxinItemSummary.fromJson(item))
          .toList(),
    );
  }

  final int ageMonth;
  final String title;
  final List<ErxinItemSummary> items;
}

class ErxinItemSummary {
  const ErxinItemSummary({
    required this.itemNo,
    required this.itemTitle,
    required this.testItem,
    required this.ageMonth,
    required this.domainCode,
    required this.domainName,
    required this.parentReportAllowed,
    required this.attentionIfFailed,
  });

  factory ErxinItemSummary.fromJson(Map<String, dynamic> json) {
    return ErxinItemSummary(
      itemNo: _intFrom(json['itemNo']),
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      ageMonth: _intFrom(json['ageMonth']),
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      parentReportAllowed: json['parentReportAllowed'] == true,
      attentionIfFailed: json['attentionIfFailed'] == true,
    );
  }

  final int itemNo;
  final String itemTitle;
  final String testItem;
  final int ageMonth;
  final String domainCode;
  final String domainName;
  final bool parentReportAllowed;
  final bool attentionIfFailed;
}

class ErxinAssessmentItem extends ErxinItemSummary {
  const ErxinAssessmentItem({
    required super.itemNo,
    required super.itemTitle,
    required super.testItem,
    required super.ageMonth,
    required super.domainCode,
    required super.domainName,
    required super.parentReportAllowed,
    required super.attentionIfFailed,
    required this.method,
    required this.passCriteria,
  });

  factory ErxinAssessmentItem.fromJson(Map<String, dynamic> json) {
    return ErxinAssessmentItem(
      itemNo: _intFrom(json['itemNo']),
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      ageMonth: _intFrom(json['ageMonth']),
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      parentReportAllowed: json['parentReportAllowed'] == true,
      attentionIfFailed: json['attentionIfFailed'] == true,
      method: '${json['method'] ?? ''}',
      passCriteria: '${json['passCriteria'] ?? ''}',
    );
  }

  static const ErxinAssessmentItem empty = ErxinAssessmentItem(
    itemNo: 0,
    itemTitle: '',
    testItem: '',
    ageMonth: 0,
    domainCode: '',
    domainName: '',
    parentReportAllowed: false,
    attentionIfFailed: false,
    method: '',
    passCriteria: '',
  );

  final String method;
  final String passCriteria;
}

Uri _uri(String base, String path, [Map<String, String>? query]) {
  final Uri baseUri = Uri.parse(base);
  final String normalizedPath = path.startsWith('/') ? path : '/$path';
  return baseUri.replace(
    path: normalizedPath,
    queryParameters: query == null || query.isEmpty ? null : query,
  );
}

Map<String, String> _authHeaders(String token) {
  return <String, String>{
    'Authorization': 'Bearer $token',
    'Accept': 'application/json',
  };
}

Object? _handleResponse(
  http.Response response, {
  required String fallbackMessage,
}) {
  final Object? decoded = response.body.trim().isEmpty
      ? null
      : jsonDecode(utf8.decode(response.bodyBytes));
  if (response.statusCode == 401 || response.statusCode == 403) {
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
  if (decoded is Map<String, dynamic>) {
    final bool success = decoded['success'] == true;
    if (success) {
      return decoded['data'];
    }
    throw AssessmentScaleApiException(
      _messageFromPayload(decoded) ?? fallbackMessage,
    );
  }
  return decoded;
}

List<Map<String, dynamic>> _listFrom(Object? value) {
  if (value is List) {
    return value
        .whereType<Map>()
        .map((Map item) => Map<String, dynamic>.from(item))
        .toList();
  }
  return <Map<String, dynamic>>[];
}

int _intFrom(Object? value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  if (value is String) {
    return int.tryParse(value.trim()) ?? 0;
  }
  return 0;
}

String? _messageFromPayload(Object? payload) {
  if (payload is Map<String, dynamic>) {
    final Object? message = payload['message'] ?? payload['msg'];
    if (message != null && '$message'.trim().isNotEmpty) {
      return '$message';
    }
  }
  return null;
}
