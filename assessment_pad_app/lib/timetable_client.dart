import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'home_client.dart';

const String defaultPadTimetablePath = String.fromEnvironment(
  'PAD_TIMETABLE_PATH',
  defaultValue: '/api/v1/pad/timetable',
);

class TimetableApiException implements Exception {
  const TimetableApiException(this.message, {this.unauthorized = false});

  final String message;
  final bool unauthorized;

  @override
  String toString() => message;
}

class TimetableData {
  const TimetableData({
    required this.startDate,
    required this.endDate,
    required this.selectedPeriodGroupId,
    required this.selectedTeacherId,
    required this.selectedTeacherName,
    required this.periodGroups,
    required this.teachers,
    required this.days,
    required this.slots,
    required this.items,
    required this.summary,
  });

  factory TimetableData.fromJson(Map<String, dynamic> json) {
    return TimetableData(
      startDate: '${json['startDate'] ?? ''}',
      endDate: '${json['endDate'] ?? ''}',
      selectedPeriodGroupId: '${json['selectedPeriodGroupUuid'] ?? ''}',
      selectedTeacherId: '${json['selectedTeacherId'] ?? ''}',
      selectedTeacherName: '${json['selectedTeacherName'] ?? ''}',
      periodGroups: _listFrom(json['periodGroups'])
          .map(
            (Map<String, dynamic> item) => TimetablePeriodGroup.fromJson(item),
          )
          .toList(),
      teachers: _listFrom(json['teachers'])
          .map((Map<String, dynamic> item) => TimetableTeacher.fromJson(item))
          .toList(),
      days: _listFrom(json['days'])
          .map((Map<String, dynamic> item) => TimetableDay.fromJson(item))
          .toList(),
      slots: _listFrom(json['slots'])
          .map((Map<String, dynamic> item) => TimetableSlot.fromJson(item))
          .toList(),
      items: _listFrom(json['items'])
          .map((Map<String, dynamic> item) => TimetableItem.fromJson(item))
          .toList(),
      summary: TimetableSummary.fromJson(_mapFrom(json['summary'])),
    );
  }

  factory TimetableData.fallback() {
    final DateTime monday = _weekMonday(DateTime.now());
    final DateTime sunday = monday.add(const Duration(days: 6));
    return TimetableData(
      startDate: _formatDate(monday),
      endDate: _formatDate(sunday),
      selectedPeriodGroupId: 'default',
      selectedTeacherId: '',
      selectedTeacherName: '当前老师',
      periodGroups: const <TimetablePeriodGroup>[
        TimetablePeriodGroup(
          id: 'default',
          name: '默认时段',
          startTime: '08:00',
          endTime: '18:20',
          lessonCount: 11,
        ),
      ],
      teachers: const <TimetableTeacher>[],
      days: List<TimetableDay>.generate(7, (int index) {
        final DateTime day = monday.add(Duration(days: index));
        return TimetableDay(
          date: _formatDate(day),
          label: _weekdayShort(day.weekday),
          weekday: _weekdayFull(day.weekday),
        );
      }),
      slots: const <TimetableSlot>[
        TimetableSlot(
          title: '第一节',
          time: '08:00 - 08:40',
          startTime: '08:00',
          endTime: '08:40',
        ),
        TimetableSlot(
          title: '第二节',
          time: '08:50 - 09:30',
          startTime: '08:50',
          endTime: '09:30',
        ),
        TimetableSlot(
          title: '第三节',
          time: '09:40 - 10:20',
          startTime: '09:40',
          endTime: '10:20',
        ),
        TimetableSlot(
          title: '第四节',
          time: '10:30 - 11:10',
          startTime: '10:30',
          endTime: '11:10',
        ),
        TimetableSlot(
          title: '第五节',
          time: '11:20 - 12:00',
          startTime: '11:20',
          endTime: '12:00',
        ),
        TimetableSlot(
          title: '第六节',
          time: '13:30 - 14:10',
          startTime: '13:30',
          endTime: '14:10',
        ),
        TimetableSlot(
          title: '第七节',
          time: '14:20 - 15:00',
          startTime: '14:20',
          endTime: '15:00',
        ),
        TimetableSlot(
          title: '第八节',
          time: '15:10 - 15:50',
          startTime: '15:10',
          endTime: '15:50',
        ),
        TimetableSlot(
          title: '第九节',
          time: '16:00 - 16:40',
          startTime: '16:00',
          endTime: '16:40',
        ),
        TimetableSlot(
          title: '第十节',
          time: '16:50 - 17:30',
          startTime: '16:50',
          endTime: '17:30',
        ),
        TimetableSlot(
          title: '第十一节',
          time: '17:40 - 18:20',
          startTime: '17:40',
          endTime: '18:20',
        ),
      ],
      items: const <TimetableItem>[],
      summary: const TimetableSummary(),
    );
  }

  final String startDate;
  final String endDate;
  final String selectedPeriodGroupId;
  final String selectedTeacherId;
  final String selectedTeacherName;
  final List<TimetablePeriodGroup> periodGroups;
  final List<TimetableTeacher> teachers;
  final List<TimetableDay> days;
  final List<TimetableSlot> slots;
  final List<TimetableItem> items;
  final TimetableSummary summary;
}

class TimetablePeriodGroup {
  const TimetablePeriodGroup({
    required this.id,
    required this.name,
    this.sort = 0,
    this.startTime = '',
    this.endTime = '',
    this.lessonCount = 0,
    this.teacherIds = const <String>[],
  });

  factory TimetablePeriodGroup.fromJson(Map<String, dynamic> json) {
    return TimetablePeriodGroup(
      id: '${json['id'] ?? ''}',
      name: '${json['name'] ?? ''}',
      sort: _intFrom(json['sort']),
      startTime: '${json['startTime'] ?? ''}',
      endTime: '${json['endTime'] ?? ''}',
      lessonCount: _intFrom(json['lessonCount']),
      teacherIds: _rawListFrom(json['teacherIds'])
          .map((Object? item) => '${item ?? ''}')
          .where((String item) => item.trim().isNotEmpty)
          .toList(),
    );
  }

  final String id;
  final String name;
  final int sort;
  final String startTime;
  final String endTime;
  final int lessonCount;
  final List<String> teacherIds;
}

class TimetableTeacher {
  const TimetableTeacher({
    required this.id,
    required this.name,
    this.current = false,
  });

  factory TimetableTeacher.fromJson(Map<String, dynamic> json) {
    return TimetableTeacher(
      id: '${json['id'] ?? ''}',
      name: '${json['name'] ?? ''}',
      current: json['current'] == true,
    );
  }

  final String id;
  final String name;
  final bool current;
}

class TimetableDay {
  const TimetableDay({
    required this.date,
    required this.label,
    required this.weekday,
  });

  factory TimetableDay.fromJson(Map<String, dynamic> json) {
    return TimetableDay(
      date: '${json['date'] ?? ''}',
      label: '${json['label'] ?? ''}',
      weekday: '${json['weekday'] ?? ''}',
    );
  }

  final String date;
  final String label;
  final String weekday;
}

class TimetableSlot {
  const TimetableSlot({
    required this.title,
    required this.time,
    required this.startTime,
    required this.endTime,
  });

  factory TimetableSlot.fromJson(Map<String, dynamic> json) {
    return TimetableSlot(
      title: '${json['title'] ?? ''}',
      time: '${json['time'] ?? ''}',
      startTime: '${json['startTime'] ?? ''}',
      endTime: '${json['endTime'] ?? ''}',
    );
  }

  final String title;
  final String time;
  final String startTime;
  final String endTime;
}

class TimetableItem {
  const TimetableItem({
    required this.id,
    required this.date,
    required this.startTime,
    required this.endTime,
    required this.lessonName,
    required this.personName,
    required this.classroomName,
    required this.status,
    required this.statusText,
    this.teachingClassName = '',
    this.studentName = '',
    this.teacherId = '',
    this.teacherName = '',
    this.conflict = false,
  });

  factory TimetableItem.fromJson(Map<String, dynamic> json) {
    return TimetableItem(
      id: '${json['id'] ?? ''}',
      date: '${json['date'] ?? ''}',
      startTime: '${json['startTime'] ?? ''}',
      endTime: '${json['endTime'] ?? ''}',
      lessonName: '${json['lessonName'] ?? ''}',
      teachingClassName: '${json['teachingClassName'] ?? ''}',
      studentName: '${json['studentName'] ?? ''}',
      personName: '${json['personName'] ?? ''}',
      classroomName: '${json['classroomName'] ?? ''}',
      teacherId: '${json['teacherId'] ?? ''}',
      teacherName: '${json['teacherName'] ?? ''}',
      status: '${json['status'] ?? ''}',
      statusText: '${json['statusText'] ?? ''}',
      conflict: json['conflict'] == true,
    );
  }

  final String id;
  final String date;
  final String startTime;
  final String endTime;
  final String lessonName;
  final String teachingClassName;
  final String studentName;
  final String personName;
  final String classroomName;
  final String teacherId;
  final String teacherName;
  final String status;
  final String statusText;
  final bool conflict;
}

class TimetableSummary {
  const TimetableSummary({
    this.total = 0,
    this.unsigned = 0,
    this.signed = 0,
    this.partial = 0,
    this.trial = 0,
    this.conflict = 0,
  });

  factory TimetableSummary.fromJson(Map<String, dynamic> json) {
    return TimetableSummary(
      total: _intFrom(json['total']),
      unsigned: _intFrom(json['unsigned']),
      signed: _intFrom(json['signed']),
      partial: _intFrom(json['partial']),
      trial: _intFrom(json['trial']),
      conflict: _intFrom(json['conflict']),
    );
  }

  final int total;
  final int unsigned;
  final int signed;
  final int partial;
  final int trial;
  final int conflict;
}

abstract interface class TimetableClient {
  Future<TimetableData> fetchTimetable(
    String token, {
    required String startDate,
    required String endDate,
    String teacherId = '',
    String periodGroupId = '',
  });
}

class ApiTimetableClient implements TimetableClient {
  const ApiTimetableClient({
    this.educationBaseUrl = defaultEducationApiBaseUrl,
    this.timetablePath = defaultPadTimetablePath,
  });

  final String educationBaseUrl;
  final String timetablePath;

  @override
  Future<TimetableData> fetchTimetable(
    String token, {
    required String startDate,
    required String endDate,
    String teacherId = '',
    String periodGroupId = '',
  }) async {
    final Map<String, String> query = <String, String>{
      'startDate': startDate,
      'endDate': endDate,
    };
    if (teacherId.trim().isNotEmpty) {
      query['teacherId'] = teacherId.trim();
    }
    if (periodGroupId.trim().isNotEmpty) {
      query['periodGroupUuid'] = periodGroupId.trim();
    }
    final Object? data = await _getJson(
      _uri(educationBaseUrl, timetablePath).replace(queryParameters: query),
      token,
    );
    if (data is! Map) {
      throw const TimetableApiException('排课日程接口返回格式不正确');
    }
    return TimetableData.fromJson(Map<String, dynamic>.from(data));
  }

  Future<Object?> _getJson(Uri uri, String token) async {
    final http.Response response;
    try {
      response = await http.get(
        uri,
        headers: <String, String>{
          'Accept': 'application/json',
          'Authorization': 'Bearer $token',
        },
      ).timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const TimetableApiException('排课日程接口响应超时，请检查网络');
    } on Object catch (error) {
      throw TimetableApiException('无法连接排课日程接口：$error');
    }

    final Object? decoded = _decodeResponse(response.body);
    if (response.statusCode == 401) {
      throw TimetableApiException(
        _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw TimetableApiException(_messageFromPayload(decoded) ?? '排课日程加载失败');
    }
    if (decoded is Map) {
      final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
      if (envelope['success'] == false) {
        throw TimetableApiException(
          _messageFromPayload(envelope) ?? '排课日程加载失败',
        );
      }
      if (envelope.containsKey('data')) {
        return envelope['data'];
      }
    }
    return decoded;
  }
}

Uri _uri(String baseUrl, String path) {
  final String trimmedBase = baseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
  final String normalizedPath = path.startsWith('/') ? path : '/$path';
  return Uri.parse('$trimmedBase$normalizedPath');
}

Object? _decodeResponse(String body) {
  if (body.trim().isEmpty) {
    return null;
  }
  try {
    return jsonDecode(body);
  } on FormatException {
    return body;
  }
}

String? _messageFromPayload(Object? payload) {
  if (payload is Map) {
    final Object? message = payload['message'] ?? payload['msg'];
    if (message != null && '$message'.trim().isNotEmpty) {
      return '$message';
    }
  }
  if (payload is String && payload.trim().isNotEmpty) {
    return payload;
  }
  return null;
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
  return const <Object?>[];
}

Map<String, dynamic> _mapFrom(Object? raw) {
  if (raw is Map) {
    return Map<String, dynamic>.from(raw);
  }
  return <String, dynamic>{};
}

int _intFrom(Object? raw) {
  if (raw is int) {
    return raw;
  }
  if (raw is num) {
    return raw.toInt();
  }
  return int.tryParse('${raw ?? ''}') ?? 0;
}

DateTime _weekMonday(DateTime date) {
  return DateTime(date.year, date.month, date.day)
      .subtract(Duration(days: date.weekday - 1));
}

String _formatDate(DateTime date) {
  return '${date.year.toString().padLeft(4, '0')}-'
      '${date.month.toString().padLeft(2, '0')}-'
      '${date.day.toString().padLeft(2, '0')}';
}

String _weekdayShort(int weekday) {
  const List<String> labels = <String>[
    '周一',
    '周二',
    '周三',
    '周四',
    '周五',
    '周六',
    '周日'
  ];
  return labels[(weekday - 1).clamp(0, 6)];
}

String _weekdayFull(int weekday) {
  const List<String> labels = <String>[
    '星期一',
    '星期二',
    '星期三',
    '星期四',
    '星期五',
    '星期六',
    '星期日'
  ];
  return labels[(weekday - 1).clamp(0, 6)];
}
