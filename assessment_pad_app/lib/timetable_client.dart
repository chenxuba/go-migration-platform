import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'home_client.dart';

const String defaultPadTimetablePath = String.fromEnvironment(
  'PAD_TIMETABLE_PATH',
  defaultValue: '/api/v1/pad/timetable',
);
const String defaultOneToOneSelectionPath = String.fromEnvironment(
  'ONE_TO_ONE_SELECTION_PATH',
  defaultValue: '/api/v1/one-to-ones/selection-page',
);
const String defaultOneToOnePagePath = String.fromEnvironment(
  'ONE_TO_ONE_PAGE_PATH',
  defaultValue: '/api/v1/one-to-ones/page',
);
const String defaultGroupClassSelectionPath = String.fromEnvironment(
  'GROUP_CLASS_SELECTION_PATH',
  defaultValue: '/api/v1/group-classes/selection-page',
);
const String defaultCourseOptionsPath = String.fromEnvironment(
  'COURSE_OPTIONS_PATH',
  defaultValue: '/api/v1/courses/options',
);
const String defaultScheduleAssistantPath = String.fromEnvironment(
  'SCHEDULE_ASSISTANT_PATH',
  defaultValue: '/api/v1/inst-users/page',
);
const String defaultScheduleClassroomPath = String.fromEnvironment(
  'SCHEDULE_CLASSROOM_PATH',
  defaultValue: '/api/v1/classrooms',
);
const String defaultOneToOneValidatePath = String.fromEnvironment(
  'ONE_TO_ONE_VALIDATE_PATH',
  defaultValue: '/api/v1/teaching-schedules/one-to-one/validate',
);
const String defaultGroupClassValidatePath = String.fromEnvironment(
  'GROUP_CLASS_VALIDATE_PATH',
  defaultValue: '/api/v1/teaching-schedules/group-class/validate',
);
const String defaultOneToOneCreatePath = String.fromEnvironment(
  'ONE_TO_ONE_CREATE_PATH',
  defaultValue: '/api/v1/teaching-schedules/one-to-one/create',
);
const String defaultGroupClassCreatePath = String.fromEnvironment(
  'GROUP_CLASS_CREATE_PATH',
  defaultValue: '/api/v1/teaching-schedules/group-class/create',
);
const String defaultScheduleDetailPath = String.fromEnvironment(
  'SCHEDULE_DETAIL_PATH',
  defaultValue: '/api/v1/teaching-schedules/detail',
);
const String defaultScheduleBatchUpdatePath = String.fromEnvironment(
  'SCHEDULE_BATCH_UPDATE_PATH',
  defaultValue: '/api/v1/teaching-schedules/batch-update',
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
    this.classType = 0,
    this.teachingClassId = '',
    this.teachingClassName = '',
    this.studentName = '',
    this.teacherId = '',
    this.teacherName = '',
    this.batchNo = '',
    this.assistantIds,
    this.classroomId,
    this.conflict = false,
  });

  factory TimetableItem.fromJson(Map<String, dynamic> json) {
    return TimetableItem(
      id: '${json['id'] ?? ''}',
      batchNo: '${json['batchNo'] ?? ''}',
      classType: _intFrom(json['classType']),
      teachingClassId: '${json['teachingClassId'] ?? ''}',
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
      assistantIds: json.containsKey('assistantIds')
          ? _stringListFrom(json['assistantIds'])
          : null,
      classroomId: json.containsKey('classroomId')
          ? '${json['classroomId'] ?? ''}'
          : null,
      status: '${json['status'] ?? ''}',
      statusText: '${json['statusText'] ?? ''}',
      conflict: json['conflict'] == true,
    );
  }

  final String id;
  final String batchNo;
  final int classType;
  final String teachingClassId;
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
  final List<String>? assistantIds;
  final String? classroomId;
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

enum ScheduleTargetType {
  oneToOne,
  groupClass,
}

class ScheduleTargetOption {
  const ScheduleTargetOption({
    required this.id,
    required this.title,
    this.subtitle = '',
    this.lessonName = '',
    this.studentName = '',
    this.disabled = false,
  });

  factory ScheduleTargetOption.oneToOneFromJson(Map<String, dynamic> json) {
    final String studentName = '${json['studentName'] ?? ''}'.trim();
    final String lessonName = '${json['lessonName'] ?? ''}'.trim();
    final String name = '${json['name'] ?? ''}'.trim();
    final bool statusDisabled =
        json.containsKey('status') && _intFrom(json['status']) != 1;
    final bool studentStatusDisabled = json.containsKey('classStudentStatus') &&
        _intFrom(json['classStudentStatus']) != 1;
    return ScheduleTargetOption(
      id: '${json['id'] ?? ''}',
      title: studentName.isNotEmpty
          ? studentName
          : (name.isNotEmpty ? name : '未命名1v1'),
      subtitle: lessonName.isEmpty ? name : lessonName,
      lessonName: lessonName,
      studentName: studentName,
      disabled: statusDisabled || studentStatusDisabled,
    );
  }

  factory ScheduleTargetOption.groupClassFromJson(Map<String, dynamic> json) {
    final String name = '${json['name'] ?? ''}'.trim();
    final String lessonName = '${json['lessonName'] ?? ''}'.trim();
    final int studentCount = _intFrom(json['studentCount']);
    return ScheduleTargetOption(
      id: '${json['id'] ?? ''}',
      title: name.isEmpty ? '未命名班课' : name,
      subtitle: lessonName.isEmpty
          ? '$studentCount人'
          : '$lessonName · $studentCount人',
      lessonName: lessonName,
      disabled: json.containsKey('status') && _intFrom(json['status']) != 1,
    );
  }

  final String id;
  final String title;
  final String subtitle;
  final String lessonName;
  final String studentName;
  final bool disabled;
}

class ScheduleLookupOption {
  const ScheduleLookupOption({
    required this.id,
    required this.label,
  });

  final String id;
  final String label;
}

class ScheduleStaffOption {
  const ScheduleStaffOption({
    required this.id,
    required this.name,
    this.subtitle = '',
  });

  factory ScheduleStaffOption.fromJson(Map<String, dynamic> json) {
    final String roleName = '${json['roleName'] ?? ''}'.trim();
    final String departNames = '${json['departNames'] ?? ''}'.trim();
    return ScheduleStaffOption(
      id: '${json['id'] ?? ''}',
      name: '${json['nickName'] ?? ''}'.trim().isEmpty
          ? '未命名员工'
          : '${json['nickName'] ?? ''}'.trim(),
      subtitle: roleName.isNotEmpty ? roleName : departNames,
    );
  }

  final String id;
  final String name;
  final String subtitle;
}

class ScheduleClassroomOption {
  const ScheduleClassroomOption({
    required this.id,
    required this.name,
    this.subtitle = '',
  });

  factory ScheduleClassroomOption.fromJson(Map<String, dynamic> json) {
    return ScheduleClassroomOption(
      id: '${json['id'] ?? json['uuid'] ?? ''}',
      name: '${json['name'] ?? ''}'.trim().isEmpty
          ? '未命名教室'
          : '${json['name'] ?? ''}'.trim(),
      subtitle: '${json['address'] ?? ''}'.trim(),
    );
  }

  final String id;
  final String name;
  final String subtitle;
}

class ScheduleSlotRequest {
  const ScheduleSlotRequest({
    required this.teacherId,
    required this.lessonDate,
    required this.startTime,
    required this.endTime,
    this.assistantIds = const <String>[],
    this.classroomId = '',
  });

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> json = <String, dynamic>{
      'teacherId': teacherId,
      'lessonDate': lessonDate,
      'startTime': startTime,
      'endTime': endTime,
    };
    if (assistantIds.isNotEmpty) {
      json['assistantIds'] = assistantIds;
    }
    if (classroomId.trim().isNotEmpty) {
      json['classroomId'] = classroomId.trim();
    }
    return json;
  }

  final String teacherId;
  final String lessonDate;
  final String startTime;
  final String endTime;
  final List<String> assistantIds;
  final String classroomId;
}

class ScheduleValidationResult {
  const ScheduleValidationResult({
    required this.valid,
    this.message = '',
    this.items = const <ScheduleValidationItem>[],
  });

  factory ScheduleValidationResult.fromJson(Map<String, dynamic> json) {
    return ScheduleValidationResult(
      valid: json['valid'] != false,
      message: '${json['message'] ?? ''}',
      items: _listFrom(json['items']).map((Map<String, dynamic> item) {
        return ScheduleValidationItem.fromJson(item);
      }).toList(),
    );
  }

  final bool valid;
  final String message;
  final List<ScheduleValidationItem> items;
}

class ScheduleValidationItem {
  const ScheduleValidationItem({
    required this.teacherId,
    required this.lessonDate,
    required this.startTime,
    required this.endTime,
    required this.valid,
    this.message = '',
    this.conflictTypes = const <String>[],
  });

  factory ScheduleValidationItem.fromJson(Map<String, dynamic> json) {
    return ScheduleValidationItem(
      teacherId: '${json['teacherId'] ?? ''}',
      lessonDate: '${json['lessonDate'] ?? ''}',
      startTime: '${json['startTime'] ?? ''}',
      endTime: '${json['endTime'] ?? ''}',
      valid: json['valid'] != false,
      message: '${json['message'] ?? ''}',
      conflictTypes: _rawListFrom(json['conflictTypes'])
          .map((Object? item) => '${item ?? ''}'.trim())
          .where((String item) => item.isNotEmpty)
          .toList(),
    );
  }

  final String teacherId;
  final String lessonDate;
  final String startTime;
  final String endTime;
  final bool valid;
  final String message;
  final List<String> conflictTypes;
}

abstract interface class TimetableClient {
  Future<TimetableData> fetchTimetable(
    String token, {
    required String startDate,
    required String endDate,
    String teacherId = '',
    String periodGroupId = '',
  });

  Future<List<ScheduleTargetOption>> fetchOneToOneTargets(
    String token, {
    String keyword = '',
  });

  Future<List<ScheduleTargetOption>> fetchGroupClassTargets(
    String token, {
    String keyword = '',
  });

  Future<List<ScheduleLookupOption>> fetchScheduleStudentOptions(
    String token, {
    String keyword = '',
  });

  Future<List<ScheduleLookupOption>> fetchScheduleCourseOptions(
    String token, {
    String keyword = '',
  });

  Future<List<ScheduleStaffOption>> fetchScheduleAssistants(
    String token, {
    String keyword = '',
  });

  Future<List<ScheduleStaffOption>> fetchInstitutionStaffOptions(
    String token, {
    String keyword = '',
  });

  Future<List<ScheduleClassroomOption>> fetchScheduleClassrooms(
    String token, {
    String keyword = '',
  });

  Future<ScheduleValidationResult> validateScheduleSlots(
    String token, {
    required ScheduleTargetType type,
    required String targetId,
    required String teacherId,
    required List<String> assistantIds,
    required String classroomId,
    required List<ScheduleSlotRequest> slots,
    List<String> excludeIds = const <String>[],
  });

  Future<int> createSchedule(
    String token, {
    required ScheduleTargetType type,
    required String targetId,
    required String teacherId,
    required List<String> assistantIds,
    required String classroomId,
    required ScheduleSlotRequest slot,
  });

  Future<void> updateScheduleSlot(
    String token, {
    required String scheduleId,
    required String teacherId,
    required List<String>? assistantIds,
    required String? classroomId,
    required ScheduleSlotRequest slot,
  });
}

class ApiTimetableClient implements TimetableClient {
  const ApiTimetableClient({
    this.educationBaseUrl = defaultEducationApiBaseUrl,
    this.timetablePath = defaultPadTimetablePath,
    this.oneToOneSelectionPath = defaultOneToOneSelectionPath,
    this.oneToOnePagePath = defaultOneToOnePagePath,
    this.groupClassSelectionPath = defaultGroupClassSelectionPath,
    this.courseOptionsPath = defaultCourseOptionsPath,
    this.scheduleAssistantPath = defaultScheduleAssistantPath,
    this.scheduleClassroomPath = defaultScheduleClassroomPath,
    this.oneToOneValidatePath = defaultOneToOneValidatePath,
    this.groupClassValidatePath = defaultGroupClassValidatePath,
    this.oneToOneCreatePath = defaultOneToOneCreatePath,
    this.groupClassCreatePath = defaultGroupClassCreatePath,
    this.scheduleDetailPath = defaultScheduleDetailPath,
    this.scheduleBatchUpdatePath = defaultScheduleBatchUpdatePath,
  });

  final String educationBaseUrl;
  final String timetablePath;
  final String oneToOneSelectionPath;
  final String oneToOnePagePath;
  final String groupClassSelectionPath;
  final String courseOptionsPath;
  final String scheduleAssistantPath;
  final String scheduleClassroomPath;
  final String oneToOneValidatePath;
  final String groupClassValidatePath;
  final String oneToOneCreatePath;
  final String groupClassCreatePath;
  final String scheduleDetailPath;
  final String scheduleBatchUpdatePath;

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

  @override
  Future<List<ScheduleTargetOption>> fetchOneToOneTargets(
    String token, {
    String keyword = '',
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, oneToOneSelectionPath),
      token,
      <String, dynamic>{
        'queryModel': <String, dynamic>{
          'searchKey': keyword.trim(),
          'status': <int>[1],
        },
        'pageRequestModel': <String, dynamic>{
          'needTotal': false,
          'pageSize': 500,
          'pageIndex': 1,
          'skipCount': 0,
        },
      },
    );
    return _listPayload(data)
        .map((Map<String, dynamic> item) {
          return ScheduleTargetOption.oneToOneFromJson(item);
        })
        .where((ScheduleTargetOption item) => item.id.trim().isNotEmpty)
        .toList();
  }

  @override
  Future<List<ScheduleTargetOption>> fetchGroupClassTargets(
    String token, {
    String keyword = '',
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, groupClassSelectionPath),
      token,
      <String, dynamic>{
        'queryModel': <String, dynamic>{
          'className': keyword.trim(),
          'status': <int>[1],
        },
        'pageRequestModel': <String, dynamic>{
          'needTotal': false,
          'pageSize': 500,
          'pageIndex': 1,
          'skipCount': 0,
        },
      },
    );
    return _listPayload(data)
        .map((Map<String, dynamic> item) {
          return ScheduleTargetOption.groupClassFromJson(item);
        })
        .where((ScheduleTargetOption item) => item.id.trim().isNotEmpty)
        .toList();
  }

  @override
  Future<List<ScheduleLookupOption>> fetchScheduleStudentOptions(
    String token, {
    String keyword = '',
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, oneToOnePagePath),
      token,
      <String, dynamic>{
        'queryModel': <String, dynamic>{
          'status': <int>[1],
          if (keyword.trim().isNotEmpty) 'searchKey': keyword.trim(),
        },
        'pageRequestModel': <String, dynamic>{
          'needTotal': false,
          'pageSize': 500,
          'pageIndex': 1,
          'skipCount': 0,
        },
      },
    );
    final Map<String, ScheduleLookupOption> optionById =
        <String, ScheduleLookupOption>{};
    for (final Map<String, dynamic> item in _listPayload(data)) {
      final String label =
          '${item['studentName'] ?? item['name'] ?? ''}'.trim();
      if (label.isEmpty) {
        continue;
      }
      optionById[label] = ScheduleLookupOption(id: label, label: label);
    }
    return optionById.values.toList();
  }

  @override
  Future<List<ScheduleLookupOption>> fetchScheduleCourseOptions(
    String token, {
    String keyword = '',
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, courseOptionsPath),
      token,
      <String, dynamic>{
        'searchKey': keyword.trim(),
      },
    );
    final Map<String, ScheduleLookupOption> optionById =
        <String, ScheduleLookupOption>{};
    for (final Map<String, dynamic> item in _listPayload(data)) {
      final String label = '${item['name'] ?? ''}'.trim();
      if (label.isEmpty) {
        continue;
      }
      optionById[label] = ScheduleLookupOption(id: label, label: label);
    }
    return optionById.values.toList();
  }

  @override
  Future<List<ScheduleStaffOption>> fetchScheduleAssistants(
    String token, {
    String keyword = '',
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, scheduleAssistantPath),
      token,
      <String, dynamic>{
        'queryModel': <String, dynamic>{
          'searchKey': keyword.trim(),
          'isTeacher': true,
          'status': true,
        },
        'pageRequestModel': <String, dynamic>{
          'needTotal': false,
          'pageSize': 40,
          'pageIndex': 1,
        },
      },
    );
    return _listPayload(data)
        .map((Map<String, dynamic> item) => ScheduleStaffOption.fromJson(item))
        .where((ScheduleStaffOption item) => item.id.trim().isNotEmpty)
        .toList();
  }

  @override
  Future<List<ScheduleStaffOption>> fetchInstitutionStaffOptions(
    String token, {
    String keyword = '',
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, scheduleAssistantPath),
      token,
      <String, dynamic>{
        'queryModel': <String, dynamic>{
          'searchKey': keyword.trim(),
          'status': false,
        },
        'pageRequestModel': <String, dynamic>{
          'needTotal': false,
          'pageSize': 500,
          'pageIndex': 1,
          'skipCount': 0,
        },
      },
    );
    return _listPayload(data)
        .map((Map<String, dynamic> item) => ScheduleStaffOption.fromJson(item))
        .where((ScheduleStaffOption item) => item.id.trim().isNotEmpty)
        .toList();
  }

  @override
  Future<List<ScheduleClassroomOption>> fetchScheduleClassrooms(
    String token, {
    String keyword = '',
  }) async {
    final Object? data = await _getJson(
      _uri(educationBaseUrl, scheduleClassroomPath).replace(
        queryParameters: <String, String>{
          'enabledOnly': 'true',
          if (keyword.trim().isNotEmpty) 'searchKey': keyword.trim(),
        },
      ),
      token,
    );
    return _listPayload(data)
        .map(
          (Map<String, dynamic> item) => ScheduleClassroomOption.fromJson(item),
        )
        .where((ScheduleClassroomOption item) => item.id.trim().isNotEmpty)
        .toList();
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
    final Object? data = await _postJson(
      _uri(
        educationBaseUrl,
        type == ScheduleTargetType.oneToOne
            ? oneToOneValidatePath
            : groupClassValidatePath,
      ),
      token,
      _schedulePayload(
        type: type,
        targetId: targetId,
        teacherId: teacherId,
        assistantIds: assistantIds,
        classroomId: classroomId,
        slots: slots,
        excludeIds: excludeIds,
      ),
    );
    if (data is! Map) {
      throw const TimetableApiException('空闲点检测接口返回格式不正确');
    }
    return ScheduleValidationResult.fromJson(Map<String, dynamic>.from(data));
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
    final Object? data = await _postJson(
      _uri(
        educationBaseUrl,
        type == ScheduleTargetType.oneToOne
            ? oneToOneCreatePath
            : groupClassCreatePath,
      ),
      token,
      _schedulePayload(
        type: type,
        targetId: targetId,
        teacherId: teacherId,
        assistantIds: assistantIds,
        classroomId: classroomId,
        slots: <ScheduleSlotRequest>[slot],
      ),
    );
    if (data is Map) {
      return _intFrom(data['count']);
    }
    return 0;
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
    List<String>? resolvedAssistantIds = assistantIds;
    String? resolvedClassroomId = classroomId;
    if (resolvedAssistantIds == null || resolvedClassroomId == null) {
      final _ScheduleUpdateContext context =
          await _fetchScheduleUpdateContext(token, scheduleId);
      resolvedAssistantIds ??= context.assistantIds;
      resolvedClassroomId ??= context.classroomId;
    }
    final List<String> finalAssistantIds = resolvedAssistantIds;
    final String finalClassroomId = resolvedClassroomId;

    final Map<String, dynamic> payload = <String, dynamic>{
      'ids': <String>[scheduleId.trim()],
      'teacherId': teacherId.trim(),
      'assistantIds': finalAssistantIds,
      'lessonDate': slot.lessonDate,
      'startTime': slot.startTime,
      'endTime': slot.endTime,
      'allowStudentConflict': false,
    };
    if (finalClassroomId.trim().isNotEmpty) {
      payload['classroomId'] = finalClassroomId.trim();
    }
    await _postJson(
        _uri(educationBaseUrl, scheduleBatchUpdatePath), token, payload);
  }

  Future<_ScheduleUpdateContext> _fetchScheduleUpdateContext(
    String token,
    String scheduleId,
  ) async {
    final Object? data = await _getJson(
      _uri(educationBaseUrl, scheduleDetailPath).replace(
        queryParameters: <String, String>{'id': scheduleId.trim()},
      ),
      token,
    );
    if (data is! Map) {
      return const _ScheduleUpdateContext();
    }
    return _ScheduleUpdateContext.fromJson(Map<String, dynamic>.from(data));
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

  Future<Object?> _postJson(
    Uri uri,
    String token,
    Map<String, dynamic> body,
  ) async {
    final http.Response response;
    try {
      response = await http
          .post(
            uri,
            headers: <String, String>{
              'Accept': 'application/json',
              'Content-Type': 'application/json',
              'Authorization': 'Bearer $token',
            },
            body: jsonEncode(body),
          )
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const TimetableApiException('排课接口响应超时，请检查网络');
    } on Object catch (error) {
      throw TimetableApiException('无法连接排课接口：$error');
    }

    final Object? decoded = _decodeResponse(response.body);
    if (response.statusCode == 401) {
      throw TimetableApiException(
        _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw TimetableApiException(_messageFromPayload(decoded) ?? '排课接口请求失败');
    }
    if (decoded is Map) {
      final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
      if (envelope['success'] == false) {
        throw TimetableApiException(
          _messageFromPayload(envelope) ?? '排课接口请求失败',
        );
      }
      if (envelope.containsKey('code') && _intFrom(envelope['code']) != 0) {
        final int code = _intFrom(envelope['code']);
        if (code != 200) {
          throw TimetableApiException(
            _messageFromPayload(envelope) ?? '排课接口请求失败',
          );
        }
      }
      if (envelope.containsKey('data')) {
        return envelope['data'];
      }
    }
    return decoded;
  }
}

class _ScheduleUpdateContext {
  const _ScheduleUpdateContext({
    this.assistantIds = const <String>[],
    this.classroomId = '',
  });

  factory _ScheduleUpdateContext.fromJson(Map<String, dynamic> json) {
    return _ScheduleUpdateContext(
      assistantIds: _stringListFrom(json['assistantIds']),
      classroomId: '${json['classroomId'] ?? ''}',
    );
  }

  final List<String> assistantIds;
  final String classroomId;
}

Map<String, dynamic> _schedulePayload({
  required ScheduleTargetType type,
  required String targetId,
  required String teacherId,
  required List<String> assistantIds,
  required String classroomId,
  required List<ScheduleSlotRequest> slots,
  List<String> excludeIds = const <String>[],
}) {
  final String targetKey =
      type == ScheduleTargetType.oneToOne ? 'oneToOneId' : 'groupClassId';
  final Map<String, dynamic> payload = <String, dynamic>{
    targetKey: targetId.trim(),
    'teacherId': teacherId.trim(),
    'assistantIds': assistantIds,
    'schedules':
        slots.map((ScheduleSlotRequest slot) => slot.toJson()).toList(),
  };
  if (classroomId.trim().isNotEmpty) {
    payload['classroomId'] = classroomId.trim();
  }
  if (excludeIds.isNotEmpty) {
    payload['excludeIds'] = excludeIds
        .map((String item) => item.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
  }
  return payload;
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

List<Map<String, dynamic>> _listPayload(Object? raw) {
  if (raw is List) {
    return _listFrom(raw);
  }
  if (raw is Map) {
    final Map<String, dynamic> payload = Map<String, dynamic>.from(raw);
    List<Map<String, dynamic>>? emptyResult;
    for (final String key in <String>['list', 'items', 'records', 'rows']) {
      final Object? value = payload[key];
      if (value is List) {
        final List<Map<String, dynamic>> items = _listFrom(value);
        if (items.isNotEmpty) {
          return items;
        }
        emptyResult ??= items;
      }
    }
    return emptyResult ?? <Map<String, dynamic>>[];
  }
  return <Map<String, dynamic>>[];
}

List<Object?> _rawListFrom(Object? raw) {
  if (raw is List) {
    return raw;
  }
  return const <Object?>[];
}

List<String> _stringListFrom(Object? raw) {
  if (raw is List) {
    return raw
        .map((Object? item) => '${item ?? ''}'.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
  }
  if (raw is String) {
    return raw
        .split(',')
        .map((String item) => item.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
  }
  return const <String>[];
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
