import 'package:assessment_pad_app/auth_client.dart';
import 'package:assessment_pad_app/assessment_scale_client.dart';
import 'package:assessment_pad_app/assessment_scale_category_page.dart';
import 'package:assessment_pad_app/home_client.dart';
import 'package:assessment_pad_app/main.dart';
import 'package:assessment_pad_app/pep3_assessment_client.dart';
import 'package:assessment_pad_app/pep3_assessment_page.dart';
import 'package:assessment_pad_app/smart_timetable_page.dart';
import 'package:assessment_pad_app/timetable_client.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('login page opens the home dashboard after real login callback',
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

    expect(find.text('评估助手'), findsOneWidget);
    expect(find.text('机构账号登录'), findsOneWidget);
    expect(find.text('验证码登录'), findsNothing);

    await _enterWithCustomKeyboard(tester, 0, 'chenrui');
    await _enterWithCustomKeyboard(tester, 1, '123456');
    await tester.tap(find.text('登 录'));
    await tester.pumpAndSettle();

    expect(find.textContaining('启明成长中心'), findsOneWidget);
    expect(find.text('开始测评'), findsOneWidget);
    expect(find.byIcon(Icons.search_rounded), findsOneWidget);

    await tester.binding.handlePopRoute();
    await tester.pumpAndSettle();

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
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.textContaining('启明成长中心'), findsOneWidget);
    expect(find.text('机构账号登录'), findsNothing);
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
    expect(find.text('C组'), findsOneWidget);

    await tester.tap(find.text('新增排课'));
    await tester.pumpAndSettle();
    expect(find.text('选择 1v1 后立即检测本周空闲点'), findsNothing);
    expect(find.text('选择班课后立即检测本周空闲点'), findsNothing);

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

    await tester.tap(find.text('C组'));
    await tester.pumpAndSettle();

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

    final TestGesture gesture = await tester.startGesture(source);
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pumpAndSettle();
    expect(find.text('可调课'), findsWidgets);
    await gesture.moveTo(target);
    await tester.pump();
    await gesture.moveTo(target);
    await tester.pump();
    await gesture.up();
    await tester.pumpAndSettle();

    expect(timetableClient.updateCalls, 1);
    expect(timetableClient.lastUpdatedScheduleId, 'schedule-a');
    expect(timetableClient.lastUpdatedClassroomId, '101');
    expect(find.byKey(const ValueKey<String>('lesson-0-1')), findsOneWidget);
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

    expect(timetableClient.createCalls, 1);
    expect(timetableClient.lastCreatedClassroomId, '101');
    expect(find.byType(SnackBar), findsNothing);
    expect(
      find.byKey(const ValueKey<String>('schedule-top-message')),
      findsOneWidget,
    );
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
    expect(find.text('确认选择'), findsOneWidget);
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

    await tester.tap(find.text('开始测评').at(1));
    await tester.pumpAndSettle();
    await tester.tap(find.text('王安全'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
    await tester.pumpAndSettle();

    expect(find.text('王安全 * 未知'), findsOneWidget);
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
        timetableClient: _FakeTimetableClient(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('新建测评'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('开始测评').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('张一鸣'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('确认选择'));
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

    expect(find.text('PEP-3 测评工作台'), findsOneWidget);
    expect(find.text('记录册页面'), findsOneWidget);
    expect(find.textContaining('旋开瓶盖'), findsWidgets);
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

    await tester.tap(find.text('保存草稿'));
    await tester.pump(const Duration(milliseconds: 120));
    expect(pep3Client.saveDraftCalls, 2);
    expect(find.text('草稿已保存'), findsWidgets);
    expect(
      find.byWidgetPredicate(
        (Widget widget) => widget.runtimeType.toString() == 'PadTopMessage',
      ),
      findsOneWidget,
    );
    expect(find.byType(SnackBar), findsNothing);
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
  @override
  Future<HomeSession> fetchCurrentSession(String token) async {
    return const HomeSession(
      nickName: '陈老师',
      orgName: '启明成长中心',
    );
  }

  @override
  Future<HomeSummary> fetchSummary(String token) async {
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

class _FakeAssessmentScaleClient implements AssessmentScaleClient {
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

  @override
  Future<List<String>> fetchCategories(String token) async {
    return const <String>['语言与沟通能力', '社交情绪评估'];
  }

  @override
  Future<AssessmentScaleLibrary> fetchScaleLibrary(
    String token, {
    String keyword = '',
    String category = '',
  }) async {
    final String normalizedKeyword = _fakeNormalize(keyword);
    final List<AssessmentScaleItem> filtered = _items.where(
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
      summary: const AssessmentScaleLibrarySummary(
        total: 2,
        available: 2,
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
    return const AssessmentDraftPage(
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
          createdTime: '2026-05-04T09:00:00Z',
          updatedTime: '2026-05-04T10:00:00Z',
        ),
      ],
    );
  }

  @override
  Future<AssessmentStudentCandidatePage> fetchStudentCandidates(
    String token, {
    String scaleCode = '',
    String keyword = '',
    int pageIndex = 1,
    int pageSize = 20,
  }) async {
    return const AssessmentStudentCandidatePage(
      total: 2,
      current: 1,
      size: 20,
      items: <AssessmentStudentCandidate>[
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
      ],
    );
  }
}

class _FakePep3AssessmentClient implements Pep3AssessmentClient {
  _FakePep3AssessmentClient({
    this.hasDraft = false,
    this.hasPreviousRecord = false,
  });

  final bool hasDraft;
  final bool hasPreviousRecord;
  int saveDraftCalls = 0;
  int inviteCalls = 0;

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
    return Pep3AssessmentItem(
      itemNo: itemNo,
      itemTitle: itemNo == 1 ? '（1） 旋开瓶盖' : '（2） 叠积木',
      testItem: itemNo == 1 ? '旋开瓶盖' : '叠积木',
      domainCode: 'FM',
      domainName: '小肌肉',
      materials: itemNo == 1 ? '肥皂泡液' : '积木',
      materialImages: const <String>[],
      method: '观察儿童是否可以按标准完成任务。',
      guidance: '请你试试看。',
      guidanceVideo: '',
      standard: '',
      scoreOptions: _scoreOptions,
      recordFields: const <Pep3RecordField>[],
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
      return const Pep3DraftPage(
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
            updatedTime: '2026-05-05T14:01:00',
            progress: Pep3DraftProgress(
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
    return _draftDetail(id: 11);
  }

  @override
  Future<Pep3CaregiverInvite> inviteCaregiverReport(
    String token,
    int draftId,
  ) async {
    inviteCalls += 1;
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
    return const Pep3RecordPage(
      items: <Pep3RecordSummary>[
        Pep3RecordSummary(
          id: 21,
          studentId: 3,
          studentName: '张一鸣',
          assessmentCode: 'PEP3',
          assessmentName: 'PEP-3',
          birthDate: '2021-03-01',
          assessmentDate: '2026-05-04',
          examinerName: '陈老师',
          updatedTime: '2026-05-04T16:00:00',
        ),
      ],
      total: 1,
      current: 1,
      size: 1,
    );
  }

  @override
  Future<Pep3RecordDetail> fetchRecordDetail(String token, int id) async {
    return const Pep3RecordDetail(
      id: 21,
      studentId: 3,
      studentName: '张一鸣',
      assessmentCode: 'PEP3',
      assessmentName: 'PEP-3',
      birthDate: '2021-03-01',
      assessmentDate: '2026-05-04',
      examinerName: '陈老师',
      updatedTime: '2026-05-04T16:00:00',
      input: Pep3DraftInput(
        studentId: 3,
        studentName: '张一鸣',
        examinerName: '陈老师',
        birthDate: '2021-03-01',
        assessmentDate: '2026-05-04',
        remark: '',
        allowMissingItems: true,
        itemScores: <int, int>{1: 0},
        itemRecordValues: <int, Map<String, dynamic>>{},
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
        itemRecordValues: <int, Map<String, dynamic>>{},
      ),
    );
  }
}

String _fakeNormalize(String value) {
  return value.trim().toLowerCase().replaceAll(RegExp(r'[\s\-/_.]'), '');
}

class _FakeTimetableClient implements TimetableClient {
  _FakeTimetableClient({this.availabilityValid = true});

  final bool availabilityValid;
  int validateCalls = 0;
  int createCalls = 0;
  int updateCalls = 0;
  String lastValidatedClassroomId = '';
  String lastCreatedClassroomId = '';
  String lastUpdatedScheduleId = '';
  String lastUpdatedClassroomId = '';
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
      items: <TimetableItem>[
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
      ],
      summary: const TimetableSummary(
        total: 2,
        unsigned: 1,
        signed: 1,
      ),
    );
  }

  @override
  Future<List<ScheduleTargetOption>> fetchOneToOneTargets(
    String token, {
    String keyword = '',
  }) async {
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
  Future<List<ScheduleStaffOption>> fetchScheduleAssistants(
    String token, {
    String keyword = '',
  }) async {
    return const <ScheduleStaffOption>[
      ScheduleStaffOption(id: '4', name: '助教A', subtitle: '康复老师'),
      ScheduleStaffOption(id: '5', name: '助教B', subtitle: '康复老师'),
    ];
  }

  @override
  Future<List<ScheduleClassroomOption>> fetchScheduleClassrooms(
    String token, {
    String keyword = '',
  }) async {
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
