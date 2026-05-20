import 'dart:convert';

import 'package:assessment_pad_app/auth_client.dart';
import 'package:assessment_pad_app/assessment_scale_client.dart';
import 'package:assessment_pad_app/assessment_scale_category_page.dart';
import 'package:assessment_pad_app/assessment_report_list_page.dart';
import 'package:assessment_pad_app/erxin_assessment_client.dart';
import 'package:assessment_pad_app/erxin_assessment_page.dart';
import 'package:assessment_pad_app/home_client.dart';
import 'package:assessment_pad_app/iep_assessment_record_client.dart';
import 'package:assessment_pad_app/iep_plan_client.dart';
import 'package:assessment_pad_app/main.dart';
import 'package:assessment_pad_app/pad_date_range_picker.dart';
import 'package:assessment_pad_app/pep3_assessment_client.dart';
import 'package:assessment_pad_app/pep3_assessment_page.dart';
import 'package:assessment_pad_app/shuangxi_assessment_page.dart';
import 'package:assessment_pad_app/smart_timetable_page.dart';
import 'package:assessment_pad_app/timetable_client.dart';
import 'package:assessment_pad_app/vbmapp_assessment_client.dart';
import 'package:assessment_pad_app/vbmapp_assessment_page.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('login page opens the home dashboard after real login callback',
      (WidgetTester tester) async {
    final _FakeHomeClient homeClient = _FakeHomeClient();
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: homeClient,
        scaleClient: _FakeAssessmentScaleClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );

    expect(find.text('评估助手'), findsOneWidget);
    expect(find.text('机构账号登录'), findsOneWidget);
    expect(find.text('验证码登录'), findsNothing);

    await _enterWithCustomKeyboard(tester, 0, 'chenrui');
    await _enterWithCustomKeyboard(tester, 1, '123456');
    await tester.tap(find.text('登 录'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.textContaining('启明成长中心'), findsOneWidget);
    expect(find.text('开始测评'), findsOneWidget);
    expect(find.byIcon(Icons.refresh_rounded), findsWidgets);
    expect(homeClient.fetchCurrentSessionCalls, 1);
    expect(homeClient.fetchSummaryCalls, 1);

    await tester.tap(find.byTooltip('刷新首页'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(homeClient.fetchCurrentSessionCalls, 2);
    expect(homeClient.fetchSummaryCalls, 2);

    await tester.binding.handlePopRoute();
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.textContaining('启明成长中心'), findsOneWidget);
    expect(find.text('机构账号登录'), findsNothing);
  });

  testWidgets('login page redirects to home when token already exists',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        erxinClient: _FakeErxinAssessmentClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.textContaining('启明成长中心'), findsOneWidget);
    expect(find.text('机构账号登录'), findsNothing);
  });

  testWidgets('start assessment card opens assessment scales when tapping card',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        erxinClient: _FakeErxinAssessmentClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    await tester.tap(find.text('开始测评'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.text('PEP-3语言理解评核量表'), findsOneWidget);
  });

  testWidgets('assessment report list opens preview dialog from view action',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'mock-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: AssessmentReportListScreen(
              onBack: () {},
              scaleClient: _FakeAssessmentScaleClient(),
              recordClient: _FakePep3AssessmentClient(hasPreviousRecord: true),
              erxinRecordClient: _FakePep3AssessmentClient(),
              autismDevRecordClient: _FakePep3AssessmentClient(),
              shuangxiRecordClient: _FakePep3AssessmentClient(),
            ),
          ),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('张一鸣'), findsOneWidget);
    await tester.tap(find.text('查看'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('PEP-3评估报告'), findsOneWidget);
    expect(find.text('测验分数'), findsOneWidget);
    expect(find.text('分数+表现图'), findsNothing);
    expect(find.text('报告解读'), findsOneWidget);
    expect(find.text('暂无评估报告内容'), findsOneWidget);

    await tester.tap(find.text('打印'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.text('选择打印内容'), findsOneWidget);
    expect(find.byType(Checkbox), findsNWidgets(4));
    expect(find.text('未生成'), findsOneWidget);

    await tester.tap(find.text('取消'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('报告解读'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.text('报告解读尚未生成'), findsOneWidget);
    expect(find.text('生成解读'), findsWidgets);
  });

  testWidgets('assessment report list sorts by report timestamp only',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'mock-token',
    });
    final _FakePep3AssessmentClient operatedClient = _FakePep3AssessmentClient(
      hasPreviousRecord: true,
      previousRecord: const Pep3RecordSummary(
        id: 11,
        studentId: 3,
        studentName: '被操作记录',
        assessmentCode: 'PEP3',
        assessmentName: 'PEP-3',
        birthDate: '2021-03-01',
        assessmentDate: '2026-05-17',
        examinerName: '陈老师',
        createdTime: '2026-05-18T08:00:00',
        updatedTime: '2026-05-19T20:00:00',
      ),
    );
    final _FakePep3AssessmentClient newerReportClient =
        _FakePep3AssessmentClient(
      hasPreviousRecord: true,
      previousRecord: const Pep3RecordSummary(
        id: 22,
        studentId: 31,
        studentName: '新报告记录',
        assessmentCode: 'ERXIN2',
        assessmentName: '儿心量表-II',
        birthDate: '2022-05-11',
        assessmentDate: '2026-05-16',
        examinerName: '陈老师',
        createdTime: '2026-05-18T09:30:00',
        updatedTime: '2026-05-18T08:00:00',
      ),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: AssessmentReportListScreen(
              onBack: () {},
              scaleClient: _FakeAssessmentScaleClient(),
              recordClient: operatedClient,
              erxinRecordClient: newerReportClient,
              autismDevRecordClient: _FakePep3AssessmentClient(),
              shuangxiRecordClient: _FakePep3AssessmentClient(),
            ),
          ),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('被操作记录'), findsOneWidget);
    expect(find.text('新报告记录'), findsOneWidget);
    expect(
      tester.getTopLeft(find.text('新报告记录')).dy,
      lessThan(tester.getTopLeft(find.text('被操作记录')).dy),
    );
  });

  testWidgets('erxin report preview keeps AI interpretation tab non-empty',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeErxinAssessmentClient erxinClient = _FakeErxinAssessmentClient(
      interpretationDelay: const Duration(milliseconds: 180),
      interpretationFetchDelay: const Duration(milliseconds: 80),
      reportPdfBytes: Uint8List(0),
    );
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'mock-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: AssessmentReportListScreen(
              onBack: () {},
              scaleClient: _FakeAssessmentScaleClient(),
              recordClient: _FakePep3AssessmentClient(),
              erxinRecordClient: _FakePep3AssessmentClient(
                hasPreviousRecord: true,
                previousRecord: const Pep3RecordSummary(
                  id: 21,
                  studentId: 31,
                  studentName: '陈旭',
                  assessmentCode: 'ERXIN2',
                  assessmentName: '儿心量表-II',
                  birthDate: '2022-05-11',
                  assessmentDate: '2026-05-08',
                  examinerName: '陈老师',
                  updatedTime: '2026-05-08T10:00:00',
                ),
              ),
              autismDevRecordClient: _FakePep3AssessmentClient(),
              shuangxiRecordClient: _FakePep3AssessmentClient(),
              erxinClient: erxinClient,
            ),
          ),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('陈旭'), findsOneWidget);
    await tester.tap(find.text('查看'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('0岁～6岁儿童发育行为评估报告'), findsOneWidget);
    expect(find.text('评估结果记录'), findsOneWidget);
    expect(find.text('报告解读'), findsOneWidget);

    await tester.tap(find.text('打印'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.text('选择打印内容'), findsOneWidget);
    expect(find.byType(Checkbox), findsNWidgets(2));
    expect(find.text('未生成'), findsOneWidget);

    await tester.tap(find.text('取消'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('报告解读'));
    await tester.pump();

    expect(find.text('本次测评显示儿童整体发育水平需结合日常观察综合判断。'), findsNothing);
    expect(erxinClient.generateInterpretationCalls, 0);

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 900));
    await tester.pump();

    expect(find.text('报告解读尚未生成'), findsOneWidget);
    expect(find.text('生成解读'), findsWidgets);
    final Finder emptyAction = find.byKey(
      const ValueKey<String>('erxin-interpretation-empty-action'),
    );
    expect(emptyAction, findsOneWidget);
    expect(tester.getSize(emptyAction).width, closeTo(236, 0.1));
    expect(erxinClient.fetchInterpretationCalls, 1);
    expect(erxinClient.generateInterpretationCalls, 0);

    await tester.tap(find.text('生成解读').last);
    await tester.pump();

    expect(find.text('确认生成解读'), findsOneWidget);
    expect(find.text('正在读取儿心评估结果'), findsNothing);
    expect(erxinClient.generateInterpretationCalls, 0);

    await tester.tap(find.text('确认生成'));
    await tester.pump();

    expect(find.text('正在读取儿心评估结果'), findsOneWidget);
    expect(find.text('本次测评显示儿童整体发育水平需结合日常观察综合判断。'), findsNothing);

    await tester.pump(const Duration(milliseconds: 70));

    expect(find.text('AI 正在生成报告解读...'), findsOneWidget);
    expect(find.textContaining('综合解读'), findsOneWidget);

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('本次测评显示儿童整体发育水平需结合日常观察综合判断。'), findsOneWidget);
    expect(find.text('重新生成解读'), findsOneWidget);
    expect(erxinClient.generateInterpretationCalls, 1);

    await tester.tap(find.byIcon(Icons.close_rounded).last);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 900));
    await tester.pump();
    await tester.tap(find.text('查看'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();
    await tester.tap(find.text('报告解读'));
    await tester.pump();

    expect(find.text('AI 正在生成报告解读'), findsNothing);
    expect(find.text('AI 正在生成报告解读...'), findsNothing);
    expect(find.textContaining('正在读取已保存的报告解读'), findsOneWidget);

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('本次测评显示儿童整体发育水平需结合日常观察综合判断。'), findsOneWidget);
    expect(erxinClient.fetchInterpretationCalls, 2);
    expect(erxinClient.generateInterpretationCalls, 1);

    await tester.tap(find.text('评估结果记录'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 900));
    await tester.pump();
    await tester.tap(find.text('报告解读'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(erxinClient.fetchInterpretationCalls, 2);
    expect(erxinClient.generateInterpretationCalls, 1);

    await tester.tap(find.text('重新生成解读'));
    await tester.pump();

    expect(find.text('确认重新生成解读'), findsOneWidget);
    expect(find.text('正在读取儿心评估结果'), findsNothing);
    expect(erxinClient.generateInterpretationCalls, 1);

    await tester.tap(find.text('取消'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('确认重新生成解读'), findsNothing);
    expect(erxinClient.generateInterpretationCalls, 1);

    await tester.tap(find.text('重新生成解读'));
    await tester.pump();
    await tester.tap(find.text('确认重新生成'));
    await tester.pump();

    expect(find.text('正在读取儿心评估结果'), findsOneWidget);

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(erxinClient.generateInterpretationCalls, 2);
  });

  testWidgets('autismdev report list opens static report preview',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'mock-token',
    });
    final _FakePep3AssessmentClient autismDevClient = _FakePep3AssessmentClient(
      hasPreviousRecord: true,
      previousRecord: const Pep3RecordSummary(
        id: 31,
        studentId: 41,
        studentName: '林一',
        assessmentCode: 'AUTISMDEV',
        assessmentName: '孤独症儿童发展评估表',
        birthDate: '2021-02-01',
        assessmentDate: '2026-05-11',
        examinerName: '陈老师',
        updatedTime: '2026-05-11T10:00:00',
      ),
    );
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: AssessmentReportListScreen(
              onBack: () {},
              scaleClient: _FakeAssessmentScaleClient(),
              recordClient: _FakePep3AssessmentClient(),
              erxinRecordClient: _FakePep3AssessmentClient(),
              autismDevRecordClient: autismDevClient,
              shuangxiRecordClient: _FakePep3AssessmentClient(),
            ),
          ),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('林一'), findsOneWidget);
    await tester.tap(find.text('查看'));
    await tester.pump();
    await tester.pump();

    expect(find.text('孤独症儿童发展评估报告'), findsWidgets);
    expect(find.text('评估情况'), findsOneWidget);
    expect(find.text('3.1 发展能力计分汇总表'), findsOneWidget);
    expect(find.text('测评次数'), findsOneWidget);
    expect(find.text('量表版本'), findsNothing);
    expect(find.text('评估结果分析'), findsOneWidget);
    expect(find.text('训练效果'), findsOneWidget);
    expect(find.text('发展情况剖面图'), findsOneWidget);
    expect(find.text('情绪行为表现图'), findsOneWidget);
    expect(autismDevClient.fetchAutismDevResultAnalysisCalls, 1);

    await tester.tap(find.text('打印'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('选择打印内容'), findsOneWidget);
    expect(find.text('全选'), findsOneWidget);
    expect(find.text('未生成'), findsWidgets);
    expect(find.byType(Checkbox), findsNWidgets(6));
  });

  testWidgets('assessment report list opens Shuangxi report analysis tab',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'mock-token',
    });
    final _FakePep3AssessmentClient shuangxiClient = _FakePep3AssessmentClient(
      hasPreviousRecord: true,
      previousRecord: const Pep3RecordSummary(
        id: 51,
        studentId: 61,
        studentName: '双溪学生',
        assessmentCode: 'SHUANGXI_A',
        assessmentName: '双溪课程评量表A',
        birthDate: '2018-01-01',
        assessmentDate: '2026-05-18',
        examinerName: '陈老师',
        updatedTime: '2026-05-18T10:00:00',
      ),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: AssessmentReportListScreen(
              onBack: () {},
              scaleClient: _FakeAssessmentScaleClient(),
              recordClient: _FakePep3AssessmentClient(),
              erxinRecordClient: _FakePep3AssessmentClient(),
              autismDevRecordClient: _FakePep3AssessmentClient(),
              shuangxiRecordClient: shuangxiClient,
            ),
          ),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('双溪学生'), findsOneWidget);
    await tester.tap(find.text('查看'));
    await tester.pump();

    expect(find.text('双溪评估报告'), findsOneWidget);
    expect(find.text('发展侧面图'), findsOneWidget);
    expect(find.text('评量结果分析'), findsOneWidget);
    expect(shuangxiClient.downloadShuangxiDevelopmentProfilePdfCalls, 1);
    expect(shuangxiClient.fetchShuangxiResultAnalysisCalls, 1);
    await tester.pump();
    expect(find.text('暂无发展侧面图PDF'), findsOneWidget);

    await tester.tap(find.text('评量结果分析'));
    await tester.pump();

    expect(find.text('双溪心智障碍个别化教育课程（三）'), findsOneWidget);
    expect(find.text('评量结果分析表'), findsOneWidget);
    expect(find.text('领域\n（依优弱序）'), findsOneWidget);
    expect(find.text('现况分析'), findsOneWidget);
    expect(find.text('AI生成'), findsOneWidget);

    await tester.tap(find.text('打印'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('选择打印内容'), findsOneWidget);
    expect(find.byType(Checkbox), findsNWidgets(2));
    expect(find.text('未生成'), findsOneWidget);
  });

  testWidgets('Shuangxi report config saves with Shuangxi record client',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'mock-token',
    });
    final _FakePep3AssessmentClient pep3Client = _FakePep3AssessmentClient();
    final _FakePep3AssessmentClient shuangxiClient = _FakePep3AssessmentClient(
      hasPreviousRecord: true,
      previousRecord: const Pep3RecordSummary(
        id: 51,
        studentId: 61,
        studentName: '双溪学生',
        assessmentCode: 'SHUANGXI_A',
        assessmentName: '双溪课程评量表A',
        birthDate: '2018-01-01',
        assessmentDate: '2026-05-18',
        examinerName: '陈老师',
        updatedTime: '2026-05-18T10:00:00',
      ),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: AssessmentReportListScreen(
              onBack: () {},
              scaleClient: _FakeAssessmentScaleClient(),
              recordClient: pep3Client,
              erxinRecordClient: _FakePep3AssessmentClient(),
              autismDevRecordClient: _FakePep3AssessmentClient(),
              shuangxiRecordClient: shuangxiClient,
              staffClient: _FakeTimetableClient(),
            ),
          ),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('双溪学生'), findsOneWidget);
    await tester.tap(find.text('配置'));
    await tester.pumpAndSettle();

    expect(find.text('配置评估记录'), findsOneWidget);
    await tester.tap(find.text('保存'));
    await tester.pumpAndSettle();

    expect(shuangxiClient.updateRecordConfigCalls, 1);
    expect(shuangxiClient.lastUpdatedRecordConfigId, 51);
    expect(shuangxiClient.lastUpdatedRecordConfigExaminerName, '陈老师');
    expect(shuangxiClient.lastUpdatedRecordConfigAssessmentDate, '2026-05-18');
    expect(pep3Client.updateRecordConfigCalls, 0);
    expect(find.text('评估配置已保存'), findsOneWidget);
  });

  testWidgets('home header fallback does not show a fake institution',
      (WidgetTester tester) async {
    final HomeSummary summary = HomeSummary.fallback();
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: HomeHeader(
            session: HomeSession.fallback,
            weather: summary.weather,
            date: summary.date,
            weekday: summary.weekday,
            loading: false,
            errorMessage: null,
            onRefresh: () {},
            onLogout: () {},
          ),
        ),
      ),
    );

    expect(find.textContaining('启明成长中心'), findsNothing);
  });

  testWidgets('schedule card shows four skeleton rows while loading',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: ScheduleCard(
            items: <HomeScheduleItem>[],
            loading: true,
          ),
        ),
      ),
    );

    expect(find.byType(ScheduleSkeletonRow), findsNWidgets(4));
    expect(find.text('今日暂无排课'), findsNothing);
  });

  testWidgets('smart timetable shows structured skeleton while loading',
      (WidgetTester tester) async {
    final _FakeTimetableClient timetableClient = _FakeTimetableClient(
      timetableDelay: const Duration(milliseconds: 300),
    );
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );

    expect(
      find.byKey(const ValueKey<String>('smart-timetable-skeleton')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey<String>('smart-timetable-skeleton-board')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey<String>('smart-timetable-skeleton-row-0')),
      findsOneWidget,
    );
    expect(find.text('陈思语老师'), findsNothing);

    await tester.pump(const Duration(milliseconds: 350));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(
      find.byKey(const ValueKey<String>('smart-timetable-skeleton')),
      findsNothing,
    );
    expect(find.text('陈思语老师'), findsOneWidget);
  });

  testWidgets('switching period groups does not show skeleton again',
      (WidgetTester tester) async {
    final _FakeTimetableClient timetableClient = _FakeTimetableClient(
      timetableDelay: const Duration(milliseconds: 300),
    );
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );
    await tester.pump(const Duration(milliseconds: 350));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.byKey(const ValueKey<String>('smart-timetable-skeleton')),
        findsNothing);
    expect(find.text('A组'), findsOneWidget);

    await tester
        .tap(find.byKey(const ValueKey<String>('period-group-dropdown')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 900));
    await tester.pump();
    await tester.tap(
      find.byKey(const ValueKey<String>('period-group-option-group-c')),
    );
    await tester.pump(const Duration(milliseconds: 50));

    expect(find.byKey(const ValueKey<String>('smart-timetable-skeleton')),
        findsNothing);
    expect(find.text('C组'), findsOneWidget);
    expect(find.text('陈思语老师'), findsOneWidget);

    await tester.pump(const Duration(milliseconds: 350));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.byKey(const ValueKey<String>('smart-timetable-skeleton')),
        findsNothing);
    expect(find.text('周子涵老师'), findsOneWidget);
  });

  testWidgets('smart timetable layout does not overflow on wide viewport',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: _FakeTimetableClient()),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(tester.takeException(), isNull);
  });

  testWidgets('smart timetable error status does not overflow on wide viewport',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(
          timetableClient: _FakeTimetableClient(
            timetableErrorMessage: '排课日程接口响应超时，请检查网络',
          ),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 900));
    await tester.pump();

    expect(find.text('排课日程接口响应超时，请检查网络'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('smart timetable opens schedule detail dialog on lesson tap',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeTimetableClient timetableClient = _FakeTimetableClient();
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const ValueKey<String>('lesson-0-0')));
    await tester.pumpAndSettle();

    final Finder detailDialog =
        find.byKey(const ValueKey<String>('schedule-detail-dialog'));
    expect(detailDialog, findsOneWidget);
    expect(find.text('感统训练'), findsWidgets);
    expect(
      find.descendant(
        of: detailDialog,
        matching: find.textContaining('09:15 - 09:55'),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(
        of: detailDialog,
        matching: find.textContaining('上课教师'),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(of: detailDialog, matching: find.textContaining('课程')),
      findsOneWidget,
    );
    expect(
      find.descendant(
        of: detailDialog,
        matching: find.textContaining('上课学员'),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(
        of: detailDialog,
        matching: find.textContaining('试听学员'),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(
        of: detailDialog,
        matching: find.textContaining('请假学员'),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(
        of: detailDialog,
        matching: find.textContaining('对内备注'),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(
        of: detailDialog,
        matching: find.text('课前先做前庭唤醒'),
      ),
      findsOneWidget,
    );
    expect(find.text('A101'), findsNothing);
    expect(timetableClient.detailCalls, 1);
  });

  testWidgets(
      'smart timetable deletes current schedule from detail dialog and refreshes',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeTimetableClient timetableClient = _FakeTimetableClient();
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byKey(const ValueKey<String>('lesson-0-0')), findsOneWidget);

    await tester.tap(find.byKey(const ValueKey<String>('lesson-0-0')));
    await tester.pumpAndSettle();

    await tester.tap(
      find.byKey(const ValueKey<String>('schedule-detail-delete-current')),
    );
    await tester.pumpAndSettle();
    await tester.tap(
      find.byKey(const ValueKey<String>('schedule-delete-confirm-submit')),
    );
    await tester.pumpAndSettle();

    expect(timetableClient.deleteCalls, 1);
    expect(timetableClient.lastDeleteScope, ScheduleDeleteScope.current);
    expect(find.byKey(const ValueKey<String>('schedule-detail-dialog')),
        findsNothing);
    expect(find.text('感统训练'), findsNothing);
  });

  testWidgets('smart timetable filters schedules by call status',
      (WidgetTester tester) async {
    final _FakeTimetableClient timetableClient = _FakeTimetableClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('感统训练'), findsOneWidget);
    expect(find.text('语言认知课'), findsOneWidget);

    await tester.tap(
      find.byKey(const ValueKey<String>('smart-filter-call-status')),
    );
    await tester.pumpAndSettle();
    await tester.tap(
      find.byKey(
        const ValueKey<String>('smart-filter-option-callStatus-signed'),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('感统训练'), findsNothing);
    expect(find.text('语言认知课'), findsOneWidget);
  });

  testWidgets('home shortcut opens smart timetable page',
      (WidgetTester tester) async {
    final _FakeTimetableClient timetableClient = _FakeTimetableClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        timetableClient: timetableClient,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('排课日程'));
    await tester.pumpAndSettle();

    expect(find.text('排课日程'), findsWidgets);
    expect(find.text('智慧课表'), findsNothing);
    expect(find.text('时间课表'), findsNothing);
    expect(find.text('按固定时段查看老师一周排课'), findsNothing);
    expect(find.text('陈思语老师'), findsOneWidget);
    expect(find.text('A组'), findsOneWidget);

    final Rect timeRailRect =
        tester.getRect(find.byKey(const ValueKey<String>('smart-time-rail')));
    final Rect scheduleGridRect = tester
        .getRect(find.byKey(const ValueKey<String>('smart-schedule-grid')));
    final Rect boardRect = tester
        .getRect(find.byKey(const ValueKey<String>('smart-timetable-board')));
    expect((boardRect.left - timeRailRect.left).abs(), lessThan(.5));
    expect((boardRect.right - scheduleGridRect.right).abs(), lessThan(.5));
    expect((timeRailRect.top - scheduleGridRect.top).abs(), lessThan(.5));
    expect((timeRailRect.bottom - scheduleGridRect.bottom).abs(), lessThan(.5));
    expect((timeRailRect.right - scheduleGridRect.left).abs(), lessThan(.5));
    expect(
      (timeRailRect.height - scheduleGridRect.height).abs(),
      lessThan(.5),
    );

    final Finder scrollBody =
        find.byKey(const ValueKey<String>('smart-timetable-scroll-body'));
    final Rect stickyHeaderBefore = tester
        .getRect(find.byKey(const ValueKey<String>('smart-timetable-header')));
    final Rect weekHeaderBefore =
        tester.getRect(find.byKey(const ValueKey<String>('smart-week-header')));
    final Rect firstCellBefore =
        tester.getRect(find.byKey(const ValueKey<String>('schedule-cell-0-0')));

    await tester.drag(scrollBody, const Offset(0, -150));
    await tester.pumpAndSettle();

    final Rect stickyHeaderAfter = tester
        .getRect(find.byKey(const ValueKey<String>('smart-timetable-header')));
    final Rect weekHeaderAfter =
        tester.getRect(find.byKey(const ValueKey<String>('smart-week-header')));
    final Rect firstCellAfter =
        tester.getRect(find.byKey(const ValueKey<String>('schedule-cell-0-0')));
    expect(stickyHeaderAfter.top, stickyHeaderBefore.top);
    expect(weekHeaderAfter.top, weekHeaderBefore.top);
    expect(firstCellAfter.top, lessThan(firstCellBefore.top));
    expect(find.text('排课日程'), findsWidgets);
    expect(find.text('陈思语老师'), findsOneWidget);

    await tester.ensureVisible(
      find.byKey(const ValueKey<String>('schedule-cell-0-0')),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('陈思语老师').first);
    await tester.pumpAndSettle();

    expect(find.text('切换老师课表'), findsOneWidget);
    expect(find.text('周子涵老师'), findsNothing);
    expect(find.text('黄雨萱老师'), findsOneWidget);

    await tester.tap(find.text('黄雨萱老师'));
    await tester.pumpAndSettle();

    expect(find.text('黄雨萱老师'), findsOneWidget);
    expect(find.text('09:15 - 09:55'), findsOneWidget);
    expect(find.text('切换老师课表'), findsNothing);

    await tester
        .tap(find.byKey(const ValueKey<String>('period-group-dropdown')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('C组').last);
    await tester.pumpAndSettle();

    expect(find.text('C组'), findsOneWidget);
    expect(find.text('周子涵老师'), findsOneWidget);
    expect(find.text('08:30 - 09:10'), findsOneWidget);
    expect(find.text('09:15 - 09:55'), findsNothing);

    await tester.tap(find.text('周子涵老师').first);
    await tester.pumpAndSettle();

    expect(find.text('切换老师课表'), findsOneWidget);
    expect(find.text('陈思语老师'), findsNothing);

    await tester.tap(find.text('周子涵老师').last);
    await tester.pumpAndSettle();

    expect(find.text('周子涵老师'), findsOneWidget);
    expect(find.text('08:30 - 09:10'), findsOneWidget);
    expect(find.text('09:15 - 09:55'), findsNothing);
    expect(find.text('切换老师课表'), findsNothing);

    expect(find.byKey(const ValueKey<String>('lesson-0-0')), findsOneWidget);
    final Offset source =
        tester.getCenter(find.byKey(const ValueKey<String>('lesson-0-0')));
    final Offset target = tester
        .getCenter(find.byKey(const ValueKey<String>('schedule-cell-0-1')));

    final TestGesture earlyGesture = await tester.startGesture(source);
    await tester.pump(const Duration(milliseconds: 150));
    await earlyGesture.moveTo(target);
    await tester.pump();
    await earlyGesture.up();
    await tester.pumpAndSettle();

    expect(timetableClient.updateCalls, 0);
    expect(find.byKey(const ValueKey<String>('lesson-0-0')), findsOneWidget);
    expect(find.byKey(const ValueKey<String>('lesson-0-1')), findsNothing);

    final Offset dragSource =
        tester.getCenter(find.byKey(const ValueKey<String>('lesson-0-0')));
    final Offset dragTarget = tester
        .getCenter(find.byKey(const ValueKey<String>('schedule-cell-0-1')));
    final TestGesture gesture = await tester.startGesture(dragSource);
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pumpAndSettle();
    expect(find.text('可调课'), findsWidgets);
    await gesture.moveTo(dragTarget);
    await tester.pump();
    await gesture.moveTo(dragTarget);
    await tester.pump();
    await gesture.up();
    await tester.pumpAndSettle();

    final int lessonCount =
        find.byKey(const ValueKey<String>('lesson-0-0')).evaluate().length +
            find.byKey(const ValueKey<String>('lesson-0-1')).evaluate().length;
    expect(lessonCount, 1);
  });

  testWidgets('home shortcut opens training center page',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('训练中心'));
    await tester.pumpAndSettle();

    expect(find.text('训练中心'), findsWidgets);
    expect(find.text('推荐训练游戏'), findsOneWidget);
    expect(find.text('游戏合集'), findsOneWidget);
    expect(find.text('今日任务'), findsOneWidget);
    expect(find.text('最近记录'), findsOneWidget);
    expect(find.text('能力雷达'), findsNothing);
  });

  testWidgets('home shortcut opens IEP center static page',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: _FakeIepPlanClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.text('学员IEP队列'), findsOneWidget);
    expect(find.textContaining('陈旭 · 4岁0月'), findsOneWidget);
    expect(find.text('儿心量表 · 2026-05-07'), findsOneWidget);
    expect(find.text('已确认'), findsWidgets);
    expect(find.text('康复教学季度计划'), findsOneWidget);
    expect(find.text('能单脚站立保持平衡5秒以上'), findsOneWidget);
    expect(find.text('开始上课'), findsOneWidget);
    expect(find.text('计划参与者'), findsWidgets);
    expect(find.text('实施者'), findsWidgets);

    await tester.tap(find.text('5月'));
    await tester.pump();

    expect(find.text('5月 W1'), findsOneWidget);
    expect(find.text('5月 W5'), findsOneWidget);
    expect(find.text('W6'), findsNothing);
    expect(find.text('康复教学5月计划'), findsOneWidget);
    expect(find.text('训练内容'), findsOneWidget);
    expect(find.text('康复教学季度计划'), findsNothing);
    expect(
      tester.widget<Text>(find.text('IEP总计划')).style?.color,
      const Color(0xFF72594D),
    );

    await tester.tap(find.text('5月 W1'));
    await tester.pump();

    expect(
      tester.widget<Text>(find.text('5月')).style?.color,
      const Color(0xFF72594D),
    );
    expect(tester.widget<Text>(find.text('5月 W1')).style?.color, Colors.white);
    expect(find.text('康复教学周计划日记录卡5月第1周'), findsOneWidget);
    expect(find.text('完成情况'), findsOneWidget);
    expect(find.text('平衡木行走'), findsOneWidget);
    expect(find.text('康复教学5月计划'), findsNothing);

    await tester.tap(find.text('IEP总计划'));
    await tester.pump();
    await tester.tap(find.text('编辑周期'));
    await tester.pump();

    expect(find.text('编辑周期'), findsWidgets);
    expect(find.text('3个月周期'), findsOneWidget);
    expect(find.text('2026-07-31'), findsOneWidget);

    await tester.tap(find.text('周期开始'));
    await tester.pump();
    expect(find.text('选择周期开始日期'), findsOneWidget);
    expect(find.text('请选择周期开始日期，结束日期将按自然月自动计算'), findsOneWidget);

    await tester.tap(find.text('5').first);
    await tester.pump();
    await tester.tap(find.text('确定'));
    await tester.pump();

    expect(find.text('2026-05-05'), findsOneWidget);
    expect(find.text('2026-07-31'), findsOneWidget);

    await tester.tap(find.text('确认同步'));
    await tester.pump();

    expect(find.text('编辑周期'), findsOneWidget);
    expect(find.text('2026.05.05-2026.07.31'), findsOneWidget);
    expect(find.text('2026-05-05 至 2026-07-31'), findsOneWidget);

    await tester.tap(find.textContaining('提升动态平衡与协调能力'));
    await tester.pump();
    expect(find.text('编辑长期目标'), findsNothing);

    await tester.tap(find.text('姓名'));
    await tester.pump();
    await tester.tap(find.textContaining('提升动态平衡与协调能力'));
    await tester.pump();
    expect(find.text('编辑长期目标'), findsNothing);

    await tester.tap(find.textContaining('提升动态平衡与协调能力'));
    await tester.pump();
    expect(find.text('编辑长期目标'), findsOneWidget);
    expect(find.text('大肌肉 · 长期目标'), findsOneWidget);
    expect(find.text('保存并同步'), findsOneWidget);

    final Finder longGoalField = find.widgetWithText(
      TextField,
      '1. 提升动态平衡与协调能力，能在移动中稳定控制身体',
    );
    expect(longGoalField, findsOneWidget);
    await tester.enterText(longGoalField, '1. 能稳定完成平衡木行走与连续跳跃');
    await tester.pump();
    await tester.tap(find.text('仅保存当前表格'));
    await tester.pump();

    expect(find.text('编辑长期目标'), findsNothing);
    expect(find.textContaining('能稳定完成平衡木行走与连续跳跃'), findsOneWidget);

    await tester.tap(find.text('能单脚站立保持平衡5秒以上'));
    await tester.pump();
    await tester.tap(find.text('能单脚站立保持平衡5秒以上'));
    await tester.pump();

    expect(find.text('编辑短期目标'), findsOneWidget);
    expect(find.text('大肌肉 · 短期目标1'), findsOneWidget);
    expect(find.text('新增一条短期目标'), findsOneWidget);
    expect(find.text('当前短期目标'), findsOneWidget);
    expect(find.text('短期目标 1'), findsOneWidget);
    expect(find.text('短期目标 2'), findsNothing);
    expect(find.widgetWithText(TextField, '个训'), findsNothing);
    expect(find.text('个训'), findsWidgets);
    expect(find.text('集体课'), findsWidgets);

    await tester.tap(find.byIcon(Icons.delete_outline_rounded));
    await tester.pump();
    expect(find.text('短期目标 1'), findsNothing);
    await tester.tap(find.text('仅保存当前表格'));
    await tester.pump();

    expect(find.text('编辑短期目标'), findsNothing);
    expect(find.text('能单脚站立保持平衡5秒以上'), findsNothing);
    expect(find.text('能双脚连续向前跳5步以上'), findsOneWidget);

    await tester.tap(find.text('姓名'));
    await tester.pump();
    await tester.tap(find.text('能双脚连续向前跳5步以上'));
    await tester.pump();
    await tester.tap(find.text('能双脚连续向前跳5步以上'));
    await tester.pump();

    expect(find.text('编辑短期目标'), findsOneWidget);
    expect(find.text('大肌肉 · 短期目标1'), findsOneWidget);

    await tester.tap(find.text('新增一条短期目标'));
    await tester.pump();
    expect(find.text('短期目标 2'), findsOneWidget);

    final Finder newShortGoalField =
        find.byKey(const ValueKey<String>('short-goal-1-goal'));
    await tester.enterText(newShortGoalField, '能完成新增短期目标');
    await tester.pump();
    final Finder newLessonOption =
        find.byKey(const ValueKey<String>('short-goal-1-lesson-集体课'));
    await tester.ensureVisible(newLessonOption);
    await tester.tap(newLessonOption);
    await tester.pump();
    await tester.ensureVisible(find.text('仅保存当前表格'));
    await tester.tap(find.text('仅保存当前表格'));
    await tester.pump();

    expect(find.text('编辑短期目标'), findsNothing);
    expect(find.text('能完成新增短期目标'), findsOneWidget);
    expect(find.text('集体课'), findsWidgets);
  });

  test('IEP assessment record client includes Shuangxi records', () async {
    final List<String> paths = <String>[];
    final ApiIepAssessmentRecordClient client = ApiIepAssessmentRecordClient(
      educationBaseUrl: 'https://api.test',
      httpClient: MockClient((http.Request request) async {
        paths.add(request.url.path);
        final bool isShuangxi = request.url.path.contains('/shuangxi-a/');
        return http.Response.bytes(
          utf8.encode(jsonEncode(<String, dynamic>{
            'success': true,
            'data': <String, dynamic>{
              'items': isShuangxi
                  ? <Map<String, dynamic>>[
                      <String, dynamic>{
                        'id': 301,
                        'studentId': 88,
                        'studentName': '双溪学生',
                        'studentGender': '男',
                        'assessmentCode': 'SHUANGXI_A',
                        'assessmentName': '双溪课程评量表A',
                        'birthDate': '2019-12-01',
                        'assessmentDate': '2026-05-18',
                        'ageYears': 6,
                        'ageMonths': 5,
                        'ageDays': 17,
                        'examinerName': '陈老师',
                        'iepPlanStatus': '',
                        'updatedTime': '2026-05-18T13:30:00Z',
                      },
                    ]
                  : <Map<String, dynamic>>[],
              'total': isShuangxi ? 1 : 0,
            },
          })),
          200,
          headers: <String, String>{'content-type': 'application/json'},
        );
      }),
    );

    final IepAssessmentRecordPage page = await client.fetchRecordsPage(
      'token',
      pageIndex: 1,
      pageSize: 10,
    );

    expect(paths, contains('/api/v1/assessments/shuangxi-a/records/page'));
    expect(page.items, hasLength(1));
    expect(page.items.first.source, 'SHUANGXI');
    expect(page.items.first.assessmentCode, 'SHUANGXI_A');
    expect(page.items.first.studentName, '双溪学生');
  });

  test('IEP plan client maps Shuangxi records to Shuangxi AI task path',
      () async {
    String requestedPath = '';
    Map<String, dynamic> requestedPayload = <String, dynamic>{};
    final ApiIepPlanClient client = ApiIepPlanClient(
      educationBaseUrl: 'https://api.test',
      httpClient: MockClient((http.Request request) async {
        requestedPath = request.url.path;
        requestedPayload =
            Map<String, dynamic>.from(jsonDecode(request.body) as Map);
        return http.Response.bytes(
          utf8.encode(jsonEncode(<String, dynamic>{
            'success': true,
            'data': <String, dynamic>{
              'taskId': 'sx-task',
              'status': 'running',
              'durationMonths': 3,
            },
          })),
          200,
          headers: <String, String>{'content-type': 'application/json'},
        );
      }),
    );

    final IepPlanGenerationTask task = await client.createIepPlanGenerationTask(
      'token',
      record: const IepAssessmentRecordSummary(
        id: 301,
        source: 'SHUANGXI',
        studentId: 88,
        studentName: '双溪学生',
        assessmentCode: 'SHUANGXI_A',
        assessmentName: '双溪课程评量表A',
        birthDate: '2019-12-01',
        assessmentDate: '2026-05-18',
        examinerName: '陈老师',
        updatedTime: '2026-05-18T13:30:00Z',
      ),
      durationMonths: 3,
    );

    expect(
      requestedPath,
      '/api/v1/assessments/shuangxi-a/records/iep-plan/ai/tasks',
    );
    expect(requestedPayload['id'], 301);
    expect(requestedPayload['durationMonths'], 3);
    expect(task.taskId, 'sx-task');
  });

  testWidgets('IEP center lists Shuangxi record and starts IEP generation',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _ShuangxiCaptureIepPlanClient iepPlanClient =
        _ShuangxiCaptureIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeShuangxiIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.textContaining('双溪学生 · 6岁5月'), findsOneWidget);
    expect(find.text('双溪课程评量表A · 2026-05-18'), findsOneWidget);
    expect(find.text('双溪学生 暂无IEP计划'), findsOneWidget);

    await _tapIepAiGenerateAndConfirm(tester, last: true);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 30));

    expect(iepPlanClient.createdRecordSource, 'SHUANGXI');
    expect(iepPlanClient.createdRecordCode, 'SHUANGXI_A');
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();
    expect(find.text('能完成双溪课程目标'), findsOneWidget);
    if (find.text('知道了').evaluate().isNotEmpty) {
      await tester.tap(find.text('知道了'));
      await tester.pump();
    }
  });

  testWidgets('IEP center expands long and short goal cells without truncating',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: _LongGoalIepPlanClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.text(_LongGoalIepPlanClient.longGoal), findsOneWidget);
    expect(find.text(_LongGoalIepPlanClient.shortGoal), findsOneWidget);
    expect(find.text(_LongGoalIepPlanClient.shortGoal2), findsOneWidget);
    expect(find.text(_LongGoalIepPlanClient.shortGoal3), findsOneWidget);
    expect(
      tester.getRect(find.text(_LongGoalIepPlanClient.longGoal)).height,
      greaterThan(90),
    );
    expect(
      tester.getRect(find.text(_LongGoalIepPlanClient.shortGoal)).height,
      greaterThan(70),
    );
  });

  testWidgets('IEP center streams AI output then renders table and saves draft',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _EmptyThenGeneratedIepPlanClient iepPlanClient =
        _EmptyThenGeneratedIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakePendingIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.text('林一诺 暂无IEP计划'), findsOneWidget);
    expect(find.text('确认IEP'), findsNothing);

    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 30));

    expect(find.text('正在生成 林一诺 的IEP计划'), findsOneWidget);
    expect(find.text('正在读取评估和训练记录'), findsWidgets);
    expect(find.text('生成中'), findsWidgets);

    await tester.pump(const Duration(milliseconds: 260));
    expect(find.text('生成完成后将自动保存草稿，并切换为正式IEP表格预览'), findsOneWidget);
    expect(find.text('AI正在整理可预览内容'), findsOneWidget);

    await tester.pump(const Duration(milliseconds: 900));
    await tester.pump(const Duration(milliseconds: 200));
    await tester.pump();
    expect(find.text('大肌肉'), findsWidgets);
    expect(find.text('能独立跳跃3次'), findsOneWidget);

    expect(iepPlanClient.savePlanCalls, 0);
    expect(find.text('待确认'), findsWidgets);
    expect(find.text('确认IEP'), findsOneWidget);
  });

  testWidgets('IEP center hides regenerate by default for empty month and week',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: _FakeIepPlanClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.text('重新生成'), findsOneWidget);

    await tester.tap(find.text('6月'));
    await tester.pumpAndSettle();
    expect(find.text('重新生成'), findsNothing);

    await tester.tap(find.text('5月 W2'));
    await tester.pumpAndSettle();
    expect(find.text('重新生成'), findsNothing);
  });

  testWidgets('IEP center generates month plan from month tab',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeIepPlanClient iepPlanClient = _FakeIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    await tester.tap(find.text('6月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    expect(find.text('6月计划未生成'), findsOneWidget);

    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 220));
    expect(find.text('确认并生成'), findsOneWidget);
    await tester.tap(find.text('确认并生成'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 260));
    await tester.pump(const Duration(milliseconds: 900));
    await tester.pump();

    expect(iepPlanClient.generateMonthlyPlanCalls, 1);
    expect(iepPlanClient.saveMonthlyPlanCalls, 1);
    expect(find.text('康复教学5月计划'), findsOneWidget);
  });

  testWidgets('IEP center generates week plan from week tab',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _FakeIepPlanClient iepPlanClient = _FakeIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    final InkWell weekTab = tester.widget<InkWell>(
      find
          .ancestor(
            of: find.text('5月 W2'),
            matching: find.byType(InkWell),
          )
          .last,
    );
    weekTab.onTap?.call();
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    expect(find.text('5月第2周计划未生成'), findsOneWidget);

    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 260));
    await tester.pump(const Duration(milliseconds: 900));
    await tester.pump();

    expect(iepPlanClient.generateWeeklyPlanCalls, 1);
    expect(iepPlanClient.saveWeeklyPlanCalls, 1);
    expect(find.text('康复教学周计划日记录卡5月第1周'), findsOneWidget);
  });

  testWidgets('IEP center confirms before regenerating plan',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _ConfirmRegenerateIepPlanClient iepPlanClient =
        _ConfirmRegenerateIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    await tester.tap(find.text('重新生成'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 220));

    expect(find.text('确认重新生成IEP？'), findsOneWidget);
    expect(iepPlanClient.createTaskCalls, 0);

    await tester.tap(find.text('取消'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 220));

    expect(find.text('确认重新生成IEP？'), findsNothing);
    expect(iepPlanClient.createTaskCalls, 0);

    await tester.tap(find.text('重新生成'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 220));
    await tester.tap(find.text('确认重新生成'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    expect(iepPlanClient.createTaskCalls, 1);
  });

  testWidgets('IEP center reconnects existing AI task after stream disconnect',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _DisconnectThenResumeIepPlanClient iepPlanClient =
        _DisconnectThenResumeIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakePendingIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();
    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump(const Duration(milliseconds: 180));

    expect(find.textContaining('接口连接失败'), findsWidgets);
    expect(find.text('重试'), findsOneWidget);
    expect(iepPlanClient.createTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 1);

    await tester.tap(find.text('重试'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 900));

    expect(find.text('能恢复订阅并完成'), findsOneWidget);
    expect(iepPlanClient.createTaskCalls, 1);
    expect(iepPlanClient.fetchTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 2);
  });

  testWidgets(
      'IEP center restores existing AI task when switching away and back',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _DisconnectThenResumeIepPlanClient iepPlanClient =
        _DisconnectThenResumeIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    await tester.tap(find.textContaining('林一诺 · 4岁8月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.text('林一诺 暂无IEP计划'), findsOneWidget);
    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump(const Duration(milliseconds: 180));

    expect(find.textContaining('接口连接失败'), findsWidgets);
    expect(iepPlanClient.createTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 1);

    await tester.tap(find.textContaining('陈旭 · 4岁0月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 320));

    await tester.tap(find.textContaining('林一诺 · 4岁8月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 450));

    await tester.pump(const Duration(milliseconds: 900));

    expect(find.text('AI生成'), findsNothing);
    expect(find.text('能恢复订阅并完成'), findsOneWidget);
    expect(iepPlanClient.createTaskCalls, 1);
    expect(iepPlanClient.fetchTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 2);
  });

  testWidgets(
      'IEP center generate action resumes existing task instead of creating another one',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _DisconnectThenResumeIepPlanClient iepPlanClient =
        _DisconnectThenResumeIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    await tester.tap(find.textContaining('林一诺 · 4岁8月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    await tester.pump(const Duration(milliseconds: 200));

    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump(const Duration(milliseconds: 180));

    expect(find.textContaining('接口连接失败'), findsWidgets);
    expect(iepPlanClient.createTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 1);

    await tester.tap(find.textContaining('陈旭 · 4岁0月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 320));

    await tester.tap(find.textContaining('林一诺 · 4岁8月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 220));

    final Finder generateButton = find.text('AI生成');
    if (generateButton.evaluate().isNotEmpty) {
      await tester.tap(generateButton);
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 900));
    } else {
      await tester.pump(const Duration(milliseconds: 900));
    }

    expect(iepPlanClient.createTaskCalls, 1);
    expect(iepPlanClient.fetchTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 2);
    expect(find.text('能恢复订阅并完成'), findsOneWidget);
  });

  testWidgets(
      'IEP center resumes regenerate task after switching away and back',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _DisconnectThenResumeRegenerateIepPlanClient iepPlanClient =
        _DisconnectThenResumeRegenerateIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    await tester.tap(find.text('重新生成'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 220));
    await tester.tap(find.text('确认重新生成'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 180));

    expect(find.textContaining('接口连接失败'), findsWidgets);
    expect(iepPlanClient.createTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 1);

    await tester.tap(find.textContaining('林一诺 · 4岁8月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 320));

    await tester.tap(find.textContaining('陈旭 · 4岁0月'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 900));

    expect(iepPlanClient.createTaskCalls, 1);
    expect(iepPlanClient.fetchTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 2);
    expect(find.text('重新生成后恢复订阅并完成'), findsOneWidget);
  });

  testWidgets('IEP center auto resumes active task after cold start',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _ColdStartResumeIepPlanClient iepPlanClient =
        _ColdStartResumeIepPlanClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakePendingIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump(const Duration(milliseconds: 1200));

    expect(iepPlanClient.fetchActiveTaskCalls, 1);
    expect(iepPlanClient.createTaskCalls, 0);
    expect(iepPlanClient.fetchTaskCalls, 1);
    expect(iepPlanClient.watchTaskCalls, 1);
    expect(find.text('冷启动恢复成功'), findsOneWidget);
  });

  testWidgets('IEP center suppresses repeated streaming long goal prefixes',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakePendingIepAssessmentRecordClient(),
        iepPlanClient: _RepeatedLongGoalStreamIepPlanClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();
    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump(const Duration(milliseconds: 620));

    expect(find.textContaining('"longGoal":"提升动态'), findsOneWidget);
    expect(find.text('提升动态平衡能力'), findsNothing);

    await tester.pump(const Duration(milliseconds: 1400));
    expect(find.text('能双脚连续跳跃5次'), findsOneWidget);
    expect(find.text('集体课'), findsWidgets);
  });

  testWidgets('IEP center suppresses repeated streaming domain prefixes',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakePendingIepAssessmentRecordClient(),
        iepPlanClient: _RepeatedDomainStreamIepPlanClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();
    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump(const Duration(milliseconds: 700));

    expect(find.textContaining('"domain":"大肌'), findsOneWidget);
    expect(find.text('大肌肉'), findsNothing);

    await tester.pump(const Duration(milliseconds: 1200));
    expect(find.text('大肌肉'), findsOneWidget);
    expect(find.text('能双脚连续跳跃5次'), findsOneWidget);
  });

  testWidgets('IEP center keeps generated table when draft save fails',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final _EmptyThenGeneratedIepPlanClient iepPlanClient =
        _EmptyThenGeneratedIepPlanClient(failSave: true);
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakePendingIepAssessmentRecordClient(),
        iepPlanClient: iepPlanClient,
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();
    await _tapIepAiGenerateAndConfirm(tester);
    await tester.pump(const Duration(milliseconds: 1200));

    expect(iepPlanClient.savePlanCalls, 0);
    expect(find.text('能独立跳跃3次'), findsOneWidget);
    expect(find.textContaining('草稿自动保存失败'), findsNothing);
    expect(find.textContaining('FormatException'), findsNothing);
  });

  testWidgets('IEP center shows structured skeleton during route bootstrap',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(
          delay: const Duration(milliseconds: 300),
        ),
        iepPlanClient: _FakeIepPlanClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    expect(find.text('学员IEP队列'), findsOneWidget);
    expect(find.byKey(const ValueKey<String>('iep-queue-skeleton-0')),
        findsOneWidget);
    expect(find.byKey(const ValueKey<String>('iep-word-table-skeleton')),
        findsOneWidget);
    expect(find.text('正在加载评估记录'), findsNothing);
    expect(find.text('正在读取IEP计划'), findsNothing);
    expect(find.text('请选择左侧评估记录'), findsNothing);

    await tester.pump(const Duration(milliseconds: 700));
    expect(find.textContaining('陈旭 · 4岁0月'), findsOneWidget);
  });

  testWidgets('IEP center keeps skeleton until initial plan load completes',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        iepRecordClient: _FakeIepAssessmentRecordClient(),
        iepPlanClient: _SlowFirstIepPlanClient(
          planDelay: const Duration(milliseconds: 600),
        ),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('IEP中心'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pump();

    expect(find.textContaining('陈旭 · 4岁0月'), findsOneWidget);
    expect(find.byKey(const ValueKey<String>('iep-word-table-skeleton')),
        findsOneWidget);
    expect(find.text('正在读取IEP计划'), findsNothing);

    await tester.pump(const Duration(milliseconds: 700));
    expect(find.text('能单脚站立保持平衡5秒以上'), findsOneWidget);

    await tester.tap(find.textContaining('林一诺 · 4岁8月'));
    await tester.pump();

    expect(find.text('正在读取IEP计划'), findsOneWidget);
    expect(find.byKey(const ValueKey<String>('iep-word-table-skeleton')),
        findsNothing);
  });

  testWidgets('smart timetable detects availability after target selection',
      (WidgetTester tester) async {
    final _FakeTimetableClient timetableClient = _FakeTimetableClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );
    await tester.pumpAndSettle();

    await tester
        .tap(find.byKey(const ValueKey<String>('schedule-target-selector')));
    await tester.pumpAndSettle();
    expect(find.text('黄雨萱老师'), findsOneWidget);
    expect(find.text('助教A'), findsNothing);
    expect(find.text('不校验教室占用冲突'), findsOneWidget);
    expect(find.text('一楼 · 校验教室占用'), findsOneWidget);

    await tester.tap(find.text('A101'));
    await tester.pumpAndSettle();

    await tester.tap(
      find.byKey(const ValueKey<String>('schedule-target-one-to-one-a')),
    );
    await tester.pumpAndSettle();

    expect(find.text('张一鸣-个训课'), findsOneWidget);
    expect(timetableClient.validateCalls, greaterThan(0));
    expect(timetableClient.lastValidatedClassroomId, '101');
    expect(find.text('空闲时段(可排)'), findsWidgets);

    await tester
        .tap(find.byKey(const ValueKey<String>('schedule-target-clear')));
    await tester.pumpAndSettle();

    expect(find.text('空闲时段(可排)'), findsNothing);

    await tester
        .tap(find.byKey(const ValueKey<String>('schedule-target-selector')));
    await tester.pumpAndSettle();
    await tester.tap(
      find.byKey(const ValueKey<String>('schedule-target-one-to-one-a')),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const ValueKey<String>('empty-slot-0-1')));
    await tester.pumpAndSettle();

    expect(find.byKey(const ValueKey<String>('schedule-create-confirm-dialog')),
        findsOneWidget);
    await tester.tap(
      find.byKey(const ValueKey<String>('schedule-create-confirm-checkbox')),
    );
    await tester.pumpAndSettle();
    await tester.tap(
      find.byKey(const ValueKey<String>('schedule-create-confirm-submit')),
    );
    await tester.pumpAndSettle();

    expect(timetableClient.createCalls, 1);
    expect(timetableClient.lastCreatedClassroomId, '101');
    expect(find.byType(SnackBar), findsNothing);
    expect(
      find.byKey(const ValueKey<String>('schedule-top-message')),
      findsOneWidget,
    );

    await tester.tap(find.byKey(const ValueKey<String>('empty-slot-0-2')));
    await tester.pumpAndSettle();

    expect(
      find.byKey(const ValueKey<String>('schedule-create-confirm-dialog')),
      findsNothing,
    );
    expect(timetableClient.createCalls, 2);
  });

  testWidgets('smart timetable loads schedule options lazily',
      (WidgetTester tester) async {
    final _FakeTimetableClient timetableClient = _FakeTimetableClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );
    await tester.pumpAndSettle();

    expect(timetableClient.oneToOneTargetCalls, 0);
    expect(timetableClient.groupClassTargetCalls, 0);
    expect(timetableClient.assistantOptionCalls, 0);
    expect(timetableClient.classroomOptionCalls, 0);

    await tester
        .tap(find.byKey(const ValueKey<String>('schedule-target-selector')));
    await tester.pumpAndSettle();

    expect(timetableClient.oneToOneTargetCalls, 1);
    expect(timetableClient.groupClassTargetCalls, 1);
    expect(timetableClient.assistantOptionCalls, 1);
    expect(timetableClient.classroomOptionCalls, 1);
  });

  testWidgets('smart timetable blocks invalid availability slots',
      (WidgetTester tester) async {
    final _FakeTimetableClient timetableClient = _FakeTimetableClient(
      availabilityValid: false,
    );
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );
    await tester.pumpAndSettle();

    await tester
        .tap(find.byKey(const ValueKey<String>('schedule-target-selector')));
    await tester.pumpAndSettle();
    await tester.tap(
      find.byKey(const ValueKey<String>('schedule-target-one-to-one-a')),
    );
    await tester.pumpAndSettle();

    expect(find.text('老师冲突'), findsWidgets);

    await tester.tap(find.byKey(const ValueKey<String>('empty-slot-0-1')));
    await tester.pumpAndSettle();

    expect(timetableClient.createCalls, 0);
  });

  testWidgets('smart timetable loads full student and course filter options',
      (WidgetTester tester) async {
    final _FakeTimetableClient timetableClient = _FakeTimetableClient();
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      MaterialApp(
        home: SmartTimetablePage(timetableClient: timetableClient),
      ),
    );
    await tester.pumpAndSettle();

    await tester
        .tap(find.byKey(const ValueKey<String>('smart-filter-student')));
    await tester.pumpAndSettle();

    expect(
      find.byKey(const ValueKey<String>('smart-filter-option-student-王安全')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey<String>('smart-filter-option-student-张一鸣')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey<String>('smart-filter-option-student-孙吾空')),
      findsOneWidget,
    );

    await tester.tapAt(const Offset(20, 20));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey<String>('smart-filter-course')));
    await tester.pumpAndSettle();

    expect(
      find.byKey(const ValueKey<String>('smart-filter-option-course-个训课')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey<String>('smart-filter-option-course-感统训练')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey<String>('smart-filter-option-course-社交沟通课')),
      findsOneWidget,
    );
  });

  testWidgets('login page switches to qr login and back',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );

    await tester.tap(find.text('二维码登录'));
    await tester.pumpAndSettle();

    expect(find.text('二维码登录'), findsOneWidget);
    expect(find.text('账号登录'), findsOneWidget);
    expect(find.text('刷新二维码'), findsOneWidget);

    await tester.tap(find.text('账号登录'));
    await tester.pumpAndSettle();

    expect(find.text('机构账号登录'), findsOneWidget);
    expect(find.text('忘记密码'), findsOneWidget);
  });

  testWidgets('login page shows styled institution picker for multiple options',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _MultiInstitutionAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );

    await _enterWithCustomKeyboard(tester, 0, 'chenrui');
    await _enterWithCustomKeyboard(tester, 1, '123456');
    await tester.tap(find.text('登 录'));
    await tester.pumpAndSettle();

    expect(find.text('选择登录机构'), findsOneWidget);
    expect(find.text('当前账号关联多个机构，请选择本次进入的后台'), findsOneWidget);
    expect(find.text('启明成长中心'), findsOneWidget);
    expect(find.text('南山训练中心'), findsOneWidget);
    expect(find.text('超管'), findsOneWidget);
    expect(find.text('正常'), findsOneWidget);

    await tester.tap(find.text('南山训练中心'));
    await tester.pumpAndSettle();

    expect(find.textContaining('启明成长中心'), findsOneWidget);
  });

  testWidgets('wrong password does not open institution picker',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _PasswordCheckingAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );

    await _enterWithCustomKeyboard(tester, 0, 'chenrui');
    await _enterWithCustomKeyboard(tester, 1, 'wrong123');
    await tester.tap(find.text('登 录'));
    await tester.pumpAndSettle();

    expect(find.text('登录失败,用户名或密码错误'), findsOneWidget);
    expect(find.text('选择登录机构'), findsNothing);
  });

  testWidgets('desktop login fields use native input without custom keyboard',
      (WidgetTester tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
    try {
      SharedPreferences.setMockInitialValues(<String, Object>{});
      await tester.pumpWidget(
        AssessmentPadApp(
          authClient: _FakeAuthClient(),
          homeClient: _FakeHomeClient(),
          scaleClient: _FakeAssessmentScaleClient(),
          timetableClient: _FakeTimetableClient(),
        ),
      );

      await tester.tap(find.byType(TextField).first);
      await tester.pumpAndSettle();

      expect(find.byKey(const ValueKey<String>('login-key-1')), findsNothing);

      await tester.enterText(find.byType(TextField).first, 'chenrui');
      await tester.pumpAndSettle();

      expect(find.text('chenrui'), findsOneWidget);
    } finally {
      debugDefaultTargetPlatformOverride = null;
    }
  });

  testWidgets('scale search keyboard shows Chinese candidates for pinyin',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('搜索量表名称 / 编码'));
    await tester.pumpAndSettle();

    expect(find.text('输入拼音后在这里选择汉字候选'), findsOneWidget);
    final double idleKeyboardHeight = tester
        .getRect(find.byKey(const ValueKey<String>('scale-search-keyboard')))
        .height;

    await tester.tap(find.byKey(const ValueKey<String>('scale-search-key-x')));
    await tester.pumpAndSettle();
    await tester.pump(const Duration(milliseconds: 250));

    expect(find.text('暂无候选，继续输入或直接搜索编码'), findsOneWidget);
    final double emptyCandidateKeyboardHeight = tester
        .getRect(find.byKey(const ValueKey<String>('scale-search-keyboard')))
        .height;
    expect(emptyCandidateKeyboardHeight, idleKeyboardHeight);

    await tester.tap(find.byKey(const ValueKey<String>('scale-search-key-清空')));
    await tester.pumpAndSettle();
    await tester.pump(const Duration(milliseconds: 250));

    for (final String key in <String>['y', 'u', 'y', 'a', 'n']) {
      await tester.tap(find.byKey(ValueKey<String>('scale-search-key-$key')));
      await tester.pumpAndSettle();
      await tester.pump(const Duration(milliseconds: 250));
    }

    expect(find.text('1.语言'), findsOneWidget);
    final double candidateKeyboardHeight = tester
        .getRect(find.byKey(const ValueKey<String>('scale-search-keyboard')))
        .height;
    expect(candidateKeyboardHeight, idleKeyboardHeight);
    final Text searchText = tester.widget<Text>(
      find.byKey(const ValueKey<String>('scale-search-display-text')),
    );
    expect(searchText.data, '语言');

    await tester
        .tap(find.byKey(const ValueKey<String>('scale-search-key-1.语言')));
    await tester.pumpAndSettle();
    await tester.pump(const Duration(milliseconds: 250));

    expect(find.text('1.语言'), findsNothing);
    expect(find.text('PEP-3语言理解评核量表'), findsOneWidget);
  });

  testWidgets('scale search clear icon keeps custom keyboard open',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester
        .tap(find.byKey(const ValueKey<String>('scale-search-display-text')));
    await tester.pumpAndSettle();

    await tester
        .tap(find.byKey(const ValueKey<String>('scale-search-key-PEP-3')));
    await tester.pumpAndSettle();
    await tester.pump(const Duration(milliseconds: 250));

    Text searchText = tester.widget<Text>(
      find.byKey(const ValueKey<String>('scale-search-display-text')),
    );
    expect(searchText.data, 'PEP-3');
    expect(find.byKey(const ValueKey<String>('scale-search-keyboard')),
        findsOneWidget);

    await tester.tap(find.byKey(const ValueKey<String>('scale-search-clear')));
    await tester.pumpAndSettle();

    searchText = tester.widget<Text>(
      find.byKey(const ValueKey<String>('scale-search-display-text')),
    );
    expect(searchText.data, '搜索量表名称 / 编码');
    expect(find.byKey(const ValueKey<String>('scale-search-keyboard')),
        findsOneWidget);
  });

  testWidgets('scale category page shows structured skeleton while loading',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              categoriesDelay: const Duration(milliseconds: 300),
              libraryDelay: const Duration(milliseconds: 300),
            ),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pump();

    expect(find.byKey(const ValueKey<String>('category-skeleton-0')),
        findsOneWidget);
    final Finder firstScaleSkeleton =
        find.byKey(const ValueKey<String>('scale-card-skeleton-0'));
    expect(firstScaleSkeleton, findsOneWidget);
    expect(find.text('正在加载量表'), findsOneWidget);
    expect(find.text('可用'), findsOneWidget);
    expect(find.text('全部'), findsOneWidget);
    final Size skeletonCardSize = tester.getSize(firstScaleSkeleton);

    await tester.pump(const Duration(milliseconds: 200));
    await tester.pump(const Duration(milliseconds: 320));
    await tester.pumpAndSettle();

    expect(find.text('PEP-3语言理解评核量表'), findsOneWidget);
    final Finder loadedScaleCard = find.ancestor(
      of: find.text('PEP-3语言理解评核量表'),
      matching: find.byWidgetPredicate(
        (Widget widget) => widget.runtimeType.toString() == '_ScaleCard',
      ),
    );
    expect(loadedScaleCard, findsOneWidget);
    expect(
      tester.getSize(loadedScaleCard).height,
      closeTo(skeletonCardSize.height, .1),
    );
  });

  testWidgets('disabled scale start button opens student selector',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('开始测评').last);
    await tester.pumpAndSettle();

    expect(find.text('开始测评前，请先选择本次测评对象。'), findsOneWidget);
    expect(find.text('确认选择并进入测评'), findsOneWidget);
  });

  testWidgets('student selector opens immediately before candidates finish',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              studentCandidatesDelay: const Duration(milliseconds: 500),
            ),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pump(const Duration(milliseconds: 50));

    expect(find.text('选择学员'), findsOneWidget);
    expect(find.text('开始测评前，请先选择本次测评对象。'), findsOneWidget);
    expect(
      find.byKey(const ValueKey<String>('student-selector-skeleton')),
      findsOneWidget,
    );
    expect(find.byType(CircularProgressIndicator), findsNothing);
    expect(find.text('儿童姓名'), findsOneWidget);
    expect(find.text('最近测评'), findsOneWidget);
    expect(find.text('张一鸣'), findsNothing);

    await tester.pump(const Duration(milliseconds: 520));
    await tester.pumpAndSettle();

    expect(
      find.byKey(const ValueKey<String>('student-selector-skeleton')),
      findsNothing,
    );
    expect(find.text('张一鸣'), findsOneWidget);
  });

  testWidgets('student selector jumps by alphabet index',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              studentCandidates: const <AssessmentStudentCandidate>[
                AssessmentStudentCandidate(
                  id: 11,
                  shortName: '陈',
                  name: '陈小北',
                  avatarUrl: '',
                  gender: '男',
                  age: '4岁',
                  birthDate: '',
                  contactPhone: '妈妈 136****0011',
                  latestAssessment: '未测评',
                ),
                AssessmentStudentCandidate(
                  id: 12,
                  shortName: '李',
                  name: '李开心',
                  avatarUrl: '',
                  gender: '女',
                  age: '5岁',
                  birthDate: '',
                  contactPhone: '爸爸 136****0012',
                  latestAssessment: '未测评',
                ),
                AssessmentStudentCandidate(
                  id: 13,
                  shortName: '马',
                  name: '马一诺',
                  avatarUrl: '',
                  gender: '女',
                  age: '3岁',
                  birthDate: '',
                  contactPhone: '妈妈 136****0013',
                  latestAssessment: '未测评',
                ),
                AssessmentStudentCandidate(
                  id: 14,
                  shortName: '王',
                  name: '王安全',
                  avatarUrl: '',
                  gender: '男',
                  age: '',
                  birthDate: '',
                  contactPhone: '爸爸 136****0014',
                  latestAssessment: '未测评',
                ),
                AssessmentStudentCandidate(
                  id: 15,
                  shortName: '张',
                  name: '张一鸣',
                  avatarUrl: '',
                  gender: '男',
                  age: '5岁2个月',
                  birthDate: '',
                  contactPhone: '妈妈 136****0015',
                  latestAssessment: '未测评',
                ),
                AssessmentStudentCandidate(
                  id: 16,
                  shortName: '赵',
                  name: '赵可欣',
                  avatarUrl: '',
                  gender: '女',
                  age: '6岁',
                  birthDate: '',
                  contactPhone: '爸爸 136****0016',
                  latestAssessment: '未测评',
                ),
              ],
            ),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();

    expect(find.text('赵可欣'), findsNothing);

    await tester.drag(
      find.byKey(const ValueKey<String>('student-letter-index-scroll')),
      const Offset(0, -240),
    );
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey<String>('student-letter-index-Z')),
        findsOneWidget);

    await tester
        .tap(find.byKey(const ValueKey<String>('student-letter-index-Z')));
    await tester.pumpAndSettle();

    expect(find.text('赵可欣'), findsOneWidget);
  });

  testWidgets('student selector switches enrolled and intention tabs',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeAssessmentScaleClient scaleClient = _FakeAssessmentScaleClient(
      studentCandidatesByStatus: const <int, List<AssessmentStudentCandidate>>{
        AssessmentStudentStatuses.enrolled: <AssessmentStudentCandidate>[
          AssessmentStudentCandidate(
            id: 21,
            shortName: '张',
            name: '张一鸣',
            avatarUrl: '',
            gender: '男',
            age: '5岁2个月',
            birthDate: '',
            contactPhone: '妈妈 136****0021',
            latestAssessment: '未测评',
            studentStatus: AssessmentStudentStatuses.enrolled,
            studentStatusText: '在读学员',
          ),
        ],
        AssessmentStudentStatuses.intention: <AssessmentStudentCandidate>[
          AssessmentStudentCandidate(
            id: 22,
            shortName: '李',
            name: '李小满',
            avatarUrl: '',
            gender: '女',
            age: '4岁',
            birthDate: '',
            contactPhone: '妈妈 136****0022',
            latestAssessment: '未测评',
            studentStatus: AssessmentStudentStatuses.intention,
            studentStatusText: '意向学员',
          ),
        ],
      },
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: scaleClient,
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();

    expect(find.text('在读学员'), findsOneWidget);
    expect(find.text('意向学员'), findsOneWidget);
    expect(find.text('张一鸣'), findsOneWidget);
    expect(find.text('李小满'), findsNothing);
    expect(scaleClient.requestedStudentStatuses,
        contains(AssessmentStudentStatuses.enrolled));

    await tester.tap(
      find.byKey(
        const ValueKey<String>(
          'student-status-tab-${AssessmentStudentStatuses.intention}',
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('李小满'), findsOneWidget);
    expect(find.text('张一鸣'), findsNothing);
    expect(scaleClient.requestedStudentStatuses,
        contains(AssessmentStudentStatuses.intention));
  });

  testWidgets('selected student echoes unknown age when age is empty',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('王安全'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    expect(find.text('王安全 * 未知'), findsOneWidget);
  });

  testWidgets('student selector and echo show day-level age from birth date',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final DateTime today = DateTime.now();
    final DateTime birthDate = DateTime(today.year - 4, today.month, today.day)
        .subtract(const Duration(days: 3));

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              studentCandidates: <AssessmentStudentCandidate>[
                AssessmentStudentCandidate(
                  id: 24,
                  shortName: '周',
                  name: '周小天',
                  avatarUrl: '',
                  gender: '男',
                  age: '4岁',
                  birthDate: _formatDateDashForTest(birthDate),
                  contactPhone: '妈妈 136****0024',
                  latestAssessment: '未测评',
                ),
              ],
            ),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();

    expect(find.text('4岁3天'), findsOneWidget);

    await tester.tap(find.text('周小天'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    expect(find.text('周小天 * 4岁3天'), findsOneWidget);
  });

  testWidgets('Shuangxi gender update syncs selected student selector cache',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    ShuangxiAssessmentLaunchArgs? openedArgs;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              scaleItems: const <AssessmentScaleItem>[_shuangxiAScaleItem],
              studentCandidates: const <AssessmentStudentCandidate>[
                AssessmentStudentCandidate(
                  id: 41,
                  shortName: '顾',
                  name: '顾未知',
                  avatarUrl: '',
                  gender: '-',
                  age: '7岁',
                  birthDate: '2019-05-18',
                  contactPhone: '妈妈 136****0041',
                  latestAssessment: '未测评',
                ),
              ],
            ),
            onBack: () {},
          ),
        ),
        onGenerateRoute: (RouteSettings settings) {
          if (settings.name == '/shuangxi-a-assessment') {
            final ShuangxiAssessmentLaunchArgs args =
                settings.arguments! as ShuangxiAssessmentLaunchArgs;
            openedArgs = args;
            return MaterialPageRoute<void>(
              settings: settings,
              builder: (BuildContext context) => Scaffold(
                body: Center(
                  child: ElevatedButton(
                    key: const ValueKey<String>(
                      'fake-shuangxi-sync-gender',
                    ),
                    onPressed: () {
                      args.onStudentGenderUpdated?.call('男');
                      Navigator.of(context).pop();
                    },
                    child: const Text('同步性别并返回'),
                  ),
                ),
              ),
            );
          }
          return null;
        },
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('顾未知'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('开始测评').last);
    await tester.pumpAndSettle();

    expect(openedArgs?.studentId, 41);
    expect(openedArgs?.studentGender, '-');

    await tester.tap(find.byKey(const ValueKey<String>(
      'fake-shuangxi-sync-gender',
    )));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('点击切换学员'));
    await tester.pumpAndSettle();

    expect(find.text('顾未知'), findsOneWidget);
    expect(find.text('男'), findsOneWidget);
    expect(find.text('-'), findsNothing);
  });

  testWidgets('scale launch asks for birth date and updates student profile',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeAssessmentScaleClient scaleClient = _FakeAssessmentScaleClient(
      scaleItems: const <AssessmentScaleItem>[_shuangxiAScaleItem],
      studentCandidates: const <AssessmentStudentCandidate>[
        AssessmentStudentCandidate(
          id: 42,
          shortName: '林',
          name: '林缺生日',
          avatarUrl: '',
          gender: '男',
          age: '',
          birthDate: '',
          contactPhone: '妈妈 136****0042',
          latestAssessment: '未测评',
        ),
      ],
    );
    ShuangxiAssessmentLaunchArgs? openedArgs;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: scaleClient,
            onBack: () {},
          ),
        ),
        onGenerateRoute: (RouteSettings settings) {
          if (settings.name == '/shuangxi-a-assessment') {
            final ShuangxiAssessmentLaunchArgs args =
                settings.arguments! as ShuangxiAssessmentLaunchArgs;
            openedArgs = args;
            return MaterialPageRoute<void>(
              settings: settings,
              builder: (BuildContext context) => const Scaffold(
                body: Center(child: Text('双溪测评已打开')),
              ),
            );
          }
          return null;
        },
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('林缺生日'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('开始测评').last);
    await tester.pumpAndSettle();

    expect(find.text('请先设置出生日期'), findsOneWidget);
    expect(openedArgs, isNull);

    await tester.tap(find.text('选择出生日期'));
    await tester.pumpAndSettle();
    final DateTime today = DateTime.now();
    expect(find.text('${today.year}年'), findsOneWidget);
    expect(find.text('${today.month}月'), findsOneWidget);
    expect(find.text('2020年5月'), findsNothing);
    expect(find.text('未选择'), findsOneWidget);
    await tester.tap(find.text('确定').last);
    await tester.pumpAndSettle();
    expect(find.text('未选择'), findsOneWidget);

    await tester.tap(find.text('${today.day}').first);
    await tester.pumpAndSettle();
    await tester.tap(find.text('确定').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('更新并开始测评'));
    await tester.pumpAndSettle();

    expect(scaleClient.updateStudentBirthDateCount, 1);
    expect(scaleClient.lastUpdatedStudentId, 42);
    expect(scaleClient.lastUpdatedBirthDate, isNotEmpty);
    expect(openedArgs?.studentId, 42);
    expect(openedArgs?.birthDate, scaleClient.lastUpdatedBirthDate);
    expect(find.text('双溪测评已打开'), findsOneWidget);
  });

  testWidgets('pad date picker disables future dates when requested',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    DateTime? picked;

    await tester.pumpWidget(
      MaterialApp(
        home: Builder(
          builder: (BuildContext context) {
            return Scaffold(
              body: Center(
                child: ElevatedButton(
                  onPressed: () async {
                    picked = await showPadDatePicker(
                      context: context,
                      initialDate: DateTime(2026, 5, 18),
                      today: DateTime(2026, 5, 18),
                      minDate: DateTime(2026, 5, 1),
                      maxDate: DateTime(2026, 6, 30),
                      disableFutureDates: true,
                    );
                  },
                  child: const Text('打开日期'),
                ),
              ),
            );
          },
        ),
      ),
    );

    await tester.tap(find.text('打开日期'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('19'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确定').last);
    await tester.pumpAndSettle();

    expect(picked, DateTime(2026, 5, 18));
  });

  testWidgets('pad date picker jumps by year and month',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    DateTime? picked;

    await tester.pumpWidget(
      MaterialApp(
        home: Builder(
          builder: (BuildContext context) {
            return Scaffold(
              body: Center(
                child: ElevatedButton(
                  onPressed: () async {
                    picked = await showPadDatePicker(
                      context: context,
                      initialDate: DateTime(2026, 5, 18),
                      today: DateTime(2026, 5, 18),
                      minDate: DateTime(2020, 1, 1),
                      maxDate: DateTime(2026, 5, 18),
                      disableFutureDates: true,
                      initiallySelectDate: false,
                    );
                  },
                  child: const Text('打开日期'),
                ),
              ),
            );
          },
        ),
      ),
    );

    await tester.tap(find.text('打开日期'));
    await tester.pumpAndSettle();

    expect(find.text('未选择'), findsOneWidget);
    await tester.tap(find.text('2026年'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('2021'));
    await tester.pumpAndSettle();

    expect(find.text('2021年'), findsOneWidget);
    await tester.tap(find.text('3月'));
    await tester.pumpAndSettle();
    expect(find.text('2021年'), findsOneWidget);
    expect(find.text('3月'), findsOneWidget);

    await tester.tap(find.text('12'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确定').last);
    await tester.pumpAndSettle();

    expect(picked, DateTime(2021, 3, 12));
  });

  testWidgets('ERXin scale blocks launch when student is over six years old',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              scaleItems: const <AssessmentScaleItem>[_erxinScaleItem],
              studentCandidates: const <AssessmentStudentCandidate>[
                AssessmentStudentCandidate(
                  id: 31,
                  shortName: '陈',
                  name: '陈超龄',
                  avatarUrl: '',
                  gender: '男',
                  age: '7岁',
                  birthDate: '2018-01-01',
                  contactPhone: '妈妈 136****0031',
                  latestAssessment: '未测评',
                ),
              ],
            ),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('陈超龄'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('开始测评').last);
    await tester.pump();

    expect(find.textContaining('超过6岁'), findsOneWidget);
    expect(find.text('儿心量表-II 测评'), findsNothing);
  });

  testWidgets('PEP3 scale blocks launch using backend age max months',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              scaleItems: const <AssessmentScaleItem>[_pep3ScaleItem],
              studentCandidates: const <AssessmentStudentCandidate>[
                AssessmentStudentCandidate(
                  id: 32,
                  shortName: '李',
                  name: '李明轩',
                  avatarUrl: '',
                  gender: '男',
                  age: '7岁6个月',
                  birthDate: '2018-11-13',
                  contactPhone: '妈妈 136****0032',
                  latestAssessment: '未测评',
                ),
              ],
            ),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('李明轩'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('开始测评').last);
    await tester.pump();

    expect(find.textContaining('超过7岁5个月'), findsOneWidget);
    expect(find.text('PEP-3 测评工作台'), findsNothing);
  });

  testWidgets(
      'Autism development scale blocks launch using backend age max months',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              scaleItems: const <AssessmentScaleItem>[_autismDevScaleItem],
              studentCandidates: const <AssessmentStudentCandidate>[
                AssessmentStudentCandidate(
                  id: 33,
                  shortName: '赵',
                  name: '赵超龄',
                  avatarUrl: '',
                  gender: '女',
                  age: '7岁',
                  birthDate: '2018-01-01',
                  contactPhone: '妈妈 136****0033',
                  latestAssessment: '未测评',
                ),
              ],
            ),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('赵超龄'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('开始测评').last);
    await tester.pump();

    expect(find.textContaining('超过6岁'), findsOneWidget);
    expect(find.text('孤独症儿童发展评估表测评工作台'), findsNothing);
  });

  testWidgets('VB-MAPP scale opens pad assessment route',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    VbmappAssessmentLaunchArgs? openedArgs;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              scaleItems: const <AssessmentScaleItem>[_vbmappScaleItem],
              studentCandidates: const <AssessmentStudentCandidate>[
                AssessmentStudentCandidate(
                  id: 51,
                  shortName: '王',
                  name: '王小语',
                  avatarUrl: '',
                  gender: '女',
                  age: '3岁',
                  birthDate: '2023-01-01',
                  contactPhone: '妈妈 136****0051',
                  latestAssessment: '未测评',
                ),
              ],
            ),
            onBack: () {},
          ),
        ),
        onGenerateRoute: (RouteSettings settings) {
          if (settings.name == '/vbmapp-assessment') {
            openedArgs = settings.arguments! as VbmappAssessmentLaunchArgs;
            return MaterialPageRoute<void>(
              settings: settings,
              builder: (BuildContext context) => const Scaffold(
                body: Center(child: Text('VB-MAPP测评已打开')),
              ),
            );
          }
          return null;
        },
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('未选择学员'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('王小语'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('开始测评').last);
    await tester.pumpAndSettle();

    expect(openedArgs?.studentId, 51);
    expect(openedArgs?.scaleName, _vbmappScaleItem.name);
    expect(find.text('VB-MAPP测评已打开'), findsOneWidget);
  });

  testWidgets('VB-MAPP assessment page exposes all three modules',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: VbmappAssessmentPage(
            args: const VbmappAssessmentLaunchArgs(
              studentId: 51,
              studentName: '王小语',
              studentAge: '3岁',
              birthDate: '2023-01-01',
              assessmentDate: '2026-05-19',
            ),
            client: _FakeVbmappAssessmentClient(),
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('VB-MAPP语言行为评估 测评工作台'), findsOneWidget);
    expect(find.text('1 / 212'), findsWidgets);
    expect(find.text('0/170'), findsOneWidget);
    expect(find.text('0/24'), findsOneWidget);
    expect(find.text('0/18'), findsOneWidget);
    expect(find.text('当前得分'), findsOneWidget);
    expect(find.text('0.0 / 170'), findsOneWidget);
    expect(find.text('0 / 96'), findsOneWidget);
    expect(find.text('0 / 90'), findsOneWidget);
    expect(
      tester.getTopLeft(find.text('E')).dy,
      lessThan(tester.getTopLeft(find.textContaining('发出2个话语')).dy),
    );
    expect(tester.widget<Switch>(find.byType(Switch)).value, isFalse);

    await tester.tap(find.text('1分：2 个'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('1.0 / 170'), findsOneWidget);
    expect(find.text('1/170'), findsOneWidget);

    await tester.tap(find.text('障碍评估').first);
    await tester.pumpAndSettle();
    expect(find.text('171 / 212'), findsWidgets);

    await tester.tap(find.text('转衔评估').first);
    await tester.pumpAndSettle();
    expect(find.text('195 / 212'), findsWidgets);
  });

  testWidgets('VB-MAPP MAND 1M records requests and suggests score',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeVbmappAssessmentClient client = _FakeVbmappAssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: VbmappAssessmentPage(
            args: const VbmappAssessmentLaunchArgs(
              studentId: 51,
              studentName: '王小语',
              studentAge: '3岁',
              birthDate: '2023-01-01',
              assessmentDate: '2026-05-19',
            ),
            client: client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('提要求1M现场记录'), findsOneWidget);
    expect(find.text('有效 0/2'), findsOneWidget);
    expect(find.text('建议 0分'), findsOneWidget);
    expect(find.text('动机情境'), findsNothing);
    expect(find.text('形式'), findsNothing);
    expect(find.text('饼干'), findsOneWidget);
    expect(find.text('书'), findsOneWidget);
    expect(find.text('打开'), findsOneWidget);
    expect(find.text('彩虹弹簧'), findsNothing);

    final Finder requestField = find.byType(TextField).first;
    await tester.enterText(requestField, '饼干');
    await tester.tap(find.text('记录本次要求'));
    await tester.pumpAndSettle();

    expect(find.text('有效 1/2'), findsOneWidget);
    expect(find.text('建议 0.5分'), findsOneWidget);
    expect(find.text('0.5 / 170'), findsOneWidget);
    expect(find.text('饼干 -> 饼干'), findsNothing);
    expect(
      tester
          .getSize(find.byKey(const ValueKey<String>('vbmapp-mand-record-0')))
          .height,
      tester
          .getSize(find.byKey(const ValueKey<String>('vbmapp-mand-record-1')))
          .height,
    );

    await tester.enterText(requestField, '书');
    await tester.tap(find.text('记录本次要求'));
    await tester.pumpAndSettle();

    expect(find.text('有效 2/2'), findsOneWidget);
    expect(find.text('建议 1分'), findsOneWidget);
    expect(find.text('1.0 / 170'), findsOneWidget);
    expect(find.text('书 -> 书'), findsNothing);
    expect(find.byKey(const ValueKey<String>('vbmapp-mand-record-0')),
        findsOneWidget);
    expect(find.byKey(const ValueKey<String>('vbmapp-mand-record-1')),
        findsOneWidget);
    expect(
      tester
          .getTopLeft(
              find.byKey(const ValueKey<String>('vbmapp-mand-record-0')))
          .dy,
      tester
          .getTopLeft(
              find.byKey(const ValueKey<String>('vbmapp-mand-record-1')))
          .dy,
    );

    await tester.enterText(requestField, '苹果');
    await tester.tap(find.text('记录本次要求'));
    await tester.pumpAndSettle();

    expect(find.text('补充记录'), findsNothing);
    expect(find.byKey(const ValueKey<String>('vbmapp-mand-record-2')),
        findsOneWidget);
    expect(find.text('苹果'), findsOneWidget);
    expect(find.text('呈现物品 · 物品 · 无肢体辅助'), findsWidgets);
    expect(
      tester
          .getSize(find.byKey(const ValueKey<String>('vbmapp-mand-record-2')))
          .width,
      greaterThan(tester
              .getSize(
                  find.byKey(const ValueKey<String>('vbmapp-mand-record-0')))
              .width *
          1.8),
    );

    await tester.tap(find.text('是'));
    await tester.pumpAndSettle();
    await tester.enterText(requestField, '秋千');
    await tester.tap(find.text('记录本次要求'));
    await tester.pumpAndSettle();

    expect(find.byKey(const ValueKey<String>('vbmapp-mand-record-3')),
        findsOneWidget);
    expect(find.text('不计'), findsOneWidget);
    expect(find.text('呈现物品 · 物品 · 肢体辅助'), findsOneWidget);
    expect(
      tester
          .getTopLeft(
              find.byKey(const ValueKey<String>('vbmapp-mand-record-2')))
          .dy,
      tester
          .getTopLeft(
              find.byKey(const ValueKey<String>('vbmapp-mand-record-3')))
          .dy,
    );
    expect(find.text('删除'), findsNothing);

    final double recordHeightBeforeDelete = tester
        .getSize(find.byKey(const ValueKey<String>('vbmapp-mand-record-2')))
        .height;
    await tester
        .tap(find.byKey(const ValueKey<String>('vbmapp-mand-record-2')));
    await tester.pumpAndSettle();

    expect(find.text('删除'), findsOneWidget);
    expect(
      tester
          .getSize(find.byKey(const ValueKey<String>('vbmapp-mand-record-2')))
          .height,
      recordHeightBeforeDelete,
    );

    await tester
        .tap(find.byKey(const ValueKey<String>('vbmapp-mand-delete-record')));
    await tester.pumpAndSettle();

    expect(find.text('苹果'), findsNothing);
    expect(find.byKey(const ValueKey<String>('vbmapp-mand-record-3')),
        findsNothing);
    expect(
      find.descendant(
        of: find.byKey(const ValueKey<String>('vbmapp-mand-record-2')),
        matching: find.text('秋千'),
      ),
      findsOneWidget,
    );
    expect(find.text('有效 2/2'), findsOneWidget);
    expect(find.text('建议 1分'), findsOneWidget);
    expect(find.text('删除'), findsNothing);
    expect(client.saveDraftItemCalls, 5);
  });

  testWidgets('VB-MAPP MAND 2M records unprompted requests and suggests score',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeVbmappAssessmentClient client = _FakeVbmappAssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: VbmappAssessmentPage(
            args: const VbmappAssessmentLaunchArgs(
              studentId: 51,
              studentName: '王小语',
              studentAge: '3岁',
              birthDate: '2023-01-01',
              assessmentDate: '2026-05-19',
            ),
            client: client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('下一题'));
    await tester.pumpAndSettle();

    expect(find.text('提要求2M现场记录'), findsOneWidget);
    expect(find.text('有效 0/4'), findsOneWidget);
    expect(find.text('建议 0分'), findsOneWidget);
    expect(find.text('辅助'), findsOneWidget);
    expect(find.text('提问下'), findsOneWidget);
    expect(find.text('自发地'), findsOneWidget);
    expect(find.text('其他辅助'), findsNothing);
    expect(find.text('音乐'), findsOneWidget);
    expect(find.text('彩虹弹簧'), findsOneWidget);
    expect(find.text('打开'), findsNothing);
    expect(
      tester.getTopLeft(find.text('辅助')).dx,
      lessThan(tester.getTopLeft(find.text('环境')).dx),
    );
    expect(
      tester.getTopLeft(find.text('环境')).dx,
      lessThan(tester.getTopLeft(find.text('对象')).dx),
    );
    expect(find.byKey(const ValueKey<String>('vbmapp-mand-record-3')),
        findsOneWidget);
    expect(
      tester
          .getTopLeft(
              find.byKey(const ValueKey<String>('vbmapp-mand-record-0')))
          .dy,
      tester
          .getTopLeft(
              find.byKey(const ValueKey<String>('vbmapp-mand-record-1')))
          .dy,
    );
    expect(
      tester
          .getTopLeft(
              find.byKey(const ValueKey<String>('vbmapp-mand-record-2')))
          .dy,
      tester
          .getTopLeft(
              find.byKey(const ValueKey<String>('vbmapp-mand-record-3')))
          .dy,
    );

    final Finder requestField = find.byType(TextField).first;
    for (final String request in <String>['音乐', '球', '彩虹弹簧']) {
      await tester.enterText(requestField, request);
      await tester.tap(find.text('记录本次要求'));
      await tester.pumpAndSettle();
    }

    expect(find.text('有效 3/4'), findsOneWidget);
    expect(find.text('建议 0.5分'), findsOneWidget);
    expect(find.text('0.5 / 170'), findsOneWidget);

    await tester.tap(find.text('自发地'));
    await tester.pumpAndSettle();
    await tester.enterText(requestField, '泡泡');
    await tester.tap(find.text('记录本次要求'));
    await tester.pumpAndSettle();

    expect(find.text('有效 4/4'), findsOneWidget);
    expect(find.text('建议 1分'), findsOneWidget);
    expect(find.text('1.0 / 170'), findsOneWidget);
    expect(find.textContaining('提问下'), findsWidgets);
    expect(find.textContaining('自发地'), findsWidgets);
    expect(client.saveDraftItemCalls, 4);
  });

  testWidgets('ERXin workbench shows structured loading shell while loading',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: _FakeErxinAssessmentClient(
              templateSummaryDelay: const Duration(milliseconds: 300),
            ),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pump();

    expect(
      find.byKey(const ValueKey<String>('erxin-loading-shell')),
      findsOneWidget,
    );
    expect(find.text('儿心量表-II 测评工作台'), findsOneWidget);
    expect(find.text('能区进度'), findsOneWidget);
    expect(find.text('规则判断'), findsOneWidget);
    expect(find.text('当前题目说明：'), findsOneWidget);
    expect(find.text('操作方法'), findsOneWidget);
    expect(find.text('通过标准'), findsOneWidget);
    expect(find.text('题目备注'), findsOneWidget);
    expect(find.text('测评记录'), findsOneWidget);
    expect(find.text('测查推进'), findsOneWidget);
    expect(find.byType(CircularProgressIndicator), findsNothing);

    await tester.pump(const Duration(milliseconds: 350));
    await tester.pumpAndSettle();

    expect(
      find.byKey(const ValueKey<String>('erxin-loading-shell')),
      findsNothing,
    );
    expect(find.textContaining('当前题目说明：'), findsOneWidget);
  });

  testWidgets(
      'ERXin workbench continues backward when basal is not established',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: _FakeErxinAssessmentClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('全部展开'), findsNothing);
    expect(find.text('主测月龄 48月龄'), findsOneWidget);
    expect(find.text('主测月龄48月龄'), findsOneWidget);
    expect(find.text('主测月龄'), findsOneWidget);
    expect(find.textContaining('当前可见：'), findsNothing);
    expect(find.text('36月龄'), findsOneWidget);
    expect(find.text('42月龄'), findsOneWidget);
    expect(find.text('48月龄'), findsOneWidget);
    expect(
      tester.getTopLeft(find.text('48月龄')).dy,
      lessThan(tester.getTopLeft(find.text('42月龄')).dy),
    );
    expect(
      tester.getTopLeft(find.text('42月龄')).dy,
      lessThan(tester.getTopLeft(find.text('36月龄')).dy),
    );

    await _tapErxinScore(tester, '48月题', true);
    expect(find.textContaining('当前题目说明：142 42月题'), findsOneWidget);
    await _tapErxinScore(tester, '42月题', false);
    expect(find.textContaining('当前题目说明：136 36月题'), findsOneWidget);
    await _tapErxinScore(tester, '36月题', true);
    await tester.pumpAndSettle();

    expect(find.text('36月题'), findsOneWidget);
    expect(find.text('42月题'), findsOneWidget);
    expect(find.text('48月题'), findsOneWidget);
    expect(find.text('继续往前测查'), findsOneWidget);
    expect(find.textContaining('继续追加33月'), findsOneWidget);
    expect(find.text('往前42月龄'), findsOneWidget);
    expect(find.text('往前36月龄'), findsOneWidget);

    await tester.tap(find.text('往前36月龄'));
    await tester.pumpAndSettle();
    expect(find.text('大运动 · 36月龄记录'), findsOneWidget);

    await tester.tap(find.widgetWithText(FilledButton, '继续往前测查'));
    await tester.pumpAndSettle();

    expect(find.text('大运动 · 当前测查'), findsOneWidget);
    expect(find.text('大运动 · 36月龄记录'), findsNothing);
    expect(find.text('33月龄'), findsOneWidget);
  });

  testWidgets('ERXin backward progress reveals all newly added record months',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: _FakeErxinAssessmentClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await _tapErxinScore(tester, '48月题', true);
    await _tapErxinScore(tester, '42月题', false);
    await _tapErxinScore(tester, '36月题', false);
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(FilledButton, '继续往前测查'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 320));

    expect(find.text('33月龄'), findsOneWidget);
    expect(find.text('30月龄'), findsOneWidget);
    expect(find.text('往前33月龄'), findsOneWidget);
    expect(find.text('往前30月龄'), findsOneWidget);
    expect(
      find.ancestor(
        of: find.text('往前33月龄'),
        matching: find.byWidgetPredicate(
          (Widget widget) =>
              widget is Material && widget.color == const Color(0xFFFFF3BF),
        ),
      ),
      findsOneWidget,
    );
    expect(
      find.ancestor(
        of: find.text('往前30月龄'),
        matching: find.byWidgetPredicate(
          (Widget widget) =>
              widget is Material && widget.color == const Color(0xFFFFF3BF),
        ),
      ),
      findsOneWidget,
    );
  });

  testWidgets('ERXin item remarks stay scoped by selected item',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeErxinAssessmentClient client = _FakeErxinAssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await _enterErxinRemark(tester, '48题备注');
    expect(find.text('48题备注'), findsOneWidget);

    await _tapErxinItemRow(tester, '42月题');
    expect(find.text('48题备注'), findsNothing);

    await _enterErxinRemark(tester, '42题备注');
    await _tapErxinItemRow(tester, '48月题');

    expect(find.text('48题备注'), findsOneWidget);
    expect(find.text('42题备注'), findsNothing);

    await tester.tap(find.widgetWithText(InkWell, '保存草稿'));
    await tester.pumpAndSettle();

    final List<dynamic> itemRemarkList =
        client.saveDraftPayloads.last['itemRemarkList'] as List<dynamic>;
    expect(
      itemRemarkList,
      contains(
        allOf(
          containsPair('itemNo', 148),
          containsPair('remark', '48题备注'),
        ),
      ),
    );
  });

  testWidgets('ERXin remark editor opens above keyboard area',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
      tester.view.resetViewInsets();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          resizeToAvoidBottomInset: false,
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: _FakeErxinAssessmentClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('添加本题备注').last);
    await tester.pumpAndSettle();
    expect(
      tester.getTopLeft(find.byType(TextField).last).dy,
      lessThan(384),
    );

    tester.view.viewInsets = const FakeViewPadding(bottom: 320);
    await tester.pumpAndSettle();
    final Rect editorRect = tester.getRect(find.byType(TextField).last);
    expect(editorRect.bottom, lessThan(448));
  });

  testWidgets('ERXin record rows open history and confirm edits',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: _FakeErxinAssessmentClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await _tapErxinScore(tester, '36月题', true);
    await _tapErxinScore(tester, '42月题', true);
    await _tapErxinScore(tester, '48月题', true);
    await tester.pumpAndSettle();

    expect(find.text('48月题'), findsOneWidget);
    expect(find.text('前测基线'), findsOneWidget);
    expect(find.text('已建立'), findsOneWidget);

    expect(find.text('往前42月龄'), findsOneWidget);
    expect(find.text('往前42月'), findsNothing);

    await tester.tap(find.text('往前42月龄'));
    await tester.pumpAndSettle();

    expect(find.text('大运动 · 42月龄记录'), findsOneWidget);
    expect(find.text('42月题'), findsOneWidget);

    await _tapErxinScore(tester, '42月题', false);
    expect(find.text('确认修改历史记录'), findsOneWidget);

    await tester.tap(find.widgetWithText(OutlinedButton, '取消'));
    await tester.pumpAndSettle();

    expect(find.text('已建立'), findsOneWidget);
    await _tapErxinScore(tester, '42月题', false);
    await tester.tap(find.widgetWithText(FilledButton, '确认修改'));
    await tester.pumpAndSettle();

    expect(find.text('确认修改历史记录'), findsNothing);
    expect(find.text('继续往前测查'), findsOneWidget);
    expect(find.text('进入往后测查'), findsNothing);

    await tester.tap(find.widgetWithText(FilledButton, '继续往前测查'));
    await tester.pumpAndSettle();

    expect(find.text('大运动 · 当前测查'), findsOneWidget);
    expect(find.textContaining('当前题目说明：133 33月题'), findsOneWidget);
    expect(tester.getTopLeft(find.text('33月题')).dy, lessThan(560));
  });

  testWidgets('ERXin progress action leaves selected history month',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: _FakeErxinAssessmentClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await _tapErxinScore(tester, '36月题', true);
    await _tapErxinScore(tester, '42月题', true);
    await _tapErxinScore(tester, '48月题', true);
    await tester.pumpAndSettle();

    await tester.tap(find.text('往前36月龄'));
    await tester.pumpAndSettle();
    expect(find.text('大运动 · 36月龄记录'), findsOneWidget);

    await tester.tap(find.widgetWithText(FilledButton, '进入往后测查'));
    await tester.pumpAndSettle();

    expect(find.text('大运动 · 当前测查'), findsOneWidget);
    expect(find.text('大运动 · 36月龄记录'), findsNothing);
    expect(find.text('54月龄'), findsOneWidget);
    expect(find.text('60月龄'), findsOneWidget);
  });

  testWidgets('ERXin expanded progress action locates next pending item',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: _FakeErxinAssessmentClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('往前测查已展开'), findsOneWidget);
    await tester.tap(find.text('往前36月龄'));
    await tester.pumpAndSettle();
    expect(find.text('大运动 · 36月龄记录'), findsOneWidget);

    await tester.tap(find.widgetWithText(FilledButton, '往前测查已展开'));
    await tester.pumpAndSettle();

    expect(find.text('大运动 · 当前测查'), findsOneWidget);
    expect(find.textContaining('当前题目说明：148 48月题'), findsOneWidget);

    await _tapErxinScore(tester, '48月题', true);
    await _tapErxinScore(tester, '42月题', true);
    await _tapErxinScore(tester, '36月题', true);
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, '进入往后测查'));
    await tester.pumpAndSettle();

    expect(find.text('往后测查已展开'), findsOneWidget);
    await tester.tap(find.text('往前36月龄'));
    await tester.pumpAndSettle();
    expect(find.text('大运动 · 36月龄记录'), findsOneWidget);

    await tester.tap(find.widgetWithText(FilledButton, '往后测查已展开'));
    await tester.pumpAndSettle();

    expect(find.text('大运动 · 当前测查'), findsOneWidget);
    expect(find.textContaining('当前题目说明：154 54月题'), findsOneWidget);
  });

  testWidgets('ERXin workbench prompts and restores latest draft',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    const AssessmentDraftSummary draft = AssessmentDraftSummary(
      id: 77,
      studentName: '陈旭',
      assessmentCode: 'ERXIN2',
      assessmentName: '儿心量表-II',
      scaleVersion: 'WS-T-580-2017',
      examinerName: '陈老师',
      status: 'draft',
      answeredItemCount: 2,
      rawScoreCount: 0,
      completionPercent: .28,
      progressItemCount: 0,
      progressQuestionDisplayPreference: '',
      createdTime: '2026-05-08T09:00:00',
      updatedTime: '2026-05-08T10:00:00',
    );
    const ErxinDraftDetail detail = ErxinDraftDetail(
      id: 77,
      studentId: 31,
      studentName: '陈旭',
      birthDate: '2022-05-11T00:00:00+08:00',
      assessmentDate: '2026-05-08T00:00:00+08:00',
      examinerName: '陈老师',
      answeredItemCount: 2,
      completionPercent: .28,
      updatedTime: '2026-05-08T10:00:00',
      progress: ErxinDraftProgress(
        itemCount: 7,
        answeredItemCount: 2,
        missingItemCount: 5,
        completionPercent: .28,
        complete: false,
        canScore: false,
        missingItemNos: <int>[136],
      ),
      input: ErxinDraftInput(
        itemPasses: <int, bool>{148: true, 142: false},
        itemRemarks: <int, String>{142: '42题草稿备注'},
      ),
    );
    final _FakeErxinAssessmentClient client = _FakeErxinAssessmentClient(
      detectedDraft: draft,
      draftDetail: detail,
      draftDetailDelay: const Duration(milliseconds: 1),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(client.fetchDraftsPageCalls, 1);
    expect(client.fetchDraftDetailCalls, 1);
    expect(find.text('发现未完成草稿'), findsOneWidget);
    expect(find.text('继续测评'), findsOneWidget);
    expect(
      find.textContaining(
        '施测者：陈老师',
        findRichText: true,
        skipOffstage: false,
      ),
      findsOneWidget,
    );

    await tester.tap(find.text('继续测评'));
    await tester.pumpAndSettle();

    expect(find.text('发现未完成草稿'), findsNothing);
    expect(find.textContaining('当前题目说明：136 36月题'), findsOneWidget);
    expect(
      find.textContaining('施测者：陈老师', findRichText: true),
      findsOneWidget,
    );
    expect(
      find.textContaining('出生日期：2022-05-11', findRichText: true),
      findsOneWidget,
    );
    expect(
      find.textContaining('测查日期', findRichText: true),
      findsNothing,
    );
    expect(
      find.textContaining('日期：2026-05-08', findRichText: true),
      findsNothing,
    );
    expect(find.textContaining('T00:00:00', findRichText: true), findsNothing);
    expect(find.text('往前42月龄'), findsOneWidget);

    await tester.tap(find.text('往前42月龄'));
    await tester.pumpAndSettle();

    expect(find.text('大运动 · 42月龄记录'), findsOneWidget);
    expect(find.text('42题草稿备注'), findsOneWidget);

    await tester.tap(find.widgetWithText(InkWell, '保存草稿'));
    await tester.pumpAndSettle();

    expect(client.saveDraftPayloads.last['birthDate'], '2022-05-11');
    expect(client.saveDraftPayloads.last['assessmentDate'], '2026-05-08');
  });

  testWidgets('ERXin workbench can restart instead of applying detected draft',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    const AssessmentDraftSummary draft = AssessmentDraftSummary(
      id: 77,
      studentName: '陈旭',
      assessmentCode: 'ERXIN2',
      assessmentName: '儿心量表-II',
      scaleVersion: 'WS-T-580-2017',
      examinerName: '陈老师',
      status: 'draft',
      answeredItemCount: 2,
      rawScoreCount: 0,
      completionPercent: .28,
      progressItemCount: 0,
      progressQuestionDisplayPreference: '',
      createdTime: '2026-05-08T09:00:00',
      updatedTime: '2026-05-08T10:00:00',
    );
    const ErxinDraftDetail detail = ErxinDraftDetail(
      id: 77,
      studentId: 31,
      studentName: '陈旭',
      birthDate: '2022-05-11',
      assessmentDate: '2026-05-08',
      examinerName: '陈老师',
      answeredItemCount: 2,
      completionPercent: .28,
      updatedTime: '2026-05-08T10:00:00',
      progress: ErxinDraftProgress(
        itemCount: 7,
        answeredItemCount: 2,
        missingItemCount: 5,
        completionPercent: .28,
        complete: false,
        canScore: false,
        missingItemNos: <int>[136],
      ),
      input: ErxinDraftInput(
        itemPasses: <int, bool>{148: true, 142: false},
        itemRemarks: <int, String>{142: '42题草稿备注'},
      ),
    );
    final _FakeErxinAssessmentClient client = _FakeErxinAssessmentClient(
      detectedDraft: draft,
      draftDetail: detail,
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: client,
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('发现未完成草稿'), findsOneWidget);

    await tester.tap(find.text('重新测评'));
    await tester.pumpAndSettle();

    expect(find.text('发现未完成草稿'), findsNothing);
    expect(find.textContaining('当前题目说明：148 48月题'), findsOneWidget);
    expect(find.text('42题草稿备注'), findsNothing);
    expect(client.saveDraftCalls, 1);
    expect(client.saveDraftPayloads.single.containsKey('id'), isFalse);
    expect(client.saveDraftPayloads.single['itemPassList'], isEmpty);

    await tester.pumpWidget(const SizedBox.shrink());
    await tester.pumpAndSettle();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: client,
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('发现未完成草稿'), findsOneWidget);
    expect(find.textContaining('0 题', findRichText: true), findsOneWidget);

    await tester.tap(find.text('继续测评'));
    await tester.pumpAndSettle();

    expect(find.text('发现未完成草稿'), findsNothing);
    expect(find.textContaining('当前题目说明：148 48月题'), findsOneWidget);
    expect(find.text('42题草稿备注'), findsNothing);
  });

  testWidgets('ERXin draft list opens the ERXin workbench',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    const AssessmentDraftPage draftPage = AssessmentDraftPage(
      total: 1,
      current: 1,
      size: 5,
      items: <AssessmentDraftSummary>[
        AssessmentDraftSummary(
          id: 77,
          studentName: '陈旭',
          assessmentCode: 'ERXIN2',
          assessmentName: '儿心量表-II',
          scaleVersion: 'WS-T-580-2017',
          examinerName: '陈老师',
          status: 'draft',
          answeredItemCount: 2,
          rawScoreCount: 0,
          completionPercent: .28,
          progressItemCount: 0,
          progressQuestionDisplayPreference: '',
          createdTime: '2026-05-08T09:00:00',
          updatedTime: '2026-05-08T10:00:00',
        ),
      ],
    );
    const ErxinDraftDetail detail = ErxinDraftDetail(
      id: 77,
      studentId: 31,
      studentName: '陈旭',
      birthDate: '2022-05-11',
      assessmentDate: '2026-05-08',
      examinerName: '陈老师',
      answeredItemCount: 2,
      completionPercent: .28,
      updatedTime: '2026-05-08T10:00:00',
      progress: ErxinDraftProgress(
        itemCount: 7,
        answeredItemCount: 2,
        missingItemCount: 5,
        completionPercent: .28,
        complete: false,
        canScore: false,
        missingItemNos: <int>[136],
      ),
      input: ErxinDraftInput(
        itemPasses: <int, bool>{148: true, 142: false},
        itemRemarks: <int, String>{},
      ),
    );
    final _FakeErxinAssessmentClient erxinClient = _FakeErxinAssessmentClient(
      detectedDraft: draftPage.items.first,
      draftDetail: detail,
    );
    ErxinAssessmentLaunchArgs? openedArgs;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AssessmentScaleCategoryScreen(
            scaleClient: _FakeAssessmentScaleClient(
              draftPage: AssessmentDraftPage.empty,
            ),
            erxinClient: erxinClient,
            onBack: () {},
          ),
        ),
        onGenerateRoute: (RouteSettings settings) {
          if (settings.name == '/erxin-assessment') {
            final ErxinAssessmentLaunchArgs args =
                settings.arguments! as ErxinAssessmentLaunchArgs;
            openedArgs = args;
            return MaterialPageRoute<void>(
              settings: settings,
              builder: (BuildContext context) => Scaffold(
                body: ErxinAssessmentPage(
                  args: args,
                  client: erxinClient,
                  onBack: () => Navigator.of(context).maybePop(),
                ),
              ),
            );
          }
          return null;
        },
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('继续草稿').first);
    await tester.pumpAndSettle();
    await tester.tap(find.text('陈旭 · 儿心量表-II'));
    await tester.pumpAndSettle();

    expect(openedArgs?.draftId, 77);
    expect(find.text('儿心量表-II 测评工作台'), findsOneWidget);
    expect(find.textContaining('当前题目说明：136 36月题'), findsOneWidget);
    expect(erxinClient.fetchDraftsPageCalls, greaterThanOrEqualTo(1));
    expect(erxinClient.fetchDraftDetailCalls, 1);
  });

  testWidgets('ERXin save draft coalesces rapid taps',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeErxinAssessmentClient client = _FakeErxinAssessmentClient(
      saveDraftDelay: const Duration(milliseconds: 120),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: client,
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final Finder saveButton = find.widgetWithText(InkWell, '保存草稿');
    await tester.tap(saveButton);
    await tester.tap(saveButton);
    await tester.tap(saveButton);
    await tester.pump(const Duration(milliseconds: 20));

    expect(client.saveDraftCalls, 1);

    await tester.pumpAndSettle();

    expect(client.saveDraftCalls, 1);
    expect(find.text('草稿已保存'), findsWidgets);
  });

  testWidgets('ERXin save draft recreates missing server draft',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeErxinAssessmentClient client = _FakeErxinAssessmentClient(
      saveDraftDelay: const Duration(milliseconds: 120),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
            ),
            client: client,
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final Finder saveButton = find.widgetWithText(InkWell, '保存草稿');
    await tester.tap(saveButton);
    await tester.pumpAndSettle();

    client.failNextDraftUpdateAsNotFound = true;
    await tester.tap(saveButton);
    await tester.tap(saveButton);
    await tester.pumpAndSettle();

    expect(client.saveDraftCalls, 3);
    expect(client.saveDraftPayloads[1]['id'], 21);
    expect(client.saveDraftPayloads[2].containsKey('id'), isFalse);
    expect(find.textContaining('assessment draft not found'), findsNothing);
    expect(find.text('草稿已保存'), findsWidgets);
  });

  testWidgets('ERXin workbench can continue future months and submit',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeErxinAssessmentClient client = _FakeErxinAssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3岁11个月',
              birthDate: '2022-05-11',
              assessmentDate: '2026-05-08',
              examinerName: '陈老师',
            ),
            client: client,
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await _tapErxinScore(tester, '36月题', true);
    await _tapErxinScore(tester, '42月题', true);
    await _tapErxinScore(tester, '48月题', true);
    await tester.pumpAndSettle();

    expect(find.text('继续往前测查'), findsNothing);
    expect(find.text('进入往后测查'), findsOneWidget);

    await tester.tap(find.widgetWithText(FilledButton, '进入往后测查'));
    await tester.pumpAndSettle();

    expect(find.text('54月龄'), findsOneWidget);
    expect(find.text('60月龄'), findsOneWidget);

    await _tapErxinScore(tester, '54月题', false);
    await _tapErxinScore(tester, '60月题', false);
    await tester.pumpAndSettle();

    expect(find.text('本能区测查完成'), findsOneWidget);
    expect(find.text('前测基线'), findsOneWidget);
    expect(find.text('后测封顶'), findsOneWidget);
    expect(
      find.ancestor(
        of: find.text('前测基线'),
        matching: find.byType(ListView),
      ),
      findsNothing,
    );
    expect(
      find.ancestor(
        of: find.text('后测封顶'),
        matching: find.byType(ListView),
      ),
      findsNothing,
    );

    await tester.tap(find.text('前测基线'));
    await tester.pump(const Duration(milliseconds: 120));
    expect(
      find.ancestor(
        of: find.text('往前42月龄'),
        matching: find.byWidgetPredicate(
          (Widget widget) =>
              widget is Material && widget.color == const Color(0xFFFFF3BF),
        ),
      ),
      findsOneWidget,
    );
    expect(
      find.ancestor(
        of: find.text('往前36月龄'),
        matching: find.byWidgetPredicate(
          (Widget widget) =>
              widget is Material && widget.color == const Color(0xFFFFF3BF),
        ),
      ),
      findsOneWidget,
    );

    await tester.pump(const Duration(milliseconds: 900));
    await tester.tap(find.text('后测封顶'));
    await tester.pump(const Duration(milliseconds: 120));
    expect(
      find.ancestor(
        of: find.text('往后54月龄'),
        matching: find.byWidgetPredicate(
          (Widget widget) =>
              widget is Material && widget.color == const Color(0xFFFFF3BF),
        ),
      ),
      findsOneWidget,
    );
    expect(
      find.ancestor(
        of: find.text('往后60月龄'),
        matching: find.byWidgetPredicate(
          (Widget widget) =>
              widget is Material && widget.color == const Color(0xFFFFF3BF),
        ),
      ),
      findsOneWidget,
    );

    await tester.pump(const Duration(milliseconds: 900));
    await tester.tap(find.text('提交记录'));
    await tester.pumpAndSettle();

    expect(client.submitDraftCalls, 1);
  });

  testWidgets('ERXin lower boundary stop allows entering future search',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeErxinAssessmentClient client = _FakeErxinAssessmentClient(
      groups: _erxinBoundaryGroups(),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '3个月',
              birthDate: '2026-02-08',
              assessmentDate: '2026-05-08',
              examinerName: '陈老师',
            ),
            client: client,
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('主测月龄 3月龄'), findsOneWidget);
    await _tapErxinScore(tester, '3月题', true);
    await _tapErxinScore(tester, '2月题', true);
    await _tapErxinScore(tester, '1月题', false);
    await tester.pumpAndSettle();

    expect(find.text('继续往前测查'), findsNothing);
    expect(find.text('已到最低月龄'), findsOneWidget);
    expect(find.text('进入往后测查'), findsOneWidget);

    await tester.tap(find.widgetWithText(FilledButton, '进入往后测查'));
    await tester.pumpAndSettle();

    expect(find.text('4月龄'), findsOneWidget);
    expect(find.text('5月龄'), findsOneWidget);

    await _tapErxinScore(tester, '4月题', false);
    await _tapErxinScore(tester, '5月题', false);
    await tester.pumpAndSettle();

    expect(find.text('本能区测查完成'), findsOneWidget);
    await tester.tap(find.text('提交记录'));
    await tester.pumpAndSettle();

    expect(client.submitDraftCalls, 1);
  });

  testWidgets('ERXin upper boundary stop completes future search',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakeErxinAssessmentClient client = _FakeErxinAssessmentClient(
      groups: _erxinBoundaryGroups(),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErxinAssessmentPage(
            args: const ErxinAssessmentLaunchArgs(
              studentId: 31,
              studentName: '陈旭',
              studentAge: '6岁',
              birthDate: '2020-05-08',
              assessmentDate: '2026-05-08',
              examinerName: '陈老师',
            ),
            client: client,
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('主测月龄 72月龄'), findsOneWidget);
    await _tapErxinScore(tester, '72月题', true);
    await _tapErxinScore(tester, '66月题', true);
    await _tapErxinScore(tester, '60月题', true);
    await tester.pumpAndSettle();

    expect(find.text('进入往后测查'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, '进入往后测查'));
    await tester.pumpAndSettle();

    expect(find.text('78月龄'), findsOneWidget);
    expect(find.text('84月龄'), findsOneWidget);

    await _tapErxinScore(tester, '78月题', true);
    await _tapErxinScore(tester, '84月题', true);
    await tester.pumpAndSettle();

    expect(find.text('已到最高月龄'), findsOneWidget);
    expect(find.text('本能区测查完成'), findsOneWidget);

    await tester.tap(find.text('提交记录'));
    await tester.pumpAndSettle();

    expect(client.submitDraftCalls, 1);
  });

  testWidgets('selected PEP3 scale opens dedicated workbench',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        scaleClient: _FakeAssessmentScaleClient(),
        pep3Client: _FakePep3AssessmentClient(),
        erxinClient: _FakeErxinAssessmentClient(),
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('新建测评'));
    await tester.pumpAndSettle();
    final Finder pep3Card = find.ancestor(
      of: find.text('PEP-3语言理解评核量表'),
      matching: find.byWidgetPredicate(
        (Widget widget) => widget.runtimeType.toString() == '_ScaleCard',
      ),
    );
    expect(pep3Card, findsOneWidget);
    final Rect cardRect = tester.getRect(pep3Card);
    await tester.tapAt(Offset(cardRect.center.dx, cardRect.bottom - 31));
    await tester.pumpAndSettle();
    await tester.tap(find.text('张一鸣'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择并进入测评'));
    await tester.pumpAndSettle();

    expect(find.text('PEP-3 测评工作台'), findsOneWidget);
    expect(find.text('记录册页面'), findsOneWidget);
    expect(find.textContaining('旋开瓶盖'), findsWidgets);
  });

  testWidgets('PEP3 workbench header shows day-level age before full month',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final DateTime today = DateTime.now();
    final DateTime birthDate = DateTime(today.year - 4, today.month, today.day)
        .subtract(const Duration(days: 3));

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: Pep3AssessmentLaunchArgs(
              studentId: 24,
              studentName: '周小天',
              studentAge: '4岁',
              birthDate: _formatDateDashForTest(birthDate),
              assessmentDate: _formatDateDashForTest(today),
            ),
            client: _FakePep3AssessmentClient(),
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pump();

    expect(find.textContaining('4岁3天', findRichText: true), findsOneWidget);
  });

  testWidgets('PEP3 workbench prompts when latest draft exists',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: _FakePep3AssessmentClient(hasDraft: true),
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('发现未完成草稿'), findsOneWidget);
    expect(find.text('当前儿童存在一份未提交的 PEP-3 测评草稿。'), findsOneWidget);
    expect(
      find.textContaining('已完成：1 / 2', findRichText: true),
      findsOneWidget,
    );
    expect(find.text('重新测评'), findsOneWidget);
    expect(find.text('继续测评'), findsOneWidget);
  });

  testWidgets('PEP3 workbench shows skeleton while initial data loads',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: _FakePep3AssessmentClient(
              summaryFetchDelay: Duration(milliseconds: 300),
            ),
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('PEP-3 测评工作台'), findsOneWidget);
    expect(
      find.textContaining('施测者：-', findRichText: true),
      findsOneWidget,
    );
    expect(
      find.byWidgetPredicate(
        (Widget widget) =>
            widget.runtimeType.toString() == '_Pep3SidebarSkeleton',
      ),
      findsOneWidget,
    );
    expect(find.byType(CircularProgressIndicator), findsNothing);

    await tester.pump(const Duration(milliseconds: 350));
    await tester.pumpAndSettle();
    expect(find.textContaining('旋开瓶盖'), findsWidgets);
  });

  testWidgets('PEP3 header shows complete age without ellipsis',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '3岁11个月',
              birthDate: '',
              assessmentDate: '2026-05-05',
            ),
            client: _FakePep3AssessmentClient(),
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('年龄：3岁11个月', findRichText: true), findsOneWidget);
    expect(
      find.textContaining('年龄：3岁11...', findRichText: true),
      findsNothing,
    );
  });

  testWidgets('PEP3 draft dialog converts UTC updated time to local time',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    const String utcUpdatedTime = '2026-05-05T12:08:00Z';

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: _FakePep3AssessmentClient(
              hasDraft: true,
              draftUpdatedTime: utcUpdatedTime,
            ),
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final String expectedUpdatedTime =
        _fakeMinuteText(DateTime.parse(utcUpdatedTime).toLocal());
    expect(
      find.textContaining(
        '更新时间：$expectedUpdatedTime',
        findRichText: true,
      ),
      findsOneWidget,
    );
    if (!expectedUpdatedTime.endsWith('12:08')) {
      expect(find.textContaining('12:08', findRichText: true), findsNothing);
    }
  });

  testWidgets('PEP3 continue draft does not switch to full page loading',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client = _FakePep3AssessmentClient(
      hasDraft: true,
      draftDetailDelay: const Duration(milliseconds: 120),
      inviteDelay: const Duration(milliseconds: 900),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('继续测评'));
    await tester.pump();

    expect(find.text('题目填充中，请稍后...'), findsOneWidget);
    expect(
      find.byWidgetPredicate(
        (Widget widget) => widget.runtimeType.toString() == '_Pep3LoadingShell',
      ),
      findsNothing,
    );

    await tester.pump(const Duration(milliseconds: 160));
    expect(
      find.byWidgetPredicate(
        (Widget widget) => widget.runtimeType.toString() == '_Pep3LoadingShell',
      ),
      findsNothing,
    );

    await tester.pumpAndSettle();
    expect(find.text('发现未完成草稿'), findsNothing);
    expect(pep3Client.inviteCalls, 1);
  });

  testWidgets('PEP3 restart creates draft and caregiver QR immediately',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client =
        _FakePep3AssessmentClient(hasDraft: true);

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('重新测评'));
    await tester.pump();

    expect(find.text('发现未完成草稿'), findsOneWidget);
    expect(pep3Client.saveDraftCalls, 0);

    await tester.pump(const Duration(milliseconds: 300));
    await tester.pumpAndSettle();

    expect(pep3Client.saveDraftCalls, 1);
    expect(pep3Client.inviteCalls, 1);
    expect(find.text('暂无二维码'), findsNothing);
  });

  testWidgets('PEP3 caregiver action buttons only show custom message',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client = _FakePep3AssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(pep3Client.inviteCalls, 1);

    await tester.tap(find.text('发送短信给家长'));
    await tester.pump(const Duration(milliseconds: 120));
    expect(pep3Client.inviteCalls, 1);
    expect(find.text('短信发送功能暂未开放'), findsOneWidget);
    expect(find.byType(SnackBar), findsNothing);

    await tester.tap(find.text('推送微信消息'));
    await tester.pump(const Duration(milliseconds: 120));
    expect(pep3Client.inviteCalls, 1);
    expect(find.text('微信推送功能暂未开放'), findsOneWidget);
    expect(find.byType(SnackBar), findsNothing);
  });

  testWidgets('PEP3 save draft uses custom message',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client = _FakePep3AssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(pep3Client.saveDraftCalls, 1);
    expect(pep3Client.inviteCalls, 1);

    await tester.tap(find.text('保存草稿'));
    await tester.pump(const Duration(milliseconds: 120));
    expect(pep3Client.saveDraftCalls, 2);
    expect(pep3Client.inviteCalls, 1);
    expect(find.text('草稿已保存'), findsWidgets);
    expect(
      find.byWidgetPredicate(
        (Widget widget) => widget.runtimeType.toString() == 'PadTopMessage',
      ),
      findsOneWidget,
    );
    expect(find.byType(SnackBar), findsNothing);
  });

  testWidgets('PEP3 training record selection does not refresh caregiver QR',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client =
        _FakePep3AssessmentClient(includeRecordField: true);

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(pep3Client.inviteCalls, 1);

    await tester.tap(find.text('三角形'));
    await tester.pumpAndSettle();

    expect(pep3Client.saveDraftItemCalls, 1);
    expect(pep3Client.inviteCalls, 1);
  });

  testWidgets('PEP3 question switching does not refresh caregiver QR',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client = _FakePep3AssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(pep3Client.inviteCalls, 1);

    await tester.tap(find.text('下一题'));
    await tester.pumpAndSettle();

    expect(find.textContaining('叠积木'), findsWidgets);
    expect(pep3Client.inviteCalls, 1);
  });

  testWidgets('PEP3 record text stays isolated when switching focused question',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client =
        _FakePep3AssessmentClient(includeTextRecordField: true);

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('下一题'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextFormField), '第36题输入内容');
    await tester.pump();

    expect(
      tester
          .widget<EditableText>(find.byType(EditableText).last)
          .controller
          .text,
      '第36题输入内容',
    );

    await tester.tap(find.text('上一题'));
    await tester.pumpAndSettle();

    expect(
      tester
          .widget<EditableText>(find.byType(EditableText).last)
          .controller
          .text,
      isEmpty,
    );
    expect(find.text('第36题输入内容'), findsNothing);
  });

  testWidgets('PEP3 fast next loading skeleton does not overflow',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client = _FakePep3AssessmentClient(
      itemFetchDelay: const Duration(milliseconds: 300),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pump(const Duration(milliseconds: 350));
    expect(find.text('下一题'), findsOneWidget);

    await tester.tap(find.text('下一题'));
    await tester.pump();

    expect(find.text('题目加载中'), findsOneWidget);
    expect(tester.takeException(), isNull);

    await tester.pumpAndSettle();
  });

  testWidgets('PEP3 score dock stays fixed while instructions scroll',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: _FakePep3AssessmentClient(longInstructions: true),
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final Finder instructionScroll = find.byKey(
      const ValueKey<String>('pep3-question-instruction-scroll'),
    );
    final Finder scoreDock = find.byKey(
      const ValueKey<String>('pep3-question-score-dock'),
    );
    final double scoreDockTop = tester.getTopLeft(scoreDock).dy;

    await tester.drag(instructionScroll, const Offset(0, -420));
    await tester.pumpAndSettle();

    expect(scoreDock, findsOneWidget);
    expect(tester.getTopLeft(scoreDock).dy, closeTo(scoreDockTop, .1));
    expect(
      find.descendant(of: scoreDock, matching: find.text('评分')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: scoreDock, matching: find.text('2 分')),
      findsOneWidget,
    );
  });

  testWidgets('PEP3 submit validation uses custom message',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    final _FakePep3AssessmentClient pep3Client = _FakePep3AssessmentClient();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: pep3Client,
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('提交记录'));
    await tester.pump(const Duration(milliseconds: 120));

    expect(find.text('还有 2 道题未评分，请补全后再提交'), findsOneWidget);
    expect(
      find.byWidgetPredicate(
        (Widget widget) => widget.runtimeType.toString() == 'PadTopMessage',
      ),
      findsOneWidget,
    );
    expect(find.byType(SnackBar), findsNothing);
  });

  testWidgets('PEP3 workbench shows previous assessment score',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1366, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Pep3AssessmentPage(
            args: const Pep3AssessmentLaunchArgs(
              studentId: 3,
              studentName: '张一鸣',
              studentAge: '5岁2个月',
              birthDate: '2021-03-01',
              assessmentDate: '2026-05-05',
            ),
            client: _FakePep3AssessmentClient(hasPreviousRecord: true),
            homeClient: _FakeHomeClient(),
            onBack: () {},
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(
      find.text('测评日期：2026-05-05', findRichText: true),
      findsOneWidget,
    );
    expect(find.text('上次测评 2026-05-04'), findsOneWidget);
    expect(find.text('0 分 · 未通过'), findsOneWidget);
    expect(find.text('上次 05-04'), findsOneWidget);
  });
}

Future<void> _enterWithCustomKeyboard(
  WidgetTester tester,
  int fieldIndex,
  String value,
) async {
  await tester.tap(find.byType(TextField).at(fieldIndex));
  await tester.pumpAndSettle();

  for (final String character in value.split('')) {
    await tester.tap(find.byKey(ValueKey<String>('login-key-$character')));
    await tester.pump(const Duration(milliseconds: 20));
  }
}

Future<void> _tapIepAiGenerateAndConfirm(
  WidgetTester tester, {
  bool last = false,
}) async {
  final Finder generateButton = find.text('AI生成');
  final int matchCount = generateButton.evaluate().length;
  await tester
      .tap(last || matchCount > 1 ? generateButton.last : generateButton);
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 220));
  expect(find.text('确认生成'), findsOneWidget);
  await tester.tap(find.text('确认生成'));
  await tester.pump();
}

class _FakeAuthClient implements AuthClient {
  @override
  Uri buildQrLoginUri(String nonce) {
    return Uri.parse('https://example.com/qr?nonce=$nonce');
  }

  @override
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier, {
    String password = '',
  }) async {
    return <InstitutionLoginOption>[];
  }

  @override
  Future<LoginResult> login({
    required String username,
    required String password,
    InstitutionLoginOption? institution,
  }) async {
    return <LoginResult>[
      LoginResult(
        token: 'fake-token',
        loginType: 'org',
        tenantId: 'tenant-a',
        orgId: 1,
        raw: <String, dynamic>{
          'token': 'fake-token',
          'loginType': 'org',
          'tenantId': 'tenant-a',
          'orgId': 1,
        },
      ),
    ].first;
  }
}

class _FakeHomeClient implements HomeClient {
  int fetchCurrentSessionCalls = 0;
  int fetchSummaryCalls = 0;

  @override
  Future<HomeSession> fetchCurrentSession(String token) async {
    fetchCurrentSessionCalls += 1;
    return const HomeSession(
      nickName: '陈老师',
      orgName: '启明成长中心',
    );
  }

  @override
  Future<HomeSummary> fetchSummary(String token) async {
    fetchSummaryCalls += 1;
    return const HomeSummary(
      date: '2026-05-04',
      weekday: '星期一',
      assessmentStats: HomeAssessmentStats(
        enrolledStudents: 80,
        assessedStudents: 32,
        inProgressDrafts: 2,
        unassessedStudents: 48,
        completedRecords: 40,
        pendingIep: 12,
        draftIep: 4,
        generatedIep: 28,
        total: 80,
        coverageRate: 0.4,
      ),
      schedule: <HomeScheduleItem>[
        HomeScheduleItem(
          time: '09:00',
          title: '认知能力评估 · 小组课',
          place: '教室A101',
          state: '进行中',
        ),
        HomeScheduleItem(
          time: '10:30',
          title: '情绪与行为评估 · 个训',
          place: '咨询室2',
          state: '即将开始',
        ),
      ],
      weather: HomeWeather(
        city: '深圳',
        condition: 'sunny',
        displayName: '晴',
        temperature: 26,
      ),
    );
  }
}

class _FakeIepAssessmentRecordClient implements IepAssessmentRecordClient {
  _FakeIepAssessmentRecordClient({this.delay = Duration.zero});

  final Duration delay;
  int fetchRecordsPageCalls = 0;

  @override
  Future<IepAssessmentRecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 20,
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  }) async {
    fetchRecordsPageCalls += 1;
    if (delay > Duration.zero) {
      await Future<void>.delayed(delay);
    }
    return const IepAssessmentRecordPage(
      total: 2,
      current: 1,
      size: 2,
      items: <IepAssessmentRecordSummary>[
        IepAssessmentRecordSummary(
          id: 91,
          source: 'ERXIN',
          studentId: 18,
          studentName: '陈旭',
          assessmentCode: 'ERXIN2',
          assessmentName: '儿心量表-II',
          birthDate: '2022-05-11',
          assessmentDate: '2026-05-07',
          ageYears: 4,
          ageMonths: 0,
          ageDays: 0,
          examinerName: '陈瑞',
          iepPlanStatus: 'confirmed',
          updatedTime: '2026-05-07T10:30:00Z',
        ),
        IepAssessmentRecordSummary(
          id: 88,
          source: 'PEP3',
          studentId: 19,
          studentName: '林一诺',
          assessmentCode: 'PEP3',
          assessmentName: 'PEP-3',
          birthDate: '2021-08-12',
          assessmentDate: '2026-04-29',
          ageYears: 4,
          ageMonths: 8,
          ageDays: 17,
          examinerName: '陈瑞',
          iepPlanStatus: '',
          updatedTime: '2026-04-29T11:30:00Z',
        ),
      ],
    );
  }
}

class _FakePendingIepAssessmentRecordClient
    implements IepAssessmentRecordClient {
  @override
  Future<IepAssessmentRecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 20,
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  }) async {
    return const IepAssessmentRecordPage(
      total: 1,
      current: 1,
      size: 1,
      items: <IepAssessmentRecordSummary>[
        IepAssessmentRecordSummary(
          id: 88,
          source: 'PEP3',
          studentId: 19,
          studentName: '林一诺',
          studentGender: '女',
          assessmentCode: 'PEP3',
          assessmentName: 'PEP-3',
          birthDate: '2021-08-12',
          assessmentDate: '2026-04-29',
          ageYears: 4,
          ageMonths: 8,
          ageDays: 17,
          examinerName: '陈瑞',
          iepPlanStatus: '',
          updatedTime: '2026-04-29T11:30:00Z',
        ),
      ],
    );
  }
}

class _FakeShuangxiIepAssessmentRecordClient
    implements IepAssessmentRecordClient {
  @override
  Future<IepAssessmentRecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 20,
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  }) async {
    return const IepAssessmentRecordPage(
      total: 1,
      current: 1,
      size: 1,
      items: <IepAssessmentRecordSummary>[
        IepAssessmentRecordSummary(
          id: 301,
          source: 'SHUANGXI',
          studentId: 88,
          studentName: '双溪学生',
          studentGender: '男',
          assessmentCode: 'SHUANGXI_A',
          assessmentName: '双溪课程评量表A',
          birthDate: '2019-12-01',
          assessmentDate: '2026-05-18',
          ageYears: 6,
          ageMonths: 5,
          ageDays: 17,
          examinerName: '陈老师',
          iepPlanStatus: '',
          updatedTime: '2026-05-18T13:30:00Z',
        ),
      ],
    );
  }
}

class _FakeIepPlanClient implements IepPlanClient {
  DateTime _startDate = DateTime(2026, 5);
  int savePlanCalls = 0;
  int generatePlanCalls = 0;
  int generateMonthlyPlanCalls = 0;
  int generateWeeklyPlanCalls = 0;
  int saveMonthlyPlanCalls = 0;
  int saveWeeklyPlanCalls = 0;
  int createTaskCalls = 0;
  int fetchTaskCalls = 0;
  int watchTaskCalls = 0;
  IepPlan? lastSavedPlan;
  IepMonthlyPlan? lastSavedMonthlyPlan;
  IepWeeklyPlan? lastSavedWeeklyPlan;

  @override
  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  }) async {
    return null;
  }

  @override
  Future<IepPlanSaved> fetchIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    return IepPlanSaved(
      exists: true,
      status: 'confirmed',
      durationMonths: durationMonths,
      plan: IepPlan(
        title: durationMonths == 6 ? '康复教学半年计划' : '康复教学季度计划',
        student: const IepPlanStudent(
          name: '陈旭',
          gender: '-',
          birthDate: '2022-05-11',
        ),
        meta: const IepPlanMeta(
          planDate: '2026-05-07',
          participant: '陈瑞',
          implementer: '陈瑞',
          startDate: '',
          endDate: '',
        ),
        rows: const <IepPlanRow>[
          IepPlanRow(
            domain: '大肌肉',
            longGoal: '1. 提升动态平衡与协调能力，能在移动中稳定控制身体',
            shortGoal: '能单脚站立保持平衡5秒以上',
            courseForm: '个训',
            startEndDate: '2026-05-01 - 2026-05-31',
          ),
          IepPlanRow(
            domain: '大肌肉',
            longGoal: '1. 提升动态平衡与协调能力，能在移动中稳定控制身体',
            shortGoal: '能双脚连续向前跳5步以上',
            courseForm: '集体课',
            startEndDate: '2026-05-01 - 2026-05-31',
          ),
          IepPlanRow(
            domain: '大肌肉',
            longGoal: '1. 提升动态平衡与协调能力，能在移动中稳定控制身体',
            shortGoal: '能在平衡木上独立行走2米',
            courseForm: '个训',
            startEndDate: '2026-05-01 - 2026-05-31',
          ),
        ],
      ),
    );
  }

  @override
  Future<IepExecutionPlansSaved> fetchExecutionPlans(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    return const IepExecutionPlansSaved(
      exists: true,
      durationMonths: 3,
      monthlyPlans: <IepMonthlyPlanSaved>[
        IepMonthlyPlanSaved(
          targetMonthIndex: 1,
          plan: IepMonthlyPlan(
            title: '康复教学5月计划',
            student: IepPlanStudent(
              name: '陈旭',
              gender: '-',
              birthDate: '2022-05-11',
            ),
            restWeekdays: <int>[DateTime.sunday],
            meta: IepMonthlyPlanMeta(
              planDate: '2026-05-07',
              participant: '陈瑞',
              implementer: '陈瑞',
              startDate: '2026-05-01',
              endDate: '2026-05-31',
              monthLabel: '5月',
              sourceTitle: '康复教学季度计划',
            ),
            rows: <IepMonthlyPlanRow>[
              IepMonthlyPlanRow(
                domain: '大肌肉',
                longGoal: '提升动态平衡与协调能力',
                shortGoal: '能在平衡木上独立行走3米',
                candidateTrainingItems: <IepMonthlyTrainingItem>[],
                trainingItems: <IepMonthlyTrainingItem>[
                  IepMonthlyTrainingItem(
                    content: '平衡木行走训练',
                    startEndDate: '2026-05-01 - 2026-05-10',
                  ),
                ],
                courseForm: '个训',
              ),
            ],
          ),
        ),
      ],
      weeklyPlans: <IepWeeklyPlanSaved>[
        IepWeeklyPlanSaved(
          targetMonthIndex: 1,
          targetWeekIndex: 1,
          plan: IepWeeklyPlan(
            title: '康复教学周计划日记录卡5月第1周',
            student: IepPlanStudent(
              name: '陈旭',
              gender: '-',
              birthDate: '2022-05-11',
            ),
            teacherName: '陈瑞',
            courseName: '康复教学',
            trainingDate: '2026-05-01 至 2026-05-02',
            preparation: '平衡木、记录表',
            weekDates: <String>['2026-05-01', '2026-05-02'],
            restWeekdays: <int>[DateTime.sunday],
            rows: <IepWeeklyPlanRow>[
              IepWeeklyPlanRow(
                project: '平衡木行走',
                content: '在平衡木上独立行走并记录掉落次数',
                completion: <String>[],
              ),
            ],
          ),
        ),
      ],
    );
  }

  @override
  Future<IepPlanPeriodSyncResult> syncIepPlanPeriod(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int sourceDurationMonths,
    required DateTime startDate,
    String syncMode = 'dates_only',
  }) async {
    _startDate = DateTime(startDate.year, startDate.month, startDate.day);
    final DateTime endDate =
        DateTime(_startDate.year, _startDate.month + durationMonths, 0);
    final IepPlanSaved plan = await fetchIepPlan(
      token,
      record: record,
      durationMonths: durationMonths,
    );
    return IepPlanPeriodSyncResult(
      iepPlan: IepPlanSaved(
        exists: plan.exists,
        status: plan.status,
        durationMonths: plan.durationMonths,
        updatedTime: plan.updatedTime,
        plan: IepPlan(
          title: plan.plan!.title,
          student: plan.plan!.student,
          meta: IepPlanMeta(
            planDate: plan.plan!.meta.planDate,
            participant: plan.plan!.meta.participant,
            implementer: plan.plan!.meta.implementer,
            startDate: _formatDateDashForTest(_startDate),
            endDate: _formatDateDashForTest(endDate),
          ),
          rows: plan.plan!.rows,
        ),
      ),
      executionPlans: await fetchExecutionPlans(
        token,
        record: record,
        durationMonths: durationMonths,
      ),
    );
  }

  @override
  Stream<IepPlanGenerationEvent> generateIepPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async* {
    generatePlanCalls += 1;
    yield IepPlanGenerationEvent.status('正在读取评估和训练记录');
    yield IepPlanGenerationEvent.delta(
      '{"title":"康复教学季度计划","rows":[',
    );
    yield IepPlanGenerationEvent.delta(
      '{"domain":"大肌肉","longGoal":"提升动态平衡能力","shortGoal":"能独立跳跃3次","courseForm":"个训","startEndDate":"2026-05-01 - 2026-05-31"}',
    );
    yield IepPlanGenerationEvent.done(
      IepPlan(
        title: durationMonths == 6 ? '康复教学半年计划' : '康复教学季度计划',
        student: IepPlanStudent(
          name: record.studentName,
          gender: record.studentGender,
          birthDate: record.birthDate,
        ),
        meta: IepPlanMeta(
          planDate: record.assessmentDate,
          participant: record.examinerName,
          implementer: record.examinerName,
          startDate: '2026-05-01',
          endDate: durationMonths == 6 ? '2026-10-31' : '2026-07-31',
        ),
        rows: const <IepPlanRow>[
          IepPlanRow(
            domain: '大肌肉',
            longGoal: '提升动态平衡能力',
            shortGoal: '能独立跳跃3次',
            courseForm: '个训',
            startEndDate: '2026-05-01 - 2026-05-31',
          ),
        ],
      ),
    );
  }

  @override
  Future<IepPlanGenerationTask> createIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    createTaskCalls += 1;
    return IepPlanGenerationTask(
      taskId: 'fake-task-${record.id}',
      status: 'running',
      durationMonths: durationMonths,
      message: '正在读取评估和训练记录',
    );
  }

  @override
  Future<IepPlanGenerationTask> fetchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async {
    fetchTaskCalls += 1;
    return IepPlanGenerationTask(
      taskId: taskId,
      status: 'done',
      durationMonths: 3,
      savedPlan: await saveIepPlan(
        token,
        record: record,
        durationMonths: 3,
        status: 'draft',
        plan: lastSavedPlan ??
            IepPlan(
              title: '康复教学季度计划',
              student: IepPlanStudent(
                name: record.studentName,
                gender: record.studentGender,
                birthDate: record.birthDate,
              ),
              meta: IepPlanMeta(
                planDate: record.assessmentDate,
                participant: record.examinerName,
                implementer: record.examinerName,
                startDate: '2026-05-01',
                endDate: '2026-07-31',
              ),
              rows: const <IepPlanRow>[
                IepPlanRow(
                  domain: '大肌肉',
                  longGoal: '提升动态平衡能力',
                  shortGoal: '能独立跳跃3次',
                  courseForm: '个训',
                  startEndDate: '2026-05-01 - 2026-05-31',
                ),
              ],
            ),
      ),
    );
  }

  @override
  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) {
    watchTaskCalls += 1;
    return generateIepPlanStream(token, record: record, durationMonths: 3);
  }

  @override
  Future<IepPlanSaved> saveIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required String status,
    required IepPlan plan,
    bool resetExecutionPlans = false,
  }) async {
    savePlanCalls += 1;
    lastSavedPlan = plan;
    return IepPlanSaved(
      exists: true,
      status: status,
      durationMonths: durationMonths,
      plan: plan,
      updatedTime: '2026-05-10T09:30:00Z',
    );
  }

  @override
  Stream<IepExecutionPlanGenerationEvent<IepMonthlyPlan>>
      generateMonthlyPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    List<int> restWeekdays = const <int>[],
    required IepPlan sourcePlan,
  }) async* {
    generateMonthlyPlanCalls += 1;
    final int monthNumber = 4 + targetMonthIndex;
    final String monthLabel = '${monthNumber}月';
    final String monthValue = monthNumber.toString().padLeft(2, '0');
    yield IepExecutionPlanGenerationEvent<IepMonthlyPlan>.status(
        '正在准备月度计划生成上下文');
    yield IepExecutionPlanGenerationEvent<IepMonthlyPlan>.delta(
      '{"title":"康复教学${monthLabel}计划","rows":[{"shortGoal":"能在平衡木上独立行走3米"}]}',
    );
    yield IepExecutionPlanGenerationEvent<IepMonthlyPlan>.done(
      IepMonthlyPlan(
        title: '康复教学${monthLabel}计划',
        student: IepPlanStudent(
          name: record.studentName,
          gender: record.studentGender,
          birthDate: record.birthDate,
        ),
        restWeekdays: restWeekdays,
        meta: IepMonthlyPlanMeta(
          planDate: '2026-05-07',
          participant: '陈瑞',
          implementer: '陈瑞',
          startDate: '2026-$monthValue-01',
          endDate: '2026-$monthValue-30',
          monthLabel: monthLabel,
          sourceTitle: '康复教学季度计划',
        ),
        rows: const <IepMonthlyPlanRow>[
          IepMonthlyPlanRow(
            domain: '大肌肉',
            longGoal: '提升动态平衡与协调能力',
            shortGoal: '能在平衡木上独立行走3米',
            candidateTrainingItems: <IepMonthlyTrainingItem>[],
            trainingItems: <IepMonthlyTrainingItem>[
              IepMonthlyTrainingItem(
                content: '平衡木交替步态训练',
                startEndDate: '2026-05-01 - 2026-05-10',
              ),
            ],
            courseForm: '个训',
          ),
        ],
      ),
    );
  }

  @override
  Stream<IepExecutionPlanGenerationEvent<IepWeeklyPlan>>
      generateWeeklyPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required IepPlan sourcePlan,
    IepMonthlyPlan? monthlyPlan,
    List<int> restWeekdays = const <int>[],
  }) async* {
    generateWeeklyPlanCalls += 1;
    final int monthNumber = 4 + targetMonthIndex;
    final String monthLabel = '${monthNumber}月';
    yield IepExecutionPlanGenerationEvent<IepWeeklyPlan>.status('正在准备周计划生成上下文');
    yield IepExecutionPlanGenerationEvent<IepWeeklyPlan>.delta(
      '{"title":"康复教学周计划日记录卡${monthLabel}第${targetWeekIndex}周","rows":[{"project":"平衡木行走"}]}',
    );
    yield IepExecutionPlanGenerationEvent<IepWeeklyPlan>.done(
      IepWeeklyPlan(
        title: '康复教学周计划日记录卡${monthLabel}第${targetWeekIndex}周',
        student: IepPlanStudent(
          name: record.studentName,
          gender: record.studentGender,
          birthDate: record.birthDate,
        ),
        teacherName: '陈瑞',
        courseName: monthlyPlan?.title ?? '康复教学${monthLabel}计划',
        trainingDate: '2026-05-01 至 2026-05-02',
        preparation: '平衡木、记录表',
        weekDates: const <String>['2026-05-01', '2026-05-02'],
        restWeekdays: restWeekdays,
        rows: const <IepWeeklyPlanRow>[
          IepWeeklyPlanRow(
            project: '平衡木行走',
            content: '在平衡木上独立行走并记录掉落次数',
            completion: <String>[],
          ),
        ],
      ),
    );
  }

  @override
  Future<IepExecutionPlansSaved> saveMonthlyPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required IepMonthlyPlan plan,
    bool preserveWeeklyPlans = false,
  }) async {
    saveMonthlyPlanCalls += 1;
    lastSavedMonthlyPlan = plan;
    return IepExecutionPlansSaved(
      exists: true,
      durationMonths: durationMonths,
      monthlyPlans: <IepMonthlyPlanSaved>[
        IepMonthlyPlanSaved(targetMonthIndex: targetMonthIndex, plan: plan),
      ],
      weeklyPlans: preserveWeeklyPlans
          ? const <IepWeeklyPlanSaved>[
              IepWeeklyPlanSaved(
                targetMonthIndex: 1,
                targetWeekIndex: 1,
                plan: IepWeeklyPlan(
                  title: '康复教学周计划日记录卡5月第1周',
                  student: IepPlanStudent(
                    name: '陈旭',
                    gender: '-',
                    birthDate: '2022-05-11',
                  ),
                  teacherName: '陈瑞',
                  courseName: '康复教学',
                  trainingDate: '2026-05-01 至 2026-05-02',
                  preparation: '平衡木、记录表',
                  weekDates: <String>['2026-05-01', '2026-05-02'],
                  restWeekdays: <int>[DateTime.sunday],
                  rows: <IepWeeklyPlanRow>[
                    IepWeeklyPlanRow(
                      project: '平衡木行走',
                      content: '在平衡木上独立行走并记录掉落次数',
                      completion: <String>[],
                    ),
                  ],
                ),
              ),
            ]
          : const <IepWeeklyPlanSaved>[],
    );
  }

  @override
  Future<IepExecutionPlansSaved> saveWeeklyPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required IepWeeklyPlan plan,
  }) async {
    saveWeeklyPlanCalls += 1;
    lastSavedWeeklyPlan = plan;
    return IepExecutionPlansSaved(
      exists: true,
      durationMonths: durationMonths,
      monthlyPlans: lastSavedMonthlyPlan == null
          ? const <IepMonthlyPlanSaved>[]
          : <IepMonthlyPlanSaved>[
              IepMonthlyPlanSaved(
                targetMonthIndex: targetMonthIndex,
                plan: lastSavedMonthlyPlan!,
              ),
            ],
      weeklyPlans: <IepWeeklyPlanSaved>[
        IepWeeklyPlanSaved(
          targetMonthIndex: targetMonthIndex,
          targetWeekIndex: targetWeekIndex,
          plan: plan,
        ),
      ],
    );
  }

  @override
  Future<IepLessonSessionWeekState> fetchLessonSessionWeekState(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
  }) async {
    return const IepLessonSessionWeekState();
  }

  @override
  Future<IepLessonSessionWeekState> startLessonSession(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required String lessonDate,
  }) async {
    return IepLessonSessionWeekState(
      exists: true,
      currentSession: IepLessonSession(
        lessonDate: lessonDate,
        weekDateIndex: 1,
        status: 'in_progress',
        elapsedSeconds: 0,
      ),
      sessions: <IepLessonSession>[
        IepLessonSession(
          lessonDate: lessonDate,
          weekDateIndex: 1,
          status: 'in_progress',
          elapsedSeconds: 0,
        ),
      ],
    );
  }

  @override
  Future<IepLessonSessionWeekState> pauseLessonSession(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required String lessonDate,
  }) async {
    return IepLessonSessionWeekState(
      exists: true,
      sessions: <IepLessonSession>[
        IepLessonSession(
          lessonDate: lessonDate,
          weekDateIndex: 1,
          status: 'paused',
          elapsedSeconds: 60,
        ),
      ],
    );
  }

  @override
  Future<IepLessonSessionWeekState> completeLessonSession(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required String lessonDate,
  }) async {
    return IepLessonSessionWeekState(
      exists: true,
      sessions: <IepLessonSession>[
        IepLessonSession(
          lessonDate: lessonDate,
          weekDateIndex: 1,
          status: 'completed',
          elapsedSeconds: 120,
        ),
      ],
    );
  }

  @override
  Future<IepLessonSessionWeekState> heartbeatLessonSession(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required String lessonDate,
  }) async {
    return IepLessonSessionWeekState(
      exists: true,
      currentSession: IepLessonSession(
        lessonDate: lessonDate,
        weekDateIndex: 1,
        status: 'in_progress',
        elapsedSeconds: 30,
      ),
      sessions: <IepLessonSession>[
        IepLessonSession(
          lessonDate: lessonDate,
          weekDateIndex: 1,
          status: 'in_progress',
          elapsedSeconds: 30,
        ),
      ],
    );
  }

  @override
  Future<IepWordFile> downloadIepPlanWord(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required IepPlan plan,
  }) async {
    return _fakeIepWordFile('${record.studentName}-IEP.docx');
  }

  @override
  Future<IepWordFile> downloadMonthlyPlanWord(
    String token, {
    required IepAssessmentRecordSummary record,
    required IepMonthlyPlan plan,
  }) async {
    return _fakeIepWordFile('${record.studentName}-月计划.docx');
  }

  @override
  Future<IepWordFile> downloadWeeklyPlanWord(
    String token, {
    required IepAssessmentRecordSummary record,
    required IepWeeklyPlan plan,
  }) async {
    return _fakeIepWordFile('${record.studentName}-周计划.docx');
  }

  @override
  Future<Uint8List> downloadIepPlanPdf(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required IepPlan plan,
  }) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadMonthlyPlanPdf(
    String token, {
    required IepAssessmentRecordSummary record,
    required IepMonthlyPlan plan,
  }) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadWeeklyPlanPdf(
    String token, {
    required IepAssessmentRecordSummary record,
    required IepWeeklyPlan plan,
  }) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<IepExecutionPlanGenerationTask> createExecutionPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required String planType,
    required int targetMonthIndex,
    int targetWeekIndex = 0,
    required IepPlan sourcePlan,
    IepMonthlyPlan? monthlyPlan,
    List<int> restWeekdays = const <int>[],
  }) async {
    return IepExecutionPlanGenerationTask(
      taskId:
          'fake-execution-$planType-${record.id}-$targetMonthIndex-$targetWeekIndex',
      status: 'running',
      durationMonths: durationMonths,
      planType: planType,
      targetMonthIndex: targetMonthIndex,
      targetWeekIndex: targetWeekIndex,
      restWeekdays: restWeekdays,
      message: planType == 'weekly' ? '正在准备周计划生成上下文' : '正在准备月度计划生成上下文',
    );
  }

  @override
  Future<IepExecutionPlanGenerationTask> fetchExecutionPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async {
    return IepExecutionPlanGenerationTask(
      taskId: taskId,
      status: 'failed',
      durationMonths: 3,
      planType: 'monthly',
      error: '测试任务已结束',
    );
  }

  @override
  Future<IepExecutionPlanGenerationTask?>
      fetchActiveExecutionPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required String planType,
    required int targetMonthIndex,
    int targetWeekIndex = 0,
  }) async {
    return null;
  }

  @override
  Stream<IepExecutionPlanGenerationEvent<dynamic>>
      watchExecutionPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async* {
    if (taskId.contains('weekly')) {
      yield* generateWeeklyPlanStream(
        token,
        record: record,
        durationMonths: 3,
        targetMonthIndex: 1,
        targetWeekIndex: 1,
        sourcePlan: lastSavedPlan ??
            IepPlan(
              title: '康复教学季度计划',
              student: IepPlanStudent(
                name: record.studentName,
                gender: record.studentGender,
                birthDate: record.birthDate,
              ),
              meta: IepPlanMeta(
                planDate: record.assessmentDate,
                participant: record.examinerName,
                implementer: record.examinerName,
                startDate: '2026-05-01',
                endDate: '2026-07-31',
              ),
              rows: const <IepPlanRow>[],
            ),
      );
      return;
    }
    yield* generateMonthlyPlanStream(
      token,
      record: record,
      durationMonths: 3,
      targetMonthIndex: 1,
      sourcePlan: lastSavedPlan ??
          IepPlan(
            title: '康复教学季度计划',
            student: IepPlanStudent(
              name: record.studentName,
              gender: record.studentGender,
              birthDate: record.birthDate,
            ),
            meta: IepPlanMeta(
              planDate: record.assessmentDate,
              participant: record.examinerName,
              implementer: record.examinerName,
              startDate: '2026-05-01',
              endDate: '2026-07-31',
            ),
            rows: const <IepPlanRow>[],
          ),
    );
  }
}

class _SlowFirstIepPlanClient extends _FakeIepPlanClient {
  _SlowFirstIepPlanClient({required this.planDelay});

  final Duration planDelay;
  bool _delayedFirstFetch = false;

  @override
  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  }) async {
    return null;
  }

  @override
  Future<IepPlanSaved> fetchIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    if (!_delayedFirstFetch) {
      _delayedFirstFetch = true;
      await Future<void>.delayed(planDelay);
    }
    return super.fetchIepPlan(
      token,
      record: record,
      durationMonths: durationMonths,
    );
  }
}

class _LongGoalIepPlanClient extends _FakeIepPlanClient {
  static const String longGoal =
      '1. 能在移动中稳定控制身体，完成平衡木行走、单脚站立、连续跳跃等动作时保持身体中线稳定，减少明显摇晃和停顿。\n'
      '2. 能在不同场地和不同材料上迁移动态平衡能力，例如垫上、地面标线、低矮平衡木和户外边缘，完成动作后能主动回到起点等待下一轮。\n'
      '3. 能根据老师口令调整速度和方向，在安全范围内完成前进、后退、转身和跨越障碍组合动作。';
  static const String shortGoal =
      '能连续向前跳跃并保持稳定，双脚同时起跳同时落地，连续完成8次以上，中途不扶墙、不坐下休息；在老师更换地垫颜色、距离和口令节奏后，仍能完成动作并主动等待下一次指令；'
      '能够在红色、蓝色、黄色三种地垫之间按照口头提示切换路线，落地后保持身体稳定两秒以上，不出现明显跌倒、跪坐或离开训练区域的情况，完成后能主动回到起点。';
  static const String shortGoal2 = '能双脚交替上下楼梯（一步一阶，扶扶手）';
  static const String shortGoal3 = '能连续向前翻滚2次（在保护下完成）';

  @override
  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  }) async {
    return null;
  }

  @override
  Future<IepPlanSaved> fetchIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    return IepPlanSaved(
      exists: true,
      status: 'confirmed',
      durationMonths: durationMonths,
      plan: IepPlan(
        title: durationMonths == 6 ? '康复教学半年计划' : '康复教学季度计划',
        student: const IepPlanStudent(
          name: '陈旭',
          gender: '-',
          birthDate: '2022-05-11',
        ),
        meta: const IepPlanMeta(
          planDate: '2026-05-07',
          participant: '陈瑞',
          implementer: '陈瑞',
          startDate: '2026-05-01',
          endDate: '2026-07-31',
        ),
        rows: const <IepPlanRow>[
          IepPlanRow(
            domain: '大肌肉',
            longGoal: longGoal,
            shortGoal: shortGoal,
            courseForm: '个训',
            startEndDate: '2026-05-01 - 2026-05-31',
          ),
          IepPlanRow(
            domain: '大肌肉',
            longGoal: longGoal,
            shortGoal: shortGoal2,
            courseForm: '个训',
            startEndDate: '2026-06-01 - 2026-06-30',
          ),
          IepPlanRow(
            domain: '大肌肉',
            longGoal: longGoal,
            shortGoal: shortGoal3,
            courseForm: '个训',
            startEndDate: '2026-07-01 - 2026-07-31',
          ),
        ],
      ),
    );
  }
}

class _ConfirmRegenerateIepPlanClient extends _FakeIepPlanClient {
  @override
  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  }) async {
    return null;
  }

  @override
  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async* {
    watchTaskCalls += 1;
    yield IepPlanGenerationEvent.error('测试中断生成');
  }
}

class _DisconnectThenResumeIepPlanClient
    extends _EmptyThenGeneratedIepPlanClient {
  int createTaskCalls = 0;
  int fetchTaskCalls = 0;
  int watchTaskCalls = 0;

  @override
  Future<IepPlanGenerationTask> createIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    createTaskCalls += 1;
    return IepPlanGenerationTask(
      taskId: 'resume-task-${record.id}',
      status: 'running',
      durationMonths: durationMonths,
      message: '正在读取评估和训练记录',
    );
  }

  @override
  Future<IepPlanGenerationTask> fetchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async {
    fetchTaskCalls += 1;
    return IepPlanGenerationTask(
      taskId: taskId,
      status: 'running',
      durationMonths: 3,
      message: 'AI正在生成IEP计划',
      streamText: '{"rows":[{"shortGoal":"能恢复',
    );
  }

  @override
  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async* {
    watchTaskCalls += 1;
    if (watchTaskCalls == 1) {
      yield IepPlanGenerationEvent.status('AI正在生成IEP计划');
      yield IepPlanGenerationEvent.delta('{"rows":[{"shortGoal":"能恢复');
      throw const IepPlanApiException('接口连接失败，请稍后重试');
    }
    yield IepPlanGenerationEvent.delta('订阅并完成"}]}');
    final IepPlan plan = IepPlan(
      title: '康复教学季度计划',
      student: IepPlanStudent(
        name: record.studentName,
        gender: record.studentGender,
        birthDate: record.birthDate,
      ),
      meta: IepPlanMeta(
        planDate: record.assessmentDate,
        participant: record.examinerName,
        implementer: record.examinerName,
        startDate: '2026-05-01',
        endDate: '2026-07-31',
      ),
      rows: const <IepPlanRow>[
        IepPlanRow(
          domain: '大肌肉',
          longGoal: '提升动态平衡能力',
          shortGoal: '能恢复订阅并完成',
          courseForm: '个训',
          startEndDate: '2026-05-01 - 2026-05-31',
        ),
      ],
    );
    yield IepPlanGenerationEvent.done(
      plan,
      savedPlan: IepPlanSaved(
        exists: true,
        status: 'draft',
        durationMonths: 3,
        plan: plan,
        updatedTime: '2026-05-10T09:30:00Z',
      ),
    );
  }
}

class _DisconnectThenResumeRegenerateIepPlanClient extends _FakeIepPlanClient {
  @override
  int createTaskCalls = 0;

  @override
  int fetchTaskCalls = 0;

  @override
  int watchTaskCalls = 0;

  @override
  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  }) async {
    return null;
  }

  @override
  Future<IepPlanGenerationTask> createIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    createTaskCalls += 1;
    return IepPlanGenerationTask(
      taskId: 'resume-regenerate-task-${record.id}',
      status: 'running',
      durationMonths: durationMonths,
      message: '正在读取评估和训练记录',
    );
  }

  @override
  Future<IepPlanGenerationTask> fetchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async {
    fetchTaskCalls += 1;
    return IepPlanGenerationTask(
      taskId: taskId,
      status: 'running',
      durationMonths: 3,
      message: 'AI正在重新生成IEP计划',
      streamText: '{"rows":[{"shortGoal":"重新生成后恢复',
    );
  }

  @override
  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async* {
    watchTaskCalls += 1;
    if (watchTaskCalls == 1) {
      yield IepPlanGenerationEvent.status('AI正在重新生成IEP计划');
      yield IepPlanGenerationEvent.delta('{"rows":[{"shortGoal":"重新生成后恢复');
      throw const IepPlanApiException('接口连接失败，请稍后重试');
    }
    yield IepPlanGenerationEvent.delta('订阅并完成"}]}');
    final IepPlan plan = IepPlan(
      title: '康复教学季度计划',
      student: IepPlanStudent(
        name: record.studentName,
        gender: record.studentGender,
        birthDate: record.birthDate,
      ),
      meta: IepPlanMeta(
        planDate: record.assessmentDate,
        participant: record.examinerName,
        implementer: record.examinerName,
        startDate: '2026-05-01',
        endDate: '2026-07-31',
      ),
      rows: const <IepPlanRow>[
        IepPlanRow(
          domain: '大肌肉',
          longGoal: '提升动态平衡能力',
          shortGoal: '重新生成后恢复订阅并完成',
          courseForm: '个训',
          startEndDate: '2026-05-01 - 2026-05-31',
        ),
      ],
    );
    yield IepPlanGenerationEvent.done(
      plan,
      savedPlan: IepPlanSaved(
        exists: true,
        status: 'draft',
        durationMonths: 3,
        plan: plan,
        updatedTime: '2026-05-10T09:30:00Z',
      ),
    );
  }
}

class _ColdStartResumeIepPlanClient extends _EmptyThenGeneratedIepPlanClient {
  int createTaskCalls = 0;
  int fetchTaskCalls = 0;
  int fetchActiveTaskCalls = 0;
  int watchTaskCalls = 0;

  @override
  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  }) async {
    fetchActiveTaskCalls += 1;
    return const IepPlanGenerationTask(
      taskId: 'cold-start-task',
      status: 'running',
      durationMonths: 3,
      message: 'AI正在生成IEP计划',
      streamText: '{"rows":[{"shortGoal":"冷启动恢复',
    );
  }

  @override
  Future<IepPlanGenerationTask> createIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    createTaskCalls += 1;
    return await super.createIepPlanGenerationTask(
      token,
      record: record,
      durationMonths: durationMonths,
    );
  }

  @override
  Future<IepPlanGenerationTask> fetchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async {
    fetchTaskCalls += 1;
    return IepPlanGenerationTask(
      taskId: taskId,
      status: 'running',
      durationMonths: 3,
      message: 'AI正在生成IEP计划',
      streamText: '{"rows":[{"shortGoal":"冷启动恢复',
    );
  }

  @override
  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async* {
    watchTaskCalls += 1;
    yield IepPlanGenerationEvent.delta('成功"}]}');
    final IepPlan plan = IepPlan(
      title: '康复教学季度计划',
      student: IepPlanStudent(
        name: record.studentName,
        gender: record.studentGender,
        birthDate: record.birthDate,
      ),
      meta: IepPlanMeta(
        planDate: record.assessmentDate,
        participant: record.examinerName,
        implementer: record.examinerName,
        startDate: '2026-05-01',
        endDate: '2026-07-31',
      ),
      rows: const <IepPlanRow>[
        IepPlanRow(
          domain: '大肌肉',
          longGoal: '提升动态平衡能力',
          shortGoal: '冷启动恢复成功',
          courseForm: '个训',
          startEndDate: '2026-05-01 - 2026-05-31',
        ),
      ],
    );
    yield IepPlanGenerationEvent.done(
      plan,
      savedPlan: IepPlanSaved(
        exists: true,
        status: 'draft',
        durationMonths: 3,
        plan: plan,
        updatedTime: '2026-05-10T09:30:00Z',
      ),
    );
  }
}

class _RepeatedLongGoalStreamIepPlanClient
    extends _EmptyThenGeneratedIepPlanClient {
  @override
  Stream<IepPlanGenerationEvent> generateIepPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async* {
    yield IepPlanGenerationEvent.status('正在读取评估和训练记录');
    await Future<void>.delayed(const Duration(milliseconds: 60));
    yield IepPlanGenerationEvent.delta(
      '{"rows":[{"domain":"大肌肉","longGoal":"提升动态平衡能力","shortGoal":"能单脚站立保持平衡5秒以上","courseForm":"集体课","startEndDate":"2026-05-01 - 2026-05-31"},',
    );
    await Future<void>.delayed(const Duration(milliseconds: 260));
    yield IepPlanGenerationEvent.delta(
      '{"domain":"大肌肉","longGoal":"提升动态',
    );
    await Future<void>.delayed(const Duration(milliseconds: 120));
    yield IepPlanGenerationEvent.delta(
      '平衡能力","shortGoal":"能双脚连续跳跃5次","courseForm":"集体课","startEndDate":"2026-06-01 - 2026-06-30"}]}',
    );
    yield IepPlanGenerationEvent.done(
      IepPlan(
        title: '康复教学季度计划',
        student: IepPlanStudent(
          name: record.studentName,
          gender: record.studentGender,
          birthDate: record.birthDate,
        ),
        meta: IepPlanMeta(
          planDate: record.assessmentDate,
          participant: record.examinerName,
          implementer: record.examinerName,
          startDate: '2026-05-01',
          endDate: '2026-07-31',
        ),
        rows: const <IepPlanRow>[
          IepPlanRow(
            domain: '大肌肉',
            longGoal: '提升动态平衡能力',
            shortGoal: '能单脚站立保持平衡5秒以上',
            courseForm: '集体课',
            startEndDate: '2026-05-01 - 2026-05-31',
          ),
          IepPlanRow(
            domain: '大肌肉',
            longGoal: '提升动态平衡能力',
            shortGoal: '能双脚连续跳跃5次',
            courseForm: '集体课',
            startEndDate: '2026-06-01 - 2026-06-30',
          ),
        ],
      ),
    );
  }
}

class _RepeatedDomainStreamIepPlanClient
    extends _EmptyThenGeneratedIepPlanClient {
  @override
  Stream<IepPlanGenerationEvent> generateIepPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async* {
    yield IepPlanGenerationEvent.status('正在读取评估和训练记录');
    yield IepPlanGenerationEvent.delta(
      '{"rows":[{"domain":"大肌肉","longGoal":"提升动态平衡能力","shortGoal":"能单脚站立保持平衡5秒以上","courseForm":"集体课","startEndDate":"2026-05-01 - 2026-05-31"},',
    );
    await Future<void>.delayed(const Duration(milliseconds: 120));
    yield IepPlanGenerationEvent.delta(
      '{"domain":"大肌',
    );
    await Future<void>.delayed(const Duration(milliseconds: 120));
    yield IepPlanGenerationEvent.delta(
      '肉","longGoal":"提升动态平衡能力","shortGoal":"能双脚连续跳跃5次","courseForm":"集体课","startEndDate":"2026-06-01 - 2026-06-30"}]}',
    );
    yield IepPlanGenerationEvent.done(
      IepPlan(
        title: '康复教学季度计划',
        student: IepPlanStudent(
          name: record.studentName,
          gender: record.studentGender,
          birthDate: record.birthDate,
        ),
        meta: IepPlanMeta(
          planDate: record.assessmentDate,
          participant: record.examinerName,
          implementer: record.examinerName,
          startDate: '2026-05-01',
          endDate: '2026-07-31',
        ),
        rows: const <IepPlanRow>[
          IepPlanRow(
            domain: '大肌肉',
            longGoal: '提升动态平衡能力',
            shortGoal: '能单脚站立保持平衡5秒以上',
            courseForm: '集体课',
            startEndDate: '2026-05-01 - 2026-05-31',
          ),
          IepPlanRow(
            domain: '大肌肉',
            longGoal: '提升动态平衡能力',
            shortGoal: '能双脚连续跳跃5次',
            courseForm: '集体课',
            startEndDate: '2026-06-01 - 2026-06-30',
          ),
        ],
      ),
    );
  }
}

class _EmptyThenGeneratedIepPlanClient implements IepPlanClient {
  _EmptyThenGeneratedIepPlanClient({this.failSave = false});

  final bool failSave;
  int savePlanCalls = 0;
  IepPlan? lastSavedPlan;

  @override
  Future<IepPlanSaved> fetchIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    return IepPlanSaved.empty(durationMonths);
  }

  @override
  Future<IepExecutionPlansSaved> fetchExecutionPlans(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    return IepExecutionPlansSaved.empty(durationMonths);
  }

  @override
  Future<IepPlanPeriodSyncResult> syncIepPlanPeriod(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int sourceDurationMonths,
    required DateTime startDate,
    String syncMode = 'dates_only',
  }) async {
    return IepPlanPeriodSyncResult.empty(durationMonths);
  }

  @override
  Stream<IepPlanGenerationEvent> generateIepPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async* {
    yield IepPlanGenerationEvent.status('正在读取评估和训练记录');
    await Future<void>.delayed(const Duration(milliseconds: 100));
    yield IepPlanGenerationEvent.delta(
      '{"rows":[{"shortGoal":"能独',
    );
    await Future<void>.delayed(const Duration(milliseconds: 40));
    yield IepPlanGenerationEvent.delta(
      '立跳跃3次","domain":"大肌肉","longGoal":"提升动态平衡能力","courseForm":"个训","startEndDate":"2026-04-01 - 2026-04-30"}',
    );
    await Future<void>.delayed(const Duration(milliseconds: 20));
    final IepPlan plan = IepPlan(
      title: '康复教学季度计划',
      student: IepPlanStudent(
        name: record.studentName,
        gender: record.studentGender,
        birthDate: record.birthDate,
      ),
      meta: IepPlanMeta(
        planDate: record.assessmentDate,
        participant: record.examinerName,
        implementer: record.examinerName,
        startDate: '2026-04-01',
        endDate: '2026-06-30',
      ),
      rows: const <IepPlanRow>[
        IepPlanRow(
          domain: '大肌肉',
          longGoal: '提升动态平衡能力',
          shortGoal: '能独立跳跃3次',
          courseForm: '个训',
          startEndDate: '2026-04-01 - 2026-04-30',
        ),
      ],
    );
    yield IepPlanGenerationEvent.done(
      plan,
      savedPlan: IepPlanSaved(
        exists: true,
        status: 'draft',
        durationMonths: durationMonths,
        plan: plan,
        updatedTime: '2026-05-10T09:30:00Z',
      ),
    );
  }

  @override
  Future<IepPlanGenerationTask> createIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    return IepPlanGenerationTask(
      taskId: 'empty-task-${record.id}',
      status: 'running',
      durationMonths: durationMonths,
    );
  }

  @override
  Future<IepPlanGenerationTask> fetchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async {
    return IepPlanGenerationTask(
      taskId: taskId,
      status: 'failed',
      durationMonths: 3,
      error: '测试任务已结束',
    );
  }

  @override
  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  }) async {
    return null;
  }

  @override
  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) {
    return generateIepPlanStream(token, record: record, durationMonths: 3);
  }

  @override
  Future<IepPlanSaved> saveIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required String status,
    required IepPlan plan,
    bool resetExecutionPlans = false,
  }) async {
    savePlanCalls += 1;
    lastSavedPlan = plan;
    if (failSave) {
      throw const IepPlanApiException('保存接口返回内容异常');
    }
    return IepPlanSaved(
      exists: true,
      status: status,
      durationMonths: durationMonths,
      plan: plan,
      updatedTime: '2026-05-10T09:30:00Z',
    );
  }

  @override
  Stream<IepExecutionPlanGenerationEvent<IepMonthlyPlan>>
      generateMonthlyPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    List<int> restWeekdays = const <int>[],
    required IepPlan sourcePlan,
  }) async* {
    yield IepExecutionPlanGenerationEvent<IepMonthlyPlan>.error('未实现');
  }

  @override
  Stream<IepExecutionPlanGenerationEvent<IepWeeklyPlan>>
      generateWeeklyPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required IepPlan sourcePlan,
    IepMonthlyPlan? monthlyPlan,
    List<int> restWeekdays = const <int>[],
  }) async* {
    yield IepExecutionPlanGenerationEvent<IepWeeklyPlan>.error('未实现');
  }

  @override
  Future<IepExecutionPlanGenerationTask> createExecutionPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required String planType,
    required int targetMonthIndex,
    int targetWeekIndex = 0,
    required IepPlan sourcePlan,
    IepMonthlyPlan? monthlyPlan,
    List<int> restWeekdays = const <int>[],
  }) async {
    return IepExecutionPlanGenerationTask(
      taskId: 'empty-execution-$planType-${record.id}',
      status: 'running',
      durationMonths: durationMonths,
      planType: planType,
      targetMonthIndex: targetMonthIndex,
      targetWeekIndex: targetWeekIndex,
      restWeekdays: restWeekdays,
    );
  }

  @override
  Future<IepExecutionPlanGenerationTask> fetchExecutionPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async {
    return IepExecutionPlanGenerationTask(
      taskId: taskId,
      status: 'failed',
      durationMonths: 3,
      planType: 'monthly',
      error: '测试任务已结束',
    );
  }

  @override
  Future<IepExecutionPlanGenerationTask?>
      fetchActiveExecutionPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required String planType,
    required int targetMonthIndex,
    int targetWeekIndex = 0,
  }) async {
    return null;
  }

  @override
  Stream<IepExecutionPlanGenerationEvent<dynamic>>
      watchExecutionPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async* {
    if (taskId.contains('weekly')) {
      yield IepExecutionPlanGenerationEvent<dynamic>.error('未实现');
      return;
    }
    yield IepExecutionPlanGenerationEvent<dynamic>.error('未实现');
  }

  @override
  Future<IepExecutionPlansSaved> saveMonthlyPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required IepMonthlyPlan plan,
    bool preserveWeeklyPlans = false,
  }) async {
    return IepExecutionPlansSaved.empty(durationMonths);
  }

  @override
  Future<IepExecutionPlansSaved> saveWeeklyPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required IepWeeklyPlan plan,
  }) async {
    return IepExecutionPlansSaved.empty(durationMonths);
  }

  @override
  Future<IepLessonSessionWeekState> fetchLessonSessionWeekState(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
  }) async {
    return const IepLessonSessionWeekState();
  }

  @override
  Future<IepLessonSessionWeekState> startLessonSession(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required String lessonDate,
  }) async {
    return IepLessonSessionWeekState(
      exists: true,
      currentSession: IepLessonSession(
        lessonDate: lessonDate,
        weekDateIndex: 1,
        status: 'in_progress',
      ),
      sessions: <IepLessonSession>[
        IepLessonSession(
          lessonDate: lessonDate,
          weekDateIndex: 1,
          status: 'in_progress',
        ),
      ],
    );
  }

  @override
  Future<IepLessonSessionWeekState> pauseLessonSession(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required String lessonDate,
  }) async {
    return IepLessonSessionWeekState(
      exists: true,
      sessions: <IepLessonSession>[
        IepLessonSession(
          lessonDate: lessonDate,
          weekDateIndex: 1,
          status: 'paused',
        ),
      ],
    );
  }

  @override
  Future<IepLessonSessionWeekState> completeLessonSession(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required String lessonDate,
  }) async {
    return IepLessonSessionWeekState(
      exists: true,
      sessions: <IepLessonSession>[
        IepLessonSession(
          lessonDate: lessonDate,
          weekDateIndex: 1,
          status: 'completed',
        ),
      ],
    );
  }

  @override
  Future<IepLessonSessionWeekState> heartbeatLessonSession(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required String lessonDate,
  }) async {
    return IepLessonSessionWeekState(
      exists: true,
      currentSession: IepLessonSession(
        lessonDate: lessonDate,
        weekDateIndex: 1,
        status: 'in_progress',
        elapsedSeconds: 10,
      ),
      sessions: <IepLessonSession>[
        IepLessonSession(
          lessonDate: lessonDate,
          weekDateIndex: 1,
          status: 'in_progress',
          elapsedSeconds: 10,
        ),
      ],
    );
  }

  @override
  Future<IepWordFile> downloadIepPlanWord(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required IepPlan plan,
  }) async {
    return _fakeIepWordFile('${record.studentName}-IEP.docx');
  }

  @override
  Future<IepWordFile> downloadMonthlyPlanWord(
    String token, {
    required IepAssessmentRecordSummary record,
    required IepMonthlyPlan plan,
  }) async {
    return _fakeIepWordFile('${record.studentName}-月计划.docx');
  }

  @override
  Future<IepWordFile> downloadWeeklyPlanWord(
    String token, {
    required IepAssessmentRecordSummary record,
    required IepWeeklyPlan plan,
  }) async {
    return _fakeIepWordFile('${record.studentName}-周计划.docx');
  }

  @override
  Future<Uint8List> downloadIepPlanPdf(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required IepPlan plan,
  }) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadMonthlyPlanPdf(
    String token, {
    required IepAssessmentRecordSummary record,
    required IepMonthlyPlan plan,
  }) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadWeeklyPlanPdf(
    String token, {
    required IepAssessmentRecordSummary record,
    required IepWeeklyPlan plan,
  }) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }
}

class _ShuangxiCaptureIepPlanClient extends _EmptyThenGeneratedIepPlanClient {
  String createdRecordSource = '';
  String createdRecordCode = '';

  @override
  Future<IepPlanGenerationTask> createIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) {
    createdRecordSource = record.source;
    createdRecordCode = record.assessmentCode;
    return super.createIepPlanGenerationTask(
      token,
      record: record,
      durationMonths: durationMonths,
    );
  }

  @override
  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async* {
    yield IepPlanGenerationEvent.status('正在读取双溪评估结果');
    final IepPlan plan = IepPlan(
      title: '康复教学季度计划',
      student: IepPlanStudent(
        name: record.studentName,
        gender: record.studentGender,
        birthDate: record.birthDate,
      ),
      meta: IepPlanMeta(
        planDate: record.assessmentDate,
        participant: record.examinerName,
        implementer: record.examinerName,
        startDate: '2026-05-01',
        endDate: '2026-07-31',
      ),
      rows: const <IepPlanRow>[
        IepPlanRow(
          domain: '生活自理',
          longGoal: '提升日常生活自理能力',
          shortGoal: '能完成双溪课程目标',
          courseForm: '个训',
          startEndDate: '2026-05-01 - 2026-05-31',
        ),
      ],
    );
    yield IepPlanGenerationEvent.done(
      plan,
      savedPlan: IepPlanSaved(
        exists: true,
        status: 'draft',
        durationMonths: 3,
        plan: plan,
        updatedTime: '2026-05-18T14:00:00Z',
      ),
    );
  }
}

IepWordFile _fakeIepWordFile(String fileName) {
  return IepWordFile(
    fileName: fileName,
    contentType:
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    bytes: const <int>[0x50, 0x4B, 0x03, 0x04],
    fallbackName: fileName,
  );
}

String _formatDateDashForTest(DateTime value) {
  return '${value.year.toString().padLeft(4, '0')}-'
      '${value.month.toString().padLeft(2, '0')}-'
      '${value.day.toString().padLeft(2, '0')}';
}

class _FakeAssessmentScaleClient implements AssessmentScaleClient {
  _FakeAssessmentScaleClient({
    this.categoriesDelay = Duration.zero,
    this.libraryDelay = Duration.zero,
    this.studentCandidatesDelay = Duration.zero,
    AssessmentDraftPage? draftPage,
    List<AssessmentStudentCandidate>? studentCandidates,
    Map<int, List<AssessmentStudentCandidate>>? studentCandidatesByStatus,
    List<AssessmentScaleItem>? scaleItems,
  })  : studentCandidates = studentCandidates ?? _defaultStudentCandidates,
        studentCandidatesByStatus = studentCandidatesByStatus ??
            <int, List<AssessmentStudentCandidate>>{
              AssessmentStudentStatuses.enrolled:
                  studentCandidates ?? _defaultStudentCandidates,
            },
        scaleItems = scaleItems ?? _items,
        draftPage = draftPage ?? _defaultDraftPage;

  final Duration categoriesDelay;
  final Duration libraryDelay;
  final Duration studentCandidatesDelay;
  final AssessmentDraftPage draftPage;
  final List<AssessmentScaleItem> scaleItems;
  final List<AssessmentStudentCandidate> studentCandidates;
  final Map<int, List<AssessmentStudentCandidate>> studentCandidatesByStatus;
  final List<int> requestedStudentStatuses = <int>[];
  int updateStudentBirthDateCount = 0;
  int lastUpdatedStudentId = 0;
  String lastUpdatedBirthDate = '';

  static const List<AssessmentScaleItem> _items = <AssessmentScaleItem>[
    AssessmentScaleItem(
      id: 1,
      name: 'PEP-3语言理解评核量表',
      code: 'PEP3-CVP',
      category: '语言与沟通能力',
      scenario: '语言沟通',
      ageRange: '2-7岁',
      ageMinMonths: 24,
      ageMaxMonths: 84,
      duration: '25分钟',
      durationMinMinutes: 20,
      durationMaxMinutes: 30,
      currentVersion: '2026',
      itemCount: 56,
      domainCount: 1,
      monthUsage: 5,
      usageCount: 12,
      latestUse: '2026-05-04',
      dataStatus: 'ready',
      status: 'available',
      statusText: '可用',
      updatedAt: '2026-05-04 10:00:00',
      summary: '语言理解评核',
      posterUrl: '',
      executionEntry: 'pep3',
      apiPackage: 'pep3',
    ),
    AssessmentScaleItem(
      id: 2,
      name: '口语发起与互动',
      code: 'ORAL-INT',
      category: '语言与沟通能力',
      scenario: '口语表达',
      ageRange: '3-8岁',
      ageMinMonths: 36,
      ageMaxMonths: 96,
      duration: '18分钟',
      durationMinMinutes: 15,
      durationMaxMinutes: 20,
      currentVersion: '2026',
      itemCount: 42,
      domainCount: 1,
      monthUsage: 2,
      usageCount: 8,
      latestUse: '2026-05-03',
      dataStatus: 'ready',
      status: 'available',
      statusText: '可用',
      updatedAt: '2026-05-03 10:00:00',
      summary: '口语互动',
      posterUrl: '',
      executionEntry: 'oral',
      apiPackage: 'oral',
    ),
  ];

  static const AssessmentDraftPage _defaultDraftPage = AssessmentDraftPage(
    total: 1,
    current: 1,
    size: 5,
    items: <AssessmentDraftSummary>[
      AssessmentDraftSummary(
        id: 9,
        studentName: '张一鸣',
        assessmentCode: 'PEP3',
        assessmentName: 'PEP-3',
        scaleVersion: '2026',
        examinerName: '陈老师',
        status: 'draft',
        answeredItemCount: 24,
        rawScoreCount: 0,
        completionPercent: .42,
        progressItemCount: 0,
        progressQuestionDisplayPreference: '',
        createdTime: '2026-05-04T09:00:00Z',
        updatedTime: '2026-05-04T10:00:00Z',
      ),
    ],
  );

  @override
  Future<List<String>> fetchCategories(String token) async {
    if (categoriesDelay > Duration.zero) {
      await Future<void>.delayed(categoriesDelay);
    }
    return const <String>['语言与沟通能力', '社交情绪评估'];
  }

  @override
  Future<AssessmentScaleLibrary> fetchScaleLibrary(
    String token, {
    String keyword = '',
    String category = '',
  }) async {
    if (libraryDelay > Duration.zero) {
      await Future<void>.delayed(libraryDelay);
    }
    final String normalizedKeyword = _fakeNormalize(keyword);
    final List<AssessmentScaleItem> filtered = scaleItems.where(
      (AssessmentScaleItem item) {
        if (category.trim().isNotEmpty && item.category != category.trim()) {
          return false;
        }
        if (normalizedKeyword.isEmpty) {
          return true;
        }
        final String target = _fakeNormalize(
          <String>[
            item.name,
            item.code,
            item.category,
            item.scenario,
            item.ageRange,
            item.duration,
          ].join(' '),
        );
        return target.contains(normalizedKeyword);
      },
    ).toList();
    return AssessmentScaleLibrary(
      items: filtered,
      summary: AssessmentScaleLibrarySummary(
        total: scaleItems.length,
        available: scaleItems.where((AssessmentScaleItem item) {
          return item.available;
        }).length,
        unavailable: 0,
        monthUsage: 7,
        usageCount: 20,
      ),
      filterOptions: const AssessmentScaleFilterOptions(
        categories: <String>['语言与沟通能力', '社交情绪评估'],
        categoryCounts: <String, int>{
          '语言与沟通能力': 2,
          '社交情绪评估': 0,
        },
        scenarios: <String>['语言沟通', '口语表达'],
        statuses: <String>['available'],
      ),
    );
  }

  @override
  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    bool latestOnly = false,
  }) async {
    return draftPage;
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
    if (studentCandidatesDelay > Duration.zero) {
      await Future<void>.delayed(studentCandidatesDelay);
    }
    requestedStudentStatuses.add(studentStatus);
    final List<AssessmentStudentCandidate> items =
        studentCandidatesByStatus[studentStatus] ??
            <AssessmentStudentCandidate>[];
    return AssessmentStudentCandidatePage(
      total: items.length,
      current: 1,
      size: pageSize,
      items: items,
    );
  }

  @override
  Future<String> updateStudentBirthDate(
    String token, {
    required int studentId,
    required String birthDate,
  }) async {
    updateStudentBirthDateCount += 1;
    lastUpdatedStudentId = studentId;
    lastUpdatedBirthDate = birthDate;
    return birthDate;
  }
}

const List<AssessmentStudentCandidate> _defaultStudentCandidates =
    <AssessmentStudentCandidate>[
  AssessmentStudentCandidate(
    id: 3,
    shortName: '张',
    name: '张一鸣',
    avatarUrl: '',
    gender: '男',
    age: '5岁2个月',
    birthDate: '2021-03-01',
    contactPhone: '妈妈 136****0001',
    latestAssessment: '未测评',
  ),
  AssessmentStudentCandidate(
    id: 4,
    shortName: '王',
    name: '王安全',
    avatarUrl: '',
    gender: '男',
    age: '',
    birthDate: '',
    contactPhone: '爸爸 136****0002',
    latestAssessment: '未测评',
  ),
];

const AssessmentScaleItem _pep3ScaleItem = AssessmentScaleItem(
  id: 1,
  name: 'PEP-3语言理解评核量表',
  code: 'PEP3-CVP',
  category: '语言与沟通能力',
  scenario: '语言沟通',
  ageRange: '2岁6个月-7岁5个月',
  ageMinMonths: 30,
  ageMaxMonths: 89,
  duration: '25分钟',
  durationMinMinutes: 20,
  durationMaxMinutes: 30,
  currentVersion: '2026',
  itemCount: 56,
  domainCount: 1,
  monthUsage: 5,
  usageCount: 12,
  latestUse: '2026-05-04',
  dataStatus: 'ready',
  status: 'available',
  statusText: '可用',
  updatedAt: '2026-05-04 10:00:00',
  summary: '语言理解评核',
  posterUrl: '',
  executionEntry: 'pep3',
  apiPackage: 'pep3',
);

const AssessmentScaleItem _autismDevScaleItem = AssessmentScaleItem(
  id: 4,
  name: '孤独症儿童发展评估表',
  code: 'AUTISMDEV',
  category: '标准化测评',
  scenario: '现场测评',
  ageRange: '0岁-6岁',
  ageMinMonths: 0,
  ageMaxMonths: 72,
  duration: '40-60分钟',
  durationMinMinutes: 40,
  durationMaxMinutes: 60,
  currentVersion: '2010-revised-trainer',
  itemCount: 133,
  domainCount: 8,
  monthUsage: 0,
  usageCount: 0,
  latestUse: '',
  dataStatus: 'ready',
  status: 'available',
  statusText: '可用',
  updatedAt: '2026-05-13 10:00:00',
  summary: '孤独症儿童发展评估',
  posterUrl: '',
  executionEntry: 'autismdev',
  apiPackage: 'autismdev',
);

const AssessmentScaleItem _shuangxiAScaleItem = AssessmentScaleItem(
  id: 5,
  name: '双溪课程评量表A',
  code: 'SHUANGXI_A',
  category: '标准化测评',
  scenario: '现场测评',
  ageRange: '0岁-18岁',
  ageMinMonths: 0,
  ageMaxMonths: 216,
  duration: '40-60分钟',
  durationMinMinutes: 40,
  durationMaxMinutes: 60,
  currentVersion: '2026',
  itemCount: 209,
  domainCount: 7,
  monthUsage: 0,
  usageCount: 0,
  latestUse: '',
  dataStatus: 'ready',
  status: 'available',
  statusText: '可用',
  updatedAt: '2026-05-18 10:00:00',
  summary: '双溪课程评量表A',
  posterUrl: '',
  executionEntry: 'shuangxi-a',
  apiPackage: 'shuangxi',
);

const AssessmentScaleItem _erxinScaleItem = AssessmentScaleItem(
  id: 3,
  name: '0岁～6岁儿童发育行为评估量表（儿心量表-II）',
  code: 'ERXIN2',
  category: '标准化测评',
  scenario: '现场测评',
  ageRange: '0岁-6岁',
  ageMinMonths: 0,
  ageMaxMonths: 72,
  duration: '20-40分钟',
  durationMinMinutes: 20,
  durationMaxMinutes: 40,
  currentVersion: 'WS-T-580-2017',
  itemCount: 261,
  domainCount: 5,
  monthUsage: 0,
  usageCount: 0,
  latestUse: '',
  dataStatus: 'ready',
  status: 'available',
  statusText: '可用',
  updatedAt: '2026-05-08 10:00:00',
  summary: '儿心量表-II',
  posterUrl: '',
  executionEntry: 'erxin',
  apiPackage: 'erxin',
);

const AssessmentScaleItem _vbmappScaleItem = AssessmentScaleItem(
  id: 6,
  name: 'VB-MAPP语言行为里程碑评估及安置计划',
  code: 'VBMAPP',
  category: '语言行为评估',
  scenario: '现场测评',
  ageRange: '0岁-4岁',
  ageMinMonths: 0,
  ageMaxMonths: 48,
  duration: '60-120分钟',
  durationMinMinutes: 60,
  durationMaxMinutes: 120,
  currentVersion: 'VBMAPP_CN_2ND_DRAFT_2026_05',
  itemCount: 212,
  domainCount: 16,
  monthUsage: 0,
  usageCount: 0,
  latestUse: '',
  dataStatus: 'ready',
  status: 'available',
  statusText: '可用',
  updatedAt: '2026-05-19 10:00:00',
  summary: 'VB-MAPP',
  posterUrl: '',
  executionEntry: 'Pad /vbmapp-assessment',
  apiPackage: '/api/v1/assessments/vbmapp/*',
);

class _FakeVbmappAssessmentClient implements VbmappAssessmentClient {
  int saveDraftItemCalls = 0;

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
  Future<VbmappDraftDetail> fetchDraftDetail(String token, int id) {
    throw UnimplementedError();
  }

  @override
  Future<VbmappAssessmentSchema> fetchAssessmentSchema(String token) async {
    return const VbmappAssessmentSchema(
      scaleVersion: 'VBMAPP_CN_2ND_DRAFT_2026_05',
      itemSchemas: <String, VbmappItemResponseSchema>{
        'milestones::MAND_01M': VbmappItemResponseSchema(
          moduleCode: 'milestones',
          itemCode: 'MAND_01M',
          uiPattern: 'mand_event_recorder',
          recordDepth: 'structured_event_log',
          materialProfileId: 'mand_1m_request_starter_set',
          whyRecord: '',
          evidenceTargets: <String>[],
          qualityChecks: <String>[],
          scoreStrategy: 'count_qualified_unique_mand_events',
          onePointCriteria: '',
          halfPointCriteria: '',
        ),
        'milestones::MAND_02M': VbmappItemResponseSchema(
          moduleCode: 'milestones',
          itemCode: 'MAND_02M',
          uiPattern: 'mand_event_recorder',
          recordDepth: 'structured_event_log',
          materialProfileId: 'mand_2m_visible_request_set',
          whyRecord: '',
          evidenceTargets: <String>[],
          qualityChecks: <String>[],
          scoreStrategy: 'count_qualified_unique_mand_events',
          onePointCriteria: '',
          halfPointCriteria: '',
        ),
      },
      materialProfiles: <String, VbmappMaterialProfile>{
        'mand_1m_request_starter_set': VbmappMaterialProfile(
          label: '提要求1M入门强化物/动作',
          suggestedTypes: <String>['食物/饮料', '实物/活动', '动作/帮助'],
          recommendedMaterials: <VbmappMaterialSuggestion>[
            VbmappMaterialSuggestion(
                id: 'test-cookie', name: '饼干', type: '食物/饮料'),
            VbmappMaterialSuggestion(id: 'test-book', name: '书', type: '实物/活动'),
            VbmappMaterialSuggestion(
                id: 'test-open', name: '打开', type: '动作/帮助'),
          ],
          preparationChecks: <String>[],
        ),
        'mand_2m_visible_request_set': VbmappMaterialProfile(
          label: '提要求2M可见强化物/活动',
          suggestedTypes: <String>['活动', '实物玩具', '社交游戏'],
          recommendedMaterials: <VbmappMaterialSuggestion>[
            VbmappMaterialSuggestion(id: 'test-music', name: '音乐', type: '活动'),
            VbmappMaterialSuggestion(
                id: 'test-slinky', name: '彩虹弹簧', type: '实物玩具'),
            VbmappMaterialSuggestion(id: 'test-ball', name: '球', type: '实物玩具'),
            VbmappMaterialSuggestion(
                id: 'test-bubbles', name: '泡泡', type: '社交游戏'),
          ],
          preparationChecks: <String>[],
        ),
      },
    );
  }

  @override
  Future<VbmappDraftSaveResult> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    return VbmappDraftSaveResult(
      id: 17,
      studentId: (payload['studentId'] as num?)?.toInt() ?? 0,
      studentName: '${payload['studentName'] ?? ''}',
      assessmentDate: '${payload['assessmentDate'] ?? ''}',
      examinerName: '${payload['examinerName'] ?? ''}',
      status: 'draft',
      answeredItemCount: 1,
      completionPercent: 0.01,
    );
  }

  @override
  Future<VbmappDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    saveDraftItemCalls += 1;
    return VbmappDraftDetail(
      id: (payload['draftId'] as num?)?.toInt() ?? 17,
      studentId: 0,
      studentName: '',
      birthDate: '',
      assessmentDate: '',
      examinerName: '',
      milestoneScores: const <String, double>{},
      barrierScores: const <String, int>{},
      transitionScores: const <String, int>{},
      itemResponses: const <String, Map<String, Map<String, dynamic>>>{},
    );
  }

  @override
  Future<VbmappDraftSubmitResult> submitDraft(String token, int id) async {
    return VbmappDraftSubmitResult(
      draftId: id,
      recordId: 27,
      draftStatus: 'submitted',
    );
  }
}

class _FakeErxinAssessmentClient implements ErxinAssessmentClient {
  _FakeErxinAssessmentClient({
    this.templateSummaryDelay = Duration.zero,
    this.saveDraftDelay = Duration.zero,
    this.draftDetailDelay = Duration.zero,
    this.interpretationDelay = Duration.zero,
    this.interpretationFetchDelay = Duration.zero,
    ErxinReportInterpretation savedInterpretation =
        ErxinReportInterpretation.empty,
    Uint8List? reportPdfBytes,
    this.detectedDraft,
    this.draftDetail,
    List<ErxinAgeGroup>? groups,
  })  : _savedInterpretation = savedInterpretation,
        reportPdfBytes =
            reportPdfBytes ?? Uint8List.fromList(<int>[37, 80, 68, 70]),
        groups = groups ?? _groups;

  final Duration templateSummaryDelay;
  final Duration saveDraftDelay;
  final Duration draftDetailDelay;
  final Duration interpretationDelay;
  final Duration interpretationFetchDelay;
  final Uint8List reportPdfBytes;
  final AssessmentDraftSummary? detectedDraft;
  final ErxinDraftDetail? draftDetail;
  final List<ErxinAgeGroup> groups;
  ErxinReportInterpretation _savedInterpretation;

  int saveDraftCalls = 0;
  int saveDraftItemCalls = 0;
  int submitDraftCalls = 0;
  int fetchDraftsPageCalls = 0;
  int fetchDraftDetailCalls = 0;
  int fetchInterpretationCalls = 0;
  int generateInterpretationCalls = 0;
  int nextDraftId = 21;
  int _latestDraftId = 0;
  bool failNextDraftUpdateAsNotFound = false;
  final List<Map<String, dynamic>> saveDraftPayloads = <Map<String, dynamic>>[];
  final List<Map<String, dynamic>> saveDraftItemPayloads =
      <Map<String, dynamic>>[];
  final Map<int, bool> _savedItemPasses = <int, bool>{};
  final Map<int, String> _savedItemRemarks = <int, String>{};
  AssessmentDraftSummary? _latestDraftSummary;

  static const ErxinDomain _domain = ErxinDomain(
    domainCode: 'GM',
    domainName: '大运动',
    sortNo: 1,
  );

  static const List<ErxinAgeGroup> _groups = <ErxinAgeGroup>[
    ErxinAgeGroup(
      ageMonth: 30,
      title: '30月龄',
      items: <ErxinItemSummary>[
        ErxinItemSummary(
          itemNo: 130,
          itemTitle: '30月题',
          testItem: '30月题',
          ageMonth: 30,
          domainCode: 'GM',
          domainName: '大运动',
          parentReportAllowed: false,
          attentionIfFailed: false,
        ),
      ],
    ),
    ErxinAgeGroup(
      ageMonth: 33,
      title: '33月龄',
      items: <ErxinItemSummary>[
        ErxinItemSummary(
          itemNo: 133,
          itemTitle: '33月题',
          testItem: '33月题',
          ageMonth: 33,
          domainCode: 'GM',
          domainName: '大运动',
          parentReportAllowed: false,
          attentionIfFailed: false,
        ),
      ],
    ),
    ErxinAgeGroup(
      ageMonth: 36,
      title: '36月龄',
      items: <ErxinItemSummary>[
        ErxinItemSummary(
          itemNo: 136,
          itemTitle: '36月题',
          testItem: '36月题',
          ageMonth: 36,
          domainCode: 'GM',
          domainName: '大运动',
          parentReportAllowed: false,
          attentionIfFailed: false,
        ),
      ],
    ),
    ErxinAgeGroup(
      ageMonth: 42,
      title: '42月龄',
      items: <ErxinItemSummary>[
        ErxinItemSummary(
          itemNo: 142,
          itemTitle: '42月题',
          testItem: '42月题',
          ageMonth: 42,
          domainCode: 'GM',
          domainName: '大运动',
          parentReportAllowed: false,
          attentionIfFailed: false,
        ),
      ],
    ),
    ErxinAgeGroup(
      ageMonth: 48,
      title: '48月龄',
      items: <ErxinItemSummary>[
        ErxinItemSummary(
          itemNo: 148,
          itemTitle: '48月题',
          testItem: '48月题',
          ageMonth: 48,
          domainCode: 'GM',
          domainName: '大运动',
          parentReportAllowed: false,
          attentionIfFailed: false,
        ),
      ],
    ),
    ErxinAgeGroup(
      ageMonth: 54,
      title: '54月龄',
      items: <ErxinItemSummary>[
        ErxinItemSummary(
          itemNo: 154,
          itemTitle: '54月题',
          testItem: '54月题',
          ageMonth: 54,
          domainCode: 'GM',
          domainName: '大运动',
          parentReportAllowed: false,
          attentionIfFailed: false,
        ),
      ],
    ),
    ErxinAgeGroup(
      ageMonth: 60,
      title: '60月龄',
      items: <ErxinItemSummary>[
        ErxinItemSummary(
          itemNo: 160,
          itemTitle: '60月题',
          testItem: '60月题',
          ageMonth: 60,
          domainCode: 'GM',
          domainName: '大运动',
          parentReportAllowed: false,
          attentionIfFailed: false,
        ),
      ],
    ),
  ];

  @override
  Future<ErxinTemplateSummary> fetchTemplateSummary(String token) async {
    if (templateSummaryDelay > Duration.zero) {
      await Future<void>.delayed(templateSummaryDelay);
    }
    final int itemCount = groups.fold<int>(
      0,
      (int total, ErxinAgeGroup group) => total + group.items.length,
    );
    return ErxinTemplateSummary(
      templateCode: 'ERXIN2_ASSESSMENT_FORM',
      title: '儿心量表-II测评录入表',
      scaleCode: 'ERXIN2',
      scaleVersion: 'WS-T-580-2017',
      itemCount: itemCount,
      domains: const <ErxinDomain>[_domain],
      ageGroups: groups,
    );
  }

  @override
  Future<ErxinAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  }) async {
    for (final ErxinAgeGroup group in groups) {
      for (final ErxinItemSummary item in group.items) {
        if (item.itemNo == itemNo) {
          return ErxinAssessmentItem(
            itemNo: item.itemNo,
            itemTitle: item.itemTitle,
            testItem: item.testItem,
            ageMonth: item.ageMonth,
            domainCode: item.domainCode,
            domainName: item.domainName,
            parentReportAllowed: item.parentReportAllowed,
            attentionIfFailed: item.attentionIfFailed,
            method: '主试者示范后请儿童完成动作。',
            passCriteria: '儿童可独立完成即通过。',
          );
        }
      }
    }
    return ErxinAssessmentItem.empty;
  }

  @override
  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  }) async {
    fetchDraftsPageCalls += 1;
    final AssessmentDraftSummary? latest = _latestDraftSummary;
    if (latest != null) {
      return AssessmentDraftPage(
        total: 1,
        current: pageIndex,
        size: pageSize,
        items: <AssessmentDraftSummary>[latest],
      );
    }
    final AssessmentDraftSummary? draft = detectedDraft;
    if (draft == null) {
      return AssessmentDraftPage.empty;
    }
    return AssessmentDraftPage(
      total: 1,
      current: pageIndex,
      size: pageSize,
      items: <AssessmentDraftSummary>[draft],
    );
  }

  @override
  Future<ErxinDraftDetail> fetchDraftDetail(String token, int id) async {
    fetchDraftDetailCalls += 1;
    if (draftDetailDelay > Duration.zero) {
      await Future<void>.delayed(draftDetailDelay);
    }
    if (_latestDraftSummary != null && _latestDraftSummary!.id == id) {
      return _draftDetail(id: id, input: _savedInput());
    }
    if (draftDetail != null && draftDetail!.id == id) {
      return draftDetail!;
    }
    return _draftDetail(id: id);
  }

  @override
  Future<ErxinDraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    saveDraftCalls += 1;
    saveDraftPayloads.add(Map<String, dynamic>.from(payload));
    if (saveDraftDelay > Duration.zero) {
      await Future<void>.delayed(saveDraftDelay);
    }
    final Object? id = payload['id'];
    if (failNextDraftUpdateAsNotFound && id is int && id > 0) {
      failNextDraftUpdateAsNotFound = false;
      throw const AssessmentScaleApiException('assessment draft not found');
    }
    _replaceSavedInput(_inputFromPayload(payload));
    if (id is int && id > 0) {
      final ErxinDraftDetail detail =
          _draftDetail(id: id, input: _savedInput());
      _latestDraftId = id;
      _latestDraftSummary = _draftSummary(id, detail);
      return detail;
    }
    final int idToReturn = nextDraftId;
    nextDraftId += 1;
    final ErxinDraftDetail detail =
        _draftDetail(id: idToReturn, input: _savedInput());
    _latestDraftId = idToReturn;
    _latestDraftSummary = _draftSummary(idToReturn, detail);
    return detail;
  }

  @override
  Future<ErxinDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    saveDraftItemCalls += 1;
    saveDraftItemPayloads.add(Map<String, dynamic>.from(payload));
    final int itemNo = _fakeIntFrom(payload['itemNo']);
    final Object? passed = payload['passed'];
    if (itemNo > 0 && passed is bool) {
      _savedItemPasses[itemNo] = passed;
    }
    final String remark = '${payload['remark'] ?? ''}'.trim();
    if (itemNo > 0) {
      if (remark.isEmpty) {
        _savedItemRemarks.remove(itemNo);
      } else {
        _savedItemRemarks[itemNo] = remark;
      }
    }
    final int draftId = _latestDraftId > 0 ? _latestDraftId : _lastDraftId;
    final ErxinDraftDetail detail = _draftDetail(
      id: draftId,
      input: _savedInput(),
    );
    if (_latestDraftId > 0) {
      _latestDraftSummary = _draftSummary(draftId, detail);
    }
    return detail;
  }

  @override
  Future<void> submitDraft(String token, int draftId) async {
    submitDraftCalls += 1;
  }

  @override
  Future<ErxinRecordDetail> fetchRecordDetail(String token, int id) async {
    final ErxinDraftDetail detail = await fetchDraftDetail(token, id);
    return ErxinRecordDetail(
      id: detail.id,
      studentId: detail.studentId,
      studentName: detail.studentName,
      assessmentCode: 'ERXIN2',
      assessmentName: '儿心量表-II',
      birthDate: detail.birthDate,
      assessmentDate: detail.assessmentDate,
      examinerName: detail.examinerName,
      updatedTime: detail.updatedTime,
      scaleVersion: 'WS-T-580-2017',
      input: detail.input,
    );
  }

  @override
  Future<ErxinRecordDetail> updateRecordConfig(
    String token,
    int id, {
    required String examinerName,
    required String assessmentDate,
  }) async {
    final ErxinDraftInput current = _savedInput();
    final ErxinDraftInput next = ErxinDraftInput(
      studentId: 31,
      studentName: '陈旭',
      examinerName: examinerName,
      remark: current.remark,
      birthDate: '2022-05-11',
      assessmentDate: assessmentDate,
      itemPasses: Map<int, bool>.from(current.itemPasses),
      itemRemarks: Map<int, String>.from(current.itemRemarks),
    );
    _replaceSavedInput(next);
    return ErxinRecordDetail(
      id: id,
      studentId: 31,
      studentName: '陈旭',
      assessmentCode: 'ERXIN2',
      assessmentName: '儿心量表-II',
      birthDate: '2022-05-11',
      assessmentDate: assessmentDate,
      examinerName: examinerName,
      updatedTime: '2026-05-08T10:00:00',
      scaleVersion: 'WS-T-580-2017',
      input: next,
    );
  }

  @override
  Future<Uint8List> downloadRecordReportPdf(String token, int id) async {
    return reportPdfBytes;
  }

  @override
  Future<Uint8List> downloadRecordReportInterpretationPdf(
    String token,
    int id,
  ) async {
    return reportPdfBytes;
  }

  @override
  Future<ErxinReportInterpretation> fetchRecordReportInterpretation(
    String token,
    int id,
  ) async {
    fetchInterpretationCalls += 1;
    if (interpretationFetchDelay > Duration.zero) {
      await Future<void>.delayed(interpretationFetchDelay);
    }
    return _savedInterpretation;
  }

  @override
  Future<ErxinReportInterpretation> generateRecordReportInterpretation(
    String token,
    int id,
  ) async {
    generateInterpretationCalls += 1;
    if (interpretationDelay > Duration.zero) {
      await Future<void>.delayed(interpretationDelay);
    }
    _savedInterpretation = const ErxinReportInterpretation(
      title: '报告解读',
      generatedBy: 'rule',
      summary: '本次测评显示儿童整体发育水平需结合日常观察综合判断。',
      domainAnalysis: <String>['大运动表现相对稳定。'],
      suggestions: <String>['建议持续关注语言和社会行为表现。'],
      notes: <String>['本解读仅供参考。'],
    );
    return _savedInterpretation;
  }

  @override
  Stream<ErxinReportInterpretationStreamEvent>
      generateRecordReportInterpretationStream(String token, int id) async* {
    generateInterpretationCalls += 1;
    yield const ErxinReportInterpretationStreamEvent(
      type: 'status',
      message: '正在读取儿心评估结果',
    );
    await Future<void>.delayed(const Duration(milliseconds: 40));
    final String json = jsonEncode(<String, Object>{
      'title': '报告解读',
      'generatedBy': 'ai',
      'summary': '本次测评显示儿童整体发育水平需结合日常观察综合判断。',
      'domainAnalysis': <String>['大运动表现相对稳定。'],
      'suggestions': <String>['建议持续关注语言和社会行为表现。'],
      'notes': <String>['本解读仅供参考。'],
    });
    final int splitIndex = json.indexOf('日常观察');
    yield ErxinReportInterpretationStreamEvent(
      type: 'delta',
      text: json.substring(0, splitIndex),
    );
    if (interpretationDelay > Duration.zero) {
      await Future<void>.delayed(interpretationDelay);
    }
    yield ErxinReportInterpretationStreamEvent(
      type: 'delta',
      text: json.substring(splitIndex),
    );
    _savedInterpretation = ErxinReportInterpretation.fromJson(
      Map<String, dynamic>.from(jsonDecode(json) as Map),
    );
    yield ErxinReportInterpretationStreamEvent(
      type: 'done',
      data: _savedInterpretation,
    );
  }

  int get _lastDraftId => nextDraftId <= 21 ? 21 : nextDraftId - 1;

  AssessmentDraftSummary _draftSummary(
    int id,
    ErxinDraftDetail detail,
  ) {
    return AssessmentDraftSummary(
      id: id,
      studentName: detail.studentName,
      assessmentCode: 'ERXIN2',
      assessmentName: '儿心量表-II',
      scaleVersion: 'WS-T-580-2017',
      examinerName: detail.examinerName,
      status: 'draft',
      answeredItemCount: detail.answeredItemCount,
      rawScoreCount: 0,
      completionPercent: detail.completionPercent,
      progressItemCount: 0,
      progressQuestionDisplayPreference: '',
      createdTime: detail.updatedTime,
      updatedTime: detail.updatedTime,
    );
  }

  void _replaceSavedInput(ErxinDraftInput input) {
    _savedItemPasses
      ..clear()
      ..addAll(input.itemPasses);
    _savedItemRemarks
      ..clear()
      ..addAll(input.itemRemarks);
  }

  ErxinDraftInput _savedInput() {
    return ErxinDraftInput(
      itemPasses: Map<int, bool>.from(_savedItemPasses),
      itemRemarks: Map<int, String>.from(_savedItemRemarks),
    );
  }

  ErxinDraftInput _inputFromPayload(Map<String, dynamic> payload) {
    return ErxinDraftInput.fromJson(<String, dynamic>{
      'itemPassList': payload['itemPassList'],
      'itemRemarkList': payload['itemRemarkList'],
    });
  }

  ErxinDraftDetail _draftDetail({
    required int id,
    ErxinDraftInput input = ErxinDraftInput.empty,
  }) {
    final int answeredItemCount = input.itemPasses.length;
    return ErxinDraftDetail(
      id: id,
      studentId: 31,
      studentName: '陈旭',
      birthDate: '2022-05-11',
      assessmentDate: '2026-05-08',
      examinerName: '陈老师',
      answeredItemCount: answeredItemCount,
      completionPercent: answeredItemCount / 7,
      updatedTime: '2026-05-08T10:00:00',
      progress: ErxinDraftProgress(
        itemCount: 7,
        answeredItemCount: answeredItemCount,
        missingItemCount: 0,
        completionPercent: answeredItemCount / 7,
        complete: false,
        canScore: false,
        missingItemNos: <int>[],
      ),
      input: input,
    );
  }
}

const List<int> _erxinBoundaryAgeMonths = <int>[
  1,
  2,
  3,
  4,
  5,
  6,
  7,
  8,
  9,
  10,
  11,
  12,
  15,
  18,
  21,
  24,
  27,
  30,
  33,
  36,
  42,
  48,
  54,
  60,
  66,
  72,
  78,
  84,
];

List<ErxinAgeGroup> _erxinBoundaryGroups() {
  return <ErxinAgeGroup>[
    for (final int month in _erxinBoundaryAgeMonths)
      ErxinAgeGroup(
        ageMonth: month,
        title: '$month月龄',
        items: <ErxinItemSummary>[
          ErxinItemSummary(
            itemNo: 1000 + month,
            itemTitle: '$month月题',
            testItem: '$month月题',
            ageMonth: month,
            domainCode: 'GM',
            domainName: '大运动',
            parentReportAllowed: false,
            attentionIfFailed: false,
          ),
        ],
      ),
  ];
}

Future<void> _tapErxinScore(
  WidgetTester tester,
  String itemTitle,
  bool passed,
) async {
  final Finder row = find.ancestor(
    of: find.text(itemTitle),
    matching: find.byWidgetPredicate(
      (Widget widget) => widget.runtimeType.toString() == '_ItemScoreRow',
    ),
  );
  expect(row, findsOneWidget);
  final Rect rect = tester.getRect(row);
  await tester.tapAt(
    Offset(rect.right - (passed ? 156 : 48), rect.center.dy),
  );
  await tester.pumpAndSettle();
}

Future<void> _tapErxinItemRow(
  WidgetTester tester,
  String itemTitle,
) async {
  final Finder row = find.ancestor(
    of: find.text(itemTitle),
    matching: find.byWidgetPredicate(
      (Widget widget) => widget.runtimeType.toString() == '_ItemScoreRow',
    ),
  );
  expect(row, findsOneWidget);
  await tester.tap(row);
  await tester.pumpAndSettle();
}

Future<void> _enterErxinRemark(
  WidgetTester tester,
  String remark,
) async {
  await tester.tap(find.text('添加本题备注').last);
  await tester.pumpAndSettle();
  expect(find.text('题目备注'), findsWidgets);
  await tester.enterText(find.byType(TextField).last, remark);
  await tester.pump();
  await tester.tap(find.widgetWithText(FilledButton, '完成'));
  await tester.pumpAndSettle();
}

class _FakePep3AssessmentClient implements Pep3AssessmentClient {
  _FakePep3AssessmentClient({
    this.hasDraft = false,
    this.hasPreviousRecord = false,
    this.previousRecord,
    this.draftUpdatedTime = '2026-05-05T14:01:00',
    this.includeRecordField = false,
    this.includeTextRecordField = false,
    this.longInstructions = false,
    this.summaryFetchDelay = Duration.zero,
    this.itemFetchDelay = Duration.zero,
    this.draftDetailDelay = Duration.zero,
    this.inviteDelay = Duration.zero,
  });

  final bool hasDraft;
  final bool hasPreviousRecord;
  final Pep3RecordSummary? previousRecord;
  final String draftUpdatedTime;
  final bool includeRecordField;
  final bool includeTextRecordField;
  final bool longInstructions;
  final Duration summaryFetchDelay;
  final Duration itemFetchDelay;
  final Duration draftDetailDelay;
  final Duration inviteDelay;
  int saveDraftCalls = 0;
  int saveDraftItemCalls = 0;
  int inviteCalls = 0;
  int inviteCompletedCalls = 0;
  int fetchAutismDevResultAnalysisCalls = 0;
  int saveAutismDevResultAnalysisCalls = 0;
  int downloadShuangxiDevelopmentProfilePdfCalls = 0;
  int fetchShuangxiResultAnalysisCalls = 0;
  int saveShuangxiResultAnalysisCalls = 0;
  int updateRecordConfigCalls = 0;
  int lastUpdatedRecordConfigId = 0;
  String lastUpdatedRecordConfigExaminerName = '';
  String lastUpdatedRecordConfigAssessmentDate = '';
  AutismDevResultAnalysis? savedAutismDevResultAnalysis;
  ShuangxiResultAnalysis? savedShuangxiResultAnalysis;
  ErxinReportInterpretation savedAutismDevReportInterpretation =
      ErxinReportInterpretation.empty;

  static const List<Pep3ScoreOption> _scoreOptions = <Pep3ScoreOption>[
    Pep3ScoreOption(value: 2, label: '通过', description: '可独立完成'),
    Pep3ScoreOption(value: 1, label: '部分通过', description: '经提示可完成'),
    Pep3ScoreOption(value: 0, label: '未通过', description: '未能完成'),
  ];

  static const Pep3DraftProgress _progress = Pep3DraftProgress(
    itemCount: 2,
    answeredItemCount: 0,
    missingItemCount: 2,
    rawScoreCount: 0,
    caregiverRawScoreCount: 0,
    completionPercent: 0,
    complete: false,
    canScore: false,
    missingItemNos: <int>[1, 2],
  );

  @override
  Future<Pep3TemplateSummary> fetchTemplateSummary(String token) async {
    if (summaryFetchDelay > Duration.zero) {
      await Future<void>.delayed(summaryFetchDelay);
    }
    return const Pep3TemplateSummary(
      title: 'PEP-3儿童心理教育评核',
      itemCount: 2,
      scoreOptions: _scoreOptions,
      itemGroups: <Pep3ItemGroupSummary>[
        Pep3ItemGroupSummary(
          groupCode: 'page_1',
          title: '记录册第1页',
          bookletPageNo: 1,
          startItemNo: 1,
          endItemNo: 2,
          items: <Pep3ItemSummary>[
            Pep3ItemSummary(
              itemNo: 1,
              itemTitle: '（1） 旋开瓶盖',
              testItem: '旋开瓶盖',
              domainCode: 'FM',
              domainName: '小肌肉',
            ),
            Pep3ItemSummary(
              itemNo: 2,
              itemTitle: '（2） 叠积木',
              testItem: '叠积木',
              domainCode: 'FM',
              domainName: '小肌肉',
            ),
          ],
        ),
      ],
    );
  }

  @override
  Future<Pep3AssessmentItem> fetchTemplateItem(String token, int itemNo) async {
    if (itemFetchDelay > Duration.zero) {
      await Future<void>.delayed(itemFetchDelay);
    }
    final String repeatedInstruction = List<String>.filled(
      14,
      '观察儿童是否可以按标准完成任务，记录启动提示、动作过程、完成质量和需要辅助的环节。',
    ).join(' ');
    return Pep3AssessmentItem(
      itemNo: itemNo,
      itemTitle: itemNo == 1 ? '（1） 旋开瓶盖' : '（2） 叠积木',
      testItem: itemNo == 1 ? '旋开瓶盖' : '叠积木',
      domainCode: 'FM',
      domainName: '小肌肉',
      materials: itemNo == 1 ? '肥皂泡液' : '积木',
      materialImages: const <String>[],
      method: longInstructions ? repeatedInstruction : '观察儿童是否可以按标准完成任务。',
      guidance: longInstructions ? repeatedInstruction : '请你试试看。',
      guidanceVideo: '',
      standard: longInstructions ? repeatedInstruction : '',
      scoreOptions: _scoreOptions,
      recordFields: includeTextRecordField
          ? const <Pep3RecordField>[
              Pep3RecordField(
                key: 'trainingNote',
                label: '训练记录备注',
                fieldType: 'text',
                displayType: '',
                required: false,
                placeholder: '请输入训练记录',
                options: <Pep3RecordFieldOption>[],
              ),
            ]
          : includeRecordField
              ? const <Pep3RecordField>[
                  Pep3RecordField(
                    key: 'shape',
                    label: '正确位置',
                    fieldType: 'radio',
                    displayType: '',
                    required: false,
                    placeholder: '',
                    options: <Pep3RecordFieldOption>[
                      Pep3RecordFieldOption(value: 'triangle', label: '三角形'),
                      Pep3RecordFieldOption(value: 'square', label: '正方形'),
                    ],
                  ),
                ]
              : const <Pep3RecordField>[],
    );
  }

  @override
  Future<Pep3DraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  }) async {
    if (hasDraft) {
      return Pep3DraftPage(
        items: <Pep3DraftSummary>[
          Pep3DraftSummary(
            id: 11,
            studentId: 3,
            studentName: '张一鸣',
            birthDate: '2021-03-01',
            assessmentDate: '2026-05-05',
            examinerName: '陈老师',
            answeredItemCount: 1,
            completionPercent: .5,
            updatedTime: draftUpdatedTime,
            progress: const Pep3DraftProgress(
              itemCount: 2,
              answeredItemCount: 1,
              missingItemCount: 1,
              rawScoreCount: 0,
              caregiverRawScoreCount: 0,
              completionPercent: .5,
              complete: false,
              canScore: false,
              missingItemNos: <int>[2],
            ),
          ),
        ],
        total: 1,
        current: 1,
        size: 1,
      );
    }
    return const Pep3DraftPage(
      items: <Pep3DraftSummary>[],
      total: 0,
      current: 1,
      size: 0,
    );
  }

  @override
  Future<Pep3DraftDetail> fetchDraftDetail(String token, int id) async {
    if (draftDetailDelay > Duration.zero) {
      await Future<void>.delayed(draftDetailDelay);
    }
    return _draftDetail(id: id);
  }

  @override
  Future<Pep3DraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    saveDraftCalls += 1;
    return _draftDetail(id: 11);
  }

  @override
  Future<Pep3DraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    saveDraftItemCalls += 1;
    return _draftDetail(id: 11);
  }

  @override
  Future<Pep3CaregiverInvite> inviteCaregiverReport(
    String token,
    int draftId,
  ) async {
    inviteCalls += 1;
    if (inviteDelay > Duration.zero) {
      await Future<void>.delayed(inviteDelay);
    }
    inviteCompletedCalls += 1;
    return const Pep3CaregiverInvite(
      miniProgramCodeDataUrl: '',
      qrCodeValue: 'pep3-caregiver-report',
      wechatUrlLink: '',
      miniProgramPath: '',
      url: '',
    );
  }

  @override
  Future<void> submitDraft(String token, int draftId) async {}

  @override
  Future<Pep3RecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    int studentId = 0,
    String assessmentCode = 'PEP3',
    String scaleCategory = '',
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  }) async {
    if (!hasPreviousRecord) {
      return const Pep3RecordPage(
        items: <Pep3RecordSummary>[],
        total: 0,
        current: 1,
        size: 0,
      );
    }
    final Pep3RecordSummary record = previousRecord ??
        const Pep3RecordSummary(
          id: 21,
          studentId: 3,
          studentName: '张一鸣',
          assessmentCode: 'PEP3',
          assessmentName: 'PEP-3',
          birthDate: '2021-03-01',
          assessmentDate: '2026-05-04',
          examinerName: '陈老师',
          updatedTime: '2026-05-04T16:00:00',
        );
    return Pep3RecordPage(
      items: <Pep3RecordSummary>[record],
      total: 1,
      current: 1,
      size: 1,
    );
  }

  @override
  Future<Pep3RecordCategoryStats> fetchRecordCategoryStats(
    String token, {
    int studentId = 0,
    String assessmentCode = 'PEP3',
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  }) async {
    if (!hasPreviousRecord) {
      return Pep3RecordCategoryStats.empty;
    }
    return const Pep3RecordCategoryStats(
      total: 1,
      categoryCounts: <String, int>{'儿童发展评估': 1},
    );
  }

  @override
  Future<Pep3RecordDetail> fetchRecordDetail(String token, int id) async {
    return _recordDetailFromSummary(previousRecord, id: id);
  }

  @override
  Future<Pep3RecordDetail> updateRecordConfig(
    String token,
    int id, {
    required String examinerName,
    required String assessmentDate,
  }) async {
    updateRecordConfigCalls += 1;
    lastUpdatedRecordConfigId = id;
    lastUpdatedRecordConfigExaminerName = examinerName;
    lastUpdatedRecordConfigAssessmentDate = assessmentDate;
    return _recordDetailFromSummary(
      previousRecord,
      id: id,
      examinerName: examinerName,
      assessmentDate: assessmentDate,
    );
  }

  Pep3RecordDetail _recordDetailFromSummary(
    Pep3RecordSummary? summary, {
    required int id,
    String? examinerName,
    String? assessmentDate,
  }) {
    final Pep3RecordSummary record = summary ??
        const Pep3RecordSummary(
          id: 21,
          studentId: 3,
          studentName: '张一鸣',
          assessmentCode: 'PEP3',
          assessmentName: 'PEP-3',
          birthDate: '2021-03-01',
          assessmentDate: '2026-05-04',
          examinerName: '陈老师',
          updatedTime: '2026-05-04T16:00:00',
        );
    final String resolvedExaminer = examinerName ?? record.examinerName;
    final String resolvedAssessmentDate =
        assessmentDate ?? record.assessmentDate;
    return Pep3RecordDetail(
      id: id > 0 ? id : record.id,
      studentId: record.studentId,
      studentName: record.studentName,
      studentGender: record.studentGender,
      studentAvatar: record.studentAvatar,
      studentPhone: record.studentPhone,
      assessmentCode: record.assessmentCode,
      assessmentName: record.assessmentName,
      scaleCategory: record.scaleCategory,
      scaleVersion: record.scaleVersion,
      birthDate: record.birthDate,
      assessmentDate: resolvedAssessmentDate,
      ageYears: record.ageYears,
      ageMonths: record.ageMonths,
      ageDays: record.ageDays,
      normAgeMonths: record.normAgeMonths,
      assessmentSequence: record.assessmentSequence,
      examinerName: resolvedExaminer,
      createdTime: record.createdTime,
      updatedTime: '2026-05-07T10:00:00',
      input: Pep3DraftInput(
        studentId: record.studentId,
        studentName: record.studentName,
        examinerName: record.examinerName,
        birthDate: record.birthDate,
        assessmentDate: record.assessmentDate,
        remark: '',
        allowMissingItems: true,
        itemScores: const <int, int>{1: 0},
        itemScoreLabels: const <int, String>{1: '0'},
        itemRecordValues: const <int, Map<String, dynamic>>{},
      ),
    );
  }

  @override
  Future<Uint8List> downloadRecordBookletPdf(
    String token,
    int id, {
    String dimension = 'score_and_profile',
  }) async {
    return Uint8List(0);
  }

  @override
  Future<Uint8List> downloadRecordReportInterpretationPdf(
    String token,
    int id,
  ) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadAutismDevRecordProfilePdf(
    String token,
    int id, {
    required String profile,
  }) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadAutismDevAssessmentInfoPdf(
    String token,
    int id,
  ) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<AutismDevResultAnalysis> fetchAutismDevResultAnalysis(
    String token,
    int id,
  ) async {
    fetchAutismDevResultAnalysisCalls += 1;
    return savedAutismDevResultAnalysis ?? AutismDevResultAnalysis.empty;
  }

  @override
  Future<AutismDevResultAnalysis> saveAutismDevResultAnalysis(
    String token,
    int id,
    AutismDevResultAnalysis analysis,
  ) async {
    saveAutismDevResultAnalysisCalls += 1;
    savedAutismDevResultAnalysis = analysis;
    return analysis;
  }

  @override
  Future<Uint8List> downloadAutismDevResultAnalysisPdf(
    String token,
    int id,
    AutismDevResultAnalysis analysis,
  ) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadAutismDevSelectedReportPdf(
    String token,
    int id, {
    required List<String> sections,
    AutismDevResultAnalysis? analysis,
  }) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadAutismDevRecordReportInterpretationPdf(
    String token,
    int id,
  ) async {
    return Uint8List.fromList(const <int>[37, 80, 68, 70]);
  }

  @override
  Future<Uint8List> downloadShuangxiDevelopmentProfilePdf(
    String token,
    int id, {
    ShuangxiDevelopmentProfilePdfConfig config =
        const ShuangxiDevelopmentProfilePdfConfig(),
  }) async {
    downloadShuangxiDevelopmentProfilePdfCalls += 1;
    return Uint8List(0);
  }

  @override
  Future<ShuangxiResultAnalysis> fetchShuangxiResultAnalysis(
    String token,
    int id,
  ) async {
    fetchShuangxiResultAnalysisCalls += 1;
    return savedShuangxiResultAnalysis ?? ShuangxiResultAnalysis.empty;
  }

  @override
  Future<ShuangxiResultAnalysis> saveShuangxiResultAnalysis(
    String token,
    int id,
    ShuangxiResultAnalysis analysis,
  ) async {
    saveShuangxiResultAnalysisCalls += 1;
    savedShuangxiResultAnalysis = analysis;
    return analysis;
  }

  @override
  Stream<ShuangxiResultAnalysisStreamEvent>
      generateShuangxiResultAnalysisStream(String token, int id) async* {
    const ShuangxiResultAnalysis analysis = ShuangxiResultAnalysis(
      title: '双溪心智障碍个别化教育课程（三）评量结果分析表',
      generatedBy: 'ai',
      rows: <ShuangxiResultAnalysisRow>[
        ShuangxiResultAnalysisRow(
          domainCode: 'SELF_CARE',
          domain: '生活自理',
          strengths: '能配合部分生活自理流程。',
          weaknesses: '连续步骤仍需提示。',
          reason: '可能与生活练习机会和提示撤除不足有关。',
          strategy: '采用视觉流程卡和分步骤提示进行练习。',
        ),
      ],
    );
    savedShuangxiResultAnalysis = analysis;
    yield const ShuangxiResultAnalysisStreamEvent(
      type: 'status',
      message: '正在读取双溪课程评量结果',
    );
    yield const ShuangxiResultAnalysisStreamEvent(
      type: 'done',
      data: analysis,
    );
  }

  @override
  Stream<AutismDevResultAnalysisStreamEvent>
      generateAutismDevResultAnalysisStream(String token, int id) async* {
    const AutismDevResultAnalysis analysis = AutismDevResultAnalysis(
      title: '孤独症儿童评估结果分析表',
      generatedBy: 'ai',
      rows: <AutismDevResultAnalysisRow>[
        AutismDevResultAnalysisRow(
          domain: '感知觉',
          status: '感知觉现状描述。',
          strengths: '可配合熟悉刺激。',
          weaknesses: '复杂辨别稳定性不足。',
          targets: '1 能追视移动物体。',
        ),
      ],
    );
    savedAutismDevResultAnalysis = analysis;
    yield const AutismDevResultAnalysisStreamEvent(
      type: 'status',
      message: '正在读取孤独症儿童发展评估结果',
    );
    yield const AutismDevResultAnalysisStreamEvent(
      type: 'done',
      data: analysis,
    );
  }

  @override
  Future<ErxinReportInterpretation> fetchAutismDevRecordReportInterpretation(
    String token,
    int id,
  ) async {
    return savedAutismDevReportInterpretation;
  }

  @override
  Future<ErxinReportInterpretation> generateAutismDevRecordReportInterpretation(
    String token,
    int id,
  ) async {
    savedAutismDevReportInterpretation = const ErxinReportInterpretation(
      title: '孤独症儿童发展评估报告解读',
      generatedBy: 'ai',
      summary: '孤独症儿童发展评估结果显示当前表现可作为训练计划参考。',
      domainAnalysis: <String>['感知觉领域需要结合具体题目继续观察。'],
      suggestions: <String>['建议优先设置可观察的小目标。'],
      notes: <String>['本解读仅供参考。'],
    );
    return savedAutismDevReportInterpretation;
  }

  @override
  Stream<ErxinReportInterpretationStreamEvent>
      generateAutismDevRecordReportInterpretationStream(
    String token,
    int id,
  ) async* {
    final ErxinReportInterpretation interpretation =
        await generateAutismDevRecordReportInterpretation(token, id);
    yield const ErxinReportInterpretationStreamEvent(
      type: 'status',
      message: '正在读取孤独症儿童发展评估结果',
    );
    yield ErxinReportInterpretationStreamEvent(
      type: 'done',
      data: interpretation,
    );
  }

  @override
  Future<ErxinReportInterpretation> fetchRecordReportInterpretation(
    String token,
    int id,
  ) async {
    return ErxinReportInterpretation.empty;
  }

  @override
  Future<ErxinReportInterpretation> generateRecordReportInterpretation(
    String token,
    int id,
  ) async {
    return const ErxinReportInterpretation(
      title: 'PEP-3报告解读',
      generatedBy: 'ai',
      summary: 'PEP-3测评结果显示当前整体发展表现可作为教学计划参考。',
      domainAnalysis: <String>['沟通领域表现较稳定。'],
      suggestions: <String>['建议结合日常训练持续观察。'],
      notes: <String>['本解读仅供参考。'],
    );
  }

  @override
  Stream<ErxinReportInterpretationStreamEvent>
      generateRecordReportInterpretationStream(String token, int id) async* {
    yield const ErxinReportInterpretationStreamEvent(
      type: 'status',
      message: '正在读取PEP-3评估结果',
    );
    yield const ErxinReportInterpretationStreamEvent(
      type: 'done',
      data: ErxinReportInterpretation(
        title: 'PEP-3报告解读',
        generatedBy: 'ai',
        summary: 'PEP-3测评结果显示当前整体发展表现可作为教学计划参考。',
        domainAnalysis: <String>['沟通领域表现较稳定。'],
        suggestions: <String>['建议结合日常训练持续观察。'],
        notes: <String>['本解读仅供参考。'],
      ),
    );
  }

  Pep3DraftDetail _draftDetail({required int id}) {
    return Pep3DraftDetail(
      id: id,
      studentId: 3,
      studentName: '张一鸣',
      birthDate: '2021-03-01',
      assessmentDate: '2026-05-05',
      examinerName: '陈老师',
      answeredItemCount: 0,
      completionPercent: 0,
      updatedTime: '2026-05-05T09:00:00Z',
      progress: _progress,
      input: const Pep3DraftInput(
        studentId: 3,
        studentName: '张一鸣',
        examinerName: '陈老师',
        birthDate: '2021-03-01',
        assessmentDate: '2026-05-05',
        remark: '',
        allowMissingItems: true,
        itemScores: <int, int>{},
        itemScoreLabels: <int, String>{},
        itemRecordValues: <int, Map<String, dynamic>>{},
      ),
    );
  }
}

String _fakeNormalize(String value) {
  return value.trim().toLowerCase().replaceAll(RegExp(r'[\s\-/_.]'), '');
}

int _fakeIntFrom(Object? value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  return int.tryParse('$value') ?? 0;
}

class _FakeTimetableClient implements TimetableClient {
  _FakeTimetableClient({
    this.availabilityValid = true,
    this.timetableDelay = Duration.zero,
    this.timetableErrorMessage,
  });

  final bool availabilityValid;
  final Duration timetableDelay;
  final String? timetableErrorMessage;
  int validateCalls = 0;
  int createCalls = 0;
  int updateCalls = 0;
  int oneToOneTargetCalls = 0;
  int groupClassTargetCalls = 0;
  int studentFilterOptionCalls = 0;
  int courseFilterOptionCalls = 0;
  int assistantOptionCalls = 0;
  int classroomOptionCalls = 0;
  int detailCalls = 0;
  int deleteCalls = 0;
  String lastValidatedClassroomId = '';
  String lastCreatedClassroomId = '';
  String lastUpdatedScheduleId = '';
  String lastUpdatedClassroomId = '';
  ScheduleDeleteScope? lastDeleteScope;
  final Set<String> _deletedScheduleIds = <String>{};
  String? _scheduleADate;
  String? _scheduleAStartTime;
  String? _scheduleAEndTime;

  @override
  Future<TimetableData> fetchTimetable(
    String token, {
    required String startDate,
    required String endDate,
    String teacherId = '',
    String periodGroupId = '',
  }) async {
    if (timetableDelay > Duration.zero) {
      await Future<void>.delayed(timetableDelay);
    }
    if (timetableErrorMessage != null) {
      throw TimetableApiException(timetableErrorMessage!);
    }
    final String selectedGroupId =
        periodGroupId.trim().isEmpty ? 'group-a' : periodGroupId.trim();
    final Set<String> groupTeacherIds =
        selectedGroupId == 'group-c' ? <String>{'2'} : <String>{'1', '3'};
    final String requestedTeacherId = teacherId.trim();
    final String selectedTeacherId =
        groupTeacherIds.contains(requestedTeacherId)
            ? requestedTeacherId
            : groupTeacherIds.first;
    final String selectedTeacherName = switch (selectedTeacherId) {
      '2' => '周子涵老师',
      '3' => '黄雨萱老师',
      _ => '陈思语老师',
    };
    final List<TimetableItem> items = <TimetableItem>[
      if (!_deletedScheduleIds.contains('schedule-a'))
        TimetableItem(
          id: 'schedule-a',
          classType: 2,
          teachingClassId: 'one-to-one-a',
          date: _scheduleADate ?? startDate,
          startTime: _scheduleAStartTime ??
              (selectedGroupId == 'group-c' ? '08:30' : '09:15'),
          endTime: _scheduleAEndTime ??
              (selectedGroupId == 'group-c' ? '09:10' : '09:55'),
          lessonName: '感统训练',
          personName: '陈小雨',
          classroomName: 'A101',
          teacherId: selectedTeacherId,
          teacherName: selectedTeacherName,
          assistantIds: const <String>['3'],
          classroomId: '101',
          status: 'unsigned',
          statusText: '未点名',
        ),
      if (!_deletedScheduleIds.contains('schedule-b'))
        TimetableItem(
          id: 'schedule-b',
          classType: 1,
          teachingClassId: 'group-class-a',
          date: _offsetDate(startDate, 2),
          startTime: selectedGroupId == 'group-c' ? '09:20' : '10:05',
          endTime: selectedGroupId == 'group-c' ? '10:00' : '10:45',
          lessonName: '语言认知课',
          personName: '星星班',
          classroomName: 'B203',
          teacherId: selectedTeacherId,
          teacherName: selectedTeacherName,
          assistantIds: const <String>[],
          classroomId: '203',
          status: 'signed',
          statusText: '已点名',
        ),
    ];
    return TimetableData(
      startDate: startDate,
      endDate: endDate,
      selectedPeriodGroupId: selectedGroupId,
      selectedTeacherId: selectedTeacherId,
      selectedTeacherName: selectedTeacherName,
      periodGroups: const <TimetablePeriodGroup>[
        TimetablePeriodGroup(
          id: 'group-a',
          name: 'A组',
          sort: 1,
          startTime: '09:15',
          endTime: '19:40',
          lessonCount: 10,
          teacherIds: <String>['1', '3'],
        ),
        TimetablePeriodGroup(
          id: 'group-c',
          name: 'C组',
          sort: 2,
          startTime: '08:30',
          endTime: '18:50',
          lessonCount: 10,
          teacherIds: <String>['2'],
        ),
      ],
      teachers: selectedGroupId == 'group-c'
          ? const <TimetableTeacher>[
              TimetableTeacher(id: '2', name: '周子涵老师'),
            ]
          : const <TimetableTeacher>[
              TimetableTeacher(id: '1', name: '陈思语老师', current: true),
              TimetableTeacher(id: '3', name: '黄雨萱老师'),
            ],
      days: _fakeTimetableDays(startDate),
      slots: _fakeTimetableSlotsForGroup(selectedGroupId),
      items: items,
      summary: TimetableSummary(
        total: items.length,
        unsigned: items
            .where((TimetableItem item) => item.status == 'unsigned')
            .length,
        signed:
            items.where((TimetableItem item) => item.status == 'signed').length,
      ),
    );
  }

  @override
  Future<List<ScheduleTargetOption>> fetchOneToOneTargets(
    String token, {
    String keyword = '',
  }) async {
    oneToOneTargetCalls += 1;
    return const <ScheduleTargetOption>[
      ScheduleTargetOption(
        id: 'one-to-one-a',
        title: '张一鸣',
        subtitle: '个训课',
        lessonName: '个训课',
        studentName: '张一鸣',
      ),
    ];
  }

  @override
  Future<List<ScheduleTargetOption>> fetchGroupClassTargets(
    String token, {
    String keyword = '',
  }) async {
    groupClassTargetCalls += 1;
    return const <ScheduleTargetOption>[
      ScheduleTargetOption(
        id: 'group-class-a',
        title: '星星班',
        subtitle: '语言认知课 · 5人',
        lessonName: '语言认知课',
      ),
    ];
  }

  @override
  Future<List<ScheduleLookupOption>> fetchScheduleStudentOptions(
    String token, {
    String keyword = '',
  }) async {
    studentFilterOptionCalls += 1;
    return const <ScheduleLookupOption>[
      ScheduleLookupOption(id: '王安全', label: '王安全'),
      ScheduleLookupOption(id: '张一鸣', label: '张一鸣'),
      ScheduleLookupOption(id: '孙吾空', label: '孙吾空'),
    ];
  }

  @override
  Future<List<ScheduleLookupOption>> fetchScheduleCourseOptions(
    String token, {
    String keyword = '',
  }) async {
    courseFilterOptionCalls += 1;
    return const <ScheduleLookupOption>[
      ScheduleLookupOption(id: '个训课', label: '个训课'),
      ScheduleLookupOption(id: '感统训练', label: '感统训练'),
      ScheduleLookupOption(id: '社交沟通课', label: '社交沟通课'),
    ];
  }

  @override
  Future<List<ScheduleStaffOption>> fetchScheduleAssistants(
    String token, {
    String keyword = '',
  }) async {
    assistantOptionCalls += 1;
    return const <ScheduleStaffOption>[
      ScheduleStaffOption(id: '4', name: '助教A', subtitle: '康复老师'),
      ScheduleStaffOption(id: '5', name: '助教B', subtitle: '康复老师'),
    ];
  }

  @override
  Future<List<ScheduleStaffOption>> fetchInstitutionStaffOptions(
    String token, {
    String keyword = '',
  }) async {
    assistantOptionCalls += 1;
    return const <ScheduleStaffOption>[
      ScheduleStaffOption(id: '1', name: '陈老师', subtitle: '评估老师'),
      ScheduleStaffOption(id: '2', name: '李老师', subtitle: '康复老师'),
      ScheduleStaffOption(id: '3', name: '王老师', subtitle: '感觉统合老师'),
    ];
  }

  @override
  Future<List<ScheduleClassroomOption>> fetchScheduleClassrooms(
    String token, {
    String keyword = '',
  }) async {
    classroomOptionCalls += 1;
    return const <ScheduleClassroomOption>[
      ScheduleClassroomOption(id: '101', name: 'A101', subtitle: '一楼'),
      ScheduleClassroomOption(id: '203', name: 'B203', subtitle: '二楼'),
    ];
  }

  @override
  Future<ScheduleValidationResult> validateScheduleSlots(
    String token, {
    required ScheduleTargetType type,
    required String targetId,
    required String teacherId,
    required List<String> assistantIds,
    required String classroomId,
    required List<ScheduleSlotRequest> slots,
    List<String> excludeIds = const <String>[],
  }) async {
    validateCalls += 1;
    lastValidatedClassroomId = classroomId;
    return ScheduleValidationResult(
      valid: availabilityValid,
      message: availabilityValid ? '' : '老师冲突',
      items: slots.map((ScheduleSlotRequest slot) {
        return ScheduleValidationItem(
          teacherId: slot.teacherId,
          lessonDate: slot.lessonDate,
          startTime: slot.startTime,
          endTime: slot.endTime,
          valid: availabilityValid,
          message: availabilityValid ? '空闲时段可排' : '老师冲突',
          conflictTypes:
              availabilityValid ? const <String>[] : const <String>['老师'],
        );
      }).toList(),
    );
  }

  @override
  Future<int> createSchedule(
    String token, {
    required ScheduleTargetType type,
    required String targetId,
    required String teacherId,
    required List<String> assistantIds,
    required String classroomId,
    required ScheduleSlotRequest slot,
  }) async {
    createCalls += 1;
    lastCreatedClassroomId = classroomId;
    return 1;
  }

  @override
  Future<void> updateScheduleSlot(
    String token, {
    required String scheduleId,
    required String teacherId,
    required List<String>? assistantIds,
    required String? classroomId,
    required ScheduleSlotRequest slot,
  }) async {
    updateCalls += 1;
    lastUpdatedScheduleId = scheduleId;
    lastUpdatedClassroomId = classroomId ?? '';
    if (scheduleId == 'schedule-a') {
      _scheduleADate = slot.lessonDate;
      _scheduleAStartTime = slot.startTime;
      _scheduleAEndTime = slot.endTime;
    }
  }

  @override
  Future<TimetableScheduleDetail> fetchScheduleDetail(
    String token, {
    required String scheduleId,
  }) async {
    detailCalls += 1;
    if (scheduleId == 'schedule-b') {
      return const TimetableScheduleDetail(
        id: 'schedule-b',
        batchNo: 'batch-group',
        batchSize: 4,
        classType: 1,
        teachingClassId: 'group-class-a',
        teachingClassName: '星星班',
        lessonId: 'lesson-b',
        lessonName: '语言认知课',
        teacherId: '1',
        teacherName: '陈思语老师',
        assistantNames: <String>['黄雨萱老师'],
        classroomId: '203',
        classroomName: 'B203',
        lessonDate: '2026-05-07',
        startAt: '2026-05-07 10:05:00',
        endAt: '2026-05-07 10:45:00',
        durationMinutes: 40,
        callStatus: 2,
        callStatusText: '已点名',
        students: <TimetableScheduleDetailStudent>[
          TimetableScheduleDetailStudent(
            studentId: 's-2',
            studentName: '李小北',
            maskedPhone: '138****8888',
            scheduleStudentTypeText: '正式',
            classStatusText: '在读',
            callStatus: 2,
            callStatusText: '已点名',
          ),
        ],
      );
    }
    return const TimetableScheduleDetail(
      id: 'schedule-a',
      batchNo: 'batch-a',
      batchSize: 3,
      classType: 2,
      teachingClassId: 'one-to-one-a',
      teachingClassName: '陈小雨-感统训练',
      lessonId: 'lesson-a',
      lessonName: '感统训练',
      teacherId: '1',
      teacherName: '陈思语老师',
      assistantIds: <String>['3'],
      assistantNames: <String>['黄雨萱老师'],
      classroomId: '101',
      classroomName: 'A101',
      lessonDate: '2026-05-05',
      startAt: '2026-05-05 09:15:00',
      endAt: '2026-05-05 09:55:00',
      durationMinutes: 40,
      callStatus: 1,
      callStatusText: '未点名',
      remark: '课前先做前庭唤醒',
      batchMeta: TimetableScheduleBatchMeta(
        schedulingMode: 'repeat',
        repeatRule: 'weekly',
        selectedWeekdays: <String>['周一'],
        plannedClassCount: 3,
      ),
      students: <TimetableScheduleDetailStudent>[
        TimetableScheduleDetailStudent(
          studentId: 's-1',
          studentName: '陈小雨',
          maskedPhone: '136****0001',
          phoneRelationshipText: '妈妈',
          scheduleStudentTypeText: '正式',
          classStatusText: '在读',
          callStatus: 1,
          callStatusText: '未点名',
        ),
      ],
      leaveStudents: <TimetableScheduleDetailStudent>[
        TimetableScheduleDetailStudent(
          studentId: 's-3',
          studentName: '周小米',
          maskedPhone: '136****0002',
          phoneRelationshipText: '爸爸',
          scheduleStudentTypeText: '请假',
          classStatusText: '请假',
          callStatus: 1,
          callStatusText: '未点名',
        ),
      ],
    );
  }

  @override
  Future<int> cancelScheduleScoped(
    String token, {
    required String scheduleId,
    required ScheduleDeleteScope scope,
  }) async {
    deleteCalls += 1;
    lastDeleteScope = scope;
    _deletedScheduleIds.add(scheduleId);
    return 1;
  }
}

List<TimetableDay> _fakeTimetableDays(String startDate) {
  final DateTime start = DateTime.parse(startDate);
  return List<TimetableDay>.generate(7, (int index) {
    final DateTime date = start.add(Duration(days: index));
    return TimetableDay(
      date: _fakeDateText(date),
      label: _fakeWeekdayShort(date.weekday),
      weekday: _fakeWeekdayFull(date.weekday),
    );
  });
}

String _offsetDate(String startDate, int offset) {
  return _fakeDateText(DateTime.parse(startDate).add(Duration(days: offset)));
}

String _fakeDateText(DateTime date) {
  return '${date.year.toString().padLeft(4, '0')}-'
      '${date.month.toString().padLeft(2, '0')}-'
      '${date.day.toString().padLeft(2, '0')}';
}

String _fakeMinuteText(DateTime date) {
  return '${date.year.toString().padLeft(4, '0')}-'
      '${date.month.toString().padLeft(2, '0')}-'
      '${date.day.toString().padLeft(2, '0')} '
      '${date.hour.toString().padLeft(2, '0')}:'
      '${date.minute.toString().padLeft(2, '0')}';
}

String _fakeWeekdayShort(int weekday) {
  const List<String> labels = <String>[
    '周一',
    '周二',
    '周三',
    '周四',
    '周五',
    '周六',
    '周日',
  ];
  return labels[(weekday - 1).clamp(0, 6)];
}

String _fakeWeekdayFull(int weekday) {
  const List<String> labels = <String>[
    '星期一',
    '星期二',
    '星期三',
    '星期四',
    '星期五',
    '星期六',
    '星期日',
  ];
  return labels[(weekday - 1).clamp(0, 6)];
}

List<TimetableSlot> _fakeTimetableSlotsForGroup(String periodGroupId) {
  if (periodGroupId == 'group-c') {
    return _fakeTimetableSlots830;
  }
  return _fakeTimetableSlots915;
}

const List<TimetableSlot> _fakeTimetableSlots915 = <TimetableSlot>[
  TimetableSlot(
    title: '第一节',
    time: '09:15 - 09:55',
    startTime: '09:15',
    endTime: '09:55',
  ),
  TimetableSlot(
    title: '第二节',
    time: '10:05 - 10:45',
    startTime: '10:05',
    endTime: '10:45',
  ),
  TimetableSlot(
    title: '第三节',
    time: '10:55 - 11:35',
    startTime: '10:55',
    endTime: '11:35',
  ),
  TimetableSlot(
    title: '第四节',
    time: '14:00 - 14:40',
    startTime: '14:00',
    endTime: '14:40',
  ),
  TimetableSlot(
    title: '第五节',
    time: '14:50 - 15:30',
    startTime: '14:50',
    endTime: '15:30',
  ),
  TimetableSlot(
    title: '第六节',
    time: '15:40 - 16:20',
    startTime: '15:40',
    endTime: '16:20',
  ),
  TimetableSlot(
    title: '第七节',
    time: '16:30 - 17:10',
    startTime: '16:30',
    endTime: '17:10',
  ),
  TimetableSlot(
    title: '第八节',
    time: '17:20 - 18:00',
    startTime: '17:20',
    endTime: '18:00',
  ),
  TimetableSlot(
    title: '第九节',
    time: '18:10 - 18:50',
    startTime: '18:10',
    endTime: '18:50',
  ),
  TimetableSlot(
    title: '第十节',
    time: '19:00 - 19:40',
    startTime: '19:00',
    endTime: '19:40',
  ),
];

const List<TimetableSlot> _fakeTimetableSlots830 = <TimetableSlot>[
  TimetableSlot(
    title: '第一节',
    time: '08:30 - 09:10',
    startTime: '08:30',
    endTime: '09:10',
  ),
  TimetableSlot(
    title: '第二节',
    time: '09:20 - 10:00',
    startTime: '09:20',
    endTime: '10:00',
  ),
  TimetableSlot(
    title: '第三节',
    time: '10:10 - 10:50',
    startTime: '10:10',
    endTime: '10:50',
  ),
  TimetableSlot(
    title: '第四节',
    time: '11:00 - 11:40',
    startTime: '11:00',
    endTime: '11:40',
  ),
  TimetableSlot(
    title: '第五节',
    time: '14:00 - 14:40',
    startTime: '14:00',
    endTime: '14:40',
  ),
  TimetableSlot(
    title: '第六节',
    time: '14:50 - 15:30',
    startTime: '14:50',
    endTime: '15:30',
  ),
  TimetableSlot(
    title: '第七节',
    time: '15:40 - 16:20',
    startTime: '15:40',
    endTime: '16:20',
  ),
  TimetableSlot(
    title: '第八节',
    time: '16:30 - 17:10',
    startTime: '16:30',
    endTime: '17:10',
  ),
  TimetableSlot(
    title: '第九节',
    time: '17:20 - 18:00',
    startTime: '17:20',
    endTime: '18:00',
  ),
  TimetableSlot(
    title: '第十节',
    time: '18:10 - 18:50',
    startTime: '18:10',
    endTime: '18:50',
  ),
];

class _MultiInstitutionAuthClient extends _FakeAuthClient {
  @override
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier, {
    String password = '',
  }) async {
    return const <InstitutionLoginOption>[
      InstitutionLoginOption(
        userId: 1,
        instId: 11,
        orgName: '启明成长中心',
        loginName: 'chenrui',
        nickName: '陈老师',
        mobile: '19900000001',
        admin: true,
        institutionReadonly: false,
        institutionStatus: 'normal',
      ),
      InstitutionLoginOption(
        userId: 2,
        instId: 12,
        orgName: '南山训练中心',
        loginName: 'chenrui',
        nickName: '陈老师',
        mobile: '19900000002',
        admin: false,
        institutionReadonly: false,
        institutionStatus: 'warning',
      ),
    ];
  }
}

class _PasswordCheckingAuthClient extends _MultiInstitutionAuthClient {
  @override
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier, {
    String password = '',
  }) async {
    if (password != '123456') {
      throw const AuthException('登录失败,用户名或密码错误');
    }
    return super.listInstitutionOptions(identifier, password: password);
  }
}
