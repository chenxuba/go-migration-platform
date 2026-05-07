import 'dart:convert';
import 'dart:io';

import 'package:assessment_pad_app/timetable_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('fetchInstitutionStaffOptions parses items payload', () async {
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
      expect(request.uri.path, '/api/v1/inst-users/page');
      expect(decoded['queryModel'], isA<Map>());
      expect((decoded['queryModel'] as Map)['status'], false);
      expect((decoded['pageRequestModel'] as Map)['pageSize'], 500);

      request.response
        ..statusCode = HttpStatus.ok
        ..headers.contentType = ContentType.json
        ..write(
          jsonEncode(
            <String, dynamic>{
              'success': true,
              'data': <String, dynamic>{
                'items': <Map<String, dynamic>>[
                  <String, dynamic>{
                    'id': 1,
                    'nickName': '陈老师',
                    'roleName': '评估老师',
                  },
                  <String, dynamic>{
                    'id': 2,
                    'nickName': '李老师',
                    'roleName': '康复老师',
                  },
                ],
                'total': 2,
                'current': 1,
                'size': 500,
              },
            },
          ),
        );
      await request.response.close();
    });

    final ApiTimetableClient client = ApiTimetableClient(
      educationBaseUrl: 'http://127.0.0.1:${server.port}',
    );
    final List<ScheduleStaffOption> options =
        await client.fetchInstitutionStaffOptions('token');

    expect(options, hasLength(2));
    expect(options.first.name, '陈老师');
    expect(options.last.name, '李老师');
  });
}
