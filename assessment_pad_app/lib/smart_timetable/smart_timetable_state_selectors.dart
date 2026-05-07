part of '../smart_timetable_page.dart';

extension _SmartTimetableStateSelectors on _SmartTimetablePageState {
  List<TimetableItem> get _visibleTimetableItems {
    if (_selectedStudentFilters.isEmpty &&
        _selectedCourseFilters.isEmpty &&
        _selectedCallStatusFilters.isEmpty) {
      return _data.items;
    }
    return _data.items.where((TimetableItem item) {
      if (_selectedStudentFilters.isNotEmpty &&
          !_selectedStudentFilters.contains(_studentFilterValue(item))) {
        return false;
      }
      if (_selectedCourseFilters.isNotEmpty &&
          !_selectedCourseFilters.contains(_courseFilterValue(item))) {
        return false;
      }
      if (_selectedCallStatusFilters.isNotEmpty) {
        final String statusKey = _callStatusFilterValue(item);
        if (statusKey.isEmpty ||
            !_selectedCallStatusFilters.contains(statusKey)) {
          return false;
        }
      }
      return true;
    }).toList();
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

  List<_TimetableFilterOption> get _studentFilterOptions {
    final Map<String, _TimetableFilterOption> options =
        <String, _TimetableFilterOption>{};
    for (final TimetableItem item in _data.items) {
      final String value = _studentFilterValue(item);
      if (value.trim().isEmpty || options.containsKey(value)) {
        continue;
      }
      options[value] = _TimetableFilterOption(id: value, label: value);
    }
    return options.values.toList();
  }

  List<_TimetableFilterOption> get _courseFilterOptions {
    final Map<String, _TimetableFilterOption> options =
        <String, _TimetableFilterOption>{};
    for (final TimetableItem item in _data.items) {
      final String value = _courseFilterValue(item);
      if (value.trim().isEmpty || options.containsKey(value)) {
        continue;
      }
      options[value] = _TimetableFilterOption(id: value, label: value);
    }
    return options.values.toList();
  }

  List<_TimetableFilterOption> get _callStatusFilterOptions {
    return const <_TimetableFilterOption>[
      _TimetableFilterOption(id: 'unsigned', label: '未点名'),
      _TimetableFilterOption(id: 'signed', label: '已点名'),
      _TimetableFilterOption(id: 'partial', label: '部分点名'),
    ];
  }

  List<List<_LessonCell?>> get _visibleScheduleRows {
    if (_selectedStudentFilters.isEmpty &&
        _selectedCourseFilters.isEmpty &&
        _selectedCallStatusFilters.isEmpty) {
      return _scheduleRows;
    }
    return _scheduleRowsFromItems(
      _visibleTimetableItems,
      _data.days,
      _weekDays,
      _timeSlots,
    );
  }

  TimetableSummary get _visibleTimetableSummary {
    return _timetableSummaryFromItems(_visibleTimetableItems);
  }

  String _filterButtonLabel(_TimetableFilterKind kind) {
    final (
      String emptyLabel,
      Set<String> selectedIds,
      List<_TimetableFilterOption> options
    ) = switch (kind) {
      _TimetableFilterKind.student => (
          '上课学员',
          _selectedStudentFilters,
          _studentFilterOptions,
        ),
      _TimetableFilterKind.course => (
          '全部课程',
          _selectedCourseFilters,
          _courseFilterOptions,
        ),
      _TimetableFilterKind.callStatus => (
          '点名状态',
          _selectedCallStatusFilters,
          _callStatusFilterOptions,
        ),
    };
    if (selectedIds.isEmpty) {
      return emptyLabel;
    }
    final List<String> selectedLabels = options
        .where((option) => selectedIds.contains(option.id))
        .map((option) => option.label)
        .toList();
    if (selectedLabels.isEmpty) {
      return emptyLabel;
    }
    if (selectedLabels.length == 1) {
      return selectedLabels.first;
    }
    return '${selectedLabels.first}+${selectedLabels.length - 1}';
  }

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

  void _pruneSelectedAssistantsToCurrentGroup() {
    final Set<String> validAssistantIds = _currentGroupAssistantOptions
        .map((ScheduleStaffOption item) => item.id)
        .toSet();
    _selectedAssistantIds = _selectedAssistantIds
        .where((String id) => validAssistantIds.contains(id.trim()))
        .toSet();
  }

  void _pruneQuickFilters() {
    final Set<String> validStudentFilters =
        _studentFilterOptions.map((item) => item.id).toSet();
    final Set<String> validCourseFilters =
        _courseFilterOptions.map((item) => item.id).toSet();
    final Set<String> validCallStatusFilters =
        _callStatusFilterOptions.map((item) => item.id).toSet();
    _selectedStudentFilters =
        _selectedStudentFilters.where(validStudentFilters.contains).toSet();
    _selectedCourseFilters =
        _selectedCourseFilters.where(validCourseFilters.contains).toSet();
    _selectedCallStatusFilters = _selectedCallStatusFilters
        .where(validCallStatusFilters.contains)
        .toSet();
  }

  void _togglePeriodGroupDropdown() {
    _updateState(() {
      _periodGroupDropdownOpen = !_periodGroupDropdownOpen;
      _teacherDropdownOpen = false;
      _schedulePanelOpen = false;
      _openFilterKind = null;
    });
  }

  void _closePeriodGroupDropdown() {
    if (!_periodGroupDropdownOpen) {
      return;
    }
    _updateState(() {
      _periodGroupDropdownOpen = false;
    });
  }

  void _toggleFilterPanel(_TimetableFilterKind kind) {
    _updateState(() {
      _openFilterKind = _openFilterKind == kind ? null : kind;
      _periodGroupDropdownOpen = false;
      _teacherDropdownOpen = false;
      _schedulePanelOpen = false;
    });
  }

  void _closeFilterPanel() {
    if (_openFilterKind == null) {
      return;
    }
    _updateState(() {
      _openFilterKind = null;
    });
  }

  void _toggleStudentFilter(String value) {
    _toggleFilterValue(_selectedStudentFilters, value);
  }

  void _toggleCourseFilter(String value) {
    _toggleFilterValue(_selectedCourseFilters, value);
  }

  void _toggleCallStatusFilter(String value) {
    _toggleFilterValue(_selectedCallStatusFilters, value);
  }

  void _clearFilter(_TimetableFilterKind kind) {
    _updateState(() {
      if (kind == _TimetableFilterKind.student) {
        _selectedStudentFilters.clear();
      } else if (kind == _TimetableFilterKind.course) {
        _selectedCourseFilters.clear();
      } else {
        _selectedCallStatusFilters.clear();
      }
    });
  }

  void _toggleFilterValue(Set<String> target, String value) {
    final String normalized = value.trim();
    if (normalized.isEmpty) {
      return;
    }
    _updateState(() {
      if (!target.add(normalized)) {
        target.remove(normalized);
      }
    });
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
}
