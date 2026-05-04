import 'dart:async';
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
  int _scheduleOptionsSequence = 0;
  int _availabilitySequence = 0;
  bool _teacherDropdownOpen = false;
  bool _schedulePanelOpen = false;
  bool _scheduleOptionsLoading = false;
  bool _availabilityLoading = false;
  bool _loading = true;
  String? _errorMessage;
  String? _scheduleOptionsError;
  String? _availabilityMessage;
  String _selectedPeriodGroupId = '';
  String _selectedTeacherId = '';
  _ScheduleMode _scheduleMode = _ScheduleMode.oneToOne;
  ScheduleTargetOption? _selectedOneToOneTarget;
  ScheduleTargetOption? _selectedGroupClassTarget;
  ScheduleClassroomOption? _selectedClassroom;
  List<ScheduleTargetOption> _oneToOneTargets = const <ScheduleTargetOption>[];
  List<ScheduleTargetOption> _groupClassTargets =
      const <ScheduleTargetOption>[];
  List<ScheduleStaffOption> _assistantOptions = const <ScheduleStaffOption>[];
  List<ScheduleClassroomOption> _classroomOptions =
      const <ScheduleClassroomOption>[];
  Set<String> _selectedAssistantIds = <String>{};
  Map<String, _SlotAvailability> _slotAvailability =
      const <String, _SlotAvailability>{};
  Map<String, _SlotAvailability> _dragSlotAvailability =
      const <String, _SlotAvailability>{};
  Set<String> _dragCheckingSlotKeys = const <String>{};
  final Map<String, Future<_SlotAvailability>> _dragValidationFutures =
      <String, Future<_SlotAvailability>>{};
  int _dragValidationSequence = 0;
  bool _dragValidationActive = false;
  String? _creatingSlotKey;
  String? _movingScheduleId;
  OverlayEntry? _scheduleMessageEntry;
  Timer? _scheduleMessageTimer;
  bool _scheduleMessageVisible = false;
  String _scheduleMessageText = '';
  _ScheduleMessageTone _scheduleMessageTone = _ScheduleMessageTone.info;
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
    _loadScheduleOptions();
  }

  @override
  void dispose() {
    _scheduleMessageTimer?.cancel();
    _removeScheduleMessage();
    super.dispose();
  }

  ScheduleTargetOption? get _selectedScheduleTarget {
    return _scheduleMode == _ScheduleMode.oneToOne
        ? _selectedOneToOneTarget
        : _selectedGroupClassTarget;
  }

  ScheduleTargetType get _scheduleTargetType {
    return _scheduleMode == _ScheduleMode.oneToOne
        ? ScheduleTargetType.oneToOne
        : ScheduleTargetType.groupClass;
  }

  List<String> get _normalizedAssistantIds {
    final String teacherId = _selectedTeacherId.trim();
    final Set<String> validAssistantIds = _currentGroupAssistantOptions
        .map((ScheduleStaffOption item) => item.id)
        .toSet();
    return _selectedAssistantIds
        .map((String id) => id.trim())
        .where(
          (String id) =>
              id.isNotEmpty &&
              id != teacherId &&
              validAssistantIds.contains(id),
        )
        .toList();
  }

  String get _selectedClassroomId => _selectedClassroom?.id.trim() ?? '';

  List<ScheduleStaffOption> get _currentGroupAssistantOptions {
    final String teacherId = _selectedTeacherId.trim();
    final Map<String, ScheduleStaffOption> staffById =
        <String, ScheduleStaffOption>{
      for (final ScheduleStaffOption staff in _assistantOptions)
        if (staff.id.trim().isNotEmpty) staff.id.trim(): staff,
    };
    final List<ScheduleStaffOption> assistants = <ScheduleStaffOption>[];
    final Set<String> seenIds = <String>{};
    for (final _TeacherOption teacher in _teachers) {
      final String id = teacher.id.trim();
      if (id.isEmpty || id == teacherId || seenIds.contains(id)) {
        continue;
      }
      seenIds.add(id);
      final ScheduleStaffOption? staff = staffById[id];
      assistants.add(
        ScheduleStaffOption(
          id: id,
          name: (staff?.name.trim().isNotEmpty == true)
              ? staff!.name.trim()
              : teacher.name,
          subtitle: (staff?.subtitle.trim().isNotEmpty == true)
              ? staff!.subtitle.trim()
              : '当前时段组老师',
        ),
      );
    }
    return assistants;
  }

  void _pruneSelectedAssistantsToCurrentGroup() {
    final Set<String> validAssistantIds = _currentGroupAssistantOptions
        .map((ScheduleStaffOption item) => item.id)
        .toSet();
    _selectedAssistantIds = _selectedAssistantIds
        .where((String id) => validAssistantIds.contains(id.trim()))
        .toSet();
  }

  Future<String> _readAuthToken() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(_timetableAuthTokenStorageKey) ?? '';
  }

  void _resetAvailabilityFields({bool cancelPending = true}) {
    if (cancelPending) {
      _availabilitySequence += 1;
    }
    _availabilityLoading = false;
    _availabilityMessage = null;
    _slotAvailability = const <String, _SlotAvailability>{};
    _creatingSlotKey = null;
  }

  Future<void> _loadScheduleOptions() async {
    final int sequence = ++_scheduleOptionsSequence;
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      return;
    }
    if (mounted) {
      setState(() {
        _scheduleOptionsLoading = true;
        _scheduleOptionsError = null;
      });
    }
    try {
      final List<dynamic> results = await Future.wait<dynamic>(
        <Future<dynamic>>[
          widget.timetableClient.fetchOneToOneTargets(token),
          widget.timetableClient.fetchGroupClassTargets(token),
          widget.timetableClient.fetchScheduleAssistants(token),
          widget.timetableClient.fetchScheduleClassrooms(token),
        ],
      );
      if (!mounted || sequence != _scheduleOptionsSequence) {
        return;
      }
      final List<ScheduleTargetOption> oneToOneTargets =
          List<ScheduleTargetOption>.from(results[0] as List<dynamic>);
      final List<ScheduleTargetOption> groupClassTargets =
          List<ScheduleTargetOption>.from(results[1] as List<dynamic>);
      final List<ScheduleStaffOption> assistantOptions =
          List<ScheduleStaffOption>.from(results[2] as List<dynamic>);
      final List<ScheduleClassroomOption> classroomOptions =
          List<ScheduleClassroomOption>.from(results[3] as List<dynamic>);
      setState(() {
        _oneToOneTargets = oneToOneTargets;
        _groupClassTargets = groupClassTargets;
        _assistantOptions = assistantOptions;
        _classroomOptions = classroomOptions;
        _selectedOneToOneTarget = _preserveSelectedTarget(
          _selectedOneToOneTarget,
          oneToOneTargets,
        );
        _selectedGroupClassTarget = _preserveSelectedTarget(
          _selectedGroupClassTarget,
          groupClassTargets,
        );
        _pruneSelectedAssistantsToCurrentGroup();
        _selectedClassroom = _preserveSelectedClassroom(
          _selectedClassroom,
          classroomOptions,
        );
        _scheduleOptionsLoading = false;
      });
      if (_selectedScheduleTarget != null) {
        unawaited(_detectScheduleAvailability());
      }
    } on TimetableApiException catch (error) {
      if (!mounted || sequence != _scheduleOptionsSequence) {
        return;
      }
      setState(() {
        _scheduleOptionsLoading = false;
        _scheduleOptionsError = error.message;
      });
    } on Object catch (error) {
      if (!mounted || sequence != _scheduleOptionsSequence) {
        return;
      }
      setState(() {
        _scheduleOptionsLoading = false;
        _scheduleOptionsError = '排课选项加载失败：$error';
      });
    }
  }

  ScheduleTargetOption? _preserveSelectedTarget(
    ScheduleTargetOption? current,
    List<ScheduleTargetOption> options,
  ) {
    if (current == null) {
      return null;
    }
    for (final ScheduleTargetOption option in options) {
      if (option.id == current.id) {
        return option;
      }
    }
    return null;
  }

  ScheduleClassroomOption? _preserveSelectedClassroom(
    ScheduleClassroomOption? current,
    List<ScheduleClassroomOption> options,
  ) {
    if (current == null) {
      return null;
    }
    for (final ScheduleClassroomOption option in options) {
      if (option.id == current.id) {
        return option;
      }
    }
    return null;
  }

  Future<void> _loadTimetable(
      {String? teacherId, String? periodGroupId}) async {
    final int sequence = ++_loadSequence;
    final String token = await _readAuthToken();
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
      if (_selectedScheduleTarget != null) {
        unawaited(_detectScheduleAvailability());
      }
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
    _pruneSelectedAssistantsToCurrentGroup();
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
      _resetAvailabilityFields();
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
      _resetAvailabilityFields();
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
      _resetAvailabilityFields();
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
      _resetAvailabilityFields();
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
    unawaited(_moveLessonToSlot(source, targetRow, targetColumn));
  }

  Future<void> _moveLessonToSlot(
    _LessonDragData source,
    int targetRow,
    int targetColumn,
  ) async {
    if (_movingScheduleId != null) {
      return;
    }
    if (source.row == targetRow && source.column == targetColumn) {
      return;
    }
    if (source.row < 0 ||
        source.column < 0 ||
        source.row >= _scheduleRows.length ||
        source.column >= _scheduleRows[source.row].length ||
        targetRow < 0 ||
        targetColumn < 0 ||
        targetRow >= _scheduleRows.length ||
        targetColumn >= _scheduleRows[targetRow].length) {
      return;
    }

    final _LessonCell? sourceLesson = _scheduleRows[source.row][source.column];
    if (sourceLesson == null) {
      return;
    }
    if (sourceLesson.id.trim().isEmpty) {
      _showScheduleMessage('缺少日程ID，无法调课');
      return;
    }
    if (_scheduleRows[targetRow][targetColumn] != null) {
      _showScheduleMessage('目标时段已有课程，不能调课');
      return;
    }
    final _ScheduleCellSlot? targetSlot =
        _scheduleCellSlotAt(targetRow, targetColumn);
    if (targetSlot == null) {
      _showScheduleMessage('目标时段无效，不能调课');
      return;
    }
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      _showScheduleMessage('登录已失效，请重新登录');
      return;
    }
    final _SlotAvailability availability = await _ensureDragTargetValidation(
      sourceLesson,
      targetSlot,
      token,
    );
    if (!availability.valid) {
      _showScheduleMessage(availability.message);
      return;
    }

    setState(() {
      _movingScheduleId = sourceLesson.id;
    });
    try {
      await widget.timetableClient.updateScheduleSlot(
        token,
        scheduleId: sourceLesson.id,
        teacherId: _selectedTeacherId,
        assistantIds: sourceLesson.assistantIds,
        classroomId: sourceLesson.classroomId,
        slot: ScheduleSlotRequest(
          teacherId: _selectedTeacherId,
          lessonDate: targetSlot.date,
          startTime: targetSlot.startTime,
          endTime: targetSlot.endTime,
          assistantIds: sourceLesson.assistantIds ?? const <String>[],
          classroomId: sourceLesson.classroomId ?? '',
        ),
      );
      _showScheduleMessage(
        '调课成功，已刷新课表',
        tone: _ScheduleMessageTone.success,
      );
      await _loadTimetable();
    } on TimetableApiException catch (error) {
      _showScheduleMessage(error.message);
    } on Object catch (error) {
      _showScheduleMessage('调课失败：$error');
    } finally {
      if (mounted) {
        setState(() {
          _movingScheduleId = null;
        });
      }
    }
  }

  void _startLessonDrag(_LessonDragData source) {
    setState(() {
      _dragValidationActive = true;
      _dragSlotAvailability = const <String, _SlotAvailability>{};
      _dragCheckingSlotKeys = const <String>{};
    });
    unawaited(_primeDragSlotAvailability(source));
  }

  void _endLessonDrag() {
    setState(() {
      _dragValidationActive = false;
      _dragCheckingSlotKeys = const <String>{};
    });
  }

  Future<void> _primeDragSlotAvailability(_LessonDragData source) async {
    final int sequence = ++_dragValidationSequence;
    final _LessonCell? sourceLesson = _lessonAt(source.row, source.column);
    if (sourceLesson == null) {
      return;
    }
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      return;
    }
    final List<_ScheduleCellSlot> cells = _emptyScheduleSlotsForCurrentWeek();
    final Map<String, _SlotAvailability> localResults =
        <String, _SlotAvailability>{};
    final List<_ScheduleCellSlot> remoteCells = <_ScheduleCellSlot>[];
    for (final _ScheduleCellSlot cell in cells) {
      final _SlotAvailability? local = _localDragValidation(sourceLesson, cell);
      if (local != null) {
        localResults[cell.key] = local;
      } else {
        remoteCells.add(cell);
      }
    }
    if (!mounted || sequence != _dragValidationSequence) {
      return;
    }
    setState(() {
      _dragValidationActive = true;
      _dragSlotAvailability = localResults;
      _dragCheckingSlotKeys =
          remoteCells.map((_ScheduleCellSlot cell) => cell.key).toSet();
    });
    if (remoteCells.isEmpty) {
      return;
    }
    try {
      final Map<String, _SlotAvailability> remoteResults =
          await _fetchDragTargetsValidation(sourceLesson, remoteCells, token);
      if (!mounted || sequence != _dragValidationSequence) {
        return;
      }
      setState(() {
        _dragSlotAvailability = <String, _SlotAvailability>{
          ...localResults,
          ...remoteResults,
        };
        _dragCheckingSlotKeys = const <String>{};
      });
    } on Object {
      if (!mounted || sequence != _dragValidationSequence) {
        return;
      }
      setState(() {
        _dragCheckingSlotKeys = const <String>{};
      });
    }
  }

  Future<_SlotAvailability> _ensureDragTargetValidation(
    _LessonCell sourceLesson,
    _ScheduleCellSlot targetSlot,
    String token,
  ) {
    final _SlotAvailability? cached = _dragSlotAvailability[targetSlot.key];
    if (cached != null) {
      return Future<_SlotAvailability>.value(cached);
    }
    final _SlotAvailability? localResult =
        _localDragValidation(sourceLesson, targetSlot);
    if (localResult != null) {
      _setDragAvailability(targetSlot.key, localResult);
      return Future<_SlotAvailability>.value(localResult);
    }

    final String requestKey =
        _dragValidationRequestKey(sourceLesson, targetSlot);
    final Future<_SlotAvailability>? pending =
        _dragValidationFutures[requestKey];
    if (pending != null) {
      return pending;
    }
    if (mounted) {
      setState(() {
        _dragCheckingSlotKeys = <String>{
          ..._dragCheckingSlotKeys,
          targetSlot.key,
        };
      });
    }
    final Future<_SlotAvailability> future =
        _fetchDragTargetValidation(sourceLesson, targetSlot, token);
    _dragValidationFutures[requestKey] = future;
    future.then(
      (_SlotAvailability result) {
        if (!mounted) {
          return;
        }
        setState(() {
          _dragSlotAvailability = <String, _SlotAvailability>{
            ..._dragSlotAvailability,
            targetSlot.key: result,
          };
          final Set<String> next = <String>{..._dragCheckingSlotKeys}
            ..remove(targetSlot.key);
          _dragCheckingSlotKeys = next;
        });
      },
      onError: (_) {
        if (!mounted) {
          return;
        }
        setState(() {
          final Set<String> next = <String>{..._dragCheckingSlotKeys}
            ..remove(targetSlot.key);
          _dragCheckingSlotKeys = next;
        });
      },
    ).whenComplete(() {
      _dragValidationFutures.remove(requestKey);
    });
    return future;
  }

  Future<Map<String, _SlotAvailability>> _fetchDragTargetsValidation(
    _LessonCell sourceLesson,
    List<_ScheduleCellSlot> targetSlots,
    String token,
  ) async {
    if (targetSlots.isEmpty) {
      return const <String, _SlotAvailability>{};
    }
    try {
      final ScheduleValidationResult result =
          await widget.timetableClient.validateScheduleSlots(
        token,
        type: sourceLesson.scheduleTargetType!,
        targetId: sourceLesson.teachingClassId,
        teacherId: _selectedTeacherId,
        assistantIds: sourceLesson.assistantIds ?? const <String>[],
        classroomId: sourceLesson.classroomId ?? '',
        slots: targetSlots.map((_ScheduleCellSlot targetSlot) {
          return ScheduleSlotRequest(
            teacherId: _selectedTeacherId,
            lessonDate: targetSlot.date,
            startTime: targetSlot.startTime,
            endTime: targetSlot.endTime,
            assistantIds: sourceLesson.assistantIds ?? const <String>[],
            classroomId: sourceLesson.classroomId ?? '',
          );
        }).toList(),
        excludeIds: <String>[sourceLesson.id],
      );
      final Map<String, ScheduleValidationItem> itemByKey =
          <String, ScheduleValidationItem>{};
      for (final ScheduleValidationItem item in result.items) {
        itemByKey[_availabilityKey(
          item.teacherId.trim().isEmpty ? _selectedTeacherId : item.teacherId,
          item.lessonDate,
          item.startTime,
          item.endTime,
        )] = item;
      }
      final bool hasItemResult = result.items.isNotEmpty;
      final Map<String, _SlotAvailability> next = <String, _SlotAvailability>{};
      for (final _ScheduleCellSlot cell in targetSlots) {
        final ScheduleValidationItem? item = itemByKey[cell.key];
        final bool valid = hasItemResult ? (item?.valid ?? true) : result.valid;
        final String message = item?.message.trim().isNotEmpty == true
            ? item!.message
            : valid
                ? '可调课'
                : (result.message.trim().isEmpty ? '当前空点不可调课' : result.message);
        next[cell.key] = _SlotAvailability(
          valid: valid,
          message: message,
          conflictTypes: item?.conflictTypes ?? const <String>[],
        );
      }
      return next;
    } on TimetableApiException catch (error) {
      return <String, _SlotAvailability>{
        for (final _ScheduleCellSlot cell in targetSlots)
          cell.key: _SlotAvailability(valid: false, message: error.message),
      };
    } on Object catch (error) {
      return <String, _SlotAvailability>{
        for (final _ScheduleCellSlot cell in targetSlots)
          cell.key: _SlotAvailability(
            valid: false,
            message: '检测调课空点失败：$error',
          ),
      };
    }
  }

  Future<_SlotAvailability> _fetchDragTargetValidation(
    _LessonCell sourceLesson,
    _ScheduleCellSlot targetSlot,
    String token,
  ) async {
    try {
      final ScheduleValidationResult result =
          await widget.timetableClient.validateScheduleSlots(
        token,
        type: sourceLesson.scheduleTargetType!,
        targetId: sourceLesson.teachingClassId,
        teacherId: _selectedTeacherId,
        assistantIds: sourceLesson.assistantIds ?? const <String>[],
        classroomId: sourceLesson.classroomId ?? '',
        slots: <ScheduleSlotRequest>[
          ScheduleSlotRequest(
            teacherId: _selectedTeacherId,
            lessonDate: targetSlot.date,
            startTime: targetSlot.startTime,
            endTime: targetSlot.endTime,
            assistantIds: sourceLesson.assistantIds ?? const <String>[],
            classroomId: sourceLesson.classroomId ?? '',
          ),
        ],
        excludeIds: <String>[sourceLesson.id],
      );
      final ScheduleValidationItem? item =
          result.items.isNotEmpty ? result.items.first : null;
      final bool valid = item?.valid ?? result.valid;
      final String message = valid
          ? '可调课'
          : ((item?.message ?? result.message).trim().isEmpty
              ? '当前空点不可调课'
              : (item?.message ?? result.message).trim());
      return _SlotAvailability(
        valid: valid,
        message: message,
        conflictTypes: item?.conflictTypes ?? const <String>[],
      );
    } on TimetableApiException catch (error) {
      return _SlotAvailability(
        valid: false,
        message: error.message,
      );
    } on Object catch (error) {
      return _SlotAvailability(
        valid: false,
        message: '检测调课空点失败：$error',
      );
    }
  }

  _SlotAvailability? _localDragValidation(
    _LessonCell sourceLesson,
    _ScheduleCellSlot targetSlot,
  ) {
    if (sourceLesson.id.trim().isEmpty) {
      return const _SlotAvailability(valid: false, message: '缺少日程ID，无法调课');
    }
    if (sourceLesson.teachingClassId.trim().isEmpty ||
        sourceLesson.scheduleTargetType == null) {
      return const _SlotAvailability(valid: false, message: '当前课程信息不完整，无法检测');
    }
    if (_isPastIsoDate(targetSlot.date)) {
      return const _SlotAvailability(valid: false, message: '不允许调课到过去日期');
    }
    final _LessonCell? targetLesson =
        _lessonAt(targetSlot.row, targetSlot.column);
    if (targetLesson != null) {
      return const _SlotAvailability(valid: false, message: '目标时段已有课程，不能调课');
    }
    return null;
  }

  void _setDragAvailability(String key, _SlotAvailability value) {
    if (!mounted) {
      return;
    }
    setState(() {
      _dragSlotAvailability = <String, _SlotAvailability>{
        ..._dragSlotAvailability,
        key: value,
      };
    });
  }

  _LessonCell? _lessonAt(int row, int column) {
    if (row < 0 ||
        column < 0 ||
        row >= _scheduleRows.length ||
        column >= _scheduleRows[row].length) {
      return null;
    }
    return _scheduleRows[row][column];
  }

  String _dragValidationRequestKey(
    _LessonCell sourceLesson,
    _ScheduleCellSlot targetSlot,
  ) {
    return <String>[
      sourceLesson.id,
      sourceLesson.teachingClassId,
      '${sourceLesson.classType}',
      _selectedTeacherId,
      targetSlot.date,
      targetSlot.startTime,
      targetSlot.endTime,
      (sourceLesson.assistantIds ?? const <String>[]).join(','),
      sourceLesson.classroomId ?? '',
    ].join('|');
  }

  bool _isPastIsoDate(String isoDate) {
    final DateTime? parsed = DateTime.tryParse(isoDate);
    if (parsed == null) {
      return false;
    }
    final DateTime target = DateTime(parsed.year, parsed.month, parsed.day);
    final DateTime now = DateTime.now();
    final DateTime today = DateTime(now.year, now.month, now.day);
    return target.isBefore(today);
  }

  void _toggleSchedulePanel() {
    setState(() {
      _schedulePanelOpen = !_schedulePanelOpen;
      _teacherDropdownOpen = false;
    });
    if (_schedulePanelOpen &&
        !_scheduleOptionsLoading &&
        _oneToOneTargets.isEmpty &&
        _groupClassTargets.isEmpty) {
      unawaited(_loadScheduleOptions());
    }
  }

  void _closeSchedulePanel() {
    if (!_schedulePanelOpen) {
      return;
    }
    setState(() {
      _schedulePanelOpen = false;
    });
  }

  void _setScheduleMode(_ScheduleMode mode) {
    if (mode == _scheduleMode) {
      return;
    }
    setState(() {
      _scheduleMode = mode;
      _resetAvailabilityFields();
    });
    if (_selectedScheduleTarget != null) {
      unawaited(_detectScheduleAvailability());
    }
  }

  void _selectScheduleTarget(ScheduleTargetOption option) {
    if (option.disabled) {
      _showScheduleMessage('该排课对象当前状态不可排');
      return;
    }
    setState(() {
      if (_scheduleMode == _ScheduleMode.oneToOne) {
        _selectedOneToOneTarget = option;
      } else {
        _selectedGroupClassTarget = option;
      }
      _schedulePanelOpen = false;
      _resetAvailabilityFields();
    });
    unawaited(_detectScheduleAvailability());
  }

  void _clearScheduleTarget() {
    setState(() {
      if (_scheduleMode == _ScheduleMode.oneToOne) {
        _selectedOneToOneTarget = null;
      } else {
        _selectedGroupClassTarget = null;
      }
      _schedulePanelOpen = false;
      _resetAvailabilityFields();
    });
  }

  void _toggleAssistant(ScheduleStaffOption assistant) {
    final String id = assistant.id.trim();
    if (id.isEmpty) {
      return;
    }
    if (id == _selectedTeacherId.trim()) {
      _showScheduleMessage('主教与助教不能为同一人，已自动忽略');
      setState(() {
        _selectedAssistantIds.remove(id);
        _resetAvailabilityFields();
      });
      unawaited(_detectScheduleAvailability());
      return;
    }
    final bool inCurrentGroup = _currentGroupAssistantOptions.any(
      (ScheduleStaffOption item) => item.id == id,
    );
    if (!inCurrentGroup) {
      _showScheduleMessage('助教只能选择当前时段组的其他老师');
      return;
    }
    setState(() {
      final Set<String> next = Set<String>.from(_selectedAssistantIds);
      if (next.contains(id)) {
        next.remove(id);
      } else {
        next.add(id);
      }
      _selectedAssistantIds = next;
      _resetAvailabilityFields();
    });
    unawaited(_detectScheduleAvailability());
  }

  void _selectClassroom(ScheduleClassroomOption? classroom) {
    setState(() {
      _selectedClassroom = classroom;
      _resetAvailabilityFields();
    });
    unawaited(_detectScheduleAvailability());
  }

  Future<void> _detectScheduleAvailability() async {
    final int sequence = ++_availabilitySequence;
    final ScheduleTargetOption? target = _selectedScheduleTarget;
    final String teacherId = _selectedTeacherId.trim();
    if (target == null) {
      if (!mounted) {
        return;
      }
      setState(() {
        _resetAvailabilityFields(cancelPending: false);
      });
      return;
    }
    if (teacherId.isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _availabilityLoading = false;
        _availabilityMessage = '当前课表没有可用老师';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
      return;
    }
    final List<_ScheduleCellSlot> cells = _emptyScheduleSlotsForCurrentWeek();
    if (cells.isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _availabilityLoading = false;
        _availabilityMessage = '本周暂无空闲时段';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
      return;
    }
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      if (!mounted || sequence != _availabilitySequence) {
        return;
      }
      setState(() {
        _availabilityLoading = false;
        _availabilityMessage = '登录已失效，请重新登录';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
      return;
    }
    final List<String> assistantIds = _normalizedAssistantIds;
    final String classroomId = _selectedClassroomId;
    if (mounted) {
      setState(() {
        _availabilityLoading = true;
        _availabilityMessage = '检测中';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
    }
    try {
      final ScheduleValidationResult result =
          await widget.timetableClient.validateScheduleSlots(
        token,
        type: _scheduleTargetType,
        targetId: target.id,
        teacherId: teacherId,
        assistantIds: assistantIds,
        classroomId: classroomId,
        slots: cells.map((_ScheduleCellSlot cell) {
          return ScheduleSlotRequest(
            teacherId: teacherId,
            lessonDate: cell.date,
            startTime: cell.startTime,
            endTime: cell.endTime,
            assistantIds: assistantIds,
            classroomId: classroomId,
          );
        }).toList(),
      );
      if (!mounted || sequence != _availabilitySequence) {
        return;
      }
      final Map<String, ScheduleValidationItem> itemByKey =
          <String, ScheduleValidationItem>{};
      for (final ScheduleValidationItem item in result.items) {
        itemByKey[_availabilityKey(
          item.teacherId.trim().isEmpty ? teacherId : item.teacherId,
          item.lessonDate,
          item.startTime,
          item.endTime,
        )] = item;
      }
      final bool hasItemResult = result.items.isNotEmpty;
      final Map<String, _SlotAvailability> next = <String, _SlotAvailability>{};
      for (final _ScheduleCellSlot cell in cells) {
        final ScheduleValidationItem? item = itemByKey[cell.key];
        final bool valid = hasItemResult ? (item?.valid ?? true) : result.valid;
        final List<String> conflictTypes =
            item?.conflictTypes ?? const <String>[];
        final String message = item?.message.trim().isNotEmpty == true
            ? item!.message
            : valid
                ? '空闲时段可排'
                : (result.message.trim().isEmpty ? '该时间段不可排课' : result.message);
        next[cell.key] = _SlotAvailability(
          valid: valid,
          message: message,
          conflictTypes: conflictTypes,
        );
      }
      final int validCount =
          next.values.where((_SlotAvailability item) => item.valid).length;
      final int invalidCount = next.length - validCount;
      setState(() {
        _availabilityLoading = false;
        _slotAvailability = next;
        _availabilityMessage = invalidCount == 0
            ? '可排 $validCount'
            : '可排 $validCount，冲突 $invalidCount';
      });
    } on TimetableApiException catch (error) {
      if (!mounted || sequence != _availabilitySequence) {
        return;
      }
      setState(() {
        _availabilityLoading = false;
        _availabilityMessage = error.message;
        _slotAvailability = const <String, _SlotAvailability>{};
      });
    } on Object catch (error) {
      if (!mounted || sequence != _availabilitySequence) {
        return;
      }
      setState(() {
        _availabilityLoading = false;
        _availabilityMessage = '空闲点检测失败：$error';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
    }
  }

  Future<void> _handleEmptySlotTap(int row, int column) async {
    final _ScheduleCellSlot? cell = _scheduleCellSlotAt(row, column);
    if (cell == null) {
      return;
    }
    final ScheduleTargetOption? target = _selectedScheduleTarget;
    if (target == null) {
      setState(() {
        _schedulePanelOpen = true;
        _teacherDropdownOpen = false;
      });
      _showScheduleMessage('请先选择要排课的1v1或班课');
      return;
    }
    if (_availabilityLoading) {
      _showScheduleMessage('正在检测本周空闲点，请稍后');
      return;
    }
    final _SlotAvailability? availability = _slotAvailability[cell.key];
    if (availability == null) {
      _showScheduleMessage('请先等待空闲点检测完成');
      unawaited(_detectScheduleAvailability());
      return;
    }
    if (!availability.valid) {
      _showScheduleMessage(availability.message);
      return;
    }
    if (_creatingSlotKey != null) {
      return;
    }
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      _showScheduleMessage('登录已失效，请重新登录');
      return;
    }
    final List<String> assistantIds = _normalizedAssistantIds;
    final String classroomId = _selectedClassroomId;
    setState(() {
      _creatingSlotKey = cell.key;
    });
    try {
      await widget.timetableClient.createSchedule(
        token,
        type: _scheduleTargetType,
        targetId: target.id,
        teacherId: _selectedTeacherId,
        assistantIds: assistantIds,
        classroomId: classroomId,
        slot: ScheduleSlotRequest(
          teacherId: _selectedTeacherId,
          lessonDate: cell.date,
          startTime: cell.startTime,
          endTime: cell.endTime,
          assistantIds: assistantIds,
          classroomId: classroomId,
        ),
      );
      _showScheduleMessage(
        '排课成功，已刷新课表',
        tone: _ScheduleMessageTone.success,
      );
      await _loadTimetable();
    } on TimetableApiException catch (error) {
      _showScheduleMessage(error.message);
      unawaited(_detectScheduleAvailability());
    } on Object catch (error) {
      _showScheduleMessage('创建排课失败：$error');
      unawaited(_detectScheduleAvailability());
    } finally {
      if (mounted) {
        setState(() {
          _creatingSlotKey = null;
        });
      }
    }
  }

  List<_ScheduleCellSlot> _emptyScheduleSlotsForCurrentWeek() {
    final List<_ScheduleCellSlot> cells = <_ScheduleCellSlot>[];
    for (int row = 0; row < _timeSlots.length; row += 1) {
      for (int column = 0; column < _weekDays.length; column += 1) {
        if (row < _scheduleRows.length &&
            column < _scheduleRows[row].length &&
            _scheduleRows[row][column] != null) {
          continue;
        }
        final _ScheduleCellSlot? cell = _scheduleCellSlotAt(row, column);
        if (cell != null) {
          cells.add(cell);
        }
      }
    }
    return cells;
  }

  _ScheduleCellSlot? _scheduleCellSlotAt(int row, int column) {
    if (row < 0 ||
        column < 0 ||
        row >= _timeSlots.length ||
        column >= _weekDays.length) {
      return null;
    }
    final _TimeSlot slot = _timeSlots[row];
    final _WeekDay day = _weekDays[column];
    final String teacherId = _selectedTeacherId.trim();
    if (teacherId.isEmpty ||
        day.isoDate.trim().isEmpty ||
        slot.startTime.trim().isEmpty ||
        slot.endTime.trim().isEmpty) {
      return null;
    }
    return _ScheduleCellSlot(
      row: row,
      column: column,
      key: _availabilityKey(
        teacherId,
        day.isoDate,
        slot.startTime,
        slot.endTime,
      ),
      date: day.isoDate,
      startTime: slot.startTime,
      endTime: slot.endTime,
    );
  }

  void _showScheduleMessage(
    String message, {
    _ScheduleMessageTone tone = _ScheduleMessageTone.info,
  }) {
    if (!mounted || message.trim().isEmpty) {
      return;
    }
    _scheduleMessageTimer?.cancel();
    _scheduleMessageText = message.trim();
    _scheduleMessageTone = tone;
    final OverlayState overlay = Overlay.of(context, rootOverlay: true);
    if (_scheduleMessageEntry == null) {
      _scheduleMessageVisible = false;
      _scheduleMessageEntry = OverlayEntry(
        builder: (BuildContext context) {
          return _ScheduleTopMessage(
            visible: _scheduleMessageVisible,
            message: _scheduleMessageText,
            tone: _scheduleMessageTone,
          );
        },
      );
      overlay.insert(_scheduleMessageEntry!);
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted || _scheduleMessageEntry == null) {
          return;
        }
        _scheduleMessageVisible = true;
        _scheduleMessageEntry?.markNeedsBuild();
      });
    } else {
      _scheduleMessageVisible = true;
      _scheduleMessageEntry!.markNeedsBuild();
    }
    _scheduleMessageTimer = Timer(const Duration(milliseconds: 1900), () {
      _scheduleMessageVisible = false;
      _scheduleMessageEntry?.markNeedsBuild();
      _scheduleMessageTimer = Timer(
        const Duration(milliseconds: 180),
        _removeScheduleMessage,
      );
    });
  }

  void _removeScheduleMessage() {
    _scheduleMessageEntry?.remove();
    _scheduleMessageEntry = null;
  }

  @override
  Widget build(BuildContext context) {
    final _TeacherOption teacher = _teachers.isEmpty
        ? const _TeacherOption(id: '', name: '当前老师', label: '当前老师')
        : _teachers[_teacherIndex.clamp(0, _teachers.length - 1)];
    final List<ScheduleStaffOption> currentGroupAssistantOptions =
        _currentGroupAssistantOptions;
    final Set<String> currentGroupSelectedAssistantIds = _selectedAssistantIds
        .where(
          (String id) => currentGroupAssistantOptions.any(
            (ScheduleStaffOption item) => item.id == id,
          ),
        )
        .toSet();
    return Scaffold(
      body: _SmartTimetableViewport(
        child: _SmartTimetableScreen(
          teacher: teacher,
          periodGroups: _periodGroups,
          periodGroupIndex: _periodGroupIndex,
          teachers: _teachers,
          teacherIndex: _teacherIndex,
          teacherDropdownOpen: _teacherDropdownOpen,
          schedulePanelOpen: _schedulePanelOpen,
          scheduleMode: _scheduleMode,
          selectedScheduleTarget: _selectedScheduleTarget,
          oneToOneTargets: _oneToOneTargets,
          groupClassTargets: _groupClassTargets,
          assistantOptions: currentGroupAssistantOptions,
          classroomOptions: _classroomOptions,
          selectedAssistantIds: currentGroupSelectedAssistantIds,
          selectedClassroom: _selectedClassroom,
          scheduleOptionsLoading: _scheduleOptionsLoading,
          scheduleOptionsError: _scheduleOptionsError,
          availabilityLoading: _availabilityLoading,
          availabilityMessage: _availabilityMessage,
          slotAvailability: _slotAvailability,
          dragSlotAvailability: _dragSlotAvailability,
          dragCheckingSlotKeys: _dragCheckingSlotKeys,
          dragValidationActive: _dragValidationActive,
          creatingSlotKey: _creatingSlotKey,
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
          onTeacherToggle: () => setState(() {
            _teacherDropdownOpen = !_teacherDropdownOpen;
            _schedulePanelOpen = false;
          }),
          onTeacherSelected: _selectTeacher,
          onTeacherDropdownClose: () =>
              setState(() => _teacherDropdownOpen = false),
          onSchedulePanelToggle: _toggleSchedulePanel,
          onSchedulePanelClose: _closeSchedulePanel,
          onScheduleModeChanged: _setScheduleMode,
          onScheduleTargetSelected: _selectScheduleTarget,
          onScheduleTargetCleared: _clearScheduleTarget,
          onAssistantToggled: _toggleAssistant,
          onClassroomSelected: _selectClassroom,
          onScheduleOptionsRefresh: _loadScheduleOptions,
          onAvailabilityRefresh: _detectScheduleAvailability,
          onRefresh: _loadTimetable,
          onLessonMove: _moveLesson,
          onLessonDragStarted: _startLessonDrag,
          onLessonDragEnded: _endLessonDrag,
          onEmptySlotTap: _handleEmptySlotTap,
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
    required this.schedulePanelOpen,
    required this.scheduleMode,
    required this.selectedScheduleTarget,
    required this.oneToOneTargets,
    required this.groupClassTargets,
    required this.assistantOptions,
    required this.classroomOptions,
    required this.selectedAssistantIds,
    required this.selectedClassroom,
    required this.scheduleOptionsLoading,
    required this.scheduleOptionsError,
    required this.availabilityLoading,
    required this.availabilityMessage,
    required this.slotAvailability,
    required this.dragSlotAvailability,
    required this.dragCheckingSlotKeys,
    required this.dragValidationActive,
    required this.creatingSlotKey,
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
    required this.onSchedulePanelToggle,
    required this.onSchedulePanelClose,
    required this.onScheduleModeChanged,
    required this.onScheduleTargetSelected,
    required this.onScheduleTargetCleared,
    required this.onAssistantToggled,
    required this.onClassroomSelected,
    required this.onScheduleOptionsRefresh,
    required this.onAvailabilityRefresh,
    required this.onRefresh,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final _TeacherOption teacher;
  final List<_PeriodGroupOption> periodGroups;
  final int periodGroupIndex;
  final List<_TeacherOption> teachers;
  final int teacherIndex;
  final bool teacherDropdownOpen;
  final bool schedulePanelOpen;
  final _ScheduleMode scheduleMode;
  final ScheduleTargetOption? selectedScheduleTarget;
  final List<ScheduleTargetOption> oneToOneTargets;
  final List<ScheduleTargetOption> groupClassTargets;
  final List<ScheduleStaffOption> assistantOptions;
  final List<ScheduleClassroomOption> classroomOptions;
  final Set<String> selectedAssistantIds;
  final ScheduleClassroomOption? selectedClassroom;
  final bool scheduleOptionsLoading;
  final String? scheduleOptionsError;
  final bool availabilityLoading;
  final String? availabilityMessage;
  final Map<String, _SlotAvailability> slotAvailability;
  final Map<String, _SlotAvailability> dragSlotAvailability;
  final Set<String> dragCheckingSlotKeys;
  final bool dragValidationActive;
  final String? creatingSlotKey;
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
  final VoidCallback onSchedulePanelToggle;
  final VoidCallback onSchedulePanelClose;
  final ValueChanged<_ScheduleMode> onScheduleModeChanged;
  final ValueChanged<ScheduleTargetOption> onScheduleTargetSelected;
  final VoidCallback onScheduleTargetCleared;
  final ValueChanged<ScheduleStaffOption> onAssistantToggled;
  final ValueChanged<ScheduleClassroomOption?> onClassroomSelected;
  final VoidCallback onScheduleOptionsRefresh;
  final VoidCallback onAvailabilityRefresh;
  final VoidCallback onRefresh;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

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
                      onSchedulePanelToggle: onSchedulePanelToggle,
                    ),
                    const SizedBox(height: 10),
                    _TimetableSubBar(
                      compact: compact,
                      scheduleMode: scheduleMode,
                      selectedScheduleTarget: selectedScheduleTarget,
                      schedulePanelOpen: schedulePanelOpen,
                      availabilityLoading: availabilityLoading,
                      availabilityMessage: availabilityMessage,
                      periodGroups: periodGroups,
                      periodGroupIndex: periodGroupIndex,
                      errorMessage: errorMessage,
                      onSchedulePanelToggle: onSchedulePanelToggle,
                      onScheduleModeChanged: onScheduleModeChanged,
                      onScheduleTargetCleared: onScheduleTargetCleared,
                      onAvailabilityRefresh: onAvailabilityRefresh,
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
                        selectedTeacherId: teacher.id,
                        scheduleTargetSelected: selectedScheduleTarget != null,
                        availabilityLoading: availabilityLoading,
                        slotAvailability: slotAvailability,
                        dragSlotAvailability: dragSlotAvailability,
                        dragCheckingSlotKeys: dragCheckingSlotKeys,
                        dragValidationActive: dragValidationActive,
                        creatingSlotKey: creatingSlotKey,
                        onLessonMove: onLessonMove,
                        onLessonDragStarted: onLessonDragStarted,
                        onLessonDragEnded: onLessonDragEnded,
                        onEmptySlotTap: onEmptySlotTap,
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
              if (schedulePanelOpen) ...<Widget>[
                Positioned.fill(
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onTap: onSchedulePanelClose,
                  ),
                ),
                Positioned(
                  top: 134,
                  left: pagePadding,
                  width: compact ? 690 : 780,
                  child: _SchedulePickerPanel(
                    mode: scheduleMode,
                    oneToOneTargets: oneToOneTargets,
                    groupClassTargets: groupClassTargets,
                    assistantOptions: assistantOptions,
                    classroomOptions: classroomOptions,
                    selectedTarget: selectedScheduleTarget,
                    selectedAssistantIds: selectedAssistantIds,
                    selectedClassroom: selectedClassroom,
                    loading: scheduleOptionsLoading,
                    errorMessage: scheduleOptionsError,
                    onModeChanged: onScheduleModeChanged,
                    onTargetSelected: onScheduleTargetSelected,
                    onAssistantToggled: onAssistantToggled,
                    onClassroomSelected: onClassroomSelected,
                    onRefresh: onScheduleOptionsRefresh,
                    onClose: onSchedulePanelClose,
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
    required this.onSchedulePanelToggle,
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
  final VoidCallback onSchedulePanelToggle;

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
                _PrimaryButton(
                  width: primaryWidth,
                  onTap: onSchedulePanelToggle,
                ),
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
    required this.scheduleMode,
    required this.selectedScheduleTarget,
    required this.schedulePanelOpen,
    required this.availabilityLoading,
    required this.availabilityMessage,
    required this.periodGroups,
    required this.periodGroupIndex,
    required this.errorMessage,
    required this.onSchedulePanelToggle,
    required this.onScheduleModeChanged,
    required this.onScheduleTargetCleared,
    required this.onAvailabilityRefresh,
    required this.onPeriodGroupSelected,
    required this.onRefresh,
  });

  final bool compact;
  final _ScheduleMode scheduleMode;
  final ScheduleTargetOption? selectedScheduleTarget;
  final bool schedulePanelOpen;
  final bool availabilityLoading;
  final String? availabilityMessage;
  final List<_PeriodGroupOption> periodGroups;
  final int periodGroupIndex;
  final String? errorMessage;
  final VoidCallback onSchedulePanelToggle;
  final ValueChanged<_ScheduleMode> onScheduleModeChanged;
  final VoidCallback onScheduleTargetCleared;
  final VoidCallback onAvailabilityRefresh;
  final ValueChanged<int> onPeriodGroupSelected;
  final VoidCallback onRefresh;

  @override
  Widget build(BuildContext context) {
    final double periodGroupWidth = compact ? 250 : 284;
    return SizedBox(
      height: 44,
      child: Row(
        children: <Widget>[
          Expanded(
            child: _ScheduleComposerBar(
              compact: compact,
              mode: scheduleMode,
              selectedTarget: selectedScheduleTarget,
              panelOpen: schedulePanelOpen,
              availabilityLoading: availabilityLoading,
              availabilityMessage: availabilityMessage,
              onPanelToggle: onSchedulePanelToggle,
              onModeChanged: onScheduleModeChanged,
              onTargetCleared: onScheduleTargetCleared,
              onAvailabilityRefresh: onAvailabilityRefresh,
            ),
          ),
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
    required this.selectedTeacherId,
    required this.scheduleTargetSelected,
    required this.availabilityLoading,
    required this.slotAvailability,
    required this.dragSlotAvailability,
    required this.dragCheckingSlotKeys,
    required this.dragValidationActive,
    required this.creatingSlotKey,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final bool compact;
  final List<List<_LessonCell?>> rows;
  final List<_WeekDay> weekDays;
  final List<_TimeSlot> timeSlots;
  final String selectedTeacherId;
  final bool scheduleTargetSelected;
  final bool availabilityLoading;
  final Map<String, _SlotAvailability> slotAvailability;
  final Map<String, _SlotAvailability> dragSlotAvailability;
  final Set<String> dragCheckingSlotKeys;
  final bool dragValidationActive;
  final String? creatingSlotKey;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

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
                          weekDays: weekDays,
                          timeSlots: timeSlots,
                          selectedTeacherId: selectedTeacherId,
                          scheduleTargetSelected: scheduleTargetSelected,
                          availabilityLoading: availabilityLoading,
                          slotAvailability: slotAvailability,
                          dragSlotAvailability: dragSlotAvailability,
                          dragCheckingSlotKeys: dragCheckingSlotKeys,
                          dragValidationActive: dragValidationActive,
                          creatingSlotKey: creatingSlotKey,
                          onLessonMove: onLessonMove,
                          onLessonDragStarted: onLessonDragStarted,
                          onLessonDragEnded: onLessonDragEnded,
                          onEmptySlotTap: onEmptySlotTap,
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
  const _ScheduleGrid({
    required this.rows,
    required this.weekDays,
    required this.timeSlots,
    required this.selectedTeacherId,
    required this.scheduleTargetSelected,
    required this.availabilityLoading,
    required this.slotAvailability,
    required this.dragSlotAvailability,
    required this.dragCheckingSlotKeys,
    required this.dragValidationActive,
    required this.creatingSlotKey,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final List<List<_LessonCell?>> rows;
  final List<_WeekDay> weekDays;
  final List<_TimeSlot> timeSlots;
  final String selectedTeacherId;
  final bool scheduleTargetSelected;
  final bool availabilityLoading;
  final Map<String, _SlotAvailability> slotAvailability;
  final Map<String, _SlotAvailability> dragSlotAvailability;
  final Set<String> dragCheckingSlotKeys;
  final bool dragValidationActive;
  final String? creatingSlotKey;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

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
              weekDays: weekDays,
              timeSlot: row < timeSlots.length ? timeSlots[row] : null,
              selectedTeacherId: selectedTeacherId,
              scheduleTargetSelected: scheduleTargetSelected,
              availabilityLoading: availabilityLoading,
              slotAvailability: slotAvailability,
              dragSlotAvailability: dragSlotAvailability,
              dragCheckingSlotKeys: dragCheckingSlotKeys,
              dragValidationActive: dragValidationActive,
              creatingSlotKey: creatingSlotKey,
              isLastRow: row == rows.length - 1,
              onLessonMove: onLessonMove,
              onLessonDragStarted: onLessonDragStarted,
              onLessonDragEnded: onLessonDragEnded,
              onEmptySlotTap: onEmptySlotTap,
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
    required this.weekDays,
    required this.timeSlot,
    required this.selectedTeacherId,
    required this.scheduleTargetSelected,
    required this.availabilityLoading,
    required this.slotAvailability,
    required this.dragSlotAvailability,
    required this.dragCheckingSlotKeys,
    required this.dragValidationActive,
    required this.creatingSlotKey,
    required this.isLastRow,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final int rowIndex;
  final List<_LessonCell?> row;
  final List<_WeekDay> weekDays;
  final _TimeSlot? timeSlot;
  final String selectedTeacherId;
  final bool scheduleTargetSelected;
  final bool availabilityLoading;
  final Map<String, _SlotAvailability> slotAvailability;
  final Map<String, _SlotAvailability> dragSlotAvailability;
  final Set<String> dragCheckingSlotKeys;
  final bool dragValidationActive;
  final String? creatingSlotKey;
  final bool isLastRow;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _rowHeight,
      child: Row(
        children: <Widget>[
          for (int column = 0; column < row.length; column += 1)
            Expanded(
              child: Builder(
                builder: (BuildContext context) {
                  final String availabilityKey =
                      _availabilityKeyForGridCell(column);
                  return _ScheduleGridCell(
                    lesson: row[column],
                    rowIndex: rowIndex,
                    columnIndex: column,
                    availability: slotAvailability[availabilityKey],
                    dragAvailability: dragSlotAvailability[availabilityKey],
                    dragChecking:
                        dragCheckingSlotKeys.contains(availabilityKey),
                    dragValidationActive: dragValidationActive,
                    scheduleTargetSelected: scheduleTargetSelected,
                    availabilityLoading: availabilityLoading,
                    creating: creatingSlotKey == availabilityKey,
                    isToday: column < weekDays.length
                        ? weekDays[column].isToday
                        : false,
                    isLastColumn: column == row.length - 1,
                    isLastRow: isLastRow,
                    onLessonMove: onLessonMove,
                    onLessonDragStarted: onLessonDragStarted,
                    onLessonDragEnded: onLessonDragEnded,
                    onEmptySlotTap: onEmptySlotTap,
                  );
                },
              ),
            ),
        ],
      ),
    );
  }

  String _availabilityKeyForGridCell(int column) {
    if (timeSlot == null ||
        column >= weekDays.length ||
        selectedTeacherId.trim().isEmpty) {
      return '';
    }
    return _availabilityKey(
      selectedTeacherId,
      weekDays[column].isoDate,
      timeSlot!.startTime,
      timeSlot!.endTime,
    );
  }
}

class _ScheduleGridCell extends StatelessWidget {
  const _ScheduleGridCell({
    required this.lesson,
    required this.rowIndex,
    required this.columnIndex,
    required this.availability,
    required this.dragAvailability,
    required this.dragChecking,
    required this.dragValidationActive,
    required this.scheduleTargetSelected,
    required this.availabilityLoading,
    required this.creating,
    required this.isToday,
    required this.isLastColumn,
    required this.isLastRow,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final _LessonCell? lesson;
  final int rowIndex;
  final int columnIndex;
  final _SlotAvailability? availability;
  final _SlotAvailability? dragAvailability;
  final bool dragChecking;
  final bool dragValidationActive;
  final bool scheduleTargetSelected;
  final bool availabilityLoading;
  final bool creating;
  final bool isToday;
  final bool isLastColumn;
  final bool isLastRow;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

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
                _EmptyScheduleSlot(
                  rowIndex: rowIndex,
                  columnIndex: columnIndex,
                  availability: availability,
                  dragAvailability: dragAvailability,
                  dragChecking: dragChecking,
                  dragValidationActive: dragValidationActive,
                  scheduleTargetSelected: scheduleTargetSelected,
                  checking: availabilityLoading && availability == null,
                  creating: creating,
                  onTap: () => onEmptySlotTap(rowIndex, columnIndex),
                )
              else
                _DraggableLessonBlock(
                  lesson: lesson!,
                  source: _LessonDragData(row: rowIndex, column: columnIndex),
                  onDragStarted: onLessonDragStarted,
                  onDragEnded: onLessonDragEnded,
                ),
            ],
          ),
        );
      },
    );
  }
}

class _EmptyScheduleSlot extends StatelessWidget {
  const _EmptyScheduleSlot({
    required this.rowIndex,
    required this.columnIndex,
    required this.availability,
    required this.dragAvailability,
    required this.dragChecking,
    required this.dragValidationActive,
    required this.scheduleTargetSelected,
    required this.checking,
    required this.creating,
    required this.onTap,
  });

  final int rowIndex;
  final int columnIndex;
  final _SlotAvailability? availability;
  final _SlotAvailability? dragAvailability;
  final bool dragChecking;
  final bool dragValidationActive;
  final bool scheduleTargetSelected;
  final bool checking;
  final bool creating;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final bool showingDragState = dragValidationActive;
    if (!scheduleTargetSelected && !showingDragState) {
      return const SizedBox.expand();
    }
    final _SlotAvailability? state =
        showingDragState ? dragAvailability : availability;
    final bool checkingNow =
        showingDragState ? dragChecking && state == null : checking;
    final bool valid = state?.valid == true;
    final bool invalid = state?.valid == false;
    final Color background = creating
        ? const Color(0xFFFFF1E8)
        : valid
            ? const Color(0xFFEAF8EC)
            : invalid
                ? const Color(0xFFFFECEC)
                : const Color(0xFFFFF8EE);
    final Color border = valid
        ? const Color(0xFFCDEDD1)
        : invalid
            ? const Color(0xFFF3B7B7)
            : const Color(0xFFF0DDC9);
    final Color foreground = valid
        ? const Color(0xFF4B9A61)
        : invalid
            ? _SmartColors.danger
            : _SmartColors.orangeDeep;
    final String label = creating
        ? '排课中'
        : valid
            ? (showingDragState ? '可调课' : '空闲时段(可排)')
            : invalid
                ? _invalidLabel(state!, drag: showingDragState)
                : checkingNow
                    ? '检测中'
                    : '待检测';

    return Material(
      key: ValueKey<String>('empty-slot-$rowIndex-$columnIndex'),
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: background,
            border: Border.all(color: border),
            borderRadius: BorderRadius.circular(9),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              if (creating || checkingNow) ...<Widget>[
                SizedBox(
                  width: 11,
                  height: 11,
                  child: CircularProgressIndicator(
                    strokeWidth: 1.5,
                    color: foreground,
                  ),
                ),
                const SizedBox(width: 6),
              ],
              Flexible(
                child: Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: foreground,
                    fontSize: valid ? 12 : 11,
                    height: 1,
                    fontWeight: valid ? FontWeight.w800 : FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _invalidLabel(_SlotAvailability state, {required bool drag}) {
    if (drag) {
      if (state.conflictTypes.isEmpty) {
        return '不可调';
      }
      return '${state.conflictTypes.join('/')}冲突';
    }
    if (state.conflictTypes.isEmpty) {
      return '冲突(不可排)';
    }
    return '${state.conflictTypes.join('/')}冲突';
  }
}

class _DraggableLessonBlock extends StatefulWidget {
  const _DraggableLessonBlock({
    required this.lesson,
    required this.source,
    required this.onDragStarted,
    required this.onDragEnded,
  });

  final _LessonCell lesson;
  final _LessonDragData source;
  final ValueChanged<_LessonDragData> onDragStarted;
  final VoidCallback onDragEnded;

  @override
  State<_DraggableLessonBlock> createState() => _DraggableLessonBlockState();
}

class _DraggableLessonBlockState extends State<_DraggableLessonBlock> {
  Timer? _armingTimer;
  bool _arming = false;
  bool _dragging = false;
  Offset? _downPosition;

  @override
  void dispose() {
    _armingTimer?.cancel();
    super.dispose();
  }

  void _handlePointerDown(PointerDownEvent event) {
    _armingTimer?.cancel();
    _downPosition = event.position;
    _armingTimer = Timer(const Duration(milliseconds: 150), () {
      if (!mounted || _dragging) {
        return;
      }
      setState(() {
        _arming = true;
      });
      unawaited(HapticFeedback.selectionClick());
    });
  }

  void _handlePointerMove(PointerMoveEvent event) {
    final Offset? down = _downPosition;
    if (down == null || _dragging) {
      return;
    }
    if ((event.position - down).distance > 12) {
      _clearArming();
    }
  }

  void _clearArming() {
    _armingTimer?.cancel();
    _armingTimer = null;
    _downPosition = null;
    if (_arming && mounted) {
      setState(() {
        _arming = false;
      });
    }
  }

  void _handleDragStarted() {
    _armingTimer?.cancel();
    _downPosition = null;
    if (mounted) {
      setState(() {
        _arming = false;
        _dragging = true;
      });
    }
    widget.onDragStarted(widget.source);
    unawaited(HapticFeedback.mediumImpact());
  }

  void _handleDragEnded() {
    if (mounted) {
      setState(() {
        _arming = false;
        _dragging = false;
      });
    }
    widget.onDragEnded();
  }

  @override
  Widget build(BuildContext context) {
    final Widget child = _LessonBlock(
      widget.lesson,
      elevated: _arming || _dragging,
      armed: _arming,
    );
    return Listener(
      onPointerDown: _handlePointerDown,
      onPointerMove: _handlePointerMove,
      onPointerUp: (_) => _clearArming(),
      onPointerCancel: (_) => _clearArming(),
      child: LongPressDraggable<_LessonDragData>(
        key: ValueKey<String>(
          'lesson-${widget.source.row}-${widget.source.column}',
        ),
        data: widget.source,
        delay: const Duration(milliseconds: 300),
        hapticFeedbackOnStart: true,
        maxSimultaneousDrags: 1,
        dragAnchorStrategy: pointerDragAnchorStrategy,
        feedback: Material(
          color: Colors.transparent,
          child: Transform.translate(
            offset: const Offset(-78, -25),
            child: SizedBox(
              width: 156,
              height: 50,
              child:
                  _LessonBlock(widget.lesson, elevated: true, dragging: true),
            ),
          ),
        ),
        childWhenDragging: Opacity(
          opacity: .28,
          child: _LessonBlock(widget.lesson),
        ),
        onDragStarted: _handleDragStarted,
        onDragCompleted: _handleDragEnded,
        onDraggableCanceled: (_, __) => _handleDragEnded(),
        onDragEnd: (_) => _handleDragEnded(),
        child: child,
      ),
    );
  }
}

class _LessonBlock extends StatelessWidget {
  const _LessonBlock(
    this.lesson, {
    this.elevated = false,
    this.armed = false,
    this.dragging = false,
  });

  final _LessonCell lesson;
  final bool elevated;
  final bool armed;
  final bool dragging;

  @override
  Widget build(BuildContext context) {
    return AnimatedContainer(
      duration: const Duration(milliseconds: 140),
      height: 50,
      padding: const EdgeInsets.fromLTRB(9, 7, 9, 6),
      decoration: BoxDecoration(
        color: lesson.status.background,
        border: Border.all(
          color:
              armed || dragging ? _SmartColors.orangeDeep : Colors.transparent,
          width: armed || dragging ? 1.4 : 1,
        ),
        borderRadius: BorderRadius.circular(9),
        boxShadow: elevated
            ? const <BoxShadow>[
                BoxShadow(
                  color: Color(0x33B05F32),
                  blurRadius: 16,
                  offset: Offset(0, 8),
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
  const _PrimaryButton({required this.width, required this.onTap});

  final double width;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _SmartColors.orange,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: SizedBox(
          width: width,
          height: 42,
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
        ),
      ),
    );
  }
}

class _ScheduleComposerBar extends StatelessWidget {
  const _ScheduleComposerBar({
    required this.compact,
    required this.mode,
    required this.selectedTarget,
    required this.panelOpen,
    required this.availabilityLoading,
    required this.availabilityMessage,
    required this.onPanelToggle,
    required this.onModeChanged,
    required this.onTargetCleared,
    required this.onAvailabilityRefresh,
  });

  final bool compact;
  final _ScheduleMode mode;
  final ScheduleTargetOption? selectedTarget;
  final bool panelOpen;
  final bool availabilityLoading;
  final String? availabilityMessage;
  final VoidCallback onPanelToggle;
  final ValueChanged<_ScheduleMode> onModeChanged;
  final VoidCallback onTargetCleared;
  final VoidCallback onAvailabilityRefresh;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        _ScheduleModeSwitch(
          width: compact ? 96 : 104,
          mode: mode,
          onChanged: onModeChanged,
        ),
        SizedBox(width: compact ? 7 : 8),
        SizedBox(
          width: compact ? 184 : 230,
          child: _ScheduleTargetSelector(
            mode: mode,
            target: selectedTarget,
            open: panelOpen,
            onTap: onPanelToggle,
            onClear: onTargetCleared,
          ),
        ),
        if (!compact) ...<Widget>[
          const SizedBox(width: 8),
          SizedBox(
            width: 168,
            child: selectedTarget == null &&
                    availabilityMessage == null &&
                    !availabilityLoading
                ? const SizedBox.shrink()
                : _AvailabilityStatusPill(
                    loading: availabilityLoading,
                    message: availabilityMessage ?? '检测中',
                    onTap: onAvailabilityRefresh,
                  ),
          ),
        ],
      ],
    );
  }
}

class _ScheduleModeSwitch extends StatelessWidget {
  const _ScheduleModeSwitch({
    required this.width,
    required this.mode,
    required this.onChanged,
  });

  final double width;
  final _ScheduleMode mode;
  final ValueChanged<_ScheduleMode> onChanged;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      width: width,
      height: 34,
      padding: const EdgeInsets.all(3),
      borderRadius: 11,
      child: Row(
        children: <Widget>[
          Expanded(
            child: _ScheduleModeItem(
              label: '1v1',
              selected: mode == _ScheduleMode.oneToOne,
              onTap: () => onChanged(_ScheduleMode.oneToOne),
            ),
          ),
          Expanded(
            child: _ScheduleModeItem(
              label: '班课',
              selected: mode == _ScheduleMode.groupClass,
              onTap: () => onChanged(_ScheduleMode.groupClass),
            ),
          ),
        ],
      ),
    );
  }
}

class _ScheduleModeItem extends StatelessWidget {
  const _ScheduleModeItem({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 160),
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: selected ? _SmartColors.orange : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: selected ? Colors.white : _SmartColors.text,
            fontSize: 12,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _ScheduleTargetSelector extends StatelessWidget {
  const _ScheduleTargetSelector({
    required this.mode,
    required this.target,
    required this.open,
    required this.onTap,
    required this.onClear,
  });

  final _ScheduleMode mode;
  final ScheduleTargetOption? target;
  final bool open;
  final VoidCallback onTap;
  final VoidCallback onClear;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      key: const ValueKey<String>('schedule-target-selector'),
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: _ShellBox(
        height: 34,
        padding: const EdgeInsets.symmetric(horizontal: 10),
        borderRadius: 11,
        child: Row(
          children: <Widget>[
            Icon(
              mode == _ScheduleMode.oneToOne
                  ? Icons.person_add_alt_1_rounded
                  : Icons.groups_2_outlined,
              color: _SmartColors.text,
              size: 16,
            ),
            const SizedBox(width: 7),
            Expanded(
              child: Text(
                target?.title ??
                    (mode == _ScheduleMode.oneToOne ? '选择1v1' : '选择班课'),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color:
                      target == null ? _SmartColors.muted : _SmartColors.text,
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            if (target != null) ...<Widget>[
              const SizedBox(width: 4),
              InkWell(
                key: const ValueKey<String>('schedule-target-clear'),
                onTap: onClear,
                borderRadius: BorderRadius.circular(8),
                child: Container(
                  width: 22,
                  height: 22,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFF7EE),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(
                    Icons.close_rounded,
                    color: _SmartColors.muted,
                    size: 15,
                  ),
                ),
              ),
              const SizedBox(width: 4),
            ],
            AnimatedRotation(
              turns: open ? .5 : 0,
              duration: const Duration(milliseconds: 160),
              child: const Icon(
                Icons.keyboard_arrow_down_rounded,
                color: _SmartColors.text,
                size: 18,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AvailabilityStatusPill extends StatelessWidget {
  const _AvailabilityStatusPill({
    required this.loading,
    required this.message,
    required this.onTap,
  });

  final bool loading;
  final String message;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      child: InkWell(
        onTap: loading ? null : onTap,
        borderRadius: BorderRadius.circular(999),
        child: Container(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: loading ? const Color(0xFFFFF8EE) : const Color(0xFFEAF8E9),
            border: Border.all(
              color:
                  loading ? const Color(0xFFF0DDC9) : const Color(0xFFC9EACB),
            ),
            borderRadius: BorderRadius.circular(999),
          ),
          child: Row(
            children: <Widget>[
              if (loading)
                const SizedBox(
                  width: 13,
                  height: 13,
                  child: CircularProgressIndicator(
                    strokeWidth: 1.6,
                    color: _SmartColors.orangeDeep,
                  ),
                )
              else
                const Icon(
                  Icons.verified_outlined,
                  color: _SmartColors.green,
                  size: 15,
                ),
              const SizedBox(width: 6),
              Expanded(
                child: _AvailabilityStatusText(
                  loading: loading,
                  message: message,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ScheduleTopMessage extends StatelessWidget {
  const _ScheduleTopMessage({
    required this.visible,
    required this.message,
    required this.tone,
  });

  final bool visible;
  final String message;
  final _ScheduleMessageTone tone;

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: SafeArea(
        bottom: false,
        child: Align(
          alignment: Alignment.topCenter,
          child: AnimatedSlide(
            offset: visible ? Offset.zero : const Offset(0, -.75),
            duration: const Duration(milliseconds: 180),
            curve: Curves.easeOutCubic,
            child: AnimatedOpacity(
              opacity: visible ? 1 : 0,
              duration: const Duration(milliseconds: 160),
              curve: Curves.easeOutCubic,
              child: Container(
                key: const ValueKey<String>('schedule-top-message'),
                constraints: const BoxConstraints(maxWidth: 430),
                margin: const EdgeInsets.only(top: 12),
                padding: const EdgeInsets.fromLTRB(14, 10, 16, 10),
                decoration: BoxDecoration(
                  color: tone.background,
                  border: Border.all(color: tone.border),
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: const <BoxShadow>[
                    BoxShadow(
                      color: Color(0x1A4A2F22),
                      blurRadius: 18,
                      offset: Offset(0, 10),
                    ),
                  ],
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: <Widget>[
                    Icon(tone.icon, color: tone.foreground, size: 18),
                    const SizedBox(width: 8),
                    Flexible(
                      child: Text(
                        message,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          color: tone.textColor,
                          fontSize: 13,
                          height: 1.1,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _AvailabilityStatusText extends StatelessWidget {
  const _AvailabilityStatusText({
    required this.loading,
    required this.message,
  });

  final bool loading;
  final String message;

  @override
  Widget build(BuildContext context) {
    final TextStyle baseStyle = TextStyle(
      color: loading ? _SmartColors.orangeDeep : _SmartColors.green,
      fontSize: 11,
      fontWeight: FontWeight.w900,
    );
    if (loading) {
      return Text(
        message,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: baseStyle,
      );
    }

    final int conflictIndex = message.indexOf('，冲突');
    if (conflictIndex < 0) {
      return Text(
        message,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: baseStyle,
      );
    }

    return Text.rich(
      TextSpan(
        children: <InlineSpan>[
          TextSpan(text: message.substring(0, conflictIndex + 1)),
          TextSpan(
            text: message.substring(conflictIndex + 1),
            style: baseStyle.copyWith(color: _SmartColors.danger),
          ),
        ],
      ),
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: baseStyle,
    );
  }
}

class _SchedulePickerPanel extends StatelessWidget {
  const _SchedulePickerPanel({
    required this.mode,
    required this.oneToOneTargets,
    required this.groupClassTargets,
    required this.assistantOptions,
    required this.classroomOptions,
    required this.selectedTarget,
    required this.selectedAssistantIds,
    required this.selectedClassroom,
    required this.loading,
    required this.errorMessage,
    required this.onModeChanged,
    required this.onTargetSelected,
    required this.onAssistantToggled,
    required this.onClassroomSelected,
    required this.onRefresh,
    required this.onClose,
  });

  final _ScheduleMode mode;
  final List<ScheduleTargetOption> oneToOneTargets;
  final List<ScheduleTargetOption> groupClassTargets;
  final List<ScheduleStaffOption> assistantOptions;
  final List<ScheduleClassroomOption> classroomOptions;
  final ScheduleTargetOption? selectedTarget;
  final Set<String> selectedAssistantIds;
  final ScheduleClassroomOption? selectedClassroom;
  final bool loading;
  final String? errorMessage;
  final ValueChanged<_ScheduleMode> onModeChanged;
  final ValueChanged<ScheduleTargetOption> onTargetSelected;
  final ValueChanged<ScheduleStaffOption> onAssistantToggled;
  final ValueChanged<ScheduleClassroomOption?> onClassroomSelected;
  final VoidCallback onRefresh;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    final List<ScheduleTargetOption> targets =
        mode == _ScheduleMode.oneToOne ? oneToOneTargets : groupClassTargets;
    return Material(
      color: Colors.transparent,
      child: Container(
        height: 334,
        padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
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
          children: <Widget>[
            Row(
              children: <Widget>[
                _ScheduleModeSwitch(
                  width: 108,
                  mode: mode,
                  onChanged: onModeChanged,
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    mode == _ScheduleMode.oneToOne
                        ? '选择 1v1 后立即检测本周空闲点'
                        : '选择班课后立即检测本周空闲点',
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                if (loading)
                  const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(
                      strokeWidth: 1.8,
                      color: _SmartColors.orangeDeep,
                    ),
                  )
                else
                  _PanelIconButton(
                    icon: Icons.refresh_rounded,
                    onTap: onRefresh,
                  ),
                const SizedBox(width: 6),
                _PanelIconButton(
                  icon: Icons.close_rounded,
                  onTap: onClose,
                ),
              ],
            ),
            if (errorMessage != null) ...<Widget>[
              const SizedBox(height: 8),
              _PanelErrorBar(message: errorMessage!, onTap: onRefresh),
            ],
            const SizedBox(height: 10),
            Expanded(
              child: Row(
                children: <Widget>[
                  Expanded(
                    flex: 11,
                    child: _ScheduleTargetColumn(
                      title: mode == _ScheduleMode.oneToOne ? '排课对象' : '排课班级',
                      emptyText:
                          mode == _ScheduleMode.oneToOne ? '暂无可排1v1' : '暂无可排班课',
                      targets: targets,
                      selectedTarget: selectedTarget,
                      onSelected: onTargetSelected,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    flex: 8,
                    child: _ScheduleAssistantColumn(
                      assistants: assistantOptions,
                      selectedIds: selectedAssistantIds,
                      onToggled: onAssistantToggled,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    flex: 8,
                    child: _ScheduleClassroomColumn(
                      classrooms: classroomOptions,
                      selectedClassroom: selectedClassroom,
                      onSelected: onClassroomSelected,
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

class _PanelIconButton extends StatelessWidget {
  const _PanelIconButton({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(9),
      child: Container(
        width: 30,
        height: 30,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: const Color(0xFFFFF7EE),
          border: Border.all(color: _SmartColors.lineSoft),
          borderRadius: BorderRadius.circular(9),
        ),
        child: Icon(icon, color: _SmartColors.text, size: 17),
      ),
    );
  }
}

class _PanelErrorBar extends StatelessWidget {
  const _PanelErrorBar({required this.message, required this.onTap});

  final String message;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        height: 30,
        padding: const EdgeInsets.symmetric(horizontal: 10),
        decoration: BoxDecoration(
          color: const Color(0xFFFFEFEA),
          border: Border.all(color: const Color(0xFFF4C8BB)),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Row(
          children: <Widget>[
            const Icon(
              Icons.info_outline_rounded,
              color: _SmartColors.orangeDeep,
              size: 15,
            ),
            const SizedBox(width: 6),
            Expanded(
              child: Text(
                message,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _SmartColors.orangeDeep,
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScheduleTargetColumn extends StatelessWidget {
  const _ScheduleTargetColumn({
    required this.title,
    required this.emptyText,
    required this.targets,
    required this.selectedTarget,
    required this.onSelected,
  });

  final String title;
  final String emptyText;
  final List<ScheduleTargetOption> targets;
  final ScheduleTargetOption? selectedTarget;
  final ValueChanged<ScheduleTargetOption> onSelected;

  @override
  Widget build(BuildContext context) {
    return _PanelSection(
      title: title,
      child: targets.isEmpty
          ? _PanelEmptyText(emptyText)
          : ListView.separated(
              physics: const BouncingScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                final ScheduleTargetOption target = targets[index];
                return _ScheduleTargetPanelItem(
                  target: target,
                  selected: selectedTarget?.id == target.id,
                  onTap: () => onSelected(target),
                );
              },
              separatorBuilder: (_, __) => const SizedBox(height: 7),
              itemCount: targets.length,
            ),
    );
  }
}

class _ScheduleTargetPanelItem extends StatelessWidget {
  const _ScheduleTargetPanelItem({
    required this.target,
    required this.selected,
    required this.onTap,
  });

  final ScheduleTargetOption target;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      key: ValueKey<String>('schedule-target-${target.id}'),
      onTap: onTap,
      borderRadius: BorderRadius.circular(11),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 160),
        constraints: const BoxConstraints(minHeight: 48),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF1E8) : Colors.white,
          border: Border.all(
            color: selected ? _SmartColors.orange : _SmartColors.lineSoft,
          ),
          borderRadius: BorderRadius.circular(11),
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
                selected ? Icons.check_rounded : Icons.menu_book_outlined,
                color: selected ? Colors.white : _SmartColors.text,
                size: 15,
              ),
            ),
            const SizedBox(width: 9),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: <Widget>[
                  Text(
                    target.title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: target.disabled
                          ? _SmartColors.muted
                          : _SmartColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  if (target.subtitle.trim().isNotEmpty) ...<Widget>[
                    const SizedBox(height: 3),
                    Text(
                      target.subtitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _SmartColors.muted,
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScheduleAssistantColumn extends StatelessWidget {
  const _ScheduleAssistantColumn({
    required this.assistants,
    required this.selectedIds,
    required this.onToggled,
  });

  final List<ScheduleStaffOption> assistants;
  final Set<String> selectedIds;
  final ValueChanged<ScheduleStaffOption> onToggled;

  @override
  Widget build(BuildContext context) {
    return _PanelSection(
      title: '上课助教',
      child: assistants.isEmpty
          ? const _PanelEmptyText('当前组暂无其他老师')
          : ListView.separated(
              physics: const BouncingScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                final ScheduleStaffOption assistant = assistants[index];
                return _ScheduleCheckItem(
                  title: assistant.name,
                  subtitle: assistant.subtitle,
                  selected: selectedIds.contains(assistant.id),
                  onTap: () => onToggled(assistant),
                );
              },
              separatorBuilder: (_, __) => const SizedBox(height: 7),
              itemCount: assistants.length,
            ),
    );
  }
}

class _ScheduleClassroomColumn extends StatelessWidget {
  const _ScheduleClassroomColumn({
    required this.classrooms,
    required this.selectedClassroom,
    required this.onSelected,
  });

  final List<ScheduleClassroomOption> classrooms;
  final ScheduleClassroomOption? selectedClassroom;
  final ValueChanged<ScheduleClassroomOption?> onSelected;

  @override
  Widget build(BuildContext context) {
    return _PanelSection(
      title: '上课教室',
      child: ListView.separated(
        physics: const BouncingScrollPhysics(),
        itemBuilder: (BuildContext context, int index) {
          if (index == 0) {
            return _ScheduleCheckItem(
              title: '不指定教室',
              subtitle: '不校验教室占用冲突',
              selected: selectedClassroom == null,
              onTap: () => onSelected(null),
            );
          }
          final ScheduleClassroomOption classroom = classrooms[index - 1];
          return _ScheduleCheckItem(
            title: classroom.name,
            subtitle: classroom.subtitle.trim().isEmpty
                ? '校验教室占用冲突'
                : '${classroom.subtitle} · 校验教室占用',
            selected: selectedClassroom?.id == classroom.id,
            onTap: () => onSelected(classroom),
          );
        },
        separatorBuilder: (_, __) => const SizedBox(height: 7),
        itemCount: classrooms.length + 1,
      ),
    );
  }
}

class _ScheduleCheckItem extends StatelessWidget {
  const _ScheduleCheckItem({
    required this.title,
    required this.subtitle,
    required this.selected,
    required this.onTap,
  });

  final String title;
  final String subtitle;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(11),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 160),
        constraints: const BoxConstraints(minHeight: 44),
        padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 7),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF1E8) : Colors.white,
          border: Border.all(
            color: selected ? _SmartColors.orange : _SmartColors.lineSoft,
          ),
          borderRadius: BorderRadius.circular(11),
        ),
        child: Row(
          children: <Widget>[
            Icon(
              selected
                  ? Icons.check_circle_rounded
                  : Icons.radio_button_unchecked_rounded,
              color: selected ? _SmartColors.orange : _SmartColors.muted,
              size: 17,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 12,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  if (subtitle.trim().isNotEmpty) ...<Widget>[
                    const SizedBox(height: 2),
                    Text(
                      subtitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _SmartColors.muted,
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PanelSection extends StatelessWidget {
  const _PanelSection({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(9, 9, 9, 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBF7),
        border: Border.all(color: _SmartColors.lineSoft),
        borderRadius: BorderRadius.circular(13),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Padding(
            padding: const EdgeInsets.only(left: 2, bottom: 8),
            child: Text(
              title,
              style: const TextStyle(
                color: _SmartColors.ink,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          Expanded(child: child),
        ],
      ),
    );
  }
}

class _PanelEmptyText extends StatelessWidget {
  const _PanelEmptyText(this.text);

  final String text;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Text(
        text,
        style: const TextStyle(
          color: _SmartColors.muted,
          fontSize: 12,
          fontWeight: FontWeight.w800,
        ),
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

enum _ScheduleMode {
  oneToOne,
  groupClass,
}

enum _ScheduleMessageTone {
  info,
  success,
}

extension _ScheduleMessageToneView on _ScheduleMessageTone {
  IconData get icon {
    switch (this) {
      case _ScheduleMessageTone.info:
        return Icons.info_outline_rounded;
      case _ScheduleMessageTone.success:
        return Icons.check_circle_outline_rounded;
    }
  }

  Color get foreground {
    switch (this) {
      case _ScheduleMessageTone.info:
        return _SmartColors.orangeDeep;
      case _ScheduleMessageTone.success:
        return _SmartColors.green;
    }
  }

  Color get textColor {
    switch (this) {
      case _ScheduleMessageTone.info:
        return _SmartColors.ink;
      case _ScheduleMessageTone.success:
        return const Color(0xFF426D44);
    }
  }

  Color get background {
    switch (this) {
      case _ScheduleMessageTone.info:
        return const Color(0xFFFFF8EE);
      case _ScheduleMessageTone.success:
        return const Color(0xFFF0FAEF);
    }
  }

  Color get border {
    switch (this) {
      case _ScheduleMessageTone.info:
        return const Color(0xFFF0DDC9);
      case _ScheduleMessageTone.success:
        return const Color(0xFFCBEACB);
    }
  }
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
