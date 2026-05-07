part of '../smart_timetable_page.dart';

enum _ScheduleMode {
  oneToOne,
  groupClass,
}

enum _TimetableFilterKind {
  student,
  course,
  callStatus,
}

class _TimetableFilterOption {
  const _TimetableFilterOption({
    required this.id,
    required this.label,
  });

  final String id;
  final String label;
}

class _ScheduleCellSlot {
  const _ScheduleCellSlot({
    required this.row,
    required this.column,
    required this.key,
    required this.date,
    required this.startTime,
    required this.endTime,
  });

  final int row;
  final int column;
  final String key;
  final String date;
  final String startTime;
  final String endTime;
}

class _SlotAvailability {
  const _SlotAvailability({
    required this.valid,
    required this.message,
    this.conflictTypes = const <String>[],
  });

  final bool valid;
  final String message;
  final List<String> conflictTypes;
}

class _ScheduleMovePreviewData {
  const _ScheduleMovePreviewData({
    required this.dateLabel,
    required this.timeLabel,
    required this.title,
    required this.courseLabel,
    required this.studentLabel,
    required this.teacherLabel,
    required this.assistantLabel,
  });

  final String dateLabel;
  final String timeLabel;
  final String title;
  final String courseLabel;
  final String studentLabel;
  final String teacherLabel;
  final String assistantLabel;
}

class _PeriodGroupOption {
  const _PeriodGroupOption({
    required this.id,
    required this.name,
    required this.meta,
  });

  final String id;
  final String name;
  final String meta;
}

class _TeacherOption {
  const _TeacherOption({
    required this.id,
    required this.name,
    required this.label,
  });

  final String id;
  final String name;
  final String label;
}

class _WeekDay {
  const _WeekDay({
    required this.label,
    required this.date,
    required this.isoDate,
    this.isToday = false,
  });

  final String label;
  final String date;
  final String isoDate;
  final bool isToday;
}

class _TimeSlot {
  const _TimeSlot({
    required this.title,
    required this.time,
    required this.startTime,
    required this.endTime,
  });

  final String title;
  final String time;
  final String startTime;
  final String endTime;
}

class _LessonCell {
  const _LessonCell({
    required this.id,
    required this.classType,
    required this.teachingClassId,
    required this.title,
    required this.person,
    required this.status,
    required this.assistantIds,
    required this.classroomId,
    this.conflict = false,
  });

  final String id;
  final int classType;
  final String teachingClassId;
  final String title;
  final String person;
  final _LessonStatus status;
  final List<String>? assistantIds;
  final String? classroomId;
  final bool conflict;

  ScheduleTargetType? get scheduleTargetType {
    if (classType == 2) {
      return ScheduleTargetType.oneToOne;
    }
    if (classType == 1) {
      return ScheduleTargetType.groupClass;
    }
    return null;
  }
}

class _LessonDragData {
  const _LessonDragData({required this.row, required this.column});

  final int row;
  final int column;
}

enum _LessonStatus {
  unsigned,
  signed,
  partial,
  trial,
  conflict,
}

extension _LessonStatusView on _LessonStatus {
  String get label {
    switch (this) {
      case _LessonStatus.unsigned:
        return '未点名';
      case _LessonStatus.signed:
        return '已点名';
      case _LessonStatus.partial:
        return '部分';
      case _LessonStatus.trial:
        return '试听';
      case _LessonStatus.conflict:
        return '冲突';
    }
  }

  Color get foreground {
    switch (this) {
      case _LessonStatus.unsigned:
        return _SmartColors.blue;
      case _LessonStatus.signed:
        return _SmartColors.gray;
      case _LessonStatus.partial:
        return _SmartColors.amber;
      case _LessonStatus.trial:
        return _SmartColors.green;
      case _LessonStatus.conflict:
        return _SmartColors.danger;
    }
  }

  Color get background {
    switch (this) {
      case _LessonStatus.unsigned:
        return const Color(0xFFEAF3FF);
      case _LessonStatus.signed:
        return const Color(0xFFF1F3F6);
      case _LessonStatus.partial:
        return const Color(0xFFFFF1D8);
      case _LessonStatus.trial:
        return const Color(0xFFEDF8EC);
      case _LessonStatus.conflict:
        return const Color(0xFFFFE8E8);
    }
  }
}

class _SmartColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color surface = Color(0xFFFFFDF9);
  static const Color card = Color(0xFFFFFDF9);
  static const Color slot = Color(0xB8FFFDF9);
  static const Color ink = Color(0xFF3F2B22);
  static const Color text = Color(0xFF6F5B50);
  static const Color muted = Color(0xFFA7958B);
  static const Color line = Color(0xFFEAD7C9);
  static const Color lineSoft = Color(0xFFF4E8DF);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
  static const Color green = Color(0xFF6F9F70);
  static const Color blue = Color(0xFF3F82D2);
  static const Color amber = Color(0xFFD99427);
  static const Color gray = Color(0xFF98A2AD);
  static const Color danger = Color(0xFFD94F4F);
  static const Color todayHeader = Color(0xFFFFF4EB);
  static const Color todayCell = Color(0xFFFFF9F4);
}

const double _smartDesignHeight = 768;
const double _smartMinDesignWidth = 1024;
const double _smartWideDesignWidth = 1366;
const double _headerHeight = 44;
const double _rowHeight = 62;
