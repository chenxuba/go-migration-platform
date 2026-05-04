import 'package:assessment_pad_app/auth_client.dart';
import 'package:assessment_pad_app/home_client.dart';
import 'package:assessment_pad_app/main.dart';
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
        timetableClient: _FakeTimetableClient(),
      ),
    );

    expect(find.text('测评云端'), findsOneWidget);
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
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(
      AssessmentPadApp(
        authClient: _FakeAuthClient(),
        homeClient: _FakeHomeClient(),
        timetableClient: _FakeTimetableClient(),
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
    final TestGesture gesture = await tester.startGesture(source);
    await tester.pump();
    await gesture.moveTo(target);
    await tester.pump();
    await gesture.up();
    await tester.pumpAndSettle();

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
    await tester.tap(
      find.byKey(const ValueKey<String>('schedule-target-one-to-one-a')),
    );
    await tester.pumpAndSettle();

    expect(timetableClient.validateCalls, greaterThan(0));
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

class _FakeTimetableClient implements TimetableClient {
  _FakeTimetableClient({this.availabilityValid = true});

  final bool availabilityValid;
  int validateCalls = 0;
  int createCalls = 0;

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
          date: startDate,
          startTime: selectedGroupId == 'group-c' ? '08:30' : '09:15',
          endTime: selectedGroupId == 'group-c' ? '09:10' : '09:55',
          lessonName: '感统训练',
          personName: '陈小雨',
          classroomName: 'A101',
          teacherId: selectedTeacherId,
          teacherName: selectedTeacherName,
          status: 'unsigned',
          statusText: '未点名',
        ),
        TimetableItem(
          id: 'schedule-b',
          date: _offsetDate(startDate, 2),
          startTime: selectedGroupId == 'group-c' ? '09:20' : '10:05',
          endTime: selectedGroupId == 'group-c' ? '10:00' : '10:45',
          lessonName: '语言认知课',
          personName: '星星班',
          classroomName: 'B203',
          teacherId: selectedTeacherId,
          teacherName: selectedTeacherName,
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
  }) async {
    validateCalls += 1;
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
