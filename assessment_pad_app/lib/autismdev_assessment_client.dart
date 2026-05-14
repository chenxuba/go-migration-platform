import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'assessment_scale_client.dart';

const String defaultAutismDevTemplateSummaryPath = String.fromEnvironment(
  'AUTISMDEV_TEMPLATE_SUMMARY_PATH',
  defaultValue: '/api/v1/assessments/autismdev/form-template/summary',
);
const String defaultAutismDevTemplateItemPath = String.fromEnvironment(
  'AUTISMDEV_TEMPLATE_ITEM_PATH',
  defaultValue: '/api/v1/assessments/autismdev/form-template/item',
);
const String defaultAutismDevDraftSavePath = String.fromEnvironment(
  'AUTISMDEV_DRAFT_SAVE_PATH',
  defaultValue: '/api/v1/assessments/autismdev/drafts/save',
);
const String defaultAutismDevDraftItemSavePath = String.fromEnvironment(
  'AUTISMDEV_DRAFT_ITEM_SAVE_PATH',
  defaultValue: '/api/v1/assessments/autismdev/drafts/item/save',
);
const String defaultAutismDevDraftDetailPath = String.fromEnvironment(
  'AUTISMDEV_DRAFT_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/autismdev/drafts/detail',
);
const String defaultAutismDevDraftsPagePath = String.fromEnvironment(
  'AUTISMDEV_DRAFTS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/autismdev/drafts/page',
);
const String defaultAutismDevDraftSubmitPath = String.fromEnvironment(
  'AUTISMDEV_DRAFT_SUBMIT_PATH',
  defaultValue: '/api/v1/assessments/autismdev/drafts/submit',
);

class AutismDevAssessmentLaunchArgs {
  const AutismDevAssessmentLaunchArgs({
    this.draftId = 0,
    this.studentId = 0,
    this.studentName = '',
    this.studentAge = '',
    this.birthDate = '',
    this.assessmentDate = '',
    this.examinerName = '',
    this.scaleName = '孤独症儿童发展评估表',
  });

  final int draftId;
  final int studentId;
  final String studentName;
  final String studentAge;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final String scaleName;
}

abstract class AutismDevAssessmentClient {
  const AutismDevAssessmentClient();

  Future<AutismDevTemplateSummary> fetchTemplateSummary(String token);

  Future<AutismDevAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  });

  Future<AutismDevDraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  );

  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  });

  Future<AutismDevDraftDetail> fetchDraftDetail(String token, int id);

  Future<AutismDevDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  );

  Future<void> submitDraft(String token, int draftId);
}

class ApiAutismDevAssessmentClient extends AutismDevAssessmentClient {
  const ApiAutismDevAssessmentClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.templateSummaryPath = defaultAutismDevTemplateSummaryPath,
    this.templateItemPath = defaultAutismDevTemplateItemPath,
    this.draftSavePath = defaultAutismDevDraftSavePath,
    this.draftItemSavePath = defaultAutismDevDraftItemSavePath,
    this.draftDetailPath = defaultAutismDevDraftDetailPath,
    this.draftsPagePath = defaultAutismDevDraftsPagePath,
    this.draftSubmitPath = defaultAutismDevDraftSubmitPath,
    this.httpClient,
  });

  final String educationBaseUrl;
  final String templateSummaryPath;
  final String templateItemPath;
  final String draftSavePath;
  final String draftItemSavePath;
  final String draftDetailPath;
  final String draftsPagePath;
  final String draftSubmitPath;
  final http.Client? httpClient;

  @override
  Future<AutismDevTemplateSummary> fetchTemplateSummary(String token) async {
    final http.Client client = httpClient ?? http.Client();
    final Object? data = await _getJson(
      client,
      _uri(educationBaseUrl, templateSummaryPath),
      token,
      fallbackMessage: '孤独症发展评估表模板加载失败',
    );
    if (data is! Map) {
      return AutismDevTemplateSummary.empty;
    }
    return AutismDevTemplateSummary.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AutismDevAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  }) async {
    final http.Client client = httpClient ?? http.Client();
    final Object? data = await _getJson(
      client,
      _uri(
        educationBaseUrl,
        templateItemPath,
        <String, String>{'itemNo': '$itemNo'},
      ),
      token,
      fallbackMessage: '题目说明加载失败',
    );
    if (data is! Map) {
      return AutismDevAssessmentItem.empty;
    }
    return AutismDevAssessmentItem.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AutismDevDraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final http.Client client = httpClient ?? http.Client();
    final Object? data = await _postJson(
      client,
      _uri(educationBaseUrl, draftSavePath),
      token,
      payload,
      fallbackMessage: '孤独症发展评估表草稿保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('草稿保存返回格式不正确');
    }
    return AutismDevDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  }) async {
    final http.Client client = httpClient ?? http.Client();
    final Object? data = await _postJson(
      client,
      _uri(educationBaseUrl, draftsPagePath),
      token,
      <String, dynamic>{
        'pageRequestModel': <String, int>{
          'pageIndex': pageIndex,
          'pageSize': pageSize,
        },
        'queryModel': <String, dynamic>{
          'assessmentCode': 'AUTISMDEV',
          if (studentId > 0) 'studentId': studentId,
          if (latestOnly) 'latestOnly': true,
        },
        if (latestOnly) 'latestOnly': true,
      },
      fallbackMessage: '孤独症发展评估表草稿列表加载失败',
    );
    if (data is! Map) {
      return AssessmentDraftPage.empty;
    }
    return AssessmentDraftPage.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AutismDevDraftDetail> fetchDraftDetail(String token, int id) async {
    final http.Client client = httpClient ?? http.Client();
    final Object? data = await _getJson(
      client,
      _uri(
        educationBaseUrl,
        draftDetailPath,
        <String, String>{'id': '$id'},
      ),
      token,
      fallbackMessage: '孤独症发展评估表草稿详情加载失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('草稿详情返回格式不正确');
    }
    return AutismDevDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AutismDevDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final http.Client client = httpClient ?? http.Client();
    final Object? data = await _postJson(
      client,
      _uri(educationBaseUrl, draftItemSavePath),
      token,
      payload,
      fallbackMessage: '孤独症发展评估表单题保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('单题保存返回格式不正确');
    }
    return AutismDevDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<void> submitDraft(String token, int draftId) async {
    final http.Client client = httpClient ?? http.Client();
    await _postJson(
      client,
      _uri(educationBaseUrl, draftSubmitPath),
      token,
      <String, int>{'id': draftId},
      fallbackMessage: '孤独症发展评估表正式记录提交失败',
    );
  }
}

class AutismDevTemplateSummary {
  const AutismDevTemplateSummary({
    required this.templateCode,
    required this.title,
    required this.scaleCode,
    required this.scaleVersion,
    required this.itemCount,
    required this.scoreOptions,
    required this.domains,
    required this.domainGroups,
  });

  factory AutismDevTemplateSummary.fromJson(Map<String, dynamic> json) {
    return AutismDevTemplateSummary(
      templateCode: '${json['templateCode'] ?? ''}',
      title: '${json['title'] ?? ''}',
      scaleCode: '${json['scaleCode'] ?? ''}',
      scaleVersion: '${json['scaleVersion'] ?? ''}',
      itemCount: _intFrom(json['itemCount']),
      scoreOptions: _listFrom(json['scoreOptions'])
          .map((Map<String, dynamic> item) =>
              AutismDevScoreOption.fromJson(item))
          .toList(),
      domains: _listFrom(json['domains'])
          .map((Map<String, dynamic> item) => AutismDevDomain.fromJson(item))
          .toList(),
      domainGroups: _listFrom(json['domainGroups'])
          .map((Map<String, dynamic> item) =>
              AutismDevDomainGroup.fromJson(item))
          .toList(),
    );
  }

  static const AutismDevTemplateSummary empty = AutismDevTemplateSummary(
    templateCode: '',
    title: '',
    scaleCode: 'AUTISMDEV',
    scaleVersion: '',
    itemCount: 0,
    scoreOptions: <AutismDevScoreOption>[],
    domains: <AutismDevDomain>[],
    domainGroups: <AutismDevDomainGroup>[],
  );

  final String templateCode;
  final String title;
  final String scaleCode;
  final String scaleVersion;
  final int itemCount;
  final List<AutismDevScoreOption> scoreOptions;
  final List<AutismDevDomain> domains;
  final List<AutismDevDomainGroup> domainGroups;
}

class AutismDevDomain {
  const AutismDevDomain({
    required this.domainCode,
    required this.domainName,
    required this.sortNo,
    required this.itemCount,
    required this.scoreType,
  });

  factory AutismDevDomain.fromJson(Map<String, dynamic> json) {
    final String domainCode = '${json['domainCode'] ?? ''}';
    return AutismDevDomain(
      domainCode: domainCode,
      domainName: _autismDevDomainDisplayName(
        '${json['domainName'] ?? ''}',
        domainCode,
      ),
      sortNo: _intFrom(json['sortNo']),
      itemCount: _intFrom(json['itemCount']),
      scoreType: '${json['scoreType'] ?? ''}',
    );
  }

  final String domainCode;
  final String domainName;
  final int sortNo;
  final int itemCount;
  final String scoreType;
}

class AutismDevDomainGroup {
  const AutismDevDomainGroup({
    required this.groupCode,
    required this.title,
    required this.domainCode,
    required this.domainName,
    required this.scoreType,
    required this.itemCount,
    required this.items,
  });

  factory AutismDevDomainGroup.fromJson(Map<String, dynamic> json) {
    final String domainCode = '${json['domainCode'] ?? ''}';
    final String domainName = _autismDevDomainDisplayName(
      '${json['domainName'] ?? ''}',
      domainCode,
    );
    return AutismDevDomainGroup(
      groupCode: '${json['groupCode'] ?? ''}',
      title: _autismDevDomainDisplayName('${json['title'] ?? ''}', domainCode),
      domainCode: domainCode,
      domainName: domainName,
      scoreType: '${json['scoreType'] ?? ''}',
      itemCount: _intFrom(json['itemCount']),
      items: _listFrom(json['items'])
          .map((Map<String, dynamic> item) =>
              AutismDevItemSummary.fromJson(item))
          .toList(),
    );
  }

  final String groupCode;
  final String title;
  final String domainCode;
  final String domainName;
  final String scoreType;
  final int itemCount;
  final List<AutismDevItemSummary> items;
}

class AutismDevItemSummary {
  const AutismDevItemSummary({
    required this.itemNo,
    required this.domainItemNo,
    required this.itemTitle,
    required this.testItem,
    required this.assessmentRange,
    required this.materials,
    required this.method,
    required this.passCriteria,
    required this.ageSegment,
    required this.ageMinMonth,
    required this.ageMaxMonth,
    required this.domainCode,
    required this.domainName,
    required this.scoreType,
    required this.assessmentModes,
  });

  factory AutismDevItemSummary.fromJson(Map<String, dynamic> json) {
    final String domainCode = '${json['domainCode'] ?? ''}';
    return AutismDevItemSummary(
      itemNo: _intFrom(json['itemNo']),
      domainItemNo: _intFrom(json['domainItemNo']),
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      assessmentRange: '${json['assessmentRange'] ?? ''}',
      materials: '${json['materials'] ?? ''}',
      method: '${json['method'] ?? ''}',
      passCriteria: '${json['passCriteria'] ?? ''}',
      ageSegment: '${json['ageSegment'] ?? ''}',
      ageMinMonth: _intFrom(json['ageMinMonth']),
      ageMaxMonth: _intFrom(json['ageMaxMonth']),
      domainCode: domainCode,
      domainName: _autismDevDomainDisplayName(
        '${json['domainName'] ?? ''}',
        domainCode,
      ),
      scoreType: '${json['scoreType'] ?? ''}',
      assessmentModes: _stringListFrom(json['assessmentModes']),
    );
  }

  final int itemNo;
  final int domainItemNo;
  final String itemTitle;
  final String testItem;
  final String assessmentRange;
  final String materials;
  final String method;
  final String passCriteria;
  final String ageSegment;
  final int ageMinMonth;
  final int ageMaxMonth;
  final String domainCode;
  final String domainName;
  final String scoreType;
  final List<String> assessmentModes;
}

class AutismDevAssessmentItem extends AutismDevItemSummary {
  const AutismDevAssessmentItem({
    required super.itemNo,
    required super.domainItemNo,
    required super.itemTitle,
    required super.testItem,
    required super.assessmentRange,
    required super.materials,
    required super.method,
    required super.passCriteria,
    required super.ageSegment,
    required super.ageMinMonth,
    required super.ageMaxMonth,
    required super.domainCode,
    required super.domainName,
    required super.scoreType,
    required super.assessmentModes,
    required this.scoreOptions,
  });

  factory AutismDevAssessmentItem.fromJson(Map<String, dynamic> json) {
    final String domainCode = '${json['domainCode'] ?? ''}';
    return AutismDevAssessmentItem(
      itemNo: _intFrom(json['itemNo']),
      domainItemNo: _intFrom(json['domainItemNo']),
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      assessmentRange: '${json['assessmentRange'] ?? ''}',
      materials: '${json['materials'] ?? ''}',
      method: '${json['method'] ?? ''}',
      passCriteria: '${json['passCriteria'] ?? ''}',
      ageSegment: '${json['ageSegment'] ?? ''}',
      ageMinMonth: _intFrom(json['ageMinMonth']),
      ageMaxMonth: _intFrom(json['ageMaxMonth']),
      domainCode: domainCode,
      domainName: _autismDevDomainDisplayName(
        '${json['domainName'] ?? ''}',
        domainCode,
      ),
      scoreType: '${json['scoreType'] ?? ''}',
      assessmentModes: _stringListFrom(json['assessmentModes']),
      scoreOptions: _listFrom(json['scoreOptions'])
          .map((Map<String, dynamic> item) =>
              AutismDevScoreOption.fromJson(item))
          .toList(),
    );
  }

  static const AutismDevAssessmentItem empty = AutismDevAssessmentItem(
    itemNo: 0,
    domainItemNo: 0,
    itemTitle: '',
    testItem: '',
    assessmentRange: '',
    materials: '',
    method: '',
    passCriteria: '',
    ageSegment: '',
    ageMinMonth: 0,
    ageMaxMonth: 0,
    domainCode: '',
    domainName: '',
    scoreType: '',
    assessmentModes: <String>[],
    scoreOptions: <AutismDevScoreOption>[],
  );

  final List<AutismDevScoreOption> scoreOptions;
}

class AutismDevScoreOption {
  const AutismDevScoreOption({
    required this.value,
    required this.label,
    required this.description,
    required this.scoreType,
  });

  factory AutismDevScoreOption.fromJson(Map<String, dynamic> json) {
    return AutismDevScoreOption(
      value: '${json['value'] ?? ''}',
      label: '${json['label'] ?? ''}',
      description: '${json['description'] ?? ''}',
      scoreType: '${json['scoreType'] ?? ''}',
    );
  }

  final String value;
  final String label;
  final String description;
  final String scoreType;
}

class AutismDevDraftDetail {
  const AutismDevDraftDetail({
    required this.id,
    required this.studentId,
    required this.studentName,
    required this.birthDate,
    required this.assessmentDate,
    required this.examinerName,
    required this.answeredItemCount,
    required this.completionPercent,
    required this.updatedTime,
    required this.progress,
    required this.input,
  });

  factory AutismDevDraftDetail.fromJson(Map<String, dynamic> json) {
    return AutismDevDraftDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      examinerName: '${json['examinerName'] ?? ''}',
      answeredItemCount: _intFrom(json['answeredItemCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      updatedTime: '${json['updatedTime'] ?? ''}',
      progress: AutismDevDraftProgress.fromJson(_mapFrom(json['progress'])),
      input: AutismDevDraftInput.fromJson(_mapFrom(json['input'])),
    );
  }

  final int id;
  final int studentId;
  final String studentName;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final int answeredItemCount;
  final double completionPercent;
  final String updatedTime;
  final AutismDevDraftProgress progress;
  final AutismDevDraftInput input;
}

class AutismDevDraftInput {
  const AutismDevDraftInput({
    this.studentId = 0,
    this.studentName = '',
    this.examinerName = '',
    this.remark = '',
    this.birthDate = '',
    this.assessmentDate = '',
    required this.itemScores,
    required this.itemRemarks,
  });

  factory AutismDevDraftInput.fromJson(Map<String, dynamic> json) {
    final Map<int, String> itemScores = <int, String>{};
    for (final MapEntry<String, dynamic> entry
        in _mapFrom(json['itemScores']).entries) {
      final int itemNo = _intFrom(entry.key);
      final String score = '${entry.value ?? ''}'.trim().toUpperCase();
      if (itemNo > 0 && score.isNotEmpty) {
        itemScores[itemNo] = score;
      }
    }
    for (final Object? raw in _rawListFrom(json['itemScoreList'])) {
      if (raw is! Map) {
        continue;
      }
      final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
      final int itemNo = _intFrom(item['itemNo']);
      final String score = '${item['score'] ?? ''}'.trim().toUpperCase();
      if (itemNo > 0 && score.isNotEmpty) {
        itemScores[itemNo] = score;
      }
    }

    final Map<int, String> itemRemarks = <int, String>{};
    for (final MapEntry<String, dynamic> entry
        in _mapFrom(json['itemRemarks']).entries) {
      final int itemNo = _intFrom(entry.key);
      final String remark = '${entry.value ?? ''}'.trim();
      if (itemNo > 0 && remark.isNotEmpty) {
        itemRemarks[itemNo] = remark;
      }
    }
    for (final Object? raw in _rawListFrom(json['itemRemarkList'])) {
      if (raw is! Map) {
        continue;
      }
      final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
      final int itemNo = _intFrom(item['itemNo']);
      final String remark = '${item['remark'] ?? ''}'.trim();
      if (itemNo > 0 && remark.isNotEmpty) {
        itemRemarks[itemNo] = remark;
      }
    }
    for (final Object? raw in _rawListFrom(json['itemScoreList'])) {
      if (raw is! Map) {
        continue;
      }
      final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
      final int itemNo = _intFrom(item['itemNo']);
      final String remark = '${item['remark'] ?? ''}'.trim();
      if (itemNo > 0 && remark.isNotEmpty) {
        itemRemarks[itemNo] = remark;
      }
    }

    return AutismDevDraftInput(
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      remark: '${json['remark'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      itemScores: itemScores,
      itemRemarks: itemRemarks,
    );
  }

  static const AutismDevDraftInput empty = AutismDevDraftInput(
    studentId: 0,
    studentName: '',
    examinerName: '',
    remark: '',
    birthDate: '',
    assessmentDate: '',
    itemScores: <int, String>{},
    itemRemarks: <int, String>{},
  );

  final int studentId;
  final String studentName;
  final String examinerName;
  final String remark;
  final String birthDate;
  final String assessmentDate;
  final Map<int, String> itemScores;
  final Map<int, String> itemRemarks;
}

class AutismDevDraftProgress {
  const AutismDevDraftProgress({
    required this.itemCount,
    required this.answeredItemCount,
    required this.missingItemCount,
    required this.completionPercent,
    required this.complete,
    required this.canScore,
    required this.missingItemNos,
  });

  factory AutismDevDraftProgress.fromJson(Map<String, dynamic> json) {
    return AutismDevDraftProgress(
      itemCount: _intFrom(json['itemCount']),
      answeredItemCount: _intFrom(json['answeredItemCount']),
      missingItemCount: _intFrom(json['missingItemCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      complete: json['complete'] == true,
      canScore: json['canScore'] == true,
      missingItemNos: _rawListFrom(json['missingItemNos'])
          .map(_intFrom)
          .where((int itemNo) => itemNo > 0)
          .toList(),
    );
  }

  static const AutismDevDraftProgress empty = AutismDevDraftProgress(
    itemCount: 0,
    answeredItemCount: 0,
    missingItemCount: 0,
    completionPercent: 0,
    complete: false,
    canScore: false,
    missingItemNos: <int>[],
  );

  final int itemCount;
  final int answeredItemCount;
  final int missingItemCount;
  final double completionPercent;
  final bool complete;
  final bool canScore;
  final List<int> missingItemNos;
}

Future<Object?> _getJson(
  http.Client client,
  Uri uri,
  String token, {
  required String fallbackMessage,
}) async {
  final http.Response response;
  try {
    response = await client
        .get(uri, headers: _authHeaders(token))
        .timeout(const Duration(seconds: 12));
  } on TimeoutException {
    throw AssessmentScaleApiException('$fallbackMessage：接口响应超时');
  } on Object catch (error) {
    throw AssessmentScaleApiException('$fallbackMessage：$error');
  }
  return _handleResponse(response, fallbackMessage: fallbackMessage);
}

Future<Object?> _postJson(
  http.Client client,
  Uri uri,
  String token,
  Map<String, dynamic> body, {
  required String fallbackMessage,
}) async {
  final http.Response response;
  try {
    response = await client
        .post(uri, headers: _jsonHeaders(token), body: jsonEncode(body))
        .timeout(const Duration(seconds: 12));
  } on TimeoutException {
    throw AssessmentScaleApiException('$fallbackMessage：接口响应超时');
  } on Object catch (error) {
    throw AssessmentScaleApiException('$fallbackMessage：$error');
  }
  return _handleResponse(response, fallbackMessage: fallbackMessage);
}

Object? _handleResponse(
  http.Response response, {
  required String fallbackMessage,
}) {
  final String body = utf8.decode(response.bodyBytes);
  Object? decoded;
  if (body.trim().isNotEmpty) {
    try {
      decoded = jsonDecode(body);
    } on Object {
      decoded = null;
    }
  }
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
  if (decoded is Map && decoded.containsKey('data')) {
    return decoded['data'];
  }
  if (decoded is Map && decoded.containsKey('result')) {
    return decoded['result'];
  }
  return decoded;
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

Map<String, String> _jsonHeaders(String token) {
  return <String, String>{
    ..._authHeaders(token),
    'Content-Type': 'application/json',
  };
}

String? _messageFromPayload(Object? payload) {
  if (payload is Map) {
    for (final String key in <String>['message', 'error', 'msg']) {
      final Object? value = payload[key];
      if (value != null && '$value'.trim().isNotEmpty) {
        return '$value';
      }
    }
  }
  return null;
}

Map<String, dynamic> _mapFrom(Object? raw) {
  if (raw is Map<String, dynamic>) {
    return raw;
  }
  if (raw is Map) {
    return Map<String, dynamic>.from(raw);
  }
  return <String, dynamic>{};
}

List<Map<String, dynamic>> _listFrom(Object? raw) {
  if (raw is List) {
    return raw
        .whereType<Map>()
        .map((Map item) => Map<String, dynamic>.from(item))
        .toList();
  }
  return <Map<String, dynamic>>[];
}

List<Object?> _rawListFrom(Object? raw) {
  if (raw is List) {
    return raw;
  }
  return <Object?>[];
}

List<String> _stringListFrom(Object? raw) {
  if (raw is List) {
    return raw.map((Object? item) => '$item'.trim()).where((String item) {
      return item.isNotEmpty;
    }).toList();
  }
  return <String>[];
}

String _autismDevDomainDisplayName(String raw, String domainCode) {
  final String name = raw.trim();
  final String code = domainCode.trim();
  if (name.isEmpty || code.isEmpty) {
    return name;
  }
  final String escapedCode = RegExp.escape(code);
  String cleaned = name
      .replaceFirst(
        RegExp('^\\s*$escapedCode\\s*[-_/：:]*\\s*', caseSensitive: false),
        '',
      )
      .replaceFirst(
        RegExp('\\s*[（(]\\s*$escapedCode\\s*[）)]\\s*\$', caseSensitive: false),
        '',
      )
      .replaceFirst(
        RegExp('\\s*[-_/：:]*\\s*$escapedCode\\s*\$', caseSensitive: false),
        '',
      )
      .trim();
  return cleaned.isEmpty ? name : cleaned;
}

int _intFrom(Object? raw) {
  if (raw is int) {
    return raw;
  }
  if (raw is num) {
    return raw.round();
  }
  return int.tryParse('$raw') ?? 0;
}

double _doubleFrom(Object? raw) {
  if (raw is num) {
    return raw.toDouble();
  }
  return double.tryParse('$raw') ?? 0;
}

String _dateOnlyFrom(Object? raw) {
  final String value = '${raw ?? ''}'.trim();
  if (value.length >= 10) {
    return value.substring(0, 10);
  }
  return value;
}
