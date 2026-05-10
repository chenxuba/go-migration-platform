import 'dart:convert';
import 'dart:io';

import 'package:assessment_pad_app/iep_assessment_record_client.dart';
import 'package:assessment_pad_app/iep_plan_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('fetchRecordsPage merges pep3 and erxin records by assessment date',
      () async {
    final HttpServer server =
        await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    final List<String> requestedPaths = <String>[];
    addTearDown(() async {
      await server.close(force: true);
    });

    server.listen((HttpRequest request) async {
      final String body = await utf8.decoder.bind(request).join();
      final Map<String, dynamic> decoded =
          jsonDecode(body) as Map<String, dynamic>;
      final Map<dynamic, dynamic> pageRequest =
          decoded['pageRequestModel'] as Map<dynamic, dynamic>;
      final Map<dynamic, dynamic> query =
          decoded['queryModel'] as Map<dynamic, dynamic>;
      requestedPaths.add(request.uri.path);

      expect(request.method, 'POST');
      expect(request.headers.value(HttpHeaders.authorizationHeader),
          'Bearer token-1');
      expect(request.headers.value('X-Access-Token'), 'token-1');
      expect(pageRequest['pageIndex'], 1);
      expect(pageRequest['pageSize'], 4);
      expect(query['searchKey'], '陈旭');
      expect(query['assessmentDateBegin'], '2026-05-01');
      expect(query['assessmentDateEnd'], '2026-05-31');

      final List<Map<String, dynamic>> items =
          request.uri.path.contains('/erxin/')
              ? <Map<String, dynamic>>[
                  <String, dynamic>{
                    'id': 91,
                    'studentId': 18,
                    'studentName': '陈旭',
                    'assessmentCode': 'ERXIN2',
                    'assessmentName': '儿心量表-II',
                    'birthDate': '2022-05-11T00:00:00Z',
                    'assessmentDate': '2026-05-07T00:00:00Z',
                    'ageYears': 4,
                    'ageMonths': 0,
                    'examinerName': '陈瑞',
                    'iepPlanStatus': 'confirmed',
                    'updatedTime': '2026-05-07T10:30:00Z',
                  },
                ]
              : <Map<String, dynamic>>[
                  <String, dynamic>{
                    'id': 88,
                    'studentId': 19,
                    'studentName': '林一诺',
                    'assessmentCode': 'PEP3',
                    'assessmentName': 'PEP-3',
                    'birthDate': '2021-08-12',
                    'assessmentDate': '2026-04-29',
                    'ageYears': 4,
                    'ageMonths': 8,
                    'examinerName': '陈瑞',
                    'updatedTime': '2026-04-29T11:30:00Z',
                  },
                ];

      request.response
        ..statusCode = HttpStatus.ok
        ..headers.contentType = ContentType.json
        ..write(
          jsonEncode(
            <String, dynamic>{
              'success': true,
              'data': <String, dynamic>{
                'items': items,
                'total': items.length,
                'current': 1,
                'size': 4,
              },
            },
          ),
        );
      await request.response.close();
    });

    final ApiIepAssessmentRecordClient client = ApiIepAssessmentRecordClient(
      educationBaseUrl: 'http://127.0.0.1:${server.port}',
    );
    final IepAssessmentRecordPage page = await client.fetchRecordsPage(
      'token-1',
      pageIndex: 1,
      pageSize: 4,
      searchKey: ' 陈旭 ',
      assessmentDateBegin: '2026-05-01',
      assessmentDateEnd: '2026-05-31',
    );

    expect(requestedPaths, contains('/api/v1/assessments/pep3/records/page'));
    expect(requestedPaths, contains('/api/v1/assessments/erxin/records/page'));
    expect(page.total, 2);
    expect(page.items, hasLength(2));
    expect(page.items.first.source, 'ERXIN');
    expect(page.items.first.studentName, '陈旭');
    expect(page.items.first.assessmentDate, '2026-05-07');
    expect(page.items.last.source, 'PEP3');
  });

  test('plan client loads erxin iep and execution plans', () async {
    final HttpServer server =
        await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    final List<String> requestedPaths = <String>[];
    addTearDown(() async {
      await server.close(force: true);
    });

    server.listen((HttpRequest request) async {
      requestedPaths.add(request.uri.path);
      expect(request.method, 'GET');
      expect(request.uri.queryParameters['id'], '91');
      expect(request.uri.queryParameters['durationMonths'], '3');

      final Object data = request.uri.path.contains('/execution/')
          ? <String, dynamic>{
              'exists': true,
              'durationMonths': 3,
              'monthlyPlans': <Map<String, dynamic>>[
                <String, dynamic>{
                  'targetMonthIndex': 1,
                  'plan': <String, dynamic>{
                    'title': '康复教学5月计划',
                    'student': <String, dynamic>{'name': '陈旭'},
                    'meta': <String, dynamic>{
                      'planDate': '2026-05-07',
                      'participant': '陈瑞',
                      'implementer': '陈瑞',
                      'startDate': '2026-05-01',
                      'endDate': '2026-05-31',
                    },
                    'rows': <Map<String, dynamic>>[
                      <String, dynamic>{
                        'domain': '大肌肉',
                        'longGoal': '提升动态平衡',
                        'shortGoal': '能独立行走3米',
                        'trainingItems': <Map<String, dynamic>>[
                          <String, dynamic>{
                            'content': '平衡木行走训练',
                            'startEndDate': '2026-05-01 - 2026-05-10',
                          },
                        ],
                        'courseForm': '个训',
                      },
                    ],
                  },
                },
              ],
              'weeklyPlans': <Map<String, dynamic>>[
                <String, dynamic>{
                  'targetMonthIndex': 1,
                  'targetWeekIndex': 1,
                  'plan': <String, dynamic>{
                    'title': '康复教学周计划日记录卡5月第1周',
                    'student': <String, dynamic>{'name': '陈旭'},
                    'teacherName': '陈瑞',
                    'courseName': '康复教学',
                    'trainingDate': '2026-05-01 至 2026-05-02',
                    'preparation': '平衡木',
                    'weekDates': <String>['2026-05-01', '2026-05-02'],
                    'rows': <Map<String, dynamic>>[
                      <String, dynamic>{
                        'project': '平衡木行走',
                        'content': '独立行走3米',
                      },
                    ],
                  },
                },
              ],
            }
          : <String, dynamic>{
              'exists': true,
              'status': 'confirmed',
              'durationMonths': 3,
              'plan': <String, dynamic>{
                'title': '康复教学季度计划',
                'student': <String, dynamic>{
                  'name': '陈旭',
                  'gender': '-',
                  'birthDate': '2022-05-11T00:00:00Z',
                },
                'meta': <String, dynamic>{
                  'planDate': '2026-05-07',
                  'participant': '陈瑞',
                  'implementer': '陈瑞',
                  'startDate': '2026-05-01',
                  'endDate': '2026-07-31',
                },
                'rows': <Map<String, dynamic>>[
                  <String, dynamic>{
                    'domain': '大肌肉',
                    'longGoal': '提升动态平衡',
                    'shortGoal': '能独立行走3米',
                    'courseForm': '个训',
                    'startEndDate': '2026-05-01 - 2026-05-31',
                  },
                ],
              },
            };

      request.response
        ..statusCode = HttpStatus.ok
        ..headers.contentType = ContentType.json
        ..write(jsonEncode(<String, dynamic>{'success': true, 'data': data}));
      await request.response.close();
    });

    final ApiIepPlanClient client = ApiIepPlanClient(
      educationBaseUrl: 'http://127.0.0.1:${server.port}',
    );
    const IepAssessmentRecordSummary record = IepAssessmentRecordSummary(
      id: 91,
      source: 'ERXIN',
      studentId: 18,
      studentName: '陈旭',
      assessmentCode: 'ERXIN2',
      assessmentName: '儿心量表-II',
      birthDate: '2022-05-11',
      assessmentDate: '2026-05-07',
      examinerName: '陈瑞',
      updatedTime: '',
    );

    final IepPlanSaved plan = await client.fetchIepPlan(
      'token-1',
      record: record,
      durationMonths: 3,
    );
    final IepExecutionPlansSaved execution = await client
        .fetchExecutionPlans('token-1', record: record, durationMonths: 3);

    expect(requestedPaths,
        contains('/api/v1/assessments/erxin/records/iep-plan/detail'));
    expect(
        requestedPaths,
        contains(
            '/api/v1/assessments/erxin/records/iep-plan/execution/detail'));
    expect(plan.hasContent, isTrue);
    expect(plan.plan?.student.birthDate, '2022-05-11');
    expect(execution.monthPlan(1)?.rows.first.trainingItems.first.content,
        '平衡木行走训练');
    expect(execution.weekPlan(1, 1)?.rows.first.project, '平衡木行走');
  });

  test('plan client syncs erxin period with backend batch API', () async {
    final HttpServer server =
        await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    addTearDown(() async {
      await server.close(force: true);
    });

    server.listen((HttpRequest request) async {
      final String body = await utf8.decoder.bind(request).join();
      final Map<String, dynamic> decoded =
          jsonDecode(body) as Map<String, dynamic>;

      expect(request.method, 'POST');
      expect(request.uri.path,
          '/api/v1/assessments/erxin/records/iep-plan/period/sync');
      expect(request.headers.value(HttpHeaders.authorizationHeader),
          'Bearer token-1');
      expect(decoded['id'], 91);
      expect(decoded['durationMonths'], 3);
      expect(decoded['sourceDurationMonths'], 6);
      expect(decoded['startDate'], '2026-05-05');

      request.response
        ..statusCode = HttpStatus.ok
        ..headers.contentType = ContentType.json
        ..write(
          jsonEncode(
            <String, dynamic>{
              'success': true,
              'data': <String, dynamic>{
                'iepPlan': <String, dynamic>{
                  'exists': true,
                  'status': 'confirmed',
                  'durationMonths': 3,
                  'plan': <String, dynamic>{
                    'title': '康复教学季度计划',
                    'student': <String, dynamic>{'name': '陈旭'},
                    'meta': <String, dynamic>{
                      'startDate': '2026-05-05',
                      'endDate': '2026-07-31',
                    },
                    'rows': <Map<String, dynamic>>[
                      <String, dynamic>{
                        'domain': '大肌肉',
                        'shortGoal': '能独立行走3米',
                      },
                    ],
                  },
                },
                'executionPlans': <String, dynamic>{
                  'exists': true,
                  'durationMonths': 3,
                  'monthlyPlans': <Map<String, dynamic>>[],
                  'weeklyPlans': <Map<String, dynamic>>[],
                },
              },
            },
          ),
        );
      await request.response.close();
    });

    final ApiIepPlanClient client = ApiIepPlanClient(
      educationBaseUrl: 'http://127.0.0.1:${server.port}',
    );
    const IepAssessmentRecordSummary record = IepAssessmentRecordSummary(
      id: 91,
      source: 'ERXIN',
      studentId: 18,
      studentName: '陈旭',
      assessmentCode: 'ERXIN2',
      assessmentName: '儿心量表-II',
      birthDate: '2022-05-11',
      assessmentDate: '2026-05-07',
      examinerName: '陈瑞',
      updatedTime: '',
    );

    final IepPlanPeriodSyncResult result = await client.syncIepPlanPeriod(
      'token-1',
      record: record,
      durationMonths: 3,
      sourceDurationMonths: 6,
      startDate: DateTime(2026, 5, 5),
    );

    expect(result.iepPlan.plan?.meta.startDate, '2026-05-05');
    expect(result.iepPlan.plan?.meta.endDate, '2026-07-31');
    expect(result.executionPlans.durationMonths, 3);
  });
}
