import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'timetable_client.dart';

class SmartTimetablePage extends StatefulWidget {
  const SmartTimetablePage({
    this.timetableClient = const ApiTimetableClient(),
    super.key,
  });

  final TimetableClient timetableClient;

  @override
  State<SmartTimetablePage> createState() => _SmartTimetablePageState();
}

class _SmartTimetablePageState extends State<SmartTimetablePage> {
  int _periodGroupIndex = 0;
  int _teacherIndex = 0;
  int _weekOffset = 0;
  int _loadSequence = 0;
  bool _teacherDropdownOpen = false;
  bool _loading = true;
  String? _errorMessage;
  String _selectedPeriodGroupId = '';
  String _selectedTeacherId = '';
  TimetableData _data = TimetableData.fallback();
  List<_PeriodGroupOption> _periodGroups = const <_PeriodGroupOption>[];
  List<_TeacherOption> _teachers = const <_TeacherOption>[];
  List<_WeekDay> _weekDays = const <_WeekDay>[];
  List<_TimeSlot> _timeSlots = const <_TimeSlot>[];
  List<List<_LessonCell?>> _scheduleRows = const <List<_LessonCell?>>[];

  @override
  void initState() {
    super.initState();
    _applyTimetableData(_data, preserveTeacherSelection: false);
    _loadTimetable();
  }

  Future<void> _loadTimetable(
      {String? teacherId, String? periodGroupId}) async {
    final int sequence = ++_loadSequence;
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_timetableAuthTokenStorageKey) ?? '';
    if (token.trim().isEmpty) {
      if (!mounted || sequence != _loadSequence) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '登录已失效，请重新登录';
      });
      return;
    }
    final _WeekRange range = _weekRange(_weekOffset);
    if (mounted) {
      setState(() {
        _loading = true;
        _errorMessage = null;
      });
    }
    try {
      final TimetableData data = await widget.timetableClient.fetchTimetable(
        token,
        startDate: range.startDate,
        endDate: range.endDate,
        teacherId: teacherId ?? _selectedTeacherId,
        periodGroupId: periodGroupId ?? _selectedPeriodGroupId,
      );
      if (!mounted || sequence != _loadSequence) {
        return;
      }
      setState(() {
        _data = data;
        _selectedPeriodGroupId = data.selectedPeriodGroupId;
        _selectedTeacherId = data.selectedTeacherId;
        _applyTimetableData(data);
        _loading = false;
      });
    } on TimetableApiException catch (error) {
      if (!mounted || sequence != _loadSequence) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || sequence != _loadSequence) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '排课日程加载失败：$error';
      });
    }
  }

  void _applyTimetableData(
    TimetableData data, {
    bool preserveTeacherSelection = true,
  }) {
    _periodGroups = _periodGroupOptionsFromData(data);
    _periodGroupIndex = _periodGroups.indexWhere(
      (_PeriodGroupOption item) => item.id == data.selectedPeriodGroupId,
    );
    if (_periodGroupIndex < 0) {
      _periodGroupIndex = 0;
    }
    if (_periodGroups.isNotEmpty) {
      _selectedPeriodGroupId = _periodGroups[_periodGroupIndex].id;
    }
    _teachers = _teacherOptionsFromData(data);
    _teacherIndex = _teachers.indexWhere(
      (_TeacherOption item) => item.id == data.selectedTeacherId,
    );
    if (_teacherIndex < 0) {
      _teacherIndex = 0;
    }
    if (!preserveTeacherSelection && _teachers.isNotEmpty) {
      _selectedTeacherId = _teachers[_teacherIndex].id;
    }
    _weekDays = _weekDaysFromData(data);
    _timeSlots = _timeSlotsFromData(data);
    _scheduleRows = _scheduleRowsFromData(data, _weekDays, _timeSlots);
  }

  void _selectPeriodGroup(int index) {
    if (index < 0 || index >= _periodGroups.length) {
      return;
    }
    final _PeriodGroupOption group = _periodGroups[index];
    if (group.id == _selectedPeriodGroupId && !_loading) {
      return;
    }
    setState(() {
      _periodGroupIndex = index;
      _selectedPeriodGroupId = group.id;
      _selectedTeacherId = '';
      _teacherIndex = 0;
      _teacherDropdownOpen = false;
      _loading = true;
      _errorMessage = null;
    });
    _loadTimetable(teacherId: '', periodGroupId: group.id);
  }

  void _selectTeacher(int index) {
    if (index < 0 || index >= _teachers.length) {
      return;
    }
    final String teacherId = _teachers[index].id;
    final _TeacherOption teacher = _teachers[index];
    setState(() {
      _teacherIndex = index;
      _selectedTeacherId = teacherId;
      _teacherDropdownOpen = false;
      _loading = true;
      _errorMessage = null;
      _data = _loadingTimetableDataForTeacher(teacher);
      _applyTimetableData(_data);
    });
    _loadTimetable(teacherId: teacherId, periodGroupId: _selectedPeriodGroupId);
  }

  void _changeWeek(int delta) {
    setState(() {
      _weekOffset += delta;
      _teacherDropdownOpen = false;
      _loading = true;
      _errorMessage = null;
      _data = _loadingTimetableDataForTeacher(
        _teachers.isEmpty
            ? const _TeacherOption(id: '', name: '当前老师', label: '当前老师')
            : _teachers[_teacherIndex.clamp(0, _teachers.length - 1)],
      );
      _applyTimetableData(_data);
    });
    _loadTimetable();
  }

  void _backToCurrentWeek() {
    setState(() {
      _weekOffset = 0;
      _teacherDropdownOpen = false;
      _loading = true;
      _errorMessage = null;
      _data = _loadingTimetableDataForTeacher(
        _teachers.isEmpty
            ? const _TeacherOption(id: '', name: '当前老师', label: '当前老师')
            : _teachers[_teacherIndex.clamp(0, _teachers.length - 1)],
      );
      _applyTimetableData(_data);
    });
    _loadTimetable();
  }

  TimetableData _loadingTimetableDataForTeacher(_TeacherOption teacher) {
    final _WeekRange range = _weekRange(_weekOffset);
    return TimetableData(
      startDate: range.startDate,
      endDate: range.endDate,
      selectedPeriodGroupId: _selectedPeriodGroupId,
      selectedTeacherId: teacher.id,
      selectedTeacherName: teacher.name,
      periodGroups: _data.periodGroups,
      teachers: _data.teachers,
      days: _timetableDaysForRange(range),
      slots: const <TimetableSlot>[],
      items: const <TimetableItem>[],
      summary: const TimetableSummary(),
    );
  }

  void _moveLesson(_LessonDragData source, int targetRow, int targetColumn) {
    if (source.row == targetRow && source.column == targetColumn) {
      return;
    }
    setState(() {
      final _LessonCell? sourceLesson =
          _scheduleRows[source.row][source.column];
      if (sourceLesson == null) {
        return;
      }
      final _LessonCell? targetLesson = _scheduleRows[targetRow][targetColumn];
      _scheduleRows[targetRow][targetColumn] = sourceLesson;
      _scheduleRows[source.row][source.column] = targetLesson;
    });
  }

  @override
  Widget build(BuildContext context) {
    final _TeacherOption teacher = _teachers.isEmpty
        ? const _TeacherOption(id: '', name: '当前老师', label: '当前老师')
        : _teachers[_teacherIndex.clamp(0, _teachers.length - 1)];
    return Scaffold(
      body: _SmartTimetableViewport(
        child: _SmartTimetableScreen(
          teacher: teacher,
          periodGroups: _periodGroups,
          periodGroupIndex: _periodGroupIndex,
          teachers: _teachers,
          teacherIndex: _teacherIndex,
          teacherDropdownOpen: _teacherDropdownOpen,
          scheduleRows: _scheduleRows,
          weekDays: _weekDays,
          timeSlots: _timeSlots,
          summary: _data.summary,
          errorMessage: _errorMessage,
          dateRange: _dateRangeText(_weekOffset),
          onBack: () => Navigator.of(context).maybePop(),
          onPrevWeek: () => _changeWeek(-1),
          onNextWeek: () => _changeWeek(1),
          onToday: _backToCurrentWeek,
          onPeriodGroupSelected: _selectPeriodGroup,
          onTeacherToggle: () => setState(
            () => _teacherDropdownOpen = !_teacherDropdownOpen,
          ),
          onTeacherSelected: _selectTeacher,
          onTeacherDropdownClose: () =>
              setState(() => _teacherDropdownOpen = false),
          onRefresh: _loadTimetable,
          onLessonMove: _moveLesson,
        ),
      ),
    );
  }
}

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

List<List<_LessonCell?>> _scheduleRowsFromData(
  TimetableData data,
  List<_WeekDay> weekDays,
  List<_TimeSlot> timeSlots,
) {
  final Map<String, int> dayIndexByDate = <String, int>{};
  for (int index = 0;
      index < data.days.length && index < weekDays.length;
      index += 1) {
    dayIndexByDate[data.days[index].date] = index;
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
  for (final TimetableItem item in data.items) {
    final int? row = slotIndexByKey[_slotKey(item.startTime, item.endTime)];
    final int? column = dayIndexByDate[item.date];
    if (row == null ||
        column == null ||
        row >= rows.length ||
        column >= weekDays.length) {
      continue;
    }
    rows[row][column] = _LessonCell(
      title: item.lessonName.trim().isEmpty ? '未命名课程' : item.lessonName.trim(),
      person: _lessonPersonText(item),
      status: _lessonStatusFromApi(item.status),
      conflict: item.conflict,
    );
  }
  return rows;
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

const SystemUiOverlayStyle _smartImmersiveOverlayStyle = SystemUiOverlayStyle(
  statusBarColor: Colors.transparent,
  statusBarIconBrightness: Brightness.dark,
  statusBarBrightness: Brightness.light,
  systemStatusBarContrastEnforced: false,
  systemNavigationBarColor: Colors.transparent,
  systemNavigationBarDividerColor: Colors.transparent,
  systemNavigationBarIconBrightness: Brightness.dark,
  systemNavigationBarContrastEnforced: false,
);

class _SmartTimetableViewport extends StatelessWidget {
  const _SmartTimetableViewport({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: _smartImmersiveOverlayStyle,
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final Size screenSize = MediaQuery.sizeOf(context);
          final double viewportWidth = constraints.maxWidth.isFinite
              ? constraints.maxWidth
              : screenSize.width;
          final double viewportHeight = constraints.maxHeight.isFinite
              ? constraints.maxHeight
              : screenSize.height;
          final double aspect = viewportWidth / viewportHeight;
          final double designWidth =
              math.max(_smartMinDesignWidth, _smartDesignHeight * aspect);

          return ColoredBox(
            color: _SmartColors.page,
            child: Center(
              child: FittedBox(
                fit: BoxFit.contain,
                child: SizedBox(
                  width: designWidth,
                  height: _smartDesignHeight,
                  child: child,
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

class _SmartTimetableScreen extends StatelessWidget {
  const _SmartTimetableScreen({
    required this.teacher,
    required this.periodGroups,
    required this.periodGroupIndex,
    required this.teachers,
    required this.teacherIndex,
    required this.teacherDropdownOpen,
    required this.scheduleRows,
    required this.weekDays,
    required this.timeSlots,
    required this.summary,
    required this.errorMessage,
    required this.dateRange,
    required this.onBack,
    required this.onPrevWeek,
    required this.onNextWeek,
    required this.onToday,
    required this.onPeriodGroupSelected,
    required this.onTeacherToggle,
    required this.onTeacherSelected,
    required this.onTeacherDropdownClose,
    required this.onRefresh,
    required this.onLessonMove,
  });

  final _TeacherOption teacher;
  final List<_PeriodGroupOption> periodGroups;
  final int periodGroupIndex;
  final List<_TeacherOption> teachers;
  final int teacherIndex;
  final bool teacherDropdownOpen;
  final List<List<_LessonCell?>> scheduleRows;
  final List<_WeekDay> weekDays;
  final List<_TimeSlot> timeSlots;
  final TimetableSummary summary;
  final String? errorMessage;
  final String dateRange;
  final VoidCallback onBack;
  final VoidCallback onPrevWeek;
  final VoidCallback onNextWeek;
  final VoidCallback onToday;
  final ValueChanged<int> onPeriodGroupSelected;
  final VoidCallback onTeacherToggle;
  final ValueChanged<int> onTeacherSelected;
  final VoidCallback onTeacherDropdownClose;
  final VoidCallback onRefresh;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width = constraints.maxWidth.isFinite
            ? constraints.maxWidth
            : _smartWideDesignWidth;
        final bool compact = width < 1180;
        final double pagePadding = compact ? 24 : 32;

        final double teacherWidth = compact ? 190 : 224;
        final double primaryWidth = compact ? 104 : 110;

        return ColoredBox(
          color: _SmartColors.page,
          child: Stack(
            children: <Widget>[
              Padding(
                padding: EdgeInsets.fromLTRB(pagePadding, 24, pagePadding, 24),
                child: Column(
                  children: <Widget>[
                    _TimetableTopBar(
                      compact: compact,
                      teacher: teacher,
                      teacherDropdownOpen: teacherDropdownOpen,
                      teacherWidth: teacherWidth,
                      primaryWidth: primaryWidth,
                      dateRange: dateRange,
                      onBack: onBack,
                      onPrevWeek: onPrevWeek,
                      onNextWeek: onNextWeek,
                      onToday: onToday,
                      onTeacherToggle: onTeacherToggle,
                    ),
                    const SizedBox(height: 10),
                    _TimetableSubBar(
                      compact: compact,
                      periodGroups: periodGroups,
                      periodGroupIndex: periodGroupIndex,
                      errorMessage: errorMessage,
                      onPeriodGroupSelected: onPeriodGroupSelected,
                      onRefresh: onRefresh,
                    ),
                    const SizedBox(height: 4),
                    _TimetableSummary(compact: compact, summary: summary),
                    const SizedBox(height: 4),
                    Expanded(
                      child: _TimetableBoard(
                        compact: compact,
                        rows: scheduleRows,
                        weekDays: weekDays,
                        timeSlots: timeSlots,
                        onLessonMove: onLessonMove,
                      ),
                    ),
                  ],
                ),
              ),
              if (teacherDropdownOpen) ...<Widget>[
                Positioned.fill(
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onTap: onTeacherDropdownClose,
                  ),
                ),
                Positioned(
                  top: 82,
                  right: pagePadding + primaryWidth + 10,
                  width: teacherWidth,
                  child: _TeacherDropdownPanel(
                    teachers: teachers,
                    selectedIndex: teacherIndex,
                    onSelected: onTeacherSelected,
                  ),
                ),
              ],
            ],
          ),
        );
      },
    );
  }
}

class _TimetableTopBar extends StatelessWidget {
  const _TimetableTopBar({
    required this.compact,
    required this.teacher,
    required this.teacherDropdownOpen,
    required this.teacherWidth,
    required this.primaryWidth,
    required this.dateRange,
    required this.onBack,
    required this.onPrevWeek,
    required this.onNextWeek,
    required this.onToday,
    required this.onTeacherToggle,
  });

  final bool compact;
  final _TeacherOption teacher;
  final bool teacherDropdownOpen;
  final double teacherWidth;
  final double primaryWidth;
  final String dateRange;
  final VoidCallback onBack;
  final VoidCallback onPrevWeek;
  final VoidCallback onNextWeek;
  final VoidCallback onToday;
  final VoidCallback onTeacherToggle;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 56,
      child: Row(
        children: <Widget>[
          SizedBox(
            width: compact ? 188 : 210,
            child: Row(
              children: <Widget>[
                _IconShell(
                  size: 42,
                  icon: Icons.chevron_left_rounded,
                  onTap: onBack,
                ),
                const SizedBox(width: 14),
                const Expanded(
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      '排课日程',
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: _SmartColors.ink,
                        fontSize: 25,
                        height: 1.05,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                _DateSwitch(
                  width: compact ? 302 : 352,
                  dateRange: dateRange,
                  onPrev: onPrevWeek,
                  onNext: onNextWeek,
                ),
                const SizedBox(width: 10),
                _ToolbarButton(
                  width: compact ? 74 : 82,
                  icon: Icons.format_list_bulleted_rounded,
                  label: '今天',
                  onTap: onToday,
                ),
                const SizedBox(width: 10),
                SizedBox(
                  width: teacherWidth,
                  child: _TeacherSelector(
                    teacher: teacher,
                    isOpen: teacherDropdownOpen,
                    onTap: onTeacherToggle,
                  ),
                ),
                const SizedBox(width: 10),
                _PrimaryButton(width: primaryWidth),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _PeriodGroupTabs extends StatelessWidget {
  const _PeriodGroupTabs({
    required this.groups,
    required this.selectedIndex,
    required this.onSelected,
    required this.compact,
  });

  final List<_PeriodGroupOption> groups;
  final int selectedIndex;
  final ValueChanged<int> onSelected;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final List<_PeriodGroupOption> displayGroups = groups.isEmpty
        ? const <_PeriodGroupOption>[
            _PeriodGroupOption(
              id: 'default',
              name: '默认时段',
              meta: '08:00 - 18:20 · 11节',
            ),
          ]
        : groups;
    return SizedBox(
      height: 34,
      child: Row(
        children: <Widget>[
          Expanded(
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              physics: const BouncingScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                final _PeriodGroupOption group = displayGroups[index];
                return _PeriodGroupTab(
                  key: ValueKey<String>('period-group-tab-${group.id}'),
                  group: group,
                  selected: index == selectedIndex,
                  compact: compact,
                  onTap: () => onSelected(index),
                );
              },
              separatorBuilder: (_, __) => SizedBox(width: compact ? 6 : 8),
              itemCount: displayGroups.length,
            ),
          ),
        ],
      ),
    );
  }
}

class _PeriodGroupTab extends StatelessWidget {
  const _PeriodGroupTab({
    required this.group,
    required this.selected,
    required this.compact,
    required this.onTap,
    super.key,
  });

  final _PeriodGroupOption group;
  final bool selected;
  final bool compact;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(11),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(11),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          width: compact ? 74 : 82,
          height: 34,
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF0E5) : _SmartColors.card,
            border: Border.all(
              color: selected ? _SmartColors.orange : _SmartColors.line,
            ),
            borderRadius: BorderRadius.circular(11),
            boxShadow: selected
                ? const <BoxShadow>[
                    BoxShadow(
                      color: Color(0x24E96F43),
                      blurRadius: 14,
                      offset: Offset(0, 6),
                    ),
                  ]
                : null,
          ),
          child: Text(
            group.name,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              color: _SmartColors.text,
              fontSize: 12,
              height: 1,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      ),
    );
  }
}

class _TimetableSubBar extends StatelessWidget {
  const _TimetableSubBar({
    required this.compact,
    required this.periodGroups,
    required this.periodGroupIndex,
    required this.errorMessage,
    required this.onPeriodGroupSelected,
    required this.onRefresh,
  });

  final bool compact;
  final List<_PeriodGroupOption> periodGroups;
  final int periodGroupIndex;
  final String? errorMessage;
  final ValueChanged<int> onPeriodGroupSelected;
  final VoidCallback onRefresh;

  @override
  Widget build(BuildContext context) {
    final double periodGroupWidth = compact ? 250 : 284;
    return SizedBox(
      height: 44,
      child: Row(
        children: <Widget>[
          const Spacer(),
          SizedBox(width: compact ? 8 : 10),
          SizedBox(
            width: periodGroupWidth,
            child: _PeriodGroupTabs(
              compact: compact,
              groups: periodGroups,
              selectedIndex: periodGroupIndex,
              onSelected: onPeriodGroupSelected,
            ),
          ),
          SizedBox(width: compact ? 8 : 10),
          if (errorMessage != null)
            _TimetableLoadStatus(message: errorMessage!, onRefresh: onRefresh),
          if (errorMessage != null) const SizedBox(width: 10),
          const _FilterButton(icon: Icons.filter_list_rounded, label: '全部课程'),
          const SizedBox(width: 8),
          const _FilterButton(
            icon: Icons.library_books_outlined,
            label: '全部状态',
          ),
        ],
      ),
    );
  }
}

class _TimetableSummary extends StatelessWidget {
  const _TimetableSummary({required this.compact, required this.summary});

  final bool compact;
  final TimetableSummary summary;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 38,
      child: Row(
        children: <Widget>[
          SizedBox(width: compact ? 6 : 12),
          const _SummaryAccent(),
          const SizedBox(width: 11),
          Text.rich(
            TextSpan(
              children: <InlineSpan>[
                const TextSpan(text: '共 '),
                TextSpan(
                  text: '${summary.total}',
                  style: const TextStyle(color: _SmartColors.orangeDeep),
                ),
                const TextSpan(text: ' 个日程，未点名 '),
                TextSpan(
                  text: '${summary.unsigned}',
                  style: const TextStyle(color: _SmartColors.orangeDeep),
                ),
                const TextSpan(text: ' 个，冲突 '),
                TextSpan(
                  text: '${summary.conflict}',
                  style: const TextStyle(color: _SmartColors.orangeDeep),
                ),
                const TextSpan(text: ' 个'),
              ],
            ),
            style: const TextStyle(
              color: _SmartColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          const _LegendItem(color: _SmartColors.blue, label: '未点名'),
          const SizedBox(width: 14),
          const _LegendItem(color: _SmartColors.gray, label: '已点名'),
          const SizedBox(width: 14),
          const _LegendItem(color: _SmartColors.amber, label: '部分点名'),
          const SizedBox(width: 14),
          const _LegendItem(color: _SmartColors.green, label: '试听'),
          const SizedBox(width: 14),
          const _LegendItem(color: _SmartColors.danger, label: '冲突'),
          const SizedBox(width: 2),
        ],
      ),
    );
  }
}

class _TimetableBoard extends StatelessWidget {
  const _TimetableBoard({
    required this.compact,
    required this.rows,
    required this.weekDays,
    required this.timeSlots,
    required this.onLessonMove,
  });

  final bool compact;
  final List<List<_LessonCell?>> rows;
  final List<_WeekDay> weekDays;
  final List<_TimeSlot> timeSlots;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;

  @override
  Widget build(BuildContext context) {
    final double leftWidth = compact ? 112 : 118;
    final int rowCount = math.max(timeSlots.length, rows.length);
    final List<List<_LessonCell?>> displayRows =
        List<List<_LessonCell?>>.generate(rowCount, (int rowIndex) {
      final List<_LessonCell?> source =
          rowIndex < rows.length ? rows[rowIndex] : const <_LessonCell?>[];
      return List<_LessonCell?>.generate(
        weekDays.length,
        (int column) => column < source.length ? source[column] : null,
      );
    });

    return Align(
      alignment: Alignment.topCenter,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(14),
        child: Container(
          key: const ValueKey<String>('smart-timetable-board'),
          color: Colors.white,
          foregroundDecoration: BoxDecoration(
            border: Border.all(color: _SmartColors.line),
            borderRadius: BorderRadius.circular(14),
          ),
          child: Column(
            children: <Widget>[
              SizedBox(
                key: const ValueKey<String>('smart-timetable-header'),
                height: _headerHeight,
                child: Row(
                  children: <Widget>[
                    SizedBox(
                      width: leftWidth,
                      child: const _DiagonalHeaderCell(),
                    ),
                    Expanded(
                      child: _WeekHeaderRow(
                        key: const ValueKey<String>('smart-week-header'),
                        weekDays: weekDays,
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  key: const ValueKey<String>('smart-timetable-scroll-body'),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      SizedBox(
                        width: leftWidth,
                        child: _TimeRail(
                          rowCount: rowCount,
                          timeSlots: timeSlots,
                        ),
                      ),
                      Expanded(
                        child: _ScheduleGrid(
                          rows: displayRows,
                          onLessonMove: onLessonMove,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _TimeRail extends StatelessWidget {
  const _TimeRail({required this.rowCount, required this.timeSlots});

  final int rowCount;
  final List<_TimeSlot> timeSlots;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      key: const ValueKey<String>('smart-time-rail'),
      color: _SmartColors.slot,
      child: Column(
        children: <Widget>[
          for (int index = 0; index < rowCount; index += 1)
            _TimeSlotCell(
              slot: _timeSlotForIndex(index, timeSlots),
              isLast: index == rowCount - 1,
            ),
        ],
      ),
    );
  }
}

_TimeSlot _timeSlotForIndex(int index, List<_TimeSlot> timeSlots) {
  if (index < timeSlots.length) {
    return timeSlots[index];
  }
  return _TimeSlot(
    title: '第${index + 1}节',
    time: '',
    startTime: '',
    endTime: '',
  );
}

class _DiagonalHeaderCell extends StatelessWidget {
  const _DiagonalHeaderCell();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: _headerHeight,
      decoration: BoxDecoration(
        color: _SmartColors.surface,
        border: const Border(
          right: BorderSide(color: _SmartColors.line),
          bottom: BorderSide(color: _SmartColors.line),
        ),
      ),
      child: Stack(
        children: <Widget>[
          Positioned.fill(child: CustomPaint(painter: _DiagonalPainter())),
          const Positioned(
            left: 10,
            bottom: 8,
            child: Text(
              '时段',
              style: TextStyle(
                color: _SmartColors.muted,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const Positioned(
            right: 10,
            top: 8,
            child: Text(
              '日期',
              style: TextStyle(
                color: _SmartColors.muted,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _TimeSlotCell extends StatelessWidget {
  const _TimeSlotCell({required this.slot, required this.isLast});

  final _TimeSlot slot;
  final bool isLast;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: _rowHeight,
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(10, 8, 4, 0),
      decoration: BoxDecoration(
        color: _SmartColors.slot,
        border: Border(
          right: const BorderSide(color: _SmartColors.line),
          bottom: BorderSide(
            color: isLast ? Colors.transparent : _SmartColors.lineSoft,
          ),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            slot.title,
            style: const TextStyle(
              color: _SmartColors.ink,
              fontSize: 13,
              height: 1.1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 7),
          Text(
            slot.time,
            style: const TextStyle(
              color: _SmartColors.muted,
              fontSize: 11,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScheduleGrid extends StatelessWidget {
  const _ScheduleGrid({required this.rows, required this.onLessonMove});

  final List<List<_LessonCell?>> rows;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      key: const ValueKey<String>('smart-schedule-grid'),
      color: Colors.white,
      child: Column(
        children: <Widget>[
          for (int row = 0; row < rows.length; row += 1)
            _ScheduleGridRow(
              rowIndex: row,
              row: rows[row],
              isLastRow: row == rows.length - 1,
              onLessonMove: onLessonMove,
            ),
        ],
      ),
    );
  }
}

class _WeekHeaderRow extends StatelessWidget {
  const _WeekHeaderRow({required this.weekDays, super.key});

  final List<_WeekDay> weekDays;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _headerHeight,
      child: Row(
        children: <Widget>[
          for (int index = 0; index < weekDays.length; index += 1)
            Expanded(
              child: _WeekHeaderCell(
                day: weekDays[index],
                isLast: index == weekDays.length - 1,
              ),
            ),
        ],
      ),
    );
  }
}

class _WeekHeaderCell extends StatelessWidget {
  const _WeekHeaderCell({
    required this.day,
    required this.isLast,
  });

  final _WeekDay day;
  final bool isLast;

  @override
  Widget build(BuildContext context) {
    final bool isToday = day.isToday;
    return Container(
      height: _headerHeight,
      decoration: BoxDecoration(
        color: isToday ? _SmartColors.todayHeader : _SmartColors.surface,
        border: Border(
          right: BorderSide(
            color: isLast ? Colors.transparent : _SmartColors.lineSoft,
          ),
          bottom: const BorderSide(color: _SmartColors.line),
        ),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Text(
                day.label,
                style: TextStyle(
                  color: isToday ? _SmartColors.orangeDeep : _SmartColors.ink,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
              if (isToday) ...<Widget>[
                const SizedBox(width: 5),
                Container(
                  height: 16,
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFE1D1),
                    borderRadius: BorderRadius.circular(999),
                  ),
                  child: const Text(
                    '今天',
                    style: TextStyle(
                      color: _SmartColors.orangeDeep,
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ],
            ],
          ),
          const SizedBox(height: 2),
          Text(
            day.date,
            style: TextStyle(
              color: isToday ? _SmartColors.orangeDeep : _SmartColors.muted,
              fontSize: 11,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScheduleGridRow extends StatelessWidget {
  const _ScheduleGridRow({
    required this.rowIndex,
    required this.row,
    required this.isLastRow,
    required this.onLessonMove,
  });

  final int rowIndex;
  final List<_LessonCell?> row;
  final bool isLastRow;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _rowHeight,
      child: Row(
        children: <Widget>[
          for (int column = 0; column < row.length; column += 1)
            Expanded(
              child: _ScheduleGridCell(
                lesson: row[column],
                rowIndex: rowIndex,
                columnIndex: column,
                isToday: column == 0,
                isLastColumn: column == row.length - 1,
                isLastRow: isLastRow,
                onLessonMove: onLessonMove,
              ),
            ),
        ],
      ),
    );
  }
}

class _ScheduleGridCell extends StatelessWidget {
  const _ScheduleGridCell({
    required this.lesson,
    required this.rowIndex,
    required this.columnIndex,
    required this.isToday,
    required this.isLastColumn,
    required this.isLastRow,
    required this.onLessonMove,
  });

  final _LessonCell? lesson;
  final int rowIndex;
  final int columnIndex;
  final bool isToday;
  final bool isLastColumn;
  final bool isLastRow;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;

  @override
  Widget build(BuildContext context) {
    return DragTarget<_LessonDragData>(
      onWillAcceptWithDetails: (DragTargetDetails<_LessonDragData> details) =>
          details.data.row != rowIndex || details.data.column != columnIndex,
      onAcceptWithDetails: (DragTargetDetails<_LessonDragData> details) =>
          onLessonMove(details.data, rowIndex, columnIndex),
      builder: (
        BuildContext context,
        List<_LessonDragData?> candidateData,
        List<dynamic> rejectedData,
      ) {
        final bool highlighted = candidateData.isNotEmpty;
        return AnimatedContainer(
          key: ValueKey<String>('schedule-cell-$rowIndex-$columnIndex'),
          duration: const Duration(milliseconds: 120),
          height: _rowHeight,
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(
            color: highlighted
                ? const Color(0xFFFFF0E6)
                : isToday
                    ? _SmartColors.todayCell
                    : Colors.white.withOpacity(.72),
            border: Border(
              right: BorderSide(
                color:
                    isLastColumn ? Colors.transparent : _SmartColors.lineSoft,
              ),
              bottom: BorderSide(
                color: isLastRow ? Colors.transparent : _SmartColors.lineSoft,
              ),
            ),
          ),
          child: Stack(
            children: <Widget>[
              if (highlighted)
                Positioned.fill(
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      border: Border.all(
                        color: _SmartColors.orange.withOpacity(.42),
                        width: 1.2,
                      ),
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                ),
              if (lesson == null)
                const SizedBox.expand()
              else
                _DraggableLessonBlock(
                  lesson: lesson!,
                  source: _LessonDragData(row: rowIndex, column: columnIndex),
                ),
            ],
          ),
        );
      },
    );
  }
}

class _DraggableLessonBlock extends StatelessWidget {
  const _DraggableLessonBlock({required this.lesson, required this.source});

  final _LessonCell lesson;
  final _LessonDragData source;

  @override
  Widget build(BuildContext context) {
    return Draggable<_LessonDragData>(
      key: ValueKey<String>('lesson-${source.row}-${source.column}'),
      data: source,
      feedback: Material(
        color: Colors.transparent,
        child: SizedBox(
          width: 156,
          height: 50,
          child: _LessonBlock(lesson, elevated: true),
        ),
      ),
      childWhenDragging: Opacity(
        opacity: .32,
        child: _LessonBlock(lesson),
      ),
      child: _LessonBlock(lesson),
    );
  }
}

class _LessonBlock extends StatelessWidget {
  const _LessonBlock(this.lesson, {this.elevated = false});

  final _LessonCell lesson;
  final bool elevated;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 50,
      padding: const EdgeInsets.fromLTRB(9, 7, 9, 6),
      decoration: BoxDecoration(
        color: lesson.status.background,
        borderRadius: BorderRadius.circular(9),
        boxShadow: elevated
            ? const <BoxShadow>[
                BoxShadow(
                  color: Color(0x26000000),
                  blurRadius: 18,
                  offset: Offset(0, 10),
                ),
              ]
            : null,
      ),
      child: Stack(
        children: <Widget>[
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: <Widget>[
              Text(
                lesson.title,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _SmartColors.ink,
                  fontSize: 12,
                  height: 1.1,
                  fontWeight: FontWeight.w900,
                ),
              ),
              Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      lesson.person,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _SmartColors.text,
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    lesson.status.label,
                    style: TextStyle(
                      color: lesson.status.foreground,
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ],
              ),
            ],
          ),
          if (lesson.conflict)
            const Positioned(
              right: 0,
              top: 0,
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: _SmartColors.danger,
                  shape: BoxShape.circle,
                ),
                child: SizedBox(width: 7, height: 7),
              ),
            ),
        ],
      ),
    );
  }
}

class _TeacherSelector extends StatelessWidget {
  const _TeacherSelector({
    required this.teacher,
    required this.isOpen,
    required this.onTap,
  });

  final _TeacherOption teacher;
  final bool isOpen;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: _ShellBox(
        height: 42,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        child: Row(
          children: <Widget>[
            const Icon(
              Icons.person_outline_rounded,
              color: _SmartColors.ink,
              size: 18,
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    teacher.label,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.muted,
                      fontSize: 10,
                      height: 1.1,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    teacher.name,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 14,
                      height: 1.05,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 6),
            AnimatedRotation(
              turns: isOpen ? .5 : 0,
              duration: const Duration(milliseconds: 160),
              child: const Icon(
                Icons.keyboard_arrow_down_rounded,
                color: _SmartColors.ink,
                size: 18,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _TeacherDropdownPanel extends StatelessWidget {
  const _TeacherDropdownPanel({
    required this.teachers,
    required this.selectedIndex,
    required this.onSelected,
  });

  final List<_TeacherOption> teachers;
  final int selectedIndex;
  final ValueChanged<int> onSelected;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: Container(
        padding: const EdgeInsets.fromLTRB(10, 10, 10, 8),
        decoration: BoxDecoration(
          color: _SmartColors.card,
          border: Border.all(color: _SmartColors.line),
          borderRadius: BorderRadius.circular(16),
          boxShadow: const <BoxShadow>[
            BoxShadow(
              color: Color(0x1AB05F32),
              blurRadius: 24,
              offset: Offset(0, 14),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const Padding(
              padding: EdgeInsets.fromLTRB(8, 2, 8, 8),
              child: Text(
                '切换老师课表',
                style: TextStyle(
                  color: _SmartColors.ink,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            if (teachers.isEmpty)
              Container(
                height: 44,
                padding: const EdgeInsets.symmetric(horizontal: 9),
                alignment: Alignment.centerLeft,
                child: const Text(
                  '该时段组暂无老师',
                  style: TextStyle(
                    color: _SmartColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              )
            else
              for (int index = 0; index < teachers.length; index += 1)
                _TeacherDropdownItem(
                  teacher: teachers[index],
                  selected: index == selectedIndex,
                  onTap: () => onSelected(index),
                ),
          ],
        ),
      ),
    );
  }
}

class _TeacherDropdownItem extends StatelessWidget {
  const _TeacherDropdownItem({
    required this.teacher,
    required this.selected,
    required this.onTap,
  });

  final _TeacherOption teacher;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        height: 44,
        padding: const EdgeInsets.symmetric(horizontal: 9),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF1E8) : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: <Widget>[
            Container(
              width: 26,
              height: 26,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: selected ? _SmartColors.orange : const Color(0xFFFFF7EE),
                borderRadius: BorderRadius.circular(9),
              ),
              child: Icon(
                selected ? Icons.check_rounded : Icons.person_outline_rounded,
                color: selected ? Colors.white : _SmartColors.text,
                size: 16,
              ),
            ),
            const SizedBox(width: 9),
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    teacher.name,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    teacher.label,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.muted,
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _DateSwitch extends StatelessWidget {
  const _DateSwitch({
    required this.width,
    required this.dateRange,
    required this.onPrev,
    required this.onNext,
  });

  final double width;
  final String dateRange;
  final VoidCallback onPrev;
  final VoidCallback onNext;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      width: width,
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 6),
      child: Row(
        children: <Widget>[
          _MiniNavButton(
            label: '上一周',
            onTap: onPrev,
          ),
          Expanded(
            child: Text(
              dateRange,
              textAlign: TextAlign.center,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _SmartColors.ink,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          _MiniNavButton(
            label: '下一周',
            onTap: onNext,
          ),
        ],
      ),
    );
  }
}

class _MiniNavButton extends StatelessWidget {
  const _MiniNavButton({
    required this.label,
    required this.onTap,
  });

  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        width: 74,
        height: 32,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: const Color(0xFFFFF0E5),
          border: Border.all(color: const Color(0xFFF3D5C4)),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Text(
          label,
          maxLines: 1,
          overflow: TextOverflow.clip,
          style: const TextStyle(
            color: _SmartColors.orangeDeep,
            fontSize: 12,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _ToolbarButton extends StatelessWidget {
  const _ToolbarButton({
    required this.width,
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final double width;
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      width: width,
      height: 42,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Icon(icon, color: _SmartColors.text, size: 16),
            const SizedBox(width: 6),
            Text(
              label,
              style: const TextStyle(
                color: _SmartColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PrimaryButton extends StatelessWidget {
  const _PrimaryButton({required this.width});

  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 42,
      decoration: BoxDecoration(
        color: _SmartColors.orange,
        borderRadius: BorderRadius.circular(13),
      ),
      child: const Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          Icon(Icons.add_rounded, color: Colors.white, size: 18),
          SizedBox(width: 5),
          Text(
            '新增排课',
            style: TextStyle(
              color: Colors.white,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _FilterButton extends StatelessWidget {
  const _FilterButton({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      height: 34,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      borderRadius: 11,
      child: Row(
        children: <Widget>[
          Icon(icon, color: _SmartColors.text, size: 16),
          const SizedBox(width: 7),
          Text(
            label,
            style: const TextStyle(
              color: _SmartColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _TimetableLoadStatus extends StatelessWidget {
  const _TimetableLoadStatus({
    required this.message,
    required this.onRefresh,
  });

  final String message;
  final VoidCallback onRefresh;

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 230),
      child: InkWell(
        onTap: onRefresh,
        borderRadius: BorderRadius.circular(999),
        child: Container(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: const Color(0xFFFFEFEA),
            border: Border.all(color: const Color(0xFFF4C8BB)),
            borderRadius: BorderRadius.circular(999),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(
                Icons.refresh_rounded,
                color: _SmartColors.orangeDeep,
                size: 15,
              ),
              const SizedBox(width: 5),
              Flexible(
                child: Text(
                  message,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _SmartColors.orangeDeep,
                    fontSize: 11,
                    height: 1,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _SummaryAccent extends StatelessWidget {
  const _SummaryAccent();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 4,
      height: 16,
      decoration: BoxDecoration(
        color: _SmartColors.orange,
        borderRadius: BorderRadius.circular(999),
      ),
    );
  }
}

class _LegendItem extends StatelessWidget {
  const _LegendItem({required this.color, required this.label});

  final Color color;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Container(
          width: 16,
          height: 4,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(999),
          ),
        ),
        const SizedBox(width: 6),
        Text(
          label,
          style: const TextStyle(
            color: _SmartColors.text,
            fontSize: 11,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }
}

class _IconShell extends StatelessWidget {
  const _IconShell({required this.size, required this.icon, this.onTap});

  final double size;
  final IconData icon;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      width: size,
      height: size,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Icon(icon, color: _SmartColors.ink, size: 24),
      ),
    );
  }
}

class _ShellBox extends StatelessWidget {
  const _ShellBox({
    required this.child,
    this.width,
    this.height,
    this.padding,
    this.borderRadius = 13,
  });

  final Widget child;
  final double? width;
  final double? height;
  final EdgeInsetsGeometry? padding;
  final double borderRadius;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _SmartColors.card.withOpacity(.92),
      borderRadius: BorderRadius.circular(borderRadius),
      child: Container(
        width: width,
        height: height,
        padding: padding,
        decoration: BoxDecoration(
          border: Border.all(color: _SmartColors.line),
          borderRadius: BorderRadius.circular(borderRadius),
        ),
        child: child,
      ),
    );
  }
}

class _DiagonalPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint line = Paint()
      ..color = _SmartColors.line
      ..strokeWidth = 1;
    canvas.drawLine(Offset.zero, Offset(size.width, size.height), line);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
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
    this.isToday = false,
  });

  final String label;
  final String date;
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
    required this.title,
    required this.person,
    required this.status,
    this.conflict = false,
  });

  final String title;
  final String person;
  final _LessonStatus status;
  final bool conflict;
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
