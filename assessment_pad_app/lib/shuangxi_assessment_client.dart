import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'assessment_scale_client.dart';

const String defaultShuangxiTemplateSummaryPath = String.fromEnvironment(
  'SHUANGXI_TEMPLATE_SUMMARY_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/form-template/summary',
);
const String defaultShuangxiTemplateItemPath = String.fromEnvironment(
  'SHUANGXI_TEMPLATE_ITEM_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/form-template/item',
);
const String defaultShuangxiDraftSavePath = String.fromEnvironment(
  'SHUANGXI_DRAFT_SAVE_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/drafts/save',
);
const String defaultShuangxiDraftItemSavePath = String.fromEnvironment(
  'SHUANGXI_DRAFT_ITEM_SAVE_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/drafts/item/save',
);
const String defaultShuangxiDraftDetailPath = String.fromEnvironment(
  'SHUANGXI_DRAFT_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/drafts/detail',
);
const String defaultShuangxiDraftsPagePath = String.fromEnvironment(
  'SHUANGXI_DRAFTS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/drafts/page',
);
const String defaultShuangxiDraftSubmitPath = String.fromEnvironment(
  'SHUANGXI_DRAFT_SUBMIT_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/drafts/submit',
);
const String defaultShuangxiRecordsPagePath = String.fromEnvironment(
  'SHUANGXI_RECORDS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/records/page',
);
const String defaultShuangxiRecordDetailPath = String.fromEnvironment(
  'SHUANGXI_RECORD_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/records/detail',
);
const String defaultShuangxiRecordCategoryStatsPath = String.fromEnvironment(
  'SHUANGXI_RECORD_CATEGORY_STATS_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/records/category-stats',
);
const String defaultShuangxiStudentGenderUpdatePath = String.fromEnvironment(
  'SHUANGXI_STUDENT_GENDER_UPDATE_PATH',
  defaultValue: '/api/v1/assessments/scales/student-gender/update',
);

abstract class ShuangxiAssessmentClient {
  const ShuangxiAssessmentClient();

  Future<ShuangxiTemplateSummary> fetchTemplateSummary(String token);

  Future<ShuangxiAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  });

  Future<ShuangxiDraftDetail> saveDraft(
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

  Future<ShuangxiDraftDetail> fetchDraftDetail(String token, int id);

  Future<ShuangxiDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  );

  Future<void> submitDraft(String token, int draftId);

  Future<ShuangxiRecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    int studentId = 0,
    String assessmentDateEnd = '',
  });

  Future<ShuangxiRecordDetail> fetchRecordDetail(String token, int id);

  Future<String> updateStudentGender(
    String token, {
    required int studentId,
    required String gender,
  });
}

class ApiShuangxiAssessmentClient extends ShuangxiAssessmentClient {
  const ApiShuangxiAssessmentClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.templateSummaryPath = defaultShuangxiTemplateSummaryPath,
    this.templateItemPath = defaultShuangxiTemplateItemPath,
    this.draftSavePath = defaultShuangxiDraftSavePath,
    this.draftItemSavePath = defaultShuangxiDraftItemSavePath,
    this.draftDetailPath = defaultShuangxiDraftDetailPath,
    this.draftsPagePath = defaultShuangxiDraftsPagePath,
    this.draftSubmitPath = defaultShuangxiDraftSubmitPath,
    this.recordsPagePath = defaultShuangxiRecordsPagePath,
    this.recordDetailPath = defaultShuangxiRecordDetailPath,
    this.studentGenderUpdatePath = defaultShuangxiStudentGenderUpdatePath,
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
  final String recordsPagePath;
  final String recordDetailPath;
  final String studentGenderUpdatePath;
  final http.Client? httpClient;

  static final http.Client _sharedHttpClient = http.Client();
  static final Map<String, Future<ShuangxiTemplateSummary>> _summaryRequests =
      <String, Future<ShuangxiTemplateSummary>>{};

  http.Client get _client => httpClient ?? _sharedHttpClient;

  @override
  Future<ShuangxiTemplateSummary> fetchTemplateSummary(String token) async {
    final String cacheKey = '$educationBaseUrl|$templateSummaryPath';
    if (httpClient == null) {
      final Future<ShuangxiTemplateSummary>? pending =
          _summaryRequests[cacheKey];
      if (pending != null) {
        return pending;
      }
    }
    final Future<ShuangxiTemplateSummary> request =
        _fetchTemplateSummaryFromNetwork(token);
    if (httpClient == null) {
      _summaryRequests[cacheKey] = request;
    }
    try {
      return await request;
    } finally {
      if (httpClient == null &&
          identical(_summaryRequests[cacheKey], request)) {
        _summaryRequests.remove(cacheKey);
      }
    }
  }

  Future<ShuangxiTemplateSummary> _fetchTemplateSummaryFromNetwork(
    String token,
  ) async {
    final Object? data = await _getJson(
      _uri(educationBaseUrl, templateSummaryPath),
      token,
      fallbackMessage: '双溪课程评量表A模板加载失败',
    );
    if (data is! Map) {
      return ShuangxiTemplateSummary.empty;
    }
    return ShuangxiTemplateSummary.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ShuangxiAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  }) async {
    final Object? data = await _getJson(
      _uri(
        educationBaseUrl,
        templateItemPath,
        <String, String>{'itemNo': '$itemNo'},
      ),
      token,
      fallbackMessage: '双溪课程评量表A题目加载失败',
    );
    if (data is! Map) {
      return ShuangxiAssessmentItem.empty;
    }
    return ShuangxiAssessmentItem.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ShuangxiDraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, draftSavePath),
      token,
      payload,
      fallbackMessage: '双溪课程评量表A草稿保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('草稿保存返回格式不正确');
    }
    return ShuangxiDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
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
          'assessmentCode': 'SHUANGXI_A',
          if (studentId > 0) 'studentId': studentId,
          if (latestOnly) 'latestOnly': true,
        },
        if (latestOnly) 'latestOnly': true,
      },
      fallbackMessage: '双溪课程评量表A草稿列表加载失败',
    );
    if (data is! Map) {
      return AssessmentDraftPage.empty;
    }
    return AssessmentDraftPage.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ShuangxiDraftDetail> fetchDraftDetail(String token, int id) async {
    final Object? data = await _getJson(
      _uri(
        educationBaseUrl,
        draftDetailPath,
        <String, String>{'id': '$id'},
      ),
      token,
      fallbackMessage: '双溪课程评量表A草稿详情加载失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('草稿详情返回格式不正确');
    }
    return ShuangxiDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ShuangxiDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, draftItemSavePath),
      token,
      payload,
      fallbackMessage: '双溪课程评量表A单题保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('单题保存返回格式不正确');
    }
    return ShuangxiDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<void> submitDraft(String token, int draftId) async {
    await _postJson(
      _uri(educationBaseUrl, draftSubmitPath),
      token,
      <String, int>{'id': draftId},
      fallbackMessage: '双溪课程评量表A正式记录提交失败',
    );
  }

  @override
  Future<ShuangxiRecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    int studentId = 0,
    String assessmentDateEnd = '',
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, recordsPagePath),
      token,
      <String, dynamic>{
        'pageRequestModel': <String, int>{
          'pageIndex': pageIndex,
          'pageSize': pageSize,
        },
        'queryModel': <String, dynamic>{
          'assessmentCode': 'SHUANGXI_A',
          if (studentId > 0) 'studentId': studentId,
          if (assessmentDateEnd.trim().isNotEmpty)
            'assessmentDateEnd': assessmentDateEnd.trim(),
        },
      },
      fallbackMessage: '双溪课程评量表A记录列表加载失败',
    );
    if (data is! Map) {
      return ShuangxiRecordPage.empty;
    }
    return ShuangxiRecordPage.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ShuangxiRecordDetail> fetchRecordDetail(
    String token,
    int id,
  ) async {
    final Object? data = await _getJson(
      _uri(
        educationBaseUrl,
        recordDetailPath,
        <String, String>{'id': '$id'},
      ),
      token,
      fallbackMessage: '双溪课程评量表A记录详情加载失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('测评记录详情返回格式不正确');
    }
    return ShuangxiRecordDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<String> updateStudentGender(
    String token, {
    required int studentId,
    required String gender,
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, studentGenderUpdatePath),
      token,
      <String, dynamic>{
        'studentId': studentId,
        'gender': gender,
      },
      fallbackMessage: '学生性别更新失败',
    );
    if (data is Map) {
      final String updated = '${data['gender'] ?? ''}'.trim();
      if (updated.isNotEmpty) {
        return updated;
      }
    }
    return gender.trim();
  }

  Future<Object?> _getJson(
    Uri uri,
    String token, {
    required String fallbackMessage,
  }) async {
    final http.Response response;
    try {
      response = await _client
          .get(uri, headers: _headers(token))
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const AssessmentScaleApiException('双溪课程评量表A接口响应超时，请检查网络');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接双溪课程评量表A接口：$error');
    }
    return _handleResponse(response, fallbackMessage: fallbackMessage);
  }

  Future<Object?> _postJson(
    Uri uri,
    String token,
    Map<String, dynamic> payload, {
    required String fallbackMessage,
  }) async {
    final http.Response response;
    try {
      response = await _client
          .post(uri, headers: _jsonHeaders(token), body: jsonEncode(payload))
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const AssessmentScaleApiException('双溪课程评量表A接口响应超时，请检查网络');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接双溪课程评量表A接口：$error');
    }
    return _handleResponse(response, fallbackMessage: fallbackMessage);
  }

  Map<String, String> _headers(String token) {
    return <String, String>{
      'Accept': 'application/json',
      if (token.trim().isNotEmpty) 'Authorization': 'Bearer ${token.trim()}',
      if (token.trim().isNotEmpty) 'X-Access-Token': token.trim(),
    };
  }

  Map<String, String> _jsonHeaders(String token) {
    return <String, String>{
      ..._headers(token),
      'Content-Type': 'application/json',
    };
  }

  Future<Object?> _handleResponse(
    http.Response response, {
    required String fallbackMessage,
  }) async {
    final Object? decoded = _decodeResponse(utf8.decode(response.bodyBytes));
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
      if (envelope.containsKey('result')) {
        return envelope['result'];
      }
    }
    return decoded;
  }
}

class ShuangxiTemplateSummary {
  const ShuangxiTemplateSummary({
    required this.title,
    required this.itemCount,
    required this.domainCount,
    required this.skillCount,
    required this.scoreMin,
    required this.scoreMax,
    required this.domains,
    required this.scoreOptions,
  });

  factory ShuangxiTemplateSummary.fromJson(Map<String, dynamic> json) {
    return ShuangxiTemplateSummary(
      title: '${json['title'] ?? '双溪课程评量表A'}',
      itemCount: _intFrom(json['itemCount']),
      domainCount: _intFrom(json['domainCount']),
      skillCount: _intFrom(json['skillCount']),
      scoreMin: _intFrom(json['scoreMin']),
      scoreMax: _intFrom(json['scoreMax']),
      domains: _listFrom(json['domains'])
          .map(ShuangxiDomainSummary.fromJson)
          .toList(),
      scoreOptions: _listFrom(json['scoreOptions'])
          .map(ShuangxiScoreOption.fromJson)
          .toList(),
    );
  }

  static const ShuangxiTemplateSummary empty = ShuangxiTemplateSummary(
    title: '双溪课程评量表A',
    itemCount: 0,
    domainCount: 0,
    skillCount: 0,
    scoreMin: 0,
    scoreMax: 3,
    domains: <ShuangxiDomainSummary>[],
    scoreOptions: <ShuangxiScoreOption>[],
  );

  final String title;
  final int itemCount;
  final int domainCount;
  final int skillCount;
  final int scoreMin;
  final int scoreMax;
  final List<ShuangxiDomainSummary> domains;
  final List<ShuangxiScoreOption> scoreOptions;

  List<ShuangxiSkillSummary> get allSkills => domains
      .expand((ShuangxiDomainSummary domain) => domain.skills)
      .toList(growable: false);

  List<ShuangxiItemSummary> get allItems => allSkills
      .expand((ShuangxiSkillSummary skill) => skill.items)
      .toList(growable: false);
}

class ShuangxiDomainSummary {
  const ShuangxiDomainSummary({
    required this.domainCode,
    required this.domainName,
    required this.sortNo,
    required this.itemCount,
    required this.maxRawScore,
    required this.skills,
  });

  factory ShuangxiDomainSummary.fromJson(Map<String, dynamic> json) {
    return ShuangxiDomainSummary(
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      sortNo: _intFrom(json['sortNo']),
      itemCount: _intFrom(json['itemCount']),
      maxRawScore: _intFrom(json['maxRawScore']),
      skills:
          _listFrom(json['skills']).map(ShuangxiSkillSummary.fromJson).toList(),
    );
  }

  final String domainCode;
  final String domainName;
  final int sortNo;
  final int itemCount;
  final int maxRawScore;
  final List<ShuangxiSkillSummary> skills;
}

class ShuangxiSkillSummary {
  const ShuangxiSkillSummary({
    required this.skillCode,
    required this.skillName,
    required this.domainCode,
    required this.domainName,
    required this.sortNo,
    required this.itemCount,
    required this.items,
  });

  factory ShuangxiSkillSummary.fromJson(Map<String, dynamic> json) {
    return ShuangxiSkillSummary(
      skillCode: '${json['skillCode'] ?? ''}',
      skillName: '${json['skillName'] ?? ''}',
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      sortNo: _intFrom(json['sortNo']),
      itemCount: _intFrom(json['itemCount']),
      items:
          _listFrom(json['items']).map(ShuangxiItemSummary.fromJson).toList(),
    );
  }

  final String skillCode;
  final String skillName;
  final String domainCode;
  final String domainName;
  final int sortNo;
  final int itemCount;
  final List<ShuangxiItemSummary> items;
}

class ShuangxiItemSummary {
  const ShuangxiItemSummary({
    required this.itemNo,
    required this.itemCode,
    required this.itemTitle,
    required this.testItem,
    required this.domainCode,
    required this.domainName,
    required this.skillCode,
    required this.skillName,
  });

  factory ShuangxiItemSummary.fromJson(Map<String, dynamic> json) {
    return ShuangxiItemSummary(
      itemNo: _intFrom(json['itemNo']),
      itemCode: '${json['itemCode'] ?? ''}',
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      skillCode: '${json['skillCode'] ?? ''}',
      skillName: '${json['skillName'] ?? ''}',
    );
  }

  static const ShuangxiItemSummary empty = ShuangxiItemSummary(
    itemNo: 0,
    itemCode: '',
    itemTitle: '',
    testItem: '',
    domainCode: '',
    domainName: '',
    skillCode: '',
    skillName: '',
  );

  final int itemNo;
  final String itemCode;
  final String itemTitle;
  final String testItem;
  final String domainCode;
  final String domainName;
  final String skillCode;
  final String skillName;
}

class ShuangxiAssessmentItem {
  const ShuangxiAssessmentItem({
    required this.itemNo,
    required this.itemCode,
    required this.itemTitle,
    required this.testItem,
    required this.domainCode,
    required this.domainName,
    required this.skillCode,
    required this.skillName,
    required this.scoreOptions,
  });

  factory ShuangxiAssessmentItem.fromJson(Map<String, dynamic> json) {
    return ShuangxiAssessmentItem(
      itemNo: _intFrom(json['itemNo']),
      itemCode: '${json['itemCode'] ?? ''}',
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      skillCode: '${json['skillCode'] ?? ''}',
      skillName: '${json['skillName'] ?? ''}',
      scoreOptions: _listFrom(json['scoreOptions'])
          .map(ShuangxiScoreOption.fromJson)
          .toList(),
    );
  }

  static const ShuangxiAssessmentItem empty = ShuangxiAssessmentItem(
    itemNo: 0,
    itemCode: '',
    itemTitle: '',
    testItem: '',
    domainCode: '',
    domainName: '',
    skillCode: '',
    skillName: '',
    scoreOptions: <ShuangxiScoreOption>[],
  );

  final int itemNo;
  final String itemCode;
  final String itemTitle;
  final String testItem;
  final String domainCode;
  final String domainName;
  final String skillCode;
  final String skillName;
  final List<ShuangxiScoreOption> scoreOptions;
}

class ShuangxiScoreOption {
  const ShuangxiScoreOption({
    required this.value,
    required this.label,
    required this.description,
  });

  factory ShuangxiScoreOption.fromJson(Map<String, dynamic> json) {
    return ShuangxiScoreOption(
      value: _intFrom(json['value']),
      label: '${json['label'] ?? ''}',
      description: '${json['description'] ?? ''}',
    );
  }

  final int value;
  final String label;
  final String description;
}

class ShuangxiDraftDetail {
  const ShuangxiDraftDetail({
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

  factory ShuangxiDraftDetail.fromJson(Map<String, dynamic> json) {
    return ShuangxiDraftDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      examinerName: '${json['examinerName'] ?? ''}',
      answeredItemCount: _intFrom(json['answeredItemCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      updatedTime: '${json['updatedTime'] ?? ''}',
      progress: ShuangxiDraftProgress.fromJson(_mapFrom(json['progress'])),
      input: ShuangxiDraftInput.fromJson(_mapFrom(json['input'])),
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
  final ShuangxiDraftProgress progress;
  final ShuangxiDraftInput input;
}

class ShuangxiDraftInput {
  const ShuangxiDraftInput({
    this.studentId = 0,
    this.studentName = '',
    this.studentGender = '',
    this.examinerName = '',
    this.remark = '',
    this.birthDate = '',
    this.assessmentDate = '',
    required this.itemScores,
  });

  factory ShuangxiDraftInput.fromJson(Map<String, dynamic> json) {
    final Map<int, int> itemScores = <int, int>{};
    for (final MapEntry<String, dynamic> entry
        in _mapFrom(json['itemScores']).entries) {
      final int itemNo = _intFrom(entry.key);
      final int score = _intFrom(entry.value);
      if (itemNo > 0) {
        itemScores[itemNo] = score;
      }
    }
    for (final Object? raw in _rawListFrom(json['itemScoreList'])) {
      if (raw is! Map) {
        continue;
      }
      final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
      final int itemNo = _intFrom(item['itemNo']);
      final int score = _intFrom(item['score']);
      if (itemNo > 0) {
        itemScores[itemNo] = score;
      }
    }
    return ShuangxiDraftInput(
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      studentGender: '${json['studentGender'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      remark: '${json['remark'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      itemScores: itemScores,
    );
  }

  static const ShuangxiDraftInput empty = ShuangxiDraftInput(
    itemScores: <int, int>{},
  );

  final int studentId;
  final String studentName;
  final String studentGender;
  final String examinerName;
  final String remark;
  final String birthDate;
  final String assessmentDate;
  final Map<int, int> itemScores;
}

class ShuangxiDraftProgress {
  const ShuangxiDraftProgress({
    required this.itemCount,
    required this.answeredItemCount,
    required this.missingItemCount,
    required this.completionPercent,
    required this.complete,
    required this.canScore,
    required this.missingItemNos,
  });

  factory ShuangxiDraftProgress.fromJson(Map<String, dynamic> json) {
    return ShuangxiDraftProgress(
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

  static const ShuangxiDraftProgress empty = ShuangxiDraftProgress(
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

class ShuangxiRecordPage {
  const ShuangxiRecordPage({
    required this.items,
    required this.total,
    required this.current,
    required this.size,
  });

  factory ShuangxiRecordPage.fromJson(Map<String, dynamic> json) {
    return ShuangxiRecordPage(
      items:
          _listFrom(json['items']).map(ShuangxiRecordSummary.fromJson).toList(),
      total: _intFrom(json['total']),
      current: _intFrom(json['current']),
      size: _intFrom(json['size']),
    );
  }

  static const ShuangxiRecordPage empty = ShuangxiRecordPage(
    items: <ShuangxiRecordSummary>[],
    total: 0,
    current: 1,
    size: 0,
  );

  final List<ShuangxiRecordSummary> items;
  final int total;
  final int current;
  final int size;
}

class ShuangxiRecordSummary {
  const ShuangxiRecordSummary({
    required this.id,
    required this.studentId,
    required this.studentName,
    required this.assessmentCode,
    required this.assessmentName,
    required this.birthDate,
    required this.assessmentDate,
    required this.examinerName,
    required this.updatedTime,
    this.scaleVersion = '',
    this.createdTime = '',
  });

  factory ShuangxiRecordSummary.fromJson(Map<String, dynamic> json) {
    return ShuangxiRecordSummary(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      assessmentCode: '${json['assessmentCode'] ?? ''}',
      assessmentName: '${json['assessmentName'] ?? ''}',
      scaleVersion: '${json['scaleVersion'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      examinerName: '${json['examinerName'] ?? ''}',
      createdTime: '${json['createdTime'] ?? ''}',
      updatedTime: '${json['updatedTime'] ?? ''}',
    );
  }

  final int id;
  final int studentId;
  final String studentName;
  final String assessmentCode;
  final String assessmentName;
  final String scaleVersion;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final String createdTime;
  final String updatedTime;
}

class ShuangxiRecordDetail extends ShuangxiRecordSummary {
  const ShuangxiRecordDetail({
    required super.id,
    required super.studentId,
    required super.studentName,
    required super.assessmentCode,
    required super.assessmentName,
    required super.birthDate,
    required super.assessmentDate,
    required super.examinerName,
    required super.updatedTime,
    super.scaleVersion,
    super.createdTime,
    required this.input,
  });

  factory ShuangxiRecordDetail.fromJson(Map<String, dynamic> json) {
    return ShuangxiRecordDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      assessmentCode: '${json['assessmentCode'] ?? ''}',
      assessmentName: '${json['assessmentName'] ?? ''}',
      scaleVersion: '${json['scaleVersion'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      examinerName: '${json['examinerName'] ?? ''}',
      createdTime: '${json['createdTime'] ?? ''}',
      updatedTime: '${json['updatedTime'] ?? ''}',
      input: ShuangxiDraftInput.fromJson(_mapFrom(json['input'])),
    );
  }

  final ShuangxiDraftInput input;
}

Uri _uri(String baseUrl, String path, [Map<String, String>? query]) {
  final String trimmedBase = baseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
  final String normalizedPath = path.startsWith('/') ? path : '/$path';
  return Uri.parse('$trimmedBase$normalizedPath').replace(
    queryParameters: query == null || query.isEmpty ? null : query,
  );
}

Object? _decodeResponse(String body) {
  if (body.trim().isEmpty) {
    return null;
  }
  return jsonDecode(body);
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

List<Map<String, dynamic>> _listFrom(Object? value) {
  if (value is List) {
    return value
        .whereType<Map>()
        .map((Map<dynamic, dynamic> item) => Map<String, dynamic>.from(item))
        .toList();
  }
  return <Map<String, dynamic>>[];
}

Map<String, dynamic> _mapFrom(Object? value) {
  if (value is Map<String, dynamic>) {
    return value;
  }
  if (value is Map) {
    return Map<String, dynamic>.from(value);
  }
  return <String, dynamic>{};
}

List<Object?> _rawListFrom(Object? value) {
  if (value is List) {
    return value;
  }
  return <Object?>[];
}

int _intFrom(Object? value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.round();
  }
  return int.tryParse('${value ?? ''}') ?? 0;
}

double _doubleFrom(Object? value) {
  if (value is num) {
    return value.toDouble();
  }
  return double.tryParse('${value ?? ''}') ?? 0;
}

String _dateOnlyFrom(Object? value) {
  final String raw = '${value ?? ''}'.trim();
  if (raw.length >= 10) {
    return raw.substring(0, 10);
  }
  return raw;
}
