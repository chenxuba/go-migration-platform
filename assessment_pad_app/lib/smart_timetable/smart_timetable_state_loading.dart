part of '../smart_timetable_page.dart';

extension _SmartTimetableStateLoading on _SmartTimetablePageState {
  Future<String> _readAuthToken() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(_timetableAuthTokenStorageKey) ?? '';
  }

  Future<void> _loadScheduleOptions() async {
    final int sequence = ++_scheduleOptionsSequence;
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      return;
    }
    if (mounted) {
      _updateState(() {
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
      _updateState(() {
        _oneToOneTargets = oneToOneTargets;
        _groupClassTargets = groupClassTargets;
        _assistantOptions = assistantOptions;
        _classroomOptions = classroomOptions;
        _scheduleOptionsLoaded = true;
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
      _updateState(() {
        _scheduleOptionsLoading = false;
        _scheduleOptionsError = error.message;
      });
    } on Object catch (error) {
      if (!mounted || sequence != _scheduleOptionsSequence) {
        return;
      }
      _updateState(() {
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

  Future<void> _loadTimetable({
    String? teacherId,
    String? periodGroupId,
    bool showSkeleton = true,
  }) async {
    final int sequence = ++_loadSequence;
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      if (!mounted || sequence != _loadSequence) {
        return;
      }
      _updateState(() {
        _loading = false;
        _errorMessage = '登录已失效，请重新登录';
      });
      return;
    }
    final _WeekRange range = _weekRange(_weekOffset);
    if (mounted && showSkeleton) {
      _updateState(() {
        _loading = true;
        _errorMessage = null;
      });
    } else if (mounted) {
      _updateState(() {
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
      _updateState(() {
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
      _updateState(() {
        _loading = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || sequence != _loadSequence) {
        return;
      }
      _updateState(() {
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
    _updateState(() {
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
    _updateState(() {
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
    _updateState(() {
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
    _updateState(() {
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
}
