import 'dart:async';

import 'package:assessment_pad_app/main.dart' show PadViewport;
import 'package:assessment_pad_app/assessment_scale_client.dart';
import 'package:assessment_pad_app/shuangxi_assessment_client.dart';
import 'package:assessment_pad_app/shuangxi_assessment_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('Shuangxi assessment page fits compact pad layout',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'test-token',
    });
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeShuangxiAssessmentClient client =
        _FakeShuangxiAssessmentClient();
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: ShuangxiAssessmentPage(
              client: client,
              onBack: () {},
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('双溪课程评量表A 测评工作台'), findsOneWidget);
    expect(find.text('1.1.1 视觉敏锐度'), findsAtLeastNWidgets(1));
    expect(find.text('领域：感官知觉'), findsNothing);
    expect(find.textContaining('评量方式'), findsNothing);
    expect(find.text('测评进度'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Shuangxi assessment page shows workbench skeleton while loading',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'test-token',
    });
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeShuangxiAssessmentClient client =
        _FakeShuangxiAssessmentClient(holdTemplateSummary: true);
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: ShuangxiAssessmentPage(
              client: client,
              onBack: () {},
            ),
          ),
        ),
      ),
    );

    expect(find.text('测评进度'), findsOneWidget);
    expect(find.text('观察备注'), findsOneWidget);
    expect(find.text('自动下一题'), findsOneWidget);
    expect(find.byType(CircularProgressIndicator), findsNothing);
    expect(tester.takeException(), isNull);

    client.completeTemplateSummary();
    await tester.pumpAndSettle();
  });

  testWidgets('Shuangxi assessment page marks previous score option',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'test-token',
    });
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeShuangxiAssessmentClient client = _FakeShuangxiAssessmentClient(
      previousAssessmentDate: '2026-05-01',
      previousItemScores: const <int, int>{1: 2},
    );
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: ShuangxiAssessmentPage(
              client: client,
              args: const ShuangxiAssessmentLaunchArgs(
                studentId: 1,
                studentName: '小明',
                studentGender: '男',
                assessmentDate: '2026-05-18',
                examinerName: '陈老师',
              ),
              onBack: () {},
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('上次测评 2026-05-01'), findsOneWidget);
    expect(find.text('上次 05-01'), findsOneWidget);
    expect(find.textContaining('2分 · 能看到眼前一至二公尺远的小物体'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Shuangxi auto next waits briefly before advancing',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'test-token',
    });
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeShuangxiAssessmentClient client =
        _FakeShuangxiAssessmentClient();
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: ShuangxiAssessmentPage(
              client: client,
              onBack: () {},
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('0分').last);
    await tester.pump();
    expect(find.text('1.1.1 视觉敏锐度'), findsAtLeastNWidgets(1));

    await tester.pump(const Duration(milliseconds: 240));
    expect(find.text('1.1.1 视觉敏锐度'), findsAtLeastNWidgets(1));

    await tester.pump(const Duration(milliseconds: 120));
    await tester.pumpAndSettle();
    expect(find.text('1.1.2 视觉追视能力'), findsAtLeastNWidgets(1));
    expect(tester.takeException(), isNull);
  });

  testWidgets('Shuangxi sex-specific item auto applies three points',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'test-token',
    });
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeShuangxiAssessmentClient client =
        _FakeShuangxiAssessmentClient(firstItemNo: 82);
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: ShuangxiAssessmentPage(
              client: client,
              args: const ShuangxiAssessmentLaunchArgs(
                studentId: 1,
                studentName: '小明',
                studentGender: '男',
                assessmentDate: '2026-05-18',
                examinerName: '陈老师',
              ),
              onBack: () {},
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('男生不适用，按 3 分计算'), findsOneWidget);
    expect(find.text('已按规则'), findsOneWidget);

    await tester.tap(find.text('保存草稿'));
    await tester.pumpAndSettle();

    final Map<String, dynamic>? payload = client.lastSavedDraftPayload;
    expect(payload, isNotNull);
    expect(payload!['studentGender'], '男');
    final List<dynamic> itemScoreList =
        payload['itemScoreList'] as List<dynamic>? ?? <dynamic>[];
    final Map<String, dynamic> item82 = Map<String, dynamic>.from(
      itemScoreList.firstWhere(
        (dynamic raw) => (raw as Map)['itemNo'] == 82,
      ) as Map,
    );
    expect(item82['score'], 3);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Shuangxi asks for gender when student gender is unknown',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'test-token',
    });
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeShuangxiAssessmentClient client =
        _FakeShuangxiAssessmentClient(firstItemNo: 82);
    bool backed = false;
    String syncedGender = '';
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: ShuangxiAssessmentPage(
              client: client,
              args: ShuangxiAssessmentLaunchArgs(
                studentId: 1,
                studentName: '小明',
                studentGender: '-',
                assessmentDate: '2026-05-18',
                examinerName: '陈老师',
                onStudentGenderUpdated: (String gender) {
                  syncedGender = gender;
                },
              ),
              onBack: () => backed = true,
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('请先确认学生性别'), findsOneWidget);
    expect(find.text('男生不适用，按 3 分计算'), findsNothing);
    await tester.tap(find.text('男'));
    await tester.pumpAndSettle();

    expect(client.updateStudentGenderCount, 0);
    expect(find.text('请先确认学生性别'), findsOneWidget);
    await tester.tap(find.text('确定'));
    await tester.pumpAndSettle();

    expect(client.updateStudentGenderCount, 1);
    expect(client.lastUpdatedStudentId, 1);
    expect(client.lastUpdatedGender, '男');
    expect(syncedGender, '男');
    expect(backed, isFalse);
    expect(find.text('请先确认学生性别'), findsNothing);
    expect(find.text('男生不适用，按 3 分计算'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Shuangxi submit defaults missing scores to zero',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'test-token',
    });
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeShuangxiAssessmentClient client =
        _FakeShuangxiAssessmentClient();
    bool backed = false;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: ShuangxiAssessmentPage(
              client: client,
              args: const ShuangxiAssessmentLaunchArgs(
                studentId: 1,
                studentName: '小明',
                studentGender: '男',
                assessmentDate: '2026-05-18',
                examinerName: '陈老师',
              ),
              onBack: () => backed = true,
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('0分').last);
    await tester.pumpAndSettle();

    await tester.tap(find.text('提交记录'));
    await tester.pumpAndSettle();

    expect(find.text('未评分题将按 0 分提交'), findsOneWidget);
    expect(find.text('按 0 分提交'), findsOneWidget);

    final ButtonStyleButton confirmButton = tester.widget<ButtonStyleButton>(
      find.byKey(
        const ValueKey<String>('shuangxi-submit-zero-confirm-button'),
      ),
    );
    confirmButton.onPressed!();
    await tester.pumpAndSettle();
    await tester.pump(const Duration(milliseconds: 700));
    await tester.pumpAndSettle();

    expect(client.saveDraftCount, greaterThan(0));
    expect(client.submitDraftCount, 1);
    expect(backed, isTrue);

    final Map<String, dynamic>? payload = client.lastSavedDraftPayload;
    expect(payload, isNotNull);
    final List<dynamic> itemScoreList =
        (payload!['itemScoreList'] as List<dynamic>? ?? <dynamic>[]);
    expect(itemScoreList.length, 209);
    final Map<int, int> scoresByItem = <int, int>{};
    for (final dynamic raw in itemScoreList) {
      final Map<String, dynamic> item = Map<String, dynamic>.from(raw as Map);
      scoresByItem[item['itemNo'] as int] = item['score'] as int;
    }
    expect(scoresByItem.length, 209);
    expect(scoresByItem[1], 0);
    expect(scoresByItem[2], 0);
    expect(scoresByItem[209], 0);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Shuangxi observation remark accepts input and saves',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'test-token',
    });
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeShuangxiAssessmentClient client =
        _FakeShuangxiAssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: ShuangxiAssessmentPage(
              client: client,
              args: const ShuangxiAssessmentLaunchArgs(
                studentId: 1,
                studentName: '小明',
                studentGender: '男',
                assessmentDate: '2026-05-18',
                examinerName: '陈老师',
              ),
              onBack: () {},
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    const String remark = '学生能跟随提示完成，注意力维持较短。';
    final Finder remarkField =
        find.byKey(const ValueKey<String>('shuangxi-observation-remark-field'));
    await tester.tap(remarkField);
    await tester.enterText(remarkField, remark);
    await tester.pump();
    expect(find.text(remark), findsOneWidget);
    expect(tester.testTextInput.isVisible, isTrue);

    await tester.tap(find.text('测评进度'));
    await tester.pump();
    expect(tester.testTextInput.isVisible, isFalse);

    await tester.tap(find.text('保存草稿'));
    await tester.pumpAndSettle();

    final Map<String, dynamic>? payload = client.lastSavedDraftPayload;
    expect(payload, isNotNull);
    final Map<dynamic, dynamic> itemRemarks =
        payload!['itemRemarks'] as Map<dynamic, dynamic>? ??
            <dynamic, dynamic>{};
    expect(itemRemarks['1'], remark);

    await tester.tap(find.text('下一题'));
    await tester.pumpAndSettle();
    expect(find.text(remark), findsNothing);

    await tester.tap(find.text('上一题'));
    await tester.pumpAndSettle();
    expect(find.text(remark), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}

class _FakeShuangxiAssessmentClient extends ShuangxiAssessmentClient {
  _FakeShuangxiAssessmentClient({
    bool holdTemplateSummary = false,
    this.previousAssessmentDate = '',
    this.previousItemScores = const <int, int>{},
    this.firstItemNo = 1,
  }) : _templateSummaryCompleter =
            holdTemplateSummary ? Completer<ShuangxiTemplateSummary>() : null;

  final Completer<ShuangxiTemplateSummary>? _templateSummaryCompleter;
  final String previousAssessmentDate;
  final Map<int, int> previousItemScores;
  final int firstItemNo;
  Map<String, dynamic>? lastSavedDraftPayload;
  int saveDraftCount = 0;
  int submitDraftCount = 0;
  int updateStudentGenderCount = 0;
  int lastUpdatedStudentId = 0;
  String lastUpdatedGender = '';

  void completeTemplateSummary() {
    final Completer<ShuangxiTemplateSummary>? completer =
        _templateSummaryCompleter;
    if (completer != null && !completer.isCompleted) {
      completer.complete(_templateSummary());
    }
  }

  @override
  Future<ShuangxiTemplateSummary> fetchTemplateSummary(String token) async {
    final Completer<ShuangxiTemplateSummary>? completer =
        _templateSummaryCompleter;
    if (completer != null) {
      return completer.future;
    }
    return _templateSummary();
  }

  ShuangxiTemplateSummary _templateSummary() {
    if (firstItemNo == 82) {
      return const ShuangxiTemplateSummary(
        title: '双溪课程评量表A',
        itemCount: 209,
        domainCount: 7,
        skillCount: 34,
        scoreMin: 0,
        scoreMax: 3,
        scoreOptions: <ShuangxiScoreOption>[
          ShuangxiScoreOption(value: 0, label: '0分', description: '尚未出现'),
          ShuangxiScoreOption(value: 1, label: '1分', description: '大量协助'),
          ShuangxiScoreOption(value: 2, label: '2分', description: '少量协助'),
          ShuangxiScoreOption(value: 3, label: '3分', description: '独立完成'),
        ],
        domains: <ShuangxiDomainSummary>[
          ShuangxiDomainSummary(
            domainCode: 'SELF_CARE',
            domainName: '生活自理',
            sortNo: 4,
            itemCount: 24,
            maxRawScore: 72,
            skills: <ShuangxiSkillSummary>[
              ShuangxiSkillSummary(
                skillCode: '4.4',
                skillName: '身体之清洁',
                domainCode: 'SELF_CARE',
                domainName: '生活自理',
                sortNo: 4,
                itemCount: 1,
                items: <ShuangxiItemSummary>[
                  ShuangxiItemSummary(
                    itemNo: 82,
                    itemCode: '4.4.8',
                    itemTitle: '4.4.8 使用卫生棉',
                    testItem: '使用卫生棉',
                    domainCode: 'SELF_CARE',
                    domainName: '生活自理',
                    skillCode: '4.4',
                    skillName: '身体之清洁',
                  ),
                ],
              ),
            ],
          ),
        ],
      );
    }
    return const ShuangxiTemplateSummary(
      title: '双溪课程评量表A',
      itemCount: 209,
      domainCount: 7,
      skillCount: 34,
      scoreMin: 0,
      scoreMax: 3,
      scoreOptions: <ShuangxiScoreOption>[
        ShuangxiScoreOption(value: 0, label: '0分', description: '尚未出现'),
        ShuangxiScoreOption(value: 1, label: '1分', description: '大量协助'),
        ShuangxiScoreOption(value: 2, label: '2分', description: '少量协助'),
        ShuangxiScoreOption(value: 3, label: '3分', description: '独立完成'),
      ],
      domains: <ShuangxiDomainSummary>[
        ShuangxiDomainSummary(
          domainCode: 'SENSORY',
          domainName: '感官知觉',
          sortNo: 1,
          itemCount: 21,
          maxRawScore: 63,
          skills: <ShuangxiSkillSummary>[
            ShuangxiSkillSummary(
              skillCode: '1.1',
              skillName: '视觉的运用',
              domainCode: 'SENSORY',
              domainName: '感官知觉',
              sortNo: 1,
              itemCount: 2,
              items: <ShuangxiItemSummary>[
                ShuangxiItemSummary(
                  itemNo: 1,
                  itemCode: '1.1.1',
                  itemTitle: '1.1.1 视觉敏锐度',
                  testItem: '视觉敏锐度',
                  domainCode: 'SENSORY',
                  domainName: '感官知觉',
                  skillCode: '1.1',
                  skillName: '视觉的运用',
                ),
                ShuangxiItemSummary(
                  itemNo: 2,
                  itemCode: '1.1.2',
                  itemTitle: '1.1.2 视觉追视能力',
                  testItem: '视觉追视能力',
                  domainCode: 'SENSORY',
                  domainName: '感官知觉',
                  skillCode: '1.1',
                  skillName: '视觉的运用',
                ),
              ],
            ),
          ],
        ),
        ShuangxiDomainSummary(
          domainCode: 'GROSS_MOTOR',
          domainName: '粗大动作',
          sortNo: 2,
          itemCount: 25,
          maxRawScore: 75,
          skills: <ShuangxiSkillSummary>[],
        ),
        ShuangxiDomainSummary(
          domainCode: 'FINE_MOTOR',
          domainName: '精细动作',
          sortNo: 3,
          itemCount: 14,
          maxRawScore: 42,
          skills: <ShuangxiSkillSummary>[],
        ),
        ShuangxiDomainSummary(
          domainCode: 'SELF_CARE',
          domainName: '生活自理',
          sortNo: 4,
          itemCount: 24,
          maxRawScore: 72,
          skills: <ShuangxiSkillSummary>[],
        ),
        ShuangxiDomainSummary(
          domainCode: 'COMMUNICATION',
          domainName: '沟通',
          sortNo: 5,
          itemCount: 56,
          maxRawScore: 168,
          skills: <ShuangxiSkillSummary>[],
        ),
        ShuangxiDomainSummary(
          domainCode: 'COGNITION',
          domainName: '认知',
          sortNo: 6,
          itemCount: 21,
          maxRawScore: 63,
          skills: <ShuangxiSkillSummary>[],
        ),
        ShuangxiDomainSummary(
          domainCode: 'SOCIAL_SKILLS',
          domainName: '社会技能',
          sortNo: 7,
          itemCount: 48,
          maxRawScore: 144,
          skills: <ShuangxiSkillSummary>[],
        ),
      ],
    );
  }

  @override
  Future<ShuangxiAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  }) async {
    if (itemNo == 82) {
      return const ShuangxiAssessmentItem(
        itemNo: 82,
        itemCode: '4.4.8',
        itemTitle: '4.4.8 使用卫生棉',
        testItem: '使用卫生棉',
        domainCode: 'SELF_CARE',
        domainName: '生活自理',
        skillCode: '4.4',
        skillName: '身体之清洁',
        scoreOptions: <ShuangxiScoreOption>[
          ShuangxiScoreOption(value: 0, label: '0分', description: '尚未出现'),
          ShuangxiScoreOption(value: 1, label: '1分', description: '大量协助'),
          ShuangxiScoreOption(value: 2, label: '2分', description: '少量协助'),
          ShuangxiScoreOption(value: 3, label: '3分', description: '独立完成'),
        ],
      );
    }
    if (itemNo == 2) {
      return const ShuangxiAssessmentItem(
        itemNo: 2,
        itemCode: '1.1.2',
        itemTitle: '1.1.2 视觉追视能力',
        testItem: '视觉追视能力',
        domainCode: 'SENSORY',
        domainName: '感官知觉',
        skillCode: '1.1',
        skillName: '视觉的运用',
        scoreOptions: <ShuangxiScoreOption>[
          ShuangxiScoreOption(value: 0, label: '0分', description: '盲或视觉注意力短暂'),
          ShuangxiScoreOption(value: 1, label: '1分', description: '能注视物体五秒以上'),
          ShuangxiScoreOption(value: 2, label: '2分', description: '能追视移动物体'),
          ShuangxiScoreOption(value: 3, label: '3分', description: '能稳定追视移动小物体'),
        ],
      );
    }
    return const ShuangxiAssessmentItem(
      itemNo: 1,
      itemCode: '1.1.1',
      itemTitle: '1.1.1 视觉敏锐度',
      testItem: '视觉敏锐度',
      domainCode: 'SENSORY',
      domainName: '感官知觉',
      skillCode: '1.1',
      skillName: '视觉的运用',
      scoreOptions: <ShuangxiScoreOption>[
        ShuangxiScoreOption(value: 0, label: '0分', description: '盲或无视觉注意力'),
        ShuangxiScoreOption(
          value: 1,
          label: '1分',
          description: '只能看到眼前三十公分远的小物体',
        ),
        ShuangxiScoreOption(
          value: 2,
          label: '2分',
          description: '能看到眼前一至二公尺远的小物体',
        ),
        ShuangxiScoreOption(
          value: 3,
          label: '3分',
          description: '能看到眼前三公尺远的小物体',
        ),
      ],
    );
  }

  @override
  Future<ShuangxiDraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    saveDraftCount += 1;
    lastSavedDraftPayload = Map<String, dynamic>.from(payload);
    return _draftDetailFromPayload(payload, id: 101);
  }

  @override
  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  }) async {
    return AssessmentDraftPage.empty;
  }

  @override
  Future<ShuangxiDraftDetail> fetchDraftDetail(String token, int id) async {
    return _draftDetailFromPayload(<String, dynamic>{
      'id': id,
      'studentId': 1,
      'studentName': '测试学生',
      'examinerName': '测试老师',
      'birthDate': '2018-01-01',
      'assessmentDate': '2026-05-18',
      'itemScores': <String, int>{'1': 2},
    }, id: id);
  }

  @override
  Future<ShuangxiDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    return _draftDetailFromPayload(payload, id: _intFrom(payload['draftId']));
  }

  @override
  Future<void> submitDraft(String token, int draftId) async {
    submitDraftCount += 1;
  }

  @override
  Future<String> updateStudentGender(
    String token, {
    required int studentId,
    required String gender,
  }) async {
    updateStudentGenderCount += 1;
    lastUpdatedStudentId = studentId;
    lastUpdatedGender = gender;
    return gender;
  }

  @override
  Future<ShuangxiRecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    int studentId = 0,
    String assessmentDateEnd = '',
  }) async {
    if (previousItemScores.isEmpty || previousAssessmentDate.trim().isEmpty) {
      return ShuangxiRecordPage.empty;
    }
    return ShuangxiRecordPage(
      items: <ShuangxiRecordSummary>[
        ShuangxiRecordSummary(
          id: 201,
          studentId: studentId,
          studentName: '小明',
          assessmentCode: 'SHUANGXI_A',
          assessmentName: '双溪课程评量表A',
          birthDate: '2018-01-01',
          assessmentDate: previousAssessmentDate,
          examinerName: '陈老师',
          updatedTime: '2026-05-01 10:00:00',
        ),
      ],
      total: 1,
      current: pageIndex,
      size: pageSize,
    );
  }

  @override
  Future<ShuangxiRecordDetail> fetchRecordDetail(String token, int id) async {
    return ShuangxiRecordDetail(
      id: id,
      studentId: 1,
      studentName: '小明',
      assessmentCode: 'SHUANGXI_A',
      assessmentName: '双溪课程评量表A',
      birthDate: '2018-01-01',
      assessmentDate: previousAssessmentDate,
      examinerName: '陈老师',
      updatedTime: '2026-05-01 10:00:00',
      input: ShuangxiDraftInput(
        studentId: 1,
        studentName: '小明',
        examinerName: '陈老师',
        birthDate: '2018-01-01',
        assessmentDate: previousAssessmentDate,
        itemScores: Map<int, int>.from(previousItemScores),
      ),
    );
  }

  ShuangxiDraftDetail _draftDetailFromPayload(
    Map<String, dynamic> payload, {
    required int id,
  }) {
    final Map<int, int> scores = <int, int>{};
    final Map<int, String> remarks = <int, String>{};
    final Object? rawScores = payload['itemScores'];
    if (rawScores is Map) {
      for (final MapEntry<Object?, Object?> entry in rawScores.entries) {
        final int itemNo = int.tryParse('${entry.key}') ?? 0;
        final int score = int.tryParse('${entry.value}') ?? 0;
        if (itemNo > 0) {
          scores[itemNo] = score;
        }
      }
    }
    final Object? rawRemarks = payload['itemRemarks'];
    if (rawRemarks is Map) {
      for (final MapEntry<Object?, Object?> entry in rawRemarks.entries) {
        final int itemNo = int.tryParse('${entry.key}') ?? 0;
        final String remark = '${entry.value ?? ''}'.trim();
        if (itemNo > 0 && remark.isNotEmpty) {
          remarks[itemNo] = remark;
        }
      }
    }
    final Object? rawRemarkList = payload['itemRemarkList'];
    if (rawRemarkList is List) {
      for (final Object? raw in rawRemarkList) {
        if (raw is! Map) {
          continue;
        }
        final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
        final int itemNo = _intFrom(item['itemNo']);
        final String remark = '${item['remark'] ?? ''}'.trim();
        if (itemNo > 0 && remark.isNotEmpty) {
          remarks[itemNo] = remark;
        }
      }
    }
    final Object? rawList = payload['itemScoreList'];
    if (rawList is List) {
      for (final Object? raw in rawList) {
        if (raw is! Map) {
          continue;
        }
        final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
        final int itemNo = _intFrom(item['itemNo']);
        final int score = _intFrom(item['score']);
        if (itemNo > 0) {
          scores[itemNo] = score;
        }
        final String remark = '${item['remark'] ?? ''}'.trim();
        if (itemNo > 0 && remark.isNotEmpty) {
          remarks[itemNo] = remark;
        }
      }
    }
    return ShuangxiDraftDetail(
      id: id,
      studentId: _intFrom(payload['studentId']),
      studentName: '${payload['studentName'] ?? ''}',
      birthDate: '${payload['birthDate'] ?? ''}',
      assessmentDate: '${payload['assessmentDate'] ?? ''}',
      examinerName: '${payload['examinerName'] ?? ''}',
      answeredItemCount: scores.length,
      completionPercent: scores.isEmpty ? 0 : 100,
      updatedTime: '2026-05-18 10:00:00',
      progress: ShuangxiDraftProgress(
        itemCount: 209,
        answeredItemCount: scores.length,
        missingItemCount: 209 - scores.length,
        completionPercent: scores.isEmpty ? 0 : 100,
        complete: scores.length >= 209,
        canScore: scores.length >= 209,
        missingItemNos: <int>[for (int i = scores.length + 1; i <= 209; i++) i],
      ),
      input: ShuangxiDraftInput(
        studentId: _intFrom(payload['studentId']),
        studentName: '${payload['studentName'] ?? ''}',
        studentGender: '${payload['studentGender'] ?? ''}',
        examinerName: '${payload['examinerName'] ?? ''}',
        remark: '${payload['remark'] ?? ''}',
        birthDate: '${payload['birthDate'] ?? ''}',
        assessmentDate: '${payload['assessmentDate'] ?? ''}',
        itemScores: scores,
        itemRemarks: remarks,
      ),
    );
  }

  int _intFrom(Object? value) {
    if (value is int) {
      return value;
    }
    return int.tryParse('$value') ?? 0;
  }
}
