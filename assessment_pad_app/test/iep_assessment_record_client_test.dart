import 'dart:convert';
import 'dart:io';

import 'package:assessment_pad_app/iep_assessment_record_client.dart';
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
}
