part of '../smart_timetable_page.dart';

String _dateRangeText(int weekOffset) {
  final _WeekRange range = _weekRange(weekOffset);
  final DateTime start = range.start;
  final DateTime end = range.end;
  String two(int value) => value.toString().padLeft(2, '0');
  if (start.year == end.year) {
    return '${start.year}.${two(start.month)}.${two(start.day)} - '
        '${two(end.month)}.${two(end.day)}';
  }
  return '${start.year}.${two(start.month)}.${two(start.day)} - '
      '${end.year}.${two(end.month)}.${two(end.day)}';
}

_WeekRange _weekRange(int weekOffset) {
  final DateTime now = DateTime.now();
  final DateTime today = DateTime(now.year, now.month, now.day);
  final DateTime monday = today.subtract(Duration(days: today.weekday - 1)).add(
        Duration(days: weekOffset * 7),
      );
  final DateTime sunday = monday.add(const Duration(days: 6));
  return _WeekRange(start: monday, end: sunday);
}

String _formatApiDate(DateTime date) {
  return '${date.year.toString().padLeft(4, '0')}-'
      '${date.month.toString().padLeft(2, '0')}-'
      '${date.day.toString().padLeft(2, '0')}';
}

List<TimetableDay> _timetableDaysForRange(_WeekRange range) {
  return List<TimetableDay>.generate(7, (int index) {
    final DateTime day = range.start.add(Duration(days: index));
    return TimetableDay(
      date: _formatApiDate(day),
      label: _weekdayShortByNumber(day.weekday),
      weekday: _weekdayFullByNumber(day.weekday),
    );
  });
}

const String _timetableAuthTokenStorageKey = 'auth_token';

List<_PeriodGroupOption> _periodGroupOptionsFromData(TimetableData data) {
  final List<TimetablePeriodGroup> groups = data.periodGroups;
  if (groups.isEmpty) {
    return const <_PeriodGroupOption>[
      _PeriodGroupOption(
        id: 'default',
        name: '默认时段',
        meta: '08:00 - 18:20 · 11节',
      ),
    ];
  }
  return groups.map((TimetablePeriodGroup group) {
    final String name =
        group.name.trim().isEmpty ? '未命名时段组' : group.name.trim();
    final String range =
        group.startTime.trim().isNotEmpty && group.endTime.trim().isNotEmpty
            ? '${group.startTime} - ${group.endTime}'
            : '未设置时段';
    final String count =
        group.lessonCount > 0 ? '${group.lessonCount}节' : '未启用';
    return _PeriodGroupOption(
      id: group.id.trim().isEmpty ? 'default' : group.id.trim(),
      name: name,
      meta: '$range · $count',
    );
  }).toList();
}

List<_TeacherOption> _teacherOptionsFromData(TimetableData data) {
  final List<TimetableTeacher> teachers = data.teachers;
  if (teachers.isEmpty) {
    final String name = data.selectedTeacherName.trim();
    if (name.isEmpty && data.selectedTeacherId.trim().isEmpty) {
      return const <_TeacherOption>[];
    }
    return <_TeacherOption>[
      _TeacherOption(
        id: data.selectedTeacherId,
        name: name.isEmpty ? '当前老师' : name,
        label: '当前老师',
      ),
    ];
  }
  return teachers.map((TimetableTeacher teacher) {
    final bool selected = teacher.id == data.selectedTeacherId;
    return _TeacherOption(
      id: teacher.id,
      name: teacher.name.trim().isEmpty ? '未命名老师' : teacher.name.trim(),
      label: selected || teacher.current ? '当前老师' : '可切换老师',
    );
  }).toList();
}

List<_WeekDay> _weekDaysFromData(TimetableData data) {
  final List<_WeekDay> days = data.days
      .map(
        (TimetableDay day) => _WeekDay(
          label: day.label.trim().isEmpty
              ? _weekdayShortLabel(day.date)
              : day.label,
          date: _monthDayLabel(day.date),
          isoDate: day.date,
          isToday: _isTodayDate(day.date),
        ),
      )
      .toList();
  if (days.isNotEmpty) {
    return days;
  }
  final _WeekRange range = _weekRange(0);
  return List<_WeekDay>.generate(7, (int index) {
    final DateTime date = range.start.add(Duration(days: index));
    return _WeekDay(
      label: _weekdayShortByNumber(date.weekday),
      date:
          '${date.month.toString().padLeft(2, '0')}/${date.day.toString().padLeft(2, '0')}',
      isoDate: _formatApiDate(date),
      isToday: _isSameDay(date, DateTime.now()),
    );
  });
}

List<_TimeSlot> _timeSlotsFromData(TimetableData data) {
  return data.slots
      .map(
        (TimetableSlot slot) => _TimeSlot(
          title: slot.title,
          time: slot.time.trim().isEmpty
              ? '${slot.startTime} - ${slot.endTime}'
              : slot.time,
          startTime: slot.startTime,
          endTime: slot.endTime,
        ),
      )
      .toList();
}

List<List<_LessonCell?>> _scheduleRowsFromItems(
  List<TimetableItem> items,
  List<TimetableDay> sourceDays,
  List<_WeekDay> weekDays,
  List<_TimeSlot> timeSlots,
) {
  final Map<String, int> dayIndexByDate = <String, int>{};
  for (int index = 0;
      index < sourceDays.length && index < weekDays.length;
      index += 1) {
    dayIndexByDate[sourceDays[index].date] = index;
  }
  final Map<String, int> slotIndexByKey = <String, int>{};
  for (int index = 0; index < timeSlots.length; index += 1) {
    slotIndexByKey[
        _slotKey(timeSlots[index].startTime, timeSlots[index].endTime)] = index;
  }
  final List<List<_LessonCell?>> rows = List<List<_LessonCell?>>.generate(
    timeSlots.length,
    (_) => List<_LessonCell?>.filled(weekDays.length, null),
  );
  for (final TimetableItem item in items) {
    final int? row = slotIndexByKey[_slotKey(item.startTime, item.endTime)];
    final int? column = dayIndexByDate[item.date];
    if (row == null ||
        column == null ||
        row >= rows.length ||
        column >= weekDays.length) {
      continue;
    }
    rows[row][column] = _LessonCell(
      id: item.id,
      classType: item.classType,
      teachingClassId: item.teachingClassId,
      title: item.lessonName.trim().isEmpty ? '未命名课程' : item.lessonName.trim(),
      person: _lessonPersonText(item),
      status: _lessonStatusFromApi(item.status),
      assistantIds: item.assistantIds,
      classroomId: item.classroomId,
      conflict: item.conflict,
    );
  }
  return rows;
}

List<List<_LessonCell?>> _scheduleRowsFromData(
  TimetableData data,
  List<_WeekDay> weekDays,
  List<_TimeSlot> timeSlots,
) {
  return _scheduleRowsFromItems(data.items, data.days, weekDays, timeSlots);
}

TimetableSummary _timetableSummaryFromItems(List<TimetableItem> items) {
  int unsigned = 0;
  int signed = 0;
  int partial = 0;
  int trial = 0;
  int conflict = 0;
  for (final TimetableItem item in items) {
    switch (_lessonStatusFromApi(item.status)) {
      case _LessonStatus.unsigned:
        unsigned += 1;
        break;
      case _LessonStatus.signed:
        signed += 1;
        break;
      case _LessonStatus.partial:
        partial += 1;
        break;
      case _LessonStatus.trial:
        trial += 1;
        break;
      case _LessonStatus.conflict:
        conflict += 1;
        break;
    }
  }
  return TimetableSummary(
    total: items.length,
    unsigned: unsigned,
    signed: signed,
    partial: partial,
    trial: trial,
    conflict: conflict,
  );
}

String _studentFilterValue(TimetableItem item) {
  final String student = item.studentName.trim();
  if (student.isNotEmpty) {
    return student;
  }
  final String person = item.personName.trim();
  if (person.isNotEmpty) {
    return person;
  }
  final String teachingClass = item.teachingClassName.trim();
  if (teachingClass.isNotEmpty) {
    return teachingClass;
  }
  final String lessonName = item.lessonName.trim();
  if (lessonName.isNotEmpty) {
    return lessonName;
  }
  return '未命名对象';
}

String _courseFilterValue(TimetableItem item) {
  final String lessonName = item.lessonName.trim();
  return lessonName.isEmpty ? '未命名课程' : lessonName;
}

String _callStatusFilterValue(TimetableItem item) {
  switch (_lessonStatusFromApi(item.status)) {
    case _LessonStatus.signed:
      return 'signed';
    case _LessonStatus.partial:
      return 'partial';
    case _LessonStatus.unsigned:
      return 'unsigned';
    case _LessonStatus.trial:
    case _LessonStatus.conflict:
      return '';
  }
}

String _lessonPersonText(TimetableItem item) {
  final String person = item.personName.trim().isEmpty
      ? (item.studentName.trim().isEmpty
          ? item.teachingClassName.trim()
          : item.studentName.trim())
      : item.personName.trim();
  final String classroom = item.classroomName.trim();
  if (person.isEmpty && classroom.isEmpty) {
    return '未分配';
  }
  if (classroom.isEmpty) {
    return person;
  }
  if (person.isEmpty) {
    return classroom;
  }
  return '$person · $classroom';
}

_LessonStatus _lessonStatusFromApi(String status) {
  switch (status.trim()) {
    case 'signed':
      return _LessonStatus.signed;
    case 'partial':
      return _LessonStatus.partial;
    case 'trial':
      return _LessonStatus.trial;
    case 'conflict':
      return _LessonStatus.conflict;
    default:
      return _LessonStatus.unsigned;
  }
}

String _slotKey(String startTime, String endTime) {
  return '${startTime.trim()}-${endTime.trim()}';
}

String _availabilityKey(
  String teacherId,
  String lessonDate,
  String startTime,
  String endTime,
) {
  return '${teacherId.trim()}|${lessonDate.trim()}|'
      '${startTime.trim()}|${endTime.trim()}';
}

String _monthDayLabel(String rawDate) {
  final DateTime? date = DateTime.tryParse(rawDate);
  if (date == null) {
    return rawDate;
  }
  return '${date.month.toString().padLeft(2, '0')}/${date.day.toString().padLeft(2, '0')}';
}

bool _isTodayDate(String rawDate) {
  final DateTime? date = DateTime.tryParse(rawDate);
  if (date == null) {
    return false;
  }
  return _isSameDay(date, DateTime.now());
}

bool _isSameDay(DateTime left, DateTime right) {
  return left.year == right.year &&
      left.month == right.month &&
      left.day == right.day;
}

String _weekdayShortLabel(String rawDate) {
  final DateTime? date = DateTime.tryParse(rawDate);
  if (date == null) {
    return '';
  }
  return _weekdayShortByNumber(date.weekday);
}

String _weekdayShortByNumber(int weekday) {
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

String _weekdayFullByNumber(int weekday) {
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

class _WeekRange {
  const _WeekRange({required this.start, required this.end});

  final DateTime start;
  final DateTime end;

  String get startDate => _formatApiDate(start);

  String get endDate => _formatApiDate(end);
}
